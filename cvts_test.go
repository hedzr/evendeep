package evendeep_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/hedzr/evendeep"
	"github.com/hedzr/evendeep/internal/times"
)

func TestCvt_Bool(t *testing.T) {
	var cvt evendeep.Cvt

	for i, c := range []struct {
		src    any
		expect any
	}{
		{0, false},
		{1, true},
		{-2, true},
		{-3.13, true},
		{"0", false},
		{"1", true},
		{"off", false},
		{"on", true},
		{"f", false},
		{"t", true},
		{"false", false},
		{"true", true},
		{"female", false},
		{"male", true},
		{"n", false},
		{"y", true},
		{"no", false},
		{"yes", true},
		{"bad", false},
		{"ok", true},
		{"close", false},
		{"open", true},

		{"", false},
		{nil, false},
		{struct{}{}, false},
	} {
		actual := cvt.Bool(c.src)
		if !reflect.DeepEqual(actual, c.expect) {
			t.Logf("%5d. expecting %v, but got %v | src is %v", i, c.expect, actual, c.src)
		}
	}
}

func TestDur(t *testing.T) {
	v := times.MustParseDuration("1d13h2m19s743ms892ns")
	t.Log(v)
	v = times.MustParseDuration("37h2m19.743000892s")
	t.Log(v)
	v = times.MustParseDuration("1d13h2m19.743000892s")
	t.Log(v)
}

func TestCvt_String(t *testing.T) {
	var cvt evendeep.Cvt

	for i, c := range []struct {
		src    any
		expect any
	}{
		{0, "0"},
		{1, "1"},
		{-2, "-2"},
		{-3.13, "-3.13"},
		{3.13, "3.13"},
		{"0", "0"},
		{"1", "1"},
		{"off", "off"},

		{3.14 + 2.718i, "(3.14+2.718i)"},
		{3.14 - 2.718i, "(3.14-2.718i)"},

		{[]any{8.2, 1, true}, `[8.2,1,true]`},
		{map[any]any{8.2: 2.72, 1: uint64(128), true: false}, `{"8.2":2.72,"1":128,"true":false}`},

		{times.MustSmartParseTime("2013-1-29 3:13:59.71026508"), "2013-01-29T03:13:59.71026508Z"},
		{times.MustParseDuration("1d13h2m19s743ms892ns"),
			37*time.Hour + 2*time.Minute + 19*time.Second + 743*time.Millisecond + 892*time.Nanosecond},

		{"", ""},
		{nil, "<nil>"},
		{struct{}{}, "{}"},
	} {
		actual := cvt.String(c.src)
		if !reflect.DeepEqual(actual, c.expect) {
			t.Logf("%5d. expecting '%v', but got '%v' | src is '%v'", i, c.expect, actual, c.src)
		}
	}
}

func TestCvt_Int(t *testing.T) {
	var cvt evendeep.Cvt

	for i, c := range []struct {
		src    any
		expect int64
	}{
		{0, 0},
		{1, 1},
		{-2, -2},
		{-3.13, -3},
		{3.13, 3},
		{"0", 0},
		{"1", 1},

		{"off", 0},

		{3.14 + 2.718i, 3},
		{3.14 - 2.718i, 3},

		{[]any{8.2, 1, true}, 0},
		{map[any]any{8.2: 2.72, 1: uint64(128), true: false}, 0},

		{times.MustSmartParseTime("2013-1-29 3:13:59.71026508"), 0},
		{times.MustParseDuration("1d13h2m19s743ms892ns"), 0},

		{"", 0},
		{nil, 0},
		{struct{}{}, 0},
	} {
		actual := cvt.Int(c.src)
		if !reflect.DeepEqual(actual, c.expect) {
			t.Logf("%5d. expecting '%v', but got '%v' | src is '%v'", i, c.expect, actual, c.src)
		}
	}
}

func TestCvt_Uint(t *testing.T) {
	var cvt evendeep.Cvt

	for i, c := range []struct {
		src    any
		expect uint64
	}{
		{0, 0},
		{1, 1},
		{-2, 18446744073709551614},
		{-3.13, 18446744073709551613},
		{3.13, 3},
		{"0", 0},
		{"1", 1},

		{"off", 0},

		{3.14 + 2.718i, 3},
		{3.14 - 2.718i, 3},

		{[]any{8.2, 1, true}, 0},
		{map[any]any{8.2: 2.72, 1: uint64(128), true: false}, 0},

		{times.MustSmartParseTime("2013-1-29 3:13:59.71026508"), 0},
		{times.MustParseDuration("1d13h2m19s743ms892ns"), 0},

		{"", 0},
		{nil, 0},
		{struct{}{}, 0},
	} {
		actual := cvt.Uint(c.src)
		if !reflect.DeepEqual(actual, c.expect) {
			t.Logf("%5d. expecting '%v', but got '%v' | src is '%v'", i, c.expect, actual, c.src)
		}
	}
}

func TestCvt_Float(t *testing.T) {
	var cvt evendeep.Cvt

	for i, c := range []struct {
		src    any
		expect float64
	}{
		{0, 0},
		{1, 1},
		{-2, -2},
		{-3.13, -3.13},
		{3.13, 3.13},
		{"0", 0},
		{"1", 1},

		{"off", 0},

		{3.14 + 2.718i, 3.14},
		{3.14 - 2.718i, 3.14},

		{[]any{8.2, 1, true}, 0},
		{map[any]any{8.2: 2.72, 1: uint64(128), true: false}, 0},

		{times.MustSmartParseTime("2013-1-29 3:13:59.71026508"), 0},
		{times.MustParseDuration("1d13h2m19s743ms892ns"), 0},

		{"", 0},
		{nil, 0},
		{struct{}{}, 0},
	} {
		actual := cvt.Float64(c.src)
		if !reflect.DeepEqual(actual, c.expect) {
			t.Logf("%5d. expecting '%v', but got '%v' | src is '%v'", i, c.expect, actual, c.src)
		}
	}
}

