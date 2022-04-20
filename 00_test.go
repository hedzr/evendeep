package evendeep

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	mrand "math/rand"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/hedzr/evendeep/diff"
	"github.com/hedzr/evendeep/flags"
	"github.com/hedzr/evendeep/flags/cms"
	"github.com/hedzr/evendeep/internal/cl"
	"github.com/hedzr/evendeep/internal/dbglog"
	"github.com/hedzr/evendeep/internal/tool"
	"github.com/hedzr/log"

	"gopkg.in/hedzr/errors.v3"
)

func TestRegisterInitRoutines(t *testing.T) {
	registerInitRoutines(nil)
	registerInitRoutines(func() {})
	registerLazyInitRoutines(nil)
	registerLazyInitRoutines(func() {})
}

// TestLogNormal _
func TestLogNormal(t *testing.T) {
	// config := log.NewLoggerConfigWith(true, "logrus", "trace")
	// logger := logrus.NewWithConfig(config)
	log.Printf("hello")
	log.Infof("hello info")
	log.Warnf("hello warn")
	log.Errorf("hello error")
	log.Debugf("hello debug")
	log.Tracef("hello trace")

	dbglog.Log("but again")
}

// TestErrorsTmpl _
func TestErrorsTmpl(t *testing.T) {
	var errTmpl = errors.New("expecting %v but got %v")

	var err error
	err = errTmpl.FormatWith("789", "123")
	t.Logf("The error is expected: %v", err)
	err = errTmpl.FormatWith(true, false)
	t.Logf("The error is expected: %v", err)
}

// TestErrorsIs _
func TestErrorsIs(t *testing.T) {
	_, err := strconv.ParseFloat("hello", 64)
	t.Logf("err = %+v", err)

	// e1:=errors2.Unwrap(err)
	// t.Logf("e1 = %+v", e1)

	t.Logf("errors.Is(err, strconv.ErrSyntax): %v", errors.Is(err, strconv.ErrSyntax))
	t.Logf("errors.Is(err, &strconv.NumError{}): %v", errors.Is(err, &strconv.NumError{}))

	var e2 *strconv.NumError
	if errors.As(err, &e2) {
		t.Logf("As() ok, e2 = %v", e2)
	} else {
		t.Logf("As() not ok")
	}
}

func TestSliceLen(t *testing.T) {
	var str []string
	var v reflect.Value = reflect.ValueOf(&str)

	// make value to adopt element's type - in this case string type

	v = v.Elem()

	v = reflect.Append(v, reflect.ValueOf("abc"))
	v = reflect.Append(v, reflect.ValueOf("def"))
	v = reflect.Append(v, reflect.ValueOf("ghi"), reflect.ValueOf("jkl"))

	fmt.Println("Our value is a type of :", v.Kind())
	fmt.Printf("len : %v, %v\n", v.Len(), tool.Typfmtv(&v))

	vSlice := v.Slice(0, v.Len())
	vSliceElems := vSlice.Interface()

	fmt.Println("With the elements of : ", vSliceElems)

	v = reflect.AppendSlice(v, reflect.ValueOf([]string{"mno", "pqr", "stu"}))

	vSlice = v.Slice(0, v.Len())
	vSliceElems = vSlice.Interface()

	fmt.Println("After AppendSlice : ", vSliceElems)
}

// TestCpChan _
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

// func TestVisibleFields(t *testing.T) {
//	var obj = new(Employee2)
//	typ := reflect.TypeOf(obj)
//	for _, sf := range reflect.VisibleFields(typ.Elem()) {
//		fmt.Println(sf)
//	}
// }

func TestUnexported(t *testing.T) {
	var s = struct{ foo int }{42}
	var i int

	rs := reflect.ValueOf(&s).Elem() // s, but writable
	rf := rs.Field(0)                // s.foo
	ri := reflect.ValueOf(&i).Elem() // i, but writeable

	// These both fail with "reflect.Value.Set using value obtained using unexported field":
	// ri.Set(rf)
	// rf.Set(ri)

	// Cheat:
	rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()

	// Now these both work:
	ri.Set(rf)
	i = 9
	rf.Set(ri)

	fmt.Println(s, i, runtime.Version())

	rf = rs.Field(0)
	cl.SetUnexportedField(rf, reflect.ValueOf(123))
	fmt.Println(s, i, runtime.Version())
}

