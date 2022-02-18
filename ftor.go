package deepcopy

import (
	"github.com/hedzr/log"
	"gopkg.in/hedzr/errors.v2"
	"reflect"
	"strings"
	"time"
)

func copyPointer(c *cpController, params *paramsPackage, from, to reflect.Value) (err error) {
	// from is a pointer

	src := c.indirect(from)
	tgt := c.indirect(to)

	if tgt.CanSet() {
		//var newCopy reflect.Value
		//if src.IsValid() {
		//	newCopy = reflect.New(src.Type())
		//} else {
		//	newCopy = src
		//}
		////deepOfSource := c.makeClone(src)
		////newCopy.Elem().Set(deepOfSource)
		//
		//tgt.Set(newCopy) // simple now

		paramsChild := newParams(withOwners(&from, &to, -1))
		params.addChildField(paramsChild)
		err = c.copyTo(paramsChild, src, to)
	} else {
		functorLog("    pointer - tgt is invalid/cannot-set, cannot be set or ignore it: %v -> %v. src.valid: %v", from.Kind(), to.Kind(), src.IsValid())
	}
	return
}

func copyUintptr(c *cpController, params *paramsPackage, from, to reflect.Value) (err error) {
	if to.CanSet() {
		to.Set(from)
	} else {
		//to.SetPointer(from.Pointer())
		functorLog("    copy uintptr not support: %v -> %v", from.Kind(), to.Kind())
	}
	return
}

