package times

import (
	"reflect"
	"testing"
	"time"
)

func TestMustSmartParseTime(t *testing.T) {
	data := []string{
		"1979-01-29 11:52:00.678910129",
	}

	for _, str := range data {
		tm := MustSmartParseTime(str)
		tm1 := MustSmartParseTimePtr(str)
		if !reflect.DeepEqual(tm, *tm1) {
			t.Fatalf("failed")
		}
		t.Logf("%v", tm)
	}
}

func TestAddKnownTimeFormats(t *testing.T) {
	src := "1979-01-29 11:52:00.678910129"
	var tm, tm1 time.Time
	var err error

	AddKnownTimeFormats("15")

	for i, c := range []struct {
		src       string
		expecting any
	}{
		{"1979-01-29 11:52:00.678910129", 0},
		{"1979-1-29 11:52:0.67891", 0},
		{"11", 0},
		{"0000-01-01 11:00:00 +0000", 0},
	} {
		tm, err = time.Parse("2006-1-2 15:4:5.999999999", c.src)
		tm1, err = SmartParseTime(c.src)
		if err != nil {
			t.Fatalf("%5d. time.Parse(%q) failed, err: %v.", i, c.src, err)
		}
		t.Logf("%5d. time: %v, smart: %v", i, tm, tm1)
	}

	tm, err = SmartParseTime(src)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("time: %v", tm)
}

// func TestRoundedSince(t *testing.T) {
// 	tm := MustSmartParseTime("5:11:22")
// 	since := RoundedSince("h", tm)
// 	tm1 := tm.Add(since)
// 	// tm1 := MustSmartParseTime("5:0:0")
// 	t.Logf("since: %q, tm: %v, tm1: %v", since, tm, tm1)
// 	// if tm.UnixNano() != tm1.UnixNano() {
// 	// 	t.Fail()
// 	// }
// 	// if !reflect.DeepEqual(tm, tm1) {
// 	// 	t.Fail()
// 	// }
// }

func TestShortDur(t *testing.T) {
	d, h, m, s := 13*24*time.Hour, 5*time.Hour, 4*time.Minute, 3*time.Second
	ds := []time.Duration{
		d + h + m + s, d + h + m, d + h + s, d + m + s, d + h, d + m, d + s, d,
		h + m + s, h + m, h + s, m + s, h, m, s, 0,
	}

	t.Logf("%-32s %-32s %-32s %-32s\n", "d.String()", "shortDur(d)", "shortDur(d,true)", "Parsed(shortDur(d))")
	t.Logf("%-32s %-32s %-32s %-32s\n", "----------", "-----------", "----------------", "-------------------")
	for _, d := range ds {
		t.Logf("%-32v %-32v %-32v %-32v\n", d, shortDur(d, false), shortDur(d, true), MustParseDuration(shortDur(d, false)))
	}
}

func TestShortDur2(t *testing.T) {
	h, m, s, n := 5*time.Second, 4*time.Millisecond, 3*time.Microsecond, 701*time.Nanosecond
	ds := []time.Duration{
		h + m + s + n, h + m + n, h + s + n, m + s + n, h + n, m + n, s + n, n, 0,
		3*24*time.Hour + 8*time.Hour + 9*time.Minute + 11*time.Second + m + s + n,
	}

	t.Logf("%-32s %-32s %-32s %-32s\n", "d.String()", "shortDur(d)", "shortDur(d,true)", "Parsed(shortDur(d))")
	t.Logf("%-32s %-32s %-32s %-32s\n", "----------", "-----------", "----------------", "-------------------")
	for _, d := range ds {
		t.Logf("%-32v %-32v %-32v %-32v\n", d, shortDur(d, false), shortDur(d, true), MustParseDuration(shortDur(d, false)))
	}
}
