package evendeep

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"

	"github.com/hedzr/evendeep/flags"
	"github.com/hedzr/evendeep/flags/cms"
	"github.com/hedzr/evendeep/internal/dbglog"
	"github.com/hedzr/evendeep/internal/syscalls"
	"github.com/hedzr/evendeep/internal/tool"
	"github.com/hedzr/evendeep/typ"
	"github.com/hedzr/log"

	"gopkg.in/hedzr/errors.v3"
)

const timeConstString = "time"

// RegisterDefaultConverters registers the ValueConverter list into
// default converters registry.
//
// It takes effects on DefaultCopyController, MakeClone, DeepCopy,
// and New, ....
func RegisterDefaultConverters(ss ...ValueConverter) {
	defValueConverters = append(defValueConverters, ss...)
	initGlobalOperators()
}

// RegisterDefaultCopiers registers the ValueCopier list into
// default copiers registry.
//
// It takes effects on DefaultCopyController, MakeClone, DeepCopy,
// and New, ....
func RegisterDefaultCopiers(ss ...ValueCopier) {
	defValueCopiers = append(defValueCopiers, ss...)
	initGlobalOperators()
}

func initConverters() {
	dbglog.Log("initializing default converters and copiers ...")
	defValueConverters = ValueConverters{ // Transform()
		&fromStringConverter{},
		&toStringConverter{},

		// &toFuncConverter{},
		&fromFuncConverter{},

		&toDurationConverter{},
		&fromDurationConverter{},
		&toTimeConverter{},
		&fromTimeConverter{},

		&fromBytesBufferConverter{},
		&fromMapConverter{},
	}
	defValueCopiers = ValueCopiers{ // CopyTo()
		&fromStringConverter{},
		&toStringConverter{},

		&toFuncConverter{},
		&fromFuncConverter{},

		&toDurationConverter{},
		&fromDurationConverter{},
		&toTimeConverter{},
		&fromTimeConverter{},

		&fromBytesBufferConverter{},
		&fromMapConverter{},
	}

	lenValueConverters = len(defValueConverters)
	lenValueCopiers = len(defValueCopiers)
}

var defValueConverters ValueConverters
var defValueCopiers ValueCopiers
var lenValueConverters, lenValueCopiers int

func defaultValueConverters() ValueConverters { return defValueConverters }
func defaultValueCopiers() ValueCopiers       { return defValueCopiers }

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

//

