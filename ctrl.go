package evendeep

import (
	"github.com/hedzr/log"

	"github.com/hedzr/evendeep/dbglog"
	"github.com/hedzr/evendeep/flags"
	"github.com/hedzr/evendeep/flags/cms"
	"github.com/hedzr/evendeep/internal/tool"
	"github.com/hedzr/evendeep/typ"

	"gopkg.in/hedzr/errors.v3"

	"reflect"
	"unsafe"
)

type cpController struct {
	copyUnexportedFields       bool
	copyFunctionResultToTarget bool
	passSourceAsFunctionInArgs bool
	autoExpandStruct           bool // navigate into nested struct?
	autoNewStruct              bool // create new instance if field is a ptr
	tryApplyConverterAtFirst   bool // ValueConverters first, or ValueCopiers?
	wipeSlice1st               bool // wipe Slice or Map before copy/merge from source field

	makeNewClone bool        // make a new clone by copying to a fresh new object
	flags        flags.Flags // CopyMergeStrategies globally
	ignoreNames  []string    // optional ignored names with wild-matching
	funcInputs   []typ.Any   // preset input args for function invoking
	rethrow      bool        // panic when error occurs

	advanceTargetFieldPointerEvenIfSourceIgnored bool

	tagKeyName string // struct tag name for cmd.CopyMergeStrategy, default is "" and assumes using "copy" as key name

	valueConverters ValueConverters
	valueCopiers    ValueCopiers

	sourceExtractor SourceValueExtractor // simple struct field value extractor in single depth
	targetSetter    TargetValueSetter    //

	// targetOriented indicates both sourceExtractor and target object are available.
	//
	// When targetOriented is true or cms.ByName has been specified, Copier
	// do traverse on a struct with target-oriented way. That is, copier will
	// pick up a source field and the corresponding target field with same name,
	// or prefer one after name transformed.
	// See also Name Conversions.
	targetOriented bool // loop for target struct fields? default is for source.
}

// SourceValueExtractor provides a hook for handling
// the extraction from source field.
//
// SourceValueExtractor can work for non-nested struct.
type SourceValueExtractor func(targetName string) typ.Any

// TargetValueSetter provide a hook for handling the setup
// to a target field.
//
// In the TargetValueSetter you could return evendeep.ErrShouldFallback to
// call the evendeep standard processing.
//
// TargetValueSetter can work for struct and map.
//
// NOTE that the sourceNames[0] is current field name, and the whole
// sourceNames slice includes the path of the nested struct(s),
// in reversal order.
type TargetValueSetter func(value *reflect.Value, sourceNames ...string) (err error)

// CopyTo makes a deep clone of a source object or merges it into the target.
func (c *cpController) CopyTo(fromObjOrPtr, toObjPtr interface{}, opts ...Opt) (err error) {
	if fromObjOrPtr == nil || toObjPtr == nil {
		return
	}

	lazyInitRoutines()

	for _, opt := range opts {
		opt(c)
	}

	var (
		from0 = reflect.ValueOf(fromObjOrPtr)
		to0   = reflect.ValueOf(toObjPtr)
		from  = tool.Rindirect(from0)
		to    = tool.Rindirect(to0)
		root  = newParams(withOwners(c, nil, &from0, &to0, &from, &to))
	)

	dbglog.Log("          flags: %v", c.flags)
	dbglog.Log("flags (verbose): %+v", c.flags)
	dbglog.Log("      from.type: %v | input: %v", tool.Typfmtv(&from), tool.Typfmtv(&from0))
	dbglog.Log("        to.type: %v | input: %v", tool.Typfmtv(&to), tool.Typfmtv(&to0))

	err = c.copyTo(root, from, to)
	return
}

func (c *cpController) copyTo(params *Params, from, to reflect.Value) (err error) {
	err = c.copyToInternal(params, from, to,
		func(c *cpController, params *Params, from, to reflect.Value) (err error) {
			kind, pkgPath := from.Kind(), from.Type().PkgPath()
			if c.sourceExtractor != nil && to.IsValid() && !tool.IsNil(to) {
				// use tool.IsNil because we are checking for:
				// 1. to,IsNil() if 'to' is an addressable value (such as slice, map, or ptr)
				// 2. false if 'to' is not an addressable value (such as struct, int, ...)
				kind, pkgPath, c.targetOriented = to.Kind(), to.Type().PkgPath(), true
			} else {
				c.targetOriented = false
			}
			if kind != reflect.Struct || !packageisreserved(pkgPath) {
				if fn, ok := copyToRoutines[kind]; ok && fn != nil {
					err = fn(c, params, from, to)
					return
				}
			}

			// source is primitive type, or in a reserved package such as time, os, ...
			dbglog.Log("   - from.type: %v - fallback to copyDefaultHandler | to.type: %v", kind, tool.Typfmtv(&to))
			err = copyDefaultHandler(c, params, from, to)
			return
		})
	return
}

