package deepcopy

import (
	"bytes"
	"fmt"
	"github.com/hedzr/localtest/deepdiff/d4l3k/messagediff"
	"github.com/hedzr/log"
	"gopkg.in/hedzr/errors.v3"
	"reflect"
	"testing"
	"time"
	"unsafe"
)

// TestNormal _
func TestNormal(t *testing.T) {
	// config := log.NewLoggerConfigWith(true, "logrus", "trace")
	// logger := logrus.NewWithConfig(config)
	log.Printf("hello")
	log.Infof("hello info")
	log.Warnf("hello warn")
	log.Errorf("hello error")
	log.Debugf("hello debug")
	log.Tracef("hello trace")

	functorLog("but again")
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
func RunTestCasesWith(tc *TestCase) (desc string, subtest func(t *testing.T)) {
	desc = tc.description
	subtest = func(t *testing.T) {
		c := NewFlatDeepCopier(tc.opts...)

		err := c.CopyTo(&tc.src, &tc.dst)

		verifier := tc.verifier
		if verifier == nil {
			verifier = runtestcasesverifier(t)
		}

		//t.Logf("\nexpect: %+v\n   got: %+v.", tc.expect, tc.dst)
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
		t.Run(fmt.Sprintf("%3d. %s", ix, tc.description), func(t *testing.T) {

			c := NewFlatDeepCopier(tc.opts...)

			err := c.CopyTo(&tc.src, &tc.dst)

			verifier := tc.verifier
			if verifier == nil {
				verifier = runtestcasesverifier(t)
			}

			//t.Logf("\nexpect: %+v\n   got: %+v.", tc.expect, tc.dst)
			if err = verifier(tc.src, tc.dst, tc.expect, err); err == nil {
				return
			}

			t.Fatalf("%3d. %s FAILED, %+v", ix, tc.description, err)
		})

	}
}

// RunTestCasesWithOpts _
func RunTestCasesWithOpts(t *testing.T, cases []TestCase, opts ...Opt) {
	for ix, tc := range cases {
		t.Run(fmt.Sprintf("%3d. %s", ix, tc.description), func(t *testing.T) {

			c := NewFlatDeepCopier(append(opts, tc.opts...)...)

			err := c.CopyTo(&tc.src, &tc.dst)

			verifier := tc.verifier
			if verifier == nil {
				verifier = runtestcasesverifier(t)
			}

			//t.Logf("\nexpect: %+v\n   got: %+v.", tc.expect, tc.dst)
			if err = verifier(tc.src, tc.dst, tc.expect, err); err == nil {
				return
			}

			t.Fatalf("%3d. %s FAILED, %+v", ix, tc.description, err)
		})

	}
}

func runtestcasesverifier(t *testing.T) Verifier {
	return func(src, dst, expect interface{}, e error) (err error) {
		a, b := reflect.ValueOf(dst), reflect.ValueOf(expect)
		aa, _ := rdecode(a)
		bb, _ := rdecode(b)
		av, bv := aa.Interface(), bb.Interface()
		t.Logf("got.type: %v, expect.type: %v", aa.Type(), bb.Type())
		t.Logf("\nexpect: %+v (%v)\n   got: %+v (%v)",
			bv, typfmtv(&bb), av, typfmtv(&aa))

		diff, equal := messagediff.PrettyDiff(expect, dst)
		if !equal {
			fmt.Println(diff)
			err = errors.New("messagediff.PrettyDiff identified its not equal:\ndifferents:\n%v", diff)
		}

		if reflect.DeepEqual(av, bv) {
			return
		}

		err = errors.New("reflect.DeepEqual test its not equal")
		return
	}
}

// Employee type for testing
type Employee struct {
	Name      string
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
	flags     []byte
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

type Attr struct {
	Attrs []string `copy:",slicemerge"`
}

type Base struct {
	Name      string
	Birthday  *time.Time
	Age       int
	EmployeID int64
}

type Employee2 struct {
	Base
	Avatar  string
	Image   []byte
	Attr    *Attr
	Valid   bool
	Deleted bool
}

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
