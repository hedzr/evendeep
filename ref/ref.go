package ref

import (
	"fmt"
	"math"
	"reflect"
	"unsafe"
)

//
// ref.go - the routines about reflect operations
//

// Rdecode decodes an interface{} or a pointer to something to its
// underlying data type.
//
// Suppose we have an interface{} pointer Value which stored a
// pointer to an integer, Rdecode will extract or retrieve the Value
// of that integer.
//
// See our TestRdecode() in ref_test.go
//
//	var b = 11
//	var i interface{} = &b
//	var v = reflect.ValueOf(&i)
//	var n = Rdecode(v)
//	println(n.Type())    // = int
//
// `prev` returns the previous Value before we arrived at the
// final `ret` Value.
// In another word, the value of `prev` Value is a pointer which
// points to the value of `ret` Value. Or, it's a interface{}
// wrapped about `ret`.
//
// A interface{} will be unboxed to its underlying datatype after
// Rdecode invoked.
func Rdecode(reflectValue reflect.Value) (ret, prev reflect.Value) {
	return Rskip(reflectValue, reflect.Ptr, reflect.Interface)
}

// Rdecodesimple is a shortcut to Rdecode without `prev` returned.
func Rdecodesimple(reflectValue reflect.Value) (ret reflect.Value) {
	ret, _ = Rdecode(reflectValue)
	return
}

func Rskip(reflectValue reflect.Value, kinds ...reflect.Kind) (ret, prev reflect.Value) {
	ret, prev = reflectValue, reflectValue
retry:
	k := ret.Kind()
	for _, kk := range kinds {
		if k == kk {
			prev = ret
			ret = ret.Elem()
			goto retry
		}
	}
	return
}

// Rdecodetype try to strip off ptr and interface{} from a type.
//
// It might not work properly on some cases because interface{} cannot
// be stripped with calling typ.Elem().
//
// In this case, use rdecodesimple(value).Type() instead
// of Rdecodetypesimple(value.Type()).
func Rdecodetype(reflectType reflect.Type) (ret, prev reflect.Type) {
	return Rskiptype(reflectType, reflect.Ptr, reflect.Interface)
}

// Rdecodetypesimple try to strip off ptr and interface{} from a type.
//
// It might not work properly on some cases because interface{} cannot
// be stripped with calling typ.Elem().
// For this case, use Rdecodesimple(value).Type() instead
// of Rdecodetypesimple(value.Type()).
func Rdecodetypesimple(reflectType reflect.Type) (ret reflect.Type) {
	ret, _ = Rdecodetype(reflectType)
	return
}

func Rskiptype(reflectType reflect.Type, kinds ...reflect.Kind) (ret, prev reflect.Type) {
	ret, prev = reflectType, reflectType
retry:
	k := ret.Kind()
	for _, kk := range kinds {
		if k == kk {
			if canElem(k) {
				prev = ret
				ret = ret.Elem()
				goto retry
			}
		}
	}
	return
}

func canElem(k reflect.Kind) bool {
	switch k { //nolint:exhaustive //others unlisted cases can be ignored
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Ptr, reflect.Slice:
		return true
		// case reflect.Interface:
		// 	return true
	}
	return false
}

func Rindirect(reflectValue reflect.Value) reflect.Value {
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	return reflectValue
}

func RindirectType(reflectType reflect.Type) reflect.Type {
	for reflectType.Kind() == reflect.Ptr { // || reflectType.Kind() == reflect.Slice {
		reflectType = reflectType.Elem()
	}
	return reflectType
}

func Rwant(reflectValue reflect.Value, kinds ...reflect.Kind) reflect.Value {
	k := reflectValue.Kind()
retry:
	for _, kk := range kinds {
		if k == kk {
			return reflectValue
		}
	}

	if k == reflect.Interface || k == reflect.Ptr {
		reflectValue = reflectValue.Elem()
		k = reflectValue.Kind()
		goto retry
	}

	return reflectValue
}

// IsMap tests if a given 'v' is a map (or pointer of map, ...)
func IsMap(v any) bool {
	if v == nil {
		return false
	}

	var rv reflect.Value
	if vv, ok := v.(reflect.Value); ok {
		rv = vv
	} else {
		rv = reflect.ValueOf(v)
	}
	tv := Rdecodetypesimple(rv.Type())
	return tv.Kind() == reflect.Map
}

