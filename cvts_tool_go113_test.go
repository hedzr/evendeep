//go:build go1.13
// +build go1.13

package evendeep

import (
	"reflect"
	"testing"

	"gopkg.in/hedzr/errors.v3"

	"github.com/hedzr/evendeep/dbglog"
	"github.com/hedzr/evendeep/ref"
)

func TestFromFuncConverterGo113AndHigher(t *testing.T) { //nolint:revive
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
	a0 := A{
		func() C { return C{7, true} },
		func() bool { return false },
	}
	b0 := B{nil, true}
	b1 := B{&C{7, true}, false}

	var boolTgt bool
	intTgt := 1
	stringTgt := "world"

	lazyInitRoutines()

	for ix, fnCase := range []struct {
		fn     interface{} //nolint:revive
		target interface{} //nolint:revive
		expect interface{} //nolint:revive
		err    error
	}{
		{func() ([]int, error) { return []int{2, 3}, nil }, &[]int{1}, []int{1, 2, 3}, nil},

		// {func() ([2]int, error) { return [2]int{2, 3}, nil }, &[2]int{1}, [2]int{2, 3}},

		{
			func() A { return a0 },
			&b0,
			b1,
			nil,
		},

		{
			func() map[string]interface{} { return map[string]interface{}{"hello": "world"} }, //nolint:revive
			&map[string]interface{}{"k": 1, "hello": "bob"},                                   //nolint:revive
			map[string]interface{}{"hello": "world", "k": 1},                                  //nolint:revive
			nil,
		},

		{func() string { return "hello" }, &stringTgt, "hello", nil},
		{func() string { return "hello" }, &intTgt, 1, ErrCannotSet}, //nolint:revive,lll // string -> number, implicit converting will be tried, and failure if it's really not a number
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
			ff, tt := ref.Rdecodesimple(fnv), ref.Rdecodesimple(tgtv)
			dbglog.Log("---- CASE %d. %v -> %v", ix, ref.Typfmtv(&ff), ref.Typfmtv(&tt))
			t.Logf("---- CASE %d. %v -> %v", ix, ref.Typfmtv(&ff), ref.Typfmtv(&tt))

			c := fromFuncConverter{}
			ctx := newValueConverterContextForTest(nil)
			err := c.CopyTo(ctx, fnv, tgtv)

			if err != nil {
				if ret := tt.Interface(); reflect.DeepEqual(ret, fnCase.expect) {
					return
				}

				if fnCase.err != nil && errors.Is(err, fnCase.err) {
					return
				}
				t.Fatalf("has error: %v, but expecting %v", err, fnCase.err)
			} else if ret := tt.Interface(); reflect.DeepEqual(ret, fnCase.expect) == false { //nolint:revive
				t.Fatalf("unexpect result: expect %v but got %v", fnCase.expect, ret)
			}
		}
	}
}
