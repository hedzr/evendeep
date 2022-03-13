package deepcopy

import (
	"bytes"
	"fmt"
	"github.com/hedzr/log"
	"gopkg.in/hedzr/errors.v3"
	"reflect"
	"time"
	"unsafe"
)

func initConverters() {
	functorLog("initializing default converters and copiers ...")
	defValueConverters = ValueConverters{
		&fromStringConverter{},
		&toStringConverter{},
		&fromBytesBufferConverter{},
		&toDurationFromString{},
		&fromFuncConverter{},
	}
	defValueCopiers = ValueCopiers{
		&fromStringConverter{},
		&toStringConverter{},
		&fromBytesBufferConverter{},
		&toDurationFromString{},
		&fromFuncConverter{},
	}
}

var defValueConverters ValueConverters
var defValueCopiers ValueCopiers

func defaultValueConverters() ValueConverters { return defValueConverters }
func defaultValueCopiers() ValueCopiers       { return defValueCopiers }

func (valueConverters ValueConverters) findConverters(params *Params, from, to reflect.Type) (converter ValueConverter, ctx *ValueConverterContext) {
	var yes bool
	for i := len(valueConverters) - 1; i >= 0; i-- {
		// FILO: the last added converter has the first priority
		cvt := valueConverters[i]
		if cvt != nil {
			if ctx, yes = cvt.Match(params, from, to); yes {
				converter = cvt
				break
			}
		}
	}
	return
}

func (valueCopiers ValueCopiers) findCopiers(params *Params, from, to reflect.Type) (copier ValueCopier, ctx *ValueConverterContext) {
	var yes bool
	for i := len(valueCopiers) - 1; i >= 0; i-- {
		// FILO: the last added converter has the first priority
		cpr := valueCopiers[i]
		if cpr != nil {
			if ctx, yes = cpr.Match(params, from, to); yes {
				copier = cpr
				break
			}
		}
	}
	return
}

// ValueConverter _
type ValueConverter interface {
	Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error)
	Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool)
}

// ValueCopier _
type ValueCopier interface {
	CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error)
	Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool)
}

// NameConverter _
type NameConverter interface {
	ToGoName(ctx *NameConverterContext, fieldName string) (goName string)
	ToFieldName(ctx *NameConverterContext, goName string) (fieldName string)
}

// ValueConverters _
type ValueConverters []ValueConverter

// ValueCopiers _
type ValueCopiers []ValueCopier

// NameConverters _
type NameConverters []NameConverter

// NameConverterContext _
type NameConverterContext struct {
	*Params
}

// ValueConverterContext _
type ValueConverterContext struct {
	*Params
}

func (ctx *ValueConverterContext) IsCopyFunctionResultToTarget() bool {
	if ctx == nil || ctx.Params == nil || ctx.controller == nil {
		return false
	}
	return ctx.controller.copyFunctionResultToTarget
}

// Preprocess find out a converter to transform source to target.
// If no comfortable converter found, the return processed is false.
func (ctx *ValueConverterContext) Preprocess(source reflect.Value, targetType reflect.Type, cvtOuter ValueConverter) (processed bool, target reflect.Value, err error) {
	if ctx != nil && ctx.Params != nil && ctx.Params.controller != nil {
		sourceType := source.Type()
		if cvt, ctx := ctx.controller.valueConverters.findConverters(ctx.Params, sourceType, targetType); cvt != nil && cvt != cvtOuter {
			target, err = cvt.Transform(ctx, source, targetType)
			processed = true
			return
		}
	}
	return
}

type toStringConverter struct{}

func (c *toStringConverter) processUnexportedField(ctx *ValueConverterContext, target, newval reflect.Value) (processed bool) {
	if ctx == nil || ctx.Params == nil {
		return
	}
	processed = ctx.Params.processUnexportedField(target, newval)
	return
}

func (c *toStringConverter) postCopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	if source.IsValid() {
		if canConvert(&source, target.Type()) {
			nv := source.Convert(target.Type())
			if c.processUnexportedField(ctx, target, nv) {
				return
			}
			target.Set(nv)
		} else {
			nv := fmt.Sprintf("%v", source.Interface())
			if c.processUnexportedField(ctx, target, reflect.ValueOf(nv)) {
				return
			}
			target.Set(reflect.ValueOf(nv))
		}
	} else {
		target = reflect.Zero(target.Type())
	}
	return
}

func (c *toStringConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	tt := target.Type()
	if ret, e := c.Transform(ctx, source, tt); e == nil {
		if c.processUnexportedField(ctx, target, ret) {
			return
		}
		target.Set(ret)
	} else {
		err = c.postCopyTo(ctx, source, target)
	}
	return
}

