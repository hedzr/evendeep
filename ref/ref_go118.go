package ref

import (
	"reflect"
)

// IsNilT for go1.18+
func IsNilT[T any](id T) (ret bool) {
	v := reflect.ValueOf(id)
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
	return
}
