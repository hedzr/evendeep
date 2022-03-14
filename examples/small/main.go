package main

import (
	"bytes"
	"github.com/hedzr/deepcopy"
	"github.com/hedzr/log"
	"reflect"
	"unsafe"
)

func main() {

	nn := []int{2, 9, 77, 111, 23, 29}
	var a [2]string
	a[0] = "Hello"
	a[1] = "World"
	var a3 = [3]string{"Hello", "World"}

	x0 := X0{}
	x1 := X1{
		A: uintptr(unsafe.Pointer(&x0)),
		H: make(chan int, 5),
		M: unsafe.Pointer(&x0),
		// E: []*X0{&x0},
		N: nn[1:5],
		O: a,
		Q: a,
	}

	expect1 := &X2{
		A: uintptr(unsafe.Pointer(&x0)),
		H: x1.H,
		M: unsafe.Pointer(&x0),
		// E: []*X0{&x0},
		N: nn[1:5],
		O: a,
		Q: a3,
	}
	x2 := X2{N: []int{23, 8}}
	expect2 := &X2{
		A: uintptr(unsafe.Pointer(&x0)),
		H: x1.H,
		M: unsafe.Pointer(&x0),
		// E: []*X0{&x0},
		N: []int{23, 8, 9, 77, 111}, // Note: [23,8] + [9,77,111,23] -> [23,8,9,77,111]
		O: a,
		Q: a3,
	}
	log.Printf("expect.Q: %v", expect1.Q)

	log.Printf("   src: %+v", x1)
	log.Printf("   tgt: %+v", X2{N: nn[1:3]})

	tgt := X2{N: nn[1:3]}

	deepcopy.Copy(x1, &tgt, deepcopy.WithStrategiesReset())
	if reflect.DeepEqual(tgt, *expect1) == false {
		log.Panicf("want %v but got %v", expect1, tgt)
	}

	deepcopy.Copy(x1, &x2) // , deepcopy.WithStrategies(deepcopy.SliceMerge))
	if reflect.DeepEqual(x2, *expect2) == false {
		log.Panicf("want %v but got %v", expect2, x2)
	}

}

// X0 type for testing
type X0 struct{}

// X1 type for testing
type X1 struct {
	A uintptr
	B map[string]interface{}
	C bytes.Buffer
	D []string
	E []*X0
	F chan struct{}
	G chan bool
	H chan int
	I func()
	J interface{}
	K *X0
	L unsafe.Pointer
	M unsafe.Pointer
	N []int
	O [2]string
	P [2]string
	Q [2]string
}

// X2 type for testing
type X2 struct {
	A uintptr
	B map[string]interface{}
	C bytes.Buffer
	D []string
	E []*X0
	F chan struct{}
	G chan bool
	H chan int
	I func()
	J interface{}
	K *X0
	L unsafe.Pointer
	M unsafe.Pointer
	N []int `copy:",slicemerge"`
	O [2]string
	P [2]string
	Q [3]string `copy:",slicecopy"`
}
