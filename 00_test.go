package deepcopy

import (
	"bytes"
	"fmt"
	"github.com/hedzr/deepcopy/cl"
	"github.com/hedzr/localtest/deepdiff/d4l3k/messagediff"
	"github.com/hedzr/log"
	"gopkg.in/hedzr/errors.v3"
	"math/big"
	mrand "math/rand"
	"reflect"
	"runtime"
	"sync"
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

//func TestVisibleFields(t *testing.T) {
//	var obj = new(Employee2)
//	typ := reflect.TypeOf(obj)
//	for _, sf := range reflect.VisibleFields(typ.Elem()) {
//		fmt.Println(sf)
//	}
//}

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

//

//

//

type randomizer struct {
	lastErr error
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

var hundred = big.NewInt(100)
var seededRand = mrand.New(mrand.NewSource(time.Now().UTC().UnixNano()))
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
