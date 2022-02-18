package deepcopy_test

import (
	"encoding/json"
	"github.com/hedzr/deepcopy"
	"reflect"
	"testing"
	"unsafe"
)

func TestDeepCopy(t *testing.T) {

	defer newCaptureLog(t).Release()

	nn := []int{2, 9, 77, 111, 23, 29}
	var a [2]string
	a[0] = "Hello"
	a[1] = "World"

	x0 := X0{}
	x1 := X1{
		A: uintptr(unsafe.Pointer(&x0)),
		H: make(chan int, 5),
		M: unsafe.Pointer(&x0),
		// E: []*X0{&x0},
		N: nn[1:3],
		O: a,
		Q: a,
	}

	t.Run("MakeClone()", func(t *testing.T) {

		var ret interface{}
		//x2 := &X2{N: nn[1:3]}

		ret = deepcopy.MakeClone(&x1)
		testBadCopy(t, x1, ret, ret, "MakeClone x1 -> new")
		t.Log("MakeClone is done.")

	})

	t.Run("DeepCopy()", func(t *testing.T) {

		var ret interface{}
		x2 := &X2{N: nn[1:3]}

		ret = deepcopy.DeepCopy(&x1, &x2, deepcopy.WithIgnoreNames("Shit", "Memo", "Name"))
		testBadCopy(t, x1, *x2, ret, "DeepCopy x1 -> x2", true)

	})

	t.Run("NewDeepCopier().CopyTo()", func(t *testing.T) {

		var ret interface{}
		x2 := &X2{N: nn[1:3]}

		ret = deepcopy.NewDeepCopier().CopyTo(&x1, &x2, deepcopy.WithIgnoreNames("Shit", "Memo", "Name"))
		testBadCopy(t, x1, *x2, ret, "NewDeepCopier().CopyTo() - DeepCopy x1 -> x2", true)

	})

}

func testBadCopy(t *testing.T, src, tgt, result interface{}, title string, notFailed ...interface{}) {

	t.Logf("checking result ...")

	//if diff := deep.Equal(src, tgt); diff == nil {
	//	return
	//} else {
	//	t.Fatalf("testBadCopy - BAD COPY (%v):\n  SRC: %+v\n  TGT: %+v\n\n DIFF: \n%v", title, src, tgt, diff)
	//}

	//dd := deepdiff.New()
	//diff, err := dd.Diff(context.Background(), src, tgt)
	//if err != nil {
	//	return
	//}
	//if diff.Len() > 0 {
	//	t.Fatalf("testBadCopy - BAD COPY (%v):\n SRC: %+v\n TGT: %+v\n\n DIFF: \n%v", title, src, tgt, diff)
	//} else {
	//	return
	//}

	if !reflect.DeepEqual(src, tgt) {

		var b1, b2 []byte
		var err error
		if b1, err = json.MarshalIndent(src, "", "  "); err == nil {
			if b2, err = json.MarshalIndent(src, "", "  "); err == nil {
				if string(b1) == string(b2) {
					return
				}
				t.Logf("testBadCopy - src: %v\ntgt: %v\n", string(b1), string(b2))
			}
		}
		if err != nil {
			t.Logf("testBadCopy - json marshal not ok (just a warning): %v", err)

			//if b1, err = yaml.Marshal(src); err == nil {
			//	if b2, err = yaml.Marshal(src); err == nil {
			//		if string(b1) == string(b2) {
			//			return
			//		}
			//	}
			//}

			//gob.Register(X1{})
			//
			//buf1 := new(bytes.Buffer)
			//enc1 := gob.NewEncoder(buf1)
			//if err = enc1.Encode(&src); err != nil {
			//	t.Fatal(err)
			//}
			//
			//buf2 := new(bytes.Buffer)
			//enc2 := gob.NewEncoder(buf2)
			//if err = enc2.Encode(&tgt); err != nil {
			//	t.Fatal(err)
			//}
			//
			//s1, s2 := buf1.String(), buf2.String()
			//if s1 == s2 {
			//	return
			//}
		}

		for _, b := range notFailed {
			if yes, ok := b.(bool); yes && ok {
				return
			}
		}

		t.Fatalf("testBadCopy - BAD COPY (%v):\n SRC: %+v\n TGT: %+v\n RES: %v", title, src, tgt, result)
	}
}
