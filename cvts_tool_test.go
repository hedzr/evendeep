package evendeep

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/hedzr/evendeep/dbglog"
	"github.com/hedzr/evendeep/flags/cms"
	"github.com/hedzr/evendeep/internal/cl"
	"github.com/hedzr/evendeep/ref"
	"github.com/hedzr/evendeep/typ"
)

type sample struct {
	a int
	b string
}

func TestUintptrAndUnsafePointer(t *testing.T) {
	s := &sample{a: 1, b: "test"}

	// Getting the address of field b in struct s
	p := unsafe.Pointer(uintptr(unsafe.Pointer(s)) + unsafe.Offsetof(s.b))

	// Typecasting it to a string pointer and printing the value of it
	_, _ = fmt.Println(*(*string)(p))

	u := uintptr(unsafe.Pointer(s))
	us := fmt.Sprintf("%v", u)
	t.Logf("us = 0x%v", us)
	v := reflect.ValueOf(us)
	ret := rToUIntegerHex(v, reflect.TypeOf(uintptr(unsafe.Pointer(s))))
	t.Logf("ret.type: %v, %v / 0x%x", ret.Type(), ret.Interface(), ret.Interface())

	// t.Logf("ret.type: %v, %v", ret.Type(), ret.Pointer())
}

func TestGetPointerAsUintptr(t *testing.T) {
	s := &sample{a: 1, b: "test"}

	v := reflect.ValueOf(s)
	u := getPointerAsUintptr(v)
	_, _ = fmt.Println(u)
	t.Log()
}

func TestForBool(t *testing.T) {
	var b1, b2 bool
	b2 = true

	v1, v2 := reflect.ValueOf(b1), reflect.ValueOf(b2)
	if rForBool(v1).Interface() != "false" {
		t.Fail()
	}

	if rForBool(v2).Interface() != "true" {
		t.Fail()
	}
}

func TestToBool(t *testing.T) {
	for _, vi := range []typ.Any{
		false,
		0,
		uint(0),
		math.Float64frombits(0),
		complex(math.Float64frombits(0), math.Float64frombits(0)),
		[0]int{},
		[1]int{0},
		(func())(nil), // nolint:gocritic // cannot remove paran
		struct{}{},
		"f",
		"false",
		"off",
		"no",
		"famale",
	} {
		v1 := reflect.ValueOf(vi)
		v2, _ := ref.Rdecode(v1)
		if rToBool(v2).Interface() != false { //nolint:revive
			t.Fatalf("for %v (%v) toBool failed", vi, ref.Typfmtv(&v2))
		}
	}

	for _, vi := range []typ.Any{
		true,
		-1,
		uint(1),
		math.Float64frombits(1),
		complex(math.Float64frombits(1), math.Float64frombits(0)),
		[1]int{1},
		[1]int{3},
		map[int]int{1: 1},
		struct{ v int }{1},
		"1", "t", "y", "m",
		"true",
		"on",
		"yes",
		"male",
	} {
		v1 := reflect.ValueOf(vi)
		v2, _ := ref.Rdecode(v1)
		if rToBool(v2).Interface() != true { //nolint:revive
			t.Fatalf("for %v (%v) toBool failed", vi, ref.Typfmtv(&v2))
		}
	}
}

func TestForInteger(t *testing.T) {
	for _, src := range []typ.Any{
		13579,
		uint(13579),
	} {
		v1 := reflect.ValueOf(src)
		v1 = ref.Rdecodesimple(v1)
		if rForInteger(v1).Interface() != "13579" {
			t.Fail()
		}
	}

	var z typ.Any
	v1 := reflect.ValueOf(z)
	v1 = ref.Rdecodesimple(v1)
	if x := rForInteger(v1).Interface(); x != "0" {
		t.Fatalf("failed, x = %v", x)
	}

	z = "bug" //nolint:goconst
	v1 = reflect.ValueOf(z)
	v1 = ref.Rdecodesimple(v1)
	if x := rForInteger(v1).Interface(); x != "0" {
		t.Fatalf("failed, x = %v", x)
	}
}

func TestToInteger(t *testing.T) {
	for _, dt := range []reflect.Type{
		reflect.TypeOf((*int)(nil)).Elem(),
		reflect.TypeOf((*int64)(nil)).Elem(),
		reflect.TypeOf((*int32)(nil)).Elem(),
		reflect.TypeOf((*int16)(nil)).Elem(),
		reflect.TypeOf((*int8)(nil)).Elem(),
	} {
		for vv, ii := range map[string]int64{
			"123":  123,
			"-123": -123,
			"8.75": 9,
			"8.49": 8,
		} {
			v := reflect.ValueOf(vv)
			ret, err := rToInteger(v, dt)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if ret.Int() != ii {
				t.Fatalf("expect %v but got %v", ii, ret.Int())
			}
		}
	}
}

