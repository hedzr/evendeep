// g/o:build !go1.18
// +/build !go1.18

package tool

// ReverseSlice reverse any slice/array
func ReverseSlice(s interface{}) { ReverseAnySlice(s) }

// ReverseStringSlice reverse a string slice
func ReverseStringSlice(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s

	// // reverse it
	// i := 0
	// j := len(a) - 1
	// for i < j {
	// 	a[i], a[j] = a[j], a[i]
	// 	i++
	// 	j--
	// }
}
