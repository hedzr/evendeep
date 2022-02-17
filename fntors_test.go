package deepcopy

import (
	"reflect"
	"testing"
)

func testDeepEqual(t *testing.T, got, expect interface{}) {
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("expecting %v but got %v", expect, got)
	}
}

func TestCopySlice_cloneMode(t *testing.T) {
	// defer newCaptureLog(t).Release()

	c := newCloner()

	var so = []int{9, 77}
	var to = []int{}
	var err error

	var src = reflect.ValueOf(&so)
	var tgt = reflect.ValueOf(&to)

	err = copySlice(c, nil, src, tgt)
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to)
		testDeepEqual(t, to, []int{9, 77})
	}

	to = []int{2, 77}
	err = copySlice(c, nil, src, reflect.ValueOf(&to))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to)
		testDeepEqual(t, to, []int{2, 77, 9, 77})
	}

}

func TestCopySlice_mergeMode(t *testing.T) {
	// defer newCaptureLog(t).Release()

	c := newDeepCopier()

	var so = []int{9, 77}
	var to = []int{}
	var err error

	var src = reflect.ValueOf(&so)
	var tgt = reflect.ValueOf(&to)

	err = copySlice(c, nil, src, tgt)
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to)
		testDeepEqual(t, to, []int{9, 77})
	}

	to = []int{2, 77}
	err = copySlice(c, nil, src, reflect.ValueOf(&to))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to)
		testDeepEqual(t, to, []int{2, 77, 9})
	}

}

func TestCopyArray(t *testing.T) {
	// defer newCaptureLog(t).Release()

	c := newDeepCopier()

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
		t.Logf("tgt = %v", to)
		testDeepEqual(t, to2, [2]int{9, 77})
	}

}

func TestCopyChan(t *testing.T) {

	c := newDeepCopier()

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
