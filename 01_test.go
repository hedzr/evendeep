package evendeep

import (
	"fmt"
	"io"
	"math"
	"reflect"
	"runtime"
	"strconv"
	"testing"
	"time"
	"unsafe"

	"gopkg.in/hedzr/errors.v3"

	"github.com/hedzr/evendeep/dbglog"
	"github.com/hedzr/evendeep/internal/cl"
	"github.com/hedzr/evendeep/ref"
	logz "github.com/hedzr/logg/slog"
)

// TestLogNormal _
func TestLogNormal(t *testing.T) {
	// config := logz.NewLoggerConfigWith(true, "logrus", "trace")
	// logger := logrus.NewWithConfig(config)
	logz.Print("hello")
	logz.Info("hello info")
	logz.Warn("hello warn")
	logz.Error("hello error")
	logz.Debug("hello debug")
	logz.Trace("hello trace")

	dbglog.Log("but again")

	t.Log()
}

// TestErrorsTmpl _
func TestErrorsTmpl(t *testing.T) {
	errTmpl := errors.New("expecting %v but got %v")

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
	v := reflect.ValueOf(&str)

	// make value to adopt element's type - in this case string type

	v = v.Elem()

	v = reflect.Append(v, reflect.ValueOf("abc"))
	v = reflect.Append(v, reflect.ValueOf("def"))
	v = reflect.Append(v, reflect.ValueOf("ghi"), reflect.ValueOf("jkl"))

	_, _ = fmt.Println("Our value is a type of :", v.Kind())
	fmt.Printf("len : %v, %v\n", v.Len(), ref.Typfmtv(&v))

	vSlice := v.Slice(0, v.Len())
	vSliceElems := vSlice.Interface()

	_, _ = fmt.Println("With the elements of : ", vSliceElems)

	v = reflect.AppendSlice(v, reflect.ValueOf([]string{"mno", "pqr", "stu"}))

	vSlice = v.Slice(0, v.Len())
	vSliceElems = vSlice.Interface()

	_, _ = fmt.Println("After AppendSlice : ", vSliceElems)
	t.Log()
}

func TestUnexported(t *testing.T) {
	s := struct{ foo int }{42}
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

	_, _ = fmt.Println(s, i, runtime.Version())

	rf = rs.Field(0)
	cl.SetUnexportedField(rf, reflect.ValueOf(123))
	_, _ = fmt.Println(s, i, runtime.Version())
	t.Log()
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

	t.Logf("ival: %v (%v), isvalid/isnil/iszero: %v/%v/%v", ref.Valfmt(&v), ref.Typfmtv(&v), v.IsValid(), ref.IsNil(v), ref.IsZero(v))

	v = reflect.ValueOf(ival)
	t.Logf("ival: %v (%v), isvalid/isnil/iszero: %v/%v/%v", ref.Valfmt(&v), ref.Typfmtv(&v), v.IsValid(), ref.IsNil(v), ref.IsZero(v))

	v = reflect.ValueOf(pival)
	t.Logf("ival: %v (%v), isvalid/isnil/iszero: %v/%v/%v", ref.Valfmt(&v), ref.Typfmtv(&v), v.IsValid(), ref.IsNil(v), ref.IsZero(v))

	v = reflect.ValueOf(aval)
	t.Logf("ival: %v (%v), isvalid/isnil/iszero: %v/%v/%v", ref.Valfmt(&v), ref.Typfmtv(&v), v.IsValid(), ref.IsNil(v), ref.IsZero(v))

	v = reflect.ValueOf(paval)
	t.Logf("ival: %v (%v), isvalid/isnil/iszero: %v/%v/%v", ref.Valfmt(&v), ref.Typfmtv(&v), v.IsValid(), ref.IsNil(v), ref.IsZero(v))

	var b bool
	v = reflect.ValueOf(b)
	t.Logf("ival: %v (%v), isvalid/isnil/iszero: %v/%v/%v", ref.Valfmt(&v), ref.Typfmtv(&v), v.IsValid(), ref.IsNil(v), ref.IsZero(v))

	b = true
	v = reflect.ValueOf(b)
	t.Logf("ival: %v (%v), isvalid/isnil/iszero: %v/%v/%v", ref.Valfmt(&v), ref.Typfmtv(&v), v.IsValid(), ref.IsNil(v), ref.IsZero(v))
}

func TestDbgLog_ChildLogEnabled(t *testing.T) {
	t.Logf("child-enabled: %v", dbglog.ChildLogEnabled())

	caplog := dbglog.CaptureLog(t)
	defer caplog.Release()
	if w, ok := caplog.(io.Writer); ok {
		w.Write([]byte("hello"))
	}

	dbglog.SetLogEnabled()
	t.Logf("child-enabled: %v", dbglog.ChildLogEnabled())

	cl2 := dbglog.NewCaptureLog(t)
	defer cl2.Release()

	dbglog.SetLogDisabled()
	t.Logf("child-enabled: %v", dbglog.ChildLogEnabled())
}
