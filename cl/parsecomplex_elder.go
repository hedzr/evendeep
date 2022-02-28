//go:build !go1.15
// +build !go1.15

// for lower than go1.15

package cl

import (
	"gopkg.in/hedzr/errors.v3"
	"strconv"
	"strings"
)

// FormatComplex converts the complex number c to a string of the
// form (a+bi) where a and b are the real and imaginary parts,
// formatted according to the format fmt and precision prec.
//
// The format fmt and precision prec have the same meaning as in FormatFloat.
// It rounds the result assuming that the original was obtained from a complex
// value of bitSize bits, which must be 64 for complex64 and 128 for complex128.
func FormatComplex(c complex128, fmt byte, prec, bitSize int) string {
	if bitSize != 64 && bitSize != 128 {
		panic("invalid bitSize")
	}
	bitSize >>= 1 // complex64 uses float32 internally

	// Check if imaginary part has a sign. If not, add one.
	im := strconv.FormatFloat(imag(c), fmt, prec, bitSize)
	if im[0] != '+' && im[0] != '-' {
		im = "+" + im
	}

	return "(" + strconv.FormatFloat(real(c), fmt, prec, bitSize) + im + "i)"
}

// ParseComplexSimple converts a string to complex number.
//
// Examples:
//
//    c1 := cmdr.ParseComplex("3-4i")
//    c2 := cmdr.ParseComplex("3.13+4.79i")
func ParseComplexSimple(s string) (v complex128) {
	return a2complexShort(s)
}

// ParseComplex converts a string to complex number.
// If the string is not valid complex format, return err not nil.
//
// Examples:
//
//    c1 := cmdr.ParseComplex("3-4i")
//    c2 := cmdr.ParseComplex("3.13+4.79i")
func ParseComplex(s string) (v complex128, err error) {
	return a2complex(s)
}

func a2complexShort(s string) (v complex128) {
	v, _ = a2complex(s)
	return
}

func a2complex(s string) (v complex128, err error) {
	s = strings.TrimSpace(strings.TrimRightFunc(strings.TrimLeftFunc(s, func(r rune) bool {
		return r == '('
	}), func(r rune) bool {
		return r == ')'
	}))

	if i := strings.IndexAny(s, "+-"); i >= 0 {
		rr, ii := s[0:i], s[i:]
		if j := strings.Index(ii, "i"); j >= 0 {
			var ff, fi float64
			ff, err = strconv.ParseFloat(strings.TrimSpace(rr), 64)
			if err != nil {
				return
			}
			fi, err = strconv.ParseFloat(strings.TrimSpace(ii[0:j]), 64)
			if err != nil {
				return
			}

			v = complex(ff, fi)
			return
		}
		err = errors.New("for a complex number, the imaginary part should end with 'i', such as '3+4i'")
		return

		// err = errors.New("not valid complex number.")
	}

	var ff float64
	ff, err = strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return
	}
	v = complex(ff, 0)
	return
}
