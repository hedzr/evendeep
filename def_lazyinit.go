package evendeep

import (
	"reflect"
	"sync"
)

var onceLazyInitRoutines sync.Once
var copyToRoutines map[reflect.Kind]copyfn
var otherLazyRoutines []func()

type copyfn func(c *cpController, params *Params, from, to reflect.Value) (err error)

func lazyInitRoutines() {
	onceLazyInitRoutines.Do(func() {
		copyToRoutines = map[reflect.Kind]copyfn{
			reflect.Ptr:           copyPointer,
			reflect.Uintptr:       copyUintptr,
			reflect.UnsafePointer: copyUnsafePointer,
			reflect.Chan:          copyChan,
			reflect.Interface:     copyInterface,
			reflect.Struct:        copyStruct,
			reflect.Slice:         copySlice,
			reflect.Array:         copyArray,
			reflect.Map:           copyMap,
			// reflect.Func:          copyFunc,

			// Invalid Kind = iota

			// Bool
			// Int
			// Int8
			// Int16
			// Int32
			// Int64
			// Uint
			// Uint8
			// Uint16
			// Uint32
			// Uint64
			// Uintptr
			// Float32
			// Float64
			// Complex64
			// Complex128

			// Array
			// Chan
			// Func
			// Interface
			// Map
			// Ptr
			// Slice
			// Struct

			// String

			// UnsafePointer
		}

		for _, fn := range otherLazyRoutines {
			if fn != nil {
				fn()
			}
		}
	})
}

func registerInitRoutines(fn func())     { otherRoutines = append(otherRoutines, fn) }         //nolint:unused,deadcode
func registerLazyInitRoutines(fn func()) { otherLazyRoutines = append(otherLazyRoutines, fn) } //nolint:unused,deadcode