func TestForUInteger(t *testing.T) {
	for _, src := range []typ.Any{
		13579,
		uint(13579),
	} {
		v1 := reflect.ValueOf(src)
		v1 = ref.Rdecodesimple(v1)
		if rForUInteger(v1).Interface() != "13579" {
			t.Fail()
		}
	}

	var z typ.Any
	v1 := reflect.ValueOf(z)
	v1 = ref.Rdecodesimple(v1)
	if x := rForUInteger(v1).Interface(); x != "0" {
		t.Fatalf("failed, x = %v", x)
	}

	z = "bug"
	v1 = reflect.ValueOf(z)
	v1 = ref.Rdecodesimple(v1)
	if x := rForUInteger(v1).Interface(); x != "0" {
		t.Fatalf("failed, x = %v", x)
	}
}

func TestToUInteger(t *testing.T) {
	for _, dt := range []reflect.Type{
		reflect.TypeOf((*uint)(nil)).Elem(),
		reflect.TypeOf((*uint64)(nil)).Elem(),
		reflect.TypeOf((*uint32)(nil)).Elem(),
		reflect.TypeOf((*uint16)(nil)).Elem(),
		reflect.TypeOf((*uint8)(nil)).Elem(),
	} {
		for vv, ii := range map[string]uint64{
			"123":  123,
			"9":    9,
			"8.75": 9,
			"8.49": 8,
		} {
			v := reflect.ValueOf(vv)
			ret, err := rToUInteger(v, dt)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if ret.Uint() != ii {
				t.Fatalf("expect %v but got %v", ii, ret.Int())
			}
		}
	}
}

func TestForUIntegerHex(t *testing.T) {
	for _, src := range []uint64{
		0x3e67,
		uint64(0x3e67),
	} {
		// v1 := reflect.ValueOf(src)
		// v1 = rdecodesimple(v1)
		if x := rForUIntegerHex(uintptr(src)).Interface(); x != "0x3e67" {
			t.Fatalf("expect %v but got %v", "0x3e67", x)
		}
	}

	var z typ.Any
	v1 := reflect.ValueOf(z)
	v1 = ref.Rdecodesimple(v1)
	if x := rForUInteger(v1).Interface(); x != "0" {
		t.Fatalf("failed, x = %v", x)
	}

	z = "bug"
	v1 = reflect.ValueOf(z)
	v1 = ref.Rdecodesimple(v1)
	if x := rForUInteger(v1).Interface(); x != "0" {
		t.Fatalf("failed, x = %v", x)
	}

	z = "0x3e59"
	v1 = reflect.ValueOf(z)
	v1 = ref.Rdecodesimple(v1)
	uintptrType := reflect.TypeOf(uintptr(0))
	if x := uintptr(rToUIntegerHex(v1, uintptrType).Uint()); x != uintptr(0x3e59) {
		t.Fatalf("failed, x = %v", x)
	}

	vz := "0x3e59"
	v1 = reflect.ValueOf(vz)
	ptrType := reflect.TypeOf(&vz)
	t.Logf("v1.kind: %v, ptrType.kind: %v", v1.Kind(), ptrType.Kind())
	if x := uintptr(rToUIntegerHex(v1, ptrType).Uint()); x == 0 {
		t.Fatalf("failed, x = %v", x)
	}
}

func TestForFloat(t *testing.T) {
	for _, src := range []typ.Any{
		13579,
		uint(13579),
	} {
		v1 := reflect.ValueOf(src)
		v1 = ref.Rdecodesimple(v1)
		if rForFloat(v1).Interface() != "13579" {
			t.Fail()
		}
	}

	var z typ.Any
	v1 := reflect.ValueOf(z)
	v1 = ref.Rdecodesimple(v1)
	if x := rForFloat(v1).Interface(); x != "0" {
		t.Fatalf("failed, x = %v", x)
	}

	z = "bug"
	v1 = reflect.ValueOf(z)
	v1 = ref.Rdecodesimple(v1)
	if x := rForFloat(v1).Interface(); x != "0" {
		t.Fatalf("failed, x = %v", x)
	}
}

