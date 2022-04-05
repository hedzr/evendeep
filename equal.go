package evendeep

import (
	"github.com/hedzr/evendeep/diff"
	"github.com/hedzr/evendeep/typ"
)

// DeepEqual compares a and b deeply inside.
//
//    equal := evendeep.DeepEqual(a, b)
//    fmt.Println(euqal)
func DeepEqual(a, b typ.Any, opts ...diff.Opt) (equal bool) {
	_, equal = diff.New(a, b, opts...)
	return
}
