package deepcopy

import (
	"bytes"
	"fmt"
	"github.com/hedzr/log"
	"reflect"
	"testing"
	"time"
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

	functorLog("but again")
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

type TestCase struct {
	description string      // description of what test is checking
	src, dst    interface{} //
	expect      interface{} // expected output
	opts        []Opt
	verifier    func(src, dst, expect interface{}) (err error)
}

func NewTestCases(c ...TestCase) []TestCase {
	return c
}

func NewTestCase(
	description string, // description of what test is checking
	src, dst interface{}, //
	expect interface{}, // expected output
	opts []Opt,
	verifier func(src, dst, expect interface{}) (err error),
) TestCase {
	return TestCase{
		description, src, dst, expect, opts, verifier,
	}
}

type ExtrasOpt func(tc *TestCase)

func RunTestCases(t *testing.T, cases []TestCase, opts ...Opt) {
	for ix, tc := range cases {
		t.Run(fmt.Sprintf("%3d. %s", ix, tc.description), func(t *testing.T) {

			c := NewFlatDeepCopier(append(opts, tc.opts...)...)

			err := c.CopyTo(&tc.src, &tc.dst)
			if err != nil {
				t.Fatal(err)
			}

			t.Logf("expect %v, got %v.", tc.expect, tc.dst)

			if tc.verifier != nil {
				if err = tc.verifier(tc.src, tc.dst, tc.expect); err != nil {
					return
				}
			} else if reflect.DeepEqual(tc.dst, tc.expect) {
				return
			}

			t.Fatalf("%3d. %s FAILED, %+v", ix, tc.description, err)
		})

	}
}

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

type X0 struct{}

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