func TestToFloat(t *testing.T) {
	for _, dt := range []reflect.Type{
		reflect.TypeOf((*float64)(nil)).Elem(),
		// reflect.TypeOf((*float32)(nil)).Elem(),
	} {
		for vv, ii := range map[string]float64{
			"123":                                  123,
			"-123":                                 -123,
			"8.75":                                 8.75,
			strconv.FormatUint(math.MaxUint64, 10): 1.8446744073709552e+19,
			"(8.1+3.5i)":                           8.1,
		} {
			v := reflect.ValueOf(vv)
			ret, err := rToFloat(v, dt)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if ret.Float() != ii {
				t.Fatalf("expect %v but got %v", ii, ret.Float())
			}
		}
	}
}

func TestForComplex(t *testing.T) {
	for src, exp := range map[typ.Any]string{
		13579:        "(13579+0i)",
		uint(13579):  "(13579+0i)",
		1.316:        "(1.316+0i)",
		8.5 + 7.13i:  "(8.5+7.13i)",
		-8.5 - 7.13i: "(-8.5-7.13i)",
	} {
		v1 := reflect.ValueOf(src)
		v1 = ref.Rdecodesimple(v1)
		if x := rForComplex(v1).Interface(); x != exp {
			t.Fatalf("failed, x = %v, expect = %v", x, exp)
		}
	}

	var z typ.Any
	v1 := reflect.ValueOf(z)
	v1 = ref.Rdecodesimple(v1)
	if x := rForComplex(v1).Interface(); x != "(0+0i)" {
		t.Fatalf("failed, x = %v", x)
	}

	z = "bug"
	v1 = reflect.ValueOf(z)
	v1 = ref.Rdecodesimple(v1)
	if x := rForComplex(v1).Interface(); x != "(0+0i)" {
		t.Fatalf("failed, x = %v", x)
	}
}

func TestToComplex(t *testing.T) {
	for _, dt := range []reflect.Type{
		// reflect.TypeOf((*complex64)(nil)).Elem(),
		reflect.TypeOf((*complex128)(nil)).Elem(),
	} {
		for vv, ii := range map[string]complex128{
			"123+1i":     123 + 1i,
			"-123-7i":    -123 - 7i,
			"8.75-3.13i": 8.75 - 3.13i,
			// strconv.FormatUint(math.MaxUint64, 10): 1.8446744073709552e+19,
			"(8.1+3.5i)": 8.1 + 3.5i,
		} {
			v := reflect.ValueOf(vv)
			ret, err := rToComplex(v, dt)
			if err != nil {
				t.Fatalf("err: %v, for src: %q", err, vv)
			}
			if ret.Complex() != ii {
				t.Fatalf("expect %v but got %v", ii, ret.Complex())
			}
		}
	}
}

func TestBytesBufferConverter_Transform(t *testing.T) {
	var bbc fromBytesBufferConverter
	tgtType := reflect.TypeOf((*bytes.Buffer)(nil)).Elem()
	var bb bytes.Buffer
	_, _ = bb.WriteString("hello")
	src := reflect.ValueOf(bb)
	tgt, err := bbc.Transform(nil, src, tgtType)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if x, ok := tgt.Interface().(bytes.Buffer); !ok {
		t.Fatalf("unexpect target value type: %v", tgt.Type())
	} else if x.String() != "hello" { //nolint:goconst
		t.Fatalf("convert failed, want 'hello' but got %q", x.String())
	}
}

func TestToStringConverter_Transform(t *testing.T) { //nolint:revive
	var bbc toStringConverter
	tgtType := reflect.TypeOf((*string)(nil)).Elem()

	var bb bytes.Buffer
	_, _ = bb.WriteString("hello")
	var sb strings.Builder
	_, _ = sb.WriteString("hello")

	for sv, exp := range map[typ.Any]string{
		"sss":           "sss",
		true:            "true",
		false:           "false",
		123:             "123",
		-123:            "-123",
		uint(123):       "123",
		8.79:            "8.79",
		uintptr(0x3e7c): "0x3e7c",
		9 + 3i:          "(9+3i)",
		&bb:             "hello",
		&sb:             "hello",
		nil:             "",
	} {
		svv := reflect.ValueOf(sv)
		// src := rdecodesimple(svv)
		tgt, err := bbc.Transform(nil, svv, tgtType)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if x, ok := tgt.Interface().(string); !ok {
			t.Fatalf("unexpect target value type: %v", tgt.Type())
		} else if x != exp {
			t.Fatalf("convert failed, want %q but got %q", exp, x)
		}

		tgtstr := "1"
		tgt = reflect.ValueOf(&tgtstr).Elem()
		dbglog.Log("target/; %v %v", ref.Valfmt(&tgt), ref.Typfmtv(&tgt))
		err = bbc.CopyTo(nil, svv, tgt)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if x, ok := tgt.Interface().(string); !ok {
			t.Fatalf("unexpect target value type: %v", tgt.Type())
		} else if x != exp {
			t.Fatalf("convert failed, want %q but got %q", exp, x)
		}
		t.Logf("   tgtstr = %v", tgtstr)
	}

	//

	type sss struct {
		string
	}
	sss1 := sss{"hello"}
	exp := "{hello}"

	svv := reflect.ValueOf(sss1)
	// src := rdecodesimple(svv)
	tgt, err := bbc.Transform(nil, svv, tgtType)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if x, ok := tgt.Interface().(string); !ok {
		t.Fatalf("unexpect target value type: %v", tgt.Type())
	} else if x != exp {
		t.Fatalf("convert failed, want %q but got %q", exp, x)
	}
}

