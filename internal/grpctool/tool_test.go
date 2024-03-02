package grpctool

import (
	"testing"
)

func TestInt64ToTime(t *testing.T) {
	tm := Int64ToTime(0)
	t.Logf("time: %v", tm)
}

func TestInt64SecondsToTime(t *testing.T) {
	tm := Int64SecondsToTime(0)
	t.Logf("time: %v", tm)
}

func TestDecodeZigZagInt(t *testing.T) {
	i, _ := DecodeZigZagInt([]byte{0x80, 1, 0, 0, 0, 0, 0})
	t.Logf("i: %v", i)
}

func TestDecodeZigZagUint(t *testing.T) {
	i, _ := DecodeZigZagUint([]byte{0x80, 1, 0, 0, 0, 0, 0})
	t.Logf("i: %v", i)
}
