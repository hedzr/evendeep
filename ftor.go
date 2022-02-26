package deepcopy

import (
	"fmt"
	"github.com/hedzr/log"
	"gopkg.in/hedzr/errors.v3"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func copyPointer(c *cpController, params *Params, from, to reflect.Value) (err error) {
	// from is a pointer

	src := c.indirect(from)
	tgt := c.indirect(to)

	if tgt.CanSet() {
		paramsChild := newParams(withOwners(params, &from, &to, nil, nil, 0))
		defer paramsChild.revoke()
		err = c.copyTo(paramsChild, src, to)
	} else {
		functorLog("    pointer - tgt is invalid/cannot-be-set/ignored: src.valid: %v, %v (%v) -> tgt.valid: %v, %v (%v)",
			src.IsValid(), src.Type(), from.Kind(),
			tgt.IsValid(), to.Type(), to.Kind())

		newtyp := to.Type()
		if to.Type() == from.Type() {
			newtyp = newtyp.Elem() // is pointer and its same
		}

		// create new object and pointer
		toobjcopyptrv := reflect.New(newtyp)
		functorLog("    toobjcopyptrv: %v", typfmtv(&toobjcopyptrv))
		if err = c.copyTo(params, src, toobjcopyptrv); err == nil {
			to.Set(toobjcopyptrv)
		}
	}
	return
}

func copyUintptr(c *cpController, params *Params, from, to reflect.Value) (err error) {
	if to.CanSet() {
		to.Set(from)
	} else {
		//to.SetPointer(from.Pointer())
		functorLog("    copy uintptr not support: %v -> %v", from.Kind(), to.Kind())
	}
	return
}

func copyUnsafePointer(c *cpController, params *Params, from, to reflect.Value) (err error) {
	if to.CanSet() {
		to.Set(from)
	} else {
		functorLog("    copy unsafe pointer not support: %v -> %v", from.Kind(), to.Kind())
	}
	return

	////src := c.indirect(from)
	////rv := reflect.NewAt(src.Type(), unsafe.Pointer(from.UnsafeAddr())).Elem()
	////to.SetPointer(unsafe.Pointer(rv.UnsafeAddr()))
	//if to.Kind() == reflect.UnsafePointer {
	//	ptrTo := ptrOf(to)
	//	ptrFrom := ptrOf(from)
	//	ptrTo.SetPointer(unsafe.Pointer(ptrFrom.UnsafeAddr()))
	//	return
	//}
	//return errors.New("target type is %v, want reflect.UnsafePointer", to.Kind())
}

func copyFunc(c *cpController, params *Params, from, to reflect.Value) (err error) {
	if to.CanSet() {
		to.Set(from)
	} else {
		//to.SetPointer(from.Pointer())
		functorLog("    todo copy function: %v -> %v", from.Kind(), to.Kind())
	}
	return
}

func copyChan(c *cpController, params *Params, from, to reflect.Value) (err error) {
	if to.CanSet() {
		functorLog("    copy chan: %v (%v) -> %v (%v)", from.Kind(), from.Type(), to.Kind(), to.Type())
		to.Set(from)
		//functorLog("        after: %v -> %v", from.Interface(), to.Interface())
	} else {
		//to.SetPointer(from.Pointer())
		functorLog("    copy chan not support: %v (%v) -> %v (%v)", from.Kind(), from.Type(), to.Kind(), to.Type())
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

func copyInterface(c *cpController, params *Params, from, to reflect.Value) (err error) {
	if from.IsNil() {
		return // TODO omitempty,...
	}

	paramsChild := newParams(withOwners(params, &from, &to, nil, nil, 0))
	defer paramsChild.revoke()

	// unbox the interface{} to original data type
	toind, toptr := c.decode(to) // c.skip(to, reflect.Interface, reflect.Pointer)

	functorLog("from.type: %v, decode to: %v", from.Type().Kind(), paramsChild.srcDecoded.Kind())
	functorLog("  to.type: %v, decode to: %v (ptr: %v) | CanSet: %v, CanAddr: %v", to.Type().Kind(), toind.Kind(), toptr.Kind(), toind.CanSet(), toind.CanAddr())

	var merging = c.flags.isAnyFlagsOK(SliceMerge, MapMerge) || params.isAnyFlagsOK(SliceMerge, MapMerge)
	if merging || c.makeNewClone == false {
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
	if tt, ok := from.Interface().(time.Time); ok {
		to.Set(reflect.ValueOf(tt))
		return
	}

	var (
		i, amount      int
		ok             bool
		padding        string
		f, _           = rdecode(from)
		t, _           = rdecode(to)
		merging        = c.flags.isAnyFlagsOK(SliceMerge, MapMerge) || params.isAnyFlagsOK(SliceMerge, MapMerge)
		targetIterator = newStructIterator(t, withStructPtrAutoExpand(c.autoExpandStuct), withStructFieldPtrAutoNew(true))
		accessor       *fieldaccessor
		sourcefields   fieldstable
		ec             = errors.New("copyStruct errors")
	)

	sourcefields = sourcefields.getallfields(f, c.autoExpandStuct)

	if functorLogValid {
		padding = strings.Repeat("  ", params.depth()*2)
		fromT, toT := f.Type(), t.Type()
		//functorLog(" %s  %d, %d, %d", padding, params.index, params.srcOffset, params.dstOffset)
		fq := dbgMakeInfoString(fromT, params, true)
		dq := dbgMakeInfoString(toT, params, false)
		functorLog(" %s- (%v (%v)) -> dst (%v (%v))", padding, fromT, fromT.Kind(), toT, toT.Kind())
		functorLog(" %s  %s -> %s", padding, fq, dq)
	}

	defer func() {
		if e := recover(); e != nil {
			ff := sourcefields.tablerecords[i].structFieldValue
			tf := accessor.FieldValue()
			tft := accessor.FieldType()

			err = errors.New("[recovered] copyStruct unsatisfied ([%v] -> [%v]), causes: %v",
				typfmtv(ff), typfmt(*tft), e).
				WithData(e).                      // collect e if it's an error object else store it simply
				WithTaggedData(errors.TaggedData{ // record the sites
					"source-field": ff,
					"target-field": tf,
					"source":       valfmt(ff),
					"target":       valfmt(tf),
				})
			//n := log.CalcStackFrames(1)   // skip defer-recover frame at first
			//log.Skip(n).Errorf("%v", err) // skip golib frames and defer-recover frame, back to the point throwing panic
			log.Errorf("%+v", err)
		}
	}()

	for i, amount = 0, len(sourcefields.tablerecords); i < amount; i++ {
		sourcefield := sourcefields.tablerecords[i]
		accessor, ok = targetIterator.Next()
		flags := parseFieldTags(sourcefield.structField.Tag) // todo pass and apply the flags in field tag
		if flags.isFlagExists(Ignore) || !ok {
			continue
		}

		srcval, dstval := sourcefield.Value(), accessor.FieldValue()
		functorLog("%d. %s (%v) %v-> %s (%v) %v", i, sourcefield.FieldName(), valfmt(srcval), typfmtv(srcval), accessor.StructFieldName(), valfmt(dstval), typfmt(*accessor.FieldType()))

		if srcval != nil && dstval != nil && srcval.IsValid() {
			ec.Attach(invokeStructFieldTransformer(c, params, *srcval, *dstval, padding))

		} else if merging {

			newtyp := accessor.FieldType()
			functorLog("    new object for %v", typfmt(*newtyp))

			// create new object and pointer
			toobjcopyptrv := reflect.New(*newtyp)
			functorLog("    toobjcopyptrv: %v", typfmtv(&toobjcopyptrv))

			if err = invokeStructFieldTransformer(c, params, *srcval, toobjcopyptrv, padding); err != nil {
				ec.Attach(err)
			} else {
				accessor.Set(toobjcopyptrv.Elem())
			}

		} else {
			functorLog("   ignore nil/zero/invalid source or nil target")
		}
	}

	////inspectStructV(to)
	//for i, amount = 0, f.NumField(); i < amount; i++ {
	//	fv := f.Field(i + params.srcOffset)
	//	if !fv.IsValid() {
	//		functorLog("%s  IGNORED invalid source", padding)
	//		continue
	//	}
	//
	//	err = transformField(c, params, from, to, f, t, i, padding)
	//}
	return
}

func dbgMakeInfoString(typ reflect.Type, params *Params, src bool) (qstr string) {
	// var ft = typ // params.dstType
	if typ.Kind() == reflect.Struct && params != nil {
		ofs := params.dstOffset
		if src {
			ofs = params.srcOffset
		}
		idx := params.index + ofs
		field := typ.Field(idx)
		//v := val.Field(idx)
		qstr = dbgMakeFieldInfoString(field, ofs, params)
	} else {
		qstr = fmt.Sprintf("%v (%v)", typ, typ.Kind())
	}
	return
}

func dbgMakeFieldInfoString(fld reflect.StructField, ofs int, params *Params) (qstr string) {
	ft := fld.Type
	ftk := ft.Kind()
	if params != nil {
		qstr = fmt.Sprintf("Field[%v(+%v)] %q (%v (%v)) | %v", params.index, ofs, fld.Name, ft, ftk, fld)
	} else {
		qstr = fmt.Sprintf("Field%v(+%v) %q (%v (%v)) | %v", fld.Index, ofs, fld.Name, ft, ftk, fld)
	}
	return
}

// transformField _
// these codes are reserved here since its are the elder implements but have flaws.
func transformField(c *cpController, params *Params, from, to, fromDecoded, toDecoded reflect.Value, i int, padding string) (err error) {
	paramsChild := newParams(withOwners(params, &from, &to, &fromDecoded, &toDecoded, i))
	defer paramsChild.revoke()

	if functorLogValid {
		functorLog(" %s  %2d, srcFieldType: %v, srcType: %v, ofs: %v, parent.ofs: %v", padding, i, paramsChild.srcFieldType, paramsChild.srcType, paramsChild.srcOffset, paramsChild.owner.srcOffset)
		functorLog(" %s      dstFieldType: %v, dstType: %v, ofs: %v, parent.ofs: %v", padding, paramsChild.dstFieldType, paramsChild.dstType, paramsChild.dstOffset, paramsChild.owner.dstOffset)
	}
	if paramsChild.isFlagExists(Ignore) {
		functorLog("%s  IGNORED [field tag settings]: %v original %q", padding, paramsChild.fieldTags, paramsChild.srcFieldType.Tag)
		return
	}

	ff, df := params.ValueOfSource(), params.ValueOfDestination() // f.Field(params.index), t.Field(params.index)
	if functorLogValid {
		functorLog(" %s      src value: %v", padding, ff.Interface())
		functorLog(" %s      dst value: %v", padding, df.Interface())
	}

	err = invokeStructFieldTransformer(c, paramsChild, ff, df, padding)
	return
}

func invokeStructFieldTransformer(c *cpController, params *Params, ff, df reflect.Value, padding string) (err error) {
	fft, dft := ff.Type(), df.Type()
	if cvt, ctx := c.findCopiers(params, fft, dft); ctx != nil {
		err = cvt.CopyTo(ctx, ff, df) // use user-defined copy-n-merger to merge or copy source to destination
	} else if cvt, ctx := c.findConverters(params, fft, dft); ctx != nil {
		var result reflect.Value
		result, err = cvt.Transform(ctx, ff, dft) // use user-defined value converter to transform from source to destination
		df.Set(result)
	} else {

		fftk, dftk := fft.Kind(), dft.Kind()
		if fftk == reflect.Struct && ff.NumField() == 0 {
			// never get into here because tablerecords.getallfields skip empty struct
		}
		if dftk == reflect.Struct && df.NumField() == 0 {
			// structIterable.Next() might return an empty struct accessor
			// rather than field.
			log.Errorf("shouldn't get into here because we have a failover branch at the callee")
		}

		err = c.copyTo(params, ff, df) // or, use internal standard implementation version
	}
	return
}

func copySlice(c *cpController, params *Params, from, to reflect.Value) (err error) {
	if from.IsNil() { // an empty slice found
		//TODO omitempty, omitnil, omitzero, ...
		return
	}

	var tgt, tgtptr reflect.Value
	tgt, tgtptr = c.decode(to)
	if to != tgtptr {
		err = c.copyTo(params, from, tgtptr) // unwrap the pointer
		return
	}

	tk := tgt.Kind()
	if tk != reflect.Slice {
		functorLog("from slice -> %v", typfmtv(&tgt))
		// todo from slice to other types
	}

	ec := errors.New("slice copy/merge errors")
	defer ec.Defer(&err)

	for _, flag := range []CopyMergeStrategy{SliceMerge, SliceCopyAppend, SliceCopy} {
		if c.flags.isGroupedFlagOK(flag) || params.isGroupedFlagOKDeeply(flag) {

			if !to.CanAddr() {
				if params != nil && !params.isStruct() {
					to = *params.dstOwner
					functorLog("use dstOwner to get a ptr to slice, new to.type: %v, canAddr: %v, canSet: %v", to.Type().Kind(), to.CanAddr(), to.CanSet())
				}
			}

			// src, _ = c.decode(from)

			functorLog("slice merge mode: %v", flag)
			functorLog("from.type: %v", from.Type().Kind())
			functorLog("  to.type: %v, canAddr: %v, canSet: %v", typfmtv(&to), to.CanAddr(), to.CanSet())
			//functorLog(" src.type: %v, len: %v, cap: %v, srcptr.canAddr: %v", src.Type().Kind(), src.Len(), src.Cap(), srcptr.CanAddr())
			functorLog(" tgt.type: %v, tgtptr: %v .canAddr: %v", typfmtv(&tgt), typfmtv(&tgtptr), tgtptr.CanAddr())

			if fn, ok := mSliceOper[flag]; ok {
				if result, e := fn(c, params, from, tgt); e == nil {
					//tgt=ns
					//t := c.want2(to, reflect.Slice, reflect.Interface)
					//t.Set(tgt)
					//   //tgtptr.Elem().Set(result)
					if tgtptr.Kind() == reflect.Slice {
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

	//if c.flags.isGroupedFlagOK(SliceMerge) || params.isGroupedFlagOK(SliceMerge) {
	//
	//	if !to.CanAddr() {
	//		if params != nil && !params.isStruct() {
	//			to = *params.dstOwner
	//			functorLog("use dstOwner to get a ptr to slice, new to.type: %v, canAddr: %v, canSet: %v", to.Type().Kind(), to.CanAddr(), to.CanSet())
	//		}
	//	}
	//
	//	src, srcptr = c.decode(from)
	//	tgt, tgtptr = c.decode(to)
	//
	//	functorLog("slice merge mode")
	//	functorLog("from.type: %v", from.Type().Kind())
	//	functorLog("  to.type: %v, canAddr: %v, canSet: %v", to.Type().Kind(), to.CanAddr(), to.CanSet())
	//	functorLog(" src.type: %v, len: %v, cap: %v, srcptr.canAddr: %v", src.Type().Kind(), src.Len(), src.Cap(), srcptr.CanAddr())
	//	functorLog(" tgt.type: %v, len: %v, cap: %v, tgtptr.canAddr: %v", tgt.Type().Kind(), tgt.Len(), tgt.Cap(), tgtptr.CanAddr())
	//
	//	//paramsChild := newParams(withOwners(&from, &to, -1))
	//	//params.addChildField(paramsChild)
	//
	//	tl, sl := tgt.Len(), src.Len()
	//	ns := reflect.MakeSlice(tgt.Type(), 0, 0)
	//	for _, ss := range []struct {
	//		length int
	//		source reflect.Value
	//	}{
	//		{tl, tgt},
	//		{sl, src},
	//	} {
	//		for i := 0; i < ss.length; i++ {
	//			//to.Set(reflect.Append(to, src.Index(i)))
	//			found, el := false, ss.source.Index(i)
	//			elv := el.Interface()
	//			for j := 0; j < ns.Len(); j++ {
	//				tev := ns.Index(j).Interface()
	//				functorLog("  testing tgt[%v](%v) and src[%v](%v)", j, tev, i, elv)
	//				if reflect.DeepEqual(tev, elv) {
	//					found = true
	//					functorLog("found exists el at tgt[%v], for src[%v], value is %v", j, i, elv)
	//					break
	//				}
	//			}
	//			if !found {
	//				ns = reflect.Append(ns, el)
	//			}
	//		}
	//	}
	//
	//	//tgt=ns
	//	//t := c.want2(to, reflect.Slice, reflect.Interface)
	//	//t.Set(tgt)
	//	tgtptr.Elem().Set(ns)
	//
	//} else if c.flags.isFlagExists(SliceCopyAppend) || (params != nil && params.isFlagExists(SliceCopyAppend)) {
	//
	//	// ftfCopyEnh
	//
	//	src, srcptr = c.decode(from)
	//	tgt, tgtptr = c.decode(to)
	//
	//	functorLog("slice copy enh (overwrite) mode")
	//	functorLog("from.type: %v", from.Type().Kind())
	//	functorLog("  to.type: %v, canAddr: %v, canSet: %v", to.Type().Kind(), to.CanAddr(), to.CanSet())
	//	functorLog(" src.type: %v, len: %v, cap: %v, srcptr.canAddr: %v", src.Type().Kind(), src.Len(), src.Cap(), srcptr.CanAddr())
	//	functorLog(" tgt.type: %v, len: %v, cap: %v, tgtptr.canAddr: %v", tgt.Type().Kind(), tgt.Len(), tgt.Cap(), tgtptr.CanAddr())
	//
	//	//for i := 0; i < src.Len(); i++ {
	//	//	si := src.Index(i)
	//	//	var found bool
	//	//	for j := 0; j < to.Elem().Len(); j++ {
	//	//		ti := to.Elem().Index(j)
	//	//		if found = reflect.DeepEqual(si, ti); found {
	//	//			break
	//	//		}
	//	//	}
	//	//	if !found {
	//	//		to.Elem().Set(reflect.Append(to.Elem(), si))
	//	//	}
	//	//}
	//
	//	tl, sl := tgt.Len(), src.Len()
	//	ns := reflect.MakeSlice(tgt.Type(), 0, 0)
	//	for _, ss := range []struct {
	//		length int
	//		source reflect.Value
	//	}{
	//		{tl, tgt},
	//		{sl, src},
	//	} {
	//		for i := 0; i < ss.length; i++ {
	//			el := ss.source.Index(i)
	//			ns = reflect.Append(ns, el)
	//		}
	//	}
	//
	//	// t := c.want(to, reflect.Slice)
	//	//t := c.want2(to, reflect.Slice, reflect.Interface)
	//	//t.Set(tgt)
	//	if tgtptr.Kind() == reflect.Slice {
	//		tgtptr.Set(ns)
	//	} else {
	//		tgtptr.Elem().Set(ns)
	//	}
	//
	//	functorLog("    slice result: %v", tgtptr.Interface())
	//	return
	//
	//} else {
	//
	//	// copy and set each source element to target slice
	//
	//	src, srcptr = c.decode(from)
	//	tgt, tgtptr = c.decode(to)
	//
	//	functorLog("slice copy mode")
	//	functorLog("     from: %v", from.Interface())
	//	functorLog("       to: %v", to.Interface())
	//	functorLog("from.type: %v", from.Type().Kind())
	//	functorLog("  to.type: %v, canAddr: %v, canSet: %v", to.Type().Kind(), to.CanAddr(), to.CanSet())
	//	functorLog(" src.type: %v, len: %v, cap: %v, srcptr.canAddr: %v", src.Type().Kind(), src.Len(), src.Cap(), srcptr.CanAddr())
	//	functorLog(" tgt.type: %v, len: %v, cap: %v, tgtptr.canSet: %v", tgt.Type().Kind(), tgt.Len(), tgt.Cap(), tgtptr.CanSet())
	//
	//	//for i := 0; i < src.Len(); i++ {
	//	//	tgt = reflect.Append(tgt, src.Index(i))
	//	//	functorLog("    %d, append: %v, tgt: %v", i, src.Index(i).Interface(), tgt.Interface())
	//	//}
	//
	//	sl := src.Len()
	//	ns := reflect.MakeSlice(tgt.Type(), 0, 0)
	//	for _, ss := range []struct {
	//		length int
	//		source reflect.Value
	//	}{
	//		// {tl, tgt},
	//		{sl, src},
	//	} {
	//		for i := 0; i < ss.length; i++ {
	//			el := ss.source.Index(i)
	//			ns = reflect.Append(ns, el)
	//		}
	//	}
	//
	//	// t := c.want(to, reflect.Slice)
	//	//t := c.want2(to, reflect.Slice, reflect.Interface)
	//	//t.Set(tgt)
	//	if tgtptr.Kind() == reflect.Slice {
	//		tgtptr.Set(ns)
	//	} else {
	//		tgtptr.Elem().Set(ns)
	//	}
	//
	//	functorLog("    slice result: %v", tgtptr.Interface())
	//}

	return
}

type fnSliceOper func(c *cpController, params *Params, src, tgt reflect.Value) (result reflect.Value, err error)
type mapSliceOper map[CopyMergeStrategy]fnSliceOper

var mSliceOper = mapSliceOper{
	// SliceCopy: target elements will be given up, and source copied to.
	SliceCopy: func(c *cpController, params *Params, src, tgt reflect.Value) (result reflect.Value, err error) {
		sl := src.Len()
		ns := reflect.MakeSlice(tgt.Type(), 0, 0)
		functorLog("tgt slice: %v, el: %v", tgt.Type(), tgt.Type().Elem())

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
					if cc, ctx := c.findConverters(params, el.Type(), tgtelemtype); cc != nil {
						if enew, err = cc.Transform(ctx, el, tgtelemtype); err != nil {
							ec.Attach(err)
							ecTotal.Attach(ec)
							continue // ignore invalid element
						}
					} else if el.CanConvert(tgtelemtype) {
						enew = el.Convert(tgtelemtype)
					}
				}

				if el.Type() == tgtelemtype {
					ns = reflect.Append(ns, el)
				} else {
					if el.CanConvert(tgtelemtype) {
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
	},
	// SliceCopyAppend: target and source elements will be copied to new target.
	// The duplicated elements were kept.
	SliceCopyAppend: func(c *cpController, params *Params, src, tgt reflect.Value) (result reflect.Value, err error) {
		sl, tl := src.Len(), tgt.Len()
		ns := reflect.MakeSlice(tgt.Type(), 0, 0)
		functorLog("tgt slice: %v, el: %v", tgt.Type(), tgt.Type().Elem())

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
					if cc, ctx := c.findConverters(params, el.Type(), tgtelemtype); cc != nil {
						if enew, err = cc.Transform(ctx, el, tgtelemtype); err != nil {
							ec.Attach(err)
							ecTotal.Attach(ec)
							continue // ignore invalid element
						}
					} else if el.CanConvert(tgtelemtype) {
						enew = el.Convert(tgtelemtype)
						//elv = enew.Interface()
					}
				}

				if el.Type() == tgtelemtype {
					ns = reflect.Append(ns, el)
				} else {
					if el.CanConvert(tgtelemtype) {
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
	},
	// SliceMerge: target and source elements will be copied to new target
	// with uniqueness.
	SliceMerge: func(c *cpController, params *Params, src, tgt reflect.Value) (result reflect.Value, err error) {
		sl, tl := src.Len(), tgt.Len()
		ns := reflect.MakeSlice(tgt.Type(), 0, 0)
		tgtelemtype := tgt.Type().Elem()
		functorLog("tgt slice: %v, el: %v", tgt.Type(), tgt.Type().Elem())

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
					if cc, ctx := c.findConverters(params, el.Type(), tgtelemtype); cc != nil {
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
					} else if el.CanConvert(tgtelemtype) {
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
	},
}

func copyArray(c *cpController, params *Params, from, to reflect.Value) (err error) {
	if from.IsZero() {
		return
	}

	src := c.indirect(from)
	tgt, tgtptr := c.decode(to)
	sl, tl := src.Len(), tgt.Len()

	if !to.CanAddr() && params != nil {
		if !params.isStruct() {
			to = *params.dstOwner
			functorLog("use dstOwner to get a ptr to array, new to.type: %v, canAddr: %v, canSet: %v", to.Type().Kind(), to.CanAddr(), to.CanSet())
		}
	}

	//functorLog("    tgt.%v: %v", params.dstOwner.Type().Field(params.index).Name, params.dstOwner.Type().Field(params.index))
	functorLog("    from.type: %v, len: %v, cap: %v", src.Type().Kind(), src.Len(), src.Cap())
	functorLog("      to.type: %v, len: %v, cap: %v, tgtptr.canSet: %v", tgt.Type().Kind(), tgt.Len(), tgt.Cap(), tgtptr.CanSet())

	set := src.Index(0).Type()
	if set != tgt.Index(0).Type() {
		return errors.New("cannot copy %v to %v", from.Interface(), to.Interface())
	}

	cnt := sl
	if sl > tl {
		cnt = tl
	}

	for i := 0; i < cnt; i++ {
		tgt.Index(i).Set(src.Index(i))
	}

	for i := cnt; i < tl; i++ {
		v := tgt.Index(i)
		if !v.IsValid() {
			tgt.Index(i).Set(reflect.Zero(v.Type()))
			functorLog("set [%v] to zero value", i)
		}
	}

	//to.Set(pt.Elem())

	functorLog("    from: %v, to: %v", src.Interface(), tgt.Interface()) // pt.Interface())

	return
}

func copyMap(c *cpController, params *Params, from, to reflect.Value) (err error) {
	if from.IsNil() {
		return
	}

	var tgt, tgtptr reflect.Value
	tgt, tgtptr = c.decode(to)
	if to != tgtptr {
		err = c.copyTo(params, from, tgtptr) // unwrap the pointer
		return
	}

	tk := tgt.Kind()
	if tk != reflect.Map {
		functorLog("from map -> %v", typfmtv(&tgt))
		//
	}

	ec := errors.New("map copy/merge errors")
	defer ec.Defer(&err)

	for _, flag := range []CopyMergeStrategy{MapMerge, MapCopy} {
		if c.flags.isGroupedFlagOK(flag) || params.isGroupedFlagOKDeeply(flag) {
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
type mapMapOper map[CopyMergeStrategy]fnMapOper

func getMapOper() (mMapOper mapMapOper) {
	mMapOper = mapMapOper{
		MapCopy: func(c *cpController, params *Params, src, tgt reflect.Value) (err error) {

			tgt.Set(reflect.MakeMap(src.Type()))

			ec := errors.New("map merge errors")
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
		MapMerge: func(c *cpController, params *Params, src, tgt reflect.Value) (err error) {
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

func ensureMapPtrValue(tgt, key reflect.Value) (val reflect.Value, ptr bool) {
	val = tgt.MapIndex(key)
	if val.Kind() == reflect.Ptr {
		if isNil(val) {
			typ := tgt.Type().Elem()
			val = reflect.New(typ)
			vind := rindirect(val)
			tgt.SetMapIndex(key, val)
			ptr = true // val = vind
			functorLog("    val.typ: %v, key.typ: %v | %v -> %v", typfmt(typ), typfmtv(&key), valfmt(&key), valfmt(&vind))
		}
	} else if !val.IsValid() {
		typ := tgt.Type().Elem()
		val = reflect.New(typ)
		vind := rindirect(val)
		functorLog("    val.typ: %v, key.typ: %v | %v -> %v", typfmt(typ), typfmtv(&key), valfmt(&key), valfmt(&vind))
		tgt.SetMapIndex(key, vind)
		ptr = true // val = vind
	}
	return
}

func mergeOneKeyInMap(c *cpController, params *Params, src, tgt, key reflect.Value) (err error) {
	originalValue := src.MapIndex(key)

	// functorLog("  tgt.type: %v", typfmtv(&tgt))

	keyType := tgt.Type().Key()
	ptrToCopyKey := reflect.New(keyType)
	functorLog("  ptrToCopyKey.type: %v", typfmtv(&ptrToCopyKey))
	ck := ptrToCopyKey.Elem()
	if err = c.copyTo(params, key, ptrToCopyKey); err != nil {
		return
	}

	tgtval, newelemcreated := ensureMapPtrValue(tgt, ck)
	//////////// tgtvalind := rindirect(tgtval)
	if newelemcreated {
		if err = c.copyTo(params, originalValue, tgtval); err != nil {
			return
		}
		tgtval = tgtval.Elem()
		functorLog("Update Map: %v -> %v", ck.Interface(), tgtval.Interface())
	} else {
		eltyp := tgt.Type().Elem()
		eltypind, _ := rskiptype(eltyp, reflect.Ptr)
		ptrToCopyValue := reflect.New(eltypind)
		functorLog("ptrToCopyValue.type: %v", typfmtv(&ptrToCopyValue))
		cv := ptrToCopyValue.Elem()
		if err = c.copyTo(params, tgtval, ptrToCopyValue); err != nil {
			return
		}
		if err = c.copyTo(params, originalValue, ptrToCopyValue); err != nil {
			return
		}
		if cv.Type() == eltyp {
			tgt.SetMapIndex(ck, cv)
			functorLog("SetMapIndex: %v -> %v", ck.Interface(), cv.Interface())
		} else {
			tgt.SetMapIndex(ck, ptrToCopyValue)
			functorLog("SetMapIndex: %v -> %v", ck.Interface(), ptrToCopyValue.Interface())
		}
	}

	return
}

func copy1(c *cpController, params *Params, from, to reflect.Value) (err error) {
	return
}

func copy2(c *cpController, params *Params, from, to reflect.Value) (err error) {
	return
}

func copyDefaultHandler(c *cpController, params *Params, from, to reflect.Value) (err error) {
	fromind := rdecodesimple(from)
	toind := rdecodesimple(to)

	if !toind.IsValid() && to.Kind() == reflect.Ptr {
		tgt := reflect.New(to.Type().Elem())
		toind = rindirect(tgt)
		defer func() {
			if err == nil {
				to.Set(tgt)
			}
		}()
	}

	functorLog("  copyDefaultHandler: %v -> %v | %v", typfmtv(&fromind), typfmtv(&toind), typfmtv(&to))
	if fromind.CanConvert(toind.Type()) {
		if toind.CanSet() {
			toind.Set(fromind.Convert(toind.Type()))
		} else if to.CanSet() {
			to.Set(fromind.Convert(toind.Type()))
		} else {
			err = ErrUnknownState
		}
	} else if from.CanConvert(to.Type()) && to.CanSet() {
		to.Set(from.Convert(to.Type()))
	} else {
		err = ErrCannotSet.WithData(fromind.Interface(), fromind.Kind(), toind.Interface(), toind.Kind())
		log.Errorf("    %v", err)
	}
	return
}

// ErrCannotSet error
var ErrCannotSet = errors.New("cannot set: %v (%v) -> %v (%v)")

// ErrUnknownState error
var ErrUnknownState = errors.New("unknown state, cannot copy to")
