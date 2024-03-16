//go:build go1.18
// +build go1.18

package ref_test

import (
	"testing"
	"unsafe"

	"github.com/hedzr/evendeep/ref"
)

func TestIsNilT(t *testing.T) {
	// v2 := reflect.ValueOf(x2)
	// t.Logf("fmt: %v", ref.ValfmtPure(&v2))

	var x2 *int
	assertYes(t, ref.IsNilT(x2), false, true)

	var i int
	assertYes(t, ref.IsNilT(i) == false, false, true)

	var u uintptr
	assertYes(t, ref.IsNilT(u) == false, false, true)

	var uf unsafe.Pointer
	assertYes(t, ref.IsNilT(uf), false, true)

	var ss []int
	assertYes(t, ref.IsNilT(ss), false, true)

	var a any
	assertYes(t, ref.IsNilT(a) == false, false, true)
}
