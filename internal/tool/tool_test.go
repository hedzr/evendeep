package tool_test

import (
	"bytes"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/hedzr/evendeep/internal/tool"
	"github.com/hedzr/evendeep/typ"
	logz "github.com/hedzr/logg/slog"
)

func TestNestedRecovery(t *testing.T) {
	zero := 0

	defer func() {
		if e := recover(); e != nil {
			logz.Error("ERR [TestRec]:", "error", e)
		}
	}()

	func(v int) {
		defer func() {
			if e := recover(); e != nil {
				// logz.Errorf("ERR: %v", e)
				logz.Panic("ERR: ", "error", e)
			}
		}()

		v /= zero
		t.Log(v)
	}(9)
}

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
	ss := []int{8, 9, 7, 9, 3, 5}
	tool.ReverseSlice(ss)
	t.Logf("ss: %v", ss)

	ss = []int{8, 9, 7, 3, 5}
	tool.ReverseAnySlice(ss)
	t.Logf("ss: %v", ss)

	st := []string{"H", "K"}
	tool.ReverseStringSlice(st)
	t.Logf("st: %v", st)
}

func TestInspectStruct(t *testing.T) {
	a4 := prepareDataA4()
	tool.InspectStruct(reflect.ValueOf(&a4))
	t.Log()
}

func TestFindInSlice(t *testing.T) {
	a4 := []int{7, 11, 17}
	v := reflect.ValueOf(a4)
	t.Log(tool.FindInSlice(v, 7, 0))
	t.Log(tool.FindInSlice(v, 1, 0))
}

func TestEqualClassical(t *testing.T) {
	a3 := []int{11, 7, 17}
	a4 := []int{7, 11, 17}
	v := reflect.ValueOf(a4)
	v2 := reflect.ValueOf(a3)
	t.Log(tool.EqualClassical(v, v2))
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

// X0 type for testing
type X0 struct{}

// X1 type for testing
type X1 struct {
	A uintptr
	B map[string]typ.Any
	C bytes.Buffer
	D []string
	E []*X0
	F chan struct{}
	G chan bool
	H chan int
	I func()
	J typ.Any
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
	B map[string]typ.Any
	C bytes.Buffer
	D []string
	E []*X0
	F chan struct{}
	G chan bool
	H chan int
	I func()
	J typ.Any
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
