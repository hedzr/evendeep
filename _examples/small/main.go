package main

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"unsafe"

	"github.com/hedzr/evendeep"
	"github.com/hedzr/evendeep/diff"
	"github.com/hedzr/evendeep/flags/cms"
	"github.com/hedzr/evendeep/ref"
	"github.com/hedzr/evendeep/typ"
)

func main() {
	// println(logex.InDebugging())

	nn := []int{2, 9, 77, 111, 23, 29}
	var a [2]string
	a[0] = "Hello"
	a[1] = "World"
	var a3 = [3]string{"Hello", "World"}

	x0 := X0{}
	x1 := X1{
		A: uintptr(unsafe.Pointer(&x0)),
		H: make(chan int, 5), //nolint:gomnd //no need
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
	// log.Printf("expect.Q: %v", expect1.Q)
	t1 := &X2{K: &X0{}}
	log.Printf("t1.K.IsZero: %v", ref.IsZero(reflect.ValueOf(t1.K).Elem()))

	log.Printf("--------------- test 1")
	tgt := X2{N: nn[1:3]}
	log.Printf("    src: %+v", x1)
	log.Printf("    tgt: %+v", tgt)
	evendeep.Copy(x1, &tgt, evendeep.WithStrategiesReset(cms.Default))
	log.Printf("    tgt: %+v", tgt)
	log.Printf("    tgt.K: %+v", tgt.K)

	if delta, ok := evendeep.DeepDiff(*expect1, x1, diff.WithTreatEmptyStructPtrAsNilPtr(true), diff.WithCompareDifferentTypeStructs(true), diff.WithCompareDifferentSizeArrays(true)); !ok {
		log.Printf("want %v but got %v", expect1, tgt)
		log.Panicf("The diffs:\n%v", delta)
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

	// log.Infof("--------------- test 2")
	// log.Printf("   src: %+v", x1)
	// log.Printf("   tgt: %+v", x2)
	// evendeep.Copy(x1, &x2, evendeep.WithStrategies(evendeep.SliceMerge))
	// if reflect.DeepEqual(*expect2, x2) == false {
	//	if delta, ok := diff.New(*expect2, x2); !ok {
	//		log.Errorf("want: %v", *expect2)
	//		log.Errorf(" got: %v", x2)
	//		fmt.Println(delta)
	//		panic("unmatched")
	//	}
	// }

	log.Printf("--------------- test 3")
	// x2 = X2{N: []int{23, 8}}
	log.Printf("    src: %+v", x1)
	log.Printf("    tgt: %+v", x2)
	evendeep.Copy(x1, &x2)
	if delta, ok := diff.New(*expect2, x2, diff.WithTreatEmptyStructPtrAsNilPtr(true), diff.WithCompareDifferentTypeStructs(true), diff.WithCompareDifferentSizeArrays(true)); !ok {
		log.Printf("want: %v", *expect2)
		log.Printf(" got: %v", x2)
		fmt.Printf("diff: %v\n", delta)
		panic("unmatched ----- test 3")
	}
	log.Printf(" result: ok")
}

// X0 type for testing
type X0 struct{}

// X1 type for testing
type X1 struct {
	A uintptr
	B map[string]typ.Any
	C bytes.Buffer
	D []string
	E []*X0
	F chan struct{}
	G chan bool
	H chan int
	I func()
	J typ.Any
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
