package diff

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/hedzr/evendeep/dbglog"
	"github.com/hedzr/evendeep/internal/natsort"
	"github.com/hedzr/evendeep/internal/tool"
	"github.com/hedzr/evendeep/ref"
	"github.com/hedzr/evendeep/typ"
	unsafe2 "github.com/hedzr/evendeep/unsafe"
)

// New compares two value deeply and returns the Diff of them.
//
// A Diff includes added, removed, and modified records, you can
// PrettyPrint them for displaying.
//
//	delta, equal := evendeep.DeepDiff([]int{3, 0}, []int{9, 3, 0})
//	t.Logf("delta: %v", delta)
//	//        added: [2] = <zero>
//	//        modified: [0] = 9 (int) (Old: 3)
//	//        modified: [1] = 3 (int) (Old: <zero>)
func New(lhs, rhs typ.Any, opts ...Opt) (inf Diff, equal bool) {
	info1 := newInfo(opts...)
	equal = info1.diff(lhs, rhs)
	inf = info1
	return
}

// Diff includes added, removed and modified records of the two values.
type Diff interface {
	ForAdded(fn func(key string, val typ.Any))
	ForRemoved(fn func(key string, val typ.Any))
	ForModified(fn func(key string, val Update))

	PrettyPrint() string
	String() string
}

func newInfo(opts ...Opt) *info {
	inf := &info{
		added:                    make(map[string]typ.Any),
		removed:                  make(map[string]typ.Any),
		modified:                 make(map[string]Update),
		pathTable:                make(map[string]Path),
		visited:                  make(map[visit]bool),
		ignoredFields:            make(map[string]bool),
		sliceNoOrder:             false,
		stripPtr1st:              false,
		treatEmptyStructPtrAsNil: false,
		differentTypeStructs:     false,
		compares: []Comparer{
			&timeComparer{},
			&bytesBufferComparer{},
		},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(inf)
		}
	}
	return inf
}

type info struct {
	added                    map[string]typ.Any
	removed                  map[string]typ.Any
	modified                 map[string]Update
	pathTable                map[string]Path
	visited                  map[visit]bool
	ignoredFields            map[string]bool
	sliceNoOrder             bool
	stripPtr1st              bool
	treatEmptyStructPtrAsNil bool
	differentTypeStructs     bool
	ignoreUnmatchedFields    bool
	differentSizeArrays      bool
	compares                 []Comparer
}

// - Comparer

func (d *info) PutAdded(k string, v typ.Any)                { d.added[k] = v }
func (d *info) PutRemoved(k string, v typ.Any)              { d.removed[k] = v }
func (d *info) PutModified(k string, v Update)              { d.modified[k] = v }
func (d *info) PutPath(path Path, parts ...PathPart) string { return d.mkkey(path, parts...) }

// - Stringer

func (d *info) String() string { return d.PrettyPrint() }

// - Diff

func (d *info) PrettyPrint() string { //nolint:revive
	var lines []string
	if d != nil {
		d.forMap(d.added, func(key string, val typ.Any) {
			lines = append(lines, fmt.Sprintf("added: %s = %v\n", key, val))
		})
		for key, val := range d.modified {
			if val.Old == nil { //nolint:gocritic // no need to switch to 'switch' clausev
				lines = append(lines, fmt.Sprintf("modified: %s = %v (%v) (Old: nil)\n",
					key, val.New, val.Typ))
			} else if val.New == nil {
				lines = append(lines, fmt.Sprintf("modified: %s = nil (Old: %v (%v))\n",
					key, val.Old, val.Typ))
			} else {
				lines = append(lines, fmt.Sprintf("modified: %s = %v (%v) (Old: %v)\n",
					key, val.New, val.Typ, val.Old))
			}
		}
		d.forMap(d.removed, func(key string, val typ.Any) {
			lines = append(lines, fmt.Sprintf("removed: %s = %v\n", key, val))
		})
	}

	natsort.Strings(lines)
	return strings.Join(lines, "")
}

func (d *info) ForAdded(fn func(key string, val typ.Any))   { d.forMap(d.added, fn) }
func (d *info) ForRemoved(fn func(key string, val typ.Any)) { d.forMap(d.removed, fn) }
func (d *info) ForModified(fn func(key string, val Update)) {
	for k, v := range d.modified {
		fn(k, v)
	}
}

func (d *info) forMap(m map[string]typ.Any, fn func(key string, val typ.Any)) {
	for k, v := range m {
		fn(k, v)
	}
}

// - Cloneable

