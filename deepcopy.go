package evendeep

import (
	"reflect"

	"github.com/hedzr/evendeep/dbglog"
	"github.com/hedzr/evendeep/flags"
	"github.com/hedzr/evendeep/flags/cms"
	"github.com/hedzr/evendeep/ref"
)

// New gets a new instance of DeepCopier (the underlying
// is *cpController) different with DefaultCopyController.
//
// Use New:
//
//	src, tgt := 123, 0
//	err = evendeep.New().CopyTo(src, &tgt)
//
// Use package functions (With-opts might cumulate):
//
//	evendeep.Copy(src, &tgt) // or synonym: evendeep.DeepCopy(src, &tgt)
//	tgt = evendeep.MakeClone(src)
//
// Use DefaultCopyController (With-opts might cumulate):
//
//	evendeep.DefaultCopyController.CopyTo(src, &tgt)
//
// The most conventional way is:
//
//	err := evendeep.New().CopyTo(src, &tgt)
func New(opts ...Opt) DeepCopier {
	c := newDeepCopier()
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Copy is a synonym of DeepCopy.
//
// DeepCopy makes a deep clone of a source object or merges it into the target.
func Copy(fromObj, toObj interface{}, opts ...Opt) (result interface{}) {
	return DeepCopy(fromObj, toObj, opts...)
}

// DeepCopy makes a deep clone of a source object or merges it into the target.
func DeepCopy(fromObj, toObj interface{}, opts ...Opt) (result interface{}) {
	if fromObj == nil {
		return toObj
	}

	defer DefaultCopyController.SaveFlagsAndRestore()()

	if err := DefaultCopyController.CopyTo(fromObj, toObj, opts...); err == nil {
		result = toObj
	}

	return
}

// MakeClone makes a deep clone of a source object.
func MakeClone(fromObj interface{}) (result interface{}) {
	if fromObj == nil {
		return nil
	}

	var (
		from     = reflect.ValueOf(fromObj)
		fromTyp  = from.Type()
		findTyp  = ref.Rdecodetypesimple(fromTyp)
		toPtr    = reflect.New(findTyp)
		toPtrObj = toPtr.Interface()
	)
	dbglog.Log("toPtrObj: %v", toPtrObj)

	defer defaultCloneController.SaveFlagsAndRestore()()

	if err := defaultCloneController.CopyTo(fromObj, toPtrObj); err == nil {
		result = toPtr.Elem().Interface()
	}

	return
}

// Cloneable interface represents a cloneable object that supports Clone() method.
//
// The native Clone algorithm of a Cloneable object can be adapted into DeepCopier.
type Cloneable interface {
	// Clone return a pointer to copy of source object.
	// But you can return the copy itself with your will.
	Clone() interface{}
}

// DeepCopyable interface represents a cloneable object that supports DeepCopy() method.
//
// The native DeepCopy algorithm of a DeepCopyable object can be adapted into DeepCopier.
type DeepCopyable interface {
	DeepCopy() interface{}
}

// DeepCopier interface.
type DeepCopier interface {
	// CopyTo function.
	CopyTo(fromObj, toObj interface{}, opts ...Opt) (err error)
}

//nolint:gochecknoglobals //i know that
var (
	// DefaultCopyController provides standard deepcopy feature.
	// copy and merge slice or map to an existed target.
	DefaultCopyController *cpController // by newDeepCopier()

	// defaultCloneController provides standard clone feature.
	// simply clone itself to a new fresh object to make a deep clone object.
	defaultCloneController *cpController // by newCloner()
)

// NewFlatDeepCopier gets a new instance of DeepCopier (the underlying
// is *cpController) like NewDeepCopier but no merge strategies
// (SliceMerge and MapMerge).
func NewFlatDeepCopier(opts ...Opt) DeepCopier {
	// lazyInitRoutines()
	c := newCopier()
	c.flags = flags.New()
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func newDeepCopier() *cpController {
	// lazyInitRoutines()
	return &cpController{
		valueConverters: defaultValueConverters(),
		valueCopiers:    defaultValueCopiers(),

		copyUnexportedFields:       true,
		copyFunctionResultToTarget: true,
		passSourceAsFunctionInArgs: true,
		autoExpandStruct:           true,
		autoNewStruct:              true,

		flags:        flags.New(cms.SliceMerge, cms.MapMerge),
		rethrow:      true,
		makeNewClone: false,
	}
}

func newCopier() *cpController {
	return &cpController{
		valueConverters: defaultValueConverters(),
		valueCopiers:    defaultValueCopiers(),

		copyFunctionResultToTarget: true,
		passSourceAsFunctionInArgs: true,

		rethrow:      true,
		makeNewClone: false,
	}
}

func newCloner() *cpController {
	return &cpController{
		valueConverters: defaultValueConverters(),
		valueCopiers:    defaultValueCopiers(),

		copyFunctionResultToTarget: true,
		passSourceAsFunctionInArgs: true,
		autoExpandStruct:           true,
		autoNewStruct:              true,

		flags:        flags.New(cms.Default),
		rethrow:      true,
		makeNewClone: true,
	}
}
