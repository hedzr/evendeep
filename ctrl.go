package deepcopy

import (
	"github.com/hedzr/log"
	"gopkg.in/hedzr/errors.v3"
	"reflect"
)

type cpController struct {
	//keepIfSourceIsNil  bool // 源字段值为nil指针时，目标字段的值保持不变
	//keepIfSourceIsZero bool // 源字段值为未初始化的零值时，目标字段的值保持不变 // 此条尚未实现
	//keepIfNotEqual     bool // keep target field value if not equals to source
	//zeroIfEquals       bool // 源和目标字段值相同时，目标字段被清除为未初始化的零值
	//eachFieldAlways    bool

	copyFunctionResultToTarget bool

	//mergeSlice bool
	//mergeMap   bool

	makeNewClone bool // make a new clone by copying to a fresh new object
	flags        Flags
	ignoreNames  []string

	valueConverters []ValueConverter
	valueCopiers    []ValueCopier
}

// CopyTo _
func (c *cpController) CopyTo(fromObjPtr, toObjPtr interface{}, opts ...Opt) (err error) {
	for _, opt := range opts {
		opt(c)
	}

	var (
		from0 = reflect.ValueOf(fromObjPtr)
		to0   = reflect.ValueOf(toObjPtr)
		from  = c.indirect(from0)
		to    = c.indirect(to0)
	)

	functorLog("from.type: %v | input: %v", from.Type().Kind(), from0.Type().Kind())
	functorLog("  to.type: %v | input: %v", to.Type().Kind(), to0.Type().Kind())

	//if !to.CanAddr() {
	//	return errors.New("copy to value is unaddressable")
	//}

	// Return is from value is invalid
	if !from.IsValid() {
		return
	}

	err = c.copyTo(nil, from, to)
	return
}

func (c *cpController) copyTo(params *paramsPackage, from, to reflect.Value) (err error) {

	if from.CanInterface() {
		if dc, ok := from.Interface().(Cloneable); ok {
			to.Set(reflect.ValueOf(dc.Clone()))
			return
		}
		if dc, ok := from.Interface().(DeepCopyable); ok {
			to.Set(reflect.ValueOf(dc.DeepCopy()))
			return
		}
	}

	//fromType := c.indirectType(from.Type())
	//toType := c.indirectType(to.Type())

	defer func() {
		if e := recover(); e != nil {
			err = errors.New("[recovered] copyTo unsatisfied ([%v] %v -> [%v] %v), causes: %v",
				c.indirectType(from.Type()), from, c.indirectType(to.Type()), to, e).
				WithData(e)
			n := log.CalcStackFrames(1)   // skip defer-recover frame at first
			log.Skip(n).Errorf("%v", err) // skip golib frames and defer-recover frame, back to the point throwing panic

		}
	}()

	kind := from.Kind()
	//functorLog(" - from.type: %v", kind)
	if fn, ok := copyToRoutines[kind]; ok && fn != nil {
		err = fn(c, params, from, to)
		return
	}

	err = copyDefaultHandler(c, params, from, to)

	return
}

func (c *cpController) findCopiers(params *paramsPackage, from, to reflect.Value) (copier ValueCopier, ctx *ValueConverterContext) {
	var yes bool
	for _, copier = range c.valueCopiers {
		if ctx, yes = copier.Match(params, from, to); yes {
			break
		}
	}
	return
}

func (c *cpController) findConverters(params *paramsPackage, from, to reflect.Value) (converter ValueConverter, ctx *ValueConverterContext) {
	var yes bool
	for _, converter = range c.valueConverters {
		if ctx, yes = converter.Match(params, from, to); yes {
			break
		}
	}
	return
}

func (c *cpController) withFlags(flags ...CopyMergeStrategy) *cpController {
	if c.flags == nil {
		c.flags = newFlags(flags...)
	} else {
		c.flags.withFlags(flags...)
	}
	return c
}

//func (c *cpController) ensureIsSlicePtr(to reflect.Value) reflect.Value {
//	// sliceValue = from.Elem()
//	// typeKindOfSliceElem = sliceValue.Type().Elem().Kind()
//	if to.Kind() != reflect.Ptr || to.Elem().Kind() != reflect.Slice {
//		x := reflect.New(c.indirect(to).Type())
//		x.Elem().Set(to)
//		return x
//	}
//	return to
//}

// decode decodes a value to its underlying type, if it was wrapped by
// interface{} or pointer.
//
//    var b = 11
//    var i interface{} = &b
//    var v = reflect.ValueOf(&i)
//    var n = c.decode(v)
//    println(n.Type())    // = int
func (c *cpController) decode(reflectValue reflect.Value) (ret, prev reflect.Value) {
	return rdecode(reflectValue)
}

func (c *cpController) skip(reflectValue reflect.Value, kinds ...reflect.Kind) (ret, prev reflect.Value) {
	return rskip(reflectValue, kinds...)
}

func (c *cpController) want(reflectValue reflect.Value, kinds ...reflect.Kind) reflect.Value {
	return rwant(reflectValue, kinds...)
}

func (c *cpController) indirect(reflectValue reflect.Value) reflect.Value {
	return rindirect(reflectValue)
}

//// indirectAnyAndPtr converts/follows Any/any/interface{} to its underlying type (with decoding)
//func (c *cpController) indirectAnyAndPtr(reflectValue reflect.Value) reflect.Value {
//	for k := reflectValue.Kind(); k == reflect.Interface || k == reflect.Ptr; k = reflectValue.Kind() {
//		reflectValue = reflectValue.Elem()
//	}
//	return reflectValue
//}
//
//// indirectAny converts/follows Any/any/interface{} to its underlying type (with decoding)
//func (c *cpController) indirectAny(reflectValue reflect.Value) reflect.Value {
//	for reflectValue.Kind() == reflect.Interface {
//		reflectValue = reflectValue.Elem()
//	}
//	return reflectValue
//}

func (c *cpController) indirectType(reflectType reflect.Type) reflect.Type {
	return rindirectType(reflectType)
}
