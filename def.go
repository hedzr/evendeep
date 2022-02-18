package deepcopy

import (
	"reflect"
	"sync"
)

// DeepCopy _
func DeepCopy(fromObj, toObj Any, opts ...Opt) (result Any) {
	if fromObj == nil {
		return toObj
	}

	lazyInitRoutines()

	if err := DefaultCopyController.CopyTo(fromObj, toObj, opts...); err == nil {
		return toObj
	}

	return
}

// MakeClone _
func MakeClone(fromObj Any) (result Any) {
	if fromObj == nil {
		return
	}

	lazyInitRoutines()

	from := reflect.ValueOf(fromObj)
	fit := DefaultCloneController.indirect(from)
	to := reflect.New(fit.Type())
	toObj := to.Interface()
	functorLog("toObj: %v", toObj)
	if err := DefaultCloneController.CopyTo(fromObj, toObj); err == nil {
		result = to.Elem().Interface()
	}

	return
}

// Cloneable _
// The native Clone algor of a Cloneable object can be adapted into DeepCopier.
type Cloneable interface {
	Clone() interface{}
}

// DeepCopyable _
// The native DeepCopy algor of a DeepCopyable object can be adapted into DeepCopier.
type DeepCopyable interface {
	DeepCopy() interface{}
}

// DeepCopier _
type DeepCopier interface {
	// CopyTo _
	CopyTo(fromObj, toObj Any, opts ...Opt) (err error)
}

var (
	// DefaultCopyController provides standard deepcopy feature.
	// copy and merge slice or map to exist target
	DefaultCopyController = newDeepCopier()
	// DefaultCloneController provides standard clone feature.
	// simply clone itself to a new fresh object to make a deep clone object.
	DefaultCloneController = newCloner()

	// onceCpController sync.Once
)

// NewDeepCopier gets a new instance of DeepCopier (the underlying
// is *cpController) different with DefaultCopyController and
// DefaultCloneController.
func NewDeepCopier(opts ...Opt) DeepCopier {
	lazyInitRoutines()
	var c = newDeepCopier()
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func newDeepCopier() *cpController {
	return &cpController{
		valueConverters:            defaultValueConverters(),
		valueCopiers:               defaultValueCopiers(),
		copyFunctionResultToTarget: true,
		mergeSlice:                 true,
		mergeMap:                   true,
	}
}

func newCloner() *cpController {
	return &cpController{
		valueConverters:            defaultValueConverters(),
		valueCopiers:               defaultValueCopiers(),
		copyFunctionResultToTarget: true,
		makeNewClone:               true,
	}
}

func newPlainCloner() *cpController {
	return &cpController{
		valueConverters:            defaultValueConverters(),
		valueCopiers:               defaultValueCopiers(),
		copyFunctionResultToTarget: true,
		makeNewClone:               true,
	}
}

var onceCopyToRoutines sync.Once
var copyToRoutines map[reflect.Kind]copyfn

type copyfn func(c *cpController, params *paramsPackage, from, to reflect.Value) (err error)

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
