package tool

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
//    var b = 11
//    var i interface{} = &b
//    var v = reflect.ValueOf(&i)
//    var n = Rdecode(v)
//    println(n.Type())    // = int
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
//
// In this case, use rdecodesimple(value).Type() instead
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

func IsNumericType(t reflect.Type) bool     { return IsNumericKind(t.Kind()) }
func IsNumIntegerType(t reflect.Type) bool  { return IsNumIntegerKind(t.Kind()) }
func IsNumericKind(k reflect.Kind) bool     { return k >= reflect.Int && k < reflect.Array }
func IsNumSIntegerKind(k reflect.Kind) bool { return k >= reflect.Int && k <= reflect.Int64 }
func IsNumUIntegerKind(k reflect.Kind) bool { return k >= reflect.Uint && k <= reflect.Uint64 }
func IsNumIntegerKind(k reflect.Kind) bool  { return k >= reflect.Int && k <= reflect.Uint64 }
func IsNumFloatKind(k reflect.Kind) bool    { return k >= reflect.Float32 && k <= reflect.Float64 }
func IsNumComplexKind(k reflect.Kind) bool  { return k >= reflect.Complex64 && k <= reflect.Complex128 }

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

func Valfmt(v *reflect.Value) string {
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

var stringerType = reflect.TypeOf((*interface{ String() string })(nil)).Elem()
var StringType = reflect.TypeOf((*string)(nil)).Elem()
var Niltyp = reflect.TypeOf((*string)(nil))

// IsZero for go1.12+, the difference is it never panic on unavailable kinds.
// see also reflect.IsZero
func IsZero(v reflect.Value) (ret bool) {
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
	case reflect.Array:
		ret = ArrayIsZero(v)
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		ret = IsNil(v)
	case reflect.Struct:
		ret = StructIsZero(v)
	case reflect.String:
		ret = v.Len() == 0
	}
	return
}

func StructIsZero(v reflect.Value) bool {
	for i := 0; i < v.NumField(); i++ {
		if !IsZero(v.Field(i)) {
			return false
		}
	}
	return true
}

func ArrayIsZero(v reflect.Value) bool {
	for i := 0; i < v.Len(); i++ {
		if !IsZero(v.Index(i)) {
			return false
		}
	}
	return true
}

// IsNil for go1.12+, the difference is it never panic on unavailable kinds.
// see also reflect.IsNil
func IsNil(v reflect.Value) bool {
	switch k := v.Kind(); k { //nolint:exhaustive //no
	case reflect.Uintptr:
		if v.CanAddr() {
			return v.UnsafeAddr() == 0 // special: reflect.IsNil assumed nil check on an uintptr is illegal, faint!
		}
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr:
		return v.IsNil()
	case reflect.UnsafePointer:
		return v.Pointer() == 0 // for go1.11, this is a workaround even not bad
	case reflect.Interface, reflect.Slice:
		return v.IsNil()
		// case reflect.Array:
		//	// never true, for an array, it is never IsNil
		// case reflect.String:
		// case reflect.Struct:
	}
	return false
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

// CanConvertHelper _
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
