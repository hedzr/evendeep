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

// fieldsTableT is an accessor to struct fields.
type fieldsTableT struct {
	tableRecordsT
	autoExpandStruct bool
	typ              reflect.Type  // struct type
	val              reflect.Value // struct value
	fastIndices      map[string]*tableRecT
}

type tableRecordsT []*tableRecT

type tableRecT struct {
	names            []string // the path from root struct, in reverse order
	indexes          []int
	structFieldValue *reflect.Value
	structField      *reflect.StructField
}

func (rec tableRecT) FieldValue() *reflect.Value        { return rec.structFieldValue }
func (rec tableRecT) StructField() *reflect.StructField { return rec.structField }
func (rec tableRecT) FieldName() string {
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
func (rec tableRecT) ShortFieldName() string {
	if len(rec.names) > 0 {
		return rec.names[0]
	}
	return ""
}
func (rec tableRecT) ShouldIgnore() bool {
	if rec.structField != nil {
		ft := parseFieldTags(rec.structField.Tag, "")
		return ft.flags.IsGroupedFlagOK(cms.Ignore)
	}
	return false
}

func (table *fieldsTableT) shouldIgnore(field *reflect.StructField, typ reflect.Type, kind reflect.Kind) bool {
	n := typ.PkgPath()
	return packageisreserved(n) // ignore golang stdlib, such as "io", "runtime", ...
}

func (table *fieldsTableT) getAllFields(structValue reflect.Value, autoexpandstruct bool) fieldsTableT {
	table.autoExpandStruct = autoexpandstruct

	structValue, _ = tool.Rdecode(structValue)
	if structValue.Kind() != reflect.Struct {
		return *table
	}

	styp := structValue.Type()
	ret := table.getFields(&structValue, styp, "", -1)
	table.tableRecordsT = append(table.tableRecordsT, ret...)

	table.typ = styp
	table.val = structValue

	table.fastIndices = make(map[string]*tableRecT)
	for _, ni := range table.tableRecordsT {
		log.VDebugf("        - ni: %v", ni.ShortFieldName())
		table.fastIndices[ni.ShortFieldName()] = ni
	}

	return *table
}

func (table *fieldsTableT) safeGetStructFieldValueInd(structValue *reflect.Value, i int) *reflect.Value {
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

func (table *fieldsTableT) getFields(structValue *reflect.Value, structType reflect.Type, fieldname string, fi int) (ret tableRecordsT) {
	st := tool.Rdecodetypesimple(structType)
	if st.Kind() != reflect.Struct {
		return
	}

	var i, amount int
	for i, amount = 0, st.NumField(); i < amount; i++ {
		var tr *tableRecT

		sf := structType.Field(i)
		sftyp := sf.Type
		sftypind := tool.RindirectType(sftyp)
		svind := table.safeGetStructFieldValueInd(structValue, i)

		dbglog.Log(" field %d: %v (%v) (%v)", i, sf.Name, tool.Typfmt(sftyp), tool.Typfmt(sftypind))

		isStruct := sftypind.Kind() == reflect.Struct
		shouldIgnored := table.shouldIgnore(&sf, sftypind, sftypind.Kind())

		if isStruct && table.autoExpandStruct && !shouldIgnored {
			n := table.getFields(svind, sftypind, sf.Name, i)
			if len(n) > 0 {
				ret = append(ret, n...)
			} else {
				// add empty struct
				tr = table.tableRec(svind, &sf, sf.Index[0], 0, "")
				ret = append(ret, tr)
			}
		} else {
			tr = table.tableRec(svind, &sf, i, fi, fieldname)
			ret = append(ret, tr)
		}
	}
	return //nolint:nakedret
}

func (table *fieldsTableT) tableRec(svind *reflect.Value, sf *reflect.StructField, index, parentIndex int, parentFieldName string) (tr *tableRecT) {
	tr = new(tableRecT)

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
	Next(params *Params, byName bool) (accessor accessor, ok bool)

	// SourceFieldShouldBeIgnored _
	SourceFieldShouldBeIgnored(ignoredNames []string) (yes bool)
	// ShouldBeIgnored _
	ShouldBeIgnored(name string, ignoredNames []string) (yes bool)
}

type structIterableOpt func(s *structIteratorT)

//

// newStructIterator return a deep recursive iterator for the given
// struct value.
//
// The structIterable.Next() will enumerate all children fields.
func newStructIterator(structValue reflect.Value, opts ...structIterableOpt) structIterable {
	s := &structIteratorT{
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
	return func(s *structIteratorT) {
		s.autoExpandStruct = expand
	}
}

// withStructFieldPtrAutoNew allows auto-expanding the struct or its pointer
// in iterating a parent struct
func withStructFieldPtrAutoNew(create bool) structIterableOpt {
	return func(s *structIteratorT) {
		s.autoNew = create
	}
}

// withStructSource _
func withStructSource(srcstructval *reflect.Value, autoexpand bool) structIterableOpt {
	return func(s *structIteratorT) {
		if srcstructval != nil {
			s.srcFields = s.srcFields.getAllFields(*srcstructval, autoexpand)
			s.withSourceIteratorIndexIncrease(-10000) // reset srcIndex to 0
		}
	}
}

//

type structIteratorT struct {
	srcFields                fieldsTableT      // source struct fields accessors
	srcIndex                 int               // source field index
	dstStruct                reflect.Value     // target struct
	dstIndex                 int               // counter for Next()
	stack                    []*fieldAccessorT // target fields accessors
	autoExpandStruct         bool              // Next() will expand *struct to struct and get inside loop deeply
	noExpandIfSrcFieldIsFunc bool              //
	autoNew                  bool              // create new inner objects for the child ptr,map,chan,..., if necessary
}

// accessor represents a struct field accessor which can be used for getting or setting.
// Before you can operate it, do verify on ValueValid()
type accessor interface {
	Set(v reflect.Value)

	IsStruct() bool

	Type() reflect.Type
	ValueValid() bool
	FieldValue() *reflect.Value
	FieldType() *reflect.Type
	StructField() *reflect.StructField // return target struct field type
	StructFieldName() string           // return target struct field name
	NumField() int

	SourceField() *tableRecT // return source struct (or others types)
	// SourceFieldShouldBeIgnored(ignoredNames []string) bool

	IsFlagOK(f cms.CopyMergeStrategy) bool
	IsGroupedFlagOK(f ...cms.CopyMergeStrategy) bool
	IsAnyFlagsOK(f ...cms.CopyMergeStrategy) bool
	IsAllFlagsOK(f ...cms.CopyMergeStrategy) bool
}

type fieldAccessorT struct {
	structValue *reflect.Value
	structType  reflect.Type
	index       int
	structField *reflect.StructField
	isStruct    bool

	sourceTableRec *tableRecT           // a copy of structIteratorT.sourcefields
	srcStructField *reflect.StructField // source field type
	fieldTags      *fieldTags           // tag of source field
}

func setToZero(fieldValue *reflect.Value) {
	setToZeroAs(fieldValue, fieldValue.Type(), fieldValue.Kind())
}

func setToZeroAs(fieldValue *reflect.Value, typ reflect.Type, kind reflect.Kind) {
	if fieldValue.CanSet() {
		switch kind { //nolint:exhaustive
		case reflect.Bool:
			fieldValue.SetBool(false)
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
			fieldValue.SetInt(0)
		case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uintptr:
			fieldValue.SetUint(0)
		case reflect.Float64, reflect.Float32:
			fieldValue.SetFloat(0)
		case reflect.Complex128, reflect.Complex64:
			fieldValue.SetComplex(0 + 0i)
		case reflect.String:
			fieldValue.SetString("")
		case reflect.Slice:
			v := reflect.New(reflect.SliceOf(fieldValue.Type().Elem())).Elem()
			fieldValue.Set(v)
		case reflect.Array:
			for i, amount := 0, fieldValue.Len(); i < amount; i++ {
				v := fieldValue.Index(i)
				vt, vk := v.Type(), v.Kind()
				setToZeroAs(&v, vt, vk)
			}
		case reflect.Map:
			v := reflect.MakeMap(fieldValue.Type()) // .Elem()
			fieldValue.Set(v)
		case reflect.Ptr:
			fieldValue.Set(reflect.Zero(typ))
			// fieldValue.SetPointer(unsafe.Pointer(uintptr(0)))
		case reflect.Interface:
			ind := fieldValue.Elem().Type()
			fieldValue.Set(reflect.Zero(ind))

		case reflect.Struct:
			setToZeroForStruct(fieldValue, typ, kind)
		default:
			// Array, Chan, Func, Interface, Invalid, Struct, UnsafePointer
			dbglog.Log("   NOT SUPPORTED (kind: %v), cannot set to zero value.", kind)
		}
	}
}

func setToZeroForStruct(structValue *reflect.Value, typ reflect.Type, kind reflect.Kind) {
	for i, amount := 0, structValue.NumField(); i < amount; i++ {
		fv := structValue.Field(i)
		setToZero(&fv)
	}
}

func (s *fieldAccessorT) Set(v reflect.Value) {
	if s.ValueValid() {
		if s.isStruct {
			// dbglog.Log("    target struct type: %v", tool.Typfmt(s.structType))
			dbglog.Log("    setting struct.%q", s.structType.Field(s.index).Name)
			sv := tool.Rindirect(*s.structValue).Field(s.index)
			dbglog.Log("      set %v (%v) -> struct.%q", tool.Valfmt(&v), tool.Typfmtv(&v), s.structType.Field(s.index).Name)
			if v.IsValid() && !tool.IsZero(v) {
				dbglog.Log("      set to v : %v", tool.Valfmt(&v))
				sv.Set(v)
			} else {
				dbglog.Log("      setToZero")
				setToZero(&sv)
			}
		} else if s.structType.Kind() == reflect.Map {
			key := s.mapkey()
			dbglog.Log("    set %v (%v) -> map[%v]", tool.Valfmt(&v), tool.Typfmtv(&v), tool.Valfmt(&key))
			s.structValue.SetMapIndex(key, v)
		}
	}
}
func (s *fieldAccessorT) SourceField() *tableRecT { return s.sourceTableRec }
func (s *fieldAccessorT) IsFlagOK(f cms.CopyMergeStrategy) bool {
	if s.fieldTags != nil {
		return s.fieldTags.flags.IsFlagOK(f)
	}
	return false
}
func (s *fieldAccessorT) IsGroupedFlagOK(f ...cms.CopyMergeStrategy) bool {
	if s.fieldTags != nil {
		return s.fieldTags.flags.IsGroupedFlagOK(f...)
	}
	return false
}
func (s *fieldAccessorT) IsAnyFlagsOK(f ...cms.CopyMergeStrategy) bool {
	if s.fieldTags != nil {
		return s.fieldTags.flags.IsAnyFlagsOK(f...)
	}
	return false
}
func (s *fieldAccessorT) IsAllFlagsOK(f ...cms.CopyMergeStrategy) bool {
	if s.fieldTags != nil {
		return s.fieldTags.flags.IsAllFlagsOK(f...)
	}
	return false
}
func (s *fieldAccessorT) IsStruct() bool {
	return s.isStruct // s.structType != nil && s.structType.Kind() == reflect.Struct
}
func (s *fieldAccessorT) Type() reflect.Type { return s.structType }
func (s *fieldAccessorT) ValueValid() bool   { return s.structValue != nil && s.structValue.IsValid() }
func (s *fieldAccessorT) FieldValue() *reflect.Value {
	if s != nil {
		if s.isStruct {
			if s.ValueValid() {
				vind := tool.Rindirect(*s.structValue)
				if vind.IsValid() && s.index < vind.NumField() {
					r := vind.Field(s.index)
					return &r
				}
			}
		} else { // if s.structType != nil && s.structType.Kind() == reflect.Map {
			key := s.mapkey()
			val := s.structValue.MapIndex(key)
			return &val
		}
	}
	return nil
}
func (s *fieldAccessorT) FieldType() *reflect.Type { //nolint:gocritic //ptrToRefParam: consider to make non-pointer type for `*reflect.Type`
	if s != nil {
		if s.isStruct {
			sf := s.StructField()
			if sf != nil {
				return &sf.Type
			}
		} else if s.structType.Kind() == reflect.Map {
			// name := s.sourceTableRec.FieldName()
			vt := s.structType.Elem()
			return &vt
		}
	}
	return nil
}
func (s *fieldAccessorT) NumField() int {
	if s.isStruct {
		sf := s.structType
		return sf.NumField()
		// if s.ValueValid() {
		//	return s.structValue.NumField()
		// }
	}
	return 0
}
func (s *fieldAccessorT) StructField() *reflect.StructField {
	// if s.ValueValid() {
	//	r := s.structValue.Type().Field(s.index)
	//	s.structField = &r
	// }
	// return s.structField
	return s.getStructField()
}
func (s *fieldAccessorT) getStructField() *reflect.StructField {
	if s.isStruct {
		if s.structField == nil && s.index >= 0 && s.index < s.structType.NumField() {
			r := s.structType.Field(s.index)
			s.structField = &r
		}
		// if s.ValueValid() {
		//	r := s.structValue.Type().Field(s.index)
		//	s.structField = &r
		// }
		return s.structField
	}
	return nil
}
func (s *fieldAccessorT) StructFieldName() string {
	if s.isStruct {
		if fld := s.StructField(); fld != nil {
			return fld.Name
		}
	} else if keys := s.structValue.MapKeys(); s.index < len(keys) {
		ck := keys[s.index]
		if target, err := rToString(ck, ck.Type()); err == nil {
			return target.String()
		}
	}
	return ""
}
func (s *fieldAccessorT) incr() *fieldAccessorT {
	s.index++
	s.structField = nil
	return s
}
func (s *fieldAccessorT) ensurePtrField() {
	if s.isStruct && s.index < s.structType.NumField() {
		if s.structValue != nil {
			sf := s.structType.Field(s.index)
			vind := tool.Rindirect(*s.structValue)
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
}
func (s *fieldAccessorT) mapkey() reflect.Value {
	if s.sourceTableRec == nil {
		var ck reflect.Value
		if keys := s.structValue.MapKeys(); s.index < len(keys) {
			ck = keys[s.index]
		}
		return ck
	}

	name := s.sourceTableRec.ShortFieldName()
	kt := s.structType.Key() // , s.structType.Elem()
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

func (s *structIteratorT) iipush(structvalue *reflect.Value, structtype reflect.Type, index int) *fieldAccessorT {
	s.stack = append(s.stack, &fieldAccessorT{isStruct: true, structValue: structvalue, structType: structtype, index: index})
	return s.iitop()
}
func (s *structIteratorT) iiempty() bool { return len(s.stack) == 0 }
func (s *structIteratorT) iipop() {
	if len(s.stack) > 0 {
		s.stack = s.stack[0 : len(s.stack)-1]
	}
}
func (s *structIteratorT) iitop() *fieldAccessorT {
	if len(s.stack) == 0 {
		return nil
	}
	return s.stack[len(s.stack)-1]
}

// func (s *structIteratorT) iiprev() *fieldAccessorT {
//	if len(s.stack) <= 1 {
//		return nil
//	}
//	return s.stack[len(s.stack)-1-1]
// }

// func (s *structIteratorT) iiSafegetFieldType() (sf *reflect.StructField) {
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
// func (s *structIteratorT) iiCheckNilPtr(lastone *fieldAccessorT, field *reflect.StructField) {
//	lastone.ensurePtrField()
// }

// sourceStructFieldsTable _
type sourceStructFieldsTable interface {
	TableRecords() tableRecordsT
	CurrRecord() *tableRecT
	TableRecord(index int) *tableRecT
	Step(delta int)
	RecordByName(name string) *reflect.Value
	MethodCallByName(name string) (mtd reflect.Method, v *reflect.Value)
}

func (s *structIteratorT) TableRecords() tableRecordsT      { return s.srcFields.tableRecordsT }
func (s *structIteratorT) CurrRecord() *tableRecT           { return s.srcFields.tableRecordsT[s.srcIndex] }
func (s *structIteratorT) TableRecord(index int) *tableRecT { return s.srcFields.tableRecordsT[index] }
func (s *structIteratorT) Step(delta int)                   { s.withSourceIteratorIndexIncrease(delta) }

func (s *structIteratorT) RecordByName(name string) (v *reflect.Value) {
	if tr, ok := s.srcFields.fastIndices[name]; ok {
		v = tr.FieldValue()
	}
	return
}

func (s *structIteratorT) MethodCallByName(name string) (mtd reflect.Method, v *reflect.Value) {
	var exists bool
	if mtd, exists = s.srcFields.typ.MethodByName(name); exists {
		mtdv := s.srcFields.val.MethodByName(name)
		retv := mtdv.Call([]reflect.Value{})
		if len(retv) > 0 {
			if len(retv) > 1 {
				errv := retv[len(retv)-1]
				if tool.IsNil(errv) && tool.Iserrortype(mtd.Type.Out(len(retv)-1)) {
					v = &retv[0]
				}
			} else {
				v = &retv[0]
			}
		}
	}
	return
}

//

func (s *structIteratorT) withSourceIteratorIndexIncrease(srcIndexDelta int) (sourcefield *tableRecT, ok bool) {
	if s.srcIndex < 0 {
		s.srcIndex = 0
	}

	// if i < params.srcType.NumField() {
	//	t := params.srcType.Field(i)
	//	params.fieldType = &t
	//	params.fieldTags = parseFieldTags(t.Tag)
	// }

	if s.srcIndex < len(s.srcFields.tableRecordsT) {
		sourcefield, ok = s.srcFields.tableRecordsT[s.srcIndex], true
	}

	s.srcIndex += srcIndexDelta
	if s.srcIndex < 0 {
		s.srcIndex = 0
	}

	return
}

func (s *structIteratorT) Next(params *Params, byName bool) (acc accessor, ok bool) {
	var sourceTableRec *tableRecT
	var accessorTmp *fieldAccessorT
	sourceTableRec, ok = s.withSourceIteratorIndexIncrease(+1)
	if ok {
		srcStructField := sourceTableRec.StructField()
		srcIsFunc := srcStructField.Type.Kind() == reflect.Func
		kind := s.dstStruct.Kind()

		if kind == reflect.Map {
			return s.doNextMapItem(params, sourceTableRec, srcStructField)
		}

		if params != nil && byName {
			return s.doNextFieldByName(sourceTableRec, srcStructField)
		}

		accessorTmp, ok = s.doNext(srcIsFunc && !s.noExpandIfSrcFieldIsFunc)
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
		kind := s.dstStruct.Kind()

		if kind == reflect.Map {
			return s.doNextMapItem(params, sourceTableRec, nil)
		}

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

func (s *structIteratorT) doNextMapItem(params *Params, sourceTableRec *tableRecT, srcStructField *reflect.StructField) (acc accessor, ok bool) {
	checkSourceTypeIsSettable := func(srcStructField *reflect.StructField) (tk reflect.Kind, ok bool) {
		if srcStructField == nil {
			return reflect.Invalid, true
		}
		st := srcStructField.Type
		elTyp := s.dstStruct.Type().Elem()
		tk = elTyp.Kind()
		ok = st.ConvertibleTo(elTyp) || st.AssignableTo(elTyp) ||
			tk == reflect.String || tk == reflect.Interface
		return
	}

	if _, ok = checkSourceTypeIsSettable(srcStructField); ok {
		acc = &fieldAccessorT{
			structValue:    &s.dstStruct,
			structType:     s.dstStruct.Type(),
			index:          s.dstIndex,
			structField:    nil,
			isStruct:       false,
			sourceTableRec: sourceTableRec,
			srcStructField: srcStructField,
			fieldTags:      nil,
		}
		ok = srcStructField != nil || s.dstIndex < len(s.dstStruct.MapKeys())
		s.dstIndex++
	}
	return
}

func (s *structIteratorT) doNextFieldByName(sourceTableRec *tableRecT, srcStructField *reflect.StructField) (acc accessor, ok bool) {
	dbglog.Log("     looking for src field: %v", srcStructField)
	dstFieldName, ignored := s.getTargetFieldNameBySourceField(srcStructField, "")

	ok = true
	if ignored {
		return // return ok = true so that caller can continue to the next field, rather than stop looping.
	}

	dbglog.Log("     looking for field %q (src field: %q)", dstFieldName, srcStructField.Name)
	var tsf reflect.StructField
	ts := tool.Rindirect(s.dstStruct)
	tsf, ok = ts.Type().FieldByName(dstFieldName)
	if ok {
		dbglog.Log("     tsf: %v", tsf)
		// tv := ts.FieldByName(dstFieldName)
		s.dstIndex = tsf.Index[0]
		acc = &fieldAccessorT{
			structValue:    &ts,
			structType:     ts.Type(),
			index:          s.dstIndex,
			structField:    &tsf,
			isStruct:       true,
			sourceTableRec: sourceTableRec,
			srcStructField: srcStructField,
			fieldTags:      parseFieldTags(srcStructField.Tag, ""),
		}
	} else {
		log.Warnf("     [WARN] dstFieldName %q NOT FOUND, it'll be ignored", dstFieldName)
		s.dstIndex = -1
		acc = &fieldAccessorT{
			structValue:    &ts,
			structType:     ts.Type(),
			index:          s.dstIndex, // return an empty accessor
			structField:    nil,        // return an empty accessor
			isStruct:       true,
			sourceTableRec: sourceTableRec,
			srcStructField: srcStructField,
			fieldTags:      parseFieldTags(srcStructField.Tag, ""),
		}
		ok = true // return ok = true so caller can keep going to next field
	}
	// no need: dstIndex++
	return // return ok = false, the caller loop will be broken.
}

func (s *structIteratorT) doNext(srcFieldIsFuncAndTargetShouldNotExpand bool) (accessor *fieldAccessorT, ok bool) {
	var lastone *fieldAccessorT
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
				!s.typShouldBeIgnored(tind) {
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

// func (s *structIteratorT) getTargetFieldName(knownSrcName, tagKeyName string) (dstFieldName string, ignored bool) {
// 	dstFieldName = knownSrcName
//
// 	// var flagsInTag *fieldTags
// 	// var ok bool
// 	if sf := s.CurrRecord().StructField(); sf != nil {
// 		dstFieldName, ignored = s.getTargetFieldNameBySourceField(sf, tagKeyName)
// 		// 	flagsInTag = parseFieldTags(sf.Tag, tagKeyName)
// 		// 	if ignored = flagsInTag.isFlagExists(cms.Ignore); ignored {
// 		// 		return
// 		// 	}
// 		// 	ctx := &NameConverterContext{Params: nil}
// 		// 	dstFieldName, ok = flagsInTag.CalcTargetName(sf.Name, ctx)
// 		// 	if !ok {
// 		// 		if tr := s.dstStruct.FieldByName(knownSrcName); !tool.IsNil(tr) {
// 		// 			dstFieldName = knownSrcName
// 		// 			dbglog.Log("     dstName: %v, ok: %v [pre 2, fld: %v, tag: %v]", dstFieldName, ok, sf.Name, sf.Tag)
// 		// 		}
// 		// 	}
// 	}
// 	return
// }

func (s *structIteratorT) getTargetFieldNameBySourceField(knownSrcField *reflect.StructField, tagName string) (dstFieldName string, ignored bool) {
	var flagsInTag *fieldTags
	var ok bool

	flagsInTag = parseFieldTags(knownSrcField.Tag, tagName)
	if ignored = flagsInTag.isFlagExists(cms.Ignore); ignored {
		return
	}
	ctx := &NameConverterContext{Params: nil}
	dstFieldName, ok = flagsInTag.CalcTargetName(knownSrcField.Name, ctx)
	if !ok {
		if tr := s.dstStruct.FieldByName(knownSrcField.Name); !tool.IsNil(tr) {
			dstFieldName = knownSrcField.Name
			dbglog.Log("     dstName: %v, ok: %v [pre 2, fld: %v, tag: %v]", dstFieldName, ok, knownSrcField.Name, knownSrcField.Tag)
		}
	}
	return
}

func (s *structIteratorT) typShouldBeIgnored(typ reflect.Type) bool {
	n := typ.PkgPath()
	return packageisreserved(n) // ignore golang stdlib, such as "io", "runtime", ...
}

func (s *structIteratorT) SourceFieldShouldBeIgnored(ignoredNames []string) (yes bool) {
	rec := s.srcFields.tableRecordsT[s.srcIndex]
	if yes = rec.ShouldIgnore(); yes {
		return
	}
	shortName := rec.ShortFieldName()
	return s.ShouldBeIgnored(shortName, ignoredNames)
}

func (s *structIteratorT) ShouldBeIgnored(name string, ignoredNames []string) (yes bool) {
	for _, x := range ignoredNames {
		if yes = isWildMatch(name, x); yes {
			break
		}
	}
	return
}

var onceinitignoredpackages sync.Once
var _ignoredpackages ignoredpackages
var _ignoredpackageprefixes ignoredpackageprefixes

type ignoredpackages map[string]bool
type ignoredpackageprefixes []string

func (a ignoredpackages) contains(packageName string) (yes bool) {
	// for _, s := range a {
	// 	if yes = s == packageName; yes {
	// 		break
	// 	}
	// }
	_, yes = a[packageName]
	return
}
func (a ignoredpackageprefixes) contains(packageName string) (yes bool) {
	for _, s := range a {
		if yes = strings.HasPrefix(packageName, s); yes {
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
		// the following name list comes with go1.18beta1 src/.
		// Perhaps it would need to be updated in the future.
		_ignoredpackages = ignoredpackages{
			"archive":     true,
			"bufio":       true,
			"builtin":     true,
			"bytes":       true,
			"cmd":         true,
			"compress":    true,
			"constraints": true,
			"container":   true,
			"context":     true,
			"crypto":      true,
			"database":    true,
			"debug":       true,
			"embed":       true,
			"encoding":    true,
			"errors":      true,
			"expvar":      true,
			"flag":        true,
			"fmt":         true,
			"go":          true,
			"hash":        true,
			"html":        true,
			"image":       true,
			"index":       true,
			"internal":    true,
			"io":          true,
			"log":         true,
			"math":        true,
			"mime":        true,
			"net":         true,
			"os":          true,
			"path":        true,
			"plugin":      true,
			"reflect":     true,
			"regexp":      true,
			"runtime":     true,
			"sort":        true,
			"strconv":     true,
			"strings":     true,
			"sync":        true,
			"syscall":     true,
			"testdata":    true,
			"testing":     true,
			"text":        true,
			"time":        true,
			"unicode":     true,
			"unsafe":      true,
		}
	})

	shouldIgnored = packagename != "" && (_ignoredpackages.contains(packagename) ||
		_ignoredpackageprefixes.contains(packagename))
	return //nolint:nakedret
}
