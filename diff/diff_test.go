package diff_test

import (
	"github.com/hedzr/evendeep/diff"
	"github.com/hedzr/evendeep/diff/testdata" //nolint:typecheck
	"github.com/hedzr/evendeep/internal/tool"

	"reflect"
	"testing"
	"time"
)

type testStruct struct {
	A, b int
	C    []int
	D    [3]int
}

type RecursiveStruct struct {
	Key   int
	Child *RecursiveStruct
}

func newRecursiveStruct(key int) *RecursiveStruct {
	a := &RecursiveStruct{
		Key: key,
	}
	b := &RecursiveStruct{
		Key:   key,
		Child: a,
	}
	a.Child = b
	return a
}

type testCase struct {
	a, b  interface{}
	diff  string
	equal bool
	opt   diff.Opt
}

func checkTestCases(t *testing.T, testData []testCase) {
	for i, td := range testData {
		dif, equal := diff.New(td.a, td.b, td.opt)
		if dif.PrettyPrint() != td.diff {
			t.Errorf("%d. PrettyDiff(%#v, %#v) diff = %#v; not %#v", i, td.a, td.b, dif.String(), td.diff)
			continue
		}
		if equal != td.equal {
			t.Errorf("%d. PrettyDiff(%#v, %#v) equal = %#v; not %#v", i, td.a, td.b, equal, td.equal)
			continue
		}
		t.Logf("%d passed. PrettyDiff(%#v, %#v)", i, td.a, td.b)
	}
}

type timeComparer struct{}

func (c *timeComparer) Match(typ reflect.Type) bool {
	return typ.String() == "time.Time"
}

func (c *timeComparer) Equal(ctx diff.Context, lhs, rhs reflect.Value, path diff.Path) (equal bool) {
	aTime := lhs.Interface().(time.Time) //nolint:errcheck //no need
	bTime := rhs.Interface().(time.Time) //nolint:errcheck //no need
	if equal = aTime.Equal(bTime); !equal {
		ctx.PutModified(ctx.PutPath(path), diff.Update{Old: aTime.String(), New: bTime.String(), Typ: tool.Typfmtvlite(&lhs)})
	}
	return
}

func TestPrettyDiff(t *testing.T) {
	testData := []testCase{
		{
			[]int{3, 0, 9},
			[]int{9, 3, 0},
			"",
			true,
			diff.WithSliceOrderedComparison(true),
		},

		{
			[]interface{}{3, 0, 9},
			[]interface{}{9, 3, 0},
			"",
			true,
			diff.WithSliceOrderedComparison(true),
		},

		{
			true,
			false,
			// "modified:  = false\n",
			"modified:  = false (bool) (Old: true)\n",
			false,
			diff.WithComparer(&timeComparer{}),
		},
		{
			true,
			0,
			"modified:  = <zero> (int) (Old: true)\n",
			false,
			diff.WithIgnoredFields(),
		},
		{
			[]int{0, 1, 2},
			[]int{0, 1, 2, 3},
			"added: [3] = 3\n",
			false,
			nil,
		},
		{
			[]int{0, 1, 2, 3},
			[]int{0, 1, 2},
			"removed: [3] = 3\n",
			false,
			nil,
		},
		{
			[]int{0},
			[]int{1},
			// "added: [0] = 1\nremoved: [0] = <zero>\n",
			"modified: [0] = 1 (int) (Old: <zero>)\n",
			false,
			nil,
		},
		{
			&[]int{0},
			&[]int{1},
			// "added: [0] = 1\nremoved: [0] = <zero>\n",
			"modified: [0] = 1 (int) (Old: <zero>)\n",
			false,
			nil,
		},
		{
			map[string]int{"a": 1, "b": 2},
			map[string]int{"b": 4, "c": 3},
			"added: [\"c\"] = 3\nmodified: [\"b\"] = 4 (int) (Old: 2)\nremoved: [\"a\"] = 1\n",
			// "added: [\"c\"] = 3\nmodified: [\"b\"] = 4\nremoved: [\"a\"] = 1\n",
			false,
			diff.WithSliceOrderedComparison(false),
		},
		{
			testStruct{1, 2, []int{1}, [3]int{4, 5, 6}},
			testStruct{1, 3, []int{1, 2}, [3]int{4, 5, 6}},
			"added: .C.[1] = 2\nmodified: .b = 3 (int) (Old: 2)\n",
			// "added: .C[1] = 2\nmodified: .b = 3\n",
			false,
			diff.WithSliceOrderedComparison(false),
		},
		{
			testStruct{1, 3, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, [3]int{4, 5, 6}},
			testStruct{1, 3, []int{42, 43, 44, 3, 4, 5, 6, 7, 8, 9, 45, 46, 12}, [3]int{4, 5, 6}},
			"modified: .C.[0] = 42 (int) (Old: <zero>)\nmodified: .C.[1] = 43 (int) (Old: 1)\nmodified: .C.[2] = 44 (int) (Old: 2)\nmodified: .C.[10] = 45 (int) (Old: 10)\nmodified: .C.[11] = 46 (int) (Old: 11)\n",
			// "modified: .C[0] = 42\nmodified: .C[1] = 43\nmodified: .C[2] = 44\nmodified: .C[10] = 45\nmodified: .C[11] = 46\n",
			false,
			diff.WithSliceOrderedComparison(false),
		},
		{
			nil,
			nil,
			"",
			true,
			nil,
		},
		{
			&struct{}{},
			nil,
			"modified:  = nil (Old: &{} (*struct {}))\n",
			// "modified:  = <nil>\n",
			false,
			nil,
		},
		{
			nil,
			&struct{}{},
			"modified:  = &{} (*struct {}) (Old: nil)\n",
			// "modified:  = &struct {}{}\n",
			false,
			nil,
		},
		{
			time.Time{},
			time.Time{},
			"",
			true,
			nil,
		},
		{
			testdata.MakeTest(10, "duck"),
			testdata.MakeTest(20, "foo"),
			"modified: .a = 20 (int) (Old: 10)\nmodified: .b = foo (string) (Old: duck)\n",
			false,
			nil,
		},
		{
			time.Date(2018, 7, 24, 14, 6, 59, 0, &time.Location{}),
			time.Date(2018, 7, 24, 14, 6, 59, 0, time.UTC),
			"",
			true,
			nil,
		},
		{
			time.Date(2017, 1, 1, 0, 0, 0, 0, &time.Location{}),
			time.Date(2018, 7, 24, 14, 6, 59, 0, time.UTC),
			"modified:  = 2018-07-24 14:06:59 +0000 UTC (time.Time) (Old: 2017-01-01 00:00:00 +0000 UTC)\n",
			false,
			nil,
		},
	}
	checkTestCases(t, testData)
}

