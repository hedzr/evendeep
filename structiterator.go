package deepcopy

import (
	"github.com/hedzr/log"
	"reflect"
	"strings"
	"sync"
)

//

type fieldstable struct {
	records          []tablerec
	autoexpandstruct bool
}

type tablerec struct {
	names            []string // the path from root struct, in reverse order
	indexes          []int
	structFieldValue *reflect.Value
	structField      *reflect.StructField
}

func (rec tablerec) Value() *reflect.Value {
	return rec.structFieldValue
}

func (table *fieldstable) shouldIgnore(field reflect.StructField, typ reflect.Type, kind reflect.Kind) bool {
	n := typ.PkgPath()
	return packageisreserved(n) // ignore golang stdlib, such as "io", "runtime", ...
}

func (table *fieldstable) getallfields(structValue reflect.Value, autoexpandstruct bool) fieldstable {
	table.autoexpandstruct = autoexpandstruct

	structValue, _ = rdecode(structValue)
	if structValue.Kind() != reflect.Struct {
		return *table
	}

	ret := table.getfields(structValue, "", -1)
	for _, ni := range ret.records {
		//ni.names = append(ni.names, sf.Name)
		//ni.indexes = append(ni.indexes, i)
		table.records = append(table.records, ni)
	}
	return *table
}

func (table *fieldstable) tablerec(svind *reflect.Value, sf *reflect.StructField, index, parentIndex int, parentFieldName string) (tr tablerec) {
	tr.structFieldValue = svind
	tr.structField = sf
	tr.names = append(tr.names, sf.Name)
	if parentFieldName != "" {
		tr.names = append(tr.names, parentFieldName)
	}
	tr.indexes = append(tr.indexes, index)
	if parentIndex >= 0 {
		tr.indexes = append(tr.indexes, parentIndex)
	}
	return
}

func (table *fieldstable) getfields(structValue reflect.Value, fieldname string, fi int) (ret fieldstable) {
	var i, amount int
	for i, amount = 0, structValue.NumField(); i < amount; i++ {
		var tr tablerec

		sv := structValue.Field(i)
		svind := rdecodesimple(sv)

		sf := structValue.Type().Field(i)

		//functorLog("%d, %v (%v (%v))", i, sf.Name, sf.Type, sf.Type.Kind())

		if !svind.IsValid() {
			tr = table.tablerec(&svind, &sf, i, fi, fieldname)
			ret.records = append(ret.records, tr)
			continue
		}

		svindtype := svind.Type() // may panic on a zero/invalid value
		svindtypekind := svindtype.Kind()
		isStruct := sf.Anonymous || svindtypekind == reflect.Struct
		shouldIgnored := table.shouldIgnore(sf, svindtype, svindtypekind)

		if isStruct && table.autoexpandstruct && !shouldIgnored {
			n := table.getfields(svind, sf.Name, i)
			for _, ni := range n.records {
				ret.records = append(ret.records, ni)
			}
		} else {
			tr = table.tablerec(&svind, &sf, i, fi, fieldname)
			ret.records = append(ret.records, tr)
		}
	}
	return
}

//

//

type valueIterator interface {
	Next() (accessor *fieldaccessor, ok bool)
}

type valueIteratorOpt func(s *structIterator)

//

type structIterator struct {
	rootStruct       reflect.Value
	stack            []*fieldaccessor
	autoexpandstruct bool // Next will expand *struct to struct and get inside loop deeply
}

type fieldaccessor struct {
	structvalue *reflect.Value
	structtype  reflect.Type
	index       int
	structfield *reflect.StructField
}

func (s *fieldaccessor) Type() reflect.Type { return s.structtype }
func (s *fieldaccessor) ValueValid() bool   { return s.structvalue != nil && s.structvalue.IsValid() }
func (s *fieldaccessor) FieldValue() *reflect.Value {
	if s.ValueValid() {
		r := s.structvalue.Field(s.index)
		return &r
	}
	return nil
}
func (s *fieldaccessor) StructField() *reflect.StructField {
	if s.ValueValid() {
		r := s.structvalue.Type().Field(s.index)
		s.structfield = &r
	}
	return s.structfield
}
func (s *fieldaccessor) getStructField() *reflect.StructField {
	if s.ValueValid() {
		r := s.structvalue.Type().Field(s.index)
		s.structfield = &r
	}
	return s.structfield
}
func (s *fieldaccessor) StructFieldName() string {
	fld := s.StructField()
	if fld != nil {
		return fld.Name
	}
	return ""
}
func (s *fieldaccessor) NumField() int {
	if s.ValueValid() {
		return s.structvalue.NumField()
	}
	return 0
}
func (s *fieldaccessor) incr() *fieldaccessor {
	s.index++
	return s
}

