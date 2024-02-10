package ref_test

import (
	"time"

	"github.com/hedzr/evendeep/ref"
	"github.com/hedzr/evendeep/typ"

	"reflect"
	"testing"
)

func TestRdecode(t *testing.T) {
	b := 12
	c := &b

	vb := reflect.ValueOf(b)
	t.Logf("          vb (%v (%v)) : %v, &b = %v", vb.Kind(), vb.Type(), vb.Interface(), &b)

	vc := reflect.ValueOf(c)
	t.Logf("         vc (%v (%v)) : %v, &b = %v, c = %v, *c -> %v", vc.Kind(), vc.Type(), vc.Interface(), &b, c, *c)

	var ii typ.Any // nolint:gosimple

	ii = c

	vi := reflect.ValueOf(&ii)
	t.Logf("vi (%v (%v)) : %v, &b = %v", vi.Kind(), vi.Type(), vi.Interface(), &b)

	vv, prev := ref.Rdecode(vi)
	t.Logf("          vv (%v (%v)) : %v, &b = %v [ rdecode(vi) ]", vc.Kind(), vv.Type(), vv.Interface(), &b)
	value := prev.Interface()
	valptr := value.(*int) //nolint:errcheck //no need
	t.Logf("       prev (%v (%v)) : %v -> %v", prev.Kind(), prev.Type(), value, *valptr)

	// A result likes:
	//
	//    vb (int (int)) : 12, &b = 0xc00001e350
	//    vc (ptr (*int)) : 0xc00001e350, &b = 0xc00001e350, c = 0xc00001e350, *c -> 12
	//    vi (ptr (*interface {})) : 0xc00006c640, &b = 0xc00001e350
	//    vv (ptr (int)) : 12, &b = 0xc00001e350 [ rdecode(vi) ]
	//    prev (ptr (*int)) : 0xc00001e350 -> 12
}

func Test1(t *testing.T) {
	b := 12

	vb := reflect.ValueOf(b)
	t.Logf("vb (%v (%v)) : %v, &b = %v", vb.Kind(), vb.Type(), vb.Interface(), &b)

	var ii typ.Any // nolint:gosimple

	ii = b
	vi := reflect.ValueOf(&ii)
	t.Logf("vi (%v (%v)) : %v, &b = %v", vi.Kind(), vi.Type(), vi.Interface(), &b)
	vi = vi.Elem()
	t.Logf("vi (%v (%v)) : %v, &b = %v", vi.Kind(), vi.Type(), vi.Interface(), &b)
	vi = vi.Elem()
	t.Logf("vi (%v (%v)) : %v, &b = %v", vi.Kind(), vi.Type(), vi.Interface(), &b)

	// nolint:gocritic //no
	// var up = vi.Addr()
	// t.Logf("up = %v", up)
}

func TestRskipRdecodeAndSoOn(t *testing.T) {
	b := &Employee{Name: "Bob"}

	vb := reflect.ValueOf(b)
	t.Logf("vb (%v (%v)) : %v, &b = %v", vb.Kind(), vb.Type(), vb.Interface(), &b)

	var ii typ.Any

	ii = b
	vi := reflect.ValueOf(&ii)
	kind := vi.Kind()
	t.Logf("vi (%v (%v)) : %v, &b = %v", kind, vi.Type(), vi.Interface(), &b)
	vi = vi.Elem()
	t.Logf("vi (%v (%v)) : %v, &b = %v", vi.Kind(), vi.Type(), vi.Interface(), &b)
	vi = vi.Elem()
	t.Logf("vi (%v (%v)) : %v, &b = %v", vi.Kind(), vi.Type(), vi.Interface(), &b)
	vi = vi.Elem()
	t.Logf("vi (%v (%v)) : %v, &b = %v", vi.Kind(), vi.Type(), vi.Interface(), &b)

	var prev reflect.Value
	v2 := reflect.ValueOf(&ii)

	// nolint:gocritic //no
	// c := newCopier()

	v2, prev = ref.Rskip(v2, reflect.Interface, reflect.Ptr)
	t.Logf("v2 (%v (%v)) : %v, prev (%v %v)", v2.Kind(), v2.Type(), v2.Interface(), prev.Kind(), prev.Type())

	ii = &b
	v2 = reflect.ValueOf(&ii)
	v2, prev = ref.Rdecode(v2)
	t.Logf("v2 (%v (%v)) : %v, prev (%v %v)", v2.Kind(), v2.Type(), v2.Interface(), prev.Kind(), prev.Type())

	// nolint:gocritic //no
	// var up = vi.Addr()
	// t.Logf("up = %v", up)
}

func assertYes(t *testing.T, b bool, expect, got interface{}) {
	if !b {
		t.Fatalf("expecting %v but got %v", expect, got)
	}
}

func TestIsNumXXX(t *testing.T) {
	var x1 interface{} = 9527 + 0i
	var v1 = reflect.ValueOf(x1)

	assertYes(t, ref.IsNumericType(v1.Type()) == true, true, false)
	assertYes(t, ref.IsNumComplexKind(v1.Kind()) == true, true, false)

	x1 = "ok"
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumericType(v1.Type()) != true, false, true)

	//

	x1 = 13
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumIntegerType(v1.Type()) == true, true, false)

	x1 = uint(13)
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumIntegerType(v1.Type()) == true, true, false)

	x1 = float32(13)
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumIntegerType(v1.Type()) != true, false, true)

	//

	x1 = 13
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumSIntegerKind(v1.Kind()) == true, true, false)

	x1 = uint(13)
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumSIntegerKind(v1.Kind()) != true, false, true)

	x1 = float32(13)
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumSIntegerKind(v1.Kind()) != true, false, true)

	//

	x1 = uint(13)
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumUIntegerKind(v1.Kind()) == true, true, false)

	x1 = int(13)
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumUIntegerKind(v1.Kind()) != true, false, true)

	x1 = float32(13)
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumUIntegerKind(v1.Kind()) != true, false, true)

	//

	x1 = 13
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumFloatKind(v1.Kind()) != true, false, true)

	x1 = float32(13)
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumFloatKind(v1.Kind()) == true, true, false)

	x1 = "ok"
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumericType(v1.Type()) != true, false, true)
}

// Employee type for testing
type Employee struct {
	Name      string
	Birthday  *time.Time
	F11       float32
	F12       float64
	C11       complex64
	C12       complex128
	Feat      []byte
	Sptr      *string
	Nickname  *string
	Age       int64
	FakeAge   int
	EmployeID int64
	DoubleAge int32
	SuperRule string
	Notes     []string
	RetryU    uint8
	TimesU    uint16
	FxReal    uint32
	FxTime    int64
	FxTimeU   uint64
	UxA       uint
	UxB       int
	Retry     int8
	Times     int16
	Born      *int
	BornU     *uint
	// nolint:unused
	flags []byte // nolint:structcheck
	Bool1 bool
	Bool2 bool
	Ro    []int
}
