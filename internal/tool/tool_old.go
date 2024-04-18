// g/o:build !go1.18
// +/build !go1.18

package tool

// ReverseSlice reverse any slice/array.
func ReverseSlice(s interface{}) { ReverseAnySlice(s) } //nolint:revive

// ReverseStringSlice reverse a string slice.
func ReverseStringSlice(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
