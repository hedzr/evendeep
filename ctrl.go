package deepcopy

import (
	"github.com/hedzr/deepcopy/flags"
	"github.com/hedzr/deepcopy/flags/cms"
	"github.com/hedzr/deepcopy/internal/dbglog"
	"github.com/hedzr/deepcopy/typ"
	"github.com/hedzr/log"
	"gopkg.in/hedzr/errors.v3"
	"reflect"
)

type cpController struct {
	copyUnexportedFields       bool
	copyFunctionResultToTarget bool
	passSourceAsFunctionInArgs bool
	autoExpandStruct           bool

	makeNewClone bool        // make a new clone by copying to a fresh new object
	flags        flags.Flags // CopyMergeStrategies globally
	ignoreNames  []string    // optional ignored names with wild-matching
	funcInputs   []typ.Any   // preset input args for function invoking
	rethrow      bool        // panic when error occurs

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

	dbglog.Log("    flags: %v", c.flags)
	dbglog.Log("from.type: %v | input: %v", typfmtv(&from), typfmtv(&from0))
	dbglog.Log("  to.type: %v | input: %v", typfmtv(&to), typfmtv(&to0))

	err = c.copyTo(root, from, to)
	return
}

func (c *cpController) copyTo(params *Params, from, to reflect.Value) (err error) {
	err = c.copyToInternal(params, from, to,
		func(c *cpController, params *Params, from, to reflect.Value) (err error) {
			kind := from.Kind() //Log(" - from.type: %v", kind)
			if kind != reflect.Struct || !packageisreserved(from.Type().PkgPath()) {
				if fn, ok := copyToRoutines[kind]; ok && fn != nil {
					err = fn(c, params, from, to)
					return
				}
			}

			// source is primitive type, or in a reserved package such as time, os, ...
			dbglog.Log(" - from.type: %v - fallback to copyDefaultHandler | to.type: %v", kind, typfmtv(&to))
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
		if params.isGroupedFlagOKDeeply(cms.OmitIfEmpty, cms.OmitIfNil, cms.OmitIfZero) {
			return
		}
		// todo set target to zero
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

			// skip go-lib frames and defer-recover frame, back to the point throwing panic
			n := log.CalcStackFrames(1) // skip defer-recover frame at first

			if c.rethrow {
				log.Skip(n).Panicf("%+v", err)
			} else {
				log.Skip(n).Errorf("%+v", err)
			}
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

func (c *cpController) withFlags(flags1 ...cms.CopyMergeStrategy) *cpController {
	if c.flags == nil {
		c.flags = flags.New(flags1...)
	} else {
		c.flags.WithFlags(flags1...)
	}
	return c
}

//func (c *cpController) isIgnoreName(name string) (yes bool) {
//	for _, x := range c.ignoreNames {
//		if yes = isWildMatch(name, x); yes {
//			break
//		}
//	}
//	return
//}
