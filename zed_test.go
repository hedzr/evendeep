package deepcopy_test

import (
	"github.com/hedzr/deepcopy"
	"testing"
	"unsafe"
)

func TestDeepCopyExternal(t *testing.T) {

	defer newCaptureLog(t).Release()

	nn := []int{2, 9, 77, 111, 23, 29}
	var a [2]string
	a[0] = "Hello"
	a[1] = "World"

	x0 := deepcopy.X0{}
	x1 := deepcopy.X1{
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
		x2 := &deepcopy.X2{N: nn[1:3]}

		ret = deepcopy.DeepCopy(&x1, &x2, deepcopy.WithIgnoreNames("Shit", "Memo", "Name"))
		testBadCopy(t, x1, *x2, ret, "DeepCopy x1 -> x2", true)

	})

}
