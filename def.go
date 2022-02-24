package deepcopy

import (
	"reflect"
)

// DeepCopy _
func DeepCopy(fromObj, toObj interface{}, opts ...Opt) (result interface{}) {
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
func MakeClone(fromObj interface{}) (result interface{}) {
	if fromObj == nil {
		return
	}

	lazyInitRoutines()

	from := reflect.ValueOf(fromObj)
	find := rindirect(from)
	toPtr := reflect.New(find.Type())
	toPtrObj := toPtr.Interface()
	functorLog("toPtrObj: %v", toPtrObj)
	if err := DefaultCloneController.CopyTo(fromObj, toPtrObj); err == nil {
		result = toPtr.Elem().Interface()
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
	CopyTo(fromObj, toObj interface{}, opts ...Opt) (err error)
}

var (
	// DefaultCopyController provides standard deepcopy feature.
	// copy and merge slice or map to exist target
	DefaultCopyController = newCopier()
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
	var c = newCopier()
	c.flags = newFlags(SliceMerge, MapMerge)
	//c.autoExpandStuct = true
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// NewFlatDeepCopier gets a new instance of DeepCopier (the underlying
// is *cpController) like NewDeepCopier but no merge strategies
// (SliceMerge and MapMerge).
func NewFlatDeepCopier(opts ...Opt) DeepCopier {
	lazyInitRoutines()
	var c = newCopier()
	c.flags = newFlags()
	for _, opt := range opts {
		opt(c)
	}
	return c
}

//// NewCloner gets a new instance of DeepCopier (the underlying
//// is *cpController) different with DefaultCopyController and
//// DefaultCloneController.
//// It returns a cloner like MakeClone()
//func NewCloner(opts ...Opt) DeepCopier {
//	lazyInitRoutines()
//	var c = newCloner()
//	c.flags = newFlags()
//	for _, opt := range opts {
//		opt(c)
//	}
//	return c
//}

func newCopier() *cpController {
	return &cpController{
		valueConverters:            defaultValueConverters(),
		valueCopiers:               defaultValueCopiers(),
		copyFunctionResultToTarget: true,
		makeNewClone:               false,
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

//func newPlainCloner() *cpController {
//	return &cpController{
//		valueConverters:            defaultValueConverters(),
//		valueCopiers:               defaultValueCopiers(),
//		copyFunctionResultToTarget: true,
//		makeNewClone:               true,
//	}
//}
