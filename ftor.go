package evendeep

// ftor.go - functors
//

import (
	"github.com/hedzr/log"
	"github.com/hedzr/log/color"

	"github.com/hedzr/evendeep/dbglog"
	"github.com/hedzr/evendeep/flags/cms"
	"github.com/hedzr/evendeep/internal/cl"
	"github.com/hedzr/evendeep/internal/tool"
	"github.com/hedzr/evendeep/typ"

	"gopkg.in/hedzr/errors.v3"

	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

func copyPointer(c *cpController, params *Params, from, to reflect.Value) (err error) {
	// from is a pointer
	fromType := from.Type()

	if fromType == to.Type() {
		if params.isFlagExists(cms.Flat) {
			to.Set(from)
			return
		}
	}

	src := tool.Rindirect(from)
	tgt := tool.Rindirect(to)

	paramsChild := newParams(withOwners(c, params, &from, &to, nil, nil))
	defer paramsChild.revoke()

	//nolint:lll //keep it
	if tgt.CanSet() {
		if src.IsValid() {
			err = c.copyTo(paramsChild, src, to)
		} else if paramsChild.isGroupedFlagOKDeeply(cms.ClearIfInvalid) {
			// pointer - src is nil - set tgt to nil too
			newtyp := to.Type()
			zv := reflect.Zero(newtyp)
			dbglog.Log("    pointer - zv: %v (%v), to: %v (%v)", tool.Valfmt(&zv), tool.Typfmt(newtyp), tool.Valfmt(&to), tool.Typfmtv(&to))
			to.Set(zv)
			// err = newobj(c, params, src, to, tgt)
		}
	} else {
		dbglog.Log("    pointer - tgt is invalid/cannot-be-set/ignored: src: (%v) -> tgt: (%v)", tool.Typfmtv(&src), tool.Typfmtv(&to))
		err = newObj(c, paramsChild, fromType, src, to, tgt)
	}
	return
}

func newObj(c *cpController, params *Params, fromType reflect.Type, src, to, tgt reflect.Value) (err error) {
	newtyp := to.Type()
	if to.Type() == fromType {
		newtyp = newtyp.Elem() // is pointer and its same
	}
	// create new object and pointer
	toobjcopyptrv := reflect.New(newtyp)
	dbglog.Log("    toobjcopyptrv: %v", tool.Typfmtv(&toobjcopyptrv))
	if err = c.copyTo(params, src, toobjcopyptrv.Elem()); err == nil {
		val := toobjcopyptrv
		if to.Type() == fromType {
			val = val.Elem()
		}
		err = setTargetValue1(params.owner, to, tgt, val)
		// to.Set(toobjcopyptrv)
	}
	return
}

func copyInterface(c *cpController, params *Params, from, to reflect.Value) (err error) {
	if tool.IsNil(from) {
		if params.isGroupedFlagOKDeeply(cms.OmitIfNil, cms.OmitIfEmpty) {
			return
		}
		if to.CanSet() {
			to.Set(reflect.Zero(to.Type()))
			return
		} else if to.Kind() == reflect.Ptr {
			if to.Elem().CanSet() {
				to.Elem().Set(reflect.Zero(to.Elem().Type()))
				return
			}
		}
		goto badReturn
	}

	if tool.IsZero(from) {
		if params.isGroupedFlagOKDeeply(cms.OmitIfZero, cms.OmitIfEmpty) {
			return
		}
		if to.CanSet() {
			to.Set(reflect.Zero(to.Type()))
			return
		} else if to.Kind() == reflect.Ptr {
			if to.Elem().CanSet() {
				to.Elem().Set(reflect.Zero(to.Elem().Type()))
				return
			}
		}
		goto badReturn
	}
	err = copyInterfaceImpl(c, params, from, to)
	return

badReturn:
	err = ErrCannotSet.FormatWith(tool.Valfmt(&from), tool.Typfmtv(&from), tool.Valfmt(&to), tool.Typfmtv(&to))
	return
}

func copyInterfaceImpl(c *cpController, params *Params, from, to reflect.Value) (err error) {
	paramsChild := newParams(withOwners(c, params, &from, &to, nil, nil))
	defer paramsChild.revoke()

	// unbox the interface{} to original data type
	toind, toptr := tool.Rdecode(to) // c.skip(to, reflect.Interface, reflect.Pointer)

	dbglog.Log("from.type: %v, decode to: %v", from.Type().Kind(), paramsChild.srcDecoded.Kind())
	dbglog.Log("  to.type: %v, decode to: %v (ptr: %v) | CanSet: %v/%v, CanAddr: %v",
		to.Type().Kind(), toind.Kind(), toptr.Kind(), toind.CanSet(), toptr.CanSet(), toind.CanAddr())

	if !tool.KindIs(toind.Kind(), reflect.Map, reflect.Slice, reflect.Chan) {
		if !toind.CanSet() && !toptr.CanSet() {
			// log.Panicf("[copyInterface] toind.CanSet() is false!")

			// valid ptr pointed to an invalid source object,
			// valid ptr pointed to an invalid target object:
			//nolint:lll //keep it
			err = errors.New("[copyInterface] target pointer cannot be set, toind.Kind: %v, toptr.Kind: %v", toind.Kind(), toptr.Kind())
			return
		}
	}

	// var merging = c.flags.isAnyFlagsOK(SliceMerge, MapMerge) || params.isAnyFlagsOK(SliceMerge, MapMerge)
	if paramsChild.inMergeMode() || !c.makeNewClone {
		err = c.copyTo(paramsChild, *paramsChild.srcDecoded, toptr)
		return
	}
	if to.CanSet() {
		copyValue := reflect.New(paramsChild.srcDecoded.Type()).Elem()
		if err = c.copyTo(paramsChild, *paramsChild.srcDecoded, copyValue); err == nil {
			to.Set(copyValue)
		}
	}
	return
}

func copyStruct(c *cpController, params *Params, from, to reflect.Value) (err error) {
	// default is cms.ByOrdinal:
	//   loops all source fields and copy its value to the corresponding target field.
	cb := forEachSourceField
	if c.targetOriented || params.isGroupedFlagOKDeeply(cms.ByName) {
		// cmd.ByName strategy:
		//   loops all target fields and try copying value from source field by its name.
		cb = forEachTargetField
	}
	err = copyStructInternal(c, params, from, to, cb)
	return
}

func copyStructInternal( //nolint:gocognit //keep it
	c *cpController, params *Params,
	from, to reflect.Value,
	fn func(paramsChild *Params, ec errors.Error, i, amount *int, padding string) (err error),
) (err error) {
	var (
		i, amount   int
		padding     string
		ec          = errors.New("copyStructInternal errors")
		paramsChild = newParams(withOwners(c, params, &from, &to, nil, nil))
	)

	defer func() {
		if e := recover(); e != nil {
			sst := paramsChild.targetIterator.(sourceStructFieldsTable) //nolint:errcheck //no need

			ff := sst.TableRecord(i).FieldValue()
			var tf = paramsChild.dstOwner
			var tft = &paramsChild.dstType
			if paramsChild.accessor != nil {
				tf = paramsChild.accessor.FieldValue()
				tft = paramsChild.accessor.FieldType()
			}

			// .WithData(e) will collect e if it's an error object else store it simply
			ec.Attach(errors.New("[recovered] copyStruct unsatisfied ([%v] -> [%v]), causes: %v",
				tool.Typfmtv(ff), tool.Typfmtptr(tft), e).
				WithMaxObjectStringLength(maxObjectStringLen).
				WithData(e).
				WithTaggedData(errors.TaggedData{ // record the sites
					"source-field": ff,
					"target-field": tf,
					"source":       tool.Valfmt(ff),
					"target":       tool.Valfmt(tf),
				}))
			// n := log.CalcStackFrames(1)   // skip defer-recover frame at first
			// log.Skip(n).Errorf("%v", err) // skip golib frames and defer-recover frame, back to the point throwing panic
			// if c.rethrow {
			//	log.Panicf("%+v", ec)
			// } else {
			// dbglog.Err("copyStructInternal will return error: %+v", ec)
			// }
		}

		ec.Defer(&err)
		paramsChild.revoke()
		if err != nil {
			dbglog.Err("copyStructInternal will return error: %+v", err)
		}
	}()

	if dbglog.LogValid {
		// dbgFrontOfStruct(params, paramsChild, padding, func(msg string, args ...interface{}) { dbglog.Log(msg, args...) })
		dbgFrontOfStruct(paramsChild, padding, dbglog.Log)
	}

	var processed bool
	if processed, err = tryConverters(c, paramsChild, &from, paramsChild.dstDecoded, &paramsChild.dstType, true); processed { //nolint:lll //keep it
		return
	}

	switch k := paramsChild.dstDecoded.Kind(); k { //nolint:exhaustive //no need
	case reflect.Slice:
		dbglog.Log("     * struct -> slice case, ...")
		if paramsChild.dstDecoded.Len() > 0 { //nolint:gocritic // no need to switch to 'switch' clause
			err = c.copyTo(paramsChild, *paramsChild.srcOwner, paramsChild.dstDecoded.Index(0))
		} else if paramsChild.isGroupedFlagOKDeeply(cms.SliceCopyAppend, cms.SliceMerge) {
			err = cpStructToNewSliceElem0(paramsChild)
		} else {
			err = ErrCannotCopy.FormatWith(tool.Valfmt(&from), tool.Typfmtv(&from), tool.Valfmt(&to), tool.Typfmtv(&to))
		}
		ec.Attach(err)
		return

	case reflect.Array:
		dbglog.Log("     * struct -> array case, ...")
		if paramsChild.dstDecoded.Len() > 0 {
			err = c.copyTo(paramsChild, *paramsChild.srcOwner, paramsChild.dstDecoded.Index(0))
		} else {
			err = ErrCannotCopy.FormatWith(tool.Valfmt(&from), tool.Typfmtv(&from), tool.Valfmt(&to), tool.Typfmtv(&to))
		}
		ec.Attach(err)
		return

	case reflect.String:
		dbglog.Log("     * struct -> string case, ...")
		var str string
		if str, err = doMarshalling(*paramsChild.srcOwner); err == nil {
			target := reflect.ValueOf(str)
			if paramsChild.dstDecoded.CanSet() {
				paramsChild.dstDecoded.Set(target)
			} else {
				err = ErrCannotSet.FormatWith(
					tool.Valfmt(paramsChild.srcDecoded),
					tool.Typfmtv(paramsChild.srcDecoded),
					tool.Valfmt(paramsChild.dstDecoded),
					tool.Typfmtv(paramsChild.dstDecoded))
			}
		}
		return
	}

	err = fn(paramsChild, ec, &i, &amount, padding)
	ec.Attach(err)
	return
}

func cpStructToNewSliceElem0(params *Params) (err error) {
	eltyp := params.dstType.Elem()
	et, _ := tool.Rdecodetype(eltyp)
	elnew := reflect.New(et)
	slice, tgtptr, el := *params.dstDecoded, params.dstOwner, elnew.Elem()
	if eltyp != et {
		tgtptr, el = params.dstOwner, elnew
	}
	if err = params.controller.copyTo(params, *params.srcOwner, elnew); err == nil {
		result := reflect.Append(slice, el)
		if tk := tgtptr.Kind(); tk == reflect.Slice || tk == reflect.Interface {
			tgtptr.Set(result)
		} else {
			tgtptr.Elem().Set(result)
		}
	}
	return
}

//nolint:lll //keep it
func getSourceFieldName(knownDestName string, params *Params) (srcFieldName string, flagsInTag *fieldTags, ignored bool) {
	srcFieldName = knownDestName
	// var flagsInTag *fieldTags
	var ok bool
	if sf := params.accessor.StructField(); sf != nil { //nolint:nestif //keep it
		flagsInTag, ignored = params.parseFieldTags(sf.Tag)
		if ignored {
			return
		}
		srcFieldName, ok = flagsInTag.CalcSourceName(sf.Name)
		// dbglog.Log("     srcName: %v, ok: %v [pre 1, fld: %v, tag: %v]", srcFieldName, ok, sf.Name, sf.Tag)
		if !ok {
			if tr := params.accessor.SourceField(); tr != nil {
				if sf = tr.structField; sf != nil {
					flagsInTag, ignored = params.parseFieldTags(sf.Tag)
					if ignored {
						return
					}
					srcFieldName, ok = flagsInTag.CalcSourceName(sf.Name)
					// dbglog.Log("     srcName: %v, ok: %v [pre 2, fld: %v, tag: %v]", srcFieldName, ok, sf.Name, sf.Tag)
				}
			}
		}
	}
	dbglog.Log("     srcName: %v, ok: %v | dstName: %v", srcFieldName, ok, knownDestName)
	return
}

// func getTargetFieldName(knownSrcName string, params *Params) (dstFieldName string, ignored bool) {
// 	dstFieldName = knownSrcName
// 	var flagsInTag *fieldTags
// 	var ok bool
// 	if sf := params.accessor.StructField(); sf != nil {
// 		flagsInTag, ignored = params.parseFieldTags(sf.Tag)
// 		if ignored {
// 			return
// 		}
// 		ctx := &NameConverterContext{Params: params}
// 		dstFieldName, ok = flagsInTag.CalcTargetName(sf.Name, ctx)
// 		if !ok {
// 			if tr := params.accessor.SourceField(); tr != nil {
// 				if sf = tr.structField; sf != nil {
// 					flagsInTag, ignored = params.parseFieldTags(sf.Tag)
// 					if ignored {
// 						return
// 					}
// 					dstFieldName, ok = flagsInTag.CalcTargetName(sf.Name, ctx)
// 					dbglog.Log("     dstName: %v, ok: %v [pre 2, fld: %v, tag: %v]", dstFieldName, ok, sf.Name, sf.Tag)
// 				}
// 			}
// 		}
// 	}
// 	return
// }

// forEachTargetField works for cms.ByName mode enabled.
func forEachTargetField(params *Params, ec errors.Error, i, amount *int, padding string) (err error) {
	c := params.controller

	var sst = params.targetIterator.(sourceStructFieldsTable)
	var val reflect.Value
	var fcz = params.isGroupedFlagOKDeeply(cms.ClearIfMissed)
	var aun = c.autoNewStruct // autoNew mode:
	var cfrtt = c.copyFunctionResultToTarget
	// We will do new(field) for each target field if it's invalid.
	//
	// It's not cms.ClearIfInvalid - which will detect if the source
	// field is invalid or not, but target one.
	//
	//nolint:lll //keep it
	dbglog.Log("     c.autoNewStruct = %v, c.copyFunctionResultToTarget = %v, cms.ClearIfMissed is set: %v", aun, cfrtt, fcz)

	for *i, *amount = 0, len(sst.TableRecords()); params.nextTargetFieldLite(); *i++ {
		name := params.accessor.StructFieldName() // get target field name
		if params.shouldBeIgnored(name) {
			continue
		}

		if extractor := c.sourceExtractor; extractor != nil { //nolint:nestif //keep it
			v := extractor(name)
			val = reflect.ValueOf(v)
			params.accessor.Set(val)
			continue
		}

		srcFieldName, _, ignored := getSourceFieldName(name, params)
		if ignored {
			dbglog.Log(`     > source field ignored (flag found in struct tag): %v.`, srcFieldName)
			continue
		}

		ind := sst.RecordByName(srcFieldName)
		switch {
		case ind != nil:
			val = *ind
		case cfrtt:
			if _, ind = sst.MethodCallByName(srcFieldName); ind != nil {
				val = *ind
			} else if _, ind = sst.MethodCallByName(name); ind != nil {
				val = *ind
			} else {
				continue // skip the field
			}
		case fcz || (aun && !params.accessor.ValueValid()):
			tt := params.accessor.FieldType()
			val = reflect.Zero(*tt)
			dbglog.Log("     target is invalid: %v, autoNewStruct: %v", params.accessor.ValueValid(), aun)
		default:
			continue
		}
		params.accessor.Set(val)
	}
	return
}

//nolint:gocognit //unify scene
func forEachSourceField(params *Params, ec errors.Error, i, amount *int, padding string) (err error) {
	sst := params.targetIterator.(sourceStructFieldsTable) //nolint:errcheck //no need
	c := params.controller

	for *i, *amount = 0, len(sst.TableRecords()); *i < *amount; *i++ {
		if params.sourceFieldShouldBeIgnored() {
			dbglog.Log("%d. %s : IGNORED", *i, sst.CurrRecord().FieldName())
			if c.advanceTargetFieldPointerEvenIfSourceIgnored {
				_ = params.nextTargetFieldLite()
			} else {
				sst.Step(1) // step the source field index subscription
			}
			continue
		}

		var goon bool
		if goon, err = forEachSourceFieldCheckTargetSetter(params, sst); goon {
			continue
		}

		var sourceField *tableRecT
		if sourceField, goon = params.nextTargetField(); !goon {
			continue
		}

		fn, srcval, dstval := sourceField.FieldName(), sourceField.FieldValue(), params.accessor.FieldValue()

		dstfieldname := params.accessor.StructFieldName()
		// log.VDebugf will be tuned and stripped off in normal build.
		dbglog.Colored(color.LightMagenta, "%d. fld %q (%v) -> %s (%v) | (%v) -> (%v)", *i,
			fn, tool.Typfmtv(srcval), dstfieldname, tool.Typfmt(*params.accessor.FieldType()),
			tool.Valfmt(srcval), tool.Valfmt(dstval))

		// The following if clause will be stripped off completely
		// in normal build.
		// It's only available when using `-tags=verbose`.
		// The first condition 'log.VerboseEnabled' do circuit to
		// be optimized by compiler.
		if log.VerboseEnabled && !ec.IsEmpty() {
			log.Warnf("    ERR-CONTAINER NOT EMPTY: %+v", ec.Error())
		}

		if srcval != nil && dstval != nil {
			typ := params.accessor.FieldType() // target type
			if typ != nil && !tool.KindIs((*typ).Kind(), reflect.Chan, reflect.Func, reflect.Interface, reflect.UnsafePointer, reflect.Ptr, reflect.Slice) {
				if tool.IsNil(*dstval) || !(*dstval).IsValid() {
					if !tool.IsNil(*srcval) {
						dbglog.Log("      create new: dstval = nil/invalid, type: %v (%v -> nil/inalid)", tool.Typfmt(*typ), tool.Valfmt(srcval))
						_, elem := newFromTypeEspSlice(*typ)
						dstval.Set(elem)
						dbglog.Log("      create new: dstval created: %v", tool.Typfmtv(dstval))
					}
				}
			}

			if srcval.IsValid() {
				if err = invokeStructFieldTransformer(c, params, srcval, dstval, typ, padding); err != nil {
					ec.Attach(err)
					dbglog.Err("    %d. fld %q error: %v", *i, fn, err)
				} else {
					dbglog.Log("    %d. fld %q copied. from-to: %v -> %v", *i, fn, tool.Valfmt(srcval), tool.Valfmt(dstval))
				}
			}
			continue
		}

		if goon = forEachSourceFieldCheckMergeMode(params, srcval, ec, padding); goon {
			continue
		}

		dbglog.Wrn("   %d. fld %q: ignore nil/zero/invalid source or nil target", *i, fn)
	}
	return
}

func forEachSourceFieldCheckTargetSetter(params *Params, sst sourceStructFieldsTable) (goon bool, err error) {
	c := params.controller
	if c.targetSetter == nil {
		return
	}

	currec := sst.CurrRecord()
	srcval := currec.FieldValue()
	if err = c.targetSetter(srcval, currec.names...); err == nil {
		if c.advanceTargetFieldPointerEvenIfSourceIgnored {
			_ = params.nextTargetFieldLite()
		} else {
			sst.Step(1) // step the source field index
		}
		goon = true
		return // targetSetter make things done, so continue to next field
	}

	if err != ErrShouldFallback { //nolint:errorlint //want it exactly
		return // has error, break the whole copier loop
	}
	err = nil // fallback
	return
}

func forEachSourceFieldCheckMergeMode(params *Params, srcval *reflect.Value,
	ec errors.Error, padding string) (goon bool) {
	if params.inMergeMode() {
		var err error
		c := params.controller

		typ := params.accessor.FieldType()
		dbglog.Log("    new object for %v", tool.Typfmt(*typ))

		// create new object and pointer
		toobjcopyptrv := reflect.New(*typ).Elem()
		dbglog.Log("    toobjcopyptrv: %v", tool.Typfmtv(&toobjcopyptrv))

		//nolint:gocritic // no need to switch to 'switch' clause
		if err = invokeStructFieldTransformer(c, params, srcval, &toobjcopyptrv, typ, padding); err != nil {
			ec.Attach(err)
			dbglog.Err("error: %v", err)
		} else if toobjcopyptrv.Kind() == reflect.Slice {
			params.accessor.Set(toobjcopyptrv)
		} else if toobjcopyptrv.Kind() == reflect.Ptr {
			params.accessor.Set(toobjcopyptrv.Elem())
		} else {
			params.accessor.Set(toobjcopyptrv)
		}
		goon = true
	}
	return
}

func dbgFrontOfStruct(params *Params, padding string, logger func(msg string, args ...interface{})) {
	if log.VerboseEnabled {
		if params == nil {
			return
		}
		if logger == nil {
			logger = dbglog.Log
		}
		d := params.depth()
		if d > 1 {
			d -= 2
		}
		padding1 := strings.Repeat("  ", d*2) //nolint:gomnd //no need
		// fromT, toT := params.srcDecoded.Type(), params.dstDecoded.Type()
		// Log(" %s  %d, %d, %d", padding, params.index, params.srcOffset, params.dstOffset)
		// fq := dbgMakeInfoString(fromT, params.owner, true, logger)
		// dq := dbgMakeInfoString(toT, params.owner, false, logger)
		logger(" %s- src (%v) -> dst (%v)", padding1, tool.Typfmtv(params.srcDecoded), tool.Typfmtv(params.dstDecoded))
		// logger(" %s  %s -> %s", padding, fq, dq)
	}
}

func invokeStructFieldTransformer(
	c *cpController, params *Params, ff, df *reflect.Value,
	dftyp *reflect.Type, //nolint:gocritic // ptrToRefParam: consider 'dftyp' to be of non-pointer type
	padding string,
) (err error) {
	fv, dv := ff != nil && ff.IsValid(), df != nil && df.IsValid()
	fft, dft := dtypzz(ff, dftyp), dtypzz(df, dftyp)
	fftk, dftk := fft.Kind(), dft.Kind()

	var processed bool
	if processed = checkClearIfEqualOpt(params, ff, df, dft); processed {
		return
	}
	if processed = checkOmitEmptyOpt(params, ff, df, dft); processed {
		return
	}
	if processed, err = tryConverters(c, params, ff, df, dftyp, false); processed {
		return
	}

	if fftk == reflect.Struct && ff.NumField() == 0 {
		// never get into here because tableRecordsT.getAllFields skip empty struct
		log.Warnf("should never get into here, might be algor wrong ?")
	}
	if dftk == reflect.Struct && df.NumField() == 0 {
		// structIterable.Next() might return an empty struct accessor
		// rather than field.
		dbglog.Err("shouldn't get into here because we have a failover branch at the callee")
	}

	if fv && dv {
		dbglog.Log(`      c.copyTo: ff -> df`)
		err = c.copyTo(params, *ff, *df) // or, use internal standard implementation version
		dbglog.Log(`      c.copyTo.end: ff -> df. err = %v, df = %v`, err, tool.Valfmt(df))
		if log.VerboseEnabled {
			if df.CanInterface() {
				switch k := df.Kind(); k {
				case reflect.Map, reflect.Slice:
					var v typ.Any = df.Interface()
					dbglog.Log(`      c.copyTo.end: df = %v`, v)
				}
			}
		}
		return
	}

	return forInvalidValues(c, params, ff, fft, dft, fftk, dftk, fv)
}

//nolint:lll //keep it
func forInvalidValues(c *cpController, params *Params, ff *reflect.Value, fft, dft reflect.Type, fftk, dftk reflect.Kind, fv bool) (err error) {
	if !fv {
		dbglog.Log("   ff is invalid: %v", tool.Typfmtv(ff))
		nv := reflect.New(fft).Elem()
		ff = &nv
	}
	if dftk == reflect.Interface {
		dft, dftk = fft, fftk
	}
	dbglog.Log("     dft: %v", tool.Typfmt(dft))
	if dftk == reflect.Ptr {
		nv := reflect.New(dft.Elem())
		tt := nv.Elem()
		dbglog.Log("   nv.tt: %v", tool.Typfmtv(&tt))
		ff1 := tool.Rindirect(*ff)
		err = c.copyTo(params, ff1, tt) // use user-defined copy-n-merger to merge or copy source to destination
		if err == nil && !params.accessor.IsStruct() {
			params.accessor.Set(tt)
		}
	} else {
		nv := reflect.New(dft)
		ff1 := tool.Rindirect(*ff)
		err = c.copyTo(params, ff1, nv.Elem()) // use user-defined copy-n-merger to merge or copy source to destination
		if err == nil && !params.accessor.IsStruct() {
			params.accessor.Set(nv.Elem())
		}
	}
	return
}

//nolint:lll //keep it
func dtypzz(df *reflect.Value, deftyp *reflect.Type) reflect.Type { //nolint:gocritic // ptrToRefParam: consider 'dftyp' to be of non-pointer type
	if df != nil && df.IsValid() {
		return df.Type()
	}
	return *deftyp
}

func checkOmitEmptyOpt(params *Params, ff, df *reflect.Value, dft reflect.Type) (processed bool) {
	if tool.IsNilv(ff) && params.isGroupedFlagOKDeeply(cms.OmitIfNil, cms.OmitIfEmpty) {
		processed = true
	}
	if tool.IsZerov(ff) && params.isGroupedFlagOKDeeply(cms.OmitIfZero, cms.OmitIfEmpty) {
		processed = true
	}
	return
}

func checkClearIfEqualOpt(params *Params, ff, df *reflect.Value, dft reflect.Type) (processed bool) {
	if params.isFlagExists(cms.ClearIfEq) {
		if tool.EqualClassical(*ff, *df) {
			df.Set(reflect.Zero(dft))
		} else if params.isFlagExists(cms.ClearIfInvalid) && !df.IsValid() {
			df.Set(reflect.Zero(dft))
		}
		processed = true
		if params.isFlagExists(cms.KeepIfNotEq) {
			return
		}
	}
	return
}

func tryConverters(c *cpController, params *Params,
	ff, df *reflect.Value,
	dftyp *reflect.Type, //nolint:gocritic // ptrToRefParam: consider 'dftyp' to be of non-pointer type
	userDefinedOnly bool,
) (processed bool, err error) {
	fft, dft := dtypzz(ff, dftyp), dtypzz(df, dftyp)
	if c.tryApplyConverterAtFirst {
		if processed, err = findAndApplyConverters(c, params, ff, df, fft, dft, userDefinedOnly); processed {
			return
		}
		processed, err = findAndApplyCopiers(c, params, ff, df, fft, dft, userDefinedOnly)
	} else {
		if processed, err = findAndApplyCopiers(c, params, ff, df, fft, dft, userDefinedOnly); processed {
			return
		}
		processed, err = findAndApplyConverters(c, params, ff, df, fft, dft, userDefinedOnly)
	}
	return
}

//nolint:lll //keep it
func findAndApplyCopiers(c *cpController, params *Params, ff, df *reflect.Value, fft, dft reflect.Type, userDefinedOnly bool) (processed bool, err error) {
	if cvt, ctx := c.valueCopiers.findCopiers(params, fft, dft, userDefinedOnly); ctx != nil { //nolint:nestif //keep it
		dbglog.Colored(color.DarkColor, "-> using Copier %v", reflect.ValueOf(cvt).Type())

		if df.IsValid() {
			if err = cvt.CopyTo(ctx, *safeFF(ff, fft), *df); err == nil { // use user-defined copy-n-merger to merge or copy source to destination
				processed = true
			}
			return
		}

		if dft.Kind() == reflect.Interface {
			dft = fft
		}
		dbglog.Log("  dft: %v", tool.Typfmt(dft))
		nv := reflect.New(dft)
		err = cvt.CopyTo(ctx, *safeFF(ff, fft), nv) // use user-defined copy-n-merger to merge or copy source to destination
		if err == nil && !params.accessor.IsStruct() {
			params.accessor.Set(nv.Elem())
			processed = true
		}
	}
	return
}

//nolint:lll //keep it
func findAndApplyConverters(c *cpController, params *Params, ff, df *reflect.Value, fft, dft reflect.Type, userDefinedOnly bool) (processed bool, err error) {
	if cvt, ctx := c.valueConverters.findConverters(params, fft, dft, userDefinedOnly); ctx != nil {
		dbglog.Colored(color.DarkColor, "-> using Converter %v", reflect.ValueOf(cvt).Type())

		var result reflect.Value
		result, err = cvt.Transform(ctx, *safeFF(ff, fft), dft) // use user-defined value converter to transform from source to destination
		if err == nil && !df.IsValid() && !params.accessor.IsStruct() {
			params.accessor.Set(result)
			processed = true
			return
		}
		df.Set(result)
		processed = true
	}
	return
}

func safeFF(ff *reflect.Value, fft reflect.Type) *reflect.Value {
	if ff == nil {
		tmpsrc := reflect.New(fft).Elem()
		ff = &tmpsrc
	}
	return ff
}

// copySlice transforms from slice to target with slice or other types.
func copySlice(c *cpController, params *Params, from, to reflect.Value) (err error) {
	if from.IsNil() && params.isGroupedFlagOKDeeply(cms.OmitIfNil, cms.OmitIfEmpty) { // an empty slice found
		return
	}

	var tgt, tgtptr reflect.Value
	tgt, tgtptr = tool.Rdecode(to)
	if to != tgtptr { //nolint:govet //how should i do //TODO needs checked-review
		err = c.copyTo(params, from, tgtptr) // unwrap the pointer
		return
	}

	if !tool.KindIs(tgt.Kind(), reflect.Map, reflect.Slice, reflect.Chan) && !tgt.CanSet() {
		log.Panicf("[copySlice] tgtptr cannot be set")
	}
	if params.controller != c {
		log.Panicf("[copySlice] c *cpController != params.controller, what's up??")
	}

	tk, typ := tgt.Kind(), tgt.Type()
	if tk != reflect.Slice {
		dbglog.Log("[copySlice] from slice -> %v", tool.Typfmt(typ))
		var processed bool
		if processed, err = tryConverters(c, params, &from, &tgt, &typ, false); !processed {
			// log.Panicf("[copySlice] unsupported transforming: from slice -> %v,", typfmtv(&tgt))
			//nolint:lll //keep it
			err = ErrCannotCopy.WithErrors(err).FormatWith(tool.Valfmt(&from), tool.Typfmtv(&from), tool.Valfmt(&tgt), tool.Typfmtv(&tgt))
		}
		return
	}

	if tool.IsNil(tgt) && params.isGroupedFlagOKDeeply(cms.OmitIfTargetZero, cms.OmitIfTargetEmpty) {
		return
	}

	params.resultForNewSlice, err = copySliceInternal(c, params, from, to, tgt, tgtptr)
	return
}

//nolint:lll,gocognit //keep it
func copySliceInternal(c *cpController, params *Params, from, to, tgt, tgtptr reflect.Value) (result *reflect.Value, err error) {
	if from.Len() == 0 {
		dbglog.Log("  slice copy ignored because src slice is empty.")
		return
	}

	ec := errors.New("slice copy/merge errors")
	defer ec.Defer(&err)

	for _, flag := range []cms.CopyMergeStrategy{cms.SliceMerge, cms.SliceCopyAppend, cms.SliceCopy} {
		if params.isGroupedFlagOKDeeply(flag) { //nolint:nestif,gocritic // nestingReduce: invert if cond, replace body with `continue`, move old body after the statement
			dbglog.Log("Using slice merge mode: %v", flag)
			dbglog.Log("  from.type: %v, value: %v", tool.Typfmtv(&from), tool.Valfmt(&from))
			dbglog.Log("    to.type: %v, value: %v | canAddr: %v, canSet: %v", tool.Typfmtv(&to), tool.Valfmt(&to), to.CanAddr(), to.CanSet())
			// Log(" src.type: %v, len: %v, cap: %v, srcptr.canAddr: %v", src.Type().Kind(), src.Len(), src.Cap(), srcptr.CanAddr())
			dbglog.Log("   tgt.type: %v, tgtptr: %v .canAddr: %v", tool.Typfmtv(&tgt), tool.Typfmtv(&tgtptr), tgtptr.CanAddr())

			if fn, ok := getSliceOperations()[flag]; ok {
				if result, err = fn(c, params, from, tgt); err == nil {
					dbglog.Log("     result: got %v (%v)", tool.Valfmt(result), tool.Typfmtv(result))
					dbglog.Log("        tgt: contains %v (%v) | tgtptr: %v, .canset: %v", tool.Valfmt(&tgt), tool.Typfmtv(&tgt), tool.Typfmtv(&tgtptr), tgtptr.CanSet())

					if tk := tgtptr.Kind(); tk == reflect.Ptr { //nolint:gocritic //keep it
						tgtptr.Elem().Set(*result)
					} else if tk == reflect.Slice || tk == reflect.Interface {
						// dbglog.Log("        TGT: %v", tool.Valfmt(&tgt))
						if tgtptr.CanSet() {
							tgtptr.Set(*result)
						}
					} else {
						dbglog.Log("      error: cannot make copy for a slice, the target ptr is cannot be set: tgtptr.typ = %v", tool.Typfmtv(&tgtptr))
						ec.Attach(errors.New("cannot make copy for a slice, the target ptr is cannot be set: tgtptr.typ = %v", tool.Typfmtv(&tgtptr)))
					}
				} else {
					dbglog.Log("      error: ec.Attach(e), e: %v", err)
					ec.Attach(err)
				}
			} else {
				dbglog.Log("      error: cannot make copy for a slice, unknown copy-merge-strategy %v", flag)
				ec.Attach(errors.New("cannot make copy for a slice, unknown copy-merge-strategy %v", flag))
			}

			break
		}
	}

	return
}

// transferSlice never used
//
//nolint:unused //future
func transferSlice(src, tgt reflect.Value) reflect.Value {
	i, sl, tl := 0, src.Len(), tgt.Len()
	for ; i < sl && i < tl; i++ {
		if i < tl {
			if i >= sl {
				tgt.SetLen(i)
				break
			}
			tgt.Index(i).Set(src.Index(i))
		}
	}
	if i < sl && i >= tl {
		for ; i < sl; i++ {
			tgt = reflect.Append(tgt, src.Index(i))
		}
	}
	return tgt
}

type fnSliceOperator func(c *cpController, params *Params, src, tgt reflect.Value) (result *reflect.Value, err error)
type mSliceOperations map[cms.CopyMergeStrategy]fnSliceOperator

func getSliceOperations() (mapOfSliceOperations mSliceOperations) {
	mapOfSliceOperations = mSliceOperations{ //nolint:exhaustive //i have right
		cms.SliceCopy:       _sliceCopyOperation,
		cms.SliceCopyAppend: _sliceCopyAppendOperation,
		cms.SliceMerge:      _sliceMergeOperation,
	}
	return
}

// _sliceCopyOperation: for SliceCopy, target elements will be given up, and source copied to.
func _sliceCopyOperation(c *cpController, params *Params, src, tgt reflect.Value) (result *reflect.Value, err error) {
	slice := reflect.MakeSlice(tgt.Type(), 0, src.Len())
	dbglog.Log("tgt slice: %v, el: %v", tgt.Type(), tgt.Type().Elem())

	ecTotal := errors.New("slice merge errors (%v -> %v)", src.Type(), tgt.Type())
	defer ecTotal.Defer(&err)

	// if c.wipeSlice1st {
	// 	// TODO c.wipeSlice1st
	// }

	for _, ss := range []struct {
		length int
		source reflect.Value
	}{
		// {tl, tgt},
		{src.Len(), src},
	} {
		var one *reflect.Value
		one, err = _sliceCopyOne(c, params, ecTotal, slice, ss.length, ss.source, tgt)
		if one != nil && err == nil {
			slice = *one
		}
	}
	result = &slice
	return
}

// _sliceCopyAppendOperation: for SliceCopyAppend, target and source elements will be copied to new target.
// The duplicated elements were kept.
//
//nolint:lll //keep it
func _sliceCopyAppendOperation(c *cpController, params *Params, src, tgt reflect.Value) (result *reflect.Value, err error) {
	sl, tl := src.Len(), tgt.Len()
	ns := reflect.MakeSlice(tgt.Type(), 0, 0)
	dbglog.Log("tgt slice: %v, el: %v", tgt.Type(), tgt.Type().Elem())

	ecTotal := errors.New("slice merge errors (%v -> %v)", src.Type(), tgt.Type())
	defer ecTotal.Defer(&err)

	for _, ss := range []struct {
		length int
		source reflect.Value
	}{
		{tl, tgt},
		{sl, src},
	} {
		var one *reflect.Value
		one, err = _sliceCopyOne(c, params, ecTotal, ns, ss.length, ss.source, tgt)
		if one != nil && err == nil {
			ns = *one
		}
	}
	result = &ns
	return
}

//nolint:lll //keep it
func _sliceCopyOne(c *cpController, params *Params, ecTotal errors.Error, slice reflect.Value, sslength int, sssource, tgt reflect.Value) (result *reflect.Value, err error) {
	tgtelemtype := tgt.Type().Elem()
	for i := 0; i < sslength; i++ {
		var (
			el   = sssource.Index(i)
			enew = el
		)
		// elv := el.Interface()
		if el.Type() != tgtelemtype {
			if cc, ctx := c.valueConverters.findConverters(params, el.Type(), tgtelemtype, false); cc != nil {
				if enew, err = cc.Transform(ctx, el, tgtelemtype); err != nil {
					ec := errors.New("cannot convert %v to %v", el.Type(), tgtelemtype)
					ec.Attach(err)
					ecTotal.Attach(ec)
					continue // ignore invalid element
				}
			} else if tool.CanConvert(&el, tgtelemtype) {
				enew = el.Convert(tgtelemtype)
				// elv = enew.Interface()
			}
		}

		if el.Type() == tgtelemtype { //nolint:nestif //keep it
			slice = reflect.Append(slice, el)
		} else {
			if tool.CanConvert(&el, tgtelemtype) {
				slice = reflect.Append(slice, enew)
			} else {
				enew = reflect.New(tgtelemtype)
				e := c.copyTo(params, el, enew)
				if e != nil {
					ec := errors.New("cannot convert %v to %v", el.Type(), tgtelemtype)
					ec.Attach(e)
					ecTotal.Attach(ec)
				} else {
					slice = reflect.Append(slice, enew.Elem())
				}
			}
		}
		// ecTotal.Attach(ec)
	}
	result = &slice
	return
}

// _sliceMergeOperation: for SliceMerge. target and source elements will be
// copied to new target with uniqueness.
//
//nolint:gocognit //unify scene
func _sliceMergeOperation(c *cpController, params *Params, src, tgt reflect.Value) (result *reflect.Value, err error) {
	sl, tl := src.Len(), tgt.Len()
	ns := reflect.MakeSlice(tgt.Type(), 0, 0)
	tgtelemtype := tgt.Type().Elem()
	dbglog.Log("tgt slice: %v, el: %v", tgt.Type(), tgtelemtype)

	ecTotal := errors.New("slice merge errors (%v -> %v)", src.Type(), tgt.Type())
	defer ecTotal.Defer(&err)

	for _, ss := range []struct {
		length int
		source reflect.Value
	}{
		{tl, tgt},
		{sl, src},
	} {
		for i := 0; i < ss.length; i++ {
			// to.Set(reflect.Append(to, src.Index(i)))
			var (
				found bool
				cvtok bool
				el    = ss.source.Index(i)
				elt   = el.Type()
				elv   = el.Interface()
				enew  = el
				ec    = errors.New("cannot convert %v to %v", el.Type(), tgtelemtype)
			)
			if elt != tgtelemtype { //nolint:nestif //keep it
				if cc, ctx := c.valueConverters.findConverters(params, elt, tgtelemtype, false); cc != nil {
					if enew, err = cc.Transform(ctx, el, tgtelemtype); err != nil {
						var ve *strconv.NumError
						if !errors.As(err, &ve) {
							ec.Attach(err)
							ecTotal.Attach(ec)
						}
						continue // ignore invalid element
					} else {
						cvtok, elv = true, enew.Interface()
					}
				} else if tool.CanConvert(&el, tgtelemtype) {
					enew = el.Convert(tgtelemtype)
					cvtok, elv = true, enew.Interface()
				}
			}

			if found = tool.FindInSlice(ns, elv, i); !found { //nolint:nestif //keep it
				if cvtok || elt == tgtelemtype {
					ns = reflect.Append(ns, enew)
				} else {
					enew = reflect.New(tgtelemtype)
					e := c.copyTo(params, el, enew)
					if e != nil {
						ec.Attach(e)
						err = ec
					} else {
						ns = reflect.Append(ns, enew.Elem())
					}
				}
			}
			if !ec.IsEmpty() {
				ecTotal.Attach(ec)
			}
		}
	}
	result = &ns
	return
}

func copyArray(c *cpController, params *Params, from, to reflect.Value) (err error) {
	if tool.IsZero(from) && params.isGroupedFlagOKDeeply(cms.OmitIfZero, cms.OmitIfEmpty) {
		return
	}

	src := tool.Rindirect(from)
	tgt, tgtptr := tool.Rdecode(to)

	// if !to.CanAddr() && params != nil {
	//	if !params.isStruct() {
	//		//to = *params.dstOwner
	//		Log("    !! use dstOwner to get a ptr to array, new to.type: %v,
	// canAddr: %v, canSet: %v", to.Type().Kind(), to.CanAddr(), to.CanSet())
	//	}
	// }

	// if tgt.CanAddr() == false && tgtptr.CanAddr() {
	//	tgt = tgtptr
	// }

	// Log("    tgt.%v: %v", params.dstOwner.Type().Field(params.index).Name, params.dstOwner.Type().Field(params.index))
	dbglog.Log("    from.type: %v, len: %v, cap: %v", src.Type().Kind(), src.Len(), src.Cap())
	dbglog.Log("      to.type: %v, len: %v, cap: %v, tgtptr.canSet: %v, tgtptr.canaddr: %v", tgt.Type().Kind(), tgt.Len(), tgt.Cap(), tgtptr.CanSet(), tgtptr.CanAddr()) //nolint:lll //keep it

	tk, tgttyp := tgt.Kind(), tgt.Type()
	if tk != reflect.Array {
		var processed bool
		if processed, err = tryConverters(c, params, &from, &tgt, &tgttyp, false); processed {
			return
		}
		// log.Panicf("[copySlice] unsupported transforming: from slice -> %v,", typfmtv(&tgt))
		err = ErrCannotCopy.FormatWith(tool.Valfmt(&src), tool.Typfmtv(&src), tool.Valfmt(&tgt), tool.Typfmtv(&tgt))
		return
	}

	if tool.IsZero(tgt) && params.isGroupedFlagOKDeeply(cms.OmitIfTargetZero, cms.OmitIfTargetEmpty) {
		return
	}

	sl, tl := src.Len(), tgt.Len()
	eltyp := tgt.Type().Elem()
	// set := src.Index(0).Type()
	// if set != tgt.Index(0).Type() {
	//	return errors.New("cannot copy %v to %v", from.Interface(), to.Interface())
	// }

	cnt := tool.MinInt(sl, tl)
	for i := 0; i < cnt; i++ {
		se := src.Index(i)
		setyp := se.Type()
		dbglog.Log("src.el.typ: %v, tgt.el.typ: %v", tool.Typfmt(setyp), eltyp)
		if se.IsValid() {
			if setyp.AssignableTo(eltyp) {
				tgt.Index(i).Set(se)
			} else if setyp.ConvertibleTo(eltyp) {
				tgt.Index(i).Set(src.Convert(eltyp))
			}
		}
		// tgt.Index(i).Set(src.Index(i))
	}

	for i := cnt; i < tl; i++ {
		v := tgt.Index(i)
		if !v.IsValid() {
			tgt.Index(i).Set(reflect.Zero(eltyp))
			dbglog.Log("set [%v] to zero value", i)
		}
	}

	// to.Set(pt.Elem())

	dbglog.Log("    from: %v, to: %v", src.Interface(), tgt.Interface()) // pt.Interface())

	return
}

func copyMap(c *cpController, params *Params, from, to reflect.Value) (err error) {
	if from.IsNil() && params.isGroupedFlagOKDeeply(cms.OmitIfNil, cms.OmitIfEmpty) { // an empty slice found
		return
	}

	var tgt, tgtptr reflect.Value
	tgt, tgtptr = tool.Rdecode(to)
	if to != tgtptr { //nolint:govet //how should i do //TODO needs checked-review
		err = c.copyTo(params, from, tgtptr) // unwrap the pointer
		return
	}

	tk, typ := tgt.Kind(), tgt.Type()
	if tk != reflect.Map {
		dbglog.Log("from map -> %v", tool.Typfmt(typ))
		// copy map to String, Slice, Struct
		var processed bool
		if processed, err = tryConverters(c, params, &from, &tgt, &typ, false); !processed {
			err = ErrCannotCopy.WithErrors(err).
				FormatWith(tool.Valfmt(&from), tool.Typfmtv(&from), tool.Valfmt(&tgt), tool.Typfmtv(&tgt))
		}
		return
	}

	tgtNil := tool.IsNil(tgt)
	if tgtNil && params.isGroupedFlagOKDeeply(cms.OmitIfTargetZero, cms.OmitIfTargetEmpty) {
		return
	}

	ec := errors.New("map copy/merge errors")
	defer dbglog.DeferVisit(ec, &err)

	// By default, the nested copyTo() cannot print log msg via dbglog.Log
	// so we get clearer logging lines for debugging.
	// To enabled them, use build tags 'moremaplog'.
	defer dbglog.DisableLog()()

	for _, flag := range []cms.CopyMergeStrategy{cms.MapMerge, cms.MapCopy} {
		if params.isGroupedFlagOKDeeply(flag) {
			if fn, ok := getMapOperations()[flag]; ok {
				ec.Attach(fn(c, params, from, tgt, tgtptr))
			} else {
				ec.Attach(errors.New("unknown strategy for map: %v", flag))
			}
			break // once any of copy-merge strategy matched, stop iterating now
		}
	}
	return
}

type fnMapOperation func(c *cpController, params *Params, src, tgt, tgtptr reflect.Value) (err error)
type mapMapOperations map[cms.CopyMergeStrategy]fnMapOperation

func getMapOperations() (mMapOperations mapMapOperations) {
	mMapOperations = mapMapOperations{ //nolint:exhaustive //i have right
		cms.MapCopy: func(c *cpController, params *Params, src, tgt, tgtptr reflect.Value) (err error) {
			tgt.Set(reflect.MakeMap(src.Type()))

			ec := errors.New("map copy errors")
			defer ec.Defer(&err)

			for _, key := range src.MapKeys() {
				originalValue := src.MapIndex(key)
				_, copyValueElem := newFromType(tgt.Type().Elem())
				ec.Attach(c.copyTo(params, originalValue, copyValueElem))

				copyKey := reflect.New(tgt.Type().Key())
				ec.Attach(c.copyTo(params, key, copyKey.Elem()))

				if c.targetSetter != nil && copyKey.Elem().Kind() == reflect.String {
					srcval := copyValueElem
					err = c.targetSetter(&srcval, copyKey.Elem().String())
					if err != nil {
						if err != ErrShouldFallback { //nolint:errorlint //want it exactly
							return
						}
						err = nil
					} else {
						return
					}
				}
				trySetMapIndex(c, params, tgt, copyKey.Elem(), copyValueElem)
			}
			return
		},
		cms.MapMerge: func(c *cpController, params *Params, src, tgt, tgtptr reflect.Value) (err error) {
			ec := errors.New("map merge errors")
			defer ec.Defer(&err)

			for _, key := range src.MapKeys() {
				// dbglog.Log("------------ [MapMerge] mergeOneKeyInMap: key = %q (%v) ------------------",
				// 	tool.Valfmt(&key), tool.Typfmtv(&key))
				ec.Attach(mergeOneKeyInMap(c, params, src, tgt, tgtptr, key))
			}
			return
		},
	}
	return
}

func mapMergePreSetter(c *cpController, ck, cv reflect.Value) (processed bool, err error) {
	if c.targetSetter != nil && ck.Kind() == reflect.String {
		err = c.targetSetter(&cv, ck.String())
		if err != nil {
			if err == ErrShouldFallback { //nolint:errorlint //want it exactly
				processed, err = true, nil
			}
		} else {
			processed = true
		}
	}
	return
}

// mergeOneKeyInMap copy one (key, value) pair in src map to tgt map.
func mergeOneKeyInMap(c *cpController, params *Params, src, tgt, tgtptr, key reflect.Value) (err error) {
	var processed bool

	dbglog.Colored(color.LightMagenta, "      <MAP> copying key '%v': (%v) -> (?)", tool.Valfmt(&key), tool.Valfmtv(src.MapIndex(key)))

	var ck reflect.Value
	if ck, err = cloneMapKey(c, params, tgt, key); err != nil {
		return
	}

	originalValue := src.MapIndex(key)

	tgtval, newelemcreated, err2 := ensureMapPtrValue(c, params, tgt, ck, originalValue)
	if err = err2; err2 != nil {
		return
	}
	if newelemcreated {
		cv := tool.Rindirect(tgtval)
		if processed, err = mapMergePreSetter(c, ck, cv); processed {
			return
		}
		defer matchTypeAndSetMapIndex(c, params, tgt, ck, tgtval)
		if err = c.copyTo(params, originalValue, tgtval); err != nil {
			return
		}
		dbglog.Log("      <MAP> original item value: m[%v] => %v", tool.Valfmtptr(&ck), tool.Valfmt(&cv))
		return
	}
	dbglog.Log("      <VAL> duplicated/gotten: %v", tool.Valfmtptr(&tgtval))

	eltyp := tgt.Type().Elem() // get map value type
	eltypind, _ := tool.Rskiptype(eltyp, reflect.Ptr)

	var ptrToCopyValue, cv reflect.Value
	if eltypind.Kind() == reflect.Interface { //nolint:nestif //keep it
		var tt reflect.Type
		tgtvalind, _ := tool.Rdecode(tgtval)
		if !tgtval.IsValid() || tool.IsZero(tgtval) {
			srcval := src.MapIndex(ck)
			if !srcval.IsValid() || tool.IsZero(srcval) {
				tgtvalind = srcval
				tt = tgtvalind.Type()
			} else {
				tt = srcval.Type()
			}
		} else {
			tt = tgtvalind.Type()
		}
		dbglog.Log("  tgtval: [%v] %v, ind: %v | tt: %v", tool.Typfmtv(&tgtval), tool.Valfmt(&tgtval), tool.Typfmtv(&tgtvalind), tool.Typfmt(tt))
		ptrToCopyValue, cv = newFromType(tt)
		if processed, err = mapMergePreSetter(c, ck, cv); processed {
			return
		}
		defer trySetMapIndex(c, params, tgt, ck, cv)
		// defer func() {
		// 	trySetMapIndex(c, params, tgt, ck, cv)
		// 	dbglog.Log("  SetMapIndex: %v -> [%v] %v", ck.Interface(), cv.Type(), cv.Interface())
		// }()
	} else {
		ptrToCopyValue, cv = newFromType(eltypind)
		if processed, err = mapMergePreSetter(c, ck, cv); processed {
			return
		}
		defer matchTypeAndSetMapIndex(c, params, tgt, ck, ptrToCopyValue)
		// defer func() {
		// 	if cv.Type() == eltyp {
		// 		tgt.SetMapIndex(ck, cv)
		// 		dbglog.Log("  SetMapIndex: %v -> [%v] %v", ck.Interface(), cv.Type(), cv.Interface())
		// 	} else {
		// 		dbglog.Log("  SetMapIndex: %v -> [%v] %v", ck.Interface(), ptrToCopyValue.Type(), ptrToCopyValue.Interface())
		// 		tgt.SetMapIndex(ck, ptrToCopyValue)
		// 	}
		// }()
	}

	dbglog.Log("  ptrToCopyValue.type: %v, eltypind: %v", tool.Typfmtv(&ptrToCopyValue), tool.Typfmt(eltypind))
	if err = c.copyTo(params, tgtval, ptrToCopyValue); err != nil {
		return
	}
	if err = c.copyTo(params, originalValue, ptrToCopyValue); err != nil {
		return
	}

	return
}

func matchTypeAndSetMapIndex(c *cpController, params *Params, m, key, val reflect.Value) {
	ve := m.Type().Elem()
	cv := tool.Rindirect(val)
	if cv.Type() == ve {
		trySetMapIndex(c, params, m, key, cv)
		dbglog.Log("      <MAP> map.item set to val.ind: %v -> %v", tool.Valfmt(&key), tool.Valfmt(&cv))
	} else if val.Type() == ve {
		trySetMapIndex(c, params, m, key, val)
		dbglog.Log("      <MAP> map.item set to val: %v -> %v", tool.Valfmt(&key), tool.Valfmt(&val))
	} else if val.Kind() == reflect.Ptr && val.Type().Elem() == ve { // tgtval is ptr to elem? such as tgtval got *bool, and the map[ck] => bool
		trySetMapIndex(c, params, m, key, val.Elem())
		dbglog.Log("      <MAP> map.item set to val.elem: %v -> %v", tool.Valfmt(&key), tool.Valfmtv(val.Elem()))
		dbglog.Log("      <MAP> map: %v", tool.Valfmt(&m))
	} else if ve.Kind() == reflect.Interface { // the map is map[key]interface{} ?, so the val can be anything.
		if val.Kind() == reflect.Ptr {
			trySetMapIndex(c, params, m, key, val.Elem())
		} else {
			trySetMapIndex(c, params, m, key, val)
		}
	} else {
		log.Warnf("cannot setMapIndex for key '%v' since val type mismatched: desired elem typ = %v, real tgt val.typ = %v.", tool.Valfmt(&key), tool.Typfmt(ve), tool.Typfmtv(&val))
	}
}

func trySetMapIndex(c *cpController, params *Params, m, key, val reflect.Value) {
	// dbglog.Colored(color.LightMagenta, "    setting map index: %v -> %v", tool.Valfmt(&key), tool.Valfmt(&val))
	if params != nil && params.controller != nil && params.accessor != nil && params.controller.copyUnexportedFields {
		if fld := params.accessor.StructField(); fld != nil {
			// in a struct
			if !tool.IsExported(fld) {
				dbglog.Log("    unexported field %q (typ: %v): key '%v' => val '%v'",
					fld.Name, tool.Typfmt(fld.Type), tool.Valfmt(&key), tool.Valfmt(&val))
				cl.SetUnexportedFieldIfMap(m, key, val)
				return
			}
		}
	}

	if params != nil && params.resultForNewSlice != nil { // specially for value is a slice
		m.SetMapIndex(key, *params.resultForNewSlice)
		params.resultForNewSlice = nil
		return
	}

	m.SetMapIndex(key, val)
}

func newFromType(typ reflect.Type) (valptr, val reflect.Value) {
	if k := typ.Kind(); k == reflect.Ptr {
		vp := reflect.New(typ)
		v := vp.Elem()
		ptr, _ := newFromType(typ.Elem())
		v.Set(ptr)
		valptr, val = vp, v // .Elem()
		dbglog.Log("creating new object for type %v: %+v", tool.Typfmt(typ), tool.Valfmt(&valptr))
	} else if k == reflect.Slice {
		valptr = reflect.MakeSlice(typ, 0, 0)
		val = valptr
	} else if k == reflect.Map { //nolint:gocritic //keep it
		valptr = reflect.MakeMap(typ)
		val = valptr
	} else if k == reflect.Chan {
		valptr = reflect.MakeChan(typ, 0)
		val = valptr
	} else if k == reflect.Func {
		valptr = reflect.MakeFunc(typ, func(args []reflect.Value) []reflect.Value { return args })
		val = valptr
	} else {
		valptr = reflect.New(typ)
		val = valptr.Elem()
	}
	return
}

// newFromTypeEspSlice makes new instance for a given type.
// especially for a slice, newFromTypeEspSlice will make a pointer
// to a new slice, so that we can set the whole slice via this
// pointer later.
func newFromTypeEspSlice(typ reflect.Type) (val, valelem reflect.Value) {
	if k := typ.Kind(); k == reflect.Ptr {
		valelem, _ = newFromTypeEspSlice(typ.Elem())
		val = valelem.Addr()
	} else if k == reflect.Slice {
		var tp = tool.PointerTo(typ)
		val = reflect.New(tp).Elem()
		// valval := reflect.MakeSlice(typ, 0, 0)
		// val.Set(valval.Addr())
		val.Set(reflect.New(typ))
		valelem = val
	} else {
		val, valelem = newFromType(typ)
	}
	return
}

//nolint:lll //keep it
func ensureMapPtrValue(c *cpController, params *Params, m, key, originalValue reflect.Value) (val reflect.Value, ptr bool, err error) {
	val = m.MapIndex(key)
	vk := val.Kind()
	if vk == reflect.Ptr { //nolint:nestif //keep it
		if tool.IsNil(val) {
			typOfValueOfMap := m.Type().Elem()     // make new instance of type pointed by pointer
			val, _ = newFromType(typOfValueOfMap)  //
			trySetMapIndex(c, params, m, key, val) // and set the new pointer into map
			ptr = true                             //
			dbglog.Log("    ensureMapPtrValue:val.typ: %v, key.typ: %v | '%v' -> %v", tool.Typfmt(typOfValueOfMap), tool.Typfmtv(&key), tool.Valfmt(&key), tool.Valfmtptr(&val))
		} else {
			// dbglog.Log("    ensureMapPtrValue: do nothing because val's is not nil")
		}
	} else if !val.IsValid() {
		typOfValueOfMap := m.Type().Elem()
		if typOfValueOfMap.Kind() != reflect.Interface {
			var valelem reflect.Value
			val, valelem = newFromType(typOfValueOfMap)
			dbglog.Log("    ensureMapPtrValue:val.typ: %v, key.typ: %v | '%v' -> %v", tool.Typfmt(typOfValueOfMap), tool.Typfmtv(&key), tool.Valfmt(&key), tool.Valfmtptr(&valelem))
			trySetMapIndex(c, params, m, key, valelem)
			ptr = true // val = vind
		} else if originalValue.IsValid() && !tool.IsZero(originalValue) { // if original value is zero, no copying needed.
			typ := tool.Rdecodesimple(originalValue).Type()
			val, _ = newFromTypeEspSlice(typ)
			ptr = true
			dbglog.Log("    ensureMapPtrValue:val.typ: %v, key.typ: %v | '%v' -> %v", tool.Typfmt(typ), tool.Typfmtv(&key), tool.Valfmt(&key), tool.Valfmtptr(&val))
			if err = c.copyTo(params, originalValue, val); err != nil {
				return
			}
			var processed bool
			if processed, err = mapMergePreSetter(c, key, val); processed {
				return
			}
			trySetMapIndex(c, params, m, key, val)
			dbglog.Log("    ensureMapPtrValue:val.typ: %v, key.typ: %v | '%v' -> %v | DONE", tool.Typfmt(typ), tool.Typfmtv(&key), tool.Valfmt(&key), tool.Valfmtptr(&val))
		} else {
			// dbglog.Log("    ensureMapPtrValue: do nothing because val and src-val are both invalid, so needn't copy.")
		}
	}
	return
}

func cloneMapKey(c *cpController, params *Params, tgt, key reflect.Value) (ck reflect.Value, err error) {
	keyType := tgt.Type().Key()
	ptrToCopyKey := reflect.New(keyType)
	dbglog.Log("     cloneMapKey(%v): tgt(map).type: %v, tgt.key.type: %v, ptrToCopyKey.type: %v", tool.Valfmt(&key), tool.Typfmtv(&tgt), tool.Typfmt(keyType), tool.Typfmtv(&ptrToCopyKey))
	ck = ptrToCopyKey.Elem()

	emptyParams := newParams() // use an empty params for just copying map key so the current processing struct fields won't be cared in this special child copier
	if err = c.copyTo(emptyParams, key, ck); err != nil {
		dbglog.Err("     cloneMapKey(%v) error on copyTo: %+v", tool.Valfmt(&key), err) // early break-point here
		return
	}

	dbglog.Log("         <KEY> cloned: '%v'", tool.Valfmtptr(&ck))
	return
}

//

//

//

func copyUintptr(c *cpController, params *Params, from, to reflect.Value) (err error) {
	tgt := tool.Rindirect(to)
	if tgt.CanSet() {
		tgt.Set(from)
	} else {
		// to.SetPointer(from.Pointer())
		dbglog.Log("    copy uintptr not support: %v -> %v", from.Kind(), to.Kind())
		err = ErrCannotCopy.FormatWith(tool.Valfmt(&from), tool.Typfmtv(&from), tool.Valfmt(&to), tool.Typfmtv(&to))
	}
	return
}

func copyUnsafePointer(c *cpController, params *Params, from, to reflect.Value) (err error) {
	tgt := tool.Rindirect(to)
	if tgt.CanSet() {
		tgt.Set(from)
	} else {
		dbglog.Log("    copy unsafe pointer not support: %v -> %v", from.Kind(), to.Kind())
		err = ErrCannotCopy.FormatWith(tool.Valfmt(&from), tool.Typfmtv(&from), tool.Valfmt(&to), tool.Typfmtv(&to))
	}
	return
}

// copyFunc never used.
// Deprecated always.
func copyFunc(c *cpController, params *Params, from, to reflect.Value) (err error) { //nolint:unused,deadcode //reserved
	tgt := tool.Rindirect(to)
	if tgt.CanSet() {
		tgt.Set(from)
		return
	}

	k := tgt.Kind()
	if k != reflect.Func && c.copyFunctionResultToTarget {
		// from.
		return
	}

	if k == reflect.Func {
		if !params.processUnexportedField(to, from) {
			ptr := from.Pointer()
			//goland:noinspection GoVetUnsafePointer
			to.SetPointer(unsafe.Pointer(ptr))
		}
		dbglog.Log("    function pointer copied: %v (%v) -> %v", from.Kind(), from.Interface(), to.Kind())
	} else {
		err = ErrCannotCopy.FormatWith(tool.Valfmt(&from), tool.Typfmtv(&from), tool.Valfmt(&to), tool.Typfmtv(&to))
	}

	return
}

func copyChan(c *cpController, params *Params, from, to reflect.Value) (err error) {
	tgt := tool.Rindirect(to)
	if tgt.CanSet() {
		dbglog.Log("    copy chan: %v (%v) -> %v (%v)", from.Kind(), from.Type(), tgt.Kind(), tgt.Type())
		tgt.Set(from)
		// Log("        after: %v -> %v", from.Interface(), tgt.Interface())
	} else {
		v := reflect.MakeChan(from.Type(), from.Cap())
		to.Set(v)
		// // to.SetPointer(from.Pointer())
		// dbglog.Log("    copy chan not support: %v (%v) -> %v (%v)", from.Kind(), from.Type(), to.Kind(), to.Type())
		// err = ErrCannotCopy.FormatWith(tool.Valfmt(&from), tool.Typfmtv(&from), tool.Valfmt(&to), tool.Typfmtv(&to))
	}
	return
}

//

func copyDefaultHandler(c *cpController, params *Params, from, to reflect.Value) (err error) {
	sourceType, targetType := from.Type(), to.Type()

	if c != nil {
		if cvt, ctx := c.valueCopiers.findCopiers(params, sourceType, targetType, false); cvt != nil {
			err = cvt.CopyTo(ctx, from, to)
			return
		}
	}

	fromind, toind := tool.Rdecodesimple(from), tool.Rdecodesimple(to)
	dbglog.Log("  copyDefaultHandler: %v -> %v | %v", tool.Typfmtv(&fromind), tool.Typfmtv(&toind), tool.Typfmtv(&to))

	// //////////////// source is primitive types but target isn't its
	var processed bool
	var toIndType = targetType
	if toind.IsValid() {
		toIndType = toind.Type()
	}
	processed, err = copyPrimitiveToComposite(c, params, from, to, toIndType)
	if processed || err != nil {
		return
	}

	// create new
	if !toind.IsValid() && to.Kind() == reflect.Ptr {
		tgt := reflect.New(targetType.Elem())
		toind = tgt.Elem()
		defer func() {
			if err == nil {
				to.Set(tgt)
			}
		}()
	}

	// try primitive -> primitive at first
	if processed, err = tryCopyPrimitiveToPrimitive(params, from, fromind, to, toind, sourceType, targetType, toIndType); processed { //nolint:lll //keep it
		return
	}

	err = ErrCannotConvertTo.FormatWith(fromind.Interface(), fromind.Kind(), toind.Interface(), toind.Kind())
	dbglog.Err("    Error: %v", err)
	return
}

//nolint:lll //keep it
func tryCopyPrimitiveToPrimitive(params *Params, from, fromind, to, toind reflect.Value, sourceType, targetType, toIndType reflect.Type) (stop bool, err error) {
	stop = true
	if tool.CanConvert(&fromind, toIndType) {
		var val = fromind.Convert(toIndType)
		err = setTargetValue1(params, to, toind, val)
		return
	}
	if tool.CanConvert(&from, to.Type()) && to.CanSet() {
		var val = from.Convert(to.Type())
		err = setTargetValue1(params, to, toind, val)
		return
	}
	if sourceType.AssignableTo(targetType) {
		if toind.CanSet() { //nolint:gocritic // no need to switch to 'switch' clause
			toind.Set(fromind)
		} else if to.CanSet() {
			to.Set(fromind)
		} else {
			err = ErrCannotSet.FormatWith(fromind, tool.Typfmtv(&fromind), toind, tool.Typfmtv(&toind))
		}
		return
	}
	stop = false
	return
}

//nolint:lll //keep it
func copyPrimitiveToComposite(c *cpController, params *Params, from, to reflect.Value, desiredType reflect.Type) (processed bool, err error) {
	switch tk := desiredType.Kind(); tk { //nolint:exhaustive //no need
	case reflect.Slice:
		dbglog.Log("  copyPrimitiveToComposite: %v -> %v | %v", tool.Typfmtv(&from), tool.Typfmt(desiredType), tool.Typfmtv(&to))

		eltyp := desiredType.Elem()
		elnew := reflect.New(eltyp)
		if err = copyDefaultHandler(c, params, from, elnew); err != nil {
			return
		}

		elnewelem := elnew.Elem()
		dbglog.Log("    source converted: %v (%v)", tool.Valfmt(&elnewelem), tool.Typfmtv(&elnewelem))

		slice := reflect.MakeSlice(reflect.SliceOf(eltyp), 1, 1)
		slice.Index(0).Set(elnewelem)
		dbglog.Log("    source converted: %v (%v)", tool.Valfmt(&slice), tool.Typfmtv(&slice))

		err = copySlice(c, params, slice, to)
		processed = true

	case reflect.Map:
		// not support

	case reflect.Struct:
		// not support

	case reflect.Func:
		tgt := tool.Rdecodesimple(to)
		processed, err = true, copyToFuncImpl(c, from, tgt, tgt.Type())
	}

	return
}

func setTargetValue1(params *Params, to, toind, newval reflect.Value) (err error) {
	if err = setTargetValue2(params, toind, newval); err == nil {
		return //nolint:nilerr //i do
	}
	if to != toind { //nolint:govet //how should i do //TODO needs checked-review
		if err = setTargetValue2(params, to, newval); err == nil {
			return //nolint:nilerr //i do
		}
	}
	if err == nil {
		err = ErrUnknownState.WithTaggedData(errors.TaggedData{ // record the sites
			"source": tool.Valfmt(&toind),
			"target": tool.Valfmt(&newval),
		})
	}
	return
}

func setTargetValue2(params *Params, to, newval reflect.Value) (err error) {
	if copyUnexportedFields, isExported := params.dstFieldIsExportedR(); !isExported {
		if copyUnexportedFields {
			cl.SetUnexportedField(to, newval)
		} // else do nothing
		return
	}
	
	// const unexportedR = true
	// if unexportedR {
	// 	if copyUnexportedFields, isExported := params.dstFieldIsExportedR(); !isExported {
	// 		if copyUnexportedFields {
	// 			cl.SetUnexportedField(to, newval)
	// 		} // else do nothing
	// 		return
	// 	}
	// } else {
	// 	k := reflect.Invalid
	// 	if params != nil && params.dstDecoded != nil {
	// 		k = params.dstDecoded.Kind()
	// 	}
	// 
	// 	if k == reflect.Struct && params.accessor != nil && !tool.IsExported(params.accessor.StructField()) {
	// 		if params.controller.copyUnexportedFields {
	// 			cl.SetUnexportedField(to, newval)
	// 		} // else ignore the unexported field
	// 		return
	// 	}
	// }

	if to.CanSet() {
		to.Set(newval)
		return
		// } else if newval.IsValid() {
		//	to.Set(newval)
		//	return
	}

	err = ErrUnknownState.WithTaggedData(errors.TaggedData{ // record the sites
		"source": tool.Valfmt(&to),
		"target": tool.Valfmt(&newval),
	})
	return
}
