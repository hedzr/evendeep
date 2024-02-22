package times

import (
	"testing"
	"time"
)

func TestSmartDurationString(t *testing.T) {
	data := []time.Duration{
		10*time.Hour + 11*time.Second + 13*time.Microsecond,
		10*time.Hour + 11*time.Minute + 13*time.Millisecond + 15*time.Nanosecond,
		15 * time.Nanosecond,
		-15 * time.Nanosecond,
		0,
	}

	t.Logf("%-32s %-32s %-32s %-32s\n", "d.String()", "shortDur(d)", "shortDur(d,true)", "Parsed(shortDur(d))")
	t.Logf("%-32s %-32s %-32s %-32s\n", "----------", "-----------", "----------------", "-------------------")
	for _, d := range data {
		t.Logf("%-32v %-32v %-32v %-32v / %v\n", d,
			SmartDurationString(d),
			SmartDurationStringEx(d, true),
			MustParseDuration(shortDur(d, false)),
			shortDurSimple(d),
		)
	}

	f := 11.013000015
	d := DurationFromFloat(f)
	t.Logf("%v", SmartDurationString(d))
}

func TestParseDuration(t *testing.T) {
	data := []string{
		"10h0m11.000013s",
		"0s",
		"0",
		"",
	}
	for _, str := range data {
		d := MustParseDuration(str)
		t.Logf("%-32v %-32v %-32v %-32v / %v\n", d,
			SmartDurationString(d),
			SmartDurationStringEx(d, true),
			MustParseDuration(shortDur(d, false)),
			shortDurSimple(d),
		)
	}
}
