package evendeep

// ftor.go - functors
//

import (
	"github.com/hedzr/evendeep/flags/cms"
	"github.com/hedzr/evendeep/internal/cl"
	"github.com/hedzr/evendeep/internal/dbglog"
	"github.com/hedzr/evendeep/internal/tool"
	"github.com/hedzr/log"

	"gopkg.in/hedzr/errors.v3"

	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

func copyPointer(c *cpController, params *Params, from, to reflect.Value) (err error) {
	// from is a pointer

	src := tool.Rindirect(from)
	tgt := tool.Rindirect(to)

	newobj := func(c *cpController, params *Params, src, to, tgt reflect.Value) (err error) {
		newtyp := to.Type()
		if to.Type() == from.Type() {
			newtyp = newtyp.Elem() // is pointer and its same
		}
		// create new object and pointer
		toobjcopyptrv := reflect.New(newtyp)
		dbglog.Log("    toobjcopyptrv: %v", tool.Typfmtv(&toobjcopyptrv))
		if err = c.copyTo(params, src, toobjcopyptrv.Elem()); err == nil {
			val := toobjcopyptrv
			if to.Type() == from.Type() {
				val = val.Elem()
			}
			err = setTargetValue1(params.owner, to, tgt, val)
			// to.Set(toobjcopyptrv)
		}
		return
	}

	paramsChild := newParams(withOwners(c, params, &from, &to, nil, nil))
	defer paramsChild.revoke()

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
		err = newobj(c, paramsChild, src, to, tgt)
	}
	return //nolint:nakedret
}

func copyInterface(c *cpController, params *Params, from, to reflect.Value) (err error) {
	if tool.IsNil(from) { //nolint:gocritic // no need to switch to 'switch' clause
		if params.isGroupedFlagOKDeeply(cms.OmitIfNil, cms.OmitIfEmpty) {
			return
		}
		if to.CanSet() {
			to.Set(reflect.Zero(to.Type()))
			return
		}
		goto badReturn
	} else if tool.IsZero(from) {
		if params.isGroupedFlagOKDeeply(cms.OmitIfZero, cms.OmitIfEmpty) {
			return
		}
		if to.CanSet() {
			to.Set(reflect.Zero(to.Type()))
			return
		}
		goto badReturn
	} else {
		paramsChild := newParams(withOwners(c, params, &from, &to, nil, nil))
		defer paramsChild.revoke()

		// unbox the interface{} to original data type
		toind, toptr := tool.Rdecode(to) // c.skip(to, reflect.Interface, reflect.Pointer)

		dbglog.Log("from.type: %v, decode to: %v", from.Type().Kind(), paramsChild.srcDecoded.Kind())
		dbglog.Log("  to.type: %v, decode to: %v (ptr: %v) | CanSet: %v, CanAddr: %v", to.Type().Kind(), toind.Kind(), toptr.Kind(), toind.CanSet(), toind.CanAddr())

		// var merging = c.flags.isAnyFlagsOK(SliceMerge, MapMerge) || params.isAnyFlagsOK(SliceMerge, MapMerge)
		//nolint:gocritic // no need to switch to 'switch' clause
		if paramsChild.inMergeMode() || !c.makeNewClone {
			err = c.copyTo(paramsChild, *paramsChild.srcDecoded, toptr)
		} else if to.CanSet() {
			copyValue := reflect.New(paramsChild.srcDecoded.Type()).Elem()
			if err = c.copyTo(paramsChild, *paramsChild.srcDecoded, copyValue); err == nil {
				to.Set(copyValue)
			}
		} else {
			goto badReturn
		}
		return
	}

badReturn:
	err = ErrCannotSet.FormatWith(tool.Valfmt(&from), tool.Typfmtv(&from), tool.Valfmt(&to), tool.Typfmtv(&to))
	return //nolint:nakedret
}

func copyStruct(c *cpController, params *Params, from, to reflect.Value) (err error) {
	// default is cms.ByOrdinal:
	//   loops all source fields and copy its value to the corresponding target field.
	cb := forEachField
	if c.targetOriented || params.isGroupedFlagOKDeeply(cms.ByName) {
		// cmd.ByName strategy:
		//   loops all target fields and try copying value from source field by its name.
		cb = forEachTargetField
	}
	err = copyStructInternal(c, params, from, to, cb)
	return
}

