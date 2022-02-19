package deepcopy

// t.go - tools functions here

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

func ptrOf(tgt reflect.Value) reflect.Value {
	//for tgt.Kind() != reflect.Ptr {
	//	functorLog("tgt: %v, get pointer", tgt.Kind())
	//	tgt = reflect.NewAt(tgt.Elem().Type(), unsafe.Pointer(tgt.UnsafeAddr()))
	//}
	tgt = reflect.NewAt(tgt.Type(), unsafe.Pointer(tgt.UnsafeAddr()))
	return tgt
}

func ptr(tgt reflect.Value, want reflect.Type) (r reflect.Value) {
	//return reflect.PtrTo(tgt)
	//for tgt.Kind() != reflect.Ptr {
	//functorLog("tgt: %v, get pointer", tgt.Kind())
	//r = reflect.NewAt(tgt.Type(), unsafe.Pointer(tgt.Pointer()))
	//}
	if tgt.CanAddr() {
		r = tgt.Addr()
	} else {
		// in hard way
		if tgt.IsNil() {
			tmp := reflect.New(want)
			tgt.Set(tmp)
			r = tmp
		} else {
			var v = tgt.Interface()
			functorLog("     v: %v", v)
			r = reflect.ValueOf(&v)
			functorLog("   ret: %v, *%v", r.Kind(), r.Elem().Kind())
		}
	}
	if r.Type() == want {
		functorLog("ret: %v, *%v", r.Kind(), r.Elem().Kind())
		return
	}

	functorLog("NOTE an temp pointer was created as *%v", want.Kind())
	return reflect.New(want)
}

func inspectStructV(val reflect.Value) {
	if val.Kind() == reflect.Interface && !val.IsNil() {
		elm := val.Elem()
		if elm.Kind() == reflect.Ptr && !elm.IsNil() && elm.Elem().Kind() == reflect.Ptr {
			val = elm
		}
	}
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	for i, count := 0, val.NumField(); i < count; i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		address := "not-addressable"

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
			address = fmt.Sprintf("0x%X", valueField.Addr().Pointer())
		}

		fmt.Printf("kind: %v, ", valueField.Kind())
		var v interface{}
		if valueField.IsValid() && !valueField.IsZero() && valueField.CanInterface() {
			v = valueField.Interface()
		}
		fmt.Printf("%d/%d. Field Name: %s,\t Field Value: %v,\t Address: %v\t, Field type: %v\n",
			i, count, typeField.Name, v, address, typeField.Type)

		if valueField.Kind() == reflect.Struct {
			inspectStructV(valueField)
		}
	}
}

func InspectStruct(v interface{}) {
	inspectStructV(reflect.ValueOf(v))
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