func copyUnsafePointer(c *cpController, params *paramsPackage, from, to reflect.Value) (err error) {
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

func copyFunc(c *cpController, params *paramsPackage, from, to reflect.Value) (err error) {

	if to.CanSet() {
		to.Set(from)
	} else {
		//to.SetPointer(from.Pointer())
		functorLog("    todo copy function: %v -> %v", from.Kind(), to.Kind())
	}
	return
}

func copyChan(c *cpController, params *paramsPackage, from, to reflect.Value) (err error) {

	if to.CanSet() {
		to.Set(from)
	} else {
		//to.SetPointer(from.Pointer())
		functorLog("    copy chan not support: %v -> %v", from.Kind(), to.Kind())
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

func copyInterface(c *cpController, params *paramsPackage, from, to reflect.Value) (err error) {
	if from.IsNil() {
		return
	}

	paramsChild := newParams(withOwners(&from, &to, -1))
	params.addChildField(paramsChild)

	findirect := c.indirectAny(from)

	functorLog("from.type: %v, decode to: %v", from.Type().Kind(), findirect.Kind())
	functorLog("  to.type: %v, decode to: %v", to.Type().Kind(), to.Elem().Type().Kind())

	merging := c.flags.isAnyFlagsOK(SliceMerge, MapMerge) || params.isAnyFlagsOK(SliceMerge, MapMerge)
	if merging {
		err = c.copyTo(paramsChild, findirect, c.indirectAny(to))

	} else {
		copyValue := reflect.New(findirect.Type()).Elem()
		if err = c.copyTo(paramsChild, findirect, copyValue); err == nil {
			to.Set(copyValue)
		}
	}
	return
}

func copyStruct(c *cpController, params *paramsPackage, from, to reflect.Value) (err error) {
	t, ok := from.Interface().(time.Time)
	if ok {
		to.Set(reflect.ValueOf(t))
		return
	}

	var i int

	defer func() {
		if e := recover(); e != nil {
			ff, tf := from.Field(i), to.Field(i)
			err = errors.New("[recovered] copyStruct unsatisfied ([%v] %v -> [%v] %v), causes: %v",
				ff.Type(), ff, tf.Type(), tf, e).
				AttachGenerals(e)
			//n := log.CalcStackFrames(1)   // skip defer-recover frame at first
			//log.Skip(n).Errorf("%v", err) // skip golib frames and defer-recover frame, back to the point throwing panic
			log.Errorf("%+v", err)
		}
	}()

	//inspectStructV(to)
	padding := strings.Repeat("  ", params.depth())

	for ; i < from.NumField(); i++ {
		fv := from.Field(i)
		if functorLogValid {
			//tt := to.Type().Field(i)
			ft := from.Type().Field(i)
			functorLog(" %s# struct field %d %q: from.type: %v, %d, %v, %q", padding, i, ft.Name, ft.Type, ft.Index, ft.Anonymous, ft.PkgPath)
		}
		//if ft.PkgPath != "" {
		//	continue
		//}
		if !fv.IsValid() {
			functorLog("   ignored invalid source")
			continue
		}

		paramsChild := newParams(withOwners(&from, &to, i))
		if paramsChild.isFlagOK(Ignore) {
			functorLog("   ignored [field tag settings]: %v original %q", paramsChild.fieldTags, paramsChild.fieldTypeSource.Tag)
			continue
		}

		params.addChildField(paramsChild)
		ff, tf := from.Field(i), to.Field(i)
		if cvt, ctx := c.findCopiers(paramsChild, ff, tf); ctx != nil {
			err = cvt.CopyTo(ctx, ff, tf)
		} else if cvt, ctx := c.findConverters(paramsChild, ff, tf); ctx != nil {
			var result reflect.Value
			result, err = cvt.Transform(ctx, ff)
			tf.Set(result)
		} else {
			err = c.copyTo(paramsChild, ff, tf)
		}
	}
	return
}

func copySlice(c *cpController, params *paramsPackage, from, to reflect.Value) (err error) {
	if from.IsNil() {
		return
	}

	// sl, tl := src.Cap(), tgt.Cap()
	if c.flags[SliceMerge] || params.isFlagOK(SliceMerge) {

		if !to.CanAddr() {
			if !params.isStruct() {
				functorLog("use ownerTarget to get a ptr to slice")
				to = *params.ownerTarget
			}
		}

		src := c.indirectAnyAndPtr(from)
		tgt := c.indirectAny(to)

		functorLog("from.type: %v", from.Type().Kind())
		functorLog("  to.type: %v, canAddr: %v", to.Type().Kind(), to.CanAddr())
		functorLog(" src.type: %v, len: %v, cap: %v", src.Type().Kind(), src.Len(), src.Cap())
		functorLog(" tgt.type: %v, len: %v, cap: %v, canAddr: %v", tgt.Type().Kind(), tgt.Len(), tgt.Cap(), tgt.CanAddr())

		//paramsChild := newParams(withOwners(&from, &to, -1))
		//params.addChildField(paramsChild)

		tl, sl := tgt.Len(), src.Len()
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
		to.Set(ns)

	} else if params != nil && params.isFlagOK(SliceCopyEnh) {

		// ftfCopyEnh

		to = c.ensureIsSlicePtr(to)
		src := c.indirect(from)
		for i := 0; i < src.Len(); i++ {
			si := src.Index(i)
			var found bool
			for j := 0; j < to.Elem().Len(); j++ {
				ti := to.Elem().Index(j)
				if found = reflect.DeepEqual(si, ti); found {
				}
			}
			if !found {
				to.Elem().Set(reflect.Append(to.Elem(), si))
			}
		}

	} else {

		// copy and set each source element to target slice

		//to = c.ensureIsSlicePtr(to)
		src := c.indirect(from)
		for i := 0; i < src.Len(); i++ {
			to.Set(reflect.Append(to, src.Index(i)))
		}

	}
	return
}

func copyArray(c *cpController, params *paramsPackage, from, to reflect.Value) (err error) {
	if from.IsZero() {
		return
	}

	src := c.indirect(from)
	tgt := c.indirect(to)

	//functorLog("    from.type: %v, len: %v, cap: %v", src.Type().Kind(), src.Len(), src.Cap())
	//functorLog("      to.type: %v, len: %v, cap: %v", tgt.Type().Kind(), tgt.Len(), tgt.Cap())

	set := src.Index(0).Type()
	if set != tgt.Index(0).Type() {
		return errors.New("cannot copy %v to %v", from.Interface(), to.Interface())
	}

	sl, tl := src.Cap(), tgt.Cap()
	cnt := sl
	if sl > tl {
		cnt = tl
	}

	//pt := c.ensureIsSlicePtr(tgt)
	for i := 0; i < cnt; i++ {
		to.Index(i).Set(src.Index(i))
	}
	//to.Set(pt.Elem())

	functorLog("    from: %v, to: %v", src.Interface(), tgt.Interface()) // pt.Interface())

	return
}

func copyMap(c *cpController, params *paramsPackage, from, to reflect.Value) (err error) {
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

func copy1(c *cpController, params *paramsPackage, from, to reflect.Value) (err error) {
	return
}

func copy2(c *cpController, params *paramsPackage, from, to reflect.Value) (err error) {
	return
}

func copyDefaultHandler(c *cpController, params *paramsPackage, from, to reflect.Value) (err error) {
	if from.CanConvert(to.Type()) && to.CanSet() {
		to.Set(from)
	} else {
		log.Errorf("    cannot Set: %v (%v) -> %v (%v)", from.Interface(), from.Kind(), to.Interface(), to.Kind())
	}
	return
}
