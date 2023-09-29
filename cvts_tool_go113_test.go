//go:build go1.13
// +build go1.13

package evendeep

import (
	"reflect"
	"testing"

	"gopkg.in/hedzr/errors.v3"

	"github.com/hedzr/evendeep/dbglog"
	"github.com/hedzr/evendeep/internal/tool"
)

func TestFromFuncConverterGo113AndHigher(t *testing.T) {
	fn0 := func() string { return "hello" }

	type C struct {
		A int
		B bool
	}
	type A struct {
		A func() C
		B func() bool
	}
	type B struct {
		C *C
		B bool
	}
	var a0 = A{
		func() C { return C{7, true} },
		func() bool { return false },
	}
	var b0 = B{nil, true}
	var b1 = B{&C{7, true}, false}

	var boolTgt bool
	var intTgt = 1
	var stringTgt = "world"

	lazyInitRoutines()

	for ix, fnCase := range []struct {
		fn     interface{}
		target interface{}
		expect interface{}
		err    error
	}{
		{func() ([]int, error) { return []int{2, 3}, nil }, &[]int{1}, []int{1, 2, 3}, nil},

		// {func() ([2]int, error) { return [2]int{2, 3}, nil }, &[2]int{1}, [2]int{2, 3}},

		{func() A { return a0 },
			&b0,
			b1,
			nil,
		},

		{func() map[string]interface{} { return map[string]interface{}{"hello": "world"} },
			&map[string]interface{}{"k": 1, "hello": "bob"},
			map[string]interface{}{"hello": "world", "k": 1},
			nil,
		},

		{func() string { return "hello" }, &stringTgt, "hello", nil},
		{func() string { return "hello" }, &intTgt, 1, ErrCannotSet}, // string -> number, implicit converting will be tried, and failure if it's really not a number
		{func() string { return "789" }, &intTgt, 789, nil},
		{&fn0, &stringTgt, "hello", nil},

		{func() ([2]int, error) { return [2]int{2, 3}, nil }, &[2]int{1}, [2]int{2, 3}, nil},
		{func() ([2]int, error) { return [2]int{2, 3}, nil }, &[3]int{1}, [3]int{2, 3}, nil},
		{func() ([3]int, error) { return [3]int{2, 3, 5}, nil }, &[2]int{1}, [2]int{2, 3}, nil},
		{func() ([]int, error) { return []int{2, 3}, nil }, &[]int{1}, []int{1, 2, 3}, nil},

		{func() bool { return true }, &boolTgt, true, nil},
		{func() int { return 3 }, &intTgt, 3, nil},
		{func() (int, error) { return 5, nil }, &intTgt, 5, nil},
	} {
		if fnCase.fn != nil { //nolint:gocritic //nestingReduce: invert if cond, replace body with `continue`, move old body after the statement
			fnv := reflect.ValueOf(&fnCase.fn)
			tgtv := reflect.ValueOf(&fnCase.target)
			ff, tt := tool.Rdecodesimple(fnv), tool.Rdecodesimple(tgtv)
			dbglog.Log("---- CASE %d. %v -> %v", ix, tool.Typfmtv(&ff), tool.Typfmtv(&tt))
			t.Logf("---- CASE %d. %v -> %v", ix, tool.Typfmtv(&ff), tool.Typfmtv(&tt))

			c := fromFuncConverter{}
			ctx := newValueConverterContextForTest(nil)
			err := c.CopyTo(ctx, fnv, tgtv)

			if err != nil {
				if fnCase.err != nil && errors.Is(err, fnCase.err) {
					return
				}
				t.Fatalf("has error: %v, but expecting %v", err, fnCase.err)
			} else if ret := tt.Interface(); reflect.DeepEqual(ret, fnCase.expect) == false {
				t.Fatalf("unexpect result: expect %v but got %v", fnCase.expect, ret)
			}
		}
	}
}