func (valueConverters ValueConverters) findConverters(params *Params, from, to reflect.Type, userDefinedOnly bool) (converter ValueConverter, ctx *ValueConverterContext) {
	var yes bool
	var min int
	if userDefinedOnly {
		min = lenValueConverters
	}
	for i := len(valueConverters) - 1; i >= min; i-- {
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

func (valueCopiers ValueCopiers) findCopiers(params *Params, from, to reflect.Type, userDefinedOnly bool) (copier ValueCopier, ctx *ValueConverterContext) {
	var yes bool
	var min int
	if userDefinedOnly {
		min = lenValueCopiers
	}
	for i := len(valueCopiers) - 1; i >= min; i-- {
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

//

// IsCopyFunctionResultToTarget does SAFELY test if copyFunctionResultToTarget is true or not.
func (ctx *ValueConverterContext) IsCopyFunctionResultToTarget() bool {
	if ctx == nil || ctx.Params == nil || ctx.controller == nil {
		return false
	}
	return ctx.controller.copyFunctionResultToTarget
}

// IsPassSourceToTargetFunction does SAFELY test if passSourceAsFunctionInArgs is true or not.
func (ctx *ValueConverterContext) IsPassSourceToTargetFunction() bool {
	if ctx == nil || ctx.Params == nil || ctx.controller == nil {
		return false
	}
	return ctx.controller.passSourceAsFunctionInArgs
}

// Preprocess find out a converter to transform source to target.
// If no comfortable converter found, the return processed is false.
func (ctx *ValueConverterContext) Preprocess(source reflect.Value, targetType reflect.Type, cvtOuter ValueConverter) (processed bool, target reflect.Value, err error) {
	if ctx != nil && ctx.Params != nil && ctx.Params.controller != nil {
		sourceType := source.Type()
		if cvt, ctxCvt := ctx.controller.valueConverters.findConverters(ctx.Params, sourceType, targetType, false); cvt != nil && cvt != cvtOuter {
			target, err = cvt.Transform(ctxCvt, source, targetType)
			processed = true
			return
		}
	}
	return
}

//nolint:unused //future
func (ctx *ValueConverterContext) isStruct() bool {
	if ctx == nil {
		return false
	}
	return ctx.Params.isStruct()
}

//nolint:unused //future
func (ctx *ValueConverterContext) isFlagExists(ftf cms.CopyMergeStrategy) (ret bool) {
	if ctx == nil {
		return
	}
	return ctx.Params.isFlagExists(ftf)
}

// isGroupedFlagOK tests if the given flag is exists or valid.
//
// Different with isGroupedFlagOKDeeply is, isGroupedFlagOK will return
// false simply while Params.fieldTags is empty or unset.
//
// When Params.fieldTags is valid, the actual testing will be forwarded
// to Params.fieldTags.flags.isGroupedFlagOK().
//
func (ctx *ValueConverterContext) isGroupedFlagOK(ftf ...cms.CopyMergeStrategy) (ret bool) { //nolint:unused //future
	if ctx == nil {
		return flags.New().IsGroupedFlagOK(ftf...)
	}
	return ctx.Params.isGroupedFlagOK(ftf...)
}

// isGroupedFlagOKDeeply tests if the given flag is exists or valid.
//
// Different with isGroupedFlagOK is, isGroupedFlagOKDeeply will check
// whether the given flag is a leader (i.e. default choice) in a group
// or not, even if Params.fieldTags is empty or unset.
//
// When Params.fieldTags is valid, the actual testing will be forwarded
// to Params.fieldTags.flags.isGroupedFlagOK().
func (ctx *ValueConverterContext) isGroupedFlagOKDeeply(ftf ...cms.CopyMergeStrategy) (ret bool) {
	if ctx == nil {
		return flags.New().IsGroupedFlagOK(ftf...)
	}
	return ctx.Params.isGroupedFlagOKDeeply(ftf...)
}

//nolint:unused //future
func (ctx *ValueConverterContext) isAnyFlagsOK(ftf ...cms.CopyMergeStrategy) (ret bool) {
	if ctx == nil {
		return flags.New().IsAnyFlagsOK(ftf...)
	}
	return ctx.Params.isAnyFlagsOK(ftf...)
}

//nolint:unused //future
func (ctx *ValueConverterContext) isAllFlagsOK(ftf ...cms.CopyMergeStrategy) (ret bool) {
	if ctx == nil {
		return flags.New().IsAllFlagsOK(ftf...)
	}
	return ctx.Params.isAllFlagsOK(ftf...)
}

//

//

//

type cvtbase struct{}

func (c *cvtbase) safeType(tgt, tgtptr reflect.Value) reflect.Type {
	if tgt.IsValid() {
		return tgt.Type()
	}
	if tgtptr.IsValid() {
		if tgtptr.Kind() == reflect.Interface {
			return tgtptr.Type()
		}
		return tgtptr.Type().Elem()
	}
	log.Panicf("niltyp !! CANNOT fetch type: tgt = %v, tgtptr = %v", tool.Typfmtv(&tgt), tool.Typfmtv(&tgtptr))
	return tool.Niltyp
}

// processUnexportedField try to set newval into target if it's an unexported field
func (c *cvtbase) processUnexportedField(ctx *ValueConverterContext, target, newval reflect.Value) (processed bool) {
	if ctx == nil || ctx.Params == nil {
		return
	}
	processed = ctx.Params.processUnexportedField(target, newval)
	return
}

func (c *cvtbase) checkSource(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, processed bool) {
	if ctx == nil {
		return
	}

	if processed = ctx.isGroupedFlagOKDeeply(cms.Ignore); processed {
		return
	}
	if processed = tool.IsNil(source) && ctx.isGroupedFlagOKDeeply(cms.OmitIfNil, cms.OmitIfEmpty); processed {
		target = reflect.Zero(targetType)
		return
	}
	if processed = tool.IsZero(source) && ctx.isGroupedFlagOKDeeply(cms.OmitIfZero, cms.OmitIfEmpty); processed {
		target = reflect.Zero(targetType)
	}
	return
}

func (c *cvtbase) checkTarget(ctx *ValueConverterContext, target reflect.Value, targetType reflect.Type) (processed bool) {
	if processed = !target.IsValid(); processed {
		return
	}
	if processed = tool.IsNil(target) && ctx.isGroupedFlagOKDeeply(cms.OmitIfTargetNil); processed {
		return
	}
	processed = tool.IsZero(target) && ctx.isGroupedFlagOKDeeply(cms.OmitIfTargetZero)
	return
}

//

type toConverterBase struct{ cvtbase }

func (c *toConverterBase) fallback(target reflect.Value) (err error) {
	tgtType := reflect.TypeOf((*time.Duration)(nil)).Elem()
	tool.Rindirect(target).Set(reflect.Zero(tgtType))
	return
}

//

type fromConverterBase struct{ cvtbase }

func (c *fromConverterBase) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	panic("not impl")
}

func (c *fromConverterBase) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	panic("not impl")
}

func (c *fromConverterBase) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	panic("not impl")
}

//nolint:unused //future
func (c *fromConverterBase) preprocess(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (processed bool, target reflect.Value, err error) {
	if ctx != nil && ctx.Params != nil && ctx.Params.controller != nil {
		sourceType := source.Type()
		if cvt, ctxCvt := ctx.controller.valueConverters.findConverters(ctx.Params, sourceType, targetType, false); cvt != nil {
			if cvt == c {
				return
			}
			if cc, ok := cvt.(*fromConverterBase); ok && cc == c {
				return
			}

			target, err = cvt.Transform(ctxCvt, source, targetType)
			processed = true
			return
		}
	}
	return
}

func (c *fromConverterBase) postCopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	// nolint:gocritic //no
	// if source.IsValid() {
	//	if canConvert(&source, target.Type()) {
	//		nv := source.Convert(target.Type())
	//		target.Set(nv)
	//		return
	//		//} else {
	//		//	nv := fmt.Sprintf("%v", source.Interface())
	//		//	target.Set(reflect.ValueOf(nv))
	//	}
	// }
	//
	// target = reflect.Zero(target.Type())
	// return
	var nv reflect.Value
	nv, err = c.convertToOrZeroTarget(ctx, source, target.Type())
	if err == nil {
		if target.CanSet() {
			dbglog.Log("    postCopyTo: set nv(%v) into target (%v)", tool.Valfmt(&nv), tool.Valfmt(&target))
			target.Set(nv)
		} else {
			err = ErrCannotSet.FormatWith(tool.Valfmt(&target), tool.Typfmtv(&target), tool.Valfmt(&nv), tool.Typfmtv(&nv))
		}
	}
	return
}

func (c *fromConverterBase) convertToOrZeroTarget(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	if tool.CanConvert(&source, targetType) {
		nv := source.Convert(targetType)
		target = nv
	} else if ctx.isGroupedFlagOKDeeply(cms.ClearIfInvalid) {
		target = reflect.Zero(targetType)
	}
	return
}

//

//

//

type toStringConverter struct{ toConverterBase }

func (c *toStringConverter) postCopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	if source.IsValid() {
		if tool.CanConvert(&source, target.Type()) {
			nv := source.Convert(target.Type())
			if c.processUnexportedField(ctx, target, nv) {
				return
			}
			target.Set(nv)
		} else {
			newVal := fmt.Sprintf("%v", source.Interface())
			nv := reflect.ValueOf(newVal)
			if c.processUnexportedField(ctx, target, nv) {
				return
			}
			target.Set(nv)
		}
	} else {
		target = reflect.Zero(target.Type())
	}
	return
}

func (c *toStringConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	tgt, tgtptr := tool.Rdecode(target)
	tgtType := c.safeType(tgt, tgtptr) // because tgt might be invalid, so we fetch tgt type via its pointer
	dbglog.Log("  target: %v (%v), tgtptr: %v, tgt: %v, tgttyp: %v", tool.Typfmtv(&target), tool.Typfmt(target.Type()), tool.Typfmtv(&tgtptr), tool.Typfmtv(&tgt), tool.Typfmt(tgtType))

	if processed := c.checkTarget(ctx, tgt, tgtType); processed {
		return
	}

	if ret, e := c.Transform(ctx, source, tgtType); e == nil {
		if c.processUnexportedField(ctx, target, ret) {
			return
		}
		dbglog.Log("     set: %v (%v) <- %v", tool.Valfmt(&target), tool.Typfmtv(&target), tool.Valfmt(&ret))
		tgtptr.Set(ret)
	} else {
		err = c.postCopyTo(ctx, source, target)
	}
	return
}