// IsSlice tests if a given 'v' is a slice ([]int, []uint, []string, ...)
func IsSlice(v any) bool {
	if v == nil {
		return false
	}

	var rv reflect.Value
	if vv, ok := v.(reflect.Value); ok {
		rv = vv
	} else {
		rv = reflect.ValueOf(v)
	}
	tv := Rdecodetypesimple(rv.Type())
	return tv.Kind() == reflect.Slice
}

// SliceAppend combines all given slices ('vv') as one result slice.
// That is, SliceAppend([]int{1,2}, []int{2,3}) will return []int{1,2,2,3}.
func SliceAppend(vv ...any) (ret any) {
	var rv reflect.Value
	for i, v := range vv {
		if !IsSlice(v) {
			continue
		}

		var vo reflect.Value
		if vv, ok := v.(reflect.Value); ok {
			vo = vv
		} else {
			vo = reflect.ValueOf(v)
		}

		if i == 0 {
			rv = reflect.MakeSlice(vo.Type(), 0, vo.Len())
		}
		rv = reflect.AppendSlice(rv, vo)
	}
	ret = rv.Interface()
	return
}

// SliceMerge merge all given slices ('vv') as one result slice.
// That is, SliceAppend([]int{1,2}, []int{2,3}) will return []int{1,2,3}.
func SliceMerge(vv ...any) (ret any) {
	var rv reflect.Value
	// var elt reflect.Type
	for i, v := range vv {
		if !IsSlice(v) {
			continue
		}

		var vo reflect.Value
		if vv, ok := v.(reflect.Value); ok {
			vo = vv
		} else {
			vo = reflect.ValueOf(v)
		}

		if i == 0 {
			// elt = vo.Type().Elem()
			rv = reflect.MakeSlice(vo.Type(), 0, vo.Len())
		}
		for k := 0; k < vo.Len(); k++ {
			ve, found := vo.Index(k), false
			for j := 0; j < rv.Len(); j++ {
				vf := rv.Index(j)
				a, b := ve.Interface(), vf.Interface()
				if found = reflect.DeepEqual(a, b); found {
					break
				}
			}
			if !found {
				rv = reflect.Append(rv, ve)
			}
		}
	}
	ret = rv.Interface()
	return
}

// IsNumeric tests if a given 'v' is a number (int, uint, float, ...)
func IsNumeric(v any) bool {
	if v == nil {
		return false
	}

	rv := reflect.ValueOf(v)
	tv := Rdecodetypesimple(rv.Type())
	return IsNumericType(tv)
}

func IsNumericType(t reflect.Type) bool     { return IsNumericKind(t.Kind()) }                           // tests if a given 't' is a number (int, uint, float, ...)
func IsNumIntegerType(t reflect.Type) bool  { return IsNumIntegerKind(t.Kind()) }                        // tests if a given 't' is a integer
func IsNumericKind(k reflect.Kind) bool     { return k >= reflect.Int && k < reflect.Array }             // tests if a given 't' is a number
func IsNumSIntegerKind(k reflect.Kind) bool { return k >= reflect.Int && k <= reflect.Int64 }            // tests if a given 't' is a signed integer
func IsNumUIntegerKind(k reflect.Kind) bool { return k >= reflect.Uint && k <= reflect.Uint64 }          // tests if a given 't' is a unsigned integer
func IsNumIntegerKind(k reflect.Kind) bool  { return k >= reflect.Int && k <= reflect.Uint64 }           // tests if a given 't' is a integer (signed or unsigned)
func IsNumFloatKind(k reflect.Kind) bool    { return k >= reflect.Float32 && k <= reflect.Float64 }      // tests if a given 't' is a float
func IsNumComplexKind(k reflect.Kind) bool  { return k >= reflect.Complex64 && k <= reflect.Complex128 } // tests if a given 't' is a complex number

func KindIs(k reflect.Kind, list ...reflect.Kind) bool {
	for _, l := range list {
		if k == l {
			return true
		}
	}
	return false
}

