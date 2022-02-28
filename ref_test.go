package deepcopy

import (
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

	var ii interface{}

	ii = c

	vi := reflect.ValueOf(&ii)
	t.Logf("vi (%v (%v)) : %v, &b = %v", vi.Kind(), vi.Type(), vi.Interface(), &b)

	vv, prev := rdecode(vi)
	t.Logf("          vv (%v (%v)) : %v, &b = %v [ rdecode(vi) ]", vc.Kind(), vv.Type(), vv.Interface(), &b)
	value := prev.Interface()
	valptr := value.(*int)
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

	var ii interface{}

	ii = b
	vi := reflect.ValueOf(&ii)
	t.Logf("vi (%v (%v)) : %v, &b = %v", vi.Kind(), vi.Type(), vi.Interface(), &b)
	vi = vi.Elem()
	t.Logf("vi (%v (%v)) : %v, &b = %v", vi.Kind(), vi.Type(), vi.Interface(), &b)
	vi = vi.Elem()
	t.Logf("vi (%v (%v)) : %v, &b = %v", vi.Kind(), vi.Type(), vi.Interface(), &b)
	//var up = vi.Addr()
	//t.Logf("up = %v", up)
}

func Test2(t *testing.T) {
	b := &Employee{Name: "Bob"}

	vb := reflect.ValueOf(b)
	t.Logf("vb (%v (%v)) : %v, &b = %v", vb.Kind(), vb.Type(), vb.Interface(), &b)

	var ii interface{}

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

	// c := newCopier()

	v2, prev = rskip(v2, reflect.Interface, reflect.Ptr)
	t.Logf("v2 (%v (%v)) : %v, prev (%v %v)", v2.Kind(), v2.Type(), v2.Interface(), prev.Kind(), prev.Type())

	ii = &b
	v2 = reflect.ValueOf(&ii)
	v2, prev = rdecode(v2)
	t.Logf("v2 (%v (%v)) : %v, prev (%v %v)", v2.Kind(), v2.Type(), v2.Interface(), prev.Kind(), prev.Type())

	//var up = vi.Addr()
	//t.Logf("up = %v", up)
}