var tgtTypes = map[reflect.Kind]reflect.Type{
	reflect.String:     reflect.TypeOf((*string)(nil)).Elem(),
	reflect.Bool:       reflect.TypeOf((*bool)(nil)).Elem(),
	reflect.Uint:       reflect.TypeOf((*uint)(nil)).Elem(),
	reflect.Int:        reflect.TypeOf((*int)(nil)).Elem(),
	reflect.Float64:    reflect.TypeOf((*float64)(nil)).Elem(),
	reflect.Complex128: reflect.TypeOf((*complex128)(nil)).Elem(),
	reflect.Ptr:        reflect.TypeOf((*int)(nil)).Elem(),
	reflect.Uintptr:    reflect.TypeOf((*uintptr)(nil)).Elem(),
}

func TestFromStringConverter_Transform(t *testing.T) { //nolint:revive
	var bbc fromStringConverter

	for src, tgtm := range map[string]map[reflect.Kind]typ.Any{
		"sss":    {reflect.String: "sss"},
		"true":   {reflect.Bool: true},
		"false":  {reflect.Bool: false},
		"123":    {reflect.Uint: uint(123)},
		"-123":   {reflect.Int: -123},
		"8.79":   {reflect.Float64: 8.79},
		"(3+4i)": {reflect.Complex128: 3 + 4i},
		"0x3e4a": {reflect.Uintptr: uintptr(0x3e4a)},
		// "":      {reflect.Ptr: uintptr(0)},
	} {
		for kind, exp := range tgtm {
			svv := reflect.ValueOf(src)
			tgtType := tgtTypes[kind]
			// src := rdecodesimple(svv)
			tgt, err := bbc.Transform(nil, svv, tgtType)
			if err != nil {
				t.Fatalf("err: %v", err)
			}

			if x := tgt.Interface(); x != exp {
				t.Fatalf("convert failed, want %v but got %v (%v)", exp, x, ref.Typfmt(tgt.Type()))
			}

			tgt = reflect.New(tgtType).Elem()
			err = bbc.CopyTo(nil, svv, tgt)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if x := tgt.Interface(); x != exp {
				t.Fatalf("convert failed, want %v but got %v (%v)", exp, x, ref.Typfmt(tgt.Type()))
			}
		}
	}
}

