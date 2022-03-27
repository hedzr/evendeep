package deepcopy

import (
	"github.com/hedzr/deepcopy/flags"
	"github.com/hedzr/deepcopy/flags/cms"
	"github.com/hedzr/deepcopy/internal/dbglog"
	"reflect"
)

// Copy _
func Copy(fromObj, toObj interface{}, opts ...Opt) (result interface{}) {
	return DeepCopy(fromObj, toObj, opts...)
}

// DeepCopy _
func DeepCopy(fromObj, toObj interface{}, opts ...Opt) (result interface{}) {
	if fromObj == nil {
		return toObj
	}

	var saved = DefaultCopyController.flags
	defer func() { DefaultCopyController.flags = saved }()

	if err := DefaultCopyController.CopyTo(fromObj, toObj, opts...); err == nil {
		result = toObj
	}

	return
}

// MakeClone _
func MakeClone(fromObj interface{}) (result interface{}) {
	if fromObj == nil {
		return fromObj
	}

	from := reflect.ValueOf(fromObj)
	//find := rindirect(from)
	fromtyp := from.Type()
	findtyp := rdecodetypesimple(fromtyp)
	toPtr := reflect.New(findtyp)
	toPtrObj := toPtr.Interface()
	dbglog.Log("toPtrObj: %v", toPtrObj)

	var saved = defaultCloneController.flags
	defer func() { defaultCloneController.flags = saved }()

	if err := defaultCloneController.CopyTo(fromObj, toPtrObj); err == nil {
		result = toPtr.Elem().Interface()
	}

	return
}

// Cloneable _
// The native Clone algor of a Cloneable object can be adapted into DeepCopier.
type Cloneable interface {
	// Clone return a pointer to copy of source object.
	// But you can return the copy itself with your will.
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
	// copy and merge slice or map to an existed target.
	DefaultCopyController *cpController // by newDeepCopier()

	// defaultCloneController provides standard clone feature.
	// simply clone itself to a new fresh object to make a deep clone object.
	defaultCloneController *cpController // by newCloner()

	// onceCpController sync.Once
)

// New gets a new instance of DeepCopier (the underlying
// is *cpController) different with DefaultCopyController.
//
// Use New:
//
//     src, tgt := 123, 0
//     deepcopy.New().CopyTo(src, &tgt)
//
// Use package functions:
//
//     deepcopy.Copy(src, &tgt) // or synonym: deepcopy.DeepCopy(src, &tgt)
//     tgt = deepcopy.MakeClone(src)
//
// Use DefaultCopyController:
//
//     deepcopy.DefaultCopyController.CopyTo(src, &tgt)
//
func New(opts ...Opt) DeepCopier {
	//lazyInitRoutines()
	var c = newDeepCopier()
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// NewFlatDeepCopier gets a new instance of DeepCopier (the underlying
// is *cpController) like NewDeepCopier but no merge strategies
// (SliceMerge and MapMerge).
func NewFlatDeepCopier(opts ...Opt) DeepCopier {
	//lazyInitRoutines()
	var c = newCopier()
	c.flags = flags.New()
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

func newDeepCopier() *cpController {
	return &cpController{
		valueConverters: defaultValueConverters(),
		valueCopiers:    defaultValueCopiers(),

		copyUnexportedFields:       true,
		copyFunctionResultToTarget: true,
		passSourceAsFunctionInArgs: true,
		autoExpandStruct:           true,

		flags:        flags.New(cms.SliceMerge, cms.MapMerge),
		rethrow:      true,
		makeNewClone: false,
	}
}

func newCopier() *cpController {
	return &cpController{
		valueConverters:            defaultValueConverters(),
		valueCopiers:               defaultValueCopiers(),
		copyFunctionResultToTarget: true,
		passSourceAsFunctionInArgs: true,
		rethrow:                    true,
		makeNewClone:               false,
	}
}

func newCloner() *cpController {
	return &cpController{
		valueConverters:            defaultValueConverters(),
		valueCopiers:               defaultValueCopiers(),
		copyFunctionResultToTarget: true,
		passSourceAsFunctionInArgs: true,
		rethrow:                    true,
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
