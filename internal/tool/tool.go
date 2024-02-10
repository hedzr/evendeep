package tool

// tool.go - tools functions here

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"unsafe"

	"github.com/hedzr/evendeep/dbglog"
	"github.com/hedzr/evendeep/ref"
	"github.com/hedzr/evendeep/typ"
)

func PtrOf(tgt reflect.Value) reflect.Value {
	ret := reflect.NewAt(tgt.Type(), unsafe.Pointer(tgt.UnsafeAddr()))
	return ret
}

// func ptr(tgt reflect.Value, want reflect.Type) (r reflect.Value) {
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
// }

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

		var v typ.Any
		if valueField.IsValid() && !ref.IsZero(valueField) && valueField.CanInterface() {
			v = valueField.Interface()
		}
		log.Printf("%s%d/%d. Field Name: %s, Field Value: %v,\t Address: %v, Field type: %v [%s]\n",
			padding, i, count, typeField.Name, v, address, typeField.Type, valueField.Kind())

		if valueField.Kind() == reflect.Struct {
			inspectStructV(valueField, level+1)
		}
	}
}

// InspectStruct dumps wach field in a struct with its reflect information.
func InspectStruct(v typ.Any) {
	inspectStructV(reflect.ValueOf(v), 0)
}

// MinInt returns min-value of a and b.
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Contains checks if name is in names.
func Contains(names []string, name string) bool {
	for _, n := range names {
		if strings.EqualFold(n, name) {
			return true
		}
	}
	return false
}

// ContainsPartialsOnly checks if testedString has a part of in partialNames list.
func ContainsPartialsOnly(partialNames []string, testedString string) (contains bool) {
	for _, n := range partialNames {
		if strings.Contains(testedString, n) {
			return true
		}
	}
	return
}

// PartialContainsShort checks if one of names has partialNeedle as a part.
func PartialContainsShort(names []string, partialNeedle string) (contains bool) {
	for _, n := range names {
		if strings.Contains(n, partialNeedle) {
			return true
		}
	}
	return
}

// PartialContains checks if one of names has partialNeedle as a part.
func PartialContains(names []string, partialNeedle string) (index int, matched string, contains bool) {
	for ix, n := range names {
		if strings.Contains(n, partialNeedle) {
			return ix, n, true
		}
	}
	return -1, "", false
}

// ReverseAnySlice reverse any slice/array.
func ReverseAnySlice(s typ.Any) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

// FindInSlice finds a value elv is in array ns.
func FindInSlice(ns reflect.Value, elv typ.Any, i int) (found bool) {
	for j := 0; j < ns.Len(); j++ {
		tev := ns.Index(j).Interface()
		// dbglog.Log("  testing tgt[%v](%v) and src[%v](%v)", j, tev, i, elv)
		if reflect.DeepEqual(tev, elv) {
			found = true
			dbglog.Log("found an exists el at tgt[%v], for src[%v], value is: %v", j, i, elv)
			break
		}
	}
	return
}

// EqualClassical tests lhs and rhs is equal.
func EqualClassical(lhs, rhs reflect.Value) bool {
	lv, rv := lhs.IsValid(), rhs.IsValid()
	if !lv {
		return !rv
	}
	if !rv {
		return !lv
	}

	return reflect.DeepEqual(lhs.Interface(), rhs.Interface())
}
