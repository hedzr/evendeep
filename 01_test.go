package evendeep

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"strconv"
	"testing"
	"time"
	"unsafe"

	"github.com/hedzr/log"
	"gopkg.in/hedzr/errors.v3"

	"github.com/hedzr/evendeep/dbglog"
	"github.com/hedzr/evendeep/internal/cl"
	"github.com/hedzr/evendeep/internal/tool"
)

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
	t.Logf("errors.Is(err, &strconv.NumError{}): %v", errors.Is(err, &strconv.NumError{Err: strconv.ErrRange}))

	var e2 *strconv.NumError
	if errors.As(err, &e2) {
		t.Logf("As() ok, e2 = %v", e2)
	} else {
		t.Logf("As() not ok")
	}
}

func TestSliceLen(t *testing.T) {
	var str []string
	var v = reflect.ValueOf(&str)

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

func TestTm00(t *testing.T) {

	timeFloat := 13572223.479231686
	sec, dec := math.Modf(timeFloat)
	tm := time.Unix(int64(sec), int64(dec*(1e9)))
	t.Logf("tm: %v", tm)

	t.Logf("sec, %v, nano, %v", tm.Unix(), tm.UnixNano())
	t.Logf("f: %v", float64(tm.UnixNano())/1e9)
}

func TestValueValid(t *testing.T) {

	var ival int
	var pival *int
	type A struct {
		ival int //nolint:unused,structcheck //test
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
