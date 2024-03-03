package ref_test

import (
	"io"
	"strings"
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
	t.Logf("     vi (%v (%v)) : %v, &b = %v", vi.Kind(), vi.Type(), vi.Interface(), &b)

	vv, prev := ref.Rdecode(vi)
	t.Logf("          vv (%v (%v)) : %v, &b = %v [ rdecode(vi) ]", vc.Kind(), vv.Type(), vv.Interface(), &b)
	value := prev.Interface()
	valptr := value.(*int) //nolint:errcheck //no need
	t.Logf("       prev (%v (%v)) : %v -> %v", prev.Kind(), prev.Type(), value, *valptr)

	// A result likes:
	//
	//           vb (int (int)) : 12, &b = 0xc00019c1d8
	//          vc (ptr (*int)) : 0xc00019c1d8, &b = 0xc00019c1d8, c = 0xc00019c1d8, *c -> 12
	//      vi (ptr (*typ.Any)) : 0xc000184360, &b = 0xc00019c1d8
	//           vv (ptr (int)) : 12, &b = 0xc00019c1d8 [ rdecode(vi) ]
	//        prev (ptr (*int)) : 0xc00019c1d8 -> 12
}

func TestRdecodesimple(t *testing.T) {
	b := 12
	c := &b

	vb := reflect.ValueOf(b)
	t.Logf("          vb (%v (%v)) : %v, &b = %v", vb.Kind(), vb.Type(), vb.Interface(), &b)

	vc := reflect.ValueOf(c)
	t.Logf("         vc (%v (%v)) : %v, &b = %v, c = %v, *c -> %v", vc.Kind(), vc.Type(), vc.Interface(), &b, c, *c)

	var ii typ.Any // nolint:gosimple

	ii = c

	vi := reflect.ValueOf(&ii)
	t.Logf("     vi (%v (%v)) : %v, &b = %v", vi.Kind(), vi.Type(), vi.Interface(), &b)

	vv := ref.Rdecodesimple(vi)
	t.Logf("          vv (%v (%v)) : %v, &b = %v [ rdecode(vi) ]", vc.Kind(), vv.Type(), vv.Interface(), &b)

	// A result likes:
	//
	//           vb (int (int)) : 12, &b = 0xc0000123c8
	//          vc (ptr (*int)) : 0xc0000123c8, &b = 0xc0000123c8, c = 0xc0000123c8, *c -> 12
	//      vi (ptr (*typ.Any)) : 0xc0000263b0, &b = 0xc0000123c8
	//           vv (ptr (int)) : 12, &b = 0xc0000123c8 [ rdecode(vi) ]
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

	var prevt reflect.Type
	v2t := reflect.TypeOf(&ii)
	v2t, prevt = ref.Rskiptype(v2t, reflect.Interface, reflect.Ptr)
	t.Logf("v2 (%v (%v)), prev (%v)", v2t.Kind(), prevt.Kind(), prevt)

	v2t = reflect.TypeOf(&ii)
	v2t, prevt = ref.Rdecodetype(v2t)
	t.Logf("v2 (%v (%v)), prev (%v)", v2t.Kind(), prevt.Kind(), prevt)

	v2t = reflect.TypeOf(&ii)
	v2t = ref.Rdecodetypesimple(v2t)
	t.Logf("v2 (%v (%v)), prev (%v)", v2t.Kind(), prevt.Kind(), prevt)

	ii = &b
	v2 = reflect.ValueOf(&ii)
	v2, prev = ref.Rdecode(v2)
	t.Logf("v2 (%v (%v)) : %v, prev (%v %v)", v2.Kind(), v2.Type(), v2.Interface(), prev.Kind(), prev.Type())

	// nolint:gocritic //no
	// var up = vi.Addr()
	// t.Logf("up = %v", up)
}

func TestRindirect(t *testing.T) {
	b := &Employee{Name: "Bob"}
	vb := reflect.ValueOf(b)
	t.Logf("vb (%v (%v)) : %v, &b = %v", vb.Kind(), vb.Type(), vb.Interface(), &b)

	v2 := ref.Rindirect(vb)
	t.Logf("v2 (%v (%v)) : %v", v2.Kind(), v2.Type(), v2.Interface())
	v2t := ref.RindirectType(vb.Type())
	t.Logf("v2t (%v (%v))", v2t.Kind(), v2t)

	v2 = ref.Rwant(vb, reflect.Struct)
	t.Logf("v2 (%v (%v)) : %v", v2.Kind(), v2.Type(), v2.Interface())

	v2 = ref.Rwant(vb, reflect.Ptr)
	t.Logf("v2 (%v (%v)) : %v", v2.Kind(), v2.Type(), v2.Interface())

	t.Log(ref.IsMap(vb))
	t.Log(ref.IsMap(reflect.ValueOf(map[string]bool{"a": false})))

	t.Log(ref.IsSlice(vb))
	t.Log(ref.IsSlice(reflect.ValueOf(map[string]bool{"a": false})))
	t.Log(ref.IsSlice(reflect.ValueOf([]bool{})))

	t.Log(ref.IsNumeric(vb))
	t.Log(ref.IsNumeric(reflect.ValueOf([]bool{})))
	t.Log(ref.IsNumeric(reflect.ValueOf(3)))
	t.Log(ref.IsNumeric(reflect.ValueOf(nil)))

	t.Log(ref.IsNumericKind(vb.Kind()))
	t.Log(ref.IsNumericKind(reflect.ValueOf([]bool{}).Kind()))
	t.Log(ref.IsNumericKind(reflect.ValueOf(3).Kind()))

	t.Log(ref.KindIs(vb.Kind(), reflect.Ptr, reflect.Slice))
	t.Log(ref.KindIs(reflect.ValueOf([]bool{}).Kind(), reflect.Ptr, reflect.Slice))
	t.Log(ref.KindIs(reflect.ValueOf(3).Kind(), reflect.Ptr, reflect.Slice))
}

