package deepcopy

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
)

type sample struct {
	a int
	b string
}

func TestUintptrAndUnsafePointer(t *testing.T) {
	s := &sample{a: 1, b: "test"}

	//Getting the address of field b in struct s
	p := unsafe.Pointer(uintptr(unsafe.Pointer(s)) + unsafe.Offsetof(s.b))

	//Typecasting it to a string pointer and printing the value of it
	fmt.Println(*(*string)(p))

	u := uintptr(unsafe.Pointer(s))
	us := fmt.Sprintf("%v", u)
	t.Logf("us = 0x%v", us)
	v := reflect.ValueOf(us)
	ret := rToUIntegerHex(v, reflect.TypeOf(uintptr(unsafe.Pointer(s))))
	t.Logf("ret.type: %v, %v / 0x%x", ret.Type(), ret.Interface(), ret.Interface())

	//t.Logf("ret.type: %v, %v", ret.Type(), ret.Pointer())
}

func TestGetPointerAsUintptr(t *testing.T) {
	s := &sample{a: 1, b: "test"}

	v := reflect.ValueOf(s)
	u := getPointerAsUintptr(v)
	fmt.Println(u)
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

	for _, vi := range []interface{}{
		false,
		0,
		uint(0),
		math.Float64frombits(0),
		complex(math.Float64frombits(0), math.Float64frombits(0)),
		[0]int{},
		[1]int{0},
		(func())(nil),
		struct{}{},
		"f",
		"false",
		"off",
		"no",
		"famale",
	} {
		v1 := reflect.ValueOf(vi)
		v2, _ := rdecode(v1)
		if rToBool(v2).Interface() != false {
			t.Fatalf("for %v (%v) toBool failed", vi, typfmtv(&v2))
		}
	}

	for _, vi := range []interface{}{
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
		v2, _ := rdecode(v1)
		if rToBool(v2).Interface() != true {
			t.Fatalf("for %v (%v) toBool failed", vi, typfmtv(&v2))
		}
	}

}

func TestForInteger(t *testing.T) {
	for _, src := range []interface{}{
		13579,
		uint(13579),
	} {
		v1 := reflect.ValueOf(src)
		v1 = rdecodesimple(v1)
		if rForInteger(v1).Interface() != "13579" {
			t.Fail()
		}
	}

	var z interface{}
	v1 := reflect.ValueOf(z)
	v1 = rdecodesimple(v1)
	if x := rForInteger(v1).Interface(); x != "0" {
		t.Fatalf("failed, x = %v", x)
	}

	z = "bug"
	v1 = reflect.ValueOf(z)
	v1 = rdecodesimple(v1)
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
			"8.75": 8,
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
	for _, src := range []interface{}{
		13579,
		uint(13579),
	} {
		v1 := reflect.ValueOf(src)
		v1 = rdecodesimple(v1)
		if rForUInteger(v1).Interface() != "13579" {
			t.Fail()
		}
	}

	var z interface{}
	v1 := reflect.ValueOf(z)
	v1 = rdecodesimple(v1)
	if x := rForUInteger(v1).Interface(); x != "0" {
		t.Fatalf("failed, x = %v", x)
	}

	z = "bug"
	v1 = reflect.ValueOf(z)
	v1 = rdecodesimple(v1)
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
			"8.75": 8,
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
		//v1 := reflect.ValueOf(src)
		//v1 = rdecodesimple(v1)
		if x := rForUIntegerHex(uintptr(src)).Interface(); x != "0x3e67" {
			t.Fatalf("expect %v but got %v", "0x3e67", x)
		}
	}

	var z interface{}
	v1 := reflect.ValueOf(z)
	v1 = rdecodesimple(v1)
	if x := rForUInteger(v1).Interface(); x != "0" {
		t.Fatalf("failed, x = %v", x)
	}

	z = "bug"
	v1 = reflect.ValueOf(z)
	v1 = rdecodesimple(v1)
	if x := rForUInteger(v1).Interface(); x != "0" {
		t.Fatalf("failed, x = %v", x)
	}

	z = "0x3e59"
	v1 = reflect.ValueOf(z)
	v1 = rdecodesimple(v1)
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
	for _, src := range []interface{}{
		13579,
		uint(13579),
	} {
		v1 := reflect.ValueOf(src)
		v1 = rdecodesimple(v1)
		if rForFloat(v1).Interface() != "13579" {
			t.Fail()
		}
	}

	var z interface{}
	v1 := reflect.ValueOf(z)
	v1 = rdecodesimple(v1)
	if x := rForFloat(v1).Interface(); x != "0" {
		t.Fatalf("failed, x = %v", x)
	}

	z = "bug"
	v1 = reflect.ValueOf(z)
	v1 = rdecodesimple(v1)
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
	for src, exp := range map[interface{}]string{
		13579:        "(13579+0i)",
		uint(13579):  "(13579+0i)",
		1.316:        "(1.316+0i)",
		8.5 + 7.13i:  "(8.5+7.13i)",
		-8.5 - 7.13i: "(-8.5-7.13i)",
	} {
		v1 := reflect.ValueOf(src)
		v1 = rdecodesimple(v1)
		if x := rForComplex(v1).Interface(); x != exp {
			t.Fatalf("failed, x = %v, expect = %v", x, exp)
		}
	}

	var z interface{}
	v1 := reflect.ValueOf(z)
	v1 = rdecodesimple(v1)
	if x := rForComplex(v1).Interface(); x != "(0+0i)" {
		t.Fatalf("failed, x = %v", x)
	}

	z = "bug"
	v1 = reflect.ValueOf(z)
	v1 = rdecodesimple(v1)
	if x := rForComplex(v1).Interface(); x != "(0+0i)" {
		t.Fatalf("failed, x = %v", x)
	}

}

