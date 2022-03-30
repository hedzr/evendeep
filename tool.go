package deepcopy

// tool.go - tools functions here

import (
	"fmt"
	"github.com/hedzr/deepcopy/internal/dbglog"
	"reflect"
	"strings"
	"unsafe"
)

func ptrOf(tgt reflect.Value) reflect.Value {
	//for tgt.Kind() != reflect.Ptr {
	//	Log("tgt: %v, get pointer", tgt.Kind())
	//	tgt = reflect.NewAt(tgt.Elem().Type(), unsafe.Pointer(tgt.UnsafeAddr()))
	//}
	ret := reflect.NewAt(tgt.Type(), unsafe.Pointer(tgt.UnsafeAddr()))
	return ret
}

//func ptr(tgt reflect.Value, want reflect.Type) (r reflect.Value) {
//	//return reflect.PtrTo(tgt)
//	//for tgt.Kind() != reflect.Ptr {
//	//Log("tgt: %v, get pointer", tgt.Kind())
//	//r = reflect.NewAt(tgt.Type(), unsafe.Pointer(tgt.Pointer()))
//	//}
//	if tgt.CanAddr() {
//		r = tgt.Addr()
//	} else {
//		// in hard way
//		if tgt.IsNil() {
//			tmp := reflect.New(want)
//			tgt.Set(tmp)
//			r = tmp
//		} else {
//			var v = tgt.Interface()
//			Log("     v: %v", v)
//			r = reflect.ValueOf(&v)
//			Log("   ret: %v, *%v", r.Kind(), r.Elem().Kind())
//		}
//	}
//	if r.Type() == want {
//		Log("ret: %v, *%v", r.Kind(), r.Elem().Kind())
//		return
//	}
//
//	Log("NOTE an temp pointer was created as *%v", want.Kind())
//	return reflect.New(want)
//}

func testFieldValue(valueField reflect.Value) (v reflect.Value, addrStr string) {
	addrStr = "not-addressable"
	if valueField.Kind() == reflect.Interface && !valueField.IsNil() {
		elm := valueField.Elem()
		if elm.Kind() == reflect.Ptr && !elm.IsNil() && elm.Elem().Kind() == reflect.Ptr {
			valueField = elm
		}
	}
	if valueField.Kind() == reflect.Ptr {
		valueField = valueField.Elem()
	}
	if valueField.CanAddr() {
		addrStr = fmt.Sprintf("0x%X", valueField.Addr().Pointer())
	}
	v = valueField
	return
}

func inspectStructV(val reflect.Value, level int) {
	if val.Kind() == reflect.Interface && !val.IsNil() {
		elm := val.Elem()
		if elm.Kind() == reflect.Ptr && !elm.IsNil() && elm.Elem().Kind() == reflect.Ptr {
			val = elm
		}
	}
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	padding := strings.Repeat("  ", level)
	for i, count := 0, val.NumField(); i < count; i++ {
		valField := val.Field(i)
		typeField := val.Type().Field(i)
		valueField, address := testFieldValue(valField)

		var v interface{}
		if valueField.IsValid() && !isZero(valueField) && valueField.CanInterface() {
			v = valueField.Interface()
		}
		fmt.Printf("%s%d/%d. Field Name: %s, Field Value: %v,\t Address: %v, Field type: %v [%s]\n",
			padding, i, count, typeField.Name, v, address, typeField.Type, valueField.Kind())

		if valueField.Kind() == reflect.Struct {
			inspectStructV(valueField, level+1)
		}
	}
}

// InspectStruct dumps wach field in a struct with its reflect information
func InspectStruct(v interface{}) {
	inspectStructV(reflect.ValueOf(v), 0)
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

// reverseAnySlice reverse any slice/array
func reverseAnySlice(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

// reverseStringSlice reverse a string slice
func reverseStringSlice(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s

	//// reverse it
	//i := 0
	//j := len(a) - 1
	//for i < j {
	//	a[i], a[j] = a[j], a[i]
	//	i++
	//	j--
	//}
}

func findInSlice(ns reflect.Value, elv interface{}, i int) (found bool) {
	for j := 0; j < ns.Len(); j++ {
		tev := ns.Index(j).Interface()
		dbglog.Log("  testing tgt[%v](%v) and src[%v](%v)", j, tev, i, elv)
		if reflect.DeepEqual(tev, elv) {
			found = true
			dbglog.Log("found exists el at tgt[%v], for src[%v], value is %v", j, i, elv)
			break
		}
	}
	return
}

func equalClassical(lhs, rhs reflect.Value) bool {
	lv, rv := lhs.IsValid(), rhs.IsValid()
	if !lv {
		return !rv
	}
	if !rv {
		return !lv
	}

	return reflect.DeepEqual(lhs.Interface(), rhs.Interface())
}