func copyStructInternal(
	c *cpController, params *Params,
	from, to reflect.Value,
	fn func(paramsChild *Params, ec errors.Error, i, amount *int, padding string) (err error),
) (err error) {
	var (
		i, amount   int
		padding     string
		ec          = errors.New("copyStruct errors")
		paramsChild = newParams(withOwners(c, params, &from, &to, nil, nil))
	)

	defer paramsChild.revoke()
	defer ec.Defer(&err)

	defer func() {
		if e := recover(); e != nil {
			sst := paramsChild.targetIterator.(sourceStructFieldsTable) //nolint:errcheck //yes

			ff := sst.TableRecord(i).FieldValue()
			var tf = paramsChild.dstOwner
			var tft = &paramsChild.dstType
			if paramsChild.accessor != nil {
				tf = paramsChild.accessor.FieldValue()
				tft = paramsChild.accessor.FieldType()
			}

			ec.Attach(errors.New("[recovered] copyStruct unsatisfied ([%v] -> [%v]), causes: %v",
				tool.Typfmtv(ff), tool.Typfmtptr(tft), e).
				WithData(e).                      // collect e if it's an error object else store it simply
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
			log.Errorf("%+v", ec)
			// }
		}
	}()

	if dbglog.LogValid {
		// dbgFrontOfStruct(params, paramsChild, padding, func(msg string, args ...interface{}) { dbglog.Log(msg, args...) })
		dbgFrontOfStruct(paramsChild, padding, dbglog.Log)
	}

	var processed bool
	if processed, err = tryConverters(c, paramsChild, &from, paramsChild.dstDecoded, &paramsChild.dstType, true); processed {
		return
	}

	switch k := paramsChild.dstDecoded.Kind(); k { //nolint:exhaustive
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
				err = ErrCannotSet.FormatWith(tool.Valfmt(paramsChild.srcDecoded),
					tool.Typfmtv(paramsChild.srcDecoded), tool.Valfmt(paramsChild.dstDecoded),
					tool.Typfmtv(paramsChild.dstDecoded))
			}
		}
		return
	}

	err = fn(paramsChild, ec, &i, &amount, padding)
	ec.Attach(err)
	return //nolint:nakedret
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

func forEachTargetField(params *Params, ec errors.Error, i, amount *int, padding string) (err error) {
	c := params.controller

	var sst = params.targetIterator.(sourceStructFieldsTable) //nolint:errcheck
	var val reflect.Value
	var fcz = params.isGroupedFlagOKDeeply(cms.ClearIfMissed)
	var aun = c.autoNewStruct // autoNew mode:
	// we will do new(field) for each target fields if it's invalid.
	//
	// It's not cms.ClearIfInvalid - which will detect if the source
	// field is invalid or not, but target one.
	dbglog.Log("     c.autoNewStruct = %v, cms.ClearIfMissed is set: %v", c.autoNewStruct, params.isGroupedFlagOKDeeply(cms.ClearIfMissed))
	dbglog.Log("     c.copyFunctionResultToTarget = %v", c.copyFunctionResultToTarget)

	for *i, *amount = 0, len(sst.TableRecords()); params.nextTargetFieldLite(); *i++ {
		name := params.accessor.StructFieldName()
		if params.shouldBeIgnored(name) {
			continue
		}

		if ex := c.sourceExtractor; ex != nil {
			v := ex(name)
			val = reflect.ValueOf(v)
		} else {
			if ind := sst.RecordByName(name); ind != nil {
				val = *ind
			} else if c.copyFunctionResultToTarget {
				if _, ind = sst.MethodCallByName(name); ind != nil {
					val = *ind
				} else {
					continue // skip the field
				}
			} else if fcz || (aun && !params.accessor.ValueValid()) {
				tt := params.accessor.FieldType()
				val = reflect.Zero(*tt)
				dbglog.Log("     target is invalid: %v, autoNewStruct: %v", params.accessor.ValueValid(), aun)
			} else {
				continue // skip the field
			}
		}
		params.accessor.Set(val)
	}
	return
}