func TestSliceAppend(t *testing.T) {
	// a:=map[string]bool{}
	var a []int

	va := reflect.ValueOf(a)
	vb := []int{1}
	vc := reflect.ValueOf([]int{2})
	vd := reflect.ValueOf([]int{2, 3})
	b := ref.SliceAppend(va, vb, vc, vd)

	t.Log(a)
	t.Log(b)
	if !reflect.DeepEqual(b, []int{1, 2, 2, 3}) {
		t.Fail()
	}

	b = ref.SliceMerge(va, vb, vc, vd)
	t.Log("merged slice b is: ", b)
	if !reflect.DeepEqual(b, []int{1, 2, 3}) {
		t.Fail()
	}
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
	t.Logf("fmt: %v", ref.ValfmtPure(&v1))

	x1 = "ok"
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumericType(v1.Type()) != true, false, true)
	t.Logf("fmt: %v", ref.ValfmtPure(&v1))

	//

	x1 = 13
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumIntegerType(v1.Type()) == true, true, false)
	t.Logf("fmt: %v", ref.ValfmtPure(&v1))

	x1 = uint(13)
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumIntegerType(v1.Type()) == true, true, false)
	t.Logf("fmt: %v", ref.ValfmtPure(&v1))

	x1 = float32(13)
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumIntegerType(v1.Type()) != true, false, true)
	t.Logf("fmt: %v", ref.ValfmtPure(&v1))

	//

	x1 = 13
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumSIntegerKind(v1.Kind()) == true, true, false)
	t.Logf("fmt: %v", ref.ValfmtPure(&v1))

	x1 = uint(13)
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumSIntegerKind(v1.Kind()) != true, false, true)
	t.Logf("fmt: %v", ref.ValfmtPure(&v1))

	x1 = float32(13)
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumSIntegerKind(v1.Kind()) != true, false, true)
	t.Logf("fmt: %v", ref.ValfmtPure(&v1))

	//

	x1 = uint(13)
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumUIntegerKind(v1.Kind()) == true, true, false)
	t.Logf("fmt: %v", ref.ValfmtPure(&v1))

	x1 = int(13)
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumUIntegerKind(v1.Kind()) != true, false, true)
	t.Logf("fmt: %v", ref.ValfmtPure(&v1))

	x1 = float32(13)
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumUIntegerKind(v1.Kind()) != true, false, true)
	t.Logf("fmt: %v", ref.ValfmtPure(&v1))

	//

	x1 = 13
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumFloatKind(v1.Kind()) != true, false, true)
	t.Logf("fmt: %v", ref.ValfmtPure(&v1))

	x1 = float32(13)
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumFloatKind(v1.Kind()) == true, true, false)
	t.Logf("fmt: %v", ref.ValfmtPure(&v1))

	x1 = "ok"
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumericType(v1.Type()) != true, false, true)
	t.Logf("fmt: %v", ref.ValfmtPure(&v1))

	x1 = true
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumericType(v1.Type()) != true, false, true)
	t.Logf("fmt: %v", ref.ValfmtPure(&v1))

	x1 = false
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.IsNumericType(v1.Type()) != true, false, true)
	assertYes(t, ref.IsZero(v1), true, false)
	t.Logf("fmt: %v", ref.ValfmtPure(&v1))

	v1 = reflect.ValueOf(&x1)
	// t.Logf("fmt false: %v", ref.ValfmtPure(&v1))
	assertYes(t, ref.IsNumericType(v1.Type()) != true, false, true)
	// t.Logf("fmt false: %v", ref.ValfmtPure(&v1))
	assertYes(t, ref.IsZero(v1) != true, true, false)
	t.Logf("fmt false: %v", ref.ValfmtPure(&v1))

	x1 = [3]int{1, 2, 3}
	v1 = reflect.ValueOf(x1)
	// t.Logf("slice: %v", ref.ValfmtPure(&v1))
	assertYes(t, ref.IsNumericType(v1.Type()) == false, true, false)
	// t.Logf("slice: %v", ref.ValfmtPure(&v1))
	assertYes(t, ref.IsZero(v1) == false, false, true)
	// t.Logf("slice: %v", ref.ValfmtPure(&v1))
	assertYes(t, ref.Iserrortype(v1.Type()) == false, false, true)
	t.Logf("slice: %v", ref.ValfmtPure(&v1))

	x1 = io.ErrShortBuffer
	v1 = reflect.ValueOf(x1)
	assertYes(t, ref.Iserrortype(v1.Type()), true, false)
	assertYes(t, ref.IsValid(v1), false, true)
	assertYes(t, ref.IsValidv(&v1), false, true)
	assertYes(t, ref.IsNil(v1) == false, false, true)
	assertYes(t, ref.IsNilv(&v1) == false, false, true)

	var sb strings.Builder
	v1 = reflect.ValueOf(sb)
	assertYes(t, ref.IsNumericType(v1.Type()) != true, false, true)
	t.Logf("fmt: %v", ref.ValfmtPure(&v1))

	v1 = reflect.ValueOf(&sb)
	assertYes(t, ref.IsNumericType(v1.Type()) != true, false, true)
	t.Logf("fmt: %v", ref.ValfmtPure(&v1))

	ss := struct{}{}
	v1 = reflect.ValueOf(ss)
	assertYes(t, ref.IsNumericType(v1.Type()) != true, false, true)
	assertYes(t, ref.IsValid(v1), false, true)
	assertYes(t, ref.IsValidv(&v1), false, true)
	assertYes(t, ref.IsNil(v1) == false, false, true)
	assertYes(t, ref.IsNilv(&v1) == false, false, true)
	t.Logf("fmt: %v", ref.ValfmtPure(&v1))

	var x2 *int
	v2 := reflect.ValueOf(x2)
	t.Logf("fmt: %v", ref.ValfmtPure(&v2))
	assertYes(t, ref.IsValid(v2), false, true)
	assertYes(t, ref.IsValidv(&v2), false, true)
	assertYes(t, ref.IsNil(v2), false, true)
	assertYes(t, ref.IsNilv(&v2), false, true)

	var ch1 chan struct{}
	v2 = reflect.ValueOf(ch1)
	t.Logf("fmt: %v", ref.ValfmtPure(&v2))
	assertYes(t, ref.IsValid(v2), false, true)
	assertYes(t, ref.IsValidv(&v2), false, true)
	assertYes(t, ref.IsNil(v2), false, true)
	assertYes(t, ref.IsNilv(&v2), false, true)

	var x int = 7
	x2 = &x
	v2 = reflect.ValueOf(x2)
	t.Logf("fmt: %v", ref.ValfmtPure(&v2))
	assertYes(t, ref.IsValid(v2), false, true)
	assertYes(t, ref.IsValidv(&v2), false, true)
	assertYes(t, ref.IsNil(v2) == false, false, true)
	assertYes(t, ref.IsNilv(&v2) == false, false, true)

	t.Logf("fmt: %v", ref.ValfmtPure(nil))
}

