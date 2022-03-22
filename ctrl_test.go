package deepcopy_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hedzr/deepcopy"
	"github.com/hedzr/deepcopy/dbglog"
	"github.com/hedzr/deepcopy/flags/cms"
	"gitlab.com/gopriv/localtest/deepdiff/d4l3k/messagediff"
	"gopkg.in/hedzr/errors.v3"
	"math"
	"reflect"
	"testing"
	"time"
	"unsafe"
)

func TestDeepCopyForInvalidSourceOrTarget(t *testing.T) {

	invalidObj := func() interface{} {
		var x *deepcopy.X0
		return x
	}
	t.Run("invalid source", func(t *testing.T) {
		src := invalidObj()
		tgt := invalidObj()
		deepcopy.DeepCopy(src, &tgt)
		t.Logf("tgt: %+v", tgt)
	})
	t.Run("valid ptr to invalid source", func(t *testing.T) {
		src := invalidObj()
		tgt := invalidObj()
		deepcopy.DeepCopy(&src, &tgt)
		t.Logf("tgt: %+v", tgt)
	})

	nilmap := func() interface{} {
		var mm []map[string]struct{}
		return mm
	}
	t.Run("nil map", func(t *testing.T) {
		src := nilmap()
		tgt := nilmap()
		deepcopy.DeepCopy(src, &tgt)
		t.Logf("tgt: %+v", tgt)
	})
	t.Run("valid ptr to nil map", func(t *testing.T) {
		src := nilmap()
		tgt := nilmap()
		deepcopy.DeepCopy(&src, &tgt)
		t.Logf("tgt: %+v", tgt)
	})

	nilslice := func() interface{} {
		var mm []map[string]struct{}
		return mm
	}
	t.Run("nil slice", func(t *testing.T) {
		src := nilslice()
		tgt := nilslice()
		deepcopy.DeepCopy(src, &tgt)
		t.Logf("tgt: %+v", tgt)
	})
	t.Run("valid ptr to nil slice", func(t *testing.T) {
		src := nilslice()
		tgt := nilslice()
		deepcopy.DeepCopy(&src, &tgt)
		t.Logf("tgt: %+v", tgt)
	})

}

type ccs struct {
	string
	int
	*float64
}

func (s *ccs) Clone() interface{} {
	return &ccs{
		string:  s.string,
		int:     s.int,
		float64: &(*s.float64),
	}
}

func TestCloneableSource(t *testing.T) {
	cloneable := func() *ccs {
		f := deepcopy.Randtool.NextFloat64()
		return &ccs{
			string:  deepcopy.Randtool.NextStringSimple(13),
			int:     deepcopy.Randtool.NextIn(300),
			float64: &f,
		}
	}

	t.Run("invoke Cloneable interface", func(t *testing.T) {
		src := cloneable()
		tgt := cloneable()
		sav := *tgt
		deepcopy.DeepCopy(&src, &tgt)
		t.Logf("src: %v, old: %v, new tgt: %v", src, sav, tgt)
		if reflect.DeepEqual(src, tgt) == false {
			var err error
			diff, equal := messagediff.PrettyDiff(src, tgt)
			if !equal {
				fmt.Println(diff)
				err = errors.New("messagediff.PrettyDiff identified its not equal:\ndifferents:\n%v", diff)
			}
			t.Fatalf("not equal. %v", err)
		}
	})
}

type dcs struct {
	string
	int
	*float64
}

func (s *dcs) DeepCopy() interface{} {
	return &dcs{
		string:  s.string,
		int:     s.int,
		float64: &(*s.float64),
	}
}