func (d *info) Clone() *info {
	copym1 := func(m1 map[string]typ.Any) map[string]typ.Any {
		m2 := make(map[string]typ.Any)
		for k, v := range m1 {
			m2[k] = v
		}
		return m2
	}
	copym2 := func(m1 map[string]Update) map[string]Update {
		m2 := make(map[string]Update)
		for k, v := range m1 {
			m2[k] = v
		}
		return m2
	}
	copym3 := func(m1 map[string]Path) map[string]Path {
		m2 := make(map[string]Path)
		for k, v := range m1 {
			m2[k] = v
		}
		return m2
	}
	copym4 := func(m1 map[visit]bool) map[visit]bool {
		m2 := make(map[visit]bool)
		for k, v := range m1 {
			m2[k] = v
		}
		return m2
	}
	copym5 := func(m1 map[string]bool) map[string]bool {
		m2 := make(map[string]bool)
		for k, v := range m1 {
			m2[k] = v
		}
		return m2
	}
	return &info{
		added:         copym1(d.added),
		removed:       copym1(d.removed),
		modified:      copym2(d.modified),
		pathTable:     copym3(d.pathTable),
		visited:       copym4(d.visited),
		ignoredFields: copym5(d.ignoredFields),
		sliceNoOrder:  d.sliceNoOrder,
	}
}

//

func (d *info) mkkey(path Path, parts ...PathPart) (key string) {
	dp := path.appendAndNew(parts...)
	key = dp.String()
	d.pathTable[key] = dp
	return
}

func (d *info) diff(lhs, rhs typ.Any) bool {
	lv, rv := reflect.ValueOf(lhs), reflect.ValueOf(rhs)
	if d.stripPtr1st {
		lv, rv = ref.Rdecodesimple(lv), ref.Rdecodesimple(rv)
	} else {
		_, lv = ref.Rskip(lv, reflect.Interface)
		_, rv = ref.Rskip(rv, reflect.Interface)
	}
	var path Path
	return d.diffv(lv, rv, path)
}

func (d *info) diffv(lv, rv reflect.Value, path Path) (equal bool) { //nolint:revive
	var processed bool

	lvv, rvv := lv.IsValid(), rv.IsValid()
	if equal, processed = d.testinvalid(lv, rv, lvv, rvv, path); processed {
		dbglog.Log("  - Invalid object found: l = %v, r = %v, path = %v", ref.Valfmt(&lv), ref.Valfmtptr(&rv), path)
		return
	}

	lvt, rvt := lv.Type(), rv.Type()
	if lvt != rvt {
		dbglog.Log("  - Unmatched type found: l = %v, r = %v, path = %v", ref.Typfmt(lvt), ref.Typfmt(rvt), path)
		if d.differentTypeStructs && lv.Kind() == reflect.Struct && rv.Kind() == reflect.Struct {
			return d.compareStructFields(lv, rv, path)
		}
		if d.differentSizeArrays && lv.Kind() == reflect.Array && rv.Kind() == reflect.Array {
			return d.compareArrayDifferSizes(lv, rv, path)
		}
		d.PutModified(d.mkkey(path), Update{Old: ref.Valfmt(&lv), New: ref.Valfmt(&rv), Typ: ref.Typfmtvlite(&rv)})
		return
	}

	kind := lv.Kind()

	if equal, processed = d.testvisited(lv, rv, lvt, path, kind); processed {
		return
	}

	if equal, processed = d.testnil(lv, rv, lvt, path, kind); processed {
		return
	}

	if equal, processed = d.testcomparer(lv, rv, lvt, path); processed {
		return
	}

	return d.diffw(lv, rv, lvt, path, kind)
}

func (d *info) testinvalid(lv, rv reflect.Value, lvv, rvv bool, path Path) (equal, processed bool) { //nolint:revive
	if !lvv && !rvv {
		return true, true
	}

	if !lvv {
		d.PutModified(d.mkkey(path), Update{Old: nil, New: ref.Valfmt(&rv), Typ: ref.Typfmtvlite(&rv)})
		return false, true
	}
	if !rvv {
		d.PutModified(d.mkkey(path), Update{Old: ref.Valfmt(&lv), New: nil, Typ: ref.Typfmtvlite(&lv)})
		return false, true
	}
	return
}

func (d *info) testvisited(lv, rv reflect.Value, typ1 reflect.Type,
	path Path,
	kind reflect.Kind,
) (equal, processed bool) {
	if lv.CanAddr() && rv.CanAddr() &&
		ref.KindIs(kind, reflect.Array, reflect.Map, reflect.Slice, reflect.Struct) {
		addr1 := unsafe.Pointer(lv.UnsafeAddr())
		addr2 := unsafe.Pointer(rv.UnsafeAddr())
		if uintptr(addr1) > uintptr(addr2) {
			// Canonicalize order to reduce number of entries in visited.
			// Assumes non-moving garbage collector.
			addr1, addr2 = addr2, addr1
		}

		// Short circuit if references are already seen.
		v := visit{addr1, addr2, typ1}
		if d.visited[v] {
			return true, true
		}

		// Remember for later.
		d.visited[v] = true
	}
	_ = path
	return
}

