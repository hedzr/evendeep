package evendeep

import (
	"fmt"
	"reflect"
	"testing"

	"gopkg.in/hedzr/errors.v3"

	"github.com/hedzr/evendeep/flags"
	"github.com/hedzr/evendeep/flags/cms"
	"github.com/hedzr/evendeep/internal/tool"
)

func TestRegisterInitRoutines(t *testing.T) {
	registerInitRoutines(nil)
	registerInitRoutines(func() {})
	registerLazyInitRoutines(nil)
	registerLazyInitRoutines(func() {})
}

// TestCpChan _
func TestCpChan(t *testing.T) {
	var val = make(chan int, 10)
	vv := reflect.ValueOf(&val)
	vi := reflect.Indirect(vv)
	value := reflect.MakeChan(vi.Type(), vi.Cap())
	t.Logf("%v (len: %v),  vv.len: %v", value.Interface(), value.Cap(), vi.Cap())

	var sval chan string
	var strVal = reflect.ValueOf(&sval)
	indirectStr := reflect.Indirect(strVal)
	svalue := reflect.MakeChan(indirectStr.Type(), 1024)
	t.Logf("Type : [%v] \nCapacity : [%v]", svalue.Kind(), svalue.Cap())

}

// func TestVisibleFields(t *testing.T) {
//	var obj = new(Employee2)
//	typ := reflect.TypeOf(obj)
//	for _, sf := range reflect.VisibleFields(typ.Elem()) {
//		fmt.Println(sf)
//	}
// }

func TestInspectStruct(t *testing.T) {
	a4 := prepareDataA4()
	tool.InspectStruct(reflect.ValueOf(&a4))
}

func TestParamsBasics(t *testing.T) {

	t.Run("basics 1", func(t *testing.T) {
		// p1 := newParams() // nolint:ineffassign
		p1 := newParams(withOwnersSimple(nil, nil))

		p2 := newParams(withOwners(p1.controller, p1, nil, nil, nil, nil))
		t.Logf("p2: %v", p2)
		t.Logf("p1: %v", p1)
		p2.revoke()
		t.Logf("p2: %v", p2)
		t.Logf("p1: %v", p1)
	})

	t.Run("basics 2", func(t *testing.T) {
		// p1 := newParams() // nolint:ineffassign
		p1 := newParams(withOwnersSimple(nil, nil))

		p2 := newParams(withOwners(p1.controller, p1, nil, nil, nil, nil))
		defer p2.revoke()

		a, expects := prepareAFT()

		v := reflect.ValueOf(&a)
		v = tool.Rindirect(v)

		for i := 0; i < v.NumField(); i++ {
			fld := v.Type().Field(i)
			fldTags := parseFieldTags(fld.Tag, "")
			if !p2.isFlagExists(cms.Ignore) {
				t.Logf("%q flags: %v [without ignore]", fld.Tag, fldTags)
			} else {
				t.Logf("%q flags: %v [ignore]", fld.Tag, fldTags)
			}
			testDeepEqual(t.Errorf, fldTags.flags, expects[i])
		}

	})
}