func TestDeepCopyableSource(t *testing.T) {
	copyable := func() *dcs {
		f := deepcopy.Randtool.NextFloat64()
		return &dcs{
			string:  deepcopy.Randtool.NextStringSimple(13),
			int:     deepcopy.Randtool.NextIn(300),
			float64: &f,
		}
	}

	t.Run("invoke DeepCopyable interface", func(t *testing.T) {
		src := copyable()
		tgt := copyable()
		sav := *tgt
		deepcopy.DeepCopy(&src, &tgt)
		t.Logf("src: %v, old: %v, new tgt: %v", src, sav, tgt)
		if reflect.DeepEqual(src, tgt) == false {
			var err error
			diff, equal := messagediff.PrettyDiff(src, tgt)
			if !equal {
				fmt.Println(diff)
				err = errors.New("messagediff.PrettyDiff identified its not equal:\ndifferents:\n%v", diff)
			}
			t.Fatalf("not equal. %v", err)
		}
	})
}

func TestSimple(t *testing.T) {

	for _, tc := range []deepcopy.TestCase{
		deepcopy.NewTestCase(
			"primitive - int",
			8, 9, 8,
			nil,
			nil,
		),
		deepcopy.NewTestCase(
			"primitive - string",
			"hello", "world", "hello",
			[]deepcopy.Opt{
				deepcopy.WithStrategiesReset(),
			},
			nil,
		),
		deepcopy.NewTestCase(
			"primitive - string slice",
			[]string{"hello", "world"},
			&[]string{"andy"},           // target needn't addressof
			&[]string{"hello", "world"}, // SliceCopy: copy to target; SliceCopyAppend: append to target; SliceMerge: merge into slice
			[]deepcopy.Opt{
				deepcopy.WithStrategiesReset(),
			},
			nil,
		),
		deepcopy.NewTestCase(
			"primitive - string slice - merge",
			[]string{"hello", "hello", "world"}, // elements in source will be merged into target with uniqueness.
			&[]string{"andy", "andy"},           // target needn't addressof
			&[]string{"andy", "hello", "world"}, // In merge mode, any dup elems will be removed.
			[]deepcopy.Opt{
				deepcopy.WithMergeStrategyOpt,
			},
			nil,
		),
		deepcopy.NewTestCase(
			"primitive - int slice",
			[]int{7, 99},
			&[]int{5},
			&[]int{7, 99},
			[]deepcopy.Opt{
				deepcopy.WithStrategiesReset(),
			},
			nil,
		),
		deepcopy.NewTestCase(
			"primitive - int slice - merge",
			[]int{7, 99},
			&[]int{5},
			&[]int{5, 7, 99},
			[]deepcopy.Opt{
				deepcopy.WithStrategies(cms.SliceMerge),
			},
			nil,
		),
		deepcopy.NewTestCase(
			"primitive types - int slice - merge for dup",
			[]int{99, 7}, &[]int{125, 99}, &[]int{125, 99, 7},
			[]deepcopy.Opt{
				deepcopy.WithStrategies(cms.SliceMerge),
			},
			nil,
		),
		// NEED REVIEW: what is copyenh strategy
		//deepcopy.NewTestCase(
		//	"primitive types - int slice - copyenh(overwrite and extend)",
		//	[]int{13, 7, 99}, []int{125, 99}, []int{7, 99, 7},
		//	[]deepcopy.Opt{
		//		deepcopy.WithStrategies(deepcopy.SliceCopyOverwrite),
		//	},
		//	nil,
		//),
	} {
		t.Run(deepcopy.RunTestCasesWith(&tc))
	}

}

func TestTypeConvert(t *testing.T) {

	var i9 = 9
	var i5 = 5
	var ui6 = uint(6)
	var i64 int64 = 10
	var f64 float64 = 9.1

	deepcopy.RunTestCases(t,
		deepcopy.NewTestCase(
			"int -> int64",
			8, i64, int64(8),
			nil,
			nil,
		),
		deepcopy.NewTestCase(
			"int64 -> int",
			int64(8), i5, 8,
			nil,
			nil,
		),
		deepcopy.NewTestCase(
			"int64 -> uint",
			int64(8), ui6, uint(8),
			nil,
			nil,
		),
		deepcopy.NewTestCase(
			"float32 -> float64",
			float32(8.1), f64, float64(8.100000381469727),
			nil,
			nil,
		),
		deepcopy.NewTestCase(
			"complex -> complex128",
			complex64(8.1+3i), complex128(9.1), complex128(8.100000381469727+3i),
			nil,
			nil,
		),
		deepcopy.NewTestCase(
			"complex -> int - ErrCannotConvertTo test",
			complex64(8.1+3i), &i5, int(8),
			nil,
			func(src, dst, expect interface{}, e error) (err error) {
				if errors.IsDescended(deepcopy.ErrCannotConvertTo, e) {
					return
				}
				return e
			},
		),
		deepcopy.NewTestCase(
			"int -> intptr",
			8, &i9, 8,
			nil,
			func(src, dst, expect interface{}, e error) (err error) {
				if d, ok := dst.(*int); ok && e == nil {
					if *d == src {
						return
					}
				}
				return errors.DataLoss
			},
		),
	)
}