// Transform will transform source type (bool, int, ...) to target string
func (c *toStringConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	if source.IsValid() {
		var processed bool
		if processed, target, err = ctx.Preprocess(source, targetType, c); processed {
			return
		}

		target, err = rToString(source, targetType)
	} else {
		target = reflect.Zero(reflect.TypeOf((*string)(nil)).Elem())
	}
	return
}

func (c *toStringConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	if yes = target.Kind() == reflect.String; yes {
		ctx = &ValueConverterContext{params}
	}
	return
}

type fromStringConverter struct{}

func safetyp(tgt, tgtptr reflect.Value) reflect.Type {
	if tgt.IsValid() {
		return tgt.Type()
	}
	if tgtptr.IsValid() {
		return tgtptr.Type().Elem()
	}
	log.Panicf("niltyp !! CANNOT fetch type: tgt = %v, tgtptr = %v", typfmtv(&tgt), typfmtv(&tgtptr))
	return niltyp
}

var niltyp = reflect.TypeOf((*string)(nil))

func (c *fromStringConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	tgt, tgtptr := rdecode(target)
	tgttyp := safetyp(tgt, tgtptr) // because tgt might be invalid so we fetch tgt type via its pointer
	functorLog("  target: %v (%v), tgtptr: %v, tgt: %v, tgttyp: %v", typfmtv(&target), typfmt(target.Type()), typfmtv(&tgtptr), typfmtv(&tgt), typfmt(tgttyp))

	if ret, e := c.Transform(ctx, source, tgttyp); e == nil {
		if tgtptr.Kind() == reflect.Interface {
			tgtptr.Set(ret)
		} else {
			tgtptr.Elem().Set(ret)
		}
		functorLog("  tgt: %v (ret = %v)", valfmt(&tgt), valfmt(&ret))
	} else {
		functorLog("  Transform() failed: %v", e)
		functorLog("  trying to postCopyTo()")
		err = c.postCopyTo(source, target)
	}
	return
}

func (c *fromStringConverter) postCopyTo(source, target reflect.Value) (err error) {
	if source.IsValid() {
		if canConvert(&source, target.Type()) {
			nv := source.Convert(target.Type())
			target.Set(nv)
			return
			//} else {
			//	nv := fmt.Sprintf("%v", source.Interface())
			//	target.Set(reflect.ValueOf(nv))
		}
	}

	target = reflect.Zero(target.Type())
	return
}

func (c *fromStringConverter) preprocess(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (processed bool, target reflect.Value, err error) {
	if ctx != nil && ctx.Params != nil && ctx.Params.controller != nil {
		sourceType := source.Type()
		if cvt, ctx := ctx.controller.valueConverters.findConverters(ctx.Params, sourceType, targetType); cvt != nil && cvt != c {
			target, err = cvt.Transform(ctx, source, targetType)
			processed = true
			return
		}
	}
	return
}

// Transform will transform source string to target type (bool, int, ...)
func (c *fromStringConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	if source.IsValid() {
		var processed bool
		if processed, target, err = c.preprocess(ctx, source, targetType); processed {
			return
		}

		switch k := targetType.Kind(); k {
		case reflect.Bool:
			target = rToBool(source)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			target, err = rToInteger(source, targetType)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			target, err = rToUInteger(source, targetType)

		case reflect.Uintptr:
			target = rToUIntegerHex(source, targetType)
		//case reflect.UnsafePointer:
		//	// target = rToUIntegerHex(source, targetType)
		//	err = errors.InvalidArgument
		//case reflect.Ptr:
		//	//target = rToUIntegerHex(source, targetType)
		//	err = errors.InvalidArgument

		case reflect.Float32, reflect.Float64:
			target, err = rToFloat(source, targetType)
		case reflect.Complex64, reflect.Complex128:
			target, err = rToComplex(source, targetType)

		case reflect.String:
			target = source

		//reflect.Array
		//reflect.Chan
		//reflect.Func
		//reflect.Interface
		//reflect.Map
		//reflect.Slice
		//reflect.Struct

		default:
			target, err = c.defaultTypes(source, targetType)
		}
	} else {
		target, err = c.defaultTypes(source, targetType)
	}
	return
}

func (c *fromStringConverter) defaultTypes(source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	if canConvert(&source, targetType) {
		nv := source.Convert(targetType)
		// target.Set(nv)
		target = nv
	} else {
		//nv := fmt.Sprintf("%v", source.Interface())
		//// target.Set(reflect.ValueOf(nv))
		//target = reflect.ValueOf(nv)
		target = reflect.Zero(targetType)
	}
	return
}

func (c *fromStringConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	if yes = source.Kind() == reflect.String; yes {
		ctx = &ValueConverterContext{params}
	}
	return
}

//

//

//

type fromBytesBufferConverter struct{}

func (c *fromBytesBufferConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	//TO/DO implement me
	//panic("implement me")
	from := source.Interface().(bytes.Buffer)
	var to bytes.Buffer
	to.Write(from.Bytes())
	target = reflect.ValueOf(to)
	return
}

func (c *fromBytesBufferConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	from, to := source.Interface().(bytes.Buffer), target.Interface().(bytes.Buffer)
	to.Reset()
	to.Write(from.Bytes())
	return
}

func (c *fromBytesBufferConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	//st.PkgPath() . st.Name()
	if yes = source.Kind() == reflect.Struct && source.String() == "bytes.Buffer"; yes {
		ctx = &ValueConverterContext{params}
		functorLog("    src: %v, tgt: %v | Matched", source, target)
	} else {
		//functorLog("    src: %v, tgt: %v", source, target)
	}
	return
}

//

//

//

type toDurationFromString struct {
}

func (c *toDurationFromString) fallback(target reflect.Value) (err error) {
	tgtType := reflect.TypeOf((*time.Duration)(nil)).Elem()
	rindirect(target).Set(reflect.Zero(tgtType))
	return
}

func (c *toDurationFromString) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	if ret, e := c.Transform(ctx, source, target.Type()); e == nil {
		target.Set(ret)
	} else {
		err = c.fallback(target)
	}
	return
}