func forEachField(params *Params, ec errors.Error, i, amount *int, padding string) (err error) {
	sst := params.targetIterator.(sourceStructFieldsTable) //nolint:errcheck
	c := params.controller

	for *i, *amount = 0, len(sst.TableRecords()); *i < *amount; *i++ {
		if params.sourceFieldShouldBeIgnored() {
			sst.Step(1) // step the source field index(pointer) backward
			continue
		}

		var sourceField *tableRecT
		var ok bool
		if sourceField, ok = params.nextTargetField(); !ok {
			continue
		}

		srcval, dstval := sourceField.FieldValue(), params.accessor.FieldValue()
		log.VDebugf("%d. %s (%v) -> %s (%v) | (%v) -> (%v)", i,
			sourceField.FieldName(), tool.Typfmtv(srcval), params.accessor.StructFieldName(),
			tool.Typfmt(*params.accessor.FieldType()), tool.Valfmt(srcval), tool.Valfmt(dstval))

		if srcval != nil && dstval != nil && srcval.IsValid() {
			typ := params.accessor.FieldType()
			if err = invokeStructFieldTransformer(c, params, srcval, dstval, typ, padding); err != nil {
				ec.Attach(err)
				log.Errorf("error: %v", err)
			}
			continue
		}

		if params.inMergeMode() {
			typ := params.accessor.FieldType()
			dbglog.Log("    new object for %v", tool.Typfmt(*typ))

			// create new object and pointer
			toobjcopyptrv := reflect.New(*typ).Elem()
			dbglog.Log("    toobjcopyptrv: %v", tool.Typfmtv(&toobjcopyptrv))

			//nolint:gocritic // no need to switch to 'switch' clause
			if err = invokeStructFieldTransformer(c, params, srcval, &toobjcopyptrv, typ, padding); err != nil {
				ec.Attach(err)
				log.Errorf("error: %v", err)
			} else if toobjcopyptrv.Kind() == reflect.Slice {
				params.accessor.Set(toobjcopyptrv)
			} else {
				params.accessor.Set(toobjcopyptrv.Elem())
			}
			continue
		}

		dbglog.Log("   ignore nil/zero/invalid source or nil target")
	}
	return //nolint:nakedret
}

func dbgFrontOfStruct(params *Params, padding string, logger func(msg string, args ...interface{})) {
	if params == nil {
		return
	}
	if logger == nil {
		logger = dbglog.Log
	}
	if log.VerboseEnabled {
		d := params.depth()
		if d > 1 {
			d -= 2
		}
		padding1 := strings.Repeat("  ", d*2) //nolint:gomnd
		// fromT, toT := params.srcDecoded.Type(), params.dstDecoded.Type()
		// Log(" %s  %d, %d, %d", padding, params.index, params.srcOffset, params.dstOffset)
		// fq := dbgMakeInfoString(fromT, params.owner, true, logger)
		// dq := dbgMakeInfoString(toT, params.owner, false, logger)
		logger(" %s- src (%v) -> dst (%v)", padding1, tool.Typfmtv(params.srcDecoded), tool.Typfmtv(params.dstDecoded))
		// logger(" %s  %s -> %s", padding, fq, dq)
	}
}

// func dbgMakeInfoString(typ reflect.Type, params *Params, src bool, logger func(msg string, args ...interface{})) (qstr string) {
//	if typ.Kind() == reflect.Struct && params != nil && params.accessor != nil && params.accessor.StructField() != nil {
//		qstr = dbgMakeFieldInfoString(params.accessor.StructField(), params.owner, logger)
//	} else {
//		qstr = fmt.Sprintf("%v (%v)", typ, typ.Kind())
//	}
//	return
// }
//
// func dbgMakeFieldInfoString(fld *reflect.StructField, params *Params, logger func(msg string, args ...interface{})) (qstr string) {
//	ft := fld.Type
//	ftk := ft.Kind()
//	if params != nil {
//		qstr = fmt.Sprintf("Field%v %q (%v (%v)) | %v", fld.Index, fld.Name, ft, ftk, fld)
//	} else {
//		qstr = fmt.Sprintf("Field%v %q (%v (%v)) | %v", fld.Index, fld.Name, ft, ftk, fld)
//	}
//	return
// }