func TestTypeConvert2Slice(t *testing.T) {

	var i9 = 9
	var i5 = 5
	//var ui6 = uint(6)
	//var i64 int64 = 10
	//var f64 float64 = 9.1

	// slice

	var si64 = []int64{9}
	var si = []int{9}
	var sui = []uint{9}
	var sf64 = []float64{9.1}
	var sc128 = []complex128{9.1}

	opts := []deepcopy.Opt{
		deepcopy.WithStrategies(cms.SliceMerge),
	}

	deepcopy.RunTestCases(t,
		deepcopy.NewTestCase(
			"[]int -> []int64",
			[]int{8}, &si64, &[]int64{9, 8},
			opts,
			nil,
		),
		deepcopy.NewTestCase(
			"int -> []int64",
			7, &si64, &[]int64{9, 8, 7},
			opts,
			nil,
		),
		deepcopy.NewTestCase(
			"[]int64 -> []int",
			[]int64{8}, &si, &[]int{9, 8},
			opts,
			nil,
		),
		deepcopy.NewTestCase(
			"int64 -> []int",
			int64(7), &si, &[]int{9, 8, 7},
			opts,
			nil,
		),
		deepcopy.NewTestCase(
			"[]int64 -> []int (truncate the overflowed input)",
			[]int64{math.MaxInt64}, &si, &[]int{9, 8, 7, cms.MaxInt},
			opts,
			nil,
		),
		deepcopy.NewTestCase(
			"int64 -> []uint",
			int64(8), sui, []uint{9, 8},
			opts,
			nil,
		),
		deepcopy.NewTestCase(
			"int64 -> []uint",
			int64(8), &sui, &[]uint{9, 8},
			opts,
			nil,
		),
		deepcopy.NewTestCase(
			"float32 -> []float64",
			float32(8.1), &sf64, &[]float64{9.1, 8.100000381469727},
			opts,
			nil,
		),
		deepcopy.NewTestCase(
			"[]float32 -> []float64",
			[]float32{8.1}, &sf64, &[]float64{9.1, 8.100000381469727},
			opts,
			nil,
		),
		deepcopy.NewTestCase(
			"complex64 -> []complex128",
			complex64(8.1+3i), &sc128, &[]complex128{9.1, 8.100000381469727 + 3i},
			opts,
			nil,
		),
		deepcopy.NewTestCase(
			"[]complex64 -> []complex128",
			[]complex64{8.1 + 3i}, &sc128, &[]complex128{9.1 + 0i, 8.100000381469727 + 3i},
			opts,
			nil,
		),
		deepcopy.NewTestCase(
			"complex -> int - ErrCannotConvertTo test",
			complex64(8.1+3i), &i5, int(8),
			opts,
			func(src, dst, expect interface{}, e error) (err error) {
				if errors.IsDescended(deepcopy.ErrCannotConvertTo, e) {
					return
				}
				return e
			},
		),
		deepcopy.NewTestCase(
			"int -> intptr",
			8, &i9, 8,
			opts,
			func(src, dst, expect interface{}, e error) (err error) {
				if d, ok := dst.(*int); ok && e == nil {
					if *d == src {
						return
					}
				}
				return errors.DataLoss
			},
		),
	)

}

