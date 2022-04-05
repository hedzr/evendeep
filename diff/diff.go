package diff

import (
	"fmt"
	"github.com/hedzr/evendeep/internal/natsort"
	"github.com/hedzr/evendeep/internal/tool"
	"github.com/hedzr/evendeep/typ"
	"reflect"
	"strings"
	"time"
	"unsafe"
)

func New(lhs, rhs typ.Any, opts ...Opt) (inf *info, equal bool) {
	inf = newInfo()
	for _, opt := range opts {
		if opt != nil {
			opt(inf)
		}
	}
	equal = inf.diff(lhs, rhs)
	return
}

type Opt func(*info)

func WithIgnoredFields(names ...string) Opt {
	return func(i *info) {
		for _, name := range names {
			i.ignoredFields[name] = true
		}
	}
}

func WithSliceOrderedComparison(b bool) Opt {
	return func(i *info) {
		i.sliceNoOrder = b
	}
}

type Diff interface {
	ForAdded(fn func(key string, val typ.Any))
	ForRemoved(fn func(key string, val typ.Any))
	ForModified(fn func(key string, val Update))

	PrettyPrint() string
	String() string
}

func newInfo() *info {
	return &info{
		added:         make(map[string]typ.Any),
		removed:       make(map[string]typ.Any),
		modified:      make(map[string]Update),
		pathTable:     make(map[string]dottedPath),
		visited:       make(map[visit]bool),
		ignoredFields: make(map[string]bool),
		sliceNoOrder:  true,
	}
}

type info struct {
	added         map[string]typ.Any
	removed       map[string]typ.Any
	modified      map[string]Update
	pathTable     map[string]dottedPath
	visited       map[visit]bool
	ignoredFields map[string]bool
	sliceNoOrder  bool
}

func (d *info) String() string { return d.PrettyPrint() }

