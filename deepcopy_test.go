package evendeep_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/hedzr/evendeep"
	"github.com/hedzr/evendeep/flags"
	"github.com/hedzr/evendeep/flags/cms"
	"github.com/hedzr/evendeep/typ"
)

type FF interface {
	Flags() flags.Flags
}

func TestFlagsRevert(t *testing.T) {
	saved := evendeep.DefaultCopyController.Flags().Clone()
	evendeep.DefaultCopyController.Flags().WithFlags(cms.SliceCopyAppend)
	evendeep.DefaultCopyController.SetFlags(saved)

	if c, ok := evendeep.New().(FF); ok {
		nf := c.Flags()
		b := reflect.DeepEqual(evendeep.DefaultCopyController.Flags(), nf)
		evendeep.HelperAssertYes(t, b, nf, evendeep.DefaultCopyController.Flags())
	}
}

func TestWithXXX(t *testing.T) {
	copier := evendeep.NewForTest()

	type AA struct {
		TestString string
		X          string
	}
	type BB struct {
		X string
	}

	t.Run("string to duration", func(t *testing.T) {
		var dur time.Duration
		src := "9h71ms"
		// svv := reflect.ValueOf(src)
		// tvv := reflect.ValueOf(&dur) // .Elem()

		err := copier.CopyTo(src, &dur)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		t.Logf("res: %v", dur)
	})

	t.Run("ignore names test", func(t *testing.T) {
		src1 := &AA{TestString: "well", X: "ok"}
		tgt1 := &BB{X: "no"}
		err := copier.CopyTo(src1, &tgt1)
		t.Logf("res bb: %+v", tgt1)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if tgt1.X != src1.X {
			t.Fatalf("err: after 'TestString' field was ignored, AA.X should be copied as BB.X")
		}
	})

	t.Run("non-ignore names test", func(t *testing.T) {
		copier = evendeep.New()

		src1 := &AA{TestString: "well", X: "ok"}
		tgt1 := &BB{X: "no"}
		err := copier.CopyTo(src1, &tgt1)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if tgt1.X != src1.TestString {
			t.Fatalf("err: if 'TestString' field was not ignored, AA.TestString should be copied as BB.X")
		}
		t.Logf("res bb: %+v", tgt1)
	})

	t.Run("ignore field test", func(t *testing.T) {
		copier = evendeep.New()

		type AAA struct {
			X1 string `copy:"-"`
			X2 string `copy:",-"`
			Y  int
		}
		type BBB struct {
			X1 string
			X2 string
			Y  int
		}
		src1 := &AAA{X1: "ok", X2: "well", Y: 1}
		tgt1 := &BBB{X1: "no", X2: "longer", Y: -1}
		err := copier.CopyTo(src1, &tgt1, evendeep.WithSyncAdvancing(true))
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		t.Logf("res bb: %+v", tgt1)

		if tgt1.X1 == src1.X1 || tgt1.X2 == src1.X2 {
			t.Fatalf("err: 'X1','X2' fields should be ignored, the fact is bad")
		}
		if tgt1.Y != src1.Y {
			t.Fatalf("err: 'Y' field should be copied.")
		}
	})

	t.Run("ignore field test - no sync advance", func(t *testing.T) {
		copier = evendeep.New()

		type AAA struct {
			X1 string `copy:"-"`
			X2 string `copy:",-"`
			Y  int
		}
		type BBB struct {
			Y int
		}
		src1 := &AAA{X1: "ok", X2: "well", Y: 1}
		tgt1 := &BBB{Y: -1}
		err := copier.CopyTo(src1, &tgt1)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		t.Logf("res bb: %+v", tgt1)

		if tgt1.Y != src1.Y {
			t.Fatalf("err: 'Y' field should be copied.")
		}
	})

	t.Run("earlier valid test", func(t *testing.T) {
		var from *AA
		var to *BB
		ret := evendeep.DeepCopy(from, &to)
		t.Logf("to = %v, ret = %v", to, ret)

		ret = evendeep.DeepCopy(nil, &to)
		t.Logf("to = %v, ret = %v", to, ret)

		ret = evendeep.MakeClone(from)
		t.Logf("ret = %v", ret)

		ret = evendeep.MakeClone(nil)
		t.Logf("ret = %v", ret)
	})

	t.Run("return error test", func(t *testing.T) {
		type AAA struct {
			X1 string `copy:"-"`
			X2 string `copy:",-"`
			Y  int
		}
		type BBB struct {
			X1 string
			X2 string
			Y  int
		}
		src1 := &AAA{X1: "ok", X2: "well", Y: 1}
		tgt1 := &BBB{X1: "no", X2: "longer", Y: -1}

		copier = evendeep.NewForTest()
		err := copier.CopyTo(&src1, tgt1)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
	})
}

func TestIgnoreSourceField(t *testing.T) {
	type AA struct {
		A bool `copy:",ignore"`
		B int
	}
	src := &AA{A: false, B: 9}
	dst := &AA{A: true, B: 19}
	err := evendeep.New().CopyTo(src, dst, evendeep.WithSyncAdvancingOpt)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("got: %v", *dst)
	if !dst.A || dst.B != 9 {
		t.FailNow()
	}
}

func TestOmitEmptySourceField(t *testing.T) {
	type AA struct {
		A int `copy:",omitempty"`
		B int
	}
	src := &AA{A: 0, B: 9}
	dst := &AA{A: 11, B: 19}
	err := evendeep.New().CopyTo(src, dst, evendeep.WithOmitEmptyOpt)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("got: %v", *dst)
	if dst.A != 11 || dst.B != 9 {
		t.FailNow()
	}
}

