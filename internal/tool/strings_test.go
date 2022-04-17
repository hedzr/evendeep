package tool

import "testing"

func TestString_Split(t *testing.T) {
	var s String = "hello world"
	t.Log(s.Split(" "))
	// Output:
	// [hello world]
}
