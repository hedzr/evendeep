package deepcopy_test

import (
	"github.com/hedzr/deepcopy"
	"github.com/hedzr/log"
	"reflect"
	"testing"
	"unsafe"
)

func TestNormal(t *testing.T) {
	// config := log.NewLoggerConfigWith(true, "logrus", "trace")
	// logger := logrus.NewWithConfig(config)
	log.Printf("hello")
	log.Infof("hello info")
	log.Warnf("hello warn")
	log.Errorf("hello error")
	log.Debugf("hello debug")
	log.Tracef("hello trace")
}

func TestCpChan(t *testing.T) {
	var val = make(chan int, 10)
	vv := reflect.ValueOf(&val)
	vi := reflect.Indirect(vv)
	value := reflect.MakeChan(vi.Type(), vi.Cap())
	t.Logf("%v (len: %v),  vv.len: %v", value.Interface(), value.Cap(), vi.Cap())

	var sval chan string
	var strVal reflect.Value = reflect.ValueOf(&sval)
	indirectStr := reflect.Indirect(strVal)
	svalue := reflect.MakeChan(indirectStr.Type(), 1024)
	t.Logf("Type : [%v] \nCapacity : [%v]", svalue.Kind(), svalue.Cap())

}

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
	x2 := &X2{N: nn[1:3]}

	var ret interface{}

	t.Log("MakeClone() -------------")
	ret = deepcopy.MakeClone(x1)
	testBadCopy(t, x1, ret, ret, "MakeClone x1 -> new")
	t.Log("MakeClone is done.")

	t.Log("DeepCopy() -------------")
	ret = deepcopy.DeepCopy(&x1, &x2, deepcopy.WithIgnoreNames("Shit", "Memo", "Name"))
	testBadCopy(t, x1, x2, ret, "DeepCopy x1 -> x2")

	t.Log("NewDeepCopier().CopyTo() -------------")
	ret = deepcopy.NewDeepCopier().CopyTo(&x1, &x2, deepcopy.WithIgnoreNames("Shit", "Memo", "Name"))
	testBadCopy(t, x1, x2, ret, "NewDeepCopier().CopyTo() - DeepCopy x1 -> x2")

}

func testBadCopy(t *testing.T, src, tgt, result interface{}, title string) {
	if !reflect.DeepEqual(src, tgt) {
		t.Fatalf("BAD COPY (%v):\n SRC: %v\n TGT: %v\n RES: %v", title, src, tgt, result)
	}
}