func TestToDurationConverter_Transform(t *testing.T) { //nolint:revive
	var bbc fromStringConverter
	dur := 3 * time.Second
	v := reflect.ValueOf(dur)
	t.Logf("dur: %v (%v, kind: %v, name: %v, pkgpath: %v)", dur, ref.Typfmtv(&v), v.Kind(), v.Type().Name(), v.Type().PkgPath())

	tgtType := reflect.TypeOf((*time.Duration)(nil)).Elem()

	src := typ.Any(int64(13 * time.Hour))
	svv := reflect.ValueOf(src)
	tgt, err := bbc.Transform(nil, svv, tgtType)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("res: %v (%v)", tgt.Interface(), ref.Typfmtv(&tgt))

	t.Run("toDurationConverter = pre", func(t *testing.T) {
		for ix, cas := range []struct {
			src, tgt, expect typ.Any
		}{
			{"71ms", &dur, 71 * time.Millisecond},
			{"9h71ms", &dur, 9*time.Hour + 71*time.Millisecond},
			{int64(13 * time.Hour), &dur, 13 * time.Hour},
		} {
			c := newDeepCopier()
			// ctx := newValueConverterContextForTest(c)
			svv = reflect.ValueOf(cas.src)
			err = c.CopyTo(cas.src, cas.tgt)
			// tgt, err = bbc.Transform(ctx, svv, tgtType)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if !reflect.DeepEqual(dur, cas.expect) {
				t.Fatalf("err transform: expect %v but got %v", cas.expect, tgt)
			}
			t.Logf("res #%d: %v", ix, dur)
		}
	})

	//

	t.Run("fromDurationConverter - normal test", func(t *testing.T) {
		inttyp := reflect.TypeOf((*int)(nil)).Elem()
		int64typ := reflect.TypeOf((*int64)(nil)).Elem()
		stringtyp := reflect.TypeOf((*string)(nil)).Elem()
		booltyp := reflect.TypeOf((*bool)(nil)).Elem()

		var fdc fromDurationConverter

		for ix, cas := range []struct {
			src, tgt, expect interface{} //nolint:revive
			desiredType      reflect.Type
		}{
			{13 * time.Hour, &dur, "13h0m0s", stringtyp},
			{71 * time.Millisecond, &dur, int(71 * time.Millisecond), inttyp},
			{9*time.Hour + 71*time.Millisecond, &dur, int64(9*time.Hour + 71*time.Millisecond), int64typ},
			{13 * time.Hour, &dur, true, booltyp},
			{0 * time.Hour, &dur, false, booltyp},
		} {
			c := newDeepCopier()
			ctx := newValueConverterContextForTest(c)
			svv = reflect.ValueOf(cas.src)
			// err = c.CopyTo(cas.src, cas.tgt)
			tgt, err = fdc.Transform(ctx, svv, cas.desiredType)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if !reflect.DeepEqual(tgt.Interface(), cas.expect) {
				t.Fatalf("err transform: expect %v but got %v (%v)", cas.expect, tgt.Interface(), ref.Typfmt(tgt.Type()))
			}
			t.Logf("res #%d: %v (%v)", ix, tgt.Interface(), ref.Typfmt(tgt.Type()))
		}
	})

	//

	t.Run("toDurationConverter - normal test", func(t *testing.T) {
		var tdc toDurationConverter

		for ix, cas := range []struct {
			src, tgt, expect interface{} //nolint:revive
		}{
			{"71ms", &dur, 71 * time.Millisecond},
			{"9h71ms", &dur, 9*time.Hour + 71*time.Millisecond},
			{int64(13 * time.Hour), &dur, 13 * time.Hour},
			{false, &dur, 0 * time.Second},
			{true, &dur, 1 * time.Nanosecond},
		} {
			c := newDeepCopier()
			ctx := newValueConverterContextForTest(c)
			svv = reflect.ValueOf(cas.src)
			// err = c.CopyTo(cas.src, cas.tgt)
			tgt, err = tdc.Transform(ctx, svv, tgtType)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if !reflect.DeepEqual(tgt.Interface(), cas.expect) {
				t.Fatalf("err transform: expect %v but got %v (%v)", cas.expect, tgt.Interface(), ref.Typfmt(tgt.Type()))
			}
			t.Logf("res #%d: %v (%v)", ix, tgt.Interface(), ref.Typfmt(tgt.Type()))
		}
	})

	// var c = newDeepCopier()
	// c.withConverters(&toDurationConverter{})
	// var ctx = newValueConverterContextForTest(c)
	// src = "71ms"
	// svv = reflect.ValueOf(src)
	// tgt, err = bbc.Transform(ctx, svv, tgtType)
	// if err != nil {
	//	t.Fatalf("err: %v", err)
	// }
	// t.Logf("res: %v (%v)", tgt.Interface(), typfmtv(&tgt))
	//
	// src = "9h71ms"
	// svv = reflect.ValueOf(src)
	// err = bbc.CopyTo(ctx, svv, reflect.ValueOf(&dur).Elem())
	// if err != nil {
	//	t.Fatalf("err: %v", err)
	// }
	// t.Logf("res: %v", dur)
	//
	// //
	//
	// c = newDeepCopier()
	// c.withCopiers(&toDurationConverter{})
	// ctx = newValueConverterContextForTest(c)
	// src = "71ms"
	// svv = reflect.ValueOf(src)
	// tgt, err = bbc.Transform(ctx, svv, tgtType)
	// if err != nil {
	//	t.Fatalf("err: %v", err)
	// }
	// t.Logf("res: %v (%v)", tgt.Interface(), typfmtv(&tgt))
	//
	// src = "9h71ms"
	// svv = reflect.ValueOf(src)
	// err = bbc.CopyTo(ctx, svv, reflect.ValueOf(&dur).Elem())
	// if err != nil {
	//	t.Fatalf("err: %v", err)
	// }
	// t.Logf("res: %v", dur)

	//

	c := newCopier()
	c.withFlags(cms.SliceMerge)
	c.withFlags(cms.MapMerge)
}

