package evendeep_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"

	"gopkg.in/hedzr/errors.v3"

	"github.com/hedzr/evendeep"
	"github.com/hedzr/evendeep/diff"
	"github.com/hedzr/evendeep/flags/cms"
	"github.com/hedzr/evendeep/internal/dbglog"
	"github.com/hedzr/evendeep/internal/tool"
	"github.com/hedzr/evendeep/typ"
)

const (
	helloString  = "hello"
	worldString  = "world"
	aHelloString = "Hello"
	aWorldString = "World"
)

func TestDeepCopyForInvalidSourceOrTarget(t *testing.T) {
	invalidObj := func() interface{} {
		var x *evendeep.X0
		return x
	}
	t.Run("invalid source", func(t *testing.T) {
		src := invalidObj()
		tgt := invalidObj()
		evendeep.DeepCopy(src, &tgt)
		t.Logf("tgt: %+v", tgt)
	})
	t.Run("valid ptr to invalid source", func(t *testing.T) {
		src := invalidObj()
		tgt := invalidObj()
		evendeep.DeepCopy(&src, &tgt)
		t.Logf("tgt: %+v", tgt)
	})

	nilmap := func() interface{} {
		var mm []map[string]struct{}
		return mm
	}
	t.Run("nil map", func(t *testing.T) {
		src := nilmap()
		tgt := nilmap()
		evendeep.DeepCopy(src, &tgt)
		t.Logf("tgt: %+v", tgt)
	})
	t.Run("valid ptr to nil map", func(t *testing.T) {
		src := nilmap()
		tgt := nilmap()
		evendeep.DeepCopy(&src, &tgt)
		t.Logf("tgt: %+v", tgt)
	})

	nilslice := func() interface{} {
		var mm []map[string]struct{}
		return mm
	}
	t.Run("nil slice", func(t *testing.T) {
		src := nilslice()
		tgt := nilslice()
		evendeep.DeepCopy(src, &tgt)
		t.Logf("tgt: %+v", tgt)
	})
	t.Run("valid ptr to nil slice", func(t *testing.T) {
		src := nilslice()
		tgt := nilslice()
		evendeep.DeepCopy(&src, &tgt)
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
		float64: &(*s.float64), // nolint:staticcheck
	}
}

func TestCloneableSource(t *testing.T) {
	cloneable := func() *ccs {
		f := evendeep.Randtool.NextFloat64()
		return &ccs{
			string:  evendeep.Randtool.NextStringSimple(13),
			int:     evendeep.Randtool.NextIn(300),
			float64: &f,
		}
	}

	t.Run("invoke Cloneable interface", func(t *testing.T) {
		src := cloneable()
		tgt := cloneable()
		sav := *tgt
		evendeep.DeepCopy(&src, &tgt)
		t.Logf("src: %v, old: %v, new tgt: %v", src, sav, tgt)
		if reflect.DeepEqual(src, tgt) == false {
			var err error
			dif, equal := diff.New(src, tgt)
			if !equal {
				fmt.Println(dif)
				err = errors.New("diff.PrettyDiff identified its not equal:\ndifferent:\n%v", dif)
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
		float64: &(*s.float64), // nolint:staticcheck
	}
}

func TestDeepCopyableSource(t *testing.T) {
	copyable := func() *dcs {
		f := evendeep.Randtool.NextFloat64()
		return &dcs{
			string:  evendeep.Randtool.NextStringSimple(13),
			int:     evendeep.Randtool.NextIn(300),
			float64: &f,
		}
	}

	t.Run("invoke DeepCopyable interface", func(t *testing.T) {
		src := copyable()
		tgt := copyable()
		sav := *tgt
		evendeep.DeepCopy(&src, &tgt)
		t.Logf("src: %v, old: %v, new tgt: %v", src, sav, tgt)
		if reflect.DeepEqual(src, tgt) == false {
			var err error
			dif, equal := diff.New(src, tgt)
			if !equal {
				fmt.Println(dif)
				err = errors.New("diff.PrettyDiff identified its not equal:\ndifferent:\n%v", dif)
			}
			t.Fatalf("not equal. %v", err)
		}
	}) // NewTasskks creates a
}

func TestSimple(t *testing.T) {

	// var dInt = 9
	// var dStr = worldString

	for _, tc := range []evendeep.TestCase{
		evendeep.NewTestCase(
			"primitive - int",
			8, 9, 8,
			nil,
			nil,
		),
		evendeep.NewTestCase(
			"primitive - string",
			helloString, worldString, helloString,
			[]evendeep.Opt{
				evendeep.WithStrategiesReset(cms.Default),
			},
			nil,
		),
		evendeep.NewTestCase(
			"primitive - string slice",
			[]string{helloString, worldString},
			&[]string{"andy"},                   // target needn't addressof
			&[]string{helloString, worldString}, // SliceCopy: copy to target; SliceCopyAppend: append to target; SliceMerge: merge into slice
			[]evendeep.Opt{
				evendeep.WithStrategiesReset(),
			},
			nil,
		),
		evendeep.NewTestCase(
			"primitive - string slice - merge",
			[]string{helloString, helloString, worldString}, // elements in source will be merged into target with uniqueness.
			&[]string{"andy", "andy"},                       // target needn't addressof
			&[]string{"andy", helloString, worldString},     // In merge mode, any dup elems will be removed.
			[]evendeep.Opt{
				evendeep.WithMergeStrategyOpt,
			},
			nil,
		),
		evendeep.NewTestCase(
			"primitive - int slice",
			[]int{7, 99},
			&[]int{5},
			&[]int{7, 99},
			[]evendeep.Opt{
				evendeep.WithStrategiesReset(),
			},
			nil,
		),
		evendeep.NewTestCase(
			"primitive - int slice - merge",
			[]int{7, 99},
			&[]int{5},
			&[]int{5, 7, 99},
			[]evendeep.Opt{
				evendeep.WithStrategies(cms.SliceMerge),
			},
			nil,
		),
		evendeep.NewTestCase(
			"primitive types - int slice - merge for dup",
			[]int{99, 7}, &[]int{125, 99}, &[]int{125, 99, 7},
			[]evendeep.Opt{
				evendeep.WithStrategies(cms.SliceMerge),
			},
			nil,
		),
		// NEED REVIEW: what is copyenh strategy
		// evendeep.NewTestCase(
		//	"primitive types - int slice - copyenh(overwrite and extend)",
		//	[]int{13, 7, 99}, []int{125, 99}, []int{7, 99, 7},
		//	[]evendeep.Opt{
		//		evendeep.WithStrategies(evendeep.SliceCopyOverwrite),
		//	},
		//	nil,
		// ),
	} {
		t.Run(evendeep.RunTestCasesWith(&tc)) // nolint:gosec // G601: Implicit memory aliasing in for loop
	}

}

func TestTypeConvert(t *testing.T) {

	var i9 = 9
	var i5 = 5
	var ui6 = uint(6)
	var i64 int64 = 10
	var f64 = 9.1

	evendeep.RunTestCases(t,
		evendeep.NewTestCase(
			"int -> int64",
			8, i64, int64(8),
			nil,
			nil,
		),
		evendeep.NewTestCase(
			"int64 -> int",
			int64(8), i5, 8,
			nil,
			nil,
		),
		evendeep.NewTestCase(
			"int64 -> uint",
			int64(8), ui6, uint(8),
			nil,
			nil,
		),
		evendeep.NewTestCase(
			"float32 -> float64",
			float32(8.1), f64, float64(8.100000381469727),
			nil,
			nil,
		),
		evendeep.NewTestCase(
			"complex -> complex128",
			complex64(8.1+3i), complex128(9.1), complex128(8.100000381469727+3i),
			nil,
			nil,
		),
		evendeep.NewTestCase(
			"complex -> int - ErrCannotConvertTo test",
			complex64(8.1+3i), &i5, int(8),
			nil,
			func(src, dst, expect interface{}, e error) (err error) {
				if errors.IsDescended(evendeep.ErrCannotConvertTo, e) {
					return
				}
				return e
			},
		),
		evendeep.NewTestCase(
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
	// var ui6 = uint(6)
	// var i64 int64 = 10
	// var f64 float64 = 9.1

	// slice

	var si64 = []int64{9}
	var si = []int{9}
	var sui = []uint{9}
	var sf64 = []float64{9.1}
	var sc128 = []complex128{9.1}

	opts := []evendeep.Opt{
		evendeep.WithStrategies(cms.SliceMerge),
	}

	evendeep.RunTestCases(t,
		evendeep.NewTestCase(
			"[]int -> []int64",
			[]int{8}, &si64, &[]int64{9, 8},
			opts,
			nil,
		),
		evendeep.NewTestCase(
			"int -> []int64",
			7, &si64, &[]int64{9, 8, 7},
			opts,
			nil,
		),
		evendeep.NewTestCase(
			"[]int64 -> []int",
			[]int64{8}, &si, &[]int{9, 8},
			opts,
			nil,
		),
		evendeep.NewTestCase(
			"int64 -> []int",
			int64(7), &si, &[]int{9, 8, 7},
			opts,
			nil,
		),
		evendeep.NewTestCase(
			"[]int64 -> []int (truncate the overflowed input)",
			[]int64{math.MaxInt64}, &si, &[]int{9, 8, 7, cms.MaxInt},
			opts,
			nil,
		),
		evendeep.NewTestCase(
			"int64 -> []uint",
			int64(8), sui, []uint{9, 8},
			opts,
			nil,
		),
		evendeep.NewTestCase(
			"int64 -> *[]uint",
			int64(8), &sui, &[]uint{9, 8},
			opts,
			nil,
		),
		evendeep.NewTestCase(
			"float32 -> []float64",
			float32(8.1), &sf64, &[]float64{9.1, 8.100000381469727},
			opts,
			nil,
		),
		evendeep.NewTestCase(
			"[]float32 -> []float64",
			[]float32{8.1}, &sf64, &[]float64{9.1, 8.100000381469727},
			opts,
			nil,
		),
		evendeep.NewTestCase(
			"complex64 -> []complex128",
			complex64(8.1+3i), &sc128, &[]complex128{9.1, 8.100000381469727 + 3i},
			opts,
			nil,
		),
		evendeep.NewTestCase(
			"[]complex64 -> []complex128",
			[]complex64{8.1 + 3i}, &sc128, &[]complex128{9.1 + 0i, 8.100000381469727 + 3i},
			opts,
			nil,
		),
		evendeep.NewTestCase(
			"complex -> int - ErrCannotConvertTo test",
			complex64(8.1+3i), &i5, int(8),
			opts,
			func(src, dst, expect interface{}, e error) (err error) {
				if errors.IsDescended(evendeep.ErrCannotConvertTo, e) {
					return
				}
				return e
			},
		),
		evendeep.NewTestCase(
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
	// type B struct {
	//	F func(int) (int, error)
	// }
	// b1 := B{F: func(i int) (int, error) { i1 = i * 2; return i1, nil }}

	opts := []evendeep.Opt{
		evendeep.WithPassSourceToTargetFunctionOpt,
	}

	i1 := 0
	b1 := func(i []int) (int, error) { i1 = i[0] * 2; return i1, nil }
	// var e1 error
	b2 := func(i int) (int, error) {
		if i > 0 {
			return 0, errors.BadRequest
		}
		return i, nil
	}

	evendeep.RunTestCases(t,
		evendeep.NewTestCase(
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
		evendeep.NewTestCase(
			"int -> func(int)(int,error)",
			8, &b2, nil,
			opts,
			func(src, dst, expect interface{}, e error) (err error) {
				if !errors.Is(e, errors.BadRequest) {
					err = errors.BadRequest
				}
				return
			},
		),
	)
}

func TestErrorCodeIs(t *testing.T) {
	var err error = errors.BadRequest
	if !errors.Is(err, errors.BadRequest) {
		t.Fatalf("want is")
	}
	err = io.ErrClosedPipe
	if errors.Is(err, errors.BadRequest) {
		t.Fatalf("want not is")
	}
	err = errors.NotFound
	if errors.Is(err, errors.BadRequest) {
		t.Fatalf("want not is (code)")
	}
}

func TestStructStdlib(t *testing.T) {

	// timeZone, _ := time.LoadLocation("America/Phoenix")
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

	for _, tc := range []evendeep.TestCase{
		evendeep.NewTestCase(
			"stdlib - time.Time 1",
			tm1, &tgt, &tm1,
			nil,
			nil,
		),
		evendeep.NewTestCase(
			"stdlib - time.Duration 1",
			dur1, &dur, &dur1,
			nil,
			nil,
		),
		evendeep.NewTestCase(
			"stdlib - bytes.Buffer 1",
			bb1, &bb, &bb1,
			nil,
			nil,
		),
		evendeep.NewTestCase(
			"stdlib - bytes.Buffer 2",
			bb1, &b, &be,
			nil,
			nil,
		),
		evendeep.NewTestCase(
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
		t.Run(evendeep.RunTestCasesWith(&tc)) // nolint:gosec // G601: Implicit memory aliasing in for loop
	}

}

func TestStructSimple(t *testing.T) {

	nn := []int{2, 9, 77, 111, 23, 29}
	var a [2]string
	a[0] = aHelloString
	a[1] = aWorldString
	var a3 = [3]string{aHelloString, aWorldString}

	x0 := evendeep.X0{}
	x1 := evendeep.X1{
		A: uintptr(unsafe.Pointer(&x0)),
		H: make(chan int, 5),
		M: unsafe.Pointer(&x0),
		// E: []*X0{&x0},
		N: nn[1:5],
		O: a,
		Q: a,
	}

	expect1 := &evendeep.X2{
		A: uintptr(unsafe.Pointer(&x0)),
		// D: []string{},
		// E: []*evendeep.X0{},
		H: x1.H,
		K: &x0,
		M: unsafe.Pointer(&x0),
		// E: []*X0{&x0},
		N: nn[1:5],
		O: a,
		Q: a3,
	}
	x2 := evendeep.X2{N: []int{23, 8}}
	expect2 := &evendeep.X2{
		A: uintptr(unsafe.Pointer(&x0)),
		H: x1.H,
		K: &x0,
		M: unsafe.Pointer(&x0),
		// E: []*X0{&x0},
		N: []int{23, 8, 9, 77, 111}, // Note: [23,8] + [9,77,111,23] -> [23,8,9,77,111]
		O: a,
		Q: a3,
	}
	t.Logf("expect.Q: %v", expect1.Q)

	t.Logf("   src: %+v", x1)
	t.Logf("   tgt: %+v", evendeep.X2{N: nn[1:3]})

	evendeep.RunTestCases(t,
		evendeep.NewTestCase(
			"struct - 1",
			x1, &evendeep.X2{N: nn[1:3]}, expect1,
			[]evendeep.Opt{
				evendeep.WithStrategiesReset(),
				// evendeep.WithStrategies(cms.OmitIfEmpty),
				evendeep.WithAutoNewForStructFieldOpt,
			},
			nil,
			// func(src, dst, expect interface{}) (err error) {
			//	dif, equal := diff.New(expect, dst)
			//	if !equal {
			//		fmt.Println(dif)
			//	}
			//	return
			// },
		),
		evendeep.NewTestCase(
			"struct - 2 - merge",
			x1, &x2,
			expect2,
			[]evendeep.Opt{
				evendeep.WithStrategies(cms.SliceMerge),
				evendeep.WithAutoNewForStructFieldOpt,
			},
			nil,
		),
	)

}

func TestStructEmbedded(t *testing.T) {

	timeZone, _ := time.LoadLocation("America/Phoenix")
	tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
	tm2 := time.Date(2003, 9, 1, 23, 59, 59, 3579, timeZone)

	src := evendeep.Employee2{
		Base: evendeep.Base{
			Name:      "Bob",
			Birthday:  &tm,
			Age:       24,
			EmployeID: 7,
		},
		Avatar: "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
		Image:  []byte{95, 27, 43, 66, 0, 21, 210},
		Attr:   &evendeep.Attr{Attrs: []string{helloString, worldString}},
		Valid:  true,
	}

	tgt := evendeep.User{
		Name:      "Frank",
		Birthday:  &tm2,
		Age:       18,
		EmployeID: 9,
		Attr:      &evendeep.Attr{Attrs: []string{"baby"}},
		Deleted:   true,
	}

	expect1 := &evendeep.User{
		Name:      "Bob",
		Birthday:  &tm,
		Age:       24,
		EmployeID: 7,
		Avatar:    "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
		Image:     []byte{95, 27, 43, 66, 0, 21, 210},
		Attr:      &evendeep.Attr{Attrs: []string{"baby", helloString, worldString}},
		Valid:     true,
	}

	evendeep.RunTestCases(t,
		evendeep.NewTestCase(
			"struct - 1",
			src, &tgt,
			expect1,
			[]evendeep.Opt{
				evendeep.WithMergeStrategyOpt,
				evendeep.WithAutoExpandStructOpt,
			},
			nil,
			// func(src, dst, expect interface{}) (err error) {
			//	dif, equal := diff.New(expect, dst)
			//	if !equal {
			//		fmt.Println(dif)
			//	}
			//	return
			// },
		),
	)

}

func TestStructToSliceOrMap(t *testing.T) {

	timeZone, _ := time.LoadLocation("America/Phoenix")
	tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
	// timeZone2, _ := time.LoadLocation("Asia/Chongqing")
	// tm1 := time.Date(2021, 2, 28, 13, 1, 23, 800, timeZone2)
	// tm2 := time.Date(2003, 9, 1, 23, 59, 59, 3579, timeZone)
	// tm3 := time.Date(2015, 1, 29, 19, 31, 37, 77, timeZone2)

	src := evendeep.Employee2{
		Base: evendeep.Base{
			Name:      "Bob",
			Birthday:  &tm,
			Age:       24,
			EmployeID: 7,
		},
		Avatar: "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
		Image:  []byte{95, 27, 43, 66, 0, 21, 210},
		Attr:   &evendeep.Attr{Attrs: []string{helloString, worldString}},
		Valid:  true,
	}

	var slice1 []evendeep.User
	var slice2 []*evendeep.User

	var map1 = make(map[string]interface{})

	expect1 := evendeep.User{
		Name:      "Bob",
		Birthday:  &tm,
		Age:       24,
		EmployeID: 7,
		Avatar:    "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
		Image:     []byte{95, 27, 43, 66, 0, 21, 210},
		Attr:      &evendeep.Attr{Attrs: []string{helloString, worldString}},
		Valid:     true,
	}

	expect3 := map[string]interface{}{
		"Name":      "Bob",
		"Birthday":  tm,
		"Age":       24,
		"EmployeID": int64(7),
		"Avatar":    "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
		"Image":     []byte{95, 27, 43, 66, 0, 21, 210},
		"Attrs":     []string{helloString, worldString},
		"Valid":     true,
		"Deleted":   false,
	}

	t.Run("struct - slice - 1", func(t *testing.T) {
		//
	})

	var str string
	expectJSON := `{
  "Name": "Bob",
  "Birthday": "1999-03-13T05:57:11.000001901-07:00",
  "Age": 24,
  "EmployeID": 7,
  "Avatar": "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet\u0026rs=1",
  "Image": "XxsrQgAV0g==",
  "Attr": {
    "Attrs": [
      "hello",
      "world"
    ]
  },
  "Valid": true,
  "Deleted": false
}`

	evendeep.RunTestCases(t,
		evendeep.NewTestCase(
			"struct -> slice []obj",
			src, &slice1, &[]evendeep.User{expect1},
			[]evendeep.Opt{evendeep.WithMergeStrategyOpt, evendeep.WithAutoExpandStructOpt, evendeep.WithAutoNewForStructFieldOpt},
			nil,
		),

		evendeep.NewTestCase(
			"struct -> string",
			src, &str, &expectJSON,
			[]evendeep.Opt{
				evendeep.WithStringMarshaller(func(v interface{}) ([]byte, error) {
					return json.MarshalIndent(v, "", "  ")
				}),
				evendeep.WithMergeStrategyOpt,
				evendeep.WithAutoExpandStructOpt,
				evendeep.WithAutoNewForStructFieldOpt},
			nil,
		),

		evendeep.NewTestCase(
			"struct -> map[string]Any",
			src, &map1, &expect3,
			[]evendeep.Opt{evendeep.WithMergeStrategyOpt, evendeep.WithAutoExpandStructOpt, evendeep.WithAutoNewForStructFieldOpt},
			nil,
		),

		evendeep.NewTestCase(
			"struct -> slice []obj",
			src, &slice1, &[]evendeep.User{expect1},
			[]evendeep.Opt{evendeep.WithMergeStrategyOpt, evendeep.WithAutoExpandStructOpt, evendeep.WithAutoNewForStructFieldOpt},
			nil,
		),
		evendeep.NewTestCase(
			"struct -> slice []*obj",
			src, &slice2, &[]*evendeep.User{&expect1},
			[]evendeep.Opt{evendeep.WithMergeStrategyOpt, evendeep.WithAutoExpandStructOpt, evendeep.WithAutoNewForStructFieldOpt},
			nil,
		),
	)
}

func TestStructWithSourceExtractor(t *testing.T) {
	type MyValue map[string]typ.Any
	type MyKey string
	const key MyKey = "data-in-sess"
	c := context.WithValue(context.TODO(), key, MyValue{
		"A": 12,
	})

	tgt := struct {
		A int
	}{}

	err := evendeep.New().CopyTo(c, &tgt, evendeep.WithSourceValueExtractor(func(targetName string) typ.Any {
		if m, ok := c.Value(key).(MyValue); ok {
			return m[targetName]
		}
		return nil
	}))

	if tgt.A != 12 || err != nil {
		t.FailNow()
	}
}

func TestStructWithTargetSetter_struct2struct(t *testing.T) {
	type srcS struct {
		A int64
		B bool
		C string
		D float64
	}
	type dstS struct {
		MoA int32
		MoB bool
		MoC string
		MoZ string
	}
	src := &srcS{
		A: 5,
		B: true,
		C: helloString,
	}
	tgt := &dstS{
		MoA: 1,
		MoB: false,
		MoZ: worldString,
	}

	setStructByName := func(s reflect.Value, fld string, val reflect.Value) {
		var f = s.FieldByName(fld)
		if f.IsValid() {
			if val.Type().ConvertibleTo(f.Type()) {
				f.Set(val.Convert(f.Type()))
			} else {
				f.Set(val)
			}
		}
	}
	err := evendeep.New().CopyTo(src, &tgt,
		evendeep.WithTargetValueSetter(func(value *reflect.Value, sourceNames ...string) (err error) {
			if value != nil {
				name := "Mo" + strings.Join(sourceNames, ".")
				setStructByName(reflect.ValueOf(tgt).Elem(), name, *value)
			}
			return // ErrShouldFallback to call the evendeep standard processing
		}),
	)

	if err != nil || tgt.MoA != 5 || !tgt.MoB || tgt.MoC != helloString || tgt.MoZ != worldString {
		t.Errorf("err: %v, tgt: %v", err, tgt)
		t.FailNow()
	} else {
		t.Logf("new map got: %v", tgt)
	}
}

func TestStructWithTargetSetter_struct2map(t *testing.T) {
	type srcS struct {
		A int
		B bool
		C string
	}

	src := &srcS{
		A: 5,
		B: true,
		C: helloString,
	}
	tgt := map[string]typ.Any{
		"Z": worldString,
	}

	err := evendeep.New().CopyTo(src, &tgt,
		evendeep.WithTargetValueSetter(func(value *reflect.Value, sourceNames ...string) (err error) {
			if value != nil {
				name := "Mo" + strings.Join(sourceNames, ".")
				tgt[name] = value.Interface()
			}
			return // ErrShouldFallback to call the evendeep standard processing
		}),
	)

	if err != nil || tgt["MoA"] != 5 || tgt["MoB"] != true || tgt["MoC"] != helloString || tgt["Z"] != worldString {
		t.Errorf("err: %v, tgt: %v", err, tgt)
		t.FailNow()
	} else if _, ok := tgt["A"]; ok {
		t.Errorf("err: key 'A' shouldn't exists, tgt: %v", tgt)
		t.FailNow()
	} else {
		t.Logf("new map got: %v", tgt)
	}
}

func TestStructWithTargetSetter_map2struct(t *testing.T) {
	type dstS struct {
		MoA int32
		MoB bool
		MoC string
		MoZ string
	}
	src := map[string]typ.Any{
		"A": 5,
		"B": true,
		"C": helloString,
	}
	tgt := &dstS{
		MoA: 1,
		MoB: false,
		MoZ: worldString,
	}

	setStructByName := func(s reflect.Value, fldName string, value reflect.Value) {
		var f = s.FieldByName(fldName)
		if f.IsValid() {
			if value.Type().ConvertibleTo(f.Type()) {
				dbglog.Log("struct.%q <- %v", fldName, tool.Valfmt(&value))
				f.Set(value.Convert(f.Type()))
			} else {
				dbglog.Log("struct.%q <- %v", fldName, tool.Valfmt(&value))
				f.Set(value)
			}
		}
	}
	err := evendeep.New().CopyTo(src, &tgt,
		evendeep.WithTargetValueSetter(func(value *reflect.Value, sourceNames ...string) (err error) {
			if value != nil {
				name := "Mo" + strings.Join(sourceNames, ".")
				setStructByName(reflect.ValueOf(tgt).Elem(), name, *value)
				dbglog.Log("struct.%q <- %v", name, tool.Valfmt(value))
			}
			return // ErrShouldFallback to call the evendeep standard processing
		}),
	)

	if err != nil || tgt.MoA != 5 || !tgt.MoB || tgt.MoC != helloString || tgt.MoZ != worldString {
		t.Errorf("err: %v, tgt: %v", err, tgt)
		t.FailNow()
	} else {
		t.Logf("new map got: %v", tgt)
	}
}

func TestStructWithTargetSetter_map2map(t *testing.T) {
	src := map[string]typ.Any{
		"A": 5,
		"B": true,
		"C": helloString,
	}
	tgt := map[string]typ.Any{
		"Z": worldString,
	}

	err := evendeep.New().CopyTo(src, &tgt,
		evendeep.WithTargetValueSetter(func(value *reflect.Value, sourceNames ...string) (err error) {
			if value != nil {
				name := "Mo" + strings.Join(sourceNames, ".")
				tgt[name] = value.Interface()
			}
			return // ErrShouldFallback to call the evendeep standard processing
		}),
	)

	if err != nil || tgt["MoA"] != 5 || tgt["MoB"] != true || tgt["MoC"] != helloString || tgt["Z"] != worldString {
		t.Errorf("err: %v, tgt: %v", err, tgt)
		t.FailNow()
	} else if _, ok := tgt["A"]; ok {
		t.Errorf("err: key 'A' shouldn't exists, tgt: %v", tgt)
		t.FailNow()
	} else {
		t.Logf("new map got: %v", tgt)
	}
}

func TestStructWithSSS(t *testing.T) {
	//
}

type aS struct {
	A int
	b bool
	C string
}

func (s aS) B() bool { return s.b }

func TestStructWithCmsByNameStrategy(t *testing.T) {
	type bS struct {
		Z int
		B bool
		C string
	}

	src := &aS{A: 6, b: true, C: helloString}
	var tgt = bS{Z: 1}

	// use ByName strategy,
	// use copyFunctionResultsToTarget
	err := evendeep.New().CopyTo(src, &tgt, evendeep.WithByNameStrategyOpt)

	if tgt.Z != 1 || !tgt.B || tgt.C != helloString || err != nil {
		t.Fatalf("BAD COPY, tgt: %+v", tgt)
	}
}

func TestStructWithNameConversions(t *testing.T) {
	type srcS struct {
		A int    `copy:"A1"`
		B bool   `copy:"B1,std"`
		C string `copy:"C1,"`
	}

	type dstS struct {
		A1 int
		B1 bool
		C1 string
	}

	src := &srcS{A: 6, B: true, C: helloString}
	var tgt = dstS{A1: 1}

	// use ByName strategy,
	err := evendeep.New().CopyTo(src, &tgt, evendeep.WithByNameStrategyOpt)

	if tgt.A1 != 6 || !tgt.B1 || tgt.C1 != helloString || err != nil {
		t.Fatalf("BAD COPY, tgt: %+v", tgt)
	}
}

func TestStructWithNameConverter(t *testing.T) {
	// TODO enable name converter for each field, and TestStructWithNameConverter()
}

func TestSliceSimple(t *testing.T) {

	tgt := []float32{3.1, 4.5, 9.67}
	itgt := []int{13, 5}

	evendeep.RunTestCases(t,
		evendeep.NewTestCase(
			"slice (float64 -> float32)",
			[]float64{9.123, 5.2}, &tgt, &[]float32{3.1, 4.5, 9.67, 9.123, 5.2},
			[]evendeep.Opt{evendeep.WithMergeStrategyOpt},
			nil,
		),
		evendeep.NewTestCase(
			"slice (uint64 -> int)",
			[]uint64{9, 5}, &itgt, &[]int{13, 5, 9},
			[]evendeep.Opt{evendeep.WithMergeStrategyOpt},
			nil,
		),
	)

}

func TestSliceTypeConvert(t *testing.T) {

	// tgt := []float32{3.1, 4.5, 9.67}
	// itgt := []int{13, 5}
	stgt := []string{"-", "2.718280076980591"}
	stgt2 := []string{"-", "2.718280076980591", "9", "5", "3.1415927410125732"}
	itgt := []int{17}
	itgt2 := []int{13}

	// itgt2 := []int{17}
	// ftgt2 := []float64{17}

	evendeep.RunTestCases(t,
		evendeep.NewTestCase(
			"slice (uint64 -> string)",
			[]uint64{9, 5}, &stgt,
			&[]string{"-", "2.718280076980591", "9", "5"},
			[]evendeep.Opt{evendeep.WithMergeStrategyOpt},
			nil,
		),
		evendeep.NewTestCase(
			"slice (float32 -> string)",
			[]float32{math.Pi, 2.71828}, &stgt,
			// NOTE that stgt kept the new result in last subtest
			&stgt2,
			[]evendeep.Opt{evendeep.WithMergeStrategyOpt},
			nil,
		),
		evendeep.NewTestCase(
			"slice (string(with floats) -> int)",
			stgt2, &itgt,
			&[]int{17, 3, 9, 5},
			[]evendeep.Opt{evendeep.WithMergeStrategyOpt},
			nil,
		),
		evendeep.NewTestCase(
			"slice (string(with floats) -> int)",
			[]string{"-", "9.718280076980591", "9", "5", "3.1415927410125732"},
			&itgt2,
			&[]int{13, 10, 9, 5, 3},
			[]evendeep.Opt{evendeep.WithMergeStrategyOpt},
			nil,
		),

		// needs complexToAnythingConverter

		// evendeep.NewTestCase(
		//	"slice (complex -> float64)",
		//	[]complex64{math.Pi + 3i, 2.71828 + 4.19i},
		//	&ftgt2,
		//	// NOTE that stgt kept the new result in last subtest
		//	&[]float64{2.718280076980591, 17, 3.1415927410125732},
		//	[]evendeep.Opt{evendeep.WithMergeStrategy},
		//	nil,
		// ),
		// evendeep.NewTestCase(
		//	"slice (complex -> int)",
		//	[]complex64{math.Pi + 3i, 2.71828 + 4.19i},
		//	&itgt2,
		//	// NOTE that stgt kept the new result in last subtest
		//	&[]float64{3, 17},
		//	[]evendeep.Opt{evendeep.WithMergeStrategy},
		//	nil,
		// ),
	)

}

func TestMapSimple(t *testing.T) {

	src := map[int64]float64{7: 0, 3: 7.18}
	tgt := map[int]float32{1: 3.1, 2: 4.5, 3: 9.67}
	exp := map[int]float32{1: 3.1, 2: 4.5, 3: 7.18, 7: 0}

	evendeep.RunTestCases(t,
		evendeep.NewTestCase(
			"map (map[int64]float64 -> map[int]float32)",
			src, &tgt, &exp,
			[]evendeep.Opt{evendeep.WithMergeStrategyOpt, evendeep.WithAutoExpandStructOpt},
			nil,
		),
		// evendeep.NewTestCase(
		//	"slice (uint64 -> int)",
		//	[]uint64{9, 5}, &itgt, &[]int{13, 5, 9},
		//	[]evendeep.Opt{evendeep.WithMergeStrategy},
		//	nil,
		// ),
	)

}

func TestMapAndStruct(t *testing.T) {

	timeZone, _ := time.LoadLocation("America/Phoenix")
	timeZone2, _ := time.LoadLocation("Asia/Chongqing")
	tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
	tm2 := time.Date(2003, 9, 1, 23, 59, 59, 3579, timeZone)
	tm1 := time.Date(2021, 2, 28, 13, 1, 23, 800, timeZone2)
	tm3 := time.Date(2015, 1, 29, 19, 31, 37, 77, timeZone2)

	src := evendeep.Employee2{
		Base: evendeep.Base{
			Name:      "Bob",
			Birthday:  &tm,
			Age:       24,
			EmployeID: 7,
		},
		Avatar: "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
		Image:  []byte{95, 27, 43, 66, 0, 21, 210},
		Attr:   &evendeep.Attr{Attrs: []string{helloString, worldString}},
		Valid:  true,
	}

	src3 := evendeep.Employee2{
		Base: evendeep.Base{
			Name:      "Ellen",
			Birthday:  &tm2,
			Age:       55,
			EmployeID: 9,
		},
		Avatar:  "https://placeholder.com/225x168",
		Image:   []byte{181, 130, 23},
		Attr:    &evendeep.Attr{Attrs: []string{"god", "bless"}},
		Valid:   false,
		Deleted: true,
	}

	tgt := evendeep.User{
		Name:      "Mathews",
		Birthday:  &tm3,
		Age:       3,
		EmployeID: 92,
		Attr:      &evendeep.Attr{Attrs: []string{"get"}},
		Deleted:   false,
	}

	tgt2 := evendeep.User{
		Name:      "Frank",
		Birthday:  &tm2,
		Age:       18,
		EmployeID: 9,
		Attr:      &evendeep.Attr{Attrs: []string{"baby"}},
	}

	tgt3 := evendeep.User{
		Name:      "Zeuth",
		Birthday:  &tm1,
		Age:       31,
		EmployeID: 17,
		Image:     []byte{181, 130, 29},
		Attr:      &evendeep.Attr{Attrs: []string{"you"}},
	}

	expect1 := evendeep.User{
		Name:      "Bob",
		Birthday:  &tm,
		Age:       24,
		EmployeID: 7,
		Avatar:    "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
		Image:     []byte{95, 27, 43, 66, 0, 21, 210},
		Attr:      &evendeep.Attr{Attrs: []string{"get", helloString, worldString}},
		Valid:     true,
	}

	expect3 := evendeep.User{
		Name:      "Ellen",
		Birthday:  &tm2,
		Age:       55,
		EmployeID: 9,
		Avatar:    "https://placeholder.com/225x168",
		Image:     []byte{181, 130, 29, 23},
		Attr:      &evendeep.Attr{Attrs: []string{"you", "god", "bless"}},
		Deleted:   true,
	}

	srcmap := map[int64]*evendeep.Employee2{
		7: &src,
		3: &src3,
	}
	tgtmap := map[float32]*evendeep.User{
		7: &tgt,
		2: &tgt2,
		3: &tgt3,
	}
	expmap := map[float32]*evendeep.User{
		7: &expect1,
		2: &tgt2,
		3: &expect3,
	}

	evendeep.RunTestCases(t,
		evendeep.NewTestCase(
			"map (map[int64]Employee2 -> map[int]User)",
			srcmap, &tgtmap, &expmap,
			[]evendeep.Opt{
				evendeep.WithMergeStrategyOpt,
				evendeep.WithAutoExpandStructOpt,
				evendeep.WithAutoNewForStructFieldOpt,
			},
			nil,
		),
		// evendeep.NewTestCase(
		//	"slice (uint64 -> int)",
		//	[]uint64{9, 5}, &itgt, &[]int{13, 5, 9},
		//	[]evendeep.Opt{evendeep.WithMergeStrategy},
		//	nil,
		// ),
	)

}

func TestMapToString(t *testing.T) {

	timeZone, _ := time.LoadLocation("America/Phoenix")
	tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
	// timeZone2, _ := time.LoadLocation("Asia/Chongqing")
	// tm1 := time.Date(2021, 2, 28, 13, 1, 23, 800, timeZone2)
	// tm2 := time.Date(2003, 9, 1, 23, 59, 59, 3579, timeZone)
	// tm3 := time.Date(2015, 1, 29, 19, 31, 37, 77, timeZone2)

	expect2 := evendeep.User{
		Name:      "Bob",
		Birthday:  &tm,
		Age:       24,
		EmployeID: 7,
		Avatar:    "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
		Image:     []byte{95, 27, 43, 66, 0, 21, 210},
		Attr:      &evendeep.Attr{Attrs: []string{helloString, worldString}},
		Valid:     true,
	}

	expect3 := evendeep.Employee2{
		Base: evendeep.Base{
			Name:      "Bob",
			Birthday:  &tm,
			Age:       24,
			EmployeID: 7,
		},
		Avatar: "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
		Image:  []byte{95, 27, 43, 66, 0, 21, 210},
		Attr:   &evendeep.Attr{Attrs: []string{helloString, worldString}},
		Valid:  true,
	}

	var s2 evendeep.User
	var s3 evendeep.Employee2
	var str1 string

	// var map1 = make(map[string]interface{})
	var map1 = map[string]interface{}{
		"Name":      "Bob",
		"Birthday":  tm,
		"Age":       24,
		"EmployeID": int64(7),
		"Avatar":    "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
		"Image":     []byte{95, 27, 43, 66, 0, 21, 210},
		"Attr":      map[string]interface{}{"Attrs": []string{helloString, worldString}},
		"Valid":     true,
		"Deleted":   false,
	}

	expect1 := `{
  "Age": 24,
  "Attr": {
    "Attrs": [
      "hello",
      "world"
    ]
  },
  "Avatar": "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet\u0026rs=1",
  "Birthday": "1999-03-13T05:57:11.000001901-07:00",
  "Deleted": false,
  "EmployeID": 7,
  "Image": "XxsrQgAV0g==",
  "Name": "Bob",
  "Valid": true
}`

	evendeep.RunTestCases(t,
		evendeep.NewTestCase(
			"map -> string [json]",
			map1, &str1, &expect1,
			[]evendeep.Opt{
				evendeep.WithStringMarshaller(func(v interface{}) ([]byte, error) {
					return json.MarshalIndent(v, "", "  ")
				}),
				evendeep.WithMergeStrategyOpt,
				evendeep.WithAutoExpandStructOpt},
			nil,
		),

		evendeep.NewTestCase(
			"map -> struct User",
			map1, &s2, &expect2,
			[]evendeep.Opt{evendeep.WithMergeStrategyOpt, evendeep.WithAutoExpandStructOpt},
			nil,
		),
		evendeep.NewTestCase(
			"map -> struct Employee2",
			map1, &s3, &expect3,
			[]evendeep.Opt{evendeep.WithMergeStrategyOpt, evendeep.WithAutoExpandStructOpt},
			nil,
		),
	)
}

func testIfBadCopy(t *testing.T, src, tgt, result interface{}, title string, notFailed ...interface{}) {

	t.Logf("checking result ...")

	// if diff := deep.Equal(src, tgt); diff == nil {
	//	return
	// } else {
	//	t.Fatalf("testIfBadCopy - BAD COPY (%v):\n  SRC: %+v\n  TGT: %+v\n\n DIFF: \n%v", title, src, tgt, diff)
	// }

	// dd := deepdiff.New()
	// diff, err := dd.Diff(context.Background(), src, tgt)
	// if err != nil {
	//	return
	// }
	// if diff.Len() > 0 {
	//	t.Fatalf("testIfBadCopy - BAD COPY (%v):\n SRC: %+v\n TGT: %+v\n\n DIFF: \n%v", title, src, tgt, diff)
	// } else {
	//	return
	// }

	dif, equal := diff.New(src, tgt)
	if equal {
		return
	}

	fmt.Println(dif)
	err := errors.New("diff.PrettyDiff identified its not equal:\ndifferent:\n%v", dif)

	for _, b := range notFailed {
		if yes, ok := b.(bool); yes && ok {
			return
		}
	}

	t.Fatal(err)

	// if !reflect.DeepEqual(src, tgt) {
	//
	//	var b1, b2 []byte
	//	var err error
	//	if b1, err = json.MarshalIndent(src, "", "  "); err == nil {
	//		if b2, err = json.MarshalIndent(src, "", "  "); err == nil {
	//			if string(b1) == string(b2) {
	//				return
	//			}
	//			t.Logf("testIfBadCopy - src: %v\ntgt: %v\n", string(b1), string(b2))
	//		}
	//	}
	//	if err != nil {
	//		t.Logf("testIfBadCopy - json marshal not ok (just a warning): %v", err)
	//
	//		//if b1, err = yaml.Marshal(src); err == nil {
	//		//	if b2, err = yaml.Marshal(src); err == nil {
	//		//		if string(b1) == string(b2) {
	//		//			return
	//		//		}
	//		//	}
	//		//}
	//
	//		//gob.Register(X1{})
	//		//
	//		//buf1 := new(bytes.Buffer)
	//		//enc1 := gob.NewEncoder(buf1)
	//		//if err = enc1.Encode(&src); err != nil {
	//		//	t.Fatal(err)
	//		//}
	//		//
	//		//buf2 := new(bytes.Buffer)
	//		//enc2 := gob.NewEncoder(buf2)
	//		//if err = enc2.Encode(&tgt); err != nil {
	//		//	t.Fatal(err)
	//		//}
	//		//
	//		//s1, s2 := buf1.String(), buf2.String()
	//		//if s1 == s2 {
	//		//	return
	//		//}
	//	}
	//
	//	for _, b := range notFailed {
	//		if yes, ok := b.(bool); yes && ok {
	//			return
	//		}
	//	}
	//
	//	t.Fatalf("testIfBadCopy - BAD COPY (%v):\n SRC: %+v\n TGT: %+v\n RES: %v", title, src, tgt, result)
	// }
}

func TestExample1(t *testing.T) {
	timeZone, _ := time.LoadLocation("America/Phoenix")
	tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
	var src = evendeep.Employee2{
		Base: evendeep.Base{
			Name:      "Bob",
			Birthday:  &tm,
			Age:       24,
			EmployeID: 7,
		},
		Avatar: "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
		Image:  []byte{95, 27, 43, 66, 0, 21, 210},
		Attr:   &evendeep.Attr{Attrs: []string{helloString, worldString}},
		Valid:  true,
	}
	var dst evendeep.User

	// direct way but no error report: evendeep.DeepCopy(src, &dst)
	c := evendeep.New()
	if err := c.CopyTo(src, &dst); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dst, evendeep.User{
		Name:      "Bob",
		Birthday:  &tm,
		Age:       24,
		EmployeID: 7,
		Avatar:    "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
		Image:     []byte{95, 27, 43, 66, 0, 21, 210},
		Attr:      &evendeep.Attr{Attrs: []string{helloString, worldString}},
		Valid:     true,
	}) {
		t.Fatalf("bad, got %v", dst)
	}
}

type MyType struct {
	I int
}

type MyTypeToStringConverter struct{}

// Uncomment this line if you wanna take a ValueCopier implementation too:
// func (c *MyTypeToStringConverter) CopyTo(ctx *evendeep.ValueConverterContext, source, target reflect.Value) (err error) { return }

func (c *MyTypeToStringConverter) Transform(ctx *evendeep.ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	if source.IsValid() && targetType.Kind() == reflect.String {
		var str string
		if str, err = evendeep.FallbackToBuiltinStringMarshalling(source); err == nil {
			target = reflect.ValueOf(str)
		}
	}
	return
}

func (c *MyTypeToStringConverter) Match(params *evendeep.Params, source, target reflect.Type) (ctx *evendeep.ValueConverterContext, yes bool) {
	sn, sp := source.Name(), source.PkgPath()
	sk, tk := source.Kind(), target.Kind()
	if yes = sk == reflect.Struct && tk == reflect.String &&
		sn == "MyType" && sp == "github.com/hedzr/evendeep_test"; yes {
		ctx = &evendeep.ValueConverterContext{Params: params}
	}
	return
}

func TestExample2(t *testing.T) {
	var myData = MyType{I: 9}
	var dst string
	c := evendeep.NewForTest()
	_ = c.CopyTo(myData, &dst,
		evendeep.WithValueConverters(&MyTypeToStringConverter{}),
		evendeep.WithStringMarshaller(json.Marshal),
	)
	if dst != `{"I":9}` {
		t.Fatalf("bad 1, got %v", dst)
	}

	// a stub call for coverage
	evendeep.RegisterDefaultCopiers()

	var dst1 string
	evendeep.RegisterDefaultConverters(&MyTypeToStringConverter{})
	c = evendeep.NewForTest()
	_ = c.CopyTo(myData, &dst1,
		evendeep.WithStringMarshaller(json.Marshal),
	)
	if dst1 != `{"I":9}` {
		t.Fatalf("bad 2, got %v", dst)
	}

}

func TestExample3(t *testing.T) {
	timeZone, _ := time.LoadLocation("America/Phoenix")
	tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
	var originRec = evendeep.User{
		Name:      "Bob",
		Birthday:  &tm,
		Age:       24,
		EmployeID: 7,
		Avatar:    "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
		Image:     []byte{95, 27, 43, 66, 0, 21, 210},
		Attr:      &evendeep.Attr{Attrs: []string{helloString, worldString}},
		Valid:     true,
	}
	var newRecord evendeep.User
	var t0 = time.Unix(0, 0)
	var expectRec = evendeep.User{Name: "Barbara", Birthday: &t0, Attr: &evendeep.Attr{}}

	_ = evendeep.New().CopyTo(originRec, &newRecord)
	t.Logf("newRecord: %v", newRecord)

	newRecord.Name = "Barbara"
	_ = evendeep.New().CopyTo(originRec, &newRecord, evendeep.WithORMDiffOpt)
	if len(newRecord.Attr.Attrs) == len(expectRec.Attr.Attrs) {
		newRecord.Attr = expectRec.Attr
	}
	if newRecord.Birthday == nil || newRecord.Birthday.Nanosecond() == 0 {
		newRecord.Birthday = &t0
	}
	if !reflect.DeepEqual(newRecord, expectRec) {
		t.Fatalf("bad, got %v | %v", newRecord, newRecord.Birthday.Nanosecond())
	}
	t.Logf("newRecord: %v", newRecord)

}

func TestExample4(t *testing.T) {
	timeZone, _ := time.LoadLocation("America/Phoenix")
	attr := &evendeep.Attr{Attrs: []string{helloString, worldString}}
	tm := time.Date(1999, 3, 13, 5, 57, 11, 1901, timeZone)
	var originRec = evendeep.User{
		Name:      "Bob",
		Birthday:  &tm,
		Age:       24,
		EmployeID: 7,
		Avatar:    "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
		Image:     []byte{95, 27, 43, 66, 0, 21, 210},
		Attr:      attr,
		Valid:     true,
	}
	var dstRecord = new(evendeep.User)
	var t0 = time.Unix(0, 0)
	var emptyRecord = evendeep.User{Name: "Barbara", Birthday: &t0}
	var expectRecord = &evendeep.User{Name: "Barbara", Birthday: &t0,
		Age:       24,
		EmployeID: 7,
		Avatar:    "https://tse4-mm.cn.bing.net/th/id/OIP-C.SAy__OKoxrIqrXWAb7Tj1wHaEC?pid=ImgDet&rs=1",
		Image:     []byte{95, 27, 43, 66, 0, 21, 210},
		Attr:      attr,
		Valid:     true,
	}
	// var expectRecordZero = evendeep.User{Name: "Barbara", Birthday: &t0,
	// 	// Image: []byte{95, 27, 43, 66, 0, 21, 210},
	// 	Attr: &evendeep.Attr{},
	// 	// Attr:  &evendeep.Attr{Attrs: []string{"hello", worldString},
	// 	// Valid: true,
	// }

	evendeep.ResetDefaultCopyController()

	// prepare a hard copy at first
	evendeep.DeepCopy(originRec, &dstRecord)
	t.Logf("dstRecord: %v", dstRecord)
	dbglog.Log("---- dstRecord: %v", dstRecord)
	if !evendeep.DeepEqual(dstRecord, &originRec) {
		t.Fatalf("bad, \n   got: %v\nexpect: %v\n   got.Attr: %v\nexpect.Attr: %v", dstRecord, originRec, dstRecord.Attr, originRec.Attr)
	}

	// now update dstRecord with the non-empty fields.
	evendeep.DeepCopy(emptyRecord, &dstRecord, evendeep.WithOmitEmptyOpt)
	t.Logf("dstRecord (WithOmitEmptyOpt): %v", dstRecord)
	// if !evendeep.DeepEqual(dstRecord, expectRecord) {
	// 	t.Fatalf("bad, \n   got: %v\nexpect: %v\n   got.Attr: %v\nexpect.Attr: %v", dstRecord, expectRecord, dstRecord.Attr, expectRecord.Attr)
	// }
	if delta, equal := evendeep.DeepDiff(dstRecord, expectRecord); !equal {
		t.Fatalf("bad, \n   got: %v\nexpect: %v\n delta:\n%v", dstRecord, expectRecord, delta)
	}

	// evendeep.ResetDefaultCopyController()
	// // now update dstRecord with the non-empty fields.
	// evendeep.DeepCopy(emptyRecord, &dstRecord, evendeep.WithStrategies(cms.OmitIfEmpty))
	// t.Logf("dstRecord (ClearIfInvalid): %v", dstRecord)
	// if delta, equal := evendeep.DeepDiff(dstRecord, expectRecordZero); !equal {
	// 	t.Fatalf("bad, \n   got: %v\nexpect: %v\n delta:\n%v", dstRecord, expectRecordZero, delta)
	// }
}