func TestTypeConvert3Func(t *testing.T) {
	//type B struct {
	//	F func(int) (int, error)
	//}
	//b1 := B{F: func(i int) (int, error) { i1 = i * 2; return i1, nil }}

	opts := []deepcopy.Opt{
		deepcopy.WithPassSourceToTargetFunctionOpt,
	}

	i1 := 0
	b1 := func(i []int) (int, error) { i1 = i[0] * 2; return i1, nil }
	//var e1 error
	b2 := func(i int) (int, error) {
		if i > 0 {
			return 0, errors.BadRequest
		}
		return i, nil
	}

	deepcopy.RunTestCases(t,
		deepcopy.NewTestCase(
			"[]int -> func(int)(int,error)",
			[]int{8}, &b1, nil,
			opts,
			func(src, dst, expect interface{}, e error) (err error) {
				if i1 != 16 {
					err = errors.BadRequest
				}
				return
			},
		),
		deepcopy.NewTestCase(
			"int -> func(int)(int,error)",
			8, &b2, nil,
			opts,
			func(src, dst, expect interface{}, e error) (err error) {
				if e != errors.BadRequest {
					err = errors.BadRequest
				}
				return
			},
		),
	)
}

func TestStructStdlib(t *testing.T) {

	//timeZone, _ := time.LoadLocation("America/Phoenix")
	timeZone2, _ := time.LoadLocation("Asia/Chongqing")
	tm1 := time.Date(1979, 1, 29, 13, 3, 49, 19730313, timeZone2)
	var tgt time.Time
	var dur time.Duration
	var dur1 = 13*time.Second + 3*time.Nanosecond
	var bb, bb1 bytes.Buffer
	bb1.WriteString("hellp world")
	var b, be []byte
	be = bb1.Bytes()

	var bbn *bytes.Buffer = nil

	for _, tc := range []deepcopy.TestCase{
		deepcopy.NewTestCase(
			"stdlib - time.Time 1",
			tm1, &tgt, &tm1,
			nil,
			nil,
		),
		deepcopy.NewTestCase(
			"stdlib - time.Duration 1",
			dur1, &dur, &dur1,
			nil,
			nil,
		),
		deepcopy.NewTestCase(
			"stdlib - bytes.Buffer 1",
			bb1, &bb, &bb1,
			nil,
			nil,
		),
		deepcopy.NewTestCase(
			"stdlib - bytes.Buffer 2",
			bb1, &b, &be,
			nil,
			nil,
		),
		deepcopy.NewTestCase(
			"stdlib - bytes.Buffer 2 - target is nil",
			bb1, &bbn, &bb1,
			nil,
			func(src, dst, expect interface{}, e error) (err error) {
				if err = e; e != nil {
					return
				}
				if p, ok := dst.(**bytes.Buffer); ok && *p == nil {
					return
				} else {
					dbglog.Log("p = %v, ok = %v, dst = %v/%v", p, ok, dst, &bbn)
				}
				err = errors.InvalidArgument
				return
			},
		),
	} {
		t.Run(deepcopy.RunTestCasesWith(&tc))
	}

}