func (d *info) PrettyPrint() string {
	var lines []string
	if d != nil {
		d.forMap(d.added, func(key string, val typ.Any) {
			lines = append(lines, fmt.Sprintf("added: %s = %v\n", key, val))
		})
		for key, val := range d.modified {
			if val.Old == nil {
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
	copym3 := func(m1 map[string]dottedPath) map[string]dottedPath {
		m2 := make(map[string]dottedPath)
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

func (d *info) mkkey(path dottedPath, parts ...pathPart) (key string) {
	dp := path.appendAndNew(parts...)
	key = dp.String()
	d.pathTable[key] = dp
	return
}

func (d *info) diff(lhs, rhs typ.Any) bool {
	lv, rv := reflect.ValueOf(lhs), reflect.ValueOf(rhs)
	var path dottedPath
	return d.diffv(lv, rv, path)
}

func (d *info) diffv(lv, rv reflect.Value, path dottedPath) (equal bool) {
	lvv, rvv := lv.IsValid(), rv.IsValid()
	if !lvv && !rvv {
		return true
	}

	if !lvv {
		d.modified[d.mkkey(path)] = Update{Old: nil, New: tool.Valfmt(&rv), Typ: typfmtlite(&rv)}
		return
	}
	if !rvv {
		d.modified[d.mkkey(path)] = Update{Old: tool.Valfmt(&lv), New: nil, Typ: typfmtlite(&lv)}
		return
	}

	lvt, rvt := lv.Type(), rv.Type()
	if lvt != rvt {
		d.modified[d.mkkey(path)] = Update{Old: tool.Valfmt(&lv), New: tool.Valfmt(&rv), Typ: typfmtlite(&rv)}
		return
	}

	kind := lv.Kind()

	if lv.CanAddr() && rv.CanAddr() && kindis(kind, reflect.Array, reflect.Map, reflect.Slice, reflect.Struct) {
		addr1 := unsafe.Pointer(lv.UnsafeAddr())
		addr2 := unsafe.Pointer(rv.UnsafeAddr())
		if uintptr(addr1) > uintptr(addr2) {
			// Canonicalize order to reduce number of entries in visited.
			// Assumes non-moving garbage collector.
			addr1, addr2 = addr2, addr1
		}

		// Short circuit if references are already seen.
		v := visit{addr1, addr2, lvt}
		if d.visited[v] {
			return true
		}

		// Remember for later.
		d.visited[v] = true
	}

	switch kind {
	case reflect.Map, reflect.Ptr, reflect.Func, reflect.Chan, reflect.Slice:
		ln, rn := tool.IsNil(lv), tool.IsNil(lv)
		if ln && rn {
			return true
		}
		if ln || rn {
			if (kind == reflect.Slice || kind == reflect.Map) && lv.Len() == rv.Len() {
				return true
			}
			d.modified[d.mkkey(path)] = Update{Old: tool.Valfmt(&lv), New: tool.Valfmt(&rv), Typ: typfmtlite(&lv)}
			return false
		}
	}

	switch kind {
	case reflect.Array:
		equal = d.diffArray(lv, rv, path)

	case reflect.Slice:
		if d.sliceNoOrder {
			equal = d.diffSliceNoOrder(lv, rv, path)
		} else {
			equal = d.diffArray(lv, rv, path)
		}

	case reflect.Map:
		equal = true
		for _, key := range lv.MapKeys() {
			aI, bI := lv.MapIndex(key), rv.MapIndex(key)
			localPath := path.appendAndNew(mapKey{key.Interface()})
			if !bI.IsValid() {
				d.removed[d.mkkey(localPath)] = tool.Valfmt(&aI)
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
				d.added[d.mkkey(localPath)] = tool.Valfmt(&bI)
				equal = false
			}
		}

	case reflect.Struct:
		equal = true
		// If the field is time.Time, use Equal to compare
		if lvt.String() == "time.Time" {
			aTime := lv.Interface().(time.Time)
			bTime := rv.Interface().(time.Time)
			if !aTime.Equal(bTime) {
				d.modified[d.mkkey(path)] = Update{Old: aTime.String(), New: bTime.String(), Typ: typfmtlite(&lv)}
				equal = false
			}
		} else {
			for i := 0; i < lvt.NumField(); i++ {
				index := []int{i}
				field := lvt.FieldByIndex(index)
				if vk := field.Tag.Get("diff"); vk == "ignore" || vk == "-" { // skip fields marked to be ignored
					continue
				}
				if _, skip := d.ignoredFields[field.Name]; skip {
					continue
				}
				localPath := path.appendAndNew(structField(field.Name))
				aI := tool.UnsafeReflectValue(lv.FieldByIndex(index))
				bI := tool.UnsafeReflectValue(rv.FieldByIndex(index))
				if eq := d.diffv(aI, bI, localPath); !eq {
					equal = false
				}
			}
		}

	case reflect.Ptr:
		equal = d.diffv(lv.Elem(), rv.Elem(), path)

	default:
		if reflect.DeepEqual(lv.Interface(), rv.Interface()) {
			equal = true
		} else {
			d.modified[d.mkkey(path)] = Update{Old: tool.Valfmt(&lv), New: tool.Valfmt(&rv), Typ: typfmtlite(&lv)}
			equal = false
		}
	}

	return
}

func (d *info) diffArray(lv, rv reflect.Value, path dottedPath) (equal bool) {
	ll, rl := lv.Len(), rv.Len()
	equal = true
	for i := 0; i < tool.MinInt(ll, rl); i++ {
		localPath := path.appendAndNew(sliceIndex(i))
		if eq := d.diffv(lv.Index(i), rv.Index(i), localPath); !eq {
			equal = false
		}
	}
	if ll > rl {
		for i := rl; i < ll; i++ {
			localPath := path.appendAndNew(sliceIndex(i))
			v := lv.Index(i)
			d.removed[d.mkkey(localPath)] = tool.Valfmt(&v)
			equal = false
		}
	} else if ll < rl {
		for i := ll; i < rl; i++ {
			localPath := path.appendAndNew(sliceIndex(i))
			v := rv.Index(i)
			d.added[d.mkkey(localPath)] = tool.Valfmt(&v)
			equal = false
		}
	}
	return
}

func (d *info) diffSliceNoOrder(lv, rv reflect.Value, path dottedPath) (equal bool) {
	ll, rl := lv.Len(), rv.Len()
	if ll != rl {
		return d.diffArray(lv, rv, path)
	}

	equal = true
	m := make(map[int]bool)
	for i := 0; i < ll; i++ {
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
			d.removed[d.mkkey(localPath)] = tool.Valfmt(&lvit)
			equal = false
		}
	}
	for i := 0; i < ll; i++ {
		localPath := path.appendAndNew(sliceIndex(i))
		if _, ok := m[i]; ok {
			continue
		}
		rvit := rv.Index(i)
		d.added[d.mkkey(localPath)] = tool.Valfmt(&rvit)
		equal = false
	}
	return
}