func (c *toDurationFromString) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	var dur time.Duration
	dur, err = time.ParseDuration(source.String())
	if err == nil {
		target = reflect.ValueOf(dur)
	}
	return
}

func (c *toDurationFromString) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	if sk, tk := source.Kind(), target.Kind(); sk == reflect.String && tk == reflect.Int64 {
		if yes = target.Name() == "Duration" && target.PkgPath() == "time"; yes {
			ctx = &ValueConverterContext{params}
		}
	}
	return
}

type fromFuncConverter struct{}

func (c *fromFuncConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	// tsetting might not be equal to tgt when:
	//    target represents -> (ptr - interface{} - bool)
	// such as:
	//    var a interface{} = true
	//    var target = reflect.ValueOf(&a)
	//    tgt, tsetter := rdecodesimple(target), rindirect(target)
	//    assertNotEqual(tgt, tsetter)
	//    // in this case, tsetter represents 'a' and tgt represents
	//    // 'decoded bool(a)'.
	//
	src := rdecodesimple(source)
	tgt, tsetter := rdecode(target)
	tgttyp := tgt.Type()
	functorLog("  CopyTo: src: %v, tgt: %v, tsetter: %v", typfmtv(&src), typfmt(tgttyp), typfmtv(&tsetter))

	if k := tgttyp.Kind(); k != reflect.Func && ctx.IsCopyFunctionResultToTarget() {
		err = c.funcResultToTarget(ctx, src, target)
		return

	} else if k == reflect.Func {

		if c.processUnexportedField(ctx, tgt, src) {
			ptr := source.Pointer()
			target.SetPointer(unsafe.Pointer(ptr))
		}
		functorLog("    function pointer copied: %v (%v) -> %v", source.Kind(), source.Interface(), target.Kind())
	}

	//if ret, e := c.Transform(ctx, src, tgttyp); e == nil {
	//	if !target.IsValid() || isZero(target) {
	//		return errors.New("cannot set to zero or invalid target")
	//	}
	//	if canConvert(&ret, tgttyp) {
	//		nv := ret.Convert(tgttyp)
	//		if c.processUnexportedField(ctx, tgt, nv) {
	//			return
	//		}
	//		tsetter.Set(nv)
	//	}
	//}
	return
}

func (c *fromFuncConverter) funcResultToTarget(ctx *ValueConverterContext, source reflect.Value, target reflect.Value) (err error) {
	sourceType := source.Type()
	var presetInArgsLen int
	var ok bool
	var controllerIsValid = ctx != nil && ctx.Params != nil && ctx.Params.controller != nil
	if controllerIsValid {
		presetInArgsLen = len(ctx.controller.funcInputs)
	}
	if sourceType.NumIn() == presetInArgsLen {
		numOutSrc := sourceType.NumOut()
		if numOutSrc > 0 {
			srcResults := source.Call([]reflect.Value{})

			results := srcResults
			lastoutargtype := sourceType.Out(sourceType.NumOut() - 1)
			ok = iserrortype(lastoutargtype)
			if ok {
				err, ok = results[len(results)-1].Interface().(error)
				if err != nil {
					return
				}
				results = results[:len(results)-1]
			}

			if len(results) > 0 {
				if controllerIsValid {
					err = ctx.controller.copyTo(ctx.Params, results[0], target)
					return
				}

				// target, err = c.expandResults(ctx, sourceType, targetType, results)
				err = errors.New("expecting ctx.Params.controller is valid object ptr")
				return
			}
		}
	}

	err = errors.New("unmatched number of function return and preset input args: function needs %v params but preset %v input args", sourceType.NumIn(), presetInArgsLen)
	return
}