func TestStructSimple(t *testing.T) {

	nn := []int{2, 9, 77, 111, 23, 29}
	var a [2]string
	a[0] = "Hello"
	a[1] = "World"
	var a3 = [3]string{"Hello", "World"}

	x0 := deepcopy.X0{}
	x1 := deepcopy.X1{
		A: uintptr(unsafe.Pointer(&x0)),
		H: make(chan int, 5),
		M: unsafe.Pointer(&x0),
		// E: []*X0{&x0},
		N: nn[1:5],
		O: a,
		Q: a,
	}

	expect1 := &deepcopy.X2{
		A: uintptr(unsafe.Pointer(&x0)),
		H: x1.H,
		M: unsafe.Pointer(&x0),
		// E: []*X0{&x0},
		N: nn[1:5],
		O: a,
		Q: a3,
	}
	x2 := deepcopy.X2{N: []int{23, 8}}
	expect2 := &deepcopy.X2{
		A: uintptr(unsafe.Pointer(&x0)),
		H: x1.H,
		M: unsafe.Pointer(&x0),
		// E: []*X0{&x0},
		N: []int{23, 8, 9, 77, 111}, // Note: [23,8] + [9,77,111,23] -> [23,8,9,77,111]
		O: a,
		Q: a3,
	}
	t.Logf("expect.Q: %v", expect1.Q)

	t.Logf("   src: %+v", x1)
	t.Logf("   tgt: %+v", deepcopy.X2{N: nn[1:3]})

	deepcopy.RunTestCases(t,
		deepcopy.NewTestCase(
			"struct - 1",
			x1, &deepcopy.X2{N: nn[1:3]},
			expect1,
			[]deepcopy.Opt{
				deepcopy.WithStrategiesReset(),
			},
			nil,
			//func(src, dst, expect interface{}) (err error) {
			//	diff, equal := messagediff.PrettyDiff(expect, dst)
			//	if !equal {
			//		fmt.Println(diff)
			//	}
			//	return
			//},
		),
		deepcopy.NewTestCase(
			"struct - 2 - merge",
			x1, &x2,
			expect2,
			[]deepcopy.Opt{
				deepcopy.WithStrategies(cms.SliceMerge),
			},
			nil,
		),
	)

}

func TestStructEmbedded(t *testing.T) {

	timeZone, _ := time.LoadLocation("America/Phoenix")
	tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
	tm2 := time.Date(2003, 9, 1, 23, 59, 59, 3579, timeZone)

	src := deepcopy.Employee2{
		Base: deepcopy.Base{
			Name:      "Bob",
			Birthday:  &tm,
			Age:       24,
			EmployeID: 7,
		},
		Avatar: "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
		Image:  []byte{95, 27, 43, 66, 0, 21, 210},
		Attr:   &deepcopy.Attr{Attrs: []string{"hello", "world"}},
		Valid:  true,
	}

	tgt := deepcopy.User{
		Name:      "Frank",
		Birthday:  &tm2,
		Age:       18,
		EmployeID: 9,
		Attr:      &deepcopy.Attr{Attrs: []string{"baby"}},
		Deleted:   true,
	}

	expect1 := &deepcopy.User{
		Name:      "Bob",
		Birthday:  &tm,
		Age:       24,
		EmployeID: 7,
		Avatar:    "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
		Image:     []byte{95, 27, 43, 66, 0, 21, 210},
		Attr:      &deepcopy.Attr{Attrs: []string{"baby", "hello", "world"}},
		Valid:     true,
	}

	deepcopy.RunTestCases(t,
		deepcopy.NewTestCase(
			"struct - 1",
			src, &tgt,
			expect1,
			[]deepcopy.Opt{
				deepcopy.WithMergeStrategyOpt,
				deepcopy.WithAutoExpandStructOpt,
			},
			nil,
			//func(src, dst, expect interface{}) (err error) {
			//	diff, equal := messagediff.PrettyDiff(expect, dst)
			//	if !equal {
			//		fmt.Println(diff)
			//	}
			//	return
			//},
		),
	)

}

func TestStructOthers(t *testing.T) {

}

func TestSliceSimple(t *testing.T) {

	tgt := []float32{3.1, 4.5, 9.67}
	itgt := []int{13, 5}

	deepcopy.RunTestCases(t,
		deepcopy.NewTestCase(
			"slice (float64 -> float32)",
			[]float64{9.123, 5.2}, &tgt, &[]float32{3.1, 4.5, 9.67, 9.123, 5.2},
			[]deepcopy.Opt{deepcopy.WithMergeStrategyOpt},
			nil,
		),
		deepcopy.NewTestCase(
			"slice (uint64 -> int)",
			[]uint64{9, 5}, &itgt, &[]int{13, 5, 9},
			[]deepcopy.Opt{deepcopy.WithMergeStrategyOpt},
			nil,
		),
	)

}

