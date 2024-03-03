package diff

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/hedzr/evendeep/dbglog"
	"github.com/hedzr/evendeep/ref"
)

type visit struct {
	al, ar unsafe.Pointer
	typ    reflect.Type
}

type sliceIndex int

func (n sliceIndex) String() string {
	return fmt.Sprintf("[%d]", n)
}

type mapKey struct {
	Key interface{}
}

func (n mapKey) String() string {
	return fmt.Sprintf("[%#v]", n.Key)
}

type structField string

func (n structField) String() string {
	return fmt.Sprintf(".%s", string(n))
}

//

// isEmptyObject detects the emptiness of a slice or a map
func isEmptyObject(v reflect.Value) (yes bool) {
	if kind := v.Kind(); !ref.KindIs(kind, reflect.Invalid, reflect.Slice, reflect.Map, reflect.Array) {
		return
	}
	if yes = ref.IsZero(v); yes {
		return
	}
	yes = v.Len() == 0
	return
}

func isEmptyStruct(v reflect.Value) (yes bool) {
	if kind := v.Kind(); kind != reflect.Struct {
		return
	}
	yes = ref.IsZero(v)
	return
}

func isEmptyStructDeeply(v reflect.Value) (yes bool) {
	if kind := v.Kind(); kind != reflect.Struct {
		return
	}
	ve := reflect.New(v.Type()).Elem()
	inf := newInfo(
		WithTreatEmptyStructPtrAsNilPtr(true),
		WithStripPointerAtFirst(true),
		WithCompareDifferentTypeStructs(true),
		WithCompareDifferentSizeArrays(false),
		WithIgnoreUnmatchedFields(false),
	)
	dbglog.Log(" isEmptyStructDeeply(v): %+v", ref.Valfmt(&v))
	dbglog.Log("          the empty obj: %+v", ref.Valfmt(&ve))
	yes = inf.diff(v, ve)
	return
}

//

//

// func kindis(k reflect.Kind, kinds ...reflect.Kind) (yes bool) {
// 	for _, kk := range kinds {
// 		if yes = k == kk; yes {
// 			break
// 		}
// 	}
// 	return
// }

// func typfmtlite(v *reflect.Value) string {
// 	// v := reflect.ValueOf(val)
//
// 	if v == nil || !v.IsValid() {
// 		return "<invalid>"
// 	}
// 	t := v.Type()
// 	return fmt.Sprintf("%v", t)
// }
//
// func valfmtlite(val typ.Any) string {
// 	v := reflect.ValueOf(val)
// 	if !v.IsValid() {
// 		return "<invalid>"
// 	}
// 	if v.Kind() == reflect.Bool {
// 		if v.Bool() {
// 			return "true"
// 		}
// 		return "false"
// 	}
// 	if tool.IsNil(v) {
// 		return "<nil>"
// 	}
// 	if tool.IsZero(v) {
// 		return "<zero>"
// 	}
// 	if v.Kind() == reflect.String {
// 		return v.String()
// 	}
// 	if tool.HasStringer(&v) {
// 		res := v.MethodByName("String").Call(nil)
// 		return res[0].String()
// 	}
// 	if tool.IsNumericKind(v.Kind()) {
// 		return fmt.Sprintf("%v", v.Interface())
// 	}
// 	if tool.CanConvert(&v, tool.StringType) {
// 		return v.Convert(tool.StringType).String()
// 	}
// 	if v.CanInterface() {
// 		return fmt.Sprintf("%v", v.Interface())
// 	}
// 	return fmt.Sprintf("<%v>", v.Kind())
// }
