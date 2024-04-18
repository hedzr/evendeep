package times

import (
	"errors"
	"regexp"
	"time"
)

// SmartDurationString converts a time.Duration value to string.
//
// It's better than time.Duration.String() because it produce days part.
// For example, 37h will be "3d1h".
func SmartDurationString(d time.Duration) string { return SmartDurationStringEx(d, false) }

// SmartDurationStringEx converts a time.Duration value to string.
//
// The boolean param 'frac' can be true, which means converter will try
// to format a float-point number as second part.
// For example, 11s13µs will be "11.000013s".
//
// By default, such as SmartDurationString, frac shall be set true
// so that will produce ms, µs and ns part.
func SmartDurationStringEx(d time.Duration, frac bool) string { return shortDur(d, frac) }

func shortDur(d time.Duration, frac bool) string {
	var arr [32]byte
	n := shortDurFormat(&arr, d, frac)
	return string(arr[n:])
}

func shortDurFormat(buf *[32]byte, d time.Duration, useFrac bool) int { //nolint:revive
	// Largest time is:
	// 2540400h10m10.000000000s
	// 2540400h10m10s999ms999us999na
	w := len(buf)

	u := uint64(d)
	neg := d < 0
	if neg {
		u = -u
	}

	if u < uint64(time.Second) {
		// Special case: if duration is smaller than a second,
		// use smaller units, like 1.2ms
		w = fmtSeconds(buf[:w], u, w)
	} else if useFrac {
		w--
		buf[w] = 's'

		w, u = fmtFrac(buf[:w], u, 9)

		// u is now integer seconds
		w = fmtInt(buf[:w], u%60)
		u /= 60

		// u is now integer minutes
		if u > 0 {
			w--
			buf[w] = 'm'
			w = fmtInt(buf[:w], u%60)
			u /= 60

			// u is now integer hours
			// Stop at hours because days can be different lengths.
			if u > 0 {
				w--
				buf[w] = 'h'
				w = fmtInt(buf[:w], u)
			}
		}
	} else {
		var days, hours, minutes, seconds, ms, us uint64
		if u >= uint64(24*time.Hour) {
			days = u / 24 / uint64(time.Hour)
			u = u % (24 * uint64(time.Hour)) //nolint:gocritic
		}
		if u >= uint64(time.Hour) {
			hours = u / uint64(time.Hour)
			u = u % uint64(time.Hour) //nolint:gocritic
		}
		if u >= uint64(time.Minute) {
			minutes = u / uint64(time.Minute)
			u = u % uint64(time.Minute) //nolint:gocritic
		}
		if u >= uint64(time.Second) {
			seconds = u / uint64(time.Second)
			u = u % uint64(time.Second) //nolint:gocritic
		}
		if u >= uint64(time.Millisecond) {
			ms = u / uint64(time.Millisecond)
			u = u % uint64(time.Millisecond) //nolint:gocritic
		}
		if u >= uint64(time.Microsecond) {
			us = u / uint64(time.Microsecond)
			u = u % uint64(time.Microsecond) //nolint:gocritic
		}

		if u > 0 {
			w--
			buf[w] = 's'
			w--
			buf[w] = 'n'
			w = fmtInt(buf[:w], u)
		}
		if us > 0 {
			w -= 3
			copy(buf[w:], "µs")
			w = fmtInt(buf[:w], us)
		}
		if ms > 0 {
			w--
			buf[w] = 's'
			w--
			buf[w] = 'm'
			w = fmtInt(buf[:w], ms)
		}
		if seconds > 0 {
			w--
			buf[w] = 's'
			w = fmtInt(buf[:w], seconds)
		}
		if minutes > 0 {
			w--
			buf[w] = 'm'
			w = fmtInt(buf[:w], minutes)
		}
		if hours > 0 {
			w--
			buf[w] = 'h'
			w = fmtInt(buf[:w], hours)
		}
		if days > 0 {
			w--
			buf[w] = 'd'
			w = fmtInt(buf[:w], days)
		}
	}

	if neg {
		w--
		buf[w] = '-'
	}

	return w
}