func TestToDurationConverter_fallback(t *testing.T) {
	var tdfs toDurationConverter
	dur := 3 * time.Second
	v := reflect.ValueOf(&dur)
	_ = tdfs.fallback(v)
	t.Logf("dur: %v", dur)
}

func TestToTimeConverter_Transform(t *testing.T) { //nolint:revive
	t.Run("fromTimeConverter - normal test", func(t *testing.T) {
		inttyp := reflect.TypeOf((*int)(nil)).Elem()
		int64typ := reflect.TypeOf((*int64)(nil)).Elem()
		stringtyp := reflect.TypeOf((*string)(nil)).Elem()
		// booltyp := reflect.TypeOf((*bool)(nil)).Elem()
		floattyp := reflect.TypeOf((*float64)(nil)).Elem()

		var ftc fromTimeConverter
		var dur int

		for ix, cas := range []struct {
			src         string
			tgt, expect interface{} //nolint:revive
			desiredType reflect.Type
		}{
			{"2001-02-03 04:05:06.078912", &dur, "2001-02-03T04:05:06Z", stringtyp},
			{"2001-02-03 04:05:06.078912", &dur, int(981173106), inttyp},
			{"2001-02-03 04:05:06.078912", &dur, int64(981173106), int64typ},
			{"2001-02-03 04:05:06.078912", &dur, float64(981173106.078912), floattyp},
		} {
			c := newDeepCopier()
			ctx := newValueConverterContextForTest(c)
			tm, err := time.Parse("2006-01-02 15:04:05.000000", cas.src)
			t.Logf("%q parsed: %v (%v)", cas.src, tm, err)
			svv := reflect.ValueOf(tm)
			// err = c.CopyTo(cas.src, cas.tgt)
			tgt, err := ftc.Transform(ctx, svv, cas.desiredType)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if !reflect.DeepEqual(tgt.Interface(), cas.expect) {
				t.Fatalf("err transform: expect %v but got %v (%v)", cas.expect, tgt.Interface(), ref.Typfmt(tgt.Type()))
			}
			t.Logf("res #%d: %v (%v)", ix, tgt.Interface(), ref.Typfmt(tgt.Type()))
		}
	})

	t.Run("toTimeConverter - normal test", func(t *testing.T) {
		var tdc toTimeConverter
		var tm time.Time
		layout := "2006-01-02 15:04:05.999999999Z07:00"
		tgtType := reflect.TypeOf((*time.Time)(nil)).Elem()

		for ix, cas := range []struct {
			src, tgt, expect interface{} //nolint:revive
		}{
			{"2001-02-03 04:05:06.078912", &tm, "2001-02-03 04:05:06.078912Z"},
			{"2001-02-03 04:05:06.078912345", &tm, "2001-02-03 04:05:06.078912345Z"},
			{int(981173106), &tm, "2001-02-03 04:05:06Z"},
			{int64(981173106), &tm, "2001-02-03 04:05:06Z"},
			{float64(981173106.078912), &tm, "2001-02-03 04:05:06.078912019Z"},
		} {
			c := newDeepCopier()
			ctx := newValueConverterContextForTest(c)
			svv := reflect.ValueOf(cas.src)
			// err = c.CopyTo(cas.src, cas.tgt)
			tgt, err := tdc.Transform(ctx, svv, tgtType)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			got := tgt.Interface().(time.Time).UTC().Format(layout)
			if !reflect.DeepEqual(got, cas.expect) {
				t.Fatalf("err transform: expect %v but got %v (%v)", cas.expect, got, ref.Typfmt(tgt.Type()))
			}
			t.Logf("res #%d: %v (%v)", ix, got, ref.Typfmt(tgt.Type()))
		}
	})
}

func TestFromStringConverter_defaultTypes(t *testing.T) {
	var fss fromStringConverter
	src := "987"
	dst := 3.3
	svv := reflect.ValueOf(src)
	dvv := reflect.ValueOf(&dst)

	ctx := newValueConverterContextForTest(newDeepCopier())
	ret, err := fss.convertToOrZeroTarget(ctx, svv, dvv.Type().Elem())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("ret: %v (%v)", ref.Valfmt(&ret), ref.Typfmtv(&ret))
}