func TestPrettyDiffRecursive(t *testing.T) {
	testData := []testCase{
		{
			newRecursiveStruct(1),
			newRecursiveStruct(1),
			"",
			true,
			nil,
		},
		{
			newRecursiveStruct(1),
			newRecursiveStruct(2),
			"modified: .Child..Key = 2 (int) (Old: 1)\nmodified: .Key = 2 (int) (Old: 1)\n",
			false,
			nil,
		},
	}
	checkTestCases(t, testData)
}

// func TestPathString(t *testing.T) {
// 	testData := []struct {
// 		in   Path
// 		want string
// 	}{{
// 		Path{StructField("test"), SliceIndex(1), MapKey{"blue"}, MapKey{12.3}},
// 		".test[1][\"blue\"][12.3]",
// 	}}
// 	for i, td := range testData {
// 		if out := td.in.String(); out != td.want {
// 			t.Errorf("%d. %#v.String() = %#v; not %#v", i, td.in, out, td.want)
// 		}
// 	}
// }

type ignoreStruct struct {
	A int `diff:"ignore"`
	a int
	B [3]int `diff:"-"`
	b [3]int
}

func TestIgnoreTag(t *testing.T) {
	s1 := ignoreStruct{1, 1, [3]int{1, 2, 3}, [3]int{4, 5, 6}}
	s2 := ignoreStruct{2, 1, [3]int{1, 8, 3}, [3]int{4, 5, 6}}

	dif, equal := diff.New(s1, s2)
	if !equal {
		t.Errorf("Expected structs to be equal. Diff:\n%s", dif.PrettyPrint())
	}

	s2 = ignoreStruct{2, 2, [3]int{1, 8, 3}, [3]int{4, 9, 6}}
	dif, equal = diff.New(s1, s2)
	if equal {
		t.Errorf("Expected structs NOT to be equal.")
	}
	expect := "modified: .a = 2 (int) (Old: 1)\nmodified: .b.[1] = 9 (int) (Old: 5)\n"
	if dif.PrettyPrint() != expect {
		t.Errorf("Expected diff to be:\n%v\nbut got:\n%v", expect, dif)
	}
}

func TestIgnoreStructFieldOption(t *testing.T) {
	a := struct {
		X string
		Y string
	}{
		X: "x",
		Y: "y",
	}
	b := struct {
		X string
		Y string
	}{
		X: "xx",
		Y: "y",
	}

	dif, equal := diff.New(a, b, diff.WithIgnoredFields("X"))
	if !equal {
		t.Errorf("Expected structs to be equal. Diff:\n%s", dif)
	}

	dif, equal = diff.New(a, b, diff.WithIgnoredFields("Y"))
	if equal {
		t.Errorf("Expected structs NOT to be equal.")
	}
	expect := "modified: .X = xx (string) (Old: x)\n"
	if dif.PrettyPrint() != expect {
		t.Errorf("Expected diff to be:\n%v\nbut got:\n%v", expect, dif)
	}
}