func TestParamsBasics3(t *testing.T) {

	t.Run("basics 3", func(t *testing.T) {
		// p1 := newParams() // nolint:ineffassign
		p1 := newParams(withOwnersSimple(nil, nil))

		p2 := newParams(withOwners(p1.controller, p1, nil, nil, nil, nil))
		defer p2.revoke()

		type AFS1 struct {
			flags     flags.Flags     `copy:",cleareq,must"`                                   //nolint:unused,structcheck //test
			converter *ValueConverter `copy:",ignore"`                                         //nolint:unused,structcheck //test
			wouldbe   int             `copy:",must,keepneq,omitzero,slicecopyappend,mapmerge"` //nolint:unused,structcheck //test
		}
		var a AFS1
		v := reflect.ValueOf(&a)
		v = tool.Rindirect(v)
		sf, _ := v.Type().FieldByName("wouldbe")
		// sf0, _ := v.Type().FieldByName("flags")
		// sf1, _ := v.Type().FieldByName("converter")

		fldTags := parseFieldTags(sf.Tag, "")
		// ft.Parse(sf.Tag)
		// ft.Parse(sf0.Tag) // entering 'continue' branch
		// ft.Parse(sf1.Tag) // entering 'delete' branch

		var z *fieldTags // nolint:gosimple
		z = fldTags

		z.isFlagExists(cms.Flat)

		z.isFlagExists(cms.Ignore)

		z.isFlagExists(cms.SliceCopy)
		p2.isFlagExists(cms.SliceCopy)
		p2.isFlagExists(cms.SliceCopyAppend)
		p2.isFlagExists(cms.SliceMerge)

		p2.isAnyFlagsOK(cms.SliceMerge, cms.Ignore)
		p2.isAllFlagsOK(cms.SliceCopy, cms.Default)

		p2.isGroupedFlagOK(cms.SliceCopy)
		p2.isGroupedFlagOK(cms.SliceCopyAppend)
		p2.isGroupedFlagOK(cms.SliceMerge)

		p2.isGroupedFlagOKDeeply(cms.SliceCopy)
		p2.isGroupedFlagOKDeeply(cms.SliceCopyAppend)
		p2.isGroupedFlagOKDeeply(cms.SliceMerge)

		if p2.depth() != 2 {
			t.Fail()
		}

		var p3 *Params
		p3.isFlagExists(cms.SliceCopy)
		p3.isGroupedFlagOK(cms.SliceCopy)
		p3.isGroupedFlagOK(cms.SliceCopyAppend)
		p3.isGroupedFlagOK(cms.SliceMerge)

		p3.isGroupedFlagOKDeeply(cms.SliceCopy)
		p3.isGroupedFlagOKDeeply(cms.SliceCopyAppend)
		p3.isGroupedFlagOKDeeply(cms.SliceMerge)

		p3.isAnyFlagsOK(cms.SliceMerge, cms.Ignore)
		p3.isAllFlagsOK(cms.SliceCopy, cms.Default)

		var p4 Params
		p4.isFlagExists(cms.SliceCopy)
		p4.isGroupedFlagOK(cms.SliceCopy)
		p4.isGroupedFlagOK(cms.SliceCopyAppend)
		p4.isGroupedFlagOK(cms.SliceMerge)
	})
}

func TestPtrCopy(t *testing.T) {
	type AAA struct {
		P1 *int `copy:",flat"`
	}
	var a, b = 1, 2
	var pa, pb = &AAA{&a}, &AAA{&b}
	Copy(pa, pb)
	t.Logf("pb.P1: %v", *pb.P1)
	if *pb.P1 != a {
		t.Fail()
	}
}

