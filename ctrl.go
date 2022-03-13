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

	copyUnexportedFields       bool
	copyFunctionResultToTarget bool
	autoExpandStruct           bool

	//mergeSlice bool
	//mergeMap   bool

	makeNewClone bool      // make a new clone by copying to a fresh new object
	flags        Flags     // CopyMergeStrategies globally
	ignoreNames  []string  // optional ignored names with wild-matching
	funcInputs   []log.Any // preset input args for function invoking

	valueConverters ValueConverters
	valueCopiers    ValueCopiers
}

// CopyTo _
func (c *cpController) CopyTo(fromObjOrPtr, toObjPtr interface{}, opts ...Opt) (err error) {

	lazyInitRoutines()

	for _, opt := range opts {
		opt(c)
	}

	var (
		from0 = reflect.ValueOf(fromObjOrPtr)
		to0   = reflect.ValueOf(toObjPtr)
		from  = rindirect(from0)
		to    = rindirect(to0)
		root  = newParams(withOwners(c, nil, &from0, &to0, &from, &to))
	)

	functorLog("from.type: %v | input: %v", typfmtv(&from), typfmtv(&from0))
	functorLog("  to.type: %v | input: %v", typfmtv(&to), typfmtv(&to0))

	//if !to.CanAddr() {
	//	return errors.New("copy to value is unaddressable")
	//}

	err = c.copyTo(root, from, to)
	return
}

func (c *cpController) copyTo(params *Params, from, to reflect.Value) (err error) {
	err = c.copyToInternal(params, from, to,
		func(c *cpController, params *Params, from, to reflect.Value) (err error) {
			kind := from.Kind() //functorLog(" - from.type: %v", kind)
			if kind != reflect.Struct || !packageisreserved(from.Type().PkgPath()) {
				if fn, ok := copyToRoutines[kind]; ok && fn != nil {
					err = fn(c, params, from, to)
					return
				}
			}

			functorLog(" - from.type: %v - fallback to copyDefaultHandler | to.type: %v", kind, to.Type())
			err = copyDefaultHandler(c, params, from, to)
			return
		})
	return
}

func (c *cpController) copyToInternal(
	params *Params, from, to reflect.Value,
	cb func(c *cpController, params *Params, from, to reflect.Value) (err error),
) (err error) {

	// Return is from value is invalid
	if !from.IsValid() {
		return
	}

	if c.testCloneables(params, from, to) {
		return
	}

	//fromType := c.indirectType(from.Type())
	//toType := c.indirectType(to.Type())

	defer func() {
		if e := recover(); e != nil {
			err = errors.New("[recovered] copyTo unsatisfied ([%v] -> [%v]), causes: %v",
				rindirectType(from.Type()), rindirectType(to.Type()), e).
				WithData(e).
				WithTaggedData(errors.TaggedData{
					"source": from,
					"target": to,
				})
			n := log.CalcStackFrames(1)   // skip defer-recover frame at first
			log.Skip(n).Errorf("%v", err) // skip go-lib frames and defer-recover frame, back to the point throwing panic

		}
	}()

	err = cb(c, params, from, to)
	return
}

func (c *cpController) testCloneables(params *Params, from, to reflect.Value) (processed bool) {
	if from.CanInterface() {
		var fromObj interface{}
		if params != nil && params.srcOwner != nil {
			f, t := *params.srcOwner, *params.dstOwner
		retry:
			fromObj = f.Interface()
			if c.testCloneables1(params, fromObj, t) {
				return true
			}
			if k := f.Kind(); k == reflect.Ptr {
				f = f.Elem()
				if k = t.Kind(); k == reflect.Ptr {
					t = t.Elem()
				}
				goto retry
			}
		}
	}
	return
}

func (c *cpController) testCloneables1(params *Params, fromObj interface{}, to reflect.Value) (processed bool) {
	if dc, ok := fromObj.(Cloneable); ok {
		to.Set(reflect.ValueOf(dc.Clone()))
		processed = true
	} else if dc, ok := fromObj.(DeepCopyable); ok {
		to.Set(reflect.ValueOf(dc.DeepCopy()))
		processed = true
	}
	return
}

func (c *cpController) withConverters(cvt ...ValueConverter) *cpController {
	for _, cc := range cvt {
		if cc != nil {
			c.valueConverters = append(c.valueConverters, cc)
		}
	}
	return c
}

func (c *cpController) withCopiers(cvt ...ValueCopier) *cpController {
	for _, cc := range cvt {
		if cc != nil {
			c.valueCopiers = append(c.valueCopiers, cc)
		}
	}
	return c
}

func (c *cpController) withFlags(flags ...CopyMergeStrategy) *cpController {
	if c.flags == nil {
		c.flags = newFlags(flags...)
	} else {
		c.flags.withFlags(flags...)
	}
	return c
}

func (c *cpController) isIgnoreName(name string) (yes bool) {
	for _, x := range c.ignoreNames {
		if yes = isWildMatch(name, x); yes {
			break
		}
	}
	return
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

//// decode decodes a value to its underlying type, if it was wrapped by
//// interface{} or pointer.
////
////    var b = 11
////    var i interface{} = &b
////    var v = reflect.ValueOf(&i)
////    var n = c.decode(v)
////    println(n.Type())    // = int
//func (c *cpController) decode(reflectValue reflect.Value) (ret, prev reflect.Value) {
//	return rdecode(reflectValue)
//}
//
//func (c *cpController) skip(reflectValue reflect.Value, kinds ...reflect.Kind) (ret, prev reflect.Value) {
//	return rskip(reflectValue, kinds...)
//}
//
//func (c *cpController) want(reflectValue reflect.Value, kinds ...reflect.Kind) reflect.Value {
//	return rwant(reflectValue, kinds...)
//}

//func (c *cpController) indirect(reflectValue reflect.Value) reflect.Value {
//	return rindirect(reflectValue)
//}

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
//
//func (c *cpController) indirectType(reflectType reflect.Type) reflect.Type {
//	return rindirectType(reflectType)
//}
