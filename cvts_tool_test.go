package deepcopy

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"
)

type sample struct {
	a int
	b string
}

func TestUintptrAndUnsafePointer(t *testing.T) {
	s := &sample{a: 1, b: "test"}

	//Getting the address of field b in struct s
	p := unsafe.Pointer(uintptr(unsafe.Pointer(s)) + unsafe.Offsetof(s.b))

	//Typecasting it to a string pointer and printing the value of it
	fmt.Println(*(*string)(p))

	u := uintptr(unsafe.Pointer(s))
	us := fmt.Sprintf("%v", u)
	t.Logf("us = 0x%v", us)
	v := reflect.ValueOf(us)
	ret := rToUIntegerHex(v, reflect.TypeOf(uintptr(unsafe.Pointer(s))))
	t.Logf("ret.type: %v, %v / 0x%x", ret.Type(), ret.Interface(), ret.Interface())

	//t.Logf("ret.type: %v, %v", ret.Type(), ret.Pointer())
}

func TestGetPointerAsUintptr(t *testing.T) {
	s := &sample{a: 1, b: "test"}

	v := reflect.ValueOf(s)
	u := getPointerAsUintptr(v)
	fmt.Println(u)
}
