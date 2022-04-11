package evendeep

import (
	"github.com/hedzr/evendeep/flags/cms"
	"github.com/hedzr/evendeep/internal/cl"
	"github.com/hedzr/evendeep/internal/dbglog"
	"github.com/hedzr/evendeep/internal/tool"
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
func (rec tablerec) FieldName() string {
	// return strings.Join(reverseStringSlice(rec.names), ".")
	var sb strings.Builder
	for i := len(rec.names) - 1; i >= 0; i-- {
		if sb.Len() > 0 {
			sb.WriteRune('.')
		}
		sb.WriteString(rec.names[i])
	}
	return sb.String()
}
func (rec tablerec) ShortFieldName() string {
	if len(rec.names) > 0 {
		return rec.names[0]
	}
	return ""
}

func (table *fieldstable) shouldIgnore(field *reflect.StructField, typ reflect.Type, kind reflect.Kind) bool {
	n := typ.PkgPath()
	return packageisreserved(n) // ignore golang stdlib, such as "io", "runtime", ...
}

func (table *fieldstable) getallfields(structValue reflect.Value, autoexpandstruct bool) fieldstable {
	table.autoExpandStruct = autoexpandstruct

	structValue, _ = tool.Rdecode(structValue)
	if structValue.Kind() != reflect.Struct {
		return *table
	}

	styp := structValue.Type()
	ret := table.getfields(&structValue, styp, "", -1)
	table.tablerecords = append(table.tablerecords, ret...)
	// for _, ni := range ret.records {
	//	table.records = append(table.records, ni)
	// }
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
	st := tool.Rdecodetypesimple(structType)
	if st.Kind() != reflect.Struct {
		return
	}

	var i, amount int
	for i, amount = 0, st.NumField(); i < amount; i++ {
		var tr tablerec

		sf := structType.Field(i)
		sftyp := sf.Type
		sftypind := tool.RindirectType(sftyp)
		svind := table.safegetstructfieldvalueind(structValue, i)

		dbglog.Log(" field %d: %v (%v) (%v)", i, sf.Name, tool.Typfmt(sftyp), tool.Typfmt(sftypind))

		isStruct := sftypind.Kind() == reflect.Struct
		shouldIgnored := table.shouldIgnore(&sf, sftypind, sftypind.Kind())

		if isStruct && table.autoExpandStruct && !shouldIgnored {
			n := table.getfields(svind, sftypind, sf.Name, i)
			if len(n) > 0 {
				ret = append(ret, n...)
			} else {
				// add empty struct
				tr = table.tablerec(svind, &sf, sf.Index[0], 0, "")
				ret = append(ret, tr)
				// ret = append(ret, tablerec{
				//	names:            []string{sf.Name},
				//	indexes:          sf.Index,
				//	structFieldValue: svind,
				//	structField:      &sf,
				// })
			}
		} else {
			tr = table.tablerec(svind, &sf, i, fi, fieldname)
			ret = append(ret, tr)
		}
	}
	return //nolint:nakedret
}

func (table *fieldstable) tablerec(svind *reflect.Value, sf *reflect.StructField, index, parentIndex int, parentFieldName string) (tr tablerec) {
	tr.structField = sf
	if tool.IsExported(sf) {
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
	Next() (accessor accessor, ok bool)

	// SourceFieldShouldBeIgnored _
	SourceFieldShouldBeIgnored(ignoredNames []string) (yes bool)
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

type accessor interface {
	Set(v reflect.Value)

	IsStruct() bool

	Type() reflect.Type
	ValueValid() bool
	FieldValue() *reflect.Value
	FieldType() *reflect.Type
	StructField() *reflect.StructField
	StructFieldName() string
	NumField() int

	SourceField() tablerec
	// SourceFieldShouldBeIgnored(ignoredNames []string) bool

	IsFlagOK(f cms.CopyMergeStrategy) bool
	IsGroupedFlagOK(f ...cms.CopyMergeStrategy) bool
	IsAnyFlagsOK(f ...cms.CopyMergeStrategy) bool
	IsAllFlagsOK(f ...cms.CopyMergeStrategy) bool
}

type fieldaccessor struct {
	structvalue *reflect.Value
	structtype  reflect.Type
	index       int
	structfield *reflect.StructField
	isstruct    bool

	sourceTableRec tablerec             // a copy of structIterator.sourcefields
	srcStructField *reflect.StructField // source field type
	fieldTags      *fieldTags           // tag of source field
}

func (s *fieldaccessor) Set(v reflect.Value) {
	if s.ValueValid() {
		if s.isstruct {
			dbglog.Log("    setting struct.%q", s.structtype.Field(s.index).Name)
			sv := tool.Rindirect(*s.structvalue).Field(s.index)
			dbglog.Log("      set %v (%v) -> struct.%q", tool.Valfmt(&v), tool.Typfmtv(&v), s.structtype.Field(s.index).Name)
			sv.Set(v)
		} else if s.structtype.Kind() == reflect.Map {
			key := s.mapkey()
			dbglog.Log("    set %v (%v) -> map[%v]", tool.Valfmt(&v), tool.Typfmtv(&v), tool.Valfmt(&key))
			s.structvalue.SetMapIndex(key, v)
		}
	}
}
func (s *fieldaccessor) SourceField() tablerec { return s.sourceTableRec }
func (s *fieldaccessor) IsFlagOK(f cms.CopyMergeStrategy) bool {
	if s.fieldTags != nil {
		return s.fieldTags.flags.IsFlagOK(f)
	}
	return false
}
func (s *fieldaccessor) IsGroupedFlagOK(f ...cms.CopyMergeStrategy) bool {
	if s.fieldTags != nil {
		return s.fieldTags.flags.IsGroupedFlagOK(f...)
	}
	return false
}
func (s *fieldaccessor) IsAnyFlagsOK(f ...cms.CopyMergeStrategy) bool {
	if s.fieldTags != nil {
		return s.fieldTags.flags.IsAnyFlagsOK(f...)
	}
	return false
}
func (s *fieldaccessor) IsAllFlagsOK(f ...cms.CopyMergeStrategy) bool {
	if s.fieldTags != nil {
		return s.fieldTags.flags.IsAllFlagsOK(f...)
	}
	return false
}
func (s *fieldaccessor) IsStruct() bool {
	return s.isstruct // s.structtype != nil && s.structtype.Kind() == reflect.Struct
}
func (s *fieldaccessor) Type() reflect.Type { return s.structtype }
func (s *fieldaccessor) ValueValid() bool   { return s.structvalue != nil && s.structvalue.IsValid() }
func (s *fieldaccessor) FieldValue() *reflect.Value {
	if s != nil {
		if s.isstruct {
			if s.ValueValid() {
				vind := tool.Rindirect(*s.structvalue)
				if vind.IsValid() && s.index < vind.NumField() {
					r := vind.Field(s.index)
					return &r
				}
			}
		} else if s.structtype.Kind() == reflect.Map {
			key := s.mapkey()
			val := s.structvalue.MapIndex(key)
			return &val
		}
	}
	return nil
}
func (s *fieldaccessor) FieldType() *reflect.Type { //nolint:gocritic //ptrToRefParam: consider to make non-pointer type for `*reflect.Type`
	if s != nil {
		if s.isstruct {
			sf := s.StructField()
			if sf != nil {
				return &sf.Type
			}
		} else if s.structtype.Kind() == reflect.Map {
			// name := s.sourceTableRec.FieldName()
			vt := s.structtype.Elem()
			return &vt
		}
	}
	return nil
}
func (s *fieldaccessor) NumField() int {
	if s.isstruct {
		sf := s.structtype
		return sf.NumField()
		// if s.ValueValid() {
		//	return s.structvalue.NumField()
		// }
	}
	return 0
}
func (s *fieldaccessor) StructField() *reflect.StructField {
	// if s.ValueValid() {
	//	r := s.structvalue.Type().Field(s.index)
	//	s.structfield = &r
	// }
	// return s.structfield
	return s.getStructField()
}
func (s *fieldaccessor) getStructField() *reflect.StructField {
	if s.isstruct {
		if s.structfield == nil && s.index < s.structtype.NumField() {
			r := s.structtype.Field(s.index)
			s.structfield = &r
		}
		// if s.ValueValid() {
		//	r := s.structvalue.Type().Field(s.index)
		//	s.structfield = &r
		// }
		return s.structfield
	}
	return nil
}
func (s *fieldaccessor) StructFieldName() string {
	if fld := s.StructField(); fld != nil {
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
	if s.isstruct && s.index < s.structtype.NumField() {
		if s.structvalue == nil {
			return // cannot do anything except return
		}
		// if s.structvalue.IsValid() {
		//	return
		// }
		sf := s.structtype.Field(s.index)
		vind := tool.Rindirect(*s.structvalue)
		fv := vind.Field(s.index)

		switch kind := sf.Type.Kind(); kind { //nolint:exhaustive
		case reflect.Ptr:
			if tool.IsNil(fv) {
				dbglog.Log("   autoNew")
				typ := sf.Type.Elem()
				nv := reflect.New(typ)
				fv.Set(nv)
			}
		default:
		}
	}
}
func (s *fieldaccessor) mapkey() reflect.Value {
	name := s.sourceTableRec.ShortFieldName()
	kt := s.structtype.Key() // , s.structtype.Elem()
	var key reflect.Value
	if kk := kt.Kind(); kk == reflect.String || kk == reflect.Interface {
		key = reflect.ValueOf(name)
	} else {
		namev := reflect.ValueOf(name)
		kp := reflect.New(kt)
		fsc := &fromStringConverter{}
		if err := fsc.CopyTo(nil, namev, kp.Elem()); err == nil {
			key = kp.Elem()
		}
	}
	return key
}

//

//

func (s *structIterator) iipush(structvalue *reflect.Value, structtype reflect.Type, index int) *fieldaccessor {
	s.stack = append(s.stack, &fieldaccessor{isstruct: true, structvalue: structvalue, structtype: structtype, index: index})
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

// func (s *structIterator) iiprev() *fieldaccessor {
//	if len(s.stack) <= 1 {
//		return nil
//	}
//	return s.stack[len(s.stack)-1-1]
// }

// func (s *structIterator) iiSafegetFieldType() (sf *reflect.StructField) {
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
// }
//
// func (s *structIterator) iiCheckNilPtr(lastone *fieldaccessor, field *reflect.StructField) {
//	lastone.ensurePtrField()
// }

// sourceStructFieldsTable _
type sourceStructFieldsTable interface {
	TableRecords() tablerecords
	CurrRecord() tablerec
	TableRecord(index int) tablerec
	Step(delta int)
}

func (s *structIterator) TableRecords() tablerecords     { return s.srcFields.tablerecords }
func (s *structIterator) CurrRecord() tablerec           { return s.srcFields.tablerecords[s.srcIndex] }
func (s *structIterator) TableRecord(index int) tablerec { return s.srcFields.tablerecords[index] }
func (s *structIterator) Step(delta int)                 { s.withSourceIteratorIndexIncrease(delta) }

func (s *structIterator) SourceFieldShouldBeIgnored(ignoredNames []string) (yes bool) {
	shortName := s.srcFields.tablerecords[s.srcIndex].ShortFieldName()
	for _, x := range ignoredNames {
		if yes = isWildMatch(shortName, x); yes {
			break
		}
	}
	return
}

func (s *structIterator) withSourceIteratorIndexIncrease(srcIndexDelta int) (sourcefield tablerec, ok bool) {
	if s.srcIndex < 0 {
		s.srcIndex = 0
	}

	// if i < params.srcType.NumField() {
	//	t := params.srcType.Field(i)
	//	params.fieldType = &t
	//	params.fieldTags = parseFieldTags(t.Tag)
	// }

	if s.srcIndex < len(s.srcFields.tablerecords) {
		sourcefield, ok = s.srcFields.tablerecords[s.srcIndex], true
	}

	s.srcIndex += srcIndexDelta
	if s.srcIndex < 0 {
		s.srcIndex = 0
	}

	return
}

func (s *structIterator) Next() (acc accessor, ok bool) {
	var sourceTableRec tablerec
	var accessorTmp *fieldaccessor
	sourceTableRec, ok = s.withSourceIteratorIndexIncrease(+1)
	if ok {
		srcStructField := sourceTableRec.StructField()
		isfn := srcStructField.Type.Kind() == reflect.Func

		kind := s.dstStruct.Kind()
		if kind == reflect.Map {
			eltyp := s.dstStruct.Type().Elem()
			tk := eltyp.Kind()
			if srcStructField.Type.ConvertibleTo(eltyp) ||
				srcStructField.Type.AssignableTo(eltyp) ||
				tk == reflect.String || tk == reflect.Interface {
				acc = &fieldaccessor{
					structvalue:    &s.dstStruct,
					structtype:     s.dstStruct.Type(),
					index:          s.dstIndex,
					structfield:    nil,
					isstruct:       false,
					sourceTableRec: sourceTableRec,
					srcStructField: srcStructField,
					fieldTags:      nil,
				}
				s.dstIndex++
			}
			return
		}

		accessorTmp, ok = s.doNext(isfn && !s.noExpandIfSrcFieldIsFunc)
		if ok {
			accessorTmp.sourceTableRec = sourceTableRec
			accessorTmp.srcStructField = srcStructField
			accessorTmp.fieldTags = parseFieldTags(accessorTmp.srcStructField.Tag, "")

			dbglog.Log("   | Next %d | src field: %v (%v) -> %v (%v) | autoexpd: (%v, %v)",
				s.srcIndex, accessorTmp.sourceTableRec.FieldName(),
				tool.Typfmt(accessorTmp.srcStructField.Type),
				accessorTmp.StructFieldName(), tool.Typfmt(accessorTmp.Type()),
				s.srcFields.autoExpandStruct, s.autoExpandStruct,
			)
			s.dstIndex++
		}
	} else {
		accessorTmp, ok = s.doNext(false)
		if ok {
			dbglog.Log("   | Next %d | -> %v (%v)",
				s.dstIndex,
				accessorTmp.StructFieldName(), tool.Typfmt(accessorTmp.Type()))
			s.dstIndex++
		}
	}
	acc = accessorTmp
	return //nolint:nakedret
}

func (s *structIterator) doNext(srcFieldIsFuncAndTargetShouldNotExpand bool) (accessor *fieldaccessor, ok bool) {
	var lastone *fieldaccessor
	var inretry bool

	if s.iiempty() {
		vind := tool.Rindirect(s.dstStruct)
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
		// tind := field.Type // rindirectType(field.Type)
		if s.autoExpandStruct {
			tind := tool.RindirectType(field.Type)
			k1 := tind.Kind()
			dbglog.Log("typ: %v, name: %v | %v", tool.Typfmt(tind), field.Name, field)
			if s.autoNew {
				lastone.ensurePtrField()
			}
			if k1 == reflect.Struct &&
				!srcFieldIsFuncAndTargetShouldNotExpand &&
				!s.shouldIgnore(field, tind, k1) {
				fvp := lastone.FieldValue()
				lastone = s.iipush(fvp, tind, 0)
				dbglog.Log("    -- (retry) -> filed is struct, typ: %v\n", tool.Typfmt(tind))
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
	return //nolint:nakedret
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
	return //nolint:nakedret
}
