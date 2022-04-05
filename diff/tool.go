package diff

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/hedzr/evendeep/typ"
)

func kindis(k reflect.Kind, kinds ...reflect.Kind) (yes bool) {
	for _, kk := range kinds {
		if yes = k == kk; yes {
			break
		}
	}
	return
}

type dottedPath struct {
	parts []pathPart
}

func (dp dottedPath) appendAndNew(parts ...pathPart) dottedPath {
	return dottedPath{parts: append(dp.parts, parts...)}
}

func (dp dottedPath) String() string {
	var sb strings.Builder
	for _, p := range dp.parts {
		if sb.Len() > 0 {
			sb.WriteRune('.')
		}
		sb.WriteString(p.String())
	}
	return sb.String()
}

type pathPart interface {
	String() string
}

type visit struct {
	al, ar unsafe.Pointer
	typ    reflect.Type
}

type Update struct {
	Old, New typ.Any // string
	Typ      string
}

func (n Update) String() string {
	if n.Old == nil {
		return fmt.Sprintf("%#v", n.New)
	}
	if n.New == nil {
		return fmt.Sprintf("%#v -> nil", n.Old)
	}
	return fmt.Sprintf("%#v -> %#v", n.Old, n.New)
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

func typfmtlite(v *reflect.Value) string {
	// v := reflect.ValueOf(val)

	if v == nil || !v.IsValid() {
		return "<invalid>"
	}
	t := v.Type()
	return fmt.Sprintf("%v", t)
}

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