func Typfmtvlite(v *reflect.Value) string {
	if v == nil || !v.IsValid() {
		return "<invalid>" //nolint:goconst //why need const it?
	}
	t := v.Type()
	return fmt.Sprintf("%v", t) //nolint:gocritic //safe string with fmt lib
}

func Typfmtv(v *reflect.Value) string {
	if v == nil || !v.IsValid() {
		return "<invalid>"
	}
	t := v.Type()
	return fmt.Sprintf("%v (%v)", t, t.Kind())
}

func Typfmt(t reflect.Type) string {
	return fmt.Sprintf("%v (%v)", t, t.Kind())
}

func Typfmtptr(t *reflect.Type) string { //nolint:gocritic //ptrToRefParam: consider `t' to be of non-pointer type
	if t == nil {
		return "???"
	}
	return fmt.Sprintf("%v (%v)", *t, (*t).Kind())
}

// Valfmtptr will step into a ptr value at first, then Valfmt.
func Valfmtptr(v *reflect.Value) string {
	s := ValfmtptrPure(v)
	if len(s) > maxValueStringLen {
		return s[0:maxValueStringLen-3] + "..."
	}
	return s
}

// ValfmtptrPure will step into a ptr value at first, then Valfmt.
func ValfmtptrPure(v *reflect.Value) string {
	if v == nil || !v.IsValid() {
		return "<invalid>"
	}
	if v.Kind() == reflect.Ptr {
		vp := v.Elem()
		return Valfmtptr(&vp)
	}
	return Valfmt(v)
}

const maxValueStringLen = 320

func Valfmtv(v reflect.Value) string {
	return Valfmt(&v)
}

func Valfmt(v *reflect.Value) string {
	s := ValfmtPure(v)
	if len(s) > maxValueStringLen {
		return s[0:maxValueStringLen-3] + "..."
	}
	return s
}

func ValfmtPure(v *reflect.Value) string {
	if v == nil || !v.IsValid() {
		return "<invalid>"
	}
	if v.Kind() == reflect.Bool {
		if v.Bool() {
			return "true"
		}
		return "false"
	}
	if IsNil(*v) {
		return "<nil>"
	}
	if IsZero(*v) {
		return "<zero>"
	}
	if v.Kind() == reflect.String {
		return v.String()
	}
	if HasStringer(v) {
		res := v.MethodByName("String").Call(nil)
		return res[0].String()
	}
	if IsNumericKind(v.Kind()) {
		return fmt.Sprintf("%v", v.Interface())
	}
	if CanConvert(v, StringType) {
		return v.Convert(StringType).String()
	}
	if v.CanInterface() {
		return fmt.Sprintf("%v", v.Interface())
	}
	return fmt.Sprintf("<%v>", v.Kind())
}

func Iserrortype(typ reflect.Type) bool {
	return typ.Implements(errtyp)
}

var errtyp = reflect.TypeOf((*error)(nil)).Elem()

var stringerType = reflect.TypeOf((*interface{ String() string })(nil)).Elem() //nolint:gochecknoglobals //i know that
var StringType = reflect.TypeOf((*string)(nil)).Elem()                         //nolint:gochecknoglobals //i know that
var Niltyp = reflect.TypeOf((*string)(nil))                                    //nolint:gochecknoglobals //i know that

// IsZero for go1.12+, the difference is it never panic on unavailable kinds.
// see also reflect.IsZero.
func IsZero(v reflect.Value) (ret bool) {
	return IsZerov(&v)
}

// IsZerov for go1.12+, the difference is it never panic on unavailable kinds.
// see also reflect.IsZero.
func IsZerov(v *reflect.Value) (ret bool) {
	if v != nil {
		switch k := v.Kind(); k { //nolint:exhaustive //others unlisted cases can be ignored
		case reflect.Bool:
			ret = !v.Bool()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			ret = v.Int() == 0
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			ret = v.Uint() == 0
		case reflect.Float32, reflect.Float64:
			ret = math.Float64bits(v.Float()) == 0
		case reflect.Complex64, reflect.Complex128:
			c := v.Complex()
			ret = math.Float64bits(real(c)) == 0 && math.Float64bits(imag(c)) == 0
		case reflect.Slice:
			ret = v.Len() == 0
		case reflect.Array:
			ret = ArrayIsZerov(v)
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.UnsafePointer:
			ret = IsNilv(v)
		case reflect.Struct:
			ret = StructIsZerov(v)
		case reflect.String:
			ret = v.Len() == 0
		}
	}
	return
}

