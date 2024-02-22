package times

import (
	"strconv"
)

// SmartParseInt is a helper
func SmartParseInt(s string) (ret int64, err error) {
	ret, err = strconv.ParseInt(s, 0, 64)
	return
}

// MustSmartParseInt is a helper
func MustSmartParseInt(s string) (ret int64) {
	ret, _ = strconv.ParseInt(s, 0, 64)
	return
}

// SmartParseUint is a helper
func SmartParseUint(s string) (ret uint64, err error) {
	ret, err = strconv.ParseUint(s, 0, 64)
	return
}

// MstSmartParseUint is a helper
func MstSmartParseUint(s string) (ret uint64) {
	ret, _ = strconv.ParseUint(s, 0, 64)
	return
}

// ParseFloat is a helper
func ParseFloat(s string) (ret float64, err error) {
	ret, err = strconv.ParseFloat(s, 64)
	return
}

// MustParseFloat is a helper
func MustParseFloat(s string) (ret float64) {
	ret, _ = strconv.ParseFloat(s, 64)
	return
}

// ParseComplex is a helper
func ParseComplex(s string) (ret complex128, err error) {
	ret, err = strconv.ParseComplex(s, 64)
	return
}

// MustParseComplex is a helper
func MustParseComplex(s string) (ret complex128) {
	ret, _ = strconv.ParseComplex(s, 64)
	return
}