func TestSliceTypeConvert(t *testing.T) {

	//tgt := []float32{3.1, 4.5, 9.67}
	//itgt := []int{13, 5}
	stgt := []string{"-", "2.718280076980591"}
	stgt2 := []string{"-", "2.718280076980591", "9", "5", "3.1415927410125732"}
	itgt := []int{17}

	//itgt2 := []int{17}
	//ftgt2 := []float64{17}

	deepcopy.RunTestCases(t,
		deepcopy.NewTestCase(
			"slice (uint64 -> string)",
			[]uint64{9, 5}, &stgt,
			&[]string{"-", "2.718280076980591", "9", "5"},
			[]deepcopy.Opt{deepcopy.WithMergeStrategyOpt},
			nil,
		),
		deepcopy.NewTestCase(
			"slice (float32 -> string)",
			[]float32{math.Pi, 2.71828}, &stgt,
			// NOTE that stgt kept the new result in last subtest
			&stgt2,
			[]deepcopy.Opt{deepcopy.WithMergeStrategyOpt},
			nil,
		),
		deepcopy.NewTestCase(
			"slice (string(with floats) -> int)",
			stgt2, &itgt,
			&[]int{17, 2, 9, 5, 3},
			[]deepcopy.Opt{deepcopy.WithMergeStrategyOpt},
			nil,
		),

		// needs complexToAnythingConverter

		//deepcopy.NewTestCase(
		//	"slice (complex -> float64)",
		//	[]complex64{math.Pi + 3i, 2.71828 + 4.19i},
		//	&ftgt2,
		//	// NOTE that stgt kept the new result in last subtest
		//	&[]float64{2.718280076980591, 17, 3.1415927410125732},
		//	[]deepcopy.Opt{deepcopy.WithMergeStrategy},
		//	nil,
		//),
		//deepcopy.NewTestCase(
		//	"slice (complex -> int)",
		//	[]complex64{math.Pi + 3i, 2.71828 + 4.19i},
		//	&itgt2,
		//	// NOTE that stgt kept the new result in last subtest
		//	&[]float64{3, 17},
		//	[]deepcopy.Opt{deepcopy.WithMergeStrategy},
		//	nil,
		//),
	)

}

func TestMapSimple(t *testing.T) {

	src := map[int64]float64{7: 0, 3: 7.18}
	tgt := map[int]float32{1: 3.1, 2: 4.5, 3: 9.67}
	exp := map[int]float32{1: 3.1, 2: 4.5, 3: 7.18, 7: 0}

	deepcopy.RunTestCases(t,
		deepcopy.NewTestCase(
			"map (map[int64]float64 -> map[int]float32)",
			src, &tgt, &exp,
			[]deepcopy.Opt{deepcopy.WithMergeStrategyOpt, deepcopy.WithAutoExpandStructOpt},
			nil,
		),
		//deepcopy.NewTestCase(
		//	"slice (uint64 -> int)",
		//	[]uint64{9, 5}, &itgt, &[]int{13, 5, 9},
		//	[]deepcopy.Opt{deepcopy.WithMergeStrategy},
		//	nil,
		//),
	)

}