func TestPtrOf(t *testing.T) {

	var i = 100
	v := reflect.ValueOf(&i)
	vind := tool.Rindirect(v)
	vp := tool.PtrOf(vind)
	t.Logf("ptr of i: %v, &i: %v", vp.Interface(), &i)
	vp.Elem().SetInt(99)
	t.Logf("i: %v", i)

}

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
			flags     flags.Flags     `copy:",cleareq,must"`                                   //nolint:unused,structcheck
			converter *ValueConverter `copy:",ignore"`                                         //nolint:unused,structcheck
			wouldbe   int             `copy:",must,keepneq,omitzero,slicecopyappend,mapmerge"` //nolint:unused,structcheck
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

func TestDeferCatchers(t *testing.T) {

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

	t.Run("dbgFrontOfStruct", func(t *testing.T) {

		src1 := &AAA{X1: "ok", X2: "well", Y: 1}
		tgt1 := &BBB{X1: "no", X2: "longer", Y: -1}

		src, dst := reflect.ValueOf(&src1), reflect.ValueOf(&tgt1)
		svv, dvv := tool.Rdecodesimple(src), tool.Rdecodesimple(dst)
		sf1, df1 := svv.Field(1), dvv.Field(1)

		c := newCopier()

		// p1 := newParams()
		p1 := newParams(withOwnersSimple(c, nil))

		p2 := newParams(withOwners(p1.controller, p1, &sf1, &df1, nil, nil))
		defer p2.revoke()

		dbgFrontOfStruct(p2, "    ", func(msg string, args ...interface{}) { dbglog.Log(msg, args...) })
	})

	slicePanic := func() {
		n := []int{5, 7, 4}
		fmt.Println(n[4])
		fmt.Println("normally returned from a")
	}

	t.Run("defer in copyStructInternal", func(t *testing.T) {

		src1 := &AAA{X1: "ok", X2: "well", Y: 1}
		tgt1 := &BBB{X1: "no", X2: "longer", Y: -1}

		src, dst := reflect.ValueOf(&src1), reflect.ValueOf(&tgt1)
		svv, dvv := tool.Rdecodesimple(src), tool.Rdecodesimple(dst)
		// sf1, df1 := svv.Field(1), dvv.Field(1)

		c := newCopier()
		c.rethrow = false

		// p1 := newParams()
		// p1 = newParams(withOwnersSimple(c, nil))
		//
		// p2 := newParams(withOwners(p1.controller, p1, &sf1, &df1, nil, nil))
		// defer p2.revoke()
		//
		// ec := errors.New("error container")

		err := copyStructInternal(c, nil, svv, dvv,
			func(paramsChild *Params, ec errors.Error, i, amount *int, padding string) (err error) {
				paramsChild.nextTargetField()
				slicePanic()
				return
			})
		t.Log(err)

	})

	t.Run("defer rethrew in copyTo", func(t *testing.T) {

		c := newCopier()
		c.rethrow = false

		src1 := &AAA{X1: "ok", X2: "well", Y: 1}
		tgt1 := &BBB{X1: "no", X2: "longer", Y: -1}

		src, dst := reflect.ValueOf(&src1), reflect.ValueOf(&tgt1)
		svv, dvv := tool.Rdecodesimple(src), tool.Rdecodesimple(dst)
		// sf1, df1 := svv.Field(1), dvv.Field(1)

		_ = c.copyToInternal(nil, svv, dvv, func(c *cpController, params *Params, from, to reflect.Value) (err error) {

			slicePanic()
			return
		})

	})

}

//

func TestTm00(t *testing.T) {

	timeFloat := 13572223.479231686
	sec, dec := math.Modf(timeFloat)
	tm := time.Unix(int64(sec), int64(dec*(1e9)))
	t.Logf("tm: %v", tm)

	t.Logf("sec, %v, nano, %v", tm.Unix(), tm.UnixNano())
	t.Logf("f: %v", float64(tm.UnixNano())/1e9)
}

//

