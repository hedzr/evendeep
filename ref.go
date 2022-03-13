package deepcopy

import (
	"fmt"
	"math"
	"reflect"
	"unsafe"
)

//
// ref.go - the routines about reflect operations
//

// rdecode decodes an interface{} or a pointer to something to its
// underlying data type.
//
// Suppose we have an interface{} pointer Value which stored a
// pointer to an integer, rdecode will extract or retrieve the Value
// of that integer.
//
// See our TestRdecode() in ref_test.go
//
//    var b = 11
//    var i interface{} = &b
//    var v = reflect.ValueOf(&i)
//    var n = rdecode(v)
//    println(n.Type())    // = int
//
// `prev` returns the previous Value before we arrived at the
// final `ret` Value.
// In another word, the value of `prev` Value is a pointer which
// points to the value of `ret` Value. Or, it's a interface{}
// wrapped about `ret`.
//
// A interface{} will be unboxed to its underlying datatype after
// rdecode invoked.
func rdecode(reflectValue reflect.Value) (ret, prev reflect.Value) {
	return rskip(reflectValue, reflect.Ptr, reflect.Interface)
}
func rdecodesimple(reflectValue reflect.Value) (ret reflect.Value) {
	ret, _ = rdecode(reflectValue)
	return
}

func rskip(reflectValue reflect.Value, kinds ...reflect.Kind) (ret, prev reflect.Value) {
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

// rdecodetype try to strip off ptr and interface{} from a type.
//
// It might not work properly on some cases because interface{} cannot
// be stripped with calling typ.Elem().
//
// In this case, use rdecodesimple(value).Type() instead
// of rdecodetypesimple(value.Type()).
func rdecodetype(reflectType reflect.Type) (ret, prev reflect.Type) {
	return rskiptype(reflectType, reflect.Ptr, reflect.Interface)
}

// rdecodetypesimple try to strip off ptr and interface{} from a type.
//
// It might not work properly on some cases because interface{} cannot
// be stripped with calling typ.Elem().
//
// In this case, use rdecodesimple(value).Type() instead
// of rdecodetypesimple(value.Type()).
func rdecodetypesimple(reflectType reflect.Type) (ret reflect.Type) {
	ret, _ = rdecodetype(reflectType)
	return
}

func rskiptype(reflectType reflect.Type, kinds ...reflect.Kind) (ret, prev reflect.Type) {
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
	switch k {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Ptr, reflect.Slice:
		return true
	}
	return false
}

func rindirect(reflectValue reflect.Value) reflect.Value {
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	return reflectValue
}

func rindirectType(reflectType reflect.Type) reflect.Type {
	for reflectType.Kind() == reflect.Ptr { // || reflectType.Kind() == reflect.Slice {
		reflectType = reflectType.Elem()
	}
	return reflectType
}

func rwant(reflectValue reflect.Value, kinds ...reflect.Kind) reflect.Value {
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

func isNumericType(t reflect.Type) bool     { return isNumericKind(t.Kind()) }
func isNumIntegerType(t reflect.Type) bool  { return isNumIntegerKind(t.Kind()) }
func isNumericKind(k reflect.Kind) bool     { return k >= reflect.Int && k < reflect.Array }
func isNumSIntegerKind(k reflect.Kind) bool { return k >= reflect.Int && k <= reflect.Int64 }
func isNumUIntegerKind(k reflect.Kind) bool { return k >= reflect.Uint && k <= reflect.Uint64 }
func isNumIntegerKind(k reflect.Kind) bool  { return k >= reflect.Int && k <= reflect.Uint64 }
func isNumFloatKind(k reflect.Kind) bool    { return k >= reflect.Float32 && k <= reflect.Float64 }
func isNumComplexKind(k reflect.Kind) bool  { return k >= reflect.Complex64 && k <= reflect.Complex128 }

func typfmtv(v *reflect.Value) string {
	if v == nil || !v.IsValid() {
		return "<invalid>"
	}
	t := v.Type()
	return fmt.Sprintf("%v (%v)", t, t.Kind())
}

func typfmt(t reflect.Type) string {
	return fmt.Sprintf("%v (%v)", t, t.Kind())
}

func valfmt(v *reflect.Value) string {
	if v == nil || !v.IsValid() {
		return "<invalid>"
	}
	if isNil(*v) {
		return "<nil>"
	}
	if isZero(*v) {
		return "<zero>"
	}
	if v.Kind() == reflect.String {
		return v.String()
	}
	if hasStringer(v) {
		res := v.MethodByName("String").Call(nil)
		return res[0].String()
	}
	if isNumericKind(v.Kind()) {
		return fmt.Sprintf("%v", v.Interface())
	}
	if canConvert(v, stringType) {
		return v.Convert(stringType).String()
	}
	if v.CanInterface() {
		return fmt.Sprintf("%v", v.Interface())
	}
	return fmt.Sprintf("<%v>", v.Kind())
}

var stringerType = reflect.TypeOf((*interface{ String() string })(nil)).Elem()
var stringType = reflect.TypeOf((*string)(nil)).Elem()

// isZero for go1.12+, the difference is it never panic on unavailable kinds.
// see also reflect.IsZero
func isZero(v reflect.Value) (ret bool) {
	k := v.Kind()
	switch k {
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
		ret = arrayIsZero(v)
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		ret = isNil(v)
	case reflect.Struct:
		ret = structIsZero(v)
	case reflect.String:
		ret = v.Len() == 0
	}
	return
}

func structIsZero(v reflect.Value) bool {
	for i := 0; i < v.NumField(); i++ {
		if !isZero(v.Field(i)) {
			return false
		}
	}
	return true
}

func arrayIsZero(v reflect.Value) bool {
	for i := 0; i < v.Len(); i++ {
		if !isZero(v.Index(i)) {
			return false
		}
	}
	return true
}

// isNil for go1.12+, the difference is it never panic on unavailable kinds.
// see also reflect.IsNil
func isNil(v reflect.Value) bool {
	k := v.Kind()
	switch k {
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
		//case reflect.Array:
		//	// never true, for an array, it is never IsNil
		//case reflect.String:
		//case reflect.Struct:
	}
	return false
}

//func (v Value) IsNil() bool {
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
//}

// isExported reports whether the field is exported.
func isExported(f *reflect.StructField) bool {
	return f.PkgPath == ""
}

func canConvertHelper(v reflect.Value, t reflect.Type) bool {
	return canConvert(&v, t)
}

// canConvert reports whether the value v can be converted to type t.
// If v.CanConvert(t) returns true then v.Convert(t) will not panic.
func canConvert(v *reflect.Value, t reflect.Type) bool {
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

func hasStringer(v *reflect.Value) bool {
	return hasImplements(v, stringerType)
}