// processUnexportedField try to set newval into target if it's an unexported field
func (c *fromFuncConverter) processUnexportedField(ctx *ValueConverterContext, target, newval reflect.Value) (processed bool) {
	if ctx == nil || ctx.Params == nil {
		return
	}
	processed = ctx.Params.processUnexportedField(target, newval)
	return
}

var errtyp = reflect.TypeOf((*error)(nil)).Elem()

func iserrortype(typ reflect.Type) bool {
	return typ.Implements(errtyp)
}

func (c *fromFuncConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	target = reflect.New(targetType).Elem()
	err = c.CopyTo(ctx, source, target)

	//src, tgt, tgttyp := rdecodesimple(source), rdecodesimple(target), rdecodetypesimple(targetType)
	//functorLog("  Transform: src: %v, tgt: %v", typfmtv(&src), typfmt(tgttyp))
	//if k := tgttyp.Kind(); k != reflect.Func && ctx.IsCopyFunctionResultToTarget() {
	//	target, err = c.funcResultToField(ctx, src, tgttyp)
	//	return
	//
	//} else if k == reflect.Func {
	//
	//	if c.processUnexportedField(ctx, tgt, src) {
	//		ptr := source.Pointer()
	//		target.SetPointer(unsafe.Pointer(ptr))
	//	}
	//	functorLog("    function pointer copied: %v (%v) -> %v", source.Kind(), source.Interface(), target.Kind())
	//}
	return
}

//func (c *fromFuncConverter) funcResultToField(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
//	sourceType := source.Type()
//	var presetInArgsLen int
//	var ok bool
//	var controllerIsValid = ctx != nil && ctx.Params != nil && ctx.Params.controller != nil
//	if controllerIsValid {
//		presetInArgsLen = len(ctx.controller.funcInputs)
//	}
//	if sourceType.NumIn() == presetInArgsLen {
//		numOutSrc := sourceType.NumOut()
//		if numOutSrc > 0 {
//			srcResults := source.Call([]reflect.Value{})
//
//			results := srcResults
//			lastoutargtype := sourceType.Out(sourceType.NumOut() - 1)
//			ok = iserrortype(lastoutargtype)
//			if ok {
//				err, ok = results[len(results)-1].Interface().(error)
//				if err != nil {
//					return
//				}
//				results = results[:len(results)-1]
//			}
//
//			if len(results) > 0 {
//				// slice,map,struct
//				// scalar
//
//				target, err = c.expandResults(ctx, sourceType, targetType, results)
//			}
//		}
//	} else {
//		err = errors.New("unmatched number of function return and preset input args: function needs %v params but preset %v input args", sourceType.NumIn(), presetInArgsLen)
//	}
//	return
//}
//
//func (c *fromFuncConverter) expandResults(ctx *ValueConverterContext, sourceType, targetType reflect.Type, results []reflect.Value) (target reflect.Value, err error) {
//	//srclen := len(results)
//	switch kind := targetType.Kind(); kind {
//	case reflect.Bool:
//		target = rToBool(results[0])
//	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
//		target, err = rToInteger(results[0], targetType)
//	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
//		target, err = rToUInteger(results[0], targetType)
//	case reflect.Float32, reflect.Float64:
//		target, err = rToFloat(results[0], targetType)
//	case reflect.Complex64, reflect.Complex128:
//		target, err = rToComplex(results[0], targetType)
//	case reflect.String:
//		{
//			var processed bool
//			if processed, target, err = ctx.Preprocess(results[0], targetType, c); processed {
//				return
//			}
//		}
//		target, err = rToString(results[0], targetType)
//
//	case reflect.Array:
//		target, err = rToArray(ctx, results[0], targetType, -1)
//	case reflect.Slice:
//		target, err = rToSlice(ctx, results[0], targetType, -1)
//	case reflect.Map:
//		target, err = rToMap(ctx, results[0], sourceType, targetType)
//	case reflect.Struct:
//		target, err = rToStruct(ctx, results[0], sourceType, targetType)
//	case reflect.Func:
//		target, err = rToFunc(ctx, results[0], sourceType, targetType)
//
//	case reflect.Interface, reflect.Ptr, reflect.Chan:
//		if results[0].Type().ConvertibleTo(targetType) {
//			target = results[0].Convert(targetType)
//		} else {
//			err = errCannotConvertTo.FormatWith(typfmt(results[0].Type()), typfmt(targetType))
//		}
//
//	case reflect.UnsafePointer:
//		err = errCannotConvertTo.FormatWith(typfmt(results[0].Type()), typfmt(targetType))
//	}
//	return
//}

func (c *fromFuncConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	if sk := source.Kind(); sk == reflect.Func {
		yes, ctx = true, &ValueConverterContext{params}
	}
	return
}