func invokeStructFieldTransformer(
	c *cpController, params *Params, ff, df *reflect.Value,
	dftyp *reflect.Type, //nolint:gocritic // ptrToRefParam: consider `dftyp' to be of non-pointer type
	padding string,
) (err error) {
	fv, dv := ff != nil && ff.IsValid(), df != nil && df.IsValid()
	fft, dft := dtypzz(ff, dftyp), dtypzz(df, dftyp)
	fftk, dftk := fft.Kind(), dft.Kind()

	var processed bool
	if processed = checkClearIfEqualOpt(params, ff, df, dft); processed {
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
		log.Errorf("shouldn't get into here because we have a failover branch at the callee")
	}

	if fv && dv {
		err = c.copyTo(params, *ff, *df) // or, use internal standard implementation version
		return
	}

	return forInvalidValues(c, params, ff, fft, dft, fftk, dftk, fv)
}

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

func dtypzz(df *reflect.Value, deftyp *reflect.Type) reflect.Type { //nolint:gocritic // ptrToRefParam: consider `dftyp' to be of non-pointer type
	if df != nil && df.IsValid() {
		return df.Type()
	}
	return *deftyp
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
	dftyp *reflect.Type, //nolint:gocritic // ptrToRefParam: consider `dftyp' to be of non-pointer type
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

func findAndApplyCopiers(c *cpController, params *Params, ff, df *reflect.Value, fft, dft reflect.Type, userDefinedOnly bool) (processed bool, err error) {
	if cvt, ctx := c.valueCopiers.findCopiers(params, fft, dft, userDefinedOnly); ctx != nil {
		dbglog.Log("-> using Copier %v", reflect.ValueOf(cvt).Type())

		if df.IsValid() {
			if err = cvt.CopyTo(ctx, *ff, *df); err == nil { // use user-defined copy-n-merger to merge or copy source to destination
				processed = true
			}
			return
		}

		if dft.Kind() == reflect.Interface {
			dft = fft
		}
		dbglog.Log("  dft: %v", tool.Typfmt(dft))
		nv := reflect.New(dft)
		err = cvt.CopyTo(ctx, *ff, nv) // use user-defined copy-n-merger to merge or copy source to destination
		if err == nil && !params.accessor.IsStruct() {
			params.accessor.Set(nv.Elem())
			processed = true
		}
	}
	return
}

func findAndApplyConverters(c *cpController, params *Params, ff, df *reflect.Value, fft, dft reflect.Type, userDefinedOnly bool) (processed bool, err error) {
	if cvt, ctx := c.valueConverters.findConverters(params, fft, dft, userDefinedOnly); ctx != nil {
		dbglog.Log("-> using Converter %v", reflect.ValueOf(cvt).Type())
		var result reflect.Value
		result, err = cvt.Transform(ctx, *ff, dft) // use user-defined value converter to transform from source to destination
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

// copySlice transforms from slice to target with slice or other types
func copySlice(c *cpController, params *Params, from, to reflect.Value) (err error) {
	if from.IsNil() && params.isGroupedFlagOKDeeply(cms.OmitIfNil, cms.OmitIfEmpty) { // an empty slice found
		return
	}

	var tgt, tgtptr reflect.Value
	tgt, tgtptr = tool.Rdecode(to)
	if to != tgtptr {
		err = c.copyTo(params, from, tgtptr) // unwrap the pointer
		return
	}

	if params.controller != c {
		log.Panicf("[copySlice] c *cpController != params.controller, what's up??")
	}

	tk, typ := tgt.Kind(), tgt.Type()
	if tk != reflect.Slice {
		dbglog.Log("from slice -> %v", tool.Typfmt(typ))
		var processed bool
		if processed, err = tryConverters(c, params, &from, &tgt, &typ, false); !processed {
			// log.Panicf("[copySlice] unsupported transforming: from slice -> %v,", typfmtv(&tgt))
			err = ErrCannotCopy.WithErrors(err).FormatWith(tool.Valfmt(&from), tool.Typfmtv(&from), tool.Valfmt(&tgt), tool.Typfmtv(&tgt))
		}
		return
	}

	if tool.IsNil(tgt) && params.isGroupedFlagOKDeeply(cms.OmitIfTargetZero, cms.OmitIfTargetEmpty) {
		return
	}

	err = copySliceInternal(c, params, from, to, tgt, tgtptr)
	return //nolint:nakedret
}

func copySliceInternal(c *cpController, params *Params, from, to, tgt, tgtptr reflect.Value) (err error) {
	ec := errors.New("slice copy/merge errors")
	defer ec.Defer(&err)

	for _, flag := range []cms.CopyMergeStrategy{cms.SliceMerge, cms.SliceCopyAppend, cms.SliceCopy} {
		if params.isGroupedFlagOKDeeply(flag) { //nolint:gocritic // nestingReduce: invert if cond, replace body with `continue`, move old body after the statement
			// if !to.CanAddr() {
			//	if params != nil && !params.isStruct() {
			//		to = *params.dstOwner
			//		Log("use dstOwner to get a ptr to slice, new to.type: %v, canAddr: %v, canSet: %v", to.Type().Kind(), to.CanAddr(), to.CanSet())
			//	}
			// }

			// src, _ = c.decode(from)

			dbglog.Log("slice merge mode: %v", flag)
			dbglog.Log("from.type: %v", from.Type().Kind())
			dbglog.Log("  to.type: %v, canAddr: %v, canSet: %v", tool.Typfmtv(&to), to.CanAddr(), to.CanSet())
			// Log(" src.type: %v, len: %v, cap: %v, srcptr.canAddr: %v", src.Type().Kind(), src.Len(), src.Cap(), srcptr.CanAddr())
			dbglog.Log(" tgt.type: %v, tgtptr: %v .canAddr: %v", tool.Typfmtv(&tgt), tool.Typfmtv(&tgtptr), tgtptr.CanAddr())

			if fn, ok := getSliceOperations()[flag]; ok {
				if result, e := fn(c, params, from, tgt); e == nil {
					// tgt=ns
					// t := c.want2(to, reflect.Slice, reflect.Interface)
					// t.Set(tgt)
					//   //tgtptr.Elem().Set(result)
					if tk := tgtptr.Kind(); tk == reflect.Slice || tk == reflect.Interface {
						tgtptr.Set(result)
					} else {
						tgtptr.Elem().Set(result)
					}
				} else {
					ec.Attach(e)
				}
			} else {
				ec.Attach(errors.New("cannot make slice copy, unknown copy-merge-strategy %v", flag))
			}

			break
		}
	}

	return //nolint:nakedret
}

type fnSliceOperator func(c *cpController, params *Params, src, tgt reflect.Value) (result reflect.Value, err error)
type mSliceOperations map[cms.CopyMergeStrategy]fnSliceOperator

func getSliceOperations() (mapOfSliceOperations mSliceOperations) {
	mapOfSliceOperations = mSliceOperations{
		cms.SliceCopy:       _sliceCopyOperation,
		cms.SliceCopyAppend: _sliceCopyAppendOperation,
		cms.SliceMerge:      _sliceMergeOperation,
	}
	return
}

// _sliceCopyOperation: for SliceCopy, target elements will be given up, and source copied to.
func _sliceCopyOperation(c *cpController, params *Params, src, tgt reflect.Value) (result reflect.Value, err error) {
	slice := reflect.MakeSlice(tgt.Type(), 0, 0)
	dbglog.Log("tgt slice: %v, el: %v", tgt.Type(), tgt.Type().Elem())

	ecTotal := errors.New("slice merge errors (%v -> %v)", src.Type(), tgt.Type())
	defer ecTotal.Defer(&err)

	for _, ss := range []struct {
		length int
		source reflect.Value
	}{
		// {tl, tgt},
		{src.Len(), src},
	} {
		slice, err = _sliceCopyOne(c, params, ecTotal, slice, ss.length, ss.source, tgt)
	}
	result = slice
	return
}

// _sliceCopyAppendOperation: for SliceCopyAppend, target and source elements will be copied to new target.
// The duplicated elements were kept.
func _sliceCopyAppendOperation(c *cpController, params *Params, src, tgt reflect.Value) (result reflect.Value, err error) {
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
		ns, err = _sliceCopyOne(c, params, ecTotal, ns, ss.length, ss.source, tgt)
	}
	result = ns
	return
}

func _sliceCopyOne(c *cpController, params *Params, ecTotal errors.Error, slice reflect.Value, sslength int, sssource, tgt reflect.Value) (result reflect.Value, err error) {
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

		if el.Type() == tgtelemtype {
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
	result = slice
	return //nolint:nakedret
}

// _sliceMergeOperation: for SliceMerge. target and source elements will be copied to new target
// with uniqueness.
func _sliceMergeOperation(c *cpController, params *Params, src, tgt reflect.Value) (result reflect.Value, err error) {
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
				elv   = el.Interface()
				enew  = el
				ec    = errors.New("cannot convert %v to %v", el.Type(), tgtelemtype)
			)
			if el.Type() != tgtelemtype {
				if cc, ctx := c.valueConverters.findConverters(params, el.Type(), tgtelemtype, false); cc != nil {
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

			found = tool.FindInSlice(ns, elv, i)

			if !found {
				if cvtok || el.Type() == tgtelemtype {
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
			ecTotal.Attach(ec)
		}
	}
	result = ns
	return //nolint:nakedret
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
	//		Log("    !! use dstOwner to get a ptr to array, new to.type: %v, canAddr: %v, canSet: %v", to.Type().Kind(), to.CanAddr(), to.CanSet())
	//	}
	// }

	// if tgt.CanAddr() == false && tgtptr.CanAddr() {
	//	tgt = tgtptr
	// }

	// Log("    tgt.%v: %v", params.dstOwner.Type().Field(params.index).Name, params.dstOwner.Type().Field(params.index))
	dbglog.Log("    from.type: %v, len: %v, cap: %v", src.Type().Kind(), src.Len(), src.Cap())
	dbglog.Log("      to.type: %v, len: %v, cap: %v, tgtptr.canSet: %v, tgtptr.canaddr: %v", tgt.Type().Kind(), tgt.Len(), tgt.Cap(), tgtptr.CanSet(), tgtptr.CanAddr())

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

	return //nolint:nakedret
}

func copyMap(c *cpController, params *Params, from, to reflect.Value) (err error) {
	if from.IsNil() && params.isGroupedFlagOKDeeply(cms.OmitIfNil, cms.OmitIfEmpty) { // an empty slice found
		return
	}

	var tgt, tgtptr reflect.Value
	tgt, tgtptr = tool.Rdecode(to)
	if to != tgtptr {
		err = c.copyTo(params, from, tgtptr) // unwrap the pointer
		return
	}

	tk, typ := tgt.Kind(), tgt.Type()
	if tk != reflect.Map {
		dbglog.Log("from map -> %v", tool.Typfmt(typ))
		// copy map to String, Slice, Struct
		var processed bool
		if processed, err = tryConverters(c, params, &from, &tgt, &typ, false); !processed {
			err = ErrCannotCopy.WithErrors(err).FormatWith(tool.Valfmt(&from), tool.Typfmtv(&from), tool.Valfmt(&tgt), tool.Typfmtv(&tgt))
		}
		return
	}

	if tool.IsNil(tgt) && params.isGroupedFlagOKDeeply(cms.OmitIfTargetZero, cms.OmitIfTargetEmpty) {
		return
	}

	ec := errors.New("map copy/merge errors")
	defer ec.Defer(&err)

	for _, flag := range []cms.CopyMergeStrategy{cms.MapMerge, cms.MapCopy} {
		if params.isGroupedFlagOKDeeply(flag) {
			if fn, ok := getMapOperations()[flag]; ok {
				ec.Attach(fn(c, params, from, tgt))
			} else {
				ec.Attach(errors.New("unknown strategy for map: %v", flag))
			}
			break
		}
	}
	return //nolint:nakedret
}

type fnMapOperation func(c *cpController, params *Params, src, tgt reflect.Value) (err error)
type mapMapOperations map[cms.CopyMergeStrategy]fnMapOperation

func getMapOperations() (mMapOperations mapMapOperations) {
	mMapOperations = mapMapOperations{
		cms.MapCopy: func(c *cpController, params *Params, src, tgt reflect.Value) (err error) {
			tgt.Set(reflect.MakeMap(src.Type()))

			ec := errors.New("map copy errors")
			defer ec.Defer(&err)

			for _, key := range src.MapKeys() {
				originalValue := src.MapIndex(key)
				copyValue := reflect.New(tgt.Type().Elem())
				ec.Attach(c.copyTo(params, originalValue, copyValue.Elem()))

				copyKey := reflect.New(tgt.Type().Key())
				ec.Attach(c.copyTo(params, key, copyKey.Elem()))

				tgt.SetMapIndex(copyKey.Elem(), copyValue.Elem())
			}
			return
		},
		cms.MapMerge: func(c *cpController, params *Params, src, tgt reflect.Value) (err error) {
			ec := errors.New("map merge errors")
			defer ec.Defer(&err)

			for _, key := range src.MapKeys() {
				ec.Attach(mergeOneKeyInMap(c, params, src, tgt, key))
			}
			return
		},
	}
	return //nolint:nakedret
}

func mergeOneKeyInMap(c *cpController, params *Params, src, tgt, key reflect.Value) (err error) {
	originalValue := src.MapIndex(key)

	keyType := tgt.Type().Key()
	ptrToCopyKey := reflect.New(keyType)
	dbglog.Log("  tgt.type: %v, ptrToCopyKey.type: %v", tool.Typfmtv(&tgt), tool.Typfmtv(&ptrToCopyKey))
	ck := ptrToCopyKey.Elem()
	if err = c.copyTo(params, key, ptrToCopyKey); err != nil {
		return
	}

	tgtval, newelemcreated := ensureMapPtrValue(tgt, ck)
	if newelemcreated {
		if err = c.copyTo(params, originalValue, tgtval); err != nil {
			return
		}
		tgtval = tgtval.Elem()
		dbglog.Log("  Update Map: %v -> %v", ck.Interface(), tgtval.Interface())
		return
	}

	eltyp := tgt.Type().Elem() // get map value type
	eltypind, _ := tool.Rskiptype(eltyp, reflect.Ptr)

	var ptrToCopyValue, cv reflect.Value
	if eltypind.Kind() == reflect.Interface {
		tgtvalind, _ := tool.Rdecode(tgtval)
		dbglog.Log("  tgtval: [%v] %v, ind: %v", tool.Typfmtv(&tgtval), tgtval.Interface(), tool.Typfmtv(&tgtvalind))
		ptrToCopyValue = reflect.New(tgtvalind.Type())
		cv = ptrToCopyValue.Elem()
		defer func() {
			tgt.SetMapIndex(ck, cv)
			dbglog.Log("  SetMapIndex: %v -> [%v] %v", ck.Interface(), cv.Type(), cv.Interface())
		}()
	} else {
		ptrToCopyValue = reflect.New(eltypind)
		cv = ptrToCopyValue.Elem()
		defer func() {
			if cv.Type() == eltyp {
				tgt.SetMapIndex(ck, cv)
				dbglog.Log("  SetMapIndex: %v -> [%v] %v", ck.Interface(), cv.Type(), cv.Interface())
			} else {
				dbglog.Log("  SetMapIndex: %v -> [%v] %v", ck.Interface(), ptrToCopyValue.Type(), ptrToCopyValue.Interface())
				tgt.SetMapIndex(ck, ptrToCopyValue)
			}
		}()
	}

	dbglog.Log("  ptrToCopyValue.type: %v, eltypind: %v", tool.Typfmtv(&ptrToCopyValue), tool.Typfmt(eltypind))
	if err = c.copyTo(params, tgtval, ptrToCopyValue); err != nil {
		return
	}
	if err = c.copyTo(params, originalValue, ptrToCopyValue); err != nil {
		return
	}

	return //nolint:nakedret
}

func ensureMapPtrValue(tgt, key reflect.Value) (val reflect.Value, ptr bool) {
	val = tgt.MapIndex(key)
	if val.Kind() == reflect.Ptr {
		if tool.IsNil(val) {
			typ := tgt.Type().Elem()
			val = reflect.New(typ)
			vind := tool.Rindirect(val)
			tgt.SetMapIndex(key, val)
			ptr = true // val = vind
			dbglog.Log("    val.typ: %v, key.typ: %v | %v -> %v", tool.Typfmt(typ), tool.Typfmtv(&key), tool.Valfmt(&key), tool.Valfmt(&vind))
		}
	} else if !val.IsValid() {
		typ := tgt.Type().Elem()
		val = reflect.New(typ)
		vind := tool.Rindirect(val)
		dbglog.Log("    val.typ: %v, key.typ: %v | %v -> %v", tool.Typfmt(typ), tool.Typfmtv(&key), tool.Valfmt(&key), tool.Valfmt(&vind))
		tgt.SetMapIndex(key, vind)
		ptr = true // val = vind
	}
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

// copyFunc never used
// Deprecated always
func copyFunc(c *cpController, params *Params, from, to reflect.Value) (err error) { //nolint:unused,deadcode
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
		// to.SetPointer(from.Pointer())
		dbglog.Log("    copy chan not support: %v (%v) -> %v (%v)", from.Kind(), from.Type(), to.Kind(), to.Type())
		err = ErrCannotCopy.FormatWith(tool.Valfmt(&from), tool.Typfmtv(&from), tool.Valfmt(&to), tool.Typfmtv(&to))
	}
	// v := reflect.MakeChan(from.Type(), from.Cap())
	// ft := from.Type()
	// p := ptr(to, ft)
	// if from.Type().ConvertibleTo(to.Type()) {
	//	p.Set(v.Convert(to.Type()))
	// } else {
	//	p.Set(v)
	// }
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
	processed, err = copyPrimitiveToComposite(c, params, from, to, toind.Type())
	if processed || err != nil {
		return
	}

	// //////////////// primitive
	if !toind.IsValid() && to.Kind() == reflect.Ptr {
		tgt := reflect.New(targetType.Elem())
		toind = tool.Rindirect(tgt)
		defer func() {
			if err == nil {
				to.Set(tgt)
			}
		}()
	}

	// try primitive -> primitive at first
	if tool.CanConvert(&fromind, toind.Type()) {
		var val = fromind.Convert(toind.Type())
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

	err = ErrCannotConvertTo.FormatWith(fromind.Interface(), fromind.Kind(), toind.Interface(), toind.Kind())
	log.Errorf("    %v", err)
	return //nolint:nakedret
}

func copyPrimitiveToComposite(c *cpController, params *Params, from, to reflect.Value, desiredType reflect.Type) (processed bool, err error) {
	switch tk := desiredType.Kind(); tk { //nolint:exhaustive
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

	return //nolint:nakedret
}

func setTargetValue1(params *Params, to, toind, newval reflect.Value) (err error) {
	if err = setTargetValue2(params, toind, newval); err == nil {
		return
	}
	if to != toind {
		if err = setTargetValue2(params, to, newval); err == nil {
			return
		}
	}
	err = ErrUnknownState
	return
}

func setTargetValue2(params *Params, to, newval reflect.Value) (err error) {
	k := reflect.Invalid
	if params != nil && params.dstDecoded != nil {
		k = params.dstDecoded.Kind()
	}

	if k == reflect.Struct && params.accessor != nil && !tool.IsExported(params.accessor.StructField()) {
		if params.controller.copyUnexportedFields {
			cl.SetUnexportedField(to, newval)
		} // else ignore the unexported field
		return
	}
	if to.CanSet() {
		to.Set(newval)
		return
		// } else if newval.IsValid() {
		//	to.Set(newval)
		//	return
	}

	err = ErrUnknownState
	return
}
