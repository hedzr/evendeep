package deepcopy

import (
	"github.com/hedzr/log"
	"gopkg.in/hedzr/errors.v2"
	"reflect"
	"strings"
)

// CopyTo _
func (c *cpController) CopyTo(fromObj, toObj Any, opts ...Opt) (err error) {
	for _, opt := range opts {
		opt(c)
	}

	var (
		from = c.indirect(reflect.ValueOf(fromObj))
		to   = c.indirect(reflect.ValueOf(toObj))
	)

	//if !to.CanAddr() {
	//	return errors.New("copy to value is unaddressable")
	//}

	// Return is from value is invalid
	if !from.IsValid() {
		return
	}

	err = c.copyTo(from, to, nil)
	return
}

func (c *cpController) copyTo(from, to reflect.Value, params *paramsPackage) (err error) {

	if from.CanInterface() {
		if dc, ok := from.Interface().(Cloneable); ok {
			to.Set(reflect.ValueOf(dc.Clone()))
			return
		}
		if dc, ok := from.Interface().(DeepCopyable); ok {
			to.Set(reflect.ValueOf(dc.DeepCopy()))
			return
		}
	}

	//fromType := c.indirectType(from.Type())
	//toType := c.indirectType(to.Type())

	defer func() {
		if e := recover(); e != nil {
			err = errors.New("[recovered] copyTo unsatisfied ([%v] %v -> [%v] %v), causes: %v",
				c.indirectType(from.Type()), from, c.indirectType(to.Type()), to, e).
				AttachGenerals(e)
			n := log.CalcStackFrames(1)   // skip defer-recover frame at first
			log.Skip(n).Errorf("%v", err) // skip golib frames and defer-recover frame, back to the point throwing panic

		}
	}()

	kind := from.Kind()
	if fn, ok := copyToRoutines[kind]; ok && fn != nil {
		err = fn(c, params, from, to)
		return
	}

	err = copyDefaultHandler(c, params, from, to)

	return
}

func (c *cpController) indirect(reflectValue reflect.Value) reflect.Value {
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	return reflectValue
}

func (c *cpController) indirectType(reflectType reflect.Type) reflect.Type {
	for reflectType.Kind() == reflect.Ptr || reflectType.Kind() == reflect.Slice {
		reflectType = reflectType.Elem()
	}
	return reflectType
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func contains(names []string, name string) bool {
	for _, n := range names {
		if strings.EqualFold(n, name) {
			return true
		}
	}
	return false
}

func containsPartialsOnly(partialNames []string, testedString string) (contains bool) {
	for _, n := range partialNames {
		if strings.Contains(testedString, n) {
			return true
		}
	}
	return
}

func partialContainsShort(names []string, partialNeedle string) (contains bool) {
	for _, n := range names {
		if strings.Contains(n, partialNeedle) {
			return true
		}
	}
	return
}

func partialContains(names []string, partialNeedle string) (index int, matched string, contains bool) {
	for ix, n := range names {
		if strings.Contains(n, partialNeedle) {
			return ix, n, true
		}
	}
	return -1, "", false
}
