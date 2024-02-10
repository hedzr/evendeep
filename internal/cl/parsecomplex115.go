//go:build go1.15
// +build go1.15

package cl

import "strconv"

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

// ParseComplex converts a string to complex number.
// If the string is not valid complex format, return err not nil.
//
// Examples:
//
//	c1 := cmdr.ParseComplex("3-4i")
//	c2 := cmdr.ParseComplex("3.13+4.79i")
func ParseComplex(s string) (v complex128, err error) {
	return strconv.ParseComplex(s, 128) //nolint:gomnd //no need
}