func TestFromStringConverter_postCopyTo(t *testing.T) {
	var fss fromStringConverter

	src := "987"
	dst := 3.3
	svv := reflect.ValueOf(src)
	dvv := reflect.ValueOf(&dst)

	c := newDeepCopier().withFlags(cms.ClearIfInvalid)
	ctx := newValueConverterContextForTest(c)
	err := fss.postCopyTo(ctx, svv, dvv.Elem())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("ret: %v (%v)", dst, ref.Typfmtv(&dvv))
}

func TestToStringConverter_postCopyTo(t *testing.T) {
	var fss toStringConverter
	src := struct {
		fval float64
	}{3.3}
	dst := struct {
		fval string
	}{}
	svv := reflect.ValueOf(&src)
	dvv := reflect.ValueOf(&dst)
	sf1 := ref.Rindirect(svv).Field(0)
	df1 := ref.Rindirect(dvv).Field(0)
	// sft := reflect.TypeOf(src).Field(0)

	ctx := &ValueConverterContext{
		Params: &Params{
			srcOwner: &svv,
			dstOwner: &dvv,
			// field:      &sft,
			// fieldTags:  parseFieldTags(sft.Tag),
			targetIterator: newStructIterator(dvv,
				withStructPtrAutoExpand(true),
				withStructFieldPtrAutoNew(true),
				withStructSource(&svv, true),
			),
			controller: newDeepCopier(),
		},
	}

	ctx.nextTargetField() // call ctx.targetIterator.Next() to locate the first field

	sf2 := cl.GetUnexportedField(sf1)

	err := fss.postCopyTo(ctx, sf2, df1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if dst.fval != "3.3" {
		t.Fatalf("want '3.3' but got %v", dst.fval)
	}
	t.Logf("ret: %v (%v)", dst, ref.Typfmtv(&dvv))
}

type (
	si1 struct{}
	si2 struct{}
)

func (*si2) String() string { return "i2" }

func TestHasStringer(t *testing.T) {
	var i1 si1
	var i2 si2

	v := reflect.ValueOf(i1)
	t.Logf("si1: %v", ref.HasStringer(&v))
	v = reflect.ValueOf(i2)
	t.Logf("si2: %v", ref.HasStringer(&v))
	v = reflect.ValueOf(&i2)
	t.Logf("*si2: %v", ref.HasStringer(&v))
}

func TestNameToMapKey(t *testing.T) {
	name := "9527"
	// value := 789
	mapslice := []interface{}{ //nolint:revive
		map[int]interface{}{ //nolint:revive
			111: 333,
		},
		map[int]interface{}{ //nolint:revive
			9527: 333,
		},
		map[float32]interface{}{ //nolint:revive
			9527: 333,
		},
		map[complex128]interface{}{ //nolint:revive
			9527: 333,
		},
		map[string]interface{}{ //nolint:revive
			"my": 12,
		},
		map[string]interface{}{ //nolint:revive
			"9527": 33,
		},
	}

	for _, m := range mapslice {
		mv := reflect.ValueOf(&m) // nolint:gosec // G601: Implicit memory aliasing in for loop
		mvind := ref.Rdecodesimple(mv)
		t.Logf("    target map is %v", ref.Typfmtv(&mvind))
		mt := ref.Rdecodetypesimple(mvind.Type())
		key, err := nameToMapKey(name, mt)
		if err != nil {
			t.Errorf("nameToMapKey, has error: %v", err)
		} else {
			t.Logf("for target map %v, got key from nameToMapKey: %v %v", ref.Typfmt(mt), ref.Valfmt(&key), ref.Typfmt(key.Type()))
		}
	}
}

func TestFromFuncConverterAlongMainEntry(t *testing.T) {
	type A1 struct {
		Bv func() (int, error)
	}
	type B1 struct {
		Bv int
	}

	a1 := A1{func() (int, error) { return 3, nil }}
	b1 := B1{1}

	// test for fromFuncConverter along Copy -> cpController.findConverters
	Copy(&a1, &b1)

	if b1.Bv != 3 {
		t.Fatalf("expect %v but got %v", 3, b1.Bv)
	}
}

func TestFromFuncConverter(t *testing.T) { //nolint:revive
	fn0 := func() string { return "hello" }

	type C struct {
		A int
		B bool
	}
	type A struct {
		A func() C
		B func() bool
	}
	type B struct {
		C *C
		B bool
	}
	a0 := A{
		func() C { return C{7, true} },
		func() bool { return false },
	}
	b0 := B{nil, true}
	b1 := B{&C{7, true}, false}

	var boolTgt bool
	intTgt := 1
	stringTgt := "world"

	lazyInitRoutines()

	for ix, fnCase := range []struct {
		fn     interface{} //nolint:revive
		target interface{} //nolint:revive
		expect interface{} //nolint:revive
		err    error
	}{
		{func() string { return "hello" }, &intTgt, 1, ErrCannotSet},

		// {func() ([]int, error) { return []int{2, 3}, nil }, &[]int{1}, []int{1, 2, 3}, nil},

		// {func() ([2]int, error) { return [2]int{2, 3}, nil }, &[2]int{1}, [2]int{2, 3}},

		{
			func() A { return a0 },
			&b0,
			b1,
			nil,
		},

		{
			func() map[string]interface{} { return map[string]interface{}{"hello": "world"} }, //nolint:revive
			&map[string]interface{}{"k": 1, "hello": "bob"},                                   //nolint:revive
			map[string]interface{}{"hello": "world", "k": 1},                                  //nolint:revive
			nil,
		},

		{func() string { return "hello" }, &stringTgt, "hello", nil},
		{func() string { return "hello" }, &intTgt, 1, ErrCannotSet}, //nolint:revive,lll // string -> number, implicit converting will be tried, and failure if it's really not a number
		{func() string { return "789" }, &intTgt, 789, nil},
		{&fn0, &stringTgt, "hello", nil},

		{func() ([2]int, error) { return [2]int{2, 3}, nil }, &[2]int{1}, [2]int{2, 3}, nil},
		{func() ([2]int, error) { return [2]int{2, 3}, nil }, &[3]int{1}, [3]int{2, 3}, nil},
		{func() ([3]int, error) { return [3]int{2, 3, 5}, nil }, &[2]int{1}, [2]int{2, 3}, nil},
		{func() ([]int, error) { return []int{2, 3}, nil }, &[]int{1}, []int{1, 2, 3}, nil},

		{func() bool { return true }, &boolTgt, true, nil},
		{func() int { return 3 }, &intTgt, 3, nil},
		{func() (int, error) { return 5, nil }, &intTgt, 5, nil},
	} {
		if fnCase.fn != nil { //nolint:gocritic //nestingReduce: invert if cond, replace body with `continue`, move old body after the statement
			fnv := reflect.ValueOf(&fnCase.fn)
			tgtv := reflect.ValueOf(&fnCase.target)
			ff, tt := ref.Rdecodesimple(fnv), ref.Rdecodesimple(tgtv)
			dbglog.Log("---- CASE %d. %v -> %v", ix, ref.Typfmtv(&ff), ref.Typfmtv(&tt))
			t.Logf("---- CASE %d. %v -> %v", ix, ref.Typfmtv(&ff), ref.Typfmtv(&tt))

			c := fromFuncConverter{}
			ctx := newValueConverterContextForTest(nil)
			err := c.CopyTo(ctx, fnv, tgtv)

			if err != nil {
				if ret := tt.Interface(); reflect.DeepEqual(ret, fnCase.expect) {
					return
				}

				// for go1.11, 12, any error would be ignored
				// but when go1.13 and higher, we will test the returned error obj
				// can match case.err, see also TestFromFuncConverterGo113AndHigher()
				t.Fatalf("has error: %v, but expecting %v", err, fnCase.err)
			} else if ret := tt.Interface(); !reflect.DeepEqual(ret, fnCase.expect) {
				t.Fatalf("unexpect result: expect %v but got %v", fnCase.expect, ret)
			}
		}
	}
}

func newValueConverterContextForTest(c *cpController) *ValueConverterContext {
	if c == nil {
		c = newDeepCopier() //nolint:revive
	}
	return &ValueConverterContext{newParams(withOwnersSimple(c, nil))}
}

func TestToExportedName(t *testing.T) {
	for i, c := range []struct {
		src    string
		expect string
	}{
		{"IGotInternAtGeeksForGeeks", "IGotInternAtGeeksForGeeks"},
		{"iGotInternAtGeeksForGeeks", "IGotInternAtGeeksForGeeks"},
		{"i-got-intern-at-geeks-for-geeks", "IGotInternAtGeeksForGeeks"},
		{"i_got_intern_at_geeks_for_geeks", "IGotInternAtGeeksForGeeks"},
	} {
		actual := toExportedName(c.src)
		if actual != c.expect {
			t.Fatalf("%5d. expert %q -> %q but got %q", i, c.src, c.expect, actual)
		}
	}
}
