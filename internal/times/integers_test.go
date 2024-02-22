package times

import (
	"testing"
)

func TestSmartParseInt(t *testing.T) {
	_, _ = SmartParseInt("1")
	_ = MustSmartParseInt("1")
	_, _ = SmartParseUint("1")
	_ = MstSmartParseUint("1")
	_, _ = ParseFloat(".1")
	_ = MustParseFloat(".1")
	_, _ = ParseComplex(".1")
	_ = MustParseComplex(".1")
}

func TestQuote(t *testing.T) {
	data := []string{
		"  ",
		"\x1f\x09",
		"\"dp\"",
	}

	for _, str := range data {
		t.Logf("quoted: %v", quote(str))
	}
}