func TestValueValid(t *testing.T) {

	var ival int
	var pival *int
	type A struct {
		ival int //nolint:unused,structcheck
	}
	var aval A
	var paval *A

	var v reflect.Value

	t.Logf("ival: %v (%v), isvalid/isnil/iszero: %v/%v/%v", tool.Valfmt(&v), tool.Typfmtv(&v), v.IsValid(), tool.IsNil(v), tool.IsZero(v))

	v = reflect.ValueOf(ival)
	t.Logf("ival: %v (%v), isvalid/isnil/iszero: %v/%v/%v", tool.Valfmt(&v), tool.Typfmtv(&v), v.IsValid(), tool.IsNil(v), tool.IsZero(v))

	v = reflect.ValueOf(pival)
	t.Logf("ival: %v (%v), isvalid/isnil/iszero: %v/%v/%v", tool.Valfmt(&v), tool.Typfmtv(&v), v.IsValid(), tool.IsNil(v), tool.IsZero(v))

	v = reflect.ValueOf(aval)
	t.Logf("ival: %v (%v), isvalid/isnil/iszero: %v/%v/%v", tool.Valfmt(&v), tool.Typfmtv(&v), v.IsValid(), tool.IsNil(v), tool.IsZero(v))

	v = reflect.ValueOf(paval)
	t.Logf("ival: %v (%v), isvalid/isnil/iszero: %v/%v/%v", tool.Valfmt(&v), tool.Typfmtv(&v), v.IsValid(), tool.IsNil(v), tool.IsZero(v))

	var b bool
	v = reflect.ValueOf(b)
	t.Logf("ival: %v (%v), isvalid/isnil/iszero: %v/%v/%v", tool.Valfmt(&v), tool.Typfmtv(&v), v.IsValid(), tool.IsNil(v), tool.IsZero(v))

	b = true
	v = reflect.ValueOf(b)
	t.Logf("ival: %v (%v), isvalid/isnil/iszero: %v/%v/%v", tool.Valfmt(&v), tool.Typfmtv(&v), v.IsValid(), tool.IsNil(v), tool.IsZero(v))
}

//

//

//

//

func HelperAssertYes(t *testing.T, b bool, expect, got interface{}) { //nolint:thelper
	if !b {
		t.Fatalf("expecting %v but got %v", expect, got)
	}
}

// NewForTest creates a new copier with most common options.
func NewForTest() DeepCopier {
	// nolint:staticcheck
	copier := New( // nolint:ineffassign
		WithValueConverters(&toDurationConverter{}),
		WithValueCopiers(&toDurationConverter{}),

		WithCloneStyle(),
		WithCopyStyle(),

		WithCopyStrategyOpt,
		WithMergeStrategyOpt,
		WithStrategiesReset(),
		WithStrategies(cms.SliceMerge, cms.MapMerge),

		WithAutoExpandStructOpt,
		WithAutoNewForStructFieldOpt,
		WithCopyUnexportedFieldOpt,
		WithCopyFunctionResultToTargetOpt,
		WithPassSourceToTargetFunctionOpt,

		WithSyncAdvancingOpt,

		WithTryApplyConverterAtFirstOpt,
		WithByNameStrategyOpt,
		WithByOrdinalStrategyOpt,

		WithIgnoreNamesReset(),
		WithIgnoreNames("Bugs*", "Test*"),

		WithStructTagName("copy"),

		WithoutPanic(),

		WithStringMarshaller(nil),
	)

	lazyInitRoutines()
	var c1 = newCopier()
	WithStrategies(cms.SliceMerge, cms.MapMerge)(c1)
	if c1.flags.IsAnyFlagsOK(cms.ByOrdinal, cms.SliceMerge, cms.MapMerge, cms.OmitIfEmpty, cms.Default) == false {
		log.Panicf("except flag set with optional values but not matched, 1")
	}
	c1 = newDeepCopier()
	WithStrategies(cms.SliceCopyAppend, cms.MapCopy)(c1)
	if c1.flags.IsAnyFlagsOK(cms.ByOrdinal, cms.SliceCopyAppend, cms.MapCopy, cms.OmitIfEmpty, cms.Default) == false {
		log.Panicf("except flag set with optional values but not matched, 2")
	}
	c1 = newCloner()
	WithStrategies(cms.SliceCopy)(c1)
	if c1.flags.IsAnyFlagsOK(cms.ByOrdinal, cms.SliceCopy, cms.MapCopy, cms.OmitIfEmpty, cms.Default) == false {
		log.Panicf("except flag set with optional values but not matched, 3")
	}

	copier = NewFlatDeepCopier(
		WithStrategies(cms.SliceMerge, cms.MapMerge),
		WithValueConverters(&toDurationConverter{}),
		WithValueCopiers(&toDurationConverter{}),
		WithCloneStyle(),
		WithCopyStyle(),
		WithAutoExpandStructOpt,
		WithCopyStrategyOpt,
		WithStrategiesReset(),
		WithMergeStrategyOpt,
		WithCopyUnexportedField(true),
		WithCopyFunctionResultToTarget(true),
		WithIgnoreNamesReset(),
		WithIgnoreNames("Bugs*", "Test*"),
	)

	return copier
}