func TestIsExported(t *testing.T) {
	type aS struct {
		exp int
		E   bool
	}
	var a aS

	v1 := reflect.ValueOf(a)
	// v2=v1.Field(0)
	t1 := v1.Type().Field(0)
	t2 := v1.Type().Field(1)
	assertYes(t, ref.IsExported(&t1) == false, false, true)
	assertYes(t, ref.IsExported(&t2), false, true)
}

func TestCanConvertHelper(t *testing.T) {
	type aS struct {
		exp int
		E   bool
	}
	var a aS

	v1 := reflect.ValueOf(a)

	var i int
	v2 := reflect.ValueOf(i)

	assertYes(t, ref.CanConvertHelper(v1, v2.Type()) == false, false, true)
	assertYes(t, ref.CanConvert(&v1, v2.Type()) == false, false, true)

	var i64 int64
	v3 := reflect.ValueOf(i64)
	assertYes(t, ref.CanConvertHelper(v3, v2.Type()), false, true)
	assertYes(t, ref.CanConvert(&v3, v2.Type()), false, true)

	s1 := [3]int{1, 2, 3}
	s2 := []int{1, 2, 3}
	s3 := []int{}
	v1 = reflect.ValueOf(s1)
	v2 = reflect.ValueOf(s2)
	v3 = reflect.ValueOf(s3)
	assertYes(t, ref.CanConvertHelper(v1, v2.Type()) == false, false, true)
	assertYes(t, ref.CanConvertHelper(v3, v2.Type()) == true, false, true)
}

func TestTypefmt(t *testing.T) {
	x1 := float32(13)
	v1 := reflect.ValueOf(x1)

	t.Logf("fmt: %v", ref.Typfmtvlite(&v1))
	t.Logf("fmt: %v", ref.Typfmtv(&v1))

	t.Logf("fmt: %v", ref.Typfmt(v1.Type()))

	typ := v1.Type()
	t.Logf("fmt: %v", ref.Typfmtptr(&typ))

	t.Logf("fmt: %v", ref.Valfmtptr(&v1))
	t.Logf("fmt: %v", ref.ValfmtptrPure(&v1))
	t.Logf("fmt: %v", ref.Valfmtv(v1))
	t.Logf("fmt: %v", ref.Valfmt(&v1))

	t.Logf("fmt: %v", ref.ValfmtPure(&v1))
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
