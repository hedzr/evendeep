package deepcopy

import (
	"github.com/hedzr/deepcopy/cl"
	"github.com/hedzr/deepcopy/dbglog"
	"github.com/hedzr/log"
	"reflect"
	"strings"
	"sync"
)

//

// fieldstable is an accessor to struct fields.
type fieldstable struct {
	tablerecords
	autoExpandStruct bool
}

type tablerecords []tablerec

type tablerec struct {
	names            []string // the path from root struct, in reverse order
	indexes          []int
	structFieldValue *reflect.Value
	structField      *reflect.StructField
}

func (rec tablerec) FieldValue() *reflect.Value        { return rec.structFieldValue }
func (rec tablerec) StructField() *reflect.StructField { return rec.structField }
func (rec tablerec) FieldName() string                 { return strings.Join(reverseStringSlice(rec.names), ".") }
func (rec tablerec) ShortFieldName() string {
	if len(rec.names) > 0 {
		return rec.names[0]
	}
	return ""
}

func (table *fieldstable) shouldIgnore(field reflect.StructField, typ reflect.Type, kind reflect.Kind) bool {
	n := typ.PkgPath()
	return packageisreserved(n) // ignore golang stdlib, such as "io", "runtime", ...
}

func (table *fieldstable) getallfields(structValue reflect.Value, autoexpandstruct bool) fieldstable {
	table.autoExpandStruct = autoexpandstruct

	structValue, _ = rdecode(structValue)
	if structValue.Kind() != reflect.Struct {
		return *table
	}

	styp := structValue.Type()
	ret := table.getfields(&structValue, styp, "", -1)
	table.tablerecords = append(table.tablerecords, ret...)
	//for _, ni := range ret.records {
	//	table.records = append(table.records, ni)
	//}
	return *table
}

func (table *fieldstable) safegetstructfieldvalueind(structValue *reflect.Value, i int) *reflect.Value {
	if structValue != nil && structValue.IsValid() {
		for structValue.Kind() == reflect.Ptr {
			v := structValue.Elem()
			structValue = &v
		}
		if structValue != nil && structValue.IsValid() {
			sv := structValue.Field(i)
			return &sv
		}
	}
	return nil
}

func (table *fieldstable) getfields(structValue *reflect.Value, structType reflect.Type, fieldname string, fi int) (ret tablerecords) {
	st := rdecodetypesimple(structType)
	if st.Kind() != reflect.Struct {
		return
	}

	var i, amount int
	for i, amount = 0, st.NumField(); i < amount; i++ {
		var tr tablerec

		sf := structType.Field(i)
		sftyp := sf.Type
		sftypind := rindirectType(sftyp)
		svind := table.safegetstructfieldvalueind(structValue, i)

		dbglog.Log(" field %d: %v (%v) (%v)", i, sf.Name, typfmt(sftyp), typfmt(sftypind))

		isStruct := sftypind.Kind() == reflect.Struct
		shouldIgnored := table.shouldIgnore(sf, sftypind, sftypind.Kind())

		if isStruct && table.autoExpandStruct && !shouldIgnored {
			n := table.getfields(svind, sftypind, sf.Name, i)
			if len(n) > 0 {
				ret = append(ret, n...)
			} else {
				// add empty struct
				tr = table.tablerec(svind, &sf, sf.Index[0], 0, "")
				ret = append(ret, tr)
				//ret = append(ret, tablerec{
				//	names:            []string{sf.Name},
				//	indexes:          sf.Index,
				//	structFieldValue: svind,
				//	structField:      &sf,
				//})
			}
		} else {
			tr = table.tablerec(svind, &sf, i, fi, fieldname)
			ret = append(ret, tr)
		}
	}
	return
}

