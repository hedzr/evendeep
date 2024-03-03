package tool

import (
	"reflect"
	"testing"
)

func TestPointerTo(t *testing.T) {
	var ii int
	ii = 8

	tv := reflect.TypeOf(ii)
	tv1 := PointerTo(tv)

	t.Logf("%v", tv1.String())
}