//

//

//

//

//

//

// Verifier _
type Verifier func(src, dst, expect interface{}, e error) (err error)

// TestCase _
type TestCase struct {
	description string      // description of what test is checking
	src, dst    interface{} //
	expect      interface{} // expected output
	opts        []Opt
	verifier    Verifier
}

// NewTestCases _
func NewTestCases(c ...TestCase) []TestCase {
	return c
}

// NewTestCase _
func NewTestCase(
	description string, // description of what test is checking
	src, dst interface{}, //
	expect interface{}, // expected output
	opts []Opt,
	verifier Verifier,
) TestCase {
	return TestCase{
		description, src, dst, expect, opts, verifier,
	}
}

// ExtrasOpt for TestCase
type ExtrasOpt func(tc *TestCase)

// RunTestCasesWith _
func RunTestCasesWith(tc *TestCase) (desc string, helperSubtest func(t *testing.T)) {
	desc = tc.description
	helperSubtest = func(t *testing.T) { //nolint:thelper
		c := NewFlatDeepCopier(tc.opts...)

		err := c.CopyTo(&tc.src, &tc.dst)

		verifier := tc.verifier
		if verifier == nil {
			verifier = runtestcasesverifier(t)
		}

		// t.Logf("\nexpect: %+v\n   got: %+v.", tc.expect, tc.dst)
		if err = verifier(tc.src, tc.dst, tc.expect, err); err == nil {
			return
		}

		t.Fatalf("%s FAILED, %+v", tc.description, err)
	}
	return
}

// RunTestCases _
func RunTestCases(t *testing.T, cases ...TestCase) {
	for ix, tc := range cases {
		if !t.Run(fmt.Sprintf("%3d. %s", ix, tc.description), func(t *testing.T) {

			c := NewFlatDeepCopier(tc.opts...)

			dbglog.Log("- Case %3d: %v", ix, tc.description)
			err := c.CopyTo(&tc.src, &tc.dst)

			verifier := tc.verifier
			if verifier == nil {
				verifier = runtestcasesverifier(t)
			}

			// t.Logf("\nexpect: %+v\n   got: %+v.", tc.expect, tc.dst)
			if err = verifier(tc.src, tc.dst, tc.expect, err); err == nil {
				log.Printf("%3d. test passed", ix)
				return
			}

			log.Errorf("%3d. Error: %v", ix, err)
			t.Fatalf("%3d. %s FAILED, %+v", ix, tc.description, err)
		}) {
			break
		}

	}
}

// RunTestCasesWithOpts _
func RunTestCasesWithOpts(t *testing.T, cases []TestCase, opts ...Opt) {
	for ix, tc := range cases {
		if !t.Run(fmt.Sprintf("%3d. %s", ix, tc.description), func(t *testing.T) {

			c := NewFlatDeepCopier(append(opts, tc.opts...)...)

			err := c.CopyTo(&tc.src, &tc.dst)

			verifier := tc.verifier
			if verifier == nil {
				verifier = runtestcasesverifier(t)
			}

			// t.Logf("\nexpect: %+v\n   got: %+v.", tc.expect, tc.dst)
			if err = verifier(tc.src, tc.dst, tc.expect, err); err == nil {
				log.Printf("%3d. test passed", ix)
				return
			}

			log.Errorf("%3d. Error: %v", ix, err)
			t.Fatalf("%3d. %s FAILED, %+v", ix, tc.description, err)
		}) {
			break
		}

	}
}