func TestMapAndStruct(t *testing.T) {

	timeZone, _ := time.LoadLocation("America/Phoenix")
	timeZone2, _ := time.LoadLocation("Asia/Chongqing")
	tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
	tm2 := time.Date(2003, 9, 1, 23, 59, 59, 3579, timeZone)
	tm1 := time.Date(2021, 2, 28, 13, 1, 23, 800, timeZone2)
	tm3 := time.Date(2015, 1, 29, 19, 31, 37, 77, timeZone2)

	src := deepcopy.Employee2{
		Base: deepcopy.Base{
			Name:      "Bob",
			Birthday:  &tm,
			Age:       24,
			EmployeID: 7,
		},
		Avatar: "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
		Image:  []byte{95, 27, 43, 66, 0, 21, 210},
		Attr:   &deepcopy.Attr{Attrs: []string{"hello", "world"}},
		Valid:  true,
	}

	src3 := deepcopy.Employee2{
		Base: deepcopy.Base{
			Name:      "Ellen",
			Birthday:  &tm2,
			Age:       55,
			EmployeID: 9,
		},
		Avatar:  "https://placeholder.com/225x168",
		Image:   []byte{181, 130, 23},
		Attr:    &deepcopy.Attr{Attrs: []string{"god", "bless"}},
		Valid:   false,
		Deleted: true,
	}

	tgt := deepcopy.User{
		Name:      "Mathews",
		Birthday:  &tm3,
		Age:       3,
		EmployeID: 92,
		Attr:      &deepcopy.Attr{Attrs: []string{"get"}},
		Deleted:   false,
	}

	tgt2 := deepcopy.User{
		Name:      "Frank",
		Birthday:  &tm2,
		Age:       18,
		EmployeID: 9,
		Attr:      &deepcopy.Attr{Attrs: []string{"baby"}},
	}

	tgt3 := deepcopy.User{
		Name:      "Zeuth",
		Birthday:  &tm1,
		Age:       31,
		EmployeID: 17,
		Image:     []byte{181, 130, 29},
		Attr:      &deepcopy.Attr{Attrs: []string{"you"}},
	}

	expect1 := deepcopy.User{
		Name:      "Bob",
		Birthday:  &tm,
		Age:       24,
		EmployeID: 7,
		Avatar:    "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
		Image:     []byte{95, 27, 43, 66, 0, 21, 210},
		Attr:      &deepcopy.Attr{Attrs: []string{"get", "hello", "world"}},
		Valid:     true,
	}

	expect3 := deepcopy.User{
		Name:      "Ellen",
		Birthday:  &tm2,
		Age:       55,
		EmployeID: 9,
		Avatar:    "https://placeholder.com/225x168",
		Image:     []byte{181, 130, 29, 23},
		Attr:      &deepcopy.Attr{Attrs: []string{"you", "god", "bless"}},
		Deleted:   true,
	}

	srcmap := map[int64]*deepcopy.Employee2{
		7: &src,
		3: &src3,
	}
	tgtmap := map[float32]*deepcopy.User{
		7: &tgt,
		2: &tgt2,
		3: &tgt3,
	}
	expmap := map[float32]*deepcopy.User{
		7: &expect1,
		2: &tgt2,
		3: &expect3,
	}

	deepcopy.RunTestCases(t,
		deepcopy.NewTestCase(
			"map (map[int64]Employee2 -> map[int]User)",
			srcmap, &tgtmap, &expmap,
			[]deepcopy.Opt{deepcopy.WithMergeStrategyOpt, deepcopy.WithAutoExpandStructOpt},
			nil,
		),
		//deepcopy.NewTestCase(
		//	"slice (uint64 -> int)",
		//	[]uint64{9, 5}, &itgt, &[]int{13, 5, 9},
		//	[]deepcopy.Opt{deepcopy.WithMergeStrategy},
		//	nil,
		//),
	)

}

func testIfBadCopy(t *testing.T, src, tgt, result interface{}, title string, notFailed ...interface{}) {

	t.Logf("checking result ...")

	//if diff := deep.Equal(src, tgt); diff == nil {
	//	return
	//} else {
	//	t.Fatalf("testIfBadCopy - BAD COPY (%v):\n  SRC: %+v\n  TGT: %+v\n\n DIFF: \n%v", title, src, tgt, diff)
	//}

	//dd := deepdiff.New()
	//diff, err := dd.Diff(context.Background(), src, tgt)
	//if err != nil {
	//	return
	//}
	//if diff.Len() > 0 {
	//	t.Fatalf("testIfBadCopy - BAD COPY (%v):\n SRC: %+v\n TGT: %+v\n\n DIFF: \n%v", title, src, tgt, diff)
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
				t.Logf("testIfBadCopy - src: %v\ntgt: %v\n", string(b1), string(b2))
			}
		}
		if err != nil {
			t.Logf("testIfBadCopy - json marshal not ok (just a warning): %v", err)

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

		t.Fatalf("testIfBadCopy - BAD COPY (%v):\n SRC: %+v\n TGT: %+v\n RES: %v", title, src, tgt, result)
	}
}