func TestDeferCatchers(t *testing.T) {
	type AAA struct {
		X1 string `copy:"-"`
		X2 string `copy:",-"`
		X3 bool
		Y  int
		Y1 int
	}
	type BBB struct {
		X1 string // backup field to receive the copying field `X3` from source `AAA`
		X2 string // backup field to receive the copying field `Y` from source `AAA`
		X3 string `copy:"-"` // backup field to receive the copying field `Y1` from source `AAA`
		Y  string
		Y1 int
	} // the 'ignore' Tag inside target field cannot block copying on itself

	postCatcher := func(runner func()) {
		defer func() {
			if e1 := recover(); e1 != nil {
				t.Logf(`caught by postCatcher, e: %v`, e1)
			}
		}()
		runner()
	}

	// func TestFieldAccessorT_Normal_Copy(t *testing.T) {
	// 	x1 := x1data()
	// 	x2 := x2data()
	// }

	t.Run("dbgFrontOfStruct", func(t *testing.T) {
		src1 := &AAA{X1: "ok", X2: "well", Y: 1}
		tgt1 := &BBB{X1: "no", X2: "longer", Y: "-1"}

		src, dst := reflect.ValueOf(&src1), reflect.ValueOf(&tgt1)
		svv, dvv := tool.Rdecodesimple(src), tool.Rdecodesimple(dst)
		sf1, df1 := svv.Field(1), dvv.Field(1)

		c := newCopier()

		// p1 := newParams()
		p1 := newParams(
			withOwnersSimple(c, nil),
			withFlags(cms.ByName),
		)

		p2 := newParams(withOwners(p1.controller, p1, &sf1, &df1, nil, nil))
		defer p2.revoke()

		// buildtags.VerboseEnabled = true
		dbgFrontOfStruct(nil, "    ", t.Logf) // just for coverage
		dbgFrontOfStruct(p2, "    ", nil)     // just for coverage
		dbgFrontOfStruct(p2, "    ", t.Logf)
	})

	slicePanic := func() {
		n := []int{5, 7, 4}
		fmt.Println(n[4])
		fmt.Println("normally returned from a")
	}

	t.Run("defer in copyStructInternal", func(t *testing.T) {
		src1 := &AAA{X1: "ok", X2: "well", Y: 1}
		tgt1 := &BBB{X1: "no", X2: "longer", Y: "-1"}

		src, dst := reflect.ValueOf(&src1), reflect.ValueOf(&tgt1)
		svv, dvv := tool.Rdecodesimple(src), tool.Rdecodesimple(dst)
		// sf1, df1 := svv.Field(1), dvv.Field(1)

		c := newCopier()
		for _, rethrow := range []bool{false, true} {
			c.rethrow = rethrow

			// p1 := newParams()
			// p1 = newParams(withOwnersSimple(c, nil))
			//
			// p2 := newParams(withOwners(p1.controller, p1, &sf1, &df1, nil, nil))
			// defer p2.revoke()
			//
			// ec := errors.New("error container")
			postCatcher(func() {
				err := copyStructInternal(c, nil, svv, dvv,
					func(paramsChild *Params, ec errors.Error, i, amount *int, padding string) (err error) {
						paramsChild.nextTargetField()
						slicePanic()
						return
					})
				t.Log(err)
			})
		}
	})

	t.Run("defer rethrew in copyTo", func(t *testing.T) {
		c := newCopier()
		for _, rethrow := range []bool{false, true} {
			c.rethrow = rethrow

			src1 := &AAA{X1: "ok", X2: "well", Y: 1}
			tgt1 := &BBB{X1: "no", X2: "longer", Y: "-1"}

			src, dst := reflect.ValueOf(&src1), reflect.ValueOf(&tgt1)
			svv, dvv := tool.Rdecodesimple(src), tool.Rdecodesimple(dst)
			// sf1, df1 := svv.Field(1), dvv.Field(1)
			postCatcher(func() {
				_ = c.copyToInternal(nil, svv, dvv, func(c *cpController, params *Params, from, to reflect.Value) (err error) {
					slicePanic()
					return
				})
			})
		}
	})

	t.Run("invalid src or dst", func(t *testing.T) {
		c := newCopier()
		c.rethrow = false

		var src1 AAA
		tgt1 := &BBB{X1: "no", X2: "longer", Y: "-1"}

		src, dst := reflect.ValueOf(&src1), reflect.ValueOf(&tgt1)
		svv, dvv := tool.Rdecodesimple(src), tool.Rdecodesimple(dst)
		// sf1, df1 := svv.Field(1), dvv.Field(1)

		t.Logf("src: %v, %v", src.IsValid(), svv.IsValid())
		t.Logf("dst: %v, %v", dst.IsValid(), dvv.IsValid())

		// src is invalid
		var svv1 reflect.Value
		t.Logf("svv1: %v", svv1.IsValid())
		_ = c.copyToInternal(nil, svv1, dvv, func(c *cpController, params *Params, from, to reflect.Value) (err error) {
			slicePanic()
			return
		})

		// src is invalid with params has OmitIfEmpty flag
		params := newParams(withFlags(cms.OmitIfEmpty, cms.ByName))
		_ = c.copyToInternal(params, svv1, dvv, func(c *cpController, params *Params, from, to reflect.Value) (err error) {
			slicePanic()
			return
		})

		// dst is invalid
		var dvv1 reflect.Value
		t.Logf("dvv1: %v", dvv1.IsValid())
		_ = c.copyToInternal(nil, svv, dvv1, func(c *cpController, params *Params, from, to reflect.Value) (err error) {
			slicePanic()
			return
		})

		// both src and dst are valid and params is also valid
		_ = c.copyToInternal(params, svv, dvv, func(c *cpController, params *Params, from, to reflect.Value) (err error) {
			return
		})
	})

	t.Run("copy src to dst with params", func(t *testing.T) {
		lazyInitRoutines()

		c := newCopier()
		c.rethrow = false

		src1 := &AAA{X1: "ok", X2: "well", X3: true, Y: 7, Y1: 13}
		tgt1 := &BBB{X1: "no", X2: "longer", Y: "-1"}

		src, dst := reflect.ValueOf(&src1), reflect.ValueOf(&tgt1)
		svv, dvv := tool.Rindirect(src), tool.Rindirect(dst)
		// sf1, df1 := svv.Field(1), dvv.Field(1)

		root := newParams(withOwners(c, nil, &svv, &dvv, &src, &dst))
		if err := c.copyTo(root, svv, dvv); err != nil {
			t.Fatalf("error: %v", err)
		}
		t.Logf("target BBB is: %+v", tgt1)
	})
}