func newStructIterator(structValue reflect.Value, opts ...valueIteratorOpt) valueIterator {
	s := &structIterator{
		rootStruct: structValue,
		stack:      nil,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// withStructPtrAutoExpand allows auto-expanding the struct or its pointer
// in iterating a parent struct
func withStructPtrAutoExpand(expand bool) valueIteratorOpt {
	return func(s *structIterator) {
		s.autoexpandstruct = expand
	}
}

func (s *structIterator) _push(structvalue *reflect.Value, structtype reflect.Type, index int) *fieldaccessor {
	s.stack = append(s.stack, &fieldaccessor{structvalue, structtype, index, nil})
	return s._top()
}
func (s *structIterator) _empty() bool { return len(s.stack) == 0 }
func (s *structIterator) _pop() {
	if len(s.stack) > 0 {
		s.stack = s.stack[0 : len(s.stack)-1]
	}
}
func (s *structIterator) _top() *fieldaccessor {
	if len(s.stack) == 0 {
		return nil
	}
	return s.stack[len(s.stack)-1]
}
func (s *structIterator) _prev() *fieldaccessor {
	if len(s.stack) <= 1 {
		return nil
	}
	return s.stack[len(s.stack)-1-1]
}

func (s *structIterator) _safegetFieldType() (sf *reflect.StructField) {

	var reprev func(position int) (sf *reflect.StructField)
	reprev = func(position int) (sf *reflect.StructField) {
		if position >= 0 {
			prev := s.stack[position]
			var st reflect.Type
			if prev.ValueValid() == false {
				// try retrieve the field type from previous element in stack (i.e. the
				// parent struct of the current field)
				sf2 := reprev(position - 1)
				if sf2 != nil {
					//log.Printf("prev.index = %v, prev.sv.valid = %v, sf = %v", prev.index, prev.ValueValid(), sf2)
					st = rdecodetypesimple(sf2.Type)
					//log.Printf("sf2.Type/st = %v", st)
					if prev.index < st.NumField() {
						fld := st.Field(prev.index)
						sf = &fld
						//log.Printf("typ: %v, name: %v | %v", typfmt(sf.Type), sf.Name, sf)
					}
				}
			} else {
				st = prev.Type()
				if prev.index < st.NumField() {
					fld := st.Field(prev.index)
					sf = &fld
				}
			}
		}
		return
	}

	sf = reprev(len(s.stack) - 1)
	return nil
}

func (s *structIterator) Next() (accessor *fieldaccessor, ok bool) {
	var lastone *fieldaccessor

	if s._empty() {
		vind := rindirect(s.rootStruct)
		tind := vind.Type()
		lastone = s._push(&vind, tind, 0)

	} else {

	uplevel:
		lastone = s._top().incr()
		if lastone.index >= lastone.NumField() {
			if len(s.stack) <= 1 {
				return // no more fields or children can be iterated
			}
			s._pop()
			goto uplevel
		}

	}

retry:
	field, valvalid := lastone.getStructField(), true
	if field == nil {
		valvalid = false
		field = s._safegetFieldType()
		if field == nil {
			log.Warnf("Next(): cannot get field type, field value == nil, or it's an empty struct")
		} else {
			//log.Debugf("typ: %v, name: %v | %v", typfmt(field.Type), field.Name, field)
		}
	}
	if field != nil {
		tind := field.Type
		if s.autoexpandstruct {
			tind = rindirectType(field.Type)
		}
		k1 := tind.Kind()
		if k1 == reflect.Struct && !s.shouldIgnore(field, tind, k1) {
			if valvalid {
				vind := rindirect(*lastone.FieldValue())
				lastone = s._push(&vind, tind, 0)
				log.Debugf("    -- (retry) -> filed is struct: %v, typ: %v\n", valfmt(&vind), typfmt(tind))
				//field = lastone.StructField()
				//fmt.Printf("    - (retry) %d. %q (%v (%v)) %v\n", field.Index, field.Name, field.Type.Kind(), field.Type, field.Index) //, tind, vind.Interface())
			} else {
				lastone = s._push(nil, tind, 0)
				log.Debugf("    -- (retry) -> filed is struct [value invalid]: %v\n", typfmt(tind))
			}
			goto retry
		}

		//field = lastone.StructField()
		//fmt.Printf("    - %d. %q (%v (%v)) %v\n", field.Index, field.Name, field.Type.Kind(), field.Type, field.Index) //, tind, vind.Interface())
	}

	ok, accessor = true, lastone
	return
}

func (s *structIterator) shouldIgnore(field *reflect.StructField, typ reflect.Type, kind reflect.Kind) bool {
	n := typ.PkgPath()
	return packageisreserved(n) // ignore golang stdlib, such as "io", "runtime", ...
}

var onceinitignoredpackages sync.Once
var _ignoredpackages ignoredpackages
var _ignoredpackageprefixes ignoredpackageprefixes

type ignoredpackages []string
type ignoredpackageprefixes []string

func (a ignoredpackages) contains(packagename string) (yes bool) {
	for _, s := range a {
		if yes = s == packagename; yes {
			break
		}
	}
	return
}
func (a ignoredpackageprefixes) contains(packagename string) (yes bool) {
	for _, s := range a {
		if yes = strings.HasPrefix(packagename, s); yes {
			break
		}
	}
	return
}

func packageisreserved(packagename string) (shouldIgnored bool) {
	onceinitignoredpackages.Do(func() {
		_ignoredpackageprefixes = ignoredpackageprefixes{
			"github.com/golang",
			"golang.org/",
			"google.golang.org/",
		}
		// the following names comes with go1.18beta1 src/.
		// Perhaps it would need to be updated in the future.
		_ignoredpackages = ignoredpackages{
			"archive",
			"bufio",
			"builtin",
			"bytes",
			"cmd",
			"compress",
			"constraints",
			"container",
			"context",
			"crypto",
			"database",
			"debug",
			"embed",
			"encoding",
			"errors",
			"expvar",
			"flag",
			"fmt",
			"go",
			"hash",
			"html",
			"image",
			"index",
			"internal",
			"io",
			"log",
			"math",
			"mime",
			"net",
			"os",
			"path",
			"plugin",
			"reflect",
			"regexp",
			"runtime",
			"sort",
			"strconv",
			"strings",
			"sync",
			"syscall",
			"testdata",
			"testing",
			"text",
			"time",
			"unicode",
			"unsafe",
		}
	})

	shouldIgnored = packagename != "" && (_ignoredpackages.contains(packagename) ||
		_ignoredpackageprefixes.contains(packagename))
	return
}
