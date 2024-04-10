package evendeep

import (
	"bytes"
	"time"
	"unsafe"

	"github.com/hedzr/evendeep/typ"
)

// Employee type for testing
type Employee struct {
	Name       string `copy:",std"`
	Birthday   *time.Time
	F11        float32
	F12        float64
	C11        complex64
	C12        complex128
	Feat       []byte
	Sptr       *string
	Nickname   *string
	Age        int64
	FakeAge    int
	EmployeID  int64
	DoubleAge  int32
	SuperRule  string
	Notes      []string
	RetryU     uint8
	TimesU     uint16
	FxReal     uint32
	FxTime     int64
	FxTimeU    uint64
	UxA        uint
	UxB        int
	Retry      int8
	Times      int16
	Born       *int
	BornU      *uint
	flags      []byte //nolint:unused,structcheck //test
	Bool1      bool
	Bool2      bool
	Ro         []int
	Zignored01 typ.Any `copy:"-"`
}

// X0 type for testing
type X0 struct{}

// X1 type for testing
type X1 struct {
	A          uintptr
	B          map[string]typ.Any
	C          bytes.Buffer
	D          []string
	E          []*X0
	F          chan struct{}
	G          chan bool
	H          chan int
	I          func()
	J          typ.Any
	K          *X0
	L          unsafe.Pointer
	M          unsafe.Pointer
	N          []int
	O          [2]string
	P          [2]string
	Q          [2]string
	Zignored01 typ.Any `copy:"-"`
}

// X2 type for testing
type X2 struct {
	A          uintptr
	B          map[string]typ.Any
	C          bytes.Buffer
	D          []string
	E          []*X0
	F          chan struct{}
	G          chan bool
	H          chan int
	I          func()
	J          interface{} //nolint:revive
	K          *X0
	L          unsafe.Pointer
	M          unsafe.Pointer
	N          []int `copy:",slicemerge"`
	O          [2]string
	P          [2]string
	Q          [3]string `copy:",slicecopy"`
	Zignored01 typ.Any   `copy:"-"`
	Zignored02 typ.Any   `copy:"-"`
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