func TestCvt_ParseComplex(t *testing.T) {
	s := "3.14+2.718i"
	var cvt evendeep.Cvt
	actual := cvt.Complex128(s)
	t.Logf("%v", actual)
}

func TestCvt_Complex(t *testing.T) {
	var cvt evendeep.Cvt

	for i, c := range []struct {
		src    any
		expect complex128
	}{
		{0, 0},
		{1, 1},
		{-2, -2},
		{-3.13, -3.13},
		{3.13, 3.13},
		{"0", 0},
		{"1", 1},

		{"off", 0},

		{"3.14+2.718i", 3.140000104904175 + 2.7179999351501465i},
		{"3.14-2.718i", 3.140000104904175 - 2.7179999351501465i},

		{"(3.14+2.718i)", 3.140000104904175 + 2.7179999351501465i},
		{"(3.14-2.718i)", 3.140000104904175 - 2.7179999351501465i},

		{3.14 + 2.718i, 3.14 + 2.718i},
		{3.14 - 2.718i, 3.14 - 2.718i},

		{[]any{8.2, 1, true}, 0},
		{map[any]any{8.2: 2.72, 1: uint64(128), true: false}, 0},

		{times.MustSmartParseTime("2013-1-29 3:13:59.71026508"), 0},
		{times.MustParseDuration("1d13h2m19s743ms892ns"), 0},

		{"", 0},
		{nil, 0},
		{struct{}{}, 0},
	} {
		actual := cvt.Complex128(c.src)
		if !reflect.DeepEqual(actual, c.expect) {
			t.Logf("%5d. expecting '%v', but got '%v' | src is '%v'", i, c.expect, actual, c.src)
		}
	}
}

func TestCvt_Duration(t *testing.T) {
	var cvt evendeep.Cvt

	for i, c := range []struct {
		src    any
		expect time.Duration
	}{
		{0, 0},
		{1, 1},
		{-2, -2},
		{-3.13, 0},
		{3.13, 0},
		{"0", 0},
		{"1", 1},

		{"off", 0},

		// more tests for string sources is not here

		{"3.14+2.718i", 0},
		{"3.14-2.718i", 0},

		{"(3.14+2.718i)", 0},
		{"(3.14-2.718i)", 0},

		{3.14 + 2.718i, 0},
		{3.14 - 2.718i, 0},

		{[]any{8.2, 1, true}, 0},
		{map[any]any{8.2: 2.72, 1: uint64(128), true: false}, 0},

		{times.MustSmartParseTime("2013-1-29 3:13:59.71026508"), 0},
		{times.MustParseDuration("1d13h2m19s743ms892ns"), times.MustParseDuration("37h2m19.743000892s")},

		{"", 0},
		{nil, 0},
		{struct{}{}, 0},
	} {
		actual := cvt.Duration(c.src)
		if !reflect.DeepEqual(actual, c.expect) {
			t.Logf("%5d. expecting '%v', but got '%v' | src is '%v'", i, c.expect, actual, c.src)
		}
	}
}

func TestCvt_Time(t *testing.T) {
	var cvt evendeep.Cvt

	zeroUnix := time.Unix(0, 0)
	var zeroTime time.Time

	for i, c := range []struct {
		src    any
		expect time.Time
	}{
		{0, zeroUnix},
		{1, zeroUnix.Add(time.Second)},
		{-2, zeroUnix.Add(-2 * time.Second)},
		{-3.13, zeroUnix.Add(-3*time.Second - 129*time.Millisecond - 999*time.Microsecond - 999*time.Nanosecond)},
		{3.13, zeroUnix.Add(3*time.Second + 129*time.Millisecond + 999*time.Microsecond + 999*time.Nanosecond)},
		{"0", zeroTime},
		{"1", zeroTime},

		{"off", zeroTime},

		// more tests for string sources is not here

		{"3.14+2.718i", zeroTime},
		{"3.14-2.718i", zeroTime},

		{"(3.14+2.718i)", zeroTime},
		{"(3.14-2.718i)", zeroTime},

		{3.14 + 2.718i, zeroUnix.Add(3*time.Second + 140*time.Millisecond)},
		{3.14 - 2.718i, zeroUnix.Add(3*time.Second + 140*time.Millisecond)},

		{[]any{8.2, 1, true}, zeroTime},
		{map[any]any{8.2: 2.72, 1: uint64(128), true: false}, zeroTime},

		{times.MustSmartParseTime("2013-1-29 3:13:59.71026508"),
			times.MustSmartParseTime("2013-1-29 3:13:59.71026508")},
		{times.MustParseDuration("1d13h2m19s743ms892ns"), zeroTime},

		{"", zeroTime},
		{nil, zeroTime},
		{struct{}{}, zeroTime},
	} {
		actual := cvt.Time(c.src)
		if !reflect.DeepEqual(actual, c.expect) {
			t.Logf("%5d. expecting '%v', but got '%v' | src is '%v'", i, c.expect, actual, c.src)
		}
	}
}
