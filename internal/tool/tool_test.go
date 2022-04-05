package tool_test

import (
	"bytes"
	"github.com/hedzr/evendeep/internal/tool"
	"reflect"
	"testing"
	"time"
	"unsafe"
)

func TestMinInt(t *testing.T) {
	t.Log(tool.MinInt(1, 9))
	t.Log(tool.MinInt(9, 1))
}

func TestContainsStringSlice(t *testing.T) {
	t.Log(tool.Contains([]string{"a", "b", "c"}, "c"))
	t.Log(tool.Contains([]string{"a", "b", "c"}, "z"))

	t.Log(tool.ContainsPartialsOnly([]string{"ac", "could", "ldbe"}, "itcouldbe"))
	t.Log(tool.ContainsPartialsOnly([]string{"a", "b", "c"}, "z"))

	t.Log(tool.PartialContainsShort([]string{"acoludbe", "bcouldbe", "ccouldbe"}, "could"))
	t.Log(tool.PartialContainsShort([]string{"a", "b", "c"}, "z"))

	idx, matchedString, containsBool := tool.PartialContains([]string{"acoludbe", "bcouldbe", "ccouldbe"}, "could")
	t.Logf("%v,%v,%v", idx, matchedString, containsBool)

	idx, matchedString, containsBool = tool.PartialContains([]string{"acoludbe", "bcouldbe", "ccouldbe"}, "byebye")
	t.Logf("%v,%v,%v", idx, matchedString, containsBool)

}

func TestReverseSlice(t *testing.T) {
	var ss = []int{8, 9, 7, 9, 3, 5}
	tool.ReverseAnySlice(ss)
	t.Logf("ss: %v", ss)

	ss = []int{8, 9, 7, 3, 5}
	tool.ReverseAnySlice(ss)
	t.Logf("ss: %v", ss)
}

func TestInspectStruct(t *testing.T) {
	a4 := prepareDataA4()
	tool.InspectStruct(reflect.ValueOf(&a4))
}

func prepareDataA4() *A4 {

	a4 := &A4{
		A3: &A3{
			A2: &A2{
				Name2: "",
				Int2:  0,
				Bool2: false,
				A1: A1{
					Name1: "",
					Int1:  0,
					Bool1: false,
				},
			},
			Name3: "",
			Int3:  0,
			A1: A1{
				Name1: "",
				Int1:  0,
				Bool1: false,
			},
			Bool3: false,
		},
		Int4: 0,
		A1: &A1{
			Name1: "",
			Int1:  0,
			Bool1: false,
		},
	}
	return a4
}

type A1 struct {
	Name1 string
	Int1  int
	Bool1 bool
}
type A2 struct {
	Name2 string
	Int2  int
	Bool2 bool
	A1
}
type A3 struct {
	*A2
	Name3 string
	Int3  int
	A1
	Bool3 bool
}
type A4 struct {
	A3   *A3
	Int4 int
	*A1
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
