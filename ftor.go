package deepcopy

import (
	"fmt"
	"github.com/hedzr/deepcopy/cl"
	"github.com/hedzr/deepcopy/dbglog"
	"github.com/hedzr/deepcopy/flags/cms"
	"github.com/hedzr/log"
	"gopkg.in/hedzr/errors.v3"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

func copyPointer(c *cpController, params *Params, from, to reflect.Value) (err error) {
	// from is a pointer

	src := rindirect(from)
	tgt := rindirect(to)

	newobj := func(c *cpController, params *Params, src, to, tgt reflect.Value) (err error) {
		newtyp := to.Type()
		if to.Type() == from.Type() {
			newtyp = newtyp.Elem() // is pointer and its same
		}
		// create new object and pointer
		toobjcopyptrv := reflect.New(newtyp)
		dbglog.Log("    toobjcopyptrv: %v", typfmtv(&toobjcopyptrv))
		if err = c.copyTo(params, src, toobjcopyptrv); err == nil {
			val := toobjcopyptrv
			if to.Type() == from.Type() {
				val = val.Elem()
			}
			err = setTargetValue1(params.owner, to, tgt, val)
			//to.Set(toobjcopyptrv)
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
			dbglog.Log("    pointer - zv: %v (%v), to: %v (%v)", valfmt(&zv), typfmt(newtyp), valfmt(&to), typfmtv(&to))
			to.Set(zv)
			// err = newobj(c, params, src, to, tgt)
		}
	} else {
		dbglog.Log("    pointer - tgt is invalid/cannot-be-set/ignored: src.valid: %v, %v (%v) -> tgt.valid: %v, %v (%v)",
			src.IsValid(), src.Type(), from.Kind(),
			tgt.IsValid(), to.Type(), to.Kind())
		err = newobj(c, paramsChild, src, to, tgt)
	}
	return
}

func copyInterface(c *cpController, params *Params, from, to reflect.Value) (err error) {
	if isNil(from) {
		if params.isGroupedFlagOKDeeply(cms.OmitIfNil, cms.OmitIfEmpty) {
			return
		}
		to.Set(reflect.Zero(to.Type()))
		return
	}
	if isZero(from) {
		if params.isGroupedFlagOKDeeply(cms.OmitIfZero, cms.OmitIfEmpty) {
			return
		}
		to.Set(reflect.Zero(to.Type()))
		return
	}

	paramsChild := newParams(withOwners(c, params, &from, &to, nil, nil))
	defer paramsChild.revoke()

	// unbox the interface{} to original data type
	toind, toptr := rdecode(to) // c.skip(to, reflect.Interface, reflect.Pointer)

	dbglog.Log("from.type: %v, decode to: %v", from.Type().Kind(), paramsChild.srcDecoded.Kind())
	dbglog.Log("  to.type: %v, decode to: %v (ptr: %v) | CanSet: %v, CanAddr: %v", to.Type().Kind(), toind.Kind(), toptr.Kind(), toind.CanSet(), toind.CanAddr())

	// var merging = c.flags.isAnyFlagsOK(SliceMerge, MapMerge) || params.isAnyFlagsOK(SliceMerge, MapMerge)
	if paramsChild.inMergeMode() || c.makeNewClone == false {
		err = c.copyTo(paramsChild, *paramsChild.srcDecoded, toptr)

	} else {
		copyValue := reflect.New(paramsChild.srcDecoded.Type()).Elem()
		if err = c.copyTo(paramsChild, *paramsChild.srcDecoded, copyValue); err == nil {
			to.Set(copyValue)
		}
	}
	return
}

func copyStruct(c *cpController, params *Params, from, to reflect.Value) (err error) {
	//if ff, ok := from.Interface().(time.Time); ok {
	//	to.Set(reflect.ValueOf(ff))
	//	return
	//}

	err = copyStructInternal(c, params, from, to,
		func(paramsChild *Params, ec errors.Error, i, amount int, padding string) {

			srcstructtable := paramsChild.targetIterator.(sourceStructFieldsTable)

			for i, amount = 0, len(srcstructtable.gettablerecords()); i < amount; i++ {
				sourcefield := srcstructtable.getcurrrecord()

				if c.isIgnoreName(sourcefield.ShortFieldName()) {
					srcstructtable.step(1) // skip this source field
					continue
				}

				if !paramsChild.nextTargetField() {
					continue
				}

				if paramsChild.parseFieldTags(sourcefield.structField.Tag) {
					continue
				}

				srcval, dstval := sourcefield.FieldValue(), paramsChild.accessor.FieldValue()
				log.VDebugf("%d. %s (%v) %v-> %s (%v) %v", i, sourcefield.FieldName(), valfmt(srcval), typfmtv(srcval), paramsChild.accessor.StructFieldName(), valfmt(dstval), typfmt(*paramsChild.accessor.FieldType()))

				if srcval != nil && dstval != nil && srcval.IsValid() {
					if err = invokeStructFieldTransformer(c, paramsChild, *srcval, *dstval, padding); err != nil {
						ec.Attach(err)
						log.Errorf("error: %v", err)
					}
					continue
				}

				if paramsChild.inMergeMode() {
					newtyp := paramsChild.accessor.FieldType()
					dbglog.Log("    new object for %v", typfmt(*newtyp))

					// create new object and pointer
					toobjcopyptrv := reflect.New(*newtyp)
					dbglog.Log("    toobjcopyptrv: %v", typfmtv(&toobjcopyptrv))

					if err = invokeStructFieldTransformer(c, paramsChild, *srcval, toobjcopyptrv, padding); err != nil {
						ec.Attach(err)
						log.Errorf("error: %v", err)
					} else {
						paramsChild.accessor.Set(toobjcopyptrv.Elem())
					}
					continue
				}

				dbglog.Log("   ignore nil/zero/invalid source or nil target")
			}

		})
	return
}

func copyStructInternal(
	c *cpController, params *Params,
	from, to reflect.Value,
	fn func(paramsChild *Params, ec errors.Error, i, amount int, padding string),
) (err error) {

	var (
		i, amount   int
		padding     string
		ec          = errors.New("copyStruct errors")
		paramsChild = newParams(withOwners(c, params, &from, &to, nil, nil))
	)

	defer ec.Defer(&err)
	defer paramsChild.revoke()

	defer func() {
		if e := recover(); e != nil {
			srcstructtable := paramsChild.targetIterator.(sourceStructFieldsTable)

			ff := srcstructtable.gettablerec(i).FieldValue()
			tf := paramsChild.accessor.FieldValue()
			tft := paramsChild.accessor.FieldType()

			err = errors.New("[recovered] copyStruct unsatisfied ([%v] -> [%v]), causes: %v",
				typfmtv(ff), typfmt(*tft), e).
				WithData(e).                      // collect e if it's an error object else store it simply
				WithTaggedData(errors.TaggedData{ // record the sites
					"source-field": ff,
					"target-field": tf,
					"source":       valfmt(ff),
					"target":       valfmt(tf),
				})
			ec.Attach(err)
			//n := log.CalcStackFrames(1)   // skip defer-recover frame at first
			//log.Skip(n).Errorf("%v", err) // skip golib frames and defer-recover frame, back to the point throwing panic
			if c.rethrow {
				log.Panicf("%+v", err)
			} else {
				log.Errorf("%+v", err)
			}
		}
	}()

	if dbglog.LogValid {
		dbgFrontOfStruct(params, paramsChild, padding, func(msg string, args ...interface{}) { dbglog.Log(msg, args...) })
	}

	fn(paramsChild, ec, i, amount, padding)

	////inspectStructV(to)
	//for i, amount = 0, f.NumField(); i < amount; i++ {
	//	fv := f.Field(i + params.srcOffset)
	//	if !fv.IsValid() {
	//		Log("%s  IGNORED invalid source", padding)
	//		continue
	//	}
	//
	//	err = transformField(c, params, from, to, f, t, i, padding)
	//}
	return
}

func dbgFrontOfStruct(params, paramsChild *Params, padding string, fn func(msg string, args ...interface{})) {
	if fn == nil {
		fn = dbglog.Log
	}
	padding = strings.Repeat("  ", params.depth()*2)
	fromT, toT := paramsChild.srcDecoded.Type(), paramsChild.dstDecoded.Type()
	//Log(" %s  %d, %d, %d", padding, params.index, params.srcOffset, params.dstOffset)
	fq := dbgMakeInfoString(fromT, params, true, fn)
	dq := dbgMakeInfoString(toT, params, false, fn)
	fn(" %s- (%v (%v)) -> dst (%v (%v))", padding, fromT, fromT.Kind(), toT, toT.Kind())
	fn(" %s  %s -> %s", padding, fq, dq)
}

func dbgMakeInfoString(typ reflect.Type, params *Params, src bool, fn func(msg string, args ...interface{})) (qstr string) {
	// var ft = typ // params.dstType
	if typ.Kind() == reflect.Struct && params != nil && params.accessor != nil && params.accessor.srcStructField != nil {
		//ofs := params.dstOffset
		//if src {
		//	ofs = params.srcOffset
		//}
		//idx := params.index + ofs
		//field := typ.Field(idx)
		//v := val.Field(idx)
		qstr = dbgMakeFieldInfoString(params.accessor.srcStructField, params, fn)
	} else {
		qstr = fmt.Sprintf("%v (%v)", typ, typ.Kind())
	}
	return
}

func dbgMakeFieldInfoString(fld *reflect.StructField, params *Params, fn func(msg string, args ...interface{})) (qstr string) {
	ft := fld.Type
	ftk := ft.Kind()
	if params != nil {
		qstr = fmt.Sprintf("Field%v %q (%v (%v)) | %v", fld.Index, fld.Name, ft, ftk, fld)
	} else {
		qstr = fmt.Sprintf("Field%v %q (%v (%v)) | %v", fld.Index, fld.Name, ft, ftk, fld)
	}
	return
}

//// transformField _
//// these codes are reserved here since its are the elder implements but have flaws.
//func transformField(c *cpController, params *Params, from, to, fromDecoded, toDecoded reflect.Value, i int, padding string) (err error) {
//	paramsChild := newParams(withOwners(c, params, &from, &to, &fromDecoded, &toDecoded, i))
//	defer paramsChild.revoke()
//
//	if functorLogValid {
//		Log(" %s  %2d, srcFieldType: %v, srcType: %v, ofs: %v, parent.ofs: %v", padding, i, paramsChild.srcFieldType, paramsChild.srcType, paramsChild.srcOffset, paramsChild.owner.srcOffset)
//		Log(" %s      dstFieldType: %v, dstType: %v, ofs: %v, parent.ofs: %v", padding, paramsChild.dstFieldType, paramsChild.dstType, paramsChild.dstOffset, paramsChild.owner.dstOffset)
//	}
//	if paramsChild.isFlagExists(Ignore) {
//		Log("%s  IGNORED [field tag settings]: %v original %q", padding, paramsChild.fieldTags, paramsChild.srcFieldType.Tag)
//		return
//	}
//
//	ff, df := params.ValueOfSource(), params.ValueOfDestination() // f.Field(params.index), t.Field(params.index)
//	if functorLogValid {
//		Log(" %s      src value: %v", padding, ff.Interface())
//		Log(" %s      dst value: %v", padding, df.Interface())
//	}
//
//	err = invokeStructFieldTransformer(c, paramsChild, ff, df, padding)
//	return
//}

func invokeStructFieldTransformer(c *cpController, params *Params, ff, df reflect.Value, padding string) (err error) {

	var processed bool
	if processed, err = tryConverters(c, params, ff, df); processed {
		return
	}

	fft, dft := ff.Type(), df.Type()
	fftk, dftk := fft.Kind(), dft.Kind()
	if fftk == reflect.Struct && ff.NumField() == 0 {
		// never get into here because tablerecords.getallfields skip empty struct
		log.Warnf("should never get into here, might be algor wrong ?")
	}
	if dftk == reflect.Struct && df.NumField() == 0 {
		// structIterable.Next() might return an empty struct accessor
		// rather than field.
		log.Errorf("shouldn't get into here because we have a failover branch at the callee")
	}

	err = c.copyTo(params, ff, df) // or, use internal standard implementation version

	return
}

func tryConverters(c *cpController, params *Params, ff, df reflect.Value) (processed bool, err error) {
	fft, dft := ff.Type(), df.Type()

	if cvt, ctx := c.valueCopiers.findCopiers(params, fft, dft); ctx != nil {
		dbglog.Log("-> using Copier %v", reflect.ValueOf(cvt).Type())
		err = cvt.CopyTo(ctx, ff, df) // use user-defined copy-n-merger to merge or copy source to destination
		processed = true

	} else if cvt, ctx := c.valueConverters.findConverters(params, fft, dft); ctx != nil {
		dbglog.Log("-> using Converter %v", reflect.ValueOf(cvt).Type())
		var result reflect.Value
		result, err = cvt.Transform(ctx, ff, dft) // use user-defined value converter to transform from source to destination
		df.Set(result)
		processed = true

	}

	return
}

// copySlice transforms from slice to target with slice or other types
func copySlice(c *cpController, params *Params, from, to reflect.Value) (err error) {
	if from.IsNil() { // an empty slice found
		//TODO omitempty, omitnil, omitzero, ...
		return
	}

	var tgt, tgtptr reflect.Value
	tgt, tgtptr = rdecode(to)
	if to != tgtptr {
		err = c.copyTo(params, from, tgtptr) // unwrap the pointer
		return
	}

	if params.controller != c {
		log.Panicf("[copySlice] c *cpController != params.controller, what's up??")
	}

	tk := tgt.Kind()
	if tk != reflect.Slice {
		var processed bool
		if processed, err = tryConverters(c, params, from, tgt); processed {
			return
		}
		log.Panicf("[copySlice] unsupported transforming: from slice -> %v,", typfmtv(&tgt))
	}

	ec := errors.New("slice copy/merge errors")
	defer ec.Defer(&err)

	for _, flag := range []cms.CopyMergeStrategy{cms.SliceMerge, cms.SliceCopyAppend, cms.SliceCopy} {
		if params.isGroupedFlagOKDeeply(flag) {

			//if !to.CanAddr() {
			//	if params != nil && !params.isStruct() {
			//		to = *params.dstOwner
			//		Log("use dstOwner to get a ptr to slice, new to.type: %v, canAddr: %v, canSet: %v", to.Type().Kind(), to.CanAddr(), to.CanSet())
			//	}
			//}

			// src, _ = c.decode(from)

			dbglog.Log("slice merge mode: %v", flag)
			dbglog.Log("from.type: %v", from.Type().Kind())
			dbglog.Log("  to.type: %v, canAddr: %v, canSet: %v", typfmtv(&to), to.CanAddr(), to.CanSet())
			//Log(" src.type: %v, len: %v, cap: %v, srcptr.canAddr: %v", src.Type().Kind(), src.Len(), src.Cap(), srcptr.CanAddr())
			dbglog.Log(" tgt.type: %v, tgtptr: %v .canAddr: %v", typfmtv(&tgt), typfmtv(&tgtptr), tgtptr.CanAddr())

			if fn, ok := getMapOfSliceOperations()[flag]; ok {
				if result, e := fn(c, params, from, tgt); e == nil {
					//tgt=ns
					//t := c.want2(to, reflect.Slice, reflect.Interface)
					//t.Set(tgt)
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

	return
}

type fnSliceOperator func(c *cpController, params *Params, src, tgt reflect.Value) (result reflect.Value, err error)
type mSliceOperations map[cms.CopyMergeStrategy]fnSliceOperator

func getMapOfSliceOperations() (mapOfSliceOperations mSliceOperations) {
	mapOfSliceOperations = mSliceOperations{
		cms.SliceCopy:       _sliceCopyOperation,
		cms.SliceCopyAppend: _sliceCopyAppendOperation,
		cms.SliceMerge:      _sliceMergeOperation,
	}
	return
}

// _sliceCopyOperation: for SliceCopy, target elements will be given up, and source copied to.
func _sliceCopyOperation(c *cpController, params *Params, src, tgt reflect.Value) (result reflect.Value, err error) {
	sl := src.Len()
	ns := reflect.MakeSlice(tgt.Type(), 0, 0)
	dbglog.Log("tgt slice: %v, el: %v", tgt.Type(), tgt.Type().Elem())

	ecTotal := errors.New("slice merge errors (%v -> %v)", src.Type(), tgt.Type())
	defer ecTotal.Defer(&err)

	for _, ss := range []struct {
		length int
		source reflect.Value
	}{
		// {tl, tgt},
		{sl, src},
	} {
		var tgtelemtype = tgt.Type().Elem()
		for i := 0; i < ss.length; i++ {
			var (
				el   = ss.source.Index(i)
				enew = el
				ec   = errors.New("cannot convert %v to %v", el.Type(), tgtelemtype)
			)
			if el.Type() != tgtelemtype {
				if cc, ctx := c.valueConverters.findConverters(params, el.Type(), tgtelemtype); cc != nil {
					if enew, err = cc.Transform(ctx, el, tgtelemtype); err != nil {
						ec.Attach(err)
						ecTotal.Attach(ec)
						continue // ignore invalid element
					}
				} else if canConvert(&el, tgtelemtype) {
					enew = el.Convert(tgtelemtype)
				}
			}

			if el.Type() == tgtelemtype {
				ns = reflect.Append(ns, el)
			} else {
				if canConvert(&el, tgtelemtype) {
					ns = reflect.Append(ns, enew)
				} else {
					ec := errors.New("cannot convert %v to %v", el.Type(), tgtelemtype)
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
		tgtelemtype := tgt.Type().Elem()
		for i := 0; i < ss.length; i++ {
			var (
				el   = ss.source.Index(i)
				enew = el
				ec   = errors.New("cannot convert %v to %v", el.Type(), tgtelemtype)
			)
			//elv := el.Interface()
			if el.Type() != tgtelemtype {
				if cc, ctx := c.valueConverters.findConverters(params, el.Type(), tgtelemtype); cc != nil {
					if enew, err = cc.Transform(ctx, el, tgtelemtype); err != nil {
						ec.Attach(err)
						ecTotal.Attach(ec)
						continue // ignore invalid element
					}
				} else if canConvert(&el, tgtelemtype) {
					enew = el.Convert(tgtelemtype)
					//elv = enew.Interface()
				}
			}

			if el.Type() == tgtelemtype {
				ns = reflect.Append(ns, el)
			} else {
				if canConvert(&el, tgtelemtype) {
					ns = reflect.Append(ns, enew)
				} else {
					ec := errors.New("cannot convert %v to %v", el.Type(), tgtelemtype)
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
	return
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
			//to.Set(reflect.Append(to, src.Index(i)))
			var (
				found bool
				cvtok bool
				el    = ss.source.Index(i)
				elv   = el.Interface()
				enew  = el
				ec    = errors.New("cannot convert %v to %v", el.Type(), tgtelemtype)
			)
			if el.Type() != tgtelemtype {
				if cc, ctx := c.valueConverters.findConverters(params, el.Type(), tgtelemtype); cc != nil {
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
				} else if canConvert(&el, tgtelemtype) {
					enew = el.Convert(tgtelemtype)
					cvtok, elv = true, enew.Interface()
				}
			}

			found = findInSlice(ns, elv, i)

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
	return
}

func copyArray(c *cpController, params *Params, from, to reflect.Value) (err error) {
	if isZero(from) {
		return
	}

	src := rindirect(from)
	tgt, tgtptr := rdecode(to)
	sl, tl := src.Len(), tgt.Len()

	//if !to.CanAddr() && params != nil {
	//	if !params.isStruct() {
	//		//to = *params.dstOwner
	//		Log("    !! use dstOwner to get a ptr to array, new to.type: %v, canAddr: %v, canSet: %v", to.Type().Kind(), to.CanAddr(), to.CanSet())
	//	}
	//}

	//if tgt.CanAddr() == false && tgtptr.CanAddr() {
	//	tgt = tgtptr
	//}

	//Log("    tgt.%v: %v", params.dstOwner.Type().Field(params.index).Name, params.dstOwner.Type().Field(params.index))
	dbglog.Log("    from.type: %v, len: %v, cap: %v", src.Type().Kind(), src.Len(), src.Cap())
	dbglog.Log("      to.type: %v, len: %v, cap: %v, tgtptr.canSet: %v, tgtptr.canaddr: %v", tgt.Type().Kind(), tgt.Len(), tgt.Cap(), tgtptr.CanSet(), tgtptr.CanAddr())

	eltyp := tgt.Type().Elem()
	//set := src.Index(0).Type()
	//if set != tgt.Index(0).Type() {
	//	return errors.New("cannot copy %v to %v", from.Interface(), to.Interface())
	//}

	cnt := minInt(sl, tl)
	for i := 0; i < cnt; i++ {
		se := src.Index(i)
		setyp := se.Type()
		dbglog.Log("src.el.typ: %v, tgt.el.typ: %v", typfmt(setyp), eltyp)
		if se.IsValid() {
			if setyp.AssignableTo(eltyp) {
				tgt.Index(i).Set(se)
			} else if setyp.ConvertibleTo(eltyp) {
				tgt.Index(i).Set(src.Convert(eltyp))
			}
		}
		//tgt.Index(i).Set(src.Index(i))
	}

	for i := cnt; i < tl; i++ {
		v := tgt.Index(i)
		if !v.IsValid() {
			tgt.Index(i).Set(reflect.Zero(eltyp))
			dbglog.Log("set [%v] to zero value", i)
		}
	}

	//to.Set(pt.Elem())

	dbglog.Log("    from: %v, to: %v", src.Interface(), tgt.Interface()) // pt.Interface())

	return
}

func copyMap(c *cpController, params *Params, from, to reflect.Value) (err error) {
	if from.IsNil() {
		return
	}

	var tgt, tgtptr reflect.Value
	tgt, tgtptr = rdecode(to)
	if to != tgtptr {
		err = c.copyTo(params, from, tgtptr) // unwrap the pointer
		return
	}

	tk := tgt.Kind()
	if tk != reflect.Map {
		dbglog.Log("from map -> %v", typfmtv(&tgt))
		//
	}

	ec := errors.New("map copy/merge errors")
	defer ec.Defer(&err)

	for _, flag := range []cms.CopyMergeStrategy{cms.MapMerge, cms.MapCopy} {
		if params.isGroupedFlagOKDeeply(flag) {
			if fn, ok := getMapOper()[flag]; ok {
				ec.Attach(fn(c, params, from, tgt))
			} else {
				ec.Attach(errors.New("unknown strategy for map: %v", flag))
			}
			break
		}
	}
	return
}

type fnMapOper func(c *cpController, params *Params, src, tgt reflect.Value) (err error)
type mapMapOper map[cms.CopyMergeStrategy]fnMapOper

func getMapOper() (mMapOper mapMapOper) {
	mMapOper = mapMapOper{
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
	return
}

func mergeOneKeyInMap(c *cpController, params *Params, src, tgt, key reflect.Value) (err error) {
	originalValue := src.MapIndex(key)

	// Log("  tgt.type: %v", typfmtv(&tgt))

	keyType := tgt.Type().Key()
	ptrToCopyKey := reflect.New(keyType)
	dbglog.Log("  ptrToCopyKey.type: %v", typfmtv(&ptrToCopyKey))
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
	} else {
		eltyp := tgt.Type().Elem() // get map value type
		eltypind, _ := rskiptype(eltyp, reflect.Ptr)

		var ptrToCopyValue, cv reflect.Value
		if eltypind.Kind() == reflect.Interface {
			tgtvalind, _ := rdecode(tgtval)
			dbglog.Log("  tgtval: [%v] %v, ind: %v", typfmtv(&tgtval), tgtval.Interface(), typfmtv(&tgtvalind))
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

		dbglog.Log("  ptrToCopyValue.type: %v, eltypind: %v", typfmtv(&ptrToCopyValue), typfmt(eltypind))
		if err = c.copyTo(params, tgtval, ptrToCopyValue); err != nil {
			return
		}
		if err = c.copyTo(params, originalValue, ptrToCopyValue); err != nil {
			return
		}

	}

	return
}

func ensureMapPtrValue(tgt, key reflect.Value) (val reflect.Value, ptr bool) {
	val = tgt.MapIndex(key)
	if val.Kind() == reflect.Ptr {
		if isNil(val) {
			typ := tgt.Type().Elem()
			val = reflect.New(typ)
			vind := rindirect(val)
			tgt.SetMapIndex(key, val)
			ptr = true // val = vind
			dbglog.Log("    val.typ: %v, key.typ: %v | %v -> %v", typfmt(typ), typfmtv(&key), valfmt(&key), valfmt(&vind))
		}
	} else if !val.IsValid() {
		typ := tgt.Type().Elem()
		val = reflect.New(typ)
		vind := rindirect(val)
		dbglog.Log("    val.typ: %v, key.typ: %v | %v -> %v", typfmt(typ), typfmtv(&key), valfmt(&key), valfmt(&vind))
		tgt.SetMapIndex(key, vind)
		ptr = true // val = vind
	}
	return
}

//

//

//

func copyUintptr(c *cpController, params *Params, from, to reflect.Value) (err error) {
	if to.CanSet() {
		to.Set(from)
	} else {
		//to.SetPointer(from.Pointer())
		dbglog.Log("    copy uintptr not support: %v -> %v", from.Kind(), to.Kind())
	}
	return
}

func copyUnsafePointer(c *cpController, params *Params, from, to reflect.Value) (err error) {
	if to.CanSet() {
		to.Set(from)
	} else {
		dbglog.Log("    copy unsafe pointer not support: %v -> %v", from.Kind(), to.Kind())
	}
	return
}

func copyFunc(c *cpController, params *Params, from, to reflect.Value) (err error) {
	if to.CanSet() {
		to.Set(from)
		return
	}

	toind := rindirect(to)
	if k := toind.Kind(); k != reflect.Func && c.copyFunctionResultToTarget {
		// from.
		return

	} else if k == reflect.Func {

		if !params.processUnexportedField(to, from) {
			ptr := from.Pointer()
			to.SetPointer(unsafe.Pointer(ptr))
		}
		dbglog.Log("    function pointer copied: %v (%v) -> %v", from.Kind(), from.Interface(), to.Kind())
	}

	return
}

func copyChan(c *cpController, params *Params, from, to reflect.Value) (err error) {
	if to.CanSet() {
		dbglog.Log("    copy chan: %v (%v) -> %v (%v)", from.Kind(), from.Type(), to.Kind(), to.Type())
		to.Set(from)
		//Log("        after: %v -> %v", from.Interface(), to.Interface())
	} else {
		//to.SetPointer(from.Pointer())
		dbglog.Log("    copy chan not support: %v (%v) -> %v (%v)", from.Kind(), from.Type(), to.Kind(), to.Type())
	}
	// v := reflect.MakeChan(from.Type(), from.Cap())
	//ft := from.Type()
	//p := ptr(to, ft)
	//if from.Type().ConvertibleTo(to.Type()) {
	//	p.Set(v.Convert(to.Type()))
	//} else {
	//	p.Set(v)
	//}
	return
}

//

func copy1(c *cpController, params *Params, from, to reflect.Value) (err error) {
	return
}

func copy2(c *cpController, params *Params, from, to reflect.Value) (err error) {
	return
}

//func copyUnexportedStructFields(c *cpController, params *Params, from, to reflect.Value) {
//	if from.Kind() != reflect.Struct || to.Kind() != reflect.Struct || !from.Type().AssignableTo(to.Type()) {
//		return
//	}
//
//	// create a shallow copy of 'to' to get all fields
//	tmp := indirect(reflect.New(to.Type()))
//	tmp.Set(from)
//
//	// revert exported fields
//	for i := 0; i < to.NumField(); i++ {
//		if tmp.Field(i).CanSet() {
//			tmp.Field(i).Set(to.Field(i))
//		}
//	}
//	to.Set(tmp)
//}

func copyDefaultHandler(c *cpController, params *Params, from, to reflect.Value) (err error) {
	if c != nil {
		sourceType, targetType := from.Type(), to.Type()
		if cvt, ctx := c.valueCopiers.findCopiers(params, sourceType, targetType); cvt != nil {
			err = cvt.CopyTo(ctx, from, to)
			return
		}
	}

	sourceType, targetType := from.Type(), to.Type()
	fromind, toind := rdecodesimple(from), rdecodesimple(to)
	dbglog.Log("  copyDefaultHandler: %v -> %v | %v", typfmtv(&fromind), typfmtv(&toind), typfmtv(&to))

	////////////////// source is primitive types but target isn't its
	var processed bool
	processed, err = copyPrimitiveToComposite(c, params, from, to, toind.Type())
	if processed || err != nil {
		return
	}

	////////////////// primitive
	if !toind.IsValid() && to.Kind() == reflect.Ptr {
		tgt := reflect.New(targetType.Elem())
		toind = rindirect(tgt)
		defer func() {
			if err == nil {
				to.Set(tgt)
			}
		}()
	}

	// try primitive -> primitive at first
	if canConvert(&fromind, toind.Type()) {
		var val = fromind.Convert(toind.Type())
		err = setTargetValue1(params, to, toind, val)
		return
	}
	if canConvert(&from, to.Type()) && to.CanSet() {
		var val = from.Convert(to.Type())
		err = setTargetValue1(params, to, toind, val)
		return
	}
	if sourceType.AssignableTo(targetType) {
		if toind.CanSet() {
			toind.Set(fromind)
		} else if to.CanSet() {
			to.Set(fromind)
		} else {
			err = ErrCannotSet.FormatWith(fromind, typfmtv(&fromind), toind, typfmtv(&toind))
		}
		return
	}

	err = ErrCannotConvertTo.FormatWith(fromind.Interface(), fromind.Kind(), toind.Interface(), toind.Kind())
	log.Errorf("    %v", err)
	return
}

func copyPrimitiveToComposite(c *cpController, params *Params, from, to reflect.Value, desiredType reflect.Type) (processed bool, err error) {

	switch tk := desiredType.Kind(); tk {
	case reflect.Slice:
		dbglog.Log("  copyPrimitiveToComposite: %v -> %v | %v", typfmtv(&from), typfmt(desiredType), typfmtv(&to))

		eltyp := desiredType.Elem()
		elnew := reflect.New(eltyp)
		if err = copyDefaultHandler(c, params, from, elnew); err != nil {
			return
		}

		elnewelem := elnew.Elem()
		dbglog.Log("    source converted: %v (%v)", valfmt(&elnewelem), typfmtv(&elnewelem))

		slice := reflect.MakeSlice(reflect.SliceOf(eltyp), 1, 1)
		slice.Index(0).Set(elnewelem)
		dbglog.Log("    source converted: %v (%v)", valfmt(&slice), typfmtv(&slice))

		err = copySlice(c, params, slice, to)
		processed = true

	case reflect.Map:
		// not support

	case reflect.Struct:
		// not support

	case reflect.Func:
		tgt := rdecodesimple(to)
		processed, err = true, copyToFuncImpl(c, from, tgt, tgt.Type())

	}

	return
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

	if k == reflect.Struct && params.accessor != nil && !isExported(params.accessor.StructField()) {
		if params.controller.copyUnexportedFields {
			cl.SetUnexportedField(to, newval)
		} //else ignore the unexported field
		return
	}
	if to.CanSet() {
		to.Set(newval)
		return
		//} else if newval.IsValid() {
		//	to.Set(newval)
		//	return
	}

	err = ErrUnknownState
	return
}