func fmtSeconds(buf []byte, u uint64, w int) (nw int) {
	// Special case: if duration is smaller than a second,
	// use smaller units, like 1.2ms
	w-- //nolint:revive
	buf[w] = 's'
	w = fmtMsec(buf[:w], u, w) //nolint:revive
	return w
}

func fmtMsec(buf []byte, u uint64, w int) (nw int) {
	var prec int
	w-- //nolint:revive
	switch {
	case u == 0:
		buf[w] = '0'
		return w
	case u < uint64(time.Microsecond):
		// print nanoseconds
		prec = 0
		buf[w] = 'n'
	case u < uint64(time.Millisecond):
		// print microseconds
		prec = 3
		// U+00B5 'µ' micro sign == 0xC2 0xB5
		w-- //nolint:revive // Need room for two bytes.
		copy(buf[w:], "µ")
	default:
		// print milliseconds
		prec = 6
		buf[w] = 'm'
	}
	w, u = fmtFrac(buf[:w], u, prec) //nolint:revive
	w = fmtInt(buf[:w], u)           //nolint:revive
	return w
}

// fmtFrac formats the fraction of v/10**prec (e.g., ".12345") into the
// tail of buf, omitting trailing zeros. It omits the decimal
// point too when the fraction is 0. It returns the index where the
// output bytes begin and the value v/10**prec.
func fmtFrac(buf []byte, v uint64, prec int) (nw int, nv uint64) {
	// Omit trailing zeros up to and including decimal point.
	w := len(buf)
	printed := false
	for i := 0; i < prec; i++ {
		digit := v % 10
		printed = printed || digit != 0
		if printed {
			w--
			buf[w] = byte(digit) + '0'
		}
		v /= 10 //nolint:revive
	}
	if printed {
		w--
		buf[w] = '.'
	}
	return w, v
}

// fmtInt formats v into the tail of buf.
// It returns the index where the output begins.
func fmtInt(buf []byte, v uint64) int {
	w := len(buf)
	if v == 0 {
		w--
		buf[w] = '0'
	} else {
		for v > 0 {
			w--
			buf[w] = byte(v%10) + '0'
			v /= 10 //nolint:revive
		}
	}
	return w
}

func shortDurSimple(d time.Duration) string {
	s := d.String()

	// if strings.HasSuffix(s, "m0s") {
	// 	s = s[:len(s)-2]
	// }
	// if strings.HasSuffix(s, "h0m") {
	// 	s = s[:len(s)-2]
	// }

	for _, z := range []struct{ reg, rep string }{
		{`(m?[hmnus])0[hmnus]?s?`, "$1"},
	} {
		re := regexp.MustCompile(z.reg)
		s = re.ReplaceAllString(s, z.rep)
	}
	return s
}

// DurationFromFloat converts a float-point number to
// time.Duration value.
//
// It treats the float number as seconds. The fraction part
// will be transformed as ms or smaller parts.
func DurationFromFloat(f float64) time.Duration {
	return time.Duration(f * float64(time.Second))
}

// MustParseDuration parses a duration string.
// A duration string is a possibly signed sequence of
// decimal numbers, each with optional fraction and a unit suffix,
// such as "300ms", "-1.5h" or "2h45m".
// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
func MustParseDuration(s string) (dur time.Duration) {
	dur, _ = ParseDuration(s)
	return
}

