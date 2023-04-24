package tool

import (
	"reflect"
	"testing"
)

func TestPtrOf(t *testing.T) {

	var i = 100
	v := reflect.ValueOf(&i)
	vind := Rindirect(v)
	vp := PtrOf(vind)
	t.Logf("ptr of i: %v, &i: %v", vp.Interface(), &i)
	vp.Elem().SetInt(99)
	t.Logf("i: %v", i)

}
