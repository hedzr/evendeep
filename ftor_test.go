package evendeep

import (
	"strconv"

	"github.com/hedzr/evendeep/flags/cms"
	"github.com/hedzr/evendeep/internal/tool"
	"github.com/hedzr/evendeep/typ"

	"reflect"
	"testing"
)

func testDeepEqual(printer func(msg string, args ...interface{}), got, expect typ.Any) {
	// a,b:=reflect.ValueOf(got),reflect.ValueOf(expect)
	// switch kind:=a.Kind();kind {
	// case reflect.Map:
	// case reflect.Slice:
	// }

	if !reflect.DeepEqual(got, expect) {
		printer("expecting %v but got %v", expect, got)
	}
}

func TestTestDeepEqual(t *testing.T) {
	// defer dbglog.NewCaptureLog(t).Release()
	var mm = []map[string]bool{
		nil, nil,
	}

	for i := 0; i < 2; i++ {
		mm[i] = make(map[string]bool)
		for s := range map[string]bool{"std": true, "sliceopy": true, "mapcopy": true, "omitempty": true} {
			mm[i][s] = true
		}
	}

	testDeepEqual(t.Errorf, mm[0], mm[1])
}

func TestCopyChan(t *testing.T) {

	c := newCopier()
	// params := newParams(withOwnersSimple(c, nil))

	var err error
	var so = make(chan struct{})
	var to chan struct{}

	err = copyChan(c, nil, reflect.ValueOf(so), reflect.ValueOf(&to))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to)
		// testDeepEqual(t, to2, [2]int{9, 77})
	}
}

func TestCopyUnsafePointer(t *testing.T) {
	// defer dbglog.NewCaptureLog(t).Release()

	// c := newDeepCopier()
	// params := newParams(withOwnersSimple(c, nil))
	//
	// var so = struct{ foo int }{42}
	// var to int
	// reflect.NewAt()
	// copyUnsafePointer(c, from, to)
}

func TestCopySlice_differModes(t *testing.T) {
	// defer dbglog.NewCaptureLog(t).Release()

	c := newCloner()
	params := newParams(withOwnersSimple(c, nil))

	// flags.LazyInitFieldTagsFlags()

	var so = []int{9, 77}
	var to = []int{}
	var err error

	var src = reflect.ValueOf(&so)
	var tgt = reflect.ValueOf(&to)

	err = copySlice(c, params, tool.Rdecodesimple(src), tool.Rdecodesimple(tgt))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to)
		testDeepEqual(t.Errorf, to, []int{9, 77})
	}

	to = []int{1}
	tgt = reflect.ValueOf(&to)
	err = copySlice(c, params, tool.Rdecodesimple(src), tool.Rdecodesimple(tgt))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to)
		testDeepEqual(t.Errorf, to, []int{9, 77})
	}

	to = []int{1}
	tgt = reflect.ValueOf(&to)
	err = copySlice(c, newParams(withFlags(cms.SliceCopyAppend), withOwnersSimple(c, nil)), tool.Rdecodesimple(src), tool.Rdecodesimple(tgt))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to)
		testDeepEqual(t.Errorf, to, []int{1, 9, 77})
	}

	to = []int{}
	tgt = reflect.ValueOf(&to)
	err = copySlice(c, newParams(withFlags(cms.SliceCopyAppend), withOwnersSimple(c, nil)), tool.Rdecodesimple(src), tool.Rdecodesimple(tgt))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to)
		testDeepEqual(t.Errorf, to, []int{9, 77})
	}

	to = []int{2, 9, 1}
	tgt = reflect.ValueOf(&to)
	err = copySlice(c, newParams(withFlags(cms.SliceCopyAppend), withOwnersSimple(c, nil)), tool.Rdecodesimple(src), tool.Rdecodesimple(tgt))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to)
		testDeepEqual(t.Errorf, to, []int{2, 9, 1, 9, 77})
	}

	so = []int{15, 2}
	src = reflect.ValueOf(&so)
	to = []int{2, 9, 1}
	tgt = reflect.ValueOf(&to)
	err = copySlice(c, newParams(withFlags(cms.SliceMerge), withOwnersSimple(c, nil)), tool.Rdecodesimple(src), tool.Rdecodesimple(tgt))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to)
		testDeepEqual(t.Errorf, to, []int{2, 9, 1, 15})
	}

	to = []int{3, 77, 2, 15}
	tgt = reflect.ValueOf(&to)
	err = copySlice(c, newParams(withFlags(cms.SliceMerge), withOwnersSimple(c, nil)), tool.Rdecodesimple(src), tool.Rdecodesimple(tgt))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to)
		testDeepEqual(t.Errorf, to, []int{3, 77, 2, 15})
	}

}

func TestCopySlice_mergeMode(t *testing.T) {
	// defer dbglog.NewCaptureLog(t).Release()

	c := newCopier().withFlags(cms.SliceMerge, cms.MapMerge)
	params := newParams(withOwnersSimple(c, nil))

	var so = []int{9, 77}
	var to = []int{}
	var err error

	var src = reflect.ValueOf(&so)
	var tgt = reflect.ValueOf(&to)

	err = copySlice(c, params, tool.Rdecodesimple(src), tool.Rdecodesimple(tgt))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to)
		testDeepEqual(t.Errorf, to, []int{9, 77})
	}

	to = []int{2, 77}
	tgt = reflect.ValueOf(&to)
	err = copySlice(c, params, tool.Rdecodesimple(src), tool.Rdecodesimple(tgt))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to)
		testDeepEqual(t.Errorf, to, []int{2, 77, 9})
	}

}

func TestCopyArray(t *testing.T) {
	// defer dbglog.NewCaptureLog(t).Release()

	c := newCopier().withFlags()
	params := newParams(withOwnersSimple(c, nil))

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
		testDeepEqual(t.Errorf, to, [5]int{9, 77, 13})
	}

	to2 := [2]int{77, 2}
	err = copyArray(c, params, src, reflect.ValueOf(&to2))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to2)
		testDeepEqual(t.Errorf, to2, [2]int{9, 77})
	}

	to2 = [2]int{}
	err = copyArray(c, params, src, reflect.ValueOf(&to2))
	if err != nil {
		t.Errorf("bad: %v", err)
	} else {
		t.Logf("tgt = %v", to2)
		testDeepEqual(t.Errorf, to2, [2]int{9, 77})
	}

}

func TestCopyStructSlice(t *testing.T) {

}

func TestPointerOfPre(t *testing.T) {
	type A struct {
		A int
	}
	var a = &A{9}
	var b = &a
	t.Logf("a = %v, %p", a, a)
	t.Logf("b = %v", b)
	av := reflect.ValueOf(a)
	ptr1 := av.Pointer()
	t.Logf("a.pointer = %v", strconv.FormatUint(uint64(ptr1), 16))
	np := reflect.New(av.Type())
	t.Logf("np = %v, typ = %v", tool.Valfmt(&np), tool.Typfmtv(&np))

	typ := av.Type() // type of *A
	val := reflect.New(typ)
	valElem := val.Elem()
	ptr, _ := newFromType(typ.Elem())
	valElem.Set(ptr)
	t.Logf("ptr = %+v", *(val.Interface().(**A)))

	// np.Elem().Set(av.Addr())
	// t.Logf("np = %v, typ = %v", tool.Valfmt(&np), tool.Typfmtv(&np))
	//
	// avp, ok := pointerOf(av)
	// if !ok {
	// 	t.Fail()
	// }
	// t.Logf("avp = %v", tool.Valfmt(&avp))
}