func (d *info) testnil(lv, rv reflect.Value, typ1 reflect.Type, path Path, kind reflect.Kind) (equal, processed bool) { //nolint:revive
	_ = typ1
	switch kind { //nolint:exhaustive //no need
	case reflect.Map, reflect.Ptr, reflect.Func, reflect.Chan, reflect.Slice:
		ln, rn := ref.IsNil(lv), ref.IsNil(rv)
		if ln && rn {
			return true, true
		}
		if ln || rn {
			if kind == reflect.Slice || kind == reflect.Map {
				// le, re := tool.IsZero(lv), tool.IsZero(rv)
				// if equal = le == re; equal {
				// 	return true, true
				// }

				// By default, the nil array are equal to an empty array

				// if d.treatEmptyStructPtrAsNil {
				if (ln) && isEmptyObject(rv) {
					return true, true
				} else if (rn) && isEmptyObject(lv) {
					return true, true
				}
				// }
			} else if kind == reflect.Struct && d.treatEmptyStructPtrAsNil {
				if ln && isEmptyStruct(rv) {
					return true, true
				}
				if rn && isEmptyStruct(lv) {
					return true, true
				}
			}
			d.PutModified(d.mkkey(path), Update{Old: ref.Valfmt(&lv), New: ref.Valfmt(&rv), Typ: ref.Typfmtvlite(&lv)})
			return false, true
		}
	}
	return
}

func (d *info) testcomparer(lv, rv reflect.Value, typ1 reflect.Type, path Path) (equal, processed bool) {
	var c Comparer
	if c, processed = d.findComparer(typ1); processed {
		equal = c.Equal(d, lv, rv, path)
	}
	return
}

func (d *info) findComparer(typ1 reflect.Type) (c Comparer, ok bool) {
	for _, c = range d.compares {
		if ok = c.Match(typ1); ok {
			break
		}
	}
	return
}

func (d *info) diffw(lv, rv reflect.Value, typ1 reflect.Type, path Path, kind reflect.Kind) (equal bool) { //nolint:revive
	switch kind { //nolint:exhaustive //no
	case reflect.Array:
		equal = d.diffArray(lv, rv, path)

	case reflect.Slice:
		if d.sliceNoOrder {
			equal = d.diffSliceNoOrder(lv, rv, path)
		} else {
			equal = d.diffArray(lv, rv, path)
		}

	case reflect.Map:
		equal = d.diffMap(lv, rv, path)

	case reflect.Struct:
		equal = d.diffStruct(lv, rv, typ1, path)

	case reflect.Ptr:
		equal = d.diffv(lv.Elem(), rv.Elem(), path)

	case reflect.Interface:
		equal = d.diffv(lv.Elem(), rv.Elem(), path)

	case reflect.Chan:
		equal = lv.Type() == rv.Type() && lv.Cap() == rv.Cap()

	default:
		a, b := lv.Interface(), rv.Interface()
		if equal = reflect.DeepEqual(a, b); !equal {
			d.PutModified(d.mkkey(path), Update{Old: ref.Valfmt(&lv), New: ref.Valfmt(&rv), Typ: ref.Typfmtvlite(&lv)})
		}
	}

	return
}

func (d *info) diffArray(lv, rv reflect.Value, path Path) (equal bool) { //nolint:revive
	ll, rl := lv.Len(), rv.Len()
	equal = true
	for i := 0; i < tool.MinInt(ll, rl); i++ {
		localPath := path.appendAndNew(sliceIndex(i))
		aI, bI := lv.Index(i), rv.Index(i)
		if eq := d.diffv(aI, bI, localPath); !eq {
			dbglog.Log("    diffArray: [%d] not equal %v - %v", i, ref.Valfmt(&aI), ref.Valfmt(&bI))
			equal = false
		}
	}
	if ll > rl {
		for i := rl; i < ll; i++ {
			v := lv.Index(i)
			if d.differentSizeArrays && ref.IsZero(v) {
				continue
			}
			localPath := path.appendAndNew(sliceIndex(i))
			d.PutRemoved(d.mkkey(localPath), ref.Valfmt(&v))
			equal = false
		}
	} else if ll < rl {
		for i := ll; i < rl; i++ {
			v := rv.Index(i)
			if d.differentSizeArrays && ref.IsZero(v) {
				continue
			}
			localPath := path.appendAndNew(sliceIndex(i))
			d.PutAdded(d.mkkey(localPath), ref.Valfmt(&v))
			equal = false
		}
	}
	return
}

