package deepcopy

import (
	"fmt"
	"github.com/hedzr/log"
	"gopkg.in/hedzr/errors.v3"
	"reflect"
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
		functorLog("    pointer - tgt is invalid/cannot-be-set/ignored: %v -> %v. src.valid: %v", from.Kind(), to.Kind(), src.IsValid())
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
		return
	}

	paramsChild := newParams(withOwners(params, &from, &to, nil, nil, 0))
	defer paramsChild.revoke()

	// unbox the interface{} to original data type
	toind, toptr := c.decode(to) // c.skip(to, reflect.Interface, reflect.Pointer)

	functorLog("from.type: %v, decode to: %v", from.Type().Kind(), paramsChild.srcDecoded.Kind())
	functorLog("  to.type: %v, decode to: %v | CanSet: %v, CanAddr: %v", to.Type().Kind(), toind.Kind(), toind.CanSet(), toind.CanAddr())

	merging := c.flags.isAnyFlagsOK(SliceMerge, MapMerge) || params.isAnyFlagsOK(SliceMerge, MapMerge)
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
		i, amount int
		padding   string
		f, _      = rdecode(from)
		t, _      = rdecode(to)
	)

	if functorLogValid {
		padding = strings.Repeat("  ", params.depth()*2)
		fromT, toT := f.Type(), t.Type()
		//functorLog(" %s  %d, %d, %d", padding, params.index, params.srcOffset, params.dstOffset)
		fq := dbgMakeInfoString(fromT, params, true)
		dq := dbgMakeInfoString(toT, params, false)
		functorLog(" %s- (%v (%v)) -> dst (%v (%v))", padding, fromT, fromT.Kind(), toT, toT.Kind())
		functorLog(" %s  %s -> %s", padding, fq, dq)
	}

	var ec = errors.New("copyStruct errors")
	defer func() {
		if e := recover(); e != nil {
			ff, tf := f.Field(i), t.Field(i)
			err = errors.New("[recovered] copyStruct unsatisfied ([%v] %v -> [%v] %v), causes: %v",
				ff.Type(), ff, tf.Type(), tf, e).
				WithData(e).                      // collect e if it's an error object else store it simply
				WithTaggedData(errors.TaggedData{ // record the sites
					"source-field": ff,
					"target-field": tf,
					"source":       valfmt(&ff),
					"target":       valfmt(&tf),
				})
			//n := log.CalcStackFrames(1)   // skip defer-recover frame at first
			//log.Skip(n).Errorf("%v", err) // skip golib frames and defer-recover frame, back to the point throwing panic
			log.Errorf("%+v", err)
		}
	}()

	var sourcefields fieldstable
	sourcefields = sourcefields.getallfields(f, c.autoExpandStuct)
	targetIterator := newStructIterator(t, withStructPtrAutoExpand(c.autoExpandStuct))
	for i, amount = 0, len(sourcefields.records); i < amount; i++ {
		sourcefield := sourcefields.records[i]
		flags := parseFieldTags(sourcefield.structField.Tag)
		if flags.isFlagExists(Ignore) {
			continue
		}
		accessor, ok := targetIterator.Next()
		if !ok {
			continue
		}
		srcval, dstval := sourcefield.Value(), accessor.FieldValue()
		functorLog("%d. %s (%v) -> %s (%v)", i, strings.Join(reverseStringSlice(sourcefield.names), "."), valfmt(srcval), accessor.StructFieldName(), valfmt(dstval))
		ec.Attach(invokeStructFieldTransformer(c, params, *srcval, *dstval, padding))
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
	if cvt, ctx := c.findCopiers(params, ff, df); ctx != nil {
		err = cvt.CopyTo(ctx, ff, df) // use user-defined copy-n-merger to merge or copy source to destination
	} else if cvt, ctx := c.findConverters(params, ff, df); ctx != nil {
		var result reflect.Value
		result, err = cvt.Transform(ctx, ff) // use user-defined value converter to transform from source to destination
		df.Set(result)
	} else {
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
		//
	}

	for _, flag := range []CopyMergeStrategy{SliceMerge, SliceCopyAppend, SliceCopy} {
		if c.flags.isGroupedFlagOK(flag) || params.isGroupedFlagOK(flag) {

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

			result := mSliceOper[flag](from, tgt)
			//tgt=ns
			//t := c.want2(to, reflect.Slice, reflect.Interface)
			//t.Set(tgt)
			//   //tgtptr.Elem().Set(result)
			if tgtptr.Kind() == reflect.Slice {
				tgtptr.Set(result)
			} else {
				tgtptr.Elem().Set(result)
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

var mSliceOper = map[CopyMergeStrategy]func(src, tgt reflect.Value) (result reflect.Value){
	SliceCopy: func(src, tgt reflect.Value) (result reflect.Value) {
		sl := src.Len()
		ns := reflect.MakeSlice(tgt.Type(), 0, 0)
		for _, ss := range []struct {
			length int
			source reflect.Value
		}{
			// {tl, tgt},
			{sl, src},
		} {
			for i := 0; i < ss.length; i++ {
				el := ss.source.Index(i)
				ns = reflect.Append(ns, el)
			}
		}
		result = ns
		return
	},
	SliceCopyAppend: func(src, tgt reflect.Value) (result reflect.Value) {
		sl, tl := src.Len(), tgt.Len()
		ns := reflect.MakeSlice(tgt.Type(), 0, 0)
		for _, ss := range []struct {
			length int
			source reflect.Value
		}{
			{tl, tgt},
			{sl, src},
		} {
			for i := 0; i < ss.length; i++ {
				el := ss.source.Index(i)
				ns = reflect.Append(ns, el)
			}
		}
		result = ns
		return
	},
	SliceMerge: func(src, tgt reflect.Value) (result reflect.Value) {
		sl, tl := src.Len(), tgt.Len()
		ns := reflect.MakeSlice(tgt.Type(), 0, 0)
		for _, ss := range []struct {
			length int
			source reflect.Value
		}{
			{tl, tgt},
			{sl, src},
		} {
			for i := 0; i < ss.length; i++ {
				//to.Set(reflect.Append(to, src.Index(i)))
				found, el := false, ss.source.Index(i)
				elv := el.Interface()
				for j := 0; j < ns.Len(); j++ {
					tev := ns.Index(j).Interface()
					functorLog("  testing tgt[%v](%v) and src[%v](%v)", j, tev, i, elv)
					if reflect.DeepEqual(tev, elv) {
						found = true
						functorLog("found exists el at tgt[%v], for src[%v], value is %v", j, i, elv)
						break
					}
				}
				if !found {
					ns = reflect.Append(ns, el)
				}
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
	to.Set(reflect.MakeMap(from.Type()))
	for _, key := range from.MapKeys() {
		originalValue := from.MapIndex(key)
		copyValue := reflect.New(originalValue.Type()).Elem()
		err = c.copyTo(params, originalValue, copyValue)
		copyKey := MakeClone(key.Interface())
		to.SetMapIndex(reflect.ValueOf(copyKey), copyValue)
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
	if fromind.CanConvert(toind.Type()) && toind.CanSet() {
		toind.Set(fromind)
	} else if from.CanConvert(to.Type()) && to.CanSet() {
		to.Set(from)
	} else {
		log.Errorf("    cannot Set: %v (%v) -> %v (%v)", fromind.Interface(), fromind.Kind(), toind.Interface(), toind.Kind())
	}
	return
}