func (c *cpController) copyToInternal( //nolint:gocognit //yes, it is an integrated logic
	params *Params, from, to reflect.Value,
	cb copyfn,
) (err error) {
	// Return is from value is invalid
	if !from.IsValid() {
		if params.isGroupedFlagOKDeeply(cms.OmitIfEmpty, cms.OmitIfNil, cms.OmitIfZero) {
			return // fast fail here
		}
		// todo set target to zero
		return
	}
	if !to.IsValid() {
		dbglog.Log(`target is invalid, cannot be set.`)
		return
	}

	if c.testCloneable(params, from, to) {
		dbglog.Log(`from -> to is NOT cloneable`)
		return
	}

	//nolint:lll,nestif //keep it
	if from.CanAddr() && to.CanAddr() && tool.KindIs(from.Kind(), reflect.Array, reflect.Map, reflect.Slice, reflect.Struct) {
		addr1 := unsafe.Pointer(from.UnsafeAddr())
		addr2 := unsafe.Pointer(to.UnsafeAddr())
		if uintptr(addr1) > uintptr(addr2) {
			// Canonicalize order to reduce number of entries in visited.
			// Assumes non-moving garbage collector.
			addr1, addr2 = addr2, addr1
		}

		if params != nil {
			params.visiting = visit{addr1, addr2, from.Type()}
			if params.visited == nil {
				params.visited = make(map[visit]visiteddestination)
			}
			if dest, ok := params.visited[params.visiting]; ok {
				to.Set(dest.dst)
				dbglog.Log(`The visited target found, set it from cache`)
				return
			}
			params.visited[params.visiting] = visiteddestination{}
		}
	}

	// fromType := c.indirectType(from.Type())
	// toType := c.indirectType(to.Type())

	defer func() {
		if e := recover(); e != nil {
			err = errors.New("[recovered] copyTo unsatisfied ([%v] -> [%v])",
				tool.RindirectType(from.Type()), tool.RindirectType(to.Type())).
				WithMaxObjectStringLength(maxObjectStringLen).
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

	params.resultForNewSlice = nil
	err = cb(c, params, from, to)
	return
}

func (c *cpController) testCloneable(params *Params, from, to reflect.Value) (processed bool) {
	if from.CanInterface() { //nolint:nestif //keep it
		var fromObj interface{}
		if params != nil && params.srcOwner != nil {
			f, t := *params.srcOwner, *params.dstOwner
		retry:
			fromObj = f.Interface()
			if c.testCloneable1(params, fromObj, t) {
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

func (c *cpController) testCloneable1(params *Params, fromObj interface{}, to reflect.Value) (processed bool) {
	if dc, ok := fromObj.(Cloneable); ok { //nolint:gocritic // no need to rewrite to 'switch'
		to.Set(reflect.ValueOf(dc.Clone()))
		processed = true
	} else if dc1, ok1 := fromObj.(DeepCopyable); ok1 {
		to.Set(reflect.ValueOf(dc1.DeepCopy()))
		processed = true
	}
	return
}

func (c *cpController) withConverters(cvt ...ValueConverter) *cpController { //nolint:unused //usable
	for _, cc := range cvt {
		if cc != nil {
			c.valueConverters = append(c.valueConverters, cc)
		}
	}
	return c
}

func (c *cpController) withCopiers(cvt ...ValueCopier) *cpController { //nolint:unused //future
	for _, cc := range cvt {
		if cc != nil {
			c.valueCopiers = append(c.valueCopiers, cc)
		}
	}
	return c
}

func (c *cpController) withFlags(flags1 ...cms.CopyMergeStrategy) *cpController { //nolint:unused //future
	if c.flags == nil {
		c.flags = flags.New(flags1...)
	} else {
		c.flags.WithFlags(flags1...)
	}
	return c
}

func (c *cpController) Flags() flags.Flags     { return c.flags }
func (c *cpController) SetFlags(f flags.Flags) { c.flags = f }

// SaveFlagsAndRestore is a defer-function so the best usage is:
//
//	defer c.SaveFlagsAndRestore()()
func (c *cpController) SaveFlagsAndRestore() func() {
	var saved = c.flags.Clone()
	return func() {
		c.flags = saved
	}
}

const maxObjectStringLen = 320
