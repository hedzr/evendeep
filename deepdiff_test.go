package evendeep

import (
	"github.com/hedzr/evendeep/diff"
	"github.com/hedzr/evendeep/typ"

	"testing"
)

func TestDeepDiff(t *testing.T) {
	testData := []testCase{
		{
			[]int{3, 0, 9},
			[]int{9, 3, 0},
			"",
			true,
			diff.WithSliceOrderedComparison(true),
		},
		{
			[]int{3, 0},
			[]int{9, 3, 0},
			"added: [0] = 9\n",
			false,
			diff.WithSliceOrderedComparison(true),
		},
		{
			[]int{3, 0},
			[]int{9, 3, 0},
			"added: [2] = <zero>\nmodified: [0] = 9 (int) (Old: 3)\nmodified: [1] = 3 (int) (Old: <zero>)\n",
			false,
			diff.WithSliceOrderedComparison(false),
		},
	}
	checkTestCases(t, testData)
}

func TestDeepEqual(t *testing.T) {
	equal := DeepEqual([]int{3, 0, 9}, []int{9, 3, 0}, diff.WithSliceOrderedComparison(true))
	if !equal {
		t.Errorf("expecting equal = true but got false")
	}
}

type testCase struct {
	a, b  typ.Any
	diff  string
	equal bool
	opt   diff.Opt
}

func checkTestCases(t *testing.T, testData []testCase) {
	for i, td := range testData {
		delta, equal := DeepDiff(td.a, td.b, td.opt)
		if delta.PrettyPrint() != td.diff {
			t.Errorf("%d. PrettyDiff(%#v, %#v) diff = %#v; not %#v", i, td.a, td.b, delta.String(), td.diff)
			continue
		}
		if equal != td.equal {
			t.Errorf("%d. PrettyDiff(%#v, %#v) equal = %#v; not %#v", i, td.a, td.b, equal, td.equal)
			continue
		}
		t.Logf("%d passed. PrettyDiff(%#v, %#v)\n%v", i, td.a, td.b, delta)
	}
}