func (table *fieldstable) tablerec(svind *reflect.Value, sf *reflect.StructField, index, parentIndex int, parentFieldName string) (tr tablerec) {
	tr.structField = sf
	if isExported(sf) {
		tr.structFieldValue = svind
	} else if svind.CanAddr() {
		val := cl.GetUnexportedField(*svind)
		tr.structFieldValue = &val
	}
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

//

//

// structIterable provides a struct fields iterable interface
type structIterable interface {
	// Next returns the next field as an accessor.
	//
	// Next iterates all fields by ordinal, and enter any
	// inner struct for the children, expect empty struct.
	//
	// For an empty struct, Next return it exactly rather than
	// field since it has nothing to iterate.
	Next() (accessor *fieldaccessor, ok bool)
}

type structIterableOpt func(s *structIterator)

//

// newStructIterator return a deep recursive iterator for the given
// struct value.
//
// The structIterable.Next() will enumerate all children fields.
func newStructIterator(structValue reflect.Value, opts ...structIterableOpt) structIterable {
	s := &structIterator{
		dstStruct: structValue,
		stack:     nil,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// withStructPtrAutoExpand allows auto-expanding the struct or its pointer
// in iterating a parent struct
func withStructPtrAutoExpand(expand bool) structIterableOpt {
	return func(s *structIterator) {
		s.autoExpandStruct = expand
	}
}

// withStructFieldPtrAutoNew allows auto-expanding the struct or its pointer
// in iterating a parent struct
func withStructFieldPtrAutoNew(create bool) structIterableOpt {
	return func(s *structIterator) {
		s.autoNew = create
	}
}

// withStructSource _
func withStructSource(srcstructval *reflect.Value, autoexpand bool) structIterableOpt {
	return func(s *structIterator) {
		if srcstructval != nil {
			s.srcFields = s.srcFields.getallfields(*srcstructval, autoexpand)
			s.withSourceIteratorIndexIncrease(-10000) // reset srcIndex to 0
		}
	}
}

//

type structIterator struct {
	srcFields                fieldstable      // source struct fields accessors
	srcIndex                 int              // source field index
	dstStruct                reflect.Value    // target struct
	dstIndex                 int              // counter for Next()
	stack                    []*fieldaccessor // target fields accessors
	autoExpandStruct         bool             // Next() will expand *struct to struct and get inside loop deeply
	noExpandIfSrcFieldIsFunc bool             //
	autoNew                  bool             // create new inner objects for the child ptr,map,chan,..., if necessary
}

type fieldaccessor struct {
	structvalue *reflect.Value
	structtype  reflect.Type
	index       int
	structfield *reflect.StructField

	sourceTableRec tablerec             // a copy of structIterator.sourcefields
	srcStructField *reflect.StructField // source field type
	fieldTags      *fieldTags           // tag of source field
}

func (s *fieldaccessor) Set(v reflect.Value) {
	if s.ValueValid() {
		sv := s.structvalue.Field(s.index)
		sv.Set(v)
	}
}

func (s *fieldaccessor) Type() reflect.Type { return s.structtype }
func (s *fieldaccessor) ValueValid() bool   { return s.structvalue != nil && s.structvalue.IsValid() }
func (s *fieldaccessor) FieldValue() *reflect.Value {
	if s.ValueValid() {
		vind := rindirect(*s.structvalue)
		if vind.IsValid() {
			r := vind.Field(s.index)
			return &r
		}
	}
	return nil
}
func (s *fieldaccessor) FieldType() *reflect.Type {
	sf := s.StructField()
	if sf != nil {
		return &sf.Type
	}
	return nil
}
func (s *fieldaccessor) NumField() int {
	sf := s.structtype
	return sf.NumField()
	//if s.ValueValid() {
	//	return s.structvalue.NumField()
	//}
	//return 0
}
func (s *fieldaccessor) StructField() *reflect.StructField {
	//if s.ValueValid() {
	//	r := s.structvalue.Type().Field(s.index)
	//	s.structfield = &r
	//}
	//return s.structfield
	return s.getStructField()
}
func (s *fieldaccessor) getStructField() *reflect.StructField {
	if s.structfield == nil && s.index < s.structtype.NumField() {
		r := s.structtype.Field(s.index)
		s.structfield = &r
	}
	//if s.ValueValid() {
	//	r := s.structvalue.Type().Field(s.index)
	//	s.structfield = &r
	//}
	return s.structfield
}
func (s *fieldaccessor) StructFieldName() string {
	fld := s.StructField()
	if fld != nil {
		return fld.Name
	}
	return ""
}
func (s *fieldaccessor) incr() *fieldaccessor {
	s.index++
	s.structfield = nil
	return s
}
func (s *fieldaccessor) ensurePtrField() {
	if s.index < s.structtype.NumField() {
		if s.structvalue == nil {
			return // cannot do anything except return
		}
		//if s.structvalue.IsValid() {
		//	return
		//}
		sf := s.structtype.Field(s.index)
		vind := rindirect(*s.structvalue)
		fv := vind.Field(s.index)
		kind := sf.Type.Kind()
		switch kind {
		case reflect.Ptr:
			if isNil(fv) {
				dbglog.Log("   autoNew")
				typ := sf.Type.Elem()
				nv := reflect.New(typ)
				fv.Set(nv)
			}
		}
	}
}

//

//

func (s *structIterator) iipush(structvalue *reflect.Value, structtype reflect.Type, index int) *fieldaccessor {
	s.stack = append(s.stack, &fieldaccessor{structvalue: structvalue, structtype: structtype, index: index})
	return s.iitop()
}
func (s *structIterator) iiempty() bool { return len(s.stack) == 0 }
func (s *structIterator) iipop() {
	if len(s.stack) > 0 {
		s.stack = s.stack[0 : len(s.stack)-1]
	}
}
func (s *structIterator) iitop() *fieldaccessor {
	if len(s.stack) == 0 {
		return nil
	}
	return s.stack[len(s.stack)-1]
}

//func (s *structIterator) iiprev() *fieldaccessor {
//	if len(s.stack) <= 1 {
//		return nil
//	}
//	return s.stack[len(s.stack)-1-1]
//}

//func (s *structIterator) iiSafegetFieldType() (sf *reflect.StructField) {
//
//	var reprev func(position int) (sf *reflect.StructField)
//	reprev = func(position int) (sf *reflect.StructField) {
//		if position >= 0 {
//			prev := s.stack[position]
//			var st reflect.Type
//			if prev.ValueValid() == false {
//				// try retrieve the field type from previous element in stack (i.e. the
//				// parent struct of the current field)
//				sf2 := reprev(position - 1)
//				if sf2 != nil {
//					//log.Printf("prev.index = %v, prev.sv.valid = %v, sf = %v", prev.index, prev.ValueValid(), sf2)
//					st = rdecodetypesimple(sf2.Type)
//					//log.Printf("sf2.Type/st = %v", st)
//					if prev.index < st.NumField() {
//						fld := st.Field(prev.index)
//						sf = &fld
//						//log.Printf("typ: %v, name: %v | %v", typfmt(sf.Type), sf.Name, sf)
//					}
//				}
//			} else {
//				st = prev.Type()
//				if prev.index < st.NumField() {
//					fld := st.Field(prev.index)
//					sf = &fld
//				}
//			}
//		}
//		return
//	}
//
//	sf = reprev(len(s.stack) - 1)
//	return nil
//}
//
//func (s *structIterator) iiCheckNilPtr(lastone *fieldaccessor, field *reflect.StructField) {
//	lastone.ensurePtrField()
//}

// sourceStructFieldsTable _
type sourceStructFieldsTable interface {
	gettablerecords() tablerecords
	getcurrrecord() tablerec
	gettablerec(index int) tablerec
	step(delta int)
}

func (s *structIterator) gettablerecords() tablerecords  { return s.srcFields.tablerecords }
func (s *structIterator) getcurrrecord() tablerec        { return s.srcFields.tablerecords[s.srcIndex] }
func (s *structIterator) gettablerec(index int) tablerec { return s.srcFields.tablerecords[index] }
func (s *structIterator) step(delta int)                 { s.withSourceIteratorIndexIncrease(delta) }

func (s *structIterator) withSourceIteratorIndexIncrease(srcIndexDelta int) (sourcefield tablerec, ok bool) {
	if s.srcIndex < 0 {
		s.srcIndex = 0
	}

	//if i < params.srcType.NumField() {
	//	t := params.srcType.Field(i)
	//	params.fieldType = &t
	//	params.fieldTags = parseFieldTags(t.Tag)
	//}

	if s.srcIndex < len(s.srcFields.tablerecords) {
		sourcefield, ok = s.srcFields.tablerecords[s.srcIndex], true
	}

	s.srcIndex += srcIndexDelta
	if s.srcIndex < 0 {
		s.srcIndex = 0
	}

	return
}

func (s *structIterator) Next() (accessor *fieldaccessor, ok bool) {
	var sourceTableRec tablerec
	sourceTableRec, ok = s.withSourceIteratorIndexIncrease(+1)
	if ok {
		srcStructField := sourceTableRec.StructField()
		isfn := srcStructField.Type.Kind() == reflect.Func

		accessor, ok = s.doNext(isfn && !s.noExpandIfSrcFieldIsFunc)
		if ok {
			accessor.sourceTableRec = sourceTableRec
			accessor.srcStructField = srcStructField
			accessor.fieldTags = parseFieldTags(accessor.srcStructField.Tag)

			dbglog.Log("   | Next %d | src field: %v (%v) -> %v (%v)",
				s.srcIndex, accessor.sourceTableRec.FieldName(),
				typfmt(accessor.srcStructField.Type),
				accessor.StructFieldName(), typfmt(accessor.Type()))
			s.dstIndex++
		}
	} else {
		accessor, ok = s.doNext(false)
		if ok {
			dbglog.Log("   | Next %d | -> %v (%v)",
				s.dstIndex,
				accessor.StructFieldName(), typfmt(accessor.Type()))
			s.dstIndex++
		}
	}
	return
}

func (s *structIterator) doNext(srcFieldIsFuncAndTargetShouldNotExpand bool) (accessor *fieldaccessor, ok bool) {
	var lastone *fieldaccessor
	var inretry bool

	if s.iiempty() {
		vind := rindirect(s.dstStruct)
		tind := vind.Type()
		lastone = s.iipush(&vind, tind, 0)

	} else {

	uplevel:
		lastone = s.iitop().incr()
		if lastone.index >= lastone.NumField() {
			if len(s.stack) <= 1 {
				return // no more fields or children can be iterated
			}
			s.iipop()
			goto uplevel
		}

	}

retryExpand:
	field := lastone.getStructField()
	if field != nil {
		tind := field.Type // rindirectType(field.Type)
		if s.autoExpandStruct {
			tind = rindirectType(field.Type)
			k1 := tind.Kind()
			dbglog.Log("typ: %v, name: %v | %v", typfmt(tind), field.Name, field)
			if s.autoNew {
				lastone.ensurePtrField()
			}
			if k1 == reflect.Struct &&
				!srcFieldIsFuncAndTargetShouldNotExpand &&
				!s.shouldIgnore(field, tind, k1) {

				fvp := lastone.FieldValue()
				lastone = s.iipush(fvp, tind, 0)
				dbglog.Log("    -- (retry) -> filed is struct, typ: %v\n", typfmt(tind))
				inretry = true
				goto retryExpand

			}
		} else if s.autoNew {
			lastone.ensurePtrField()
		}

	} else {
		if inretry && lastone.NumField() == 0 {
			// for an empty struct, go back and up to parent level and
			// iterate it instead of iterating its fields since there's
			// no longer fields.
			//
			// NOTE that should be cared to prevent endless loop at this
			// point.
			s.iipop()
			lastone = s.iitop()
		} else {
			log.Warnf("cannot fetching field, empty struct ? ")
		}
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
