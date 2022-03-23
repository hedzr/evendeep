package syscalls

import (
	"fmt"
	"testing"
	"unsafe"
)

type sample struct {
	a int
	b string
}

func TestUnsafePointer(t *testing.T) {
	s := &sample{a: 1, b: "test"}

	//Getting the address of field b in struct s
	p := unsafe.Pointer(uintptr(unsafe.Pointer(s)) + unsafe.Offsetof(s.b))

	//Typecasting it to a string pointer and printing the value of it
	fmt.Println(*(*string)(p))

	str := *(*string)(p)
	if str != "test" {
		t.Fail()
	}

	//Get the address as a uintptr
	startAddress := uintptr(unsafe.Pointer(s))
	fmt.Printf("Start Address of s: %d, %x, %v\n",
		startAddress, startAddress,
		UintptrToString(startAddress),
	)

	str = UintptrToString(startAddress)
	if u := UintptrFromString(str); u != startAddress {
		t.Fail()
	} else {

		b := toBytes1(u)
		fmt.Println(b)
		b = toBytes2(&u)
		fmt.Println(b)

	}
}