func TestDeepCopyGenerally(t *testing.T) {
	// defer dbglog.NewCaptureLog(t).Release()

	nn := []int{2, 9, 77, 111, 23, 29}
	var a [2]string
	a[0] = "Hello"
	a[1] = "World"

	x0 := evendeep.X0{}
	x1 := evendeep.X1{
		A: uintptr(unsafe.Pointer(&x0)),
		H: make(chan int, 5),
		M: unsafe.Pointer(&x0),
		// E: []*X0{&x0},
		N: nn[1:3],
		O: a,
		Q: a,
	}

	t.Run("MakeClone()", func(t *testing.T) {
		var ret typ.Any // nolint:gosimple
		// x2 := &X2{N: nn[1:3]}

		ret = evendeep.MakeClone(&x1)
		x1.K = &x0
		testIfBadCopy(t, x1, ret, ret, "MakeClone x1 -> new")
		t.Log("MakeClone is done.")
	})

	t.Run("DeepCopy()", func(t *testing.T) {
		var ret typ.Any
		x2 := &evendeep.X2{N: nn[1:3]}

		ret = evendeep.DeepCopy(&x1, &x2, evendeep.WithIgnoreNames("Shit", "Memo", "Name"))
		testIfBadCopy(t, x1, *x2, ret, "DeepCopy x1 -> x2", true)
	})

	t.Run("NewDeepCopier().CopyTo()", func(t *testing.T) {
		var ret typ.Any
		x2 := &evendeep.X2{N: nn[1:3]}

		ret = evendeep.New().CopyTo(&x1, &x2, evendeep.WithIgnoreNames("Shit", "Memo", "Name"))
		testIfBadCopy(t, x1, *x2, ret, "NewDeepCopier().CopyTo() - DeepCopy x1 -> x2", true)
	})
}

func TestDeepCopy(t *testing.T) {
	type AA struct {
		A bool
		B int32
		C string
	}
	type BB struct {
		A int
		B int16
		C *string
	}

	aa := AA{A: true, B: 16, C: helloString}
	var bb BB
	ret := evendeep.DeepCopy(aa, &bb,
		evendeep.WithIgnoreNames("Shit", "Memo", "Name"))
	t.Logf("ret = %v", ret)
	// ret = &{0 16 &"hello"}
	if *bb.C != helloString {
		t.FailNow()
	}
}

func TestMakeClone(t *testing.T) {
	type AA struct {
		A bool
		B int32
		C string
	}

	aa := AA{A: true, B: 16, C: helloString}
	ret := evendeep.MakeClone(aa)
	aaCopy := ret.(AA)
	t.Logf("ret = %v", aaCopy)
	// ret = {true 16 hello}
}

func TestNew(t *testing.T) {
	type AA struct {
		A bool
		B int32
		C string
	}
	type BB struct {
		A int
		B int16
		C *string
	}

	aa := AA{A: true, B: 16, C: helloString}
	var bb BB
	var ret typ.Any = evendeep.New().CopyTo(aa, &bb,
		evendeep.WithIgnoreNames("Shit", "Memo", "Name"))
	t.Logf("ret = %v", ret)
	// ret = &{0 16 &"hello"}
	if *bb.C != helloString {
		t.FailNow()
	}
}

func TestWithIgnoreNames(t *testing.T) {
	type AA struct {
		A bool
		B int32
		C string
		D string
	}
	type BB struct {
		A int
		B int16
		C *string
	}

	aa := AA{A: true, B: 16, C: helloString, D: worldString}
	var bb BB
	ret := evendeep.DeepCopy(aa, &bb,
		evendeep.WithIgnoreNames("C*"),
		evendeep.WithSyncAdvancing(false),
		evendeep.WithByOrdinalStrategyOpt,
	)
	t.Logf("ret = %v, .C = %v", ret, *bb.C)
	// ret = &{0 16 &"world"}
	if *bb.C != worldString {
		t.FailNow()
	}

	var cc BB
	ret = evendeep.DeepCopy(aa, &cc,
		evendeep.WithIgnoreNames("C*"),
		evendeep.WithSyncAdvancing(true),
		evendeep.WithByOrdinalStrategyOpt,
	)
	t.Logf("ret = %v, .C = %v", ret, *cc.C)
	// ret = &{0 16 &""}
	if *cc.C != "" {
		t.FailNow()
	}
}

func TestPlainCopyFuncField(t *testing.T) {
	type AA struct {
		Fn func()
	}

	t.Run("copy func field", func(t *testing.T) {
		a := AA{func() {
			println("yes")
		}}
		var b AA

		err := evendeep.New().CopyTo(&a, &b, evendeep.WithIgnoreNames("Shit", "Memo", "Name"))
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if b.Fn != nil {
			b.Fn()
		} else {
			t.Fatalf("bad")
		}
	})

	type BB struct {
		fn func()
	}

	t.Run("copy private func field", func(t *testing.T) {
		a := BB{func() {
			println("yes")
		}}
		var b BB

		err := evendeep.New().CopyTo(&a, &b, evendeep.WithIgnoreNames("Shit", "Memo", "Name"))
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if b.fn != nil {
			b.fn()
		} else {
			t.Fatalf("bad")
		}
	})

	type CC struct {
		Jx func(i1, i2 int) (i3 string)
	}

	t.Run("copy private func field (complex)", func(t *testing.T) {
		a := CC{func(i1, i2 int) (i3 string) {
			return fmt.Sprintf("%v+%v", i1, i2)
		}}

		v := reflect.ValueOf(&a)
		vf := v.Elem().Field(0)
		vft := vf.Type()
		t.Logf("out: %v, %v", vft.NumOut(), vft.Out(0))
	})
}