func TestToComplex(t *testing.T) {
	for _, dt := range []reflect.Type{
		//reflect.TypeOf((*complex64)(nil)).Elem(),
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
	bb.WriteString("hello")
	src := reflect.ValueOf(bb)
	tgt, err := bbc.Transform(nil, src, tgtType)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if x, ok := tgt.Interface().(bytes.Buffer); !ok {
		t.Fatalf("unexpect target value type: %v", tgt.Type())
	} else if x.String() != "hello" {
		t.Fatalf("convert failed, want 'hello' but got %q", x.String())
	}
}

func TestToStringConverter_Transform(t *testing.T) {
	var bbc toStringConverter
	tgtType := reflect.TypeOf((*string)(nil)).Elem()

	var bb bytes.Buffer
	bb.WriteString("hello")
	var sb strings.Builder
	sb.WriteString("hello")

	for sv, exp := range map[interface{}]string{
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

		var tgtstr string
		tgt = reflect.ValueOf(&tgtstr).Elem()
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
	var sss1 = sss{"hello"}
	var exp = "{hello}"

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

func TestFromStringConverter_Transform(t *testing.T) {
	var bbc fromStringConverter
	tgtTypes := map[reflect.Kind]reflect.Type{
		reflect.String:     reflect.TypeOf((*string)(nil)).Elem(),
		reflect.Bool:       reflect.TypeOf((*bool)(nil)).Elem(),
		reflect.Uint:       reflect.TypeOf((*uint)(nil)).Elem(),
		reflect.Int:        reflect.TypeOf((*int)(nil)).Elem(),
		reflect.Float64:    reflect.TypeOf((*float64)(nil)).Elem(),
		reflect.Complex128: reflect.TypeOf((*complex128)(nil)).Elem(),
		reflect.Ptr:        reflect.TypeOf((*int)(nil)).Elem(),
		reflect.Uintptr:    reflect.TypeOf((*uintptr)(nil)).Elem(),
	}

	for src, tgtm := range map[string]map[reflect.Kind]interface{}{
		"sss":    {reflect.String: "sss"},
		"true":   {reflect.Bool: true},
		"false":  {reflect.Bool: false},
		"123":    {reflect.Uint: uint(123)},
		"-123":   {reflect.Int: -123},
		"8.79":   {reflect.Float64: 8.79},
		"(3+4i)": {reflect.Complex128: 3 + 4i},
		"0x3e4a": {reflect.Uintptr: uintptr(0x3e4a)},
		//"":      {reflect.Ptr: uintptr(0)},
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
				t.Fatalf("convert failed, want %v but got %v (%v)", exp, x, typfmt(tgt.Type()))
			}

			tgt = reflect.New(tgtType).Elem()
			err = bbc.CopyTo(nil, svv, tgt)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if x := tgt.Interface(); x != exp {
				t.Fatalf("convert failed, want %v but got %v (%v)", exp, x, typfmt(tgt.Type()))
			}
		}
	}

	var dur = 3 * time.Second
	var v = reflect.ValueOf(dur)
	t.Logf("dur: %v (%v, kind: %v, name: %v, pkgpath: %v)", dur, typfmtv(&v), v.Kind(), v.Type().Name(), v.Type().PkgPath())

	tgtType := reflect.TypeOf((*time.Duration)(nil)).Elem()
	var src interface{} = int64(13 * time.Hour)
	svv := reflect.ValueOf(src)
	tgt, err := bbc.Transform(nil, svv, tgtType)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("res: %v (%v)", tgt.Interface(), typfmtv(&tgt))

	// todo parseDuration, parseDateTime, ...
	var c = newDeepCopier()
	c.withConverters(&toDurationFromString{})
	var ctx = &ValueConverterContext{
		Params: newParams(withOwners(c, nil, nil, nil, nil, nil)),
	}
	src = "71ms"
	svv = reflect.ValueOf(src)
	tgt, err = bbc.Transform(ctx, svv, tgtType)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("res: %v (%v)", tgt.Interface(), typfmtv(&tgt))

	src = "9h71ms"
	svv = reflect.ValueOf(src)
	err = bbc.CopyTo(ctx, svv, reflect.ValueOf(&dur).Elem())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("res: %v", dur)
}

func TestWithXXX(t *testing.T) {
	copier := New(
		WithValueConverters(&toDurationFromString{}),
		WithValueCopiers(&toDurationFromString{}),
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

	var dur time.Duration
	var src = "9h71ms"
	//var svv = reflect.ValueOf(src)
	//var tvv = reflect.ValueOf(&dur) // .Elem()

	err := copier.CopyTo(src, &dur)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("res: %v", dur)
}
