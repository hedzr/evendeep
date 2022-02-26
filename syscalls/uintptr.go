package syscalls

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"unsafe"
)

func UintptrToString(p uintptr) string {
	u := uintptrToUint(p)
	return "0x" + strconv.FormatUint(u, 16)
}

func UintptrFromString(s string) uintptr {
	if s[0:2] == "0x" {
		u, e := strconv.ParseUint(s[2:], 16, 64)
		if e != nil {
			return uintptr(0)
		}
		return uintptr(u)
	} else {
		u, e := strconv.ParseUint(s, 16, 64)
		if e != nil {
			return uintptr(0)
		}
		return uintptr(u)
	}
}

func uintptrToUint(u uintptr) uint64 {
	size := unsafe.Sizeof(u)
	switch size {
	case 4:
		return uint64(uint32(u))
	case 8:
		return uint64(u)
	default:
		panic(fmt.Sprintf("unknown uintptr size: %v", size))
	}
}

func toBytes1(p uintptr) []byte {
	size := unsafe.Sizeof(p)
	b := make([]byte, size)
	switch size {
	case 4:
		binary.LittleEndian.PutUint32(b, uint32(p))
	case 8:
		binary.LittleEndian.PutUint64(b, uint64(p))
	default:
		panic(fmt.Sprintf("unknown uintptr size: %v", size))
	}
	return b
}

func toBytes2(u *uintptr) []byte {
	const sizeOfUintPtr = unsafe.Sizeof(uintptr(0))
	return (*[sizeOfUintPtr]byte)(unsafe.Pointer(u))[:]
}
