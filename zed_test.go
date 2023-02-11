package evendeep_test

import (
	"bytes"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/hedzr/evendeep"
	"github.com/hedzr/evendeep/internal/tool"
)

func TestBytesBuffer(t *testing.T) {
	var v bytes.Buffer
	vv := reflect.ValueOf(v)
	t.Logf("%v.%v", vv.Type().PkgPath(), vv.Type().Name())
}

// canConvert reports whether the value v can be converted to type t.
// If v.CanConvert(t) returns true then v.Convert(t) will not panic.
func canConvert(v *reflect.Value, t reflect.Type) bool {
	vt := v.Type()
	if !vt.ConvertibleTo(t) {
		return false
	}
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
		if n > h.Len {
			return false
		}
	}
	return true
}

func TestTimeStruct(t *testing.T) {

	timeZone, _ := time.LoadLocation("America/Phoenix")
	tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
	tm2 := time.Date(2003, 9, 1, 23, 59, 59, 3579, timeZone)

	src := &tm
	dst := &tm2

	vs := reflect.ValueOf(src)
	vd := reflect.ValueOf(dst)

	t.Logf("%v -> %v", vs.Type(), vd.Type())

	if canConvert(&vs, vd.Type()) {
		if vd.Elem().CanSet() {
			vd.Elem().Set(vs.Elem())
			return
		} else {
			t.Logf("vd.CanSet == false")
		}
	} else {
		t.Logf("vs.CanConvert == false")
	}
}

func TestUintptr(t *testing.T) {

	x0 := evendeep.X0{}
	up := unsafe.Pointer(&x0)

	vv := reflect.ValueOf(&up)
	t.Logf("%v (%v) | %v", vv.Type(), vv.Type().Kind(), vv.Interface())
	vv = vv.Elem()
	t.Logf("%v (%v) | %v", vv.Type(), vv.Type().Kind(), vv.Interface())

	var a *int
	v1 := reflect.ValueOf(&a)
	t.Logf("%v (%v) | %v", v1.Type(), v1.Type().Kind(), v1.Interface())
	v1 = v1.Elem()
	t.Logf("%v (%v) | %v", v1.Type(), v1.Type().Kind(), v1.Interface())

	defer func() {
		if e := recover(); e != nil {
			t.Logf("[recover] %v", e)
		}
	}()
	v1 = v1.Elem() // should report panic: reflect: call of reflect.Value.Type on zero Value

	// these following codes would never be reached.
	t.Logf("%v (%v) | %v", v1.Type(), v1.Type().Kind(), v1.Interface())
	v1 = v1.Elem()
	t.Logf("%v (%v) | %v", v1.Type(), v1.Type().Kind(), v1.Interface())
}

func TestInspectStruct(t *testing.T) {
	em := new(evendeep.Employee)
	tool.InspectStruct(em)
}

func TestDeepCopyFromOutside(t *testing.T) {
	// defer dbglog.newCaptureLog(t).Release()

	nn := []int{2, 9, 77, 111, 23, 29}
	var a [2]string
	a[0] = "Hello"
	a[1] = "World"

	x0 := evendeep.X0{}
	x1 := evendeep.X1{
		A: uintptr(unsafe.Pointer(&x0)),
		H: make(chan int, 5),
		M: unsafe.Pointer(&x0),
		// E: []*X0{&x0},
		N: nn[1:3],
		O: a,
		Q: a,
	}

	t.Run("DeepCopy()", func(t *testing.T) {
		var ret interface{}
		x2ind := evendeep.X2{N: nn[1:3]}
		x2 := &x2ind

		ret = evendeep.DeepCopy(&x1, &x2, evendeep.WithIgnoreNames("Shit", "Memo", "Name"))
		testIfBadCopy(t, x1, x2ind, ret, "DeepCopy x1 -> x2", true)
	})
}