// Transform will transform source type (bool, int, ...) to target string
func (c *toStringConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	if source.IsValid() {
		// nolint:gocritic //no
		// var processed bool
		// if processed, target, err = ctx.Preprocess(source, targetType, c); processed {
		//	return
		// }

		var processed bool
		if target, processed = c.checkSource(ctx, source, targetType); processed {
			return
		}

		var str string
		if str, processed, err = tryMarshalling(source); processed {
			if err == nil {
				target = reflect.ValueOf(str)
			}
			return
		}

		target, err = rToString(source, targetType)
	} else if ctx == nil || ctx.isGroupedFlagOKDeeply(cms.ClearIfInvalid) {
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

//

var marshallableTypes = map[string]reflect.Type{
	// "MarshalBinary": reflect.TypeOf((*encoding.BinaryMarshaler)(nil)).Elem(),
	"MarshalText": reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem(),
	"MarshalJSON": reflect.TypeOf((*json.Marshaler)(nil)).Elem(),
}

var textMarshaller = TextMarshaller(func(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
})

func canMarshalling(source reflect.Value) (mtd reflect.Value, yes bool) {
	st := source.Type()
	for fnn, t := range marshallableTypes {
		if st.Implements(t) {
			yes, mtd = true, source.MethodByName(fnn)
			break
		}
	}
	return
}

// FallbackToBuiltinStringMarshalling exposes the builtin string
// marshalling mechanism for your customized ValueConverter or
// ValueCopier.
func FallbackToBuiltinStringMarshalling(source reflect.Value) (str string, err error) {
	return doMarshalling(source)
}

func doMarshalling(source reflect.Value) (str string, err error) {
	var data []byte
	if mtd, yes := canMarshalling(source); yes {
		ret := mtd.Call(nil)
		if err, yes = (ret[1].Interface()).(error); err == nil && yes {
			data = ret[0].Interface().([]byte) //nolint:errcheck //no need
		}
	} else {
		data, err = textMarshaller(source.Interface())
	}
	if err == nil {
		str = string(data)
	}
	return
}

func tryMarshalling(source reflect.Value) (str string, processed bool, err error) {
	var data []byte
	var mtd reflect.Value
	if mtd, processed = canMarshalling(source); processed {
		ret := mtd.Call(nil)
		if err, _ = (ret[1].Interface()).(error); err == nil {
			data = ret[0].Interface().([]byte) //nolint:errcheck //no need
		}
	}
	if err == nil {
		str = string(data)
	}
	return
}

//

type fromStringConverter struct{ fromConverterBase }

func (c *fromStringConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	tgt, tgtptr := tool.Rdecode(target)
	tgttyp := c.safeType(tgt, tgtptr) // because tgt might be invalid so we fetch tgt type via its pointer
	dbglog.Log("  target: %v (%v), tgtptr: %v, tgt: %v, tgttyp: %v", tool.Typfmtv(&target), tool.Typfmt(target.Type()), tool.Typfmtv(&tgtptr), tool.Typfmtv(&tgt), tool.Typfmt(tgttyp))

	if processed := c.checkTarget(ctx, tgt, tgttyp); processed {
		// nolint:gocritic //no
		// target.Set(ret)
		return
	}

	if ret, e := c.Transform(ctx, source, tgttyp); e == nil {
		if tgtptr.Kind() == reflect.Interface { //nolint:gocritic // no need to switch to 'switch' clause
			tgtptr.Set(ret)
		} else if tgtptr.Kind() == reflect.Ptr {
			tgtptr.Elem().Set(ret)
		} else if tgt.CanSet() {
			tgt.Set(ret)
		} else {
			err = ErrCannotSet.FormatWith(tool.Valfmt(&tgt), tool.Typfmtv(&tgt), tool.Valfmt(&ret), tool.Typfmtv(&ret))
		}
		dbglog.Log("  tgt: %v (ret = %v)", tool.Valfmt(&tgt), tool.Valfmt(&ret))
	} else if !errors.Is(e, strconv.ErrSyntax) && !errors.Is(e, strconv.ErrRange) {
		dbglog.Log("  Transform() failed: %v", e)
		dbglog.Log("  try running postCopyTo()")
		err = c.postCopyTo(ctx, source, target)
	}
	return
}

// Transform will transform source string to target type (bool, int, ...)
func (c *fromStringConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	if source.IsValid() {
		// nolint:gocritic //no
		// var processed bool
		// if processed, target, err = c.preprocess(ctx, source, targetType); processed {
		//	return
		// }

		var processed bool
		if target, processed = c.checkSource(ctx, source, targetType); processed {
			return
		}

		switch k := targetType.Kind(); k { //nolint:exhaustive //no need
		case reflect.Bool:
			target = rToBool(source)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			target, err = rToInteger(source, targetType)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			target, err = rToUInteger(source, targetType)

		case reflect.Uintptr:
			target = rToUIntegerHex(source, targetType)
		// case reflect.UnsafePointer:
		//	// target = rToUIntegerHex(source, targetType)
		//	err = errors.InvalidArgument
		// case reflect.Ptr:
		//	//target = rToUIntegerHex(source, targetType)
		//	err = errors.InvalidArgument

		case reflect.Float32, reflect.Float64:
			target, err = rToFloat(source, targetType)
		case reflect.Complex64, reflect.Complex128:
			target, err = rToComplex(source, targetType)

		case reflect.String:
			target = source

		// nolint:gocritic //no
		// reflect.Array
		// reflect.Chan
		// reflect.Func
		// reflect.Interface
		// reflect.Map
		// reflect.Slice
		// reflect.Struct

		default:
			target, err = c.convertToOrZeroTarget(ctx, source, targetType)
		}
	} else {
		target, err = c.convertToOrZeroTarget(ctx, source, targetType)
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

// fromMapConverter transforms a map to other types (esp string, slice, struct)
type fromMapConverter struct{ fromConverterBase }

func (c *fromMapConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	tgt, tgtptr := tool.Rdecode(target)
	tgttyp := c.safeType(tgt, tgtptr) // because tgt might be invalid so we fetch tgt type via its pointer
	dbglog.Log("  target: %v (%v), tgtptr: %v, tgt: %v, tgttyp: %v", tool.Typfmtv(&target), tool.Typfmt(target.Type()), tool.Typfmtv(&tgtptr), tool.Typfmtv(&tgt), tool.Typfmt(tgttyp))

	if processed := c.checkTarget(ctx, tgt, tgttyp); processed {
		// nolint:gocritic //no
		// target.Set(ret)
		return
	}

	if ctx.controller.targetSetter != nil {
		if tgttyp.Kind() == reflect.Struct {
			// DON'T get into c.Transform(), because Transform will modify
			// a temporary new instance and return it to caller, and
			// the new instance will be set to 'target'.
			// When target setter is valid, we assume the setter will
			// modify the real target directly rather than on a temporary
			// object.
			err = c.toStructDirectly(ctx, source, target, tgttyp)
			return
		}
	}

	if ret, e := c.Transform(ctx, source, tgttyp); e == nil {
		if k := tgtptr.Kind(); k == reflect.Interface { //nolint:gocritic // no need to switch to 'switch' clause
			tgtptr.Set(ret)
		} else if k == reflect.Ptr {
			tgtptr.Elem().Set(ret)
			// } else if tool.IsZero(tgt) {
		} else if tgt.CanSet() {
			tgt.Set(ret)
		} else {
			err = ErrCannotSet.FormatWith(tool.Valfmt(&tgt), tool.Typfmtv(&tgt), tool.Valfmt(&ret), tool.Typfmtv(&ret))
		}
		dbglog.Log("  tgt: %v (ret = %v)", tool.Valfmt(&tgt), tool.Valfmt(&ret))
	} else if !errors.Is(e, strconv.ErrSyntax) && !errors.Is(e, strconv.ErrRange) {
		dbglog.Log("  Transform() failed: %v", e)
		dbglog.Log("  try running postCopyTo()")
		err = c.postCopyTo(ctx, source, target)
	}
	return
}

// Transform will transform source string to target type (bool, int, ...)
func (c *fromMapConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	if source.IsValid() {
		var processed bool
		if target, processed = c.checkSource(ctx, source, targetType); processed {
			return
		}

		switch k := targetType.Kind(); k { //nolint:exhaustive //no need
		case reflect.String:
			var str string
			if str, err = doMarshalling(source); err == nil {
				target = reflect.ValueOf(str)
			}

		case reflect.Struct:
			target, err = c.toStruct(ctx, source, targetType)

		// case reflect.Slice:
		// case reflect.Array:

		default:
			target, err = c.convertToOrZeroTarget(ctx, source, targetType)
		}
	} else {
		target, err = c.convertToOrZeroTarget(ctx, source, targetType)
	}
	return
}

func (c *fromMapConverter) toStructDirectly(ctx *ValueConverterContext, source, target reflect.Value, targetType reflect.Type) (err error) {
	cc := ctx.controller

	preSetter := func(value reflect.Value, names ...string) (processed bool, err error) {
		if cc.targetSetter != nil {
			if err = cc.targetSetter(&value, names...); err == nil {
				processed = true
			} else {
				if err != ErrShouldFallback { //nolint:errorlint //want it exactly
					return // has error
				}
				err, processed = nil, false
			}
		}
		return
	}

	var ec = errors.New("map -> struct errors")
	defer ec.Defer(&err)

	keys := source.MapKeys()
	for _, key := range keys {
		src := source.MapIndex(key)
		st := src.Kind()
		if st == reflect.Interface {
			src = src.Elem()
		}

		// convert map key to string type
		key, err = rToString(key, tool.StringType)
		if err != nil {
			continue // ignore non-string key
		}
		ks := key.String()
		dbglog.Log("  key %q, src: %v (%v)", ks, tool.Valfmt(&src), tool.Typfmtv(&src))

		if cc.targetSetter != nil {
			newtyp := src.Type()
			val := reflect.New(newtyp).Elem()
			err = ctx.controller.copyTo(ctx.Params, src, val)
			dbglog.Log("  nv.%q: %v (%v) ", ks, tool.Valfmt(&val), tool.Typfmtv(&val))
			var processed bool
			if processed, err = preSetter(val, ks); err != nil || processed {
				ec.Attach(err)
				continue
			}
		}
	}
	return
}

func (c *fromMapConverter) toStruct(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	cc := ctx.controller

	preSetter := func(value reflect.Value, names ...string) (processed bool, err error) {
		if cc.targetSetter != nil {
			if err = cc.targetSetter(&value, names...); err == nil {
				processed = true
			} else {
				if err != ErrShouldFallback { //nolint:errorlint //want it exactly
					return // has error
				}
				err, processed = nil, false
			}
		}
		return
	}

	var ec = errors.New("map -> struct errors")
	defer ec.Defer(&err)

	target = reflect.New(targetType).Elem()
	keys := source.MapKeys()
	for _, key := range keys {
		src := source.MapIndex(key)
		st := src.Kind()
		if st == reflect.Interface {
			src = src.Elem()
		}

		// convert map key to string type
		key, err = rToString(key, tool.StringType)
		if err != nil {
			continue // ignore non-string key
		}
		ks := key.String()
		dbglog.Log("  key %q, src: %v (%v)", ks, tool.Valfmt(&src), tool.Typfmtv(&src))

		if cc.targetSetter != nil {
			newtyp := src.Type()
			val := reflect.New(newtyp).Elem()
			err = ctx.controller.copyTo(ctx.Params, src, val)
			dbglog.Log("  nv.%q: %v (%v) ", ks, tool.Valfmt(&val), tool.Typfmtv(&val))
			var processed bool
			if processed, err = preSetter(val, ks); err != nil || processed {
				ec.Attach(err)
				continue
			}
		}

		// use the key.(string) as the target struct field name
		tsf, ok := targetType.FieldByName(ks)
		if !ok {
			continue
		}

		fld := target.FieldByName(ks)
		// nolint:gocritic //no
		// dbglog.Log("  fld %q: ", ks)
		tsft := tsf.Type
		tsfk := tsft.Kind()
		if tsfk == reflect.Interface {
			// nolint:gocritic //no
			// tsft = tsft.Elem()
			fld = fld.Elem()
		} else if tsfk == reflect.Ptr {
			dbglog.Log("  fld.%q: %v (%v)", ks, tool.Valfmt(&fld), tool.Typfmtv(&fld))
			if fld.IsNil() {
				n := reflect.New(fld.Type().Elem())
				target.FieldByName(ks).Set(n)
				fld = target.FieldByName(ks)
			}
			// nolint:gocritic //no
			// tsft = tsft.Elem()
			fld = fld.Elem()
			dbglog.Log("  fld.%q: %v (%v)", ks, tool.Valfmt(&fld), tool.Typfmtv(&fld))
		}

		err = ctx.controller.copyTo(ctx.Params, src, fld)
		dbglog.Log("  nv.%q: %v (%v) ", ks, tool.Valfmt(&fld), tool.Typfmtv(&fld))
		ec.Attach(err)

		// nolint:gocritic //no
		// var nv reflect.Value
		// nv, err = c.fromConverterBase.convertToOrZeroTarget(ctx, src, tsft)
		// dbglog.Log("  nv.%q: %v (%v) ", ks, valfmt(&nv), typfmtv(&nv))
		// if err == nil {
		//	if fld.CanSet() {
		//		if tsfk == reflect.Ptr {
		//			n := reflect.New(fld.Type().Elem())
		//			n.Elem().Set(nv)
		//			nv = n
		//		}
		//		fld.Set(nv)
		//	} else {
		//		err = ErrCannotSet.FormatWith(valfmt(&fld), typfmtv(&fld), valfmt(&nv), typfmtv(&nv))
		//	}
		// }
	}
	dbglog.Log("  target: %v (%v) ", tool.Valfmt(&target), tool.Typfmtv(&target))
	return
}

func (c *fromMapConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	if yes = source.Kind() == reflect.Map && target.Kind() != reflect.Map; yes {
		ctx = &ValueConverterContext{params}
	}
	return
}

//

//

//

type fromBytesBufferConverter struct{ fromConverterBase }

func (c *fromBytesBufferConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	tgt, tgtptr := tool.Rdecode(target)
	tgtType := c.safeType(tgt, tgtptr) // because tgt might be invalid, so we fetch tgt type via its pointer
	// tgtType := target.Type()
	dbglog.Log(" target: %v (%v), tgtptr: %v, tgt: %v, tgttyp: %v",
		tool.Typfmtv(&target), tool.Typfmt(target.Type()),
		tool.Typfmtv(&tgtptr), tool.Typfmtv(&tgt), tool.Typfmt(tgtType))

	if processed := c.checkTarget(ctx, tgt, tgtType); processed {
		// target.Set(ret)
		return
	}

	from := source.Interface().(bytes.Buffer) //nolint:errcheck //no need
	tv := tgtptr.Interface()
	switch to := tv.(type) {
	case bytes.Buffer:
		to.Reset()
		to.Write(from.Bytes())
		// dbglog.Log("     to: %v", to.String())
	case *bytes.Buffer:
		to.Reset()
		to.Write(from.Bytes())
		// dbglog.Log("    *to: %v", to.String())
	case *[]byte:
		tgtptr.Elem().Set(reflect.ValueOf(from.Bytes()))
	case []byte:
		tgtptr.Elem().Set(reflect.ValueOf(from.Bytes()))
	}
	return
}

func (c *fromBytesBufferConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	var processed bool
	if target, processed = c.checkSource(ctx, source, targetType); processed {
		return
	}

	// TO/DO implement me
	// panic("implement me")
	from := source.Interface().(bytes.Buffer) //nolint:errcheck //no need
	var to bytes.Buffer
	_, err = to.Write(from.Bytes())
	target = reflect.ValueOf(to)
	return
}

func (c *fromBytesBufferConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	// st.PkgPath() . st.Name()
	if yes = source.Kind() == reflect.Struct && source.String() == "bytes.Buffer"; yes {
		ctx = &ValueConverterContext{params}
		dbglog.Log("    src: %v, tgt: %v | Matched", source, target)
	} else {
		dbglog.Log("    src: %v, tgt: %v", source, target)
	}
	return
}

//

//

//

type fromTimeConverter struct{ fromConverterBase }

func (c *fromTimeConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	tgt, tgtptr := tool.Rdecode(target)
	tgtType := c.safeType(tgt, tgtptr) // because tgt might be invalid, so we fetch tgt type via its pointer
	dbglog.Log("  target: %v (%v), tgtptr: %v, tgt: %v, tgttyp: %v", tool.Typfmtv(&target), tool.Typfmt(target.Type()), tool.Typfmtv(&tgtptr), tool.Typfmtv(&tgt), tool.Typfmt(tgtType))

	if processed := c.checkTarget(ctx, tgt, tgtType); processed {
		// tgtptr.Set(ret)
		return
	}

	if ret, e := c.Transform(ctx, source, tgtType); e == nil {
		if k := tgtptr.Kind(); k == reflect.Interface { //nolint:gocritic // no need to switch to 'switch' clause
			tgtptr.Set(ret)
		} else if k == reflect.Ptr {
			tgtptr.Elem().Set(ret)
		} else if tgt.CanSet() {
			tgt.Set(ret)
		} else {
			err = ErrCannotSet.FormatWith(tool.Valfmt(&tgt), tool.Typfmtv(&tgt), tool.Valfmt(&ret), tool.Typfmtv(&ret))
		}
		dbglog.Log("  tgt: %v (ret = %v)", tool.Valfmt(&tgt), tool.Valfmt(&ret))
	} else {
		dbglog.Log("  Transform() failed: %v", e)
		dbglog.Log("  trying to postCopyTo()")
		err = c.postCopyTo(ctx, source, target)
	}
	return
}

func (c *fromTimeConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	if source.IsValid() {
		var processed bool
		if target, processed = c.checkSource(ctx, source, targetType); processed {
			return
		}

		switch k := targetType.Kind(); k { //nolint:exhaustive //no need
		case reflect.Bool:
			b := tool.IsNil(source) || tool.IsZero(source)
			target = reflect.ValueOf(b)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			tm := source.Interface().(time.Time) //nolint:errcheck //no need
			t := reflect.ValueOf(tm.Unix())
			target, err = rToInteger(t, targetType)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tm := source.Interface().(time.Time) //nolint:errcheck //no need
			t := reflect.ValueOf(tm.Unix())
			target, err = rToUInteger(t, targetType)

		case reflect.Float32, reflect.Float64:
			tm := source.Interface().(time.Time) //nolint:errcheck //no need
			f := float64(tm.UnixNano()) / 1e9    //nolint:gomnd //simple case
			t := reflect.ValueOf(f)
			target, err = rToFloat(t, targetType)
		case reflect.Complex64, reflect.Complex128:
			tm := source.Interface().(time.Time) //nolint:errcheck //no need
			f := float64(tm.UnixNano()) / 1e9    //nolint:gomnd //simple case
			t := reflect.ValueOf(f)
			target, err = rToComplex(t, targetType)

		case reflect.String:
			tm := source.Interface().(time.Time) //nolint:errcheck //no need
			str := tm.Format(time.RFC3339)
			t := reflect.ValueOf(str)
			target, err = rToString(t, targetType)

		default:
			target, err = c.convertToOrZeroTarget(ctx, source, targetType)
		}
	} else {
		target, err = c.convertToOrZeroTarget(ctx, source, targetType)
	}
	return
}

func (c *fromTimeConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	if sk := source.Kind(); sk == reflect.Struct {
		if yes = source.Name() == "Time" && source.PkgPath() == timeConstString; yes {
			ctx = &ValueConverterContext{params}
		}
	}
	return
}

//

var knownTimeLayouts = []string{
	"2006-01-02 15:04:05.000000000",
	"2006-01-02 15:04:05.000000",
	"2006-01-02 15:04:05.000",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02",

	"2006-01-02 15:04:05.999999999Z07:00",
	"2006-01-02 15:04:05.999999999",
	"2006-01-02 15:04:05Z07:00",
	"2006-01-02 15:04:05",

	time.RFC3339,

	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC3339Nano,
	time.Kitchen,
	time.Stamp,
	time.StampMilli,
	time.StampMicro,
	time.StampNano,

	"01/02/2006 15:04:05.000000000",
	"01/02/2006 15:04:05.000000",
	"01/02/2006 15:04:05.000",
	"01/02/2006 15:04:05",
	"01/02/2006 15:04",
	"01/02/2006",
}

type toTimeConverter struct{ toConverterBase }

// func (c *toTimeConverter) fallback(target reflect.Value) (err error) {
//	var timeTimeTyp = reflect.TypeOf((*time.Time)(nil)).Elem()
//	rindirect(target).Set(reflect.Zero(timeTimeTyp))
//	return
// }

func (c *toTimeConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	// tgtType := target.Type()
	tgt, tgtptr := tool.Rdecode(target)
	tgtType := c.safeType(tgt, tgtptr) // because tgt might be invalid, so we fetch tgt type via its pointer
	dbglog.Log("  target: %v (%v), tgtptr: %v, tgt: %v, tgttyp: %v", tool.Typfmtv(&target), tool.Typfmt(target.Type()), tool.Typfmtv(&tgtptr), tool.Typfmtv(&tgt), tool.Typfmt(tgtType))

	if processed := c.checkTarget(ctx, tgt, tgtType); processed {
		// target.Set(ret)
		return
	}

	if ret, e := c.Transform(ctx, source, tgtType); e == nil {
		target.Set(ret)
	} else if ctx.isGroupedFlagOKDeeply(cms.ClearIfInvalid) {
		err = c.fallback(target)
	}
	return
}

func tryParseTime(s string) (tm time.Time) {
	var err error
	for _, layout := range knownTimeLayouts {
		tm, err = time.Parse(layout, s)
		if err == nil {
			return
		}
	}
	return
}

func (c *toTimeConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	if source.IsValid() { //nolint:gocritic // no need to switch to 'switch' clause
		var processed bool
		if target, processed = c.checkSource(ctx, source, targetType); processed {
			return
		}

		switch k := source.Kind(); k { //nolint:exhaustive //no need
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			tm := time.Unix(source.Int(), 0)
			target = reflect.ValueOf(tm)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tm := time.Unix(int64(source.Uint()), 0)
			target = reflect.ValueOf(tm)

		case reflect.Float32, reflect.Float64:
			sec, dec := math.Modf(source.Float())
			tm := time.Unix(int64(sec), int64(dec*(1e9)))
			target = reflect.ValueOf(tm)
		case reflect.Complex64, reflect.Complex128:
			sec, dec := math.Modf(real(source.Complex()))
			tm := time.Unix(int64(sec), int64(dec*(1e9)))
			target = reflect.ValueOf(tm)

		case reflect.String:
			tm := tryParseTime(source.String())
			target = reflect.ValueOf(tm)

		default:
			err = ErrCannotConvertTo.FormatWith(source, tool.Typfmtv(&source), targetType, targetType.Kind())
		}
	} else if ctx.isGroupedFlagOKDeeply(cms.ClearIfInvalid) {
		target = reflect.Zero(targetType)
	} else {
		err = errors.New("source (%v) is invalid", tool.Valfmt(&source))
	}
	return
}

func (c *toTimeConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	if tk := target.Kind(); tk == reflect.Struct {
		if yes = target.Name() == "Time" && target.PkgPath() == timeConstString; yes {
			ctx = &ValueConverterContext{params}
		}
	}
	return
}

//

//

//

type fromDurationConverter struct{ fromConverterBase }

func (c *fromDurationConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	tgt, tgtptr := tool.Rdecode(target)
	tgttyp := c.safeType(tgt, tgtptr) // because tgt might be invalid, so we fetch tgt type via its pointer
	dbglog.Log("  target: %v (%v), tgtptr: %v, tgt: %v, tgttyp: %v", tool.Typfmtv(&target), tool.Typfmt(target.Type()), tool.Typfmtv(&tgtptr), tool.Typfmtv(&tgt), tool.Typfmt(tgttyp))

	var processed bool
	if target, processed = c.checkSource(ctx, source, tgttyp); processed {
		return
	}

	if ret, e := c.Transform(ctx, source, tgttyp); e == nil {
		if tgtptr.Kind() == reflect.Interface { //nolint:gocritic // no need to switch to 'switch' clause
			tgtptr.Set(ret)
		} else if tgtptr.Kind() == reflect.Ptr {
			tgtptr.Elem().Set(ret)
		} else if tgt.CanSet() {
			tgt.Set(ret)
		} else {
			err = ErrCannotSet.FormatWith(tool.Valfmt(&tgt), tool.Typfmtv(&tgt), tool.Valfmt(&ret), tool.Typfmtv(&ret))
		}
		dbglog.Log("  tgt: %v (ret = %v)", tool.Valfmt(&tgt), tool.Valfmt(&ret))
	} else {
		dbglog.Log("  Transform() failed: %v", e)
		dbglog.Log("  trying to postCopyTo()")
		err = c.postCopyTo(ctx, source, target)
	}
	return
}

func (c *fromDurationConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	if source.IsValid() {
		// var processed bool
		// if processed, target, err = c.preprocess(ctx, source, targetType); processed {
		//	return
		// }

		var processed bool
		if target, processed = c.checkSource(ctx, source, targetType); processed {
			return
		}

		switch k := targetType.Kind(); k { //nolint:exhaustive //no need
		case reflect.Bool:
			target = rToBool(source)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			target, err = rToInteger(source, targetType)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			target, err = rToUInteger(source, targetType)

		case reflect.Uintptr:
			target = rToUIntegerHex(source, targetType)
		// case reflect.UnsafePointer:
		//	// target = rToUIntegerHex(source, targetType)
		//	err = errors.InvalidArgument
		// case reflect.Ptr:
		//	//target = rToUIntegerHex(source, targetType)
		//	err = errors.InvalidArgument

		case reflect.Float32, reflect.Float64:
			target, err = rToFloat(source, targetType)
		case reflect.Complex64, reflect.Complex128:
			target, err = rToComplex(source, targetType)

		case reflect.String:
			target, err = rToString(source, targetType)

		// reflect.Array
		// reflect.Chan
		// reflect.Func
		// reflect.Interface
		// reflect.Map
		// reflect.Slice
		// reflect.Struct

		default:
			target, err = c.convertToOrZeroTarget(ctx, source, targetType)
		}
	} else {
		target, err = c.convertToOrZeroTarget(ctx, source, targetType)
	}
	return
}

func (c *fromDurationConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	if sk := source.Kind(); sk == reflect.Int64 {
		if yes = source.Name() == "Duration" && source.PkgPath() == timeConstString; yes {
			ctx = &ValueConverterContext{params}
		}
	}
	return
}

//

//

//

type toDurationConverter struct{ toConverterBase }

func (c *toDurationConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	// tgtType := target.Type()
	tgt, tgtptr := tool.Rdecode(target)
	tgtType := c.safeType(tgt, tgtptr) // because tgt might be invalid, so we fetch tgt type via its pointer
	dbglog.Log("  target: %v (%v), tgtptr: %v, tgt: %v, tgttyp: %v", tool.Typfmtv(&target), tool.Typfmt(target.Type()), tool.Typfmtv(&tgtptr), tool.Typfmtv(&tgt), tool.Typfmt(tgtType))

	if processed := c.checkTarget(ctx, tgt, tgtType); processed {
		// tgtptr.Set(ret)
		return
	}

	if ret, e := c.Transform(ctx, source, tgtType); e == nil {
		target.Set(ret)
	} else if ctx.isGroupedFlagOKDeeply(cms.ClearIfInvalid) {
		err = c.fallback(target)
	}
	return
}

func (c *toDurationConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	if source.IsValid() { //nolint:gocritic // no need to switch to 'switch' clause
		var processed bool
		if target, processed = c.checkSource(ctx, source, targetType); processed {
			return
		}

		switch k := source.Kind(); k { //nolint:exhaustive //no need
		case reflect.Bool:
			if source.Bool() {
				target = reflect.ValueOf(1 * time.Nanosecond)
			} else {
				target = reflect.ValueOf(0 * time.Second)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			target = reflect.ValueOf(time.Duration(source.Int()))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			target = reflect.ValueOf(time.Duration(int64(source.Uint())))

		case reflect.Uintptr:
			target = reflect.ValueOf(time.Duration(int64(syscalls.UintptrToUint(source.Pointer()))))
		// case reflect.UnsafePointer:
		//	// target = rToUIntegerHex(source, targetType)
		//	err = errors.InvalidArgument
		// case reflect.Ptr:
		//	//target = rToUIntegerHex(source, targetType)
		//	err = errors.InvalidArgument

		case reflect.Float32, reflect.Float64:
			target = reflect.ValueOf(time.Duration(int64(source.Float())))
		case reflect.Complex64, reflect.Complex128:
			target = reflect.ValueOf(time.Duration(int64(real(source.Complex()))))

		case reflect.String:
			var dur time.Duration
			dur, err = time.ParseDuration(source.String())
			if err == nil {
				target = reflect.ValueOf(dur)
			}

		// reflect.Array
		// reflect.Chan
		// reflect.Func
		// reflect.Interface
		// reflect.Map
		// reflect.Slice
		// reflect.Struct

		default:
			err = ErrCannotConvertTo.FormatWith(source, tool.Typfmtv(&source), targetType, targetType.Kind())
		}
	} else if ctx.isGroupedFlagOKDeeply(cms.ClearIfInvalid) {
		target = reflect.Zero(targetType)
	} else {
		err = errors.New("source (%v) is invalid", tool.Valfmt(&source))
	}
	return
}

func (c *toDurationConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	if tk := target.Kind(); tk == reflect.Int64 {
		if yes = target.Name() == "Duration" && target.PkgPath() == timeConstString; yes {
			ctx = &ValueConverterContext{params}
		}
	}
	return
}

//

//

type toFuncConverter struct{ fromConverterBase }

func copyToFuncImpl(controller *cpController, source, target reflect.Value, targetType reflect.Type) (err error) {
	var presets []typ.Any
	if controller != nil {
		presets = controller.funcInputs
	}
	if targetType.NumIn() == len(presets)+1 {
		var args []reflect.Value
		for _, in := range presets {
			args = append(args, reflect.ValueOf(in))
		}
		args = append(args, source)

		res := target.Call(args)
		if len(res) > 0 {
			last := res[len(res)-1]
			if tool.Iserrortype(targetType.Out(len(res)-1)) && !tool.IsNil(last) {
				err = last.Interface().(error) //nolint:errcheck //no need
			}
		}
	}
	return
}

// processUnexportedField try to set newval into target if it's an unexported field
func (c *toFuncConverter) processUnexportedField(ctx *ValueConverterContext, target, newval reflect.Value) (processed bool) {
	if ctx == nil || ctx.Params == nil {
		return
	}
	processed = ctx.Params.processUnexportedField(target, newval)
	return
}

func (c *toFuncConverter) copyTo(ctx *ValueConverterContext, source, src, tgt, tsetter reflect.Value) (err error) {
	if ctx.isGroupedFlagOKDeeply(cms.Ignore) {
		return
	}

	tgttyp := tgt.Type()
	dbglog.Log("  copyTo: src: %v, tgt: %v,", tool.Typfmtv(&src), tool.Typfmt(tgttyp))

	if k := src.Kind(); k != reflect.Func && ctx.IsPassSourceToTargetFunction() {
		var controller *cpController
		if ctx.Params != nil && ctx.controller != nil {
			controller = ctx.controller
		}
		err = copyToFuncImpl(controller, source, tgt, tgttyp)
	} else if k == reflect.Func {
		if !c.processUnexportedField(ctx, tgt, src) {
			tsetter.Set(src)
		}
		dbglog.Log("    function pointer copied: %v (%v) -> %v", source.Kind(), source.Interface(), tgt.Kind())
	}
	return
}

func (c *toFuncConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	src := tool.Rdecodesimple(source)
	tgt, tgtptr := tool.Rdecode(target)
	tgtType := c.safeType(tgt, tgtptr) // because tgt might be invalid, so we fetch tgt type via its pointer
	// Log("  CopyTo: src: %v, tgt: %v,", typfmtv(&src), typfmt(tgtType))
	dbglog.Log("  target: %v (%v), tgtptr: %v, tgt: %v, tgttyp: %v", tool.Typfmtv(&target), tool.Typfmt(target.Type()), tool.Typfmtv(&tgtptr), tool.Typfmtv(&tgt), tool.Typfmt(tgtType))

	if processed := c.checkTarget(ctx, tgt, tgtType); processed {
		// tgtptr.Set(ret)
		return
	}

	err = c.copyTo(ctx, source, src, tgt, tgtptr)
	return
}

// func (c *toFuncConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
//
//	target = reflect.New(targetType).Elem()
//
//	src := rdecodesimple(source)
//	tgt, tgtptr := rdecode(target)
//
//	var processed bool
//	if target, processed = c.checkSource(ctx, source, targetType); processed {
//		return
//	}
//
//	err = c.copyTo(ctx, source, src, tgt, tgtptr)
//	return
// }

func (c *toFuncConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	if tk := target.Kind(); tk == reflect.Func {
		yes, ctx = true, &ValueConverterContext{params}
	}
	return
}

//

type fromFuncConverter struct{ fromConverterBase }

func (c *fromFuncConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	// tsetter might not be equal to tgt when:
	//    target represents -> (ptr - interface{} - bool)
	// such as:
	//    var a interface{} = true
	//    var target = reflect.ValueOf(&a)
	//    tgt, tsetter := rdecodesimple(target), rindirect(target)
	//    assertNotEqual(tgt, tsetter)
	//    // in this case, tsetter represents 'a' and tgt represents
	//    // 'decoded bool(a)'.
	//
	src := tool.Rdecodesimple(source)
	tgt, tgtptr := tool.Rdecode(target)
	tgtType := c.safeType(tgt, tgtptr)
	// dbglog.Log("  CopyTo: src: %v, tgt: %v, tsetter: %v", typfmtv(&src), typfmt(tgttyp), typfmtv(&tsetter))
	dbglog.Log("  target: %v (%v), tgtptr: %v, tgt: %v, tgttyp: %v", tool.Typfmtv(&target), tool.Typfmt(target.Type()), tool.Typfmtv(&tgtptr), tool.Typfmtv(&tgt), tool.Typfmt(tgtType))

	if processed := c.checkTarget(ctx, tgt, tgtType); processed {
		// target.Set(ret)
		return
	}

	if k := tgtType.Kind(); k != reflect.Func && ctx.IsCopyFunctionResultToTarget() {
		err = c.funcResultToTarget(ctx, src, target)
		return
	} else if k == reflect.Func {
		if !c.processUnexportedField(ctx, tgt, src) {
			tgtptr.Set(src)
		}
		dbglog.Log("    function pointer copied: %v (%v) -> %v", source.Kind(), source.Interface(), target.Kind())
	}

	// if ret, e := c.Transform(ctx, src, tgttyp); e == nil {
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
	// }
	return
}

func (c *fromFuncConverter) funcResultToTarget(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
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
			ok = tool.Iserrortype(lastoutargtype)
			if ok {
				v := results[len(results)-1].Interface()
				err, _ = v.(error)
				if err != nil {
					return
				}
				results = results[:len(results)-1]
			}

			if len(results) > 0 {
				if controllerIsValid {
					// if tk := target.Kind(); tk == reflect.Ptr && isNil(target) {}
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

// // processUnexportedField try to set newval into target if it's an unexported field
// func (c *fromFuncConverter) processUnexportedField(ctx *ValueConverterContext, target, newval reflect.Value) (processed bool) {
//	if ctx == nil || ctx.Params == nil {
//		return
//	}
//	processed = ctx.Params.processUnexportedField(target, newval)
//	return
// }

func (c *fromFuncConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	var processed bool
	if target, processed = c.checkSource(ctx, source, targetType); processed {
		return
	}

	target = reflect.New(targetType).Elem()
	err = c.CopyTo(ctx, source, target)

	// src, tgt, tgttyp := rdecodesimple(source), rdecodesimple(target), rdecodetypesimple(targetType)
	// Log("  Transform: src: %v, tgt: %v", typfmtv(&src), typfmt(tgttyp))
	// if k := tgttyp.Kind(); k != reflect.Func && ctx.IsCopyFunctionResultToTarget() {
	//	target, err = c.funcResultToField(ctx, src, tgttyp)
	//	return
	//
	// } else if k == reflect.Func {
	//
	//	if c.processUnexportedField(ctx, tgt, src) {
	//		ptr := source.Pointer()
	//		target.SetPointer(unsafe.Pointer(ptr))
	//	}
	//	Log("    function pointer copied: %v (%v) -> %v", source.Kind(), source.Interface(), target.Kind())
	// }
	return
}

// func (c *fromFuncConverter) funcResultToField(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
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
// }
//
// func (c *fromFuncConverter) expandResults(ctx *ValueConverterContext, sourceType, targetType reflect.Type, results []reflect.Value) (target reflect.Value, err error) {
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
// }

func (c *fromFuncConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	if sk := source.Kind(); sk == reflect.Func {
		yes, ctx = true, &ValueConverterContext{params}
	}
	return
}
