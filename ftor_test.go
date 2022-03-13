package deepcopy

import (
	"reflect"
	"testing"
)

func testDeepEqual(t *testing.T, got, expect interface{}) {
	//a,b:=reflect.ValueOf(got),reflect.ValueOf(expect)
	//switch kind:=a.Kind();kind {
	//case reflect.Map:
	//case reflect.Slice:
	//}

	if !reflect.DeepEqual(got, expect) {
		t.Errorf("expecting %v but got %v", expect, got)
	}
}

func TestTestDeepEqual(t *testing.T) {
	// defer newCaptureLog(t).Release()
	var mm = []map[string]bool{
		nil, nil,
	}

	for i := 0; i < 2; i++ {
		mm[i] = make(map[string]bool)
		for s := range map[string]bool{"std": true, "sliceopy": true, "mapcopy": true, "omitempty": true} {
			mm[i][s] = true
		}
	}

	testDeepEqual(t, mm[0], mm[1])
}

func TestCopySlice_differModes(t *testing.T) {
	// defer newCaptureLog(t).Release()

	c := newCloner()

	lazyInitFieldTagsFlags()

	var so = []int{9, 77}
	var to = []int{}
	var err error

	var src = reflect.ValueOf(&so)
	var tgt = reflect.ValueOf(&to)

	err = copySlice(c, nil, rdecodesimple(src), rdecodesimple(tgt))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to)
		testDeepEqual(t, to, []int{9, 77})
	}

	to = []int{1}
	tgt = reflect.ValueOf(&to)
	err = copySlice(c, nil, rdecodesimple(src), rdecodesimple(tgt))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to)
		testDeepEqual(t, to, []int{9, 77})
	}

	to = []int{1}
	tgt = reflect.ValueOf(&to)
	err = copySlice(c, newParams(withFlags(SliceCopyAppend)), rdecodesimple(src), rdecodesimple(tgt))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to)
		testDeepEqual(t, to, []int{1, 9, 77})
	}

	to = []int{}
	tgt = reflect.ValueOf(&to)
	err = copySlice(c, newParams(withFlags(SliceCopyAppend)), rdecodesimple(src), rdecodesimple(tgt))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to)
		testDeepEqual(t, to, []int{9, 77})
	}

	to = []int{2, 9, 1}
	tgt = reflect.ValueOf(&to)
	err = copySlice(c, newParams(withFlags(SliceCopyAppend)), rdecodesimple(src), rdecodesimple(tgt))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to)
		testDeepEqual(t, to, []int{2, 9, 1, 9, 77})
	}

	so = []int{15, 2}
	src = reflect.ValueOf(&so)
	to = []int{2, 9, 1}
	tgt = reflect.ValueOf(&to)
	err = copySlice(c, newParams(withFlags(SliceMerge)), rdecodesimple(src), rdecodesimple(tgt))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to)
		testDeepEqual(t, to, []int{2, 9, 1, 15})
	}

	to = []int{3, 77, 2, 15}
	tgt = reflect.ValueOf(&to)
	err = copySlice(c, newParams(withFlags(SliceMerge)), rdecodesimple(src), rdecodesimple(tgt))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to)
		testDeepEqual(t, to, []int{3, 77, 2, 15})
	}

}

func TestCopySlice_mergeMode(t *testing.T) {
	// defer newCaptureLog(t).Release()

	c := newCopier().withFlags(SliceMerge, MapMerge)

	var so = []int{9, 77}
	var to = []int{}
	var err error

	var src = reflect.ValueOf(&so)
	var tgt = reflect.ValueOf(&to)

	err = copySlice(c, nil, rdecodesimple(src), rdecodesimple(tgt))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to)
		testDeepEqual(t, to, []int{9, 77})
	}

	to = []int{2, 77}
	tgt = reflect.ValueOf(&to)
	err = copySlice(c, nil, rdecodesimple(src), rdecodesimple(tgt))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to)
		testDeepEqual(t, to, []int{2, 77, 9})
	}

}

func TestCopyArray(t *testing.T) {
	// defer newCaptureLog(t).Release()

	c := newCopier().withFlags()

	var so = [3]int{9, 77, 13}
	var to = [5]int{}
	var err error

	var src = reflect.ValueOf(&so)
	var tgt = reflect.ValueOf(&to)

	err = copyArray(c, nil, src, tgt)
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to)
		testDeepEqual(t, to, [5]int{9, 77, 13})
	}

	to2 := [2]int{77, 2}
	err = copyArray(c, nil, src, reflect.ValueOf(&to2))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to2)
		testDeepEqual(t, to2, [2]int{9, 77})
	}

	to2 = [2]int{}
	err = copyArray(c, nil, src, reflect.ValueOf(&to2))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to2)
		testDeepEqual(t, to2, [2]int{9, 77})
	}

}

func TestCopyChan(t *testing.T) {

	c := newCopier()

	var err error
	var so = make(chan struct{})
	var to chan struct{}

	err = copyChan(c, nil, reflect.ValueOf(so), reflect.ValueOf(to))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to)
		//testDeepEqual(t, to2, [2]int{9, 77})
	}
}

func TestCopyUnsafePointer(t *testing.T) {
	// defer newCaptureLog(t).Release()

	//c := newDeepCopier()
	//
	//var so = struct{ foo int }{42}
	//var to int
	//reflect.NewAt()
	//copyUnsafePointer(c, from, to)
}
