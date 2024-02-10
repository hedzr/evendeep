package evendeep

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hedzr/evendeep/dbglog"
	"github.com/hedzr/evendeep/diff"
	"github.com/hedzr/evendeep/flags/cms"
	"github.com/hedzr/evendeep/ref"
	"github.com/hedzr/evendeep/typ"
	logz "github.com/hedzr/logg/slog"

	"gopkg.in/hedzr/errors.v3"
)

func HelperAssertYes(t *testing.T, b bool, expect, got typ.Any) { //nolint:thelper
	if !b {
		t.Fatalf("expecting %v but got %v", expect, got)
	}
}

func TestNewForTest(t *testing.T) {
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

	var a = 1
	var b int
	if err := copier.CopyTo(a, &b); err != nil {
		t.Error("bad")
	}
}

// NewForTest creates a new copier with most common options.
func NewForTest() DeepCopier {
	var copier DeepCopier

	lazyInitRoutines()
	var c1 = newCopier()
	WithStrategies(cms.SliceMerge, cms.MapMerge)(c1)
	if c1.flags.IsAnyFlagsOK(cms.ByOrdinal, cms.SliceMerge, cms.MapMerge, cms.OmitIfEmpty, cms.Default) == false {
		logz.Panic("except flag set with optional values but not matched, 1")
	}
	c1 = newDeepCopier()
	WithStrategies(cms.SliceCopyAppend, cms.MapCopy)(c1)
	if c1.flags.IsAnyFlagsOK(cms.ByOrdinal, cms.SliceCopyAppend, cms.MapCopy, cms.OmitIfEmpty, cms.Default) == false {
		logz.Panic("except flag set with optional values but not matched, 2")
	}
	c1 = newCloner()
	WithStrategies(cms.SliceCopy)(c1)
	if c1.flags.IsAnyFlagsOK(cms.ByOrdinal, cms.SliceCopy, cms.MapCopy, cms.OmitIfEmpty, cms.Default) == false {
		logz.Panic("except flag set with optional values but not matched, 3")
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
type Verifier func(src, dst, expect typ.Any, e error) (err error)

// TestCase _
type TestCase struct {
	Description string  // description of what test is checking
	Src, Dst    typ.Any //
	Expect      typ.Any // expected output
	Opts        []Opt
	Verifier    Verifier
}

// NewTestCases _
func NewTestCases(c ...TestCase) []TestCase {
	return c
}

// NewTestCase _
func NewTestCase(
	description string, // description of what test is checking
	src, dst typ.Any, //
	expect typ.Any, // expected output
	opts []Opt,
	verifier Verifier,
) TestCase {
	return TestCase{
		Description: description, Src: src, Dst: dst,
		Expect: expect, Opts: opts, Verifier: verifier,
	}
}

// ExtrasOpt for TestCase
type ExtrasOpt func(tc *TestCase)

// RunTestCasesWith _
func RunTestCasesWith(tc *TestCase) (desc string, helperSubtest func(t *testing.T)) {
	desc = tc.Description
	helperSubtest = func(t *testing.T) { //nolint:thelper
		c := NewFlatDeepCopier(tc.Opts...)

		err := c.CopyTo(&tc.Src, &tc.Dst)

		verifier := tc.Verifier
		if verifier == nil {
			verifier = DoTestCasesVerifier(t)
		}

		// t.Logf("\nexpect: %+v\n   got: %+v.", tc.expect, tc.dst)
		if err = verifier(tc.Src, tc.Dst, tc.Expect, err); err == nil {
			return
		}

		t.Fatalf("%s FAILED, %+v", tc.Description, err)
	}
	return
}

// DefaultDeepCopyTestRunner _
func DefaultDeepCopyTestRunner(ix int, tc TestCase, opts ...Opt) func(t *testing.T) {
	return func(t *testing.T) {
		c := NewFlatDeepCopier(append(opts, tc.Opts...)...)

		dbglog.Log("- Case %3d: %v", ix, tc.Description)
		err := c.CopyTo(&tc.Src, &tc.Dst)

		verifier := tc.Verifier
		if verifier == nil {
			verifier = DoTestCasesVerifier(t)
		}

		// t.Logf("\nexpect: %+v\n   got: %+v.", tc.expect, tc.dst)
		if err = verifier(tc.Src, tc.Dst, tc.Expect, err); err == nil {
			logz.Print("test passed", "pass-index", ix)
			return
		}

		logz.Error("Error occurs", "index", ix, "error", err)
		t.Fatal("FAILED", "index", ix, "desc", tc.Description, "error", err)
	}
}

// runTestCases _
func runTestCases(t *testing.T, cases ...TestCase) {
	for ix, tc := range cases {
		if !t.Run(fmt.Sprintf("%3d. %s", ix, tc.Description),
			DefaultDeepCopyTestRunner(ix, tc)) {
			break
		}
	}
}

// runTestCasesWithOpts _
func runTestCasesWithOpts(t *testing.T, cases []TestCase, opts ...Opt) {
	for ix, tc := range cases {
		if !t.Run(fmt.Sprintf("%3d. %s", ix, tc.Description), DefaultDeepCopyTestRunner(ix, tc, opts...)) {
			break
		}
	}
}

func DoTestCasesVerifier(t *testing.T) Verifier {
	return func(src, dst, expect typ.Any, e error) (err error) {
		a, b := reflect.ValueOf(dst), reflect.ValueOf(expect)
		aa, _ := ref.Rdecode(a)
		bb, _ := ref.Rdecode(b)
		av, bv := aa.Interface(), bb.Interface()
		logz.Print("mismatched",
			logz.Group("expect", "bv", bv, "typ", ref.Typfmtv(&bb), "t", aa.Type()),
			logz.Group("got", "av", av, "typ", ref.Typfmtv(&aa), "t", bb.Type()),
			"error", e)

		dif, equal := diff.New(expect, dst,
			diff.WithSliceOrderedComparison(false),
			diff.WithStripPointerAtFirst(true),
		)
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