// ParseDuration parses a duration string.
//
// A duration string is a possibly signed sequence of
// decimal numbers, each with optional fraction and a unit suffix,
// such as "300ms", "-1.5h" or "2h45m".
// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
//
// The difference is we accept day part, such as '3d7s'.
func ParseDuration(s string) (time.Duration, error) { //nolint:revive
	// [-+]?([0-9]*(\.[0-9]*)?[a-z]+)+
	orig := s
	var d uint64
	neg := false

	// Consume [-+]?
	if s != "" {
		c := s[0]
		if c == '-' || c == '+' {
			neg = c == '-'
			s = s[1:] //nolint:revive
		}
	}
	// Special case: if all that is left is "0", this is zero.
	if s == "0" {
		return 0, nil
	}
	if s == "" {
		return 0, errors.New("time: invalid duration " + quote(orig))
	}
	for s != "" {
		var (
			v, f  uint64      // integers before, after decimal point
			scale float64 = 1 // value = v + f/scale
		)

		var err error

		// The next character must be [0-9.]
		if !(s[0] == '.' || '0' <= s[0] && s[0] <= '9') {
			return 0, errors.New("time: invalid duration " + quote(orig))
		}
		// Consume [0-9]*
		pl := len(s)
		v, s, err = leadingInt(s) //nolint:revive
		if err != nil {
			return 0, errors.New("time: invalid duration " + quote(orig))
		}
		pre := pl != len(s) // whether we consumed anything before a period

		// Consume (\.[0-9]*)?
		post := false
		if s != "" && s[0] == '.' {
			s = s[1:] //nolint:revive
			pl := len(s)
			f, scale, s = leadingFraction(s) //nolint:revive
			post = pl != len(s)
		}
		if !pre && !post {
			// no digits (e.g. ".s" or "-.s")
			return 0, errors.New("time: invalid duration " + quote(orig))
		}

		// Consume unit.
		i := 0
		for ; i < len(s); i++ {
			c := s[i]
			if c == '.' || '0' <= c && c <= '9' {
				break
			}
		}
		if i == 0 {
			return 0, errors.New("time: missing unit in duration " + quote(orig))
		}
		u := s[:i]
		s = s[i:] //nolint:revive
		unit, ok := unitMap[u]
		if !ok {
			return 0, errors.New("time: unknown unit " + quote(u) + " in duration " + quote(orig))
		}
		if v > 1<<63/unit {
			// overflow
			return 0, errors.New("time: invalid duration " + quote(orig))
		}
		v *= unit
		if f > 0 {
			// float64 is needed to be nanosecond accurate for fractions of hours.
			// v >= 0 && (f*unit/scale) <= 3.6e+12 (ns/h, h is the largest unit)
			v += uint64(float64(f) * (float64(unit) / scale))
			if v > 1<<63 {
				// overflow
				return 0, errors.New("time: invalid duration " + quote(orig))
			}
		}
		d += v
		if d > 1<<63 {
			return 0, errors.New("time: invalid duration " + quote(orig))
		}
	}
	if neg {
		return -time.Duration(d), nil
	}
	if d > 1<<63-1 {
		return 0, errors.New("time: invalid duration " + quote(orig))
	}
	return time.Duration(d), nil
}

var errLeadingInt = errors.New("time: bad [0-9]*") // never printed

// leadingInt consumes the leading [0-9]* from s.
func leadingInt[bytes []byte | string](s bytes) (x uint64, rem bytes, err error) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if x > 1<<63/10 {
			// overflow
			return 0, rem, errLeadingInt
		}
		x = x*10 + uint64(c) - '0'
		if x > 1<<63 {
			// overflow
			return 0, rem, errLeadingInt
		}
	}
	return x, s[i:], nil
}

// leadingFraction consumes the leading [0-9]* from s.
// It is used only for fractions, so does not return an error on overflow,
// it just stops accumulating precision.
func leadingFraction(s string) (x uint64, scale float64, rem string) {
	i := 0
	scale = 1
	overflow := false
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if overflow {
			continue
		}
		if x > (1<<63-1)/10 {
			// It's possible for overflow to give a positive number, so take care.
			overflow = true
			continue
		}
		y := x*10 + uint64(c) - '0'
		if y > 1<<63 {
			overflow = true
			continue
		}
		x = y
		scale *= 10
	}
	return x, scale, s[i:]
}

var unitMap = map[string]uint64{
	"ns": uint64(time.Nanosecond),
	"us": uint64(time.Microsecond),
	"µs": uint64(time.Microsecond), // U+00B5 = micro symbol
	"μs": uint64(time.Microsecond), // U+03BC = Greek letter mu
	"ms": uint64(time.Millisecond),
	"s":  uint64(time.Second),
	"m":  uint64(time.Minute),
	"h":  uint64(time.Hour),
	"d":  uint64(24 * time.Hour),
}
