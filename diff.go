package evendeep

import (
	"github.com/hedzr/evendeep/diff"
	"github.com/hedzr/evendeep/typ"
)

// DeepDiff compares a and b deeply inside.
//
//    delta, equal := evendeep.DeepDiff(a, b)
//    fmt.Println(delta)
//    fmt.Println(delta.PrettyPrint())
func DeepDiff(a, b typ.Any, opts ...diff.Opt) (delta diff.Diff, equal bool) {
	return diff.New(a, b, opts...)
}