func StructIsZero(v reflect.Value) bool { return StructIsZerov(&v) }
func StructIsZerov(v *reflect.Value) bool {
	for i := 0; i < v.NumField(); i++ {
		if !IsZero(v.Field(i)) {
			return false
		}
	}
	return true
}

func ArrayIsZero(v reflect.Value) bool { return ArrayIsZerov(&v) }
func ArrayIsZerov(v *reflect.Value) bool {
	for i := 0; i < v.Len(); i++ {
		if !IsZero(v.Index(i)) {
			return false
		}
	}
	return true
}

// IsNil for go1.12+, the difference is it never panic on unavailable kinds.
// see also reflect.IsNil.
func IsNil(v reflect.Value) bool {
	return IsNilv(&v)
}

// IsNilv for go1.12+, the difference is it never panic on unavailable kinds.
// see also reflect.IsNil.
func IsNilv(v *reflect.Value) bool {
	if v != nil {
		switch k := v.Kind(); k { //nolint:exhaustive //no need
		case reflect.Uintptr:
			if v.CanAddr() {
				return v.UnsafeAddr() == 0 // special: reflect.IsNil assumed nil check on an uintptr is illegal, faint!
			}
		case reflect.UnsafePointer:
			return v.Pointer() == 0 // for go1.11, this is a workaround even not bad
		case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr:
			return v.IsNil()
		case reflect.Interface, reflect.Slice:
			return v.IsNil()
			// case reflect.Array:
			//	// never true, for an array, it is never IsNil
			// case reflect.String:
			// case reflect.Struct:
		}
	}
	return false
}

func IsValid(v reflect.Value) bool {
	return v.IsValid()
}

func IsValidv(v *reflect.Value) bool {
	return v != nil && v.IsValid()
}

// func (v Value) IsNil() bool {
//	k := v.kind()
//	switch k {
//	case Chan, Func, Map, Pointer, UnsafePointer:
//		if v.flag&flagMethod != 0 {
//			return false
//		}
//		ptr := v.ptr
//		if v.flag&flagIndir != 0 {
//			ptr = *(*unsafe.Pointer)(ptr)
//		}
//		return ptr == nil
//	case Interface, Slice:
//		// Both interface and slice are nil if first word is 0.
//		// Both are always bigger than a word; assume flagIndir.
//		return *(*unsafe.Pointer)(v.ptr) == nil
//	}
//	panic(&ValueError{"reflect.Value.IsNil", v.kind()})
// }

// IsExported reports whether the field is exported.
func IsExported(f *reflect.StructField) bool {
	return f.PkgPath == ""
}

// CanConvertHelper is a shorthand of CanConvert.
func CanConvertHelper(v reflect.Value, t reflect.Type) bool {
	return CanConvert(&v, t)
}

// CanConvert reports whether the value v can be converted to type t.
// If v.CanConvert(t) returns true then v.Convert(t) will not panic.
func CanConvert(v *reflect.Value, t reflect.Type) bool {
	if !v.IsValid() {
		return false
	}

	vt := v.Type()
	if !vt.ConvertibleTo(t) {
		// Currently the only conversion that is OK in terms of type
		// but that can panic depending on the value is converting
		// from slice to pointer-to-array.
		if vt.Kind() == reflect.Slice && t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Array {
			n := t.Elem().Len()
			type sliceHeader struct {
				Data unsafe.Pointer
				Len  int
				Cap  int
			}
			h := (*sliceHeader)(unsafe.Pointer(v.Pointer()))
			return n <= h.Len
		}

		return false
	}
	return true
}

func hasImplements(v *reflect.Value, interfaceType reflect.Type) bool {
	vt := v.Type()
	return vt.Implements(interfaceType)
}

func HasStringer(v *reflect.Value) bool {
	return hasImplements(v, stringerType)
}
