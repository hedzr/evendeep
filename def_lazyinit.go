package deepcopy

import (
	"reflect"
	"sync"
)

var onceCopyToRoutines sync.Once
var copyToRoutines map[reflect.Kind]copyfn

type copyfn func(c *cpController, params *Params, from, to reflect.Value) (err error)

func lazyInitRoutines() { onceCopyToRoutines.Do(initRoutines) }
func initRoutines() {
	copyToRoutines = map[reflect.Kind]copyfn{
		reflect.Ptr:           copyPointer,
		reflect.Uintptr:       copyUintptr,
		reflect.UnsafePointer: copyUnsafePointer,
		reflect.Func:          copyFunc,
		reflect.Chan:          copyChan,
		reflect.Interface:     copyInterface,
		reflect.Struct:        copyStruct,
		reflect.Slice:         copySlice,
		reflect.Array:         copyArray,
		reflect.Map:           copyMap,

		//Invalid Kind = iota
		//Bool
		//Int
		//Int8
		//Int16
		//Int32
		//Int64
		//Uint
		//Uint8
		//Uint16
		//Uint32
		//Uint64
		//Uintptr
		//Float32
		//Float64
		//Complex64
		//Complex128
		//Array
		//Chan
		//Func
		//Interface
		//Map
		//Ptr
		//Slice
		//String
		//Struct
		//UnsafePointer
	}
}