func (d *info) diffSliceNoOrder(lv, rv reflect.Value, path Path) (equal bool) {
	ll, rl := lv.Len(), rv.Len()
	equal = true
	m := make(map[int]bool)
	for i := 0; i < tool.MinInt(ll, rl); i++ {
		localPath := path.appendAndNew(sliceIndex(i))
		lvit := lv.Index(i)
		var eq bool
		for j := 0; j < rl; j++ {
			if eq = d.Clone().diffv(lvit, rv.Index(j), localPath); eq {
				m[j] = true
				break
			}
		}
		if !eq {
			d.PutRemoved(d.mkkey(localPath), ref.Valfmt(&lvit))
			equal = false
		}
	}
	for i := 0; i < rl; i++ {
		localPath := path.appendAndNew(sliceIndex(i))
		if _, ok := m[i]; ok {
			continue
		}
		rvit := rv.Index(i)
		d.PutAdded(d.mkkey(localPath), ref.Valfmt(&rvit))
		equal = false
	}
	return
}

func (d *info) diffMap(lv, rv reflect.Value, path Path) (equal bool) {
	equal = true
	for _, key := range lv.MapKeys() {
		aI, bI := lv.MapIndex(key), rv.MapIndex(key)
		localPath := path.appendAndNew(mapKey{key.Interface()})
		if !bI.IsValid() {
			d.PutRemoved(d.mkkey(localPath), ref.Valfmt(&aI))
			equal = false
		} else if eq := d.diffv(aI, bI, localPath); !eq {
			equal = false
		}
	}
	for _, key := range rv.MapKeys() {
		aI := lv.MapIndex(key)
		if !aI.IsValid() {
			bI := rv.MapIndex(key)
			localPath := path.appendAndNew(mapKey{key.Interface()})
			d.PutAdded(d.mkkey(localPath), ref.Valfmt(&bI))
			equal = false
		}
	}
	return
}

func (d *info) diffStruct(lv, rv reflect.Value, typ1 reflect.Type, path Path) (equal bool) { //nolint:revive
	equal = true
	for i := 0; i < typ1.NumField(); i++ {
		index := []int{i}
		field := typ1.FieldByIndex(index)
		if vk := field.Tag.Get("diff"); vk == "ignore" || vk == "-" { // skip fields marked to be ignored
			continue
		}
		if _, skip := d.ignoredFields[field.Name]; skip {
			continue
		}
		localPath := path.appendAndNew(structField(field.Name))
		aI := unsafe2.UnsafeReflectValue(lv.FieldByIndex(index))
		bI := unsafe2.UnsafeReflectValue(rv.FieldByIndex(index))
		if d.treatEmptyStructPtrAsNil && field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
			ln, rn := ref.IsNil(aI), ref.IsNil(bI)
			var eq bool
			if eq = ln && rn; !equal {
				if eq = ln && isEmptyStruct(bI.Elem()); !eq {
					eq = rn && isEmptyStruct(aI.Elem())
				}
			}
			if !eq {
				equal = false
				d.PutModified(d.mkkey(path), Update{Old: ref.Valfmt(&aI), New: ref.Valfmt(&bI), Typ: ref.Typfmtvlite(&aI)})
			}
			continue
		}
		if eq := d.diffv(aI, bI, localPath); !eq {
			equal = false
		}
	}
	return
}

func (d *info) compareArrayDifferSizes(lv, rv reflect.Value, path Path) (equal bool) {
	return d.diffArray(lv, rv, path)
}

func (d *info) compareStructFields(lv, rv reflect.Value, path Path) (equal bool) { //nolint:revive
	for i, lt, rt := 0, lv.Type(), rv.Type(); i < lv.NumField(); i++ {
		// fldval := lv.Field(i)
		fldtyp := lt.Field(i)
		fldname := fldtyp.Name
		if _, ok := rt.FieldByName(fldname); ok {
			l := lv.Field(i)
			r := rv.FieldByName(fldname)
			localPath := path.appendAndNew(structField(fldname))
			equal = d.diffv(l, r, localPath)
			if !equal {
				if equal = !l.IsValid() && !r.IsValid(); !equal {
					return
				} // if both l and r are invalid, assumes its equivalence
			}
		} else if !d.ignoreUnmatchedFields {
			return false
		}
	}
	return
}