func runtestcasesverifier(t *testing.T) Verifier {
	return func(src, dst, expect interface{}, e error) (err error) {
		a, b := reflect.ValueOf(dst), reflect.ValueOf(expect)
		aa, _ := tool.Rdecode(a)
		bb, _ := tool.Rdecode(b)
		av, bv := aa.Interface(), bb.Interface()
		log.Printf("\nexpect: %+v (%v | %v)\n   got: %+v (%v | %v)\n   err: %v",
			bv, tool.Typfmtv(&bb), aa.Type(), av, tool.Typfmtv(&aa), bb.Type(), e)

		dif, equal := diff.New(expect, dst, diff.WithSliceOrderedComparison(false))
		if !equal {
			fmt.Println(dif)
			err = errors.New("diff.PrettyDiff identified its not equal:\ndifferent:\n%v", dif).WithErrors(e)
			return
		}

		// if !reflect.DeepEqual(av, bv) {
		//	err = errors.New("reflect.DeepEqual identified its not equal")
		// }
		err = e
		return
	}
}

//

//

//

type randomizer struct {
	lastErr error //nolint:unused,structcheck
}

// var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
const (
	// Alphabets gets the a to z and A to Z
	Alphabets = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// Digits gets 0 to 9
	Digits = "0123456789"
	// AlphabetNumerics gets Alphabets and Digits
	AlphabetNumerics = Alphabets + Digits
	// Symbols gets the ascii symbols
	Symbols = "~!@#$%^&*()-_+={}[]\\|<,>.?/\"';:`"
	// ASCII gets the ascii characters
	ASCII = AlphabetNumerics + Symbols
)

var hundred = big.NewInt(100)                                            //nolint:unused,deadcode,varcheck
var seededRand = mrand.New(mrand.NewSource(time.Now().UTC().UnixNano())) //nolint:gosec //G404: Use of weak random number generator (math/rand instead of crypto/rand)
var mu sync.Mutex
var Randtool = &randomizer{}

func (r *randomizer) Next() int {
	mu.Lock()
	defer mu.Unlock()
	return seededRand.Int()
}
func (r *randomizer) NextIn(max int) int {
	mu.Lock()
	defer mu.Unlock()
	return seededRand.Intn(max)
}
func (r *randomizer) inRange(min, max int) int {
	mu.Lock()
	defer mu.Unlock()
	return seededRand.Intn(max-min) + min
}
func (r *randomizer) NextInRange(min, max int) int { return r.inRange(min, max) }
func (r *randomizer) NextInt63n(n int64) int64 {
	mu.Lock()
	defer mu.Unlock()
	return seededRand.Int63n(n)
}
func (r *randomizer) NextIntn(n int) int {
	mu.Lock()
	defer mu.Unlock()
	return seededRand.Intn(n)
}
func (r *randomizer) NextFloat64() float64 {
	mu.Lock()
	defer mu.Unlock()
	return seededRand.Float64()
}

// NextStringSimple returns a random string with specified length 'n', just in A..Z
func (r *randomizer) NextStringSimple(n int) string {
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		n := seededRand.Intn(90-65) + 65
		b[i] = byte(n) // 'a' .. 'z'
	}
	return string(b)
}

//

//

//

// Employee type for testing
type Employee struct {
	Name      string `copy:",std"`
	Birthday  *time.Time
	F11       float32
	F12       float64
	C11       complex64
	C12       complex128
	Feat      []byte
	Sptr      *string
	Nickname  *string
	Age       int64
	FakeAge   int
	EmployeID int64
	DoubleAge int32
	SuperRule string
	Notes     []string
	RetryU    uint8
	TimesU    uint16
	FxReal    uint32
	FxTime    int64
	FxTimeU   uint64
	UxA       uint
	UxB       int
	Retry     int8
	Times     int16
	Born      *int
	BornU     *uint
	flags     []byte //nolint:unused,structcheck
	Bool1     bool
	Bool2     bool
	Ro        []int
}

// X0 type for testing
type X0 struct{}

// X1 type for testing
type X1 struct {
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

// Attr _
type Attr struct {
	Attrs []string `copy:",slicemerge"`
}

// Base _
type Base struct {
	Name      string
	Birthday  *time.Time
	Age       int
	EmployeID int64
}

// Employee2 _
type Employee2 struct {
	Base
	Avatar  string
	Image   []byte
	Attr    *Attr
	Valid   bool
	Deleted bool
}

// User _
type User struct {
	Name      string
	Birthday  *time.Time
	Age       int
	EmployeID int64
	Avatar    string
	Image     []byte
	Attr      *Attr
	Valid     bool
	Deleted   bool
}
