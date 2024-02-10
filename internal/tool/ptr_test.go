package tool

import (
	"reflect"
	"testing"

	"github.com/hedzr/evendeep/ref"
)

func TestPtrOf(t *testing.T) {

	var i = 100
	v := reflect.ValueOf(&i)
	vind := ref.Rindirect(v)
	vp := PtrOf(vind)
	t.Logf("ptr of i: %v, &i: %v", vp.Interface(), &i)
	vp.Elem().SetInt(99)
	t.Logf("i: %v", i)

}
