package deepcopy

import (
	"github.com/hedzr/log"
	"gopkg.in/hedzr/errors.v2"
	"reflect"
	"time"
)

func copyPointer(c *cpController, params *paramsPackage, from, to reflect.Value) (err error) {
	// from is a pointer

	src := c.indirect(from)
	tgt := c.indirect(to)

	//newCopy := reflect.New(src.Type())
	//deepOfSource := c.makeClone(src)
	//newCopy.Elem().Set(deepOfSource)
	if tgt.IsValid() {
		tgt.Set(src) // simple now
	} else {
		functorLog("pointer - tgt is invalid, can be set: %v -> %v", from.Kind(), to.Kind())
	}
	return
}

func copyUintptr(c *cpController, params *paramsPackage, from, to reflect.Value) (err error) {
	//to.SetPointer(from.Pointer())
	functorLog("copy uintptr not support: %v -> %v", from.Kind(), to.Kind())
	return
}

func copyUnsafePointer(c *cpController, params *paramsPackage, from, to reflect.Value) (err error) {
	functorLog("copy unsafe pointer not support: %v -> %v", from.Kind(), to.Kind())
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
	functorLog("todo copy function: %v -> %v", from.Kind(), to.Kind())
	return
}

func copyChan(c *cpController, params *paramsPackage, from, to reflect.Value) (err error) {
	functorLog("copy chan not support: %v -> %v", from.Kind(), to.Kind())
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

	find := from.Elem()
	copyValue := reflect.New(find.Type()).Elem()
	if err = c.copyTo(find, copyValue, params); err == nil {
		to.Set(copyValue)
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

	inspectStructV(to)

	for ; i < from.NumField(); i++ {
		ft := from.Type().Field(i)
		if ft.PkgPath != "" {
			continue
		}
		paramsField := &paramsPackage{
			owner:       &from,
			index:       i,
			fieldType:   &ft,
			fieldTags:   parseFieldTags(ft.Tag),
			ownerTarget: &to,
		}

		ff, tf := from.Field(i), to.Field(i)
		err = c.copyTo(ff, tf, paramsField)
	}
	return
}

func copySlice(c *cpController, params *paramsPackage, from, to reflect.Value) (err error) {
	if from.IsNil() {
		return
	}

	src := c.indirect(from)
	tgt := c.indirect(to)

	//functorLog("from.type: %v, len: %v, cap: %v", src.Type().Kind(), src.Len(), src.Cap())
	//functorLog("  to.type: %v, len: %v, cap: %v", tgt.Type().Kind(), tgt.Len(), tgt.Cap())

	// sl, tl := src.Cap(), tgt.Cap()
	if c.mergeSlice {
		tl := tgt.Len()
		ns := reflect.MakeSlice(tgt.Type(), tl, tgt.Cap())
		paramsSlice := &paramsPackage{
			owner:       &from,
			index:       0,
			ownerTarget: &ns,
		}
		for i := 0; i < tl; i++ {
			v := tgt.Index(i)
			paramsSlice.index = i
			err = c.copyTo(v, ns.Index(i), paramsSlice)
		}
		for i := 0; i < src.Len(); i++ {
			found, el := false, src.Index(i)
			elv := el.Interface()
			for j := 0; j < ns.Len(); j++ {
				tev := ns.Index(j).Interface()
				//functorLog("  testing tgt[%v](%v) and src[%v](%v)", j, tev, i, elv)
				if reflect.DeepEqual(tev, elv) {
					found = true
					//functorLog("found exists el at tgt[%v], for src[%v], value is %v", j, i, elv)
					break
				}
			}
			if !found {
				ns = reflect.Append(ns, el)
			}
			//functorLog("new tgt: %v", ns.Interface())
		}
		tgt.Set(ns)
	} else {
		if params != nil && params.fieldTags.flags[""] {
		}
		// copy and set each source element to target slice
		for i := 0; i < src.Len(); i++ {
			tgt = reflect.Append(tgt, src.Index(i))
		}

		tl := tgt.Len()
		ns := reflect.MakeSlice(tgt.Type(), tl, tgt.Cap())
		paramsSlice := &paramsPackage{
			owner:       &from,
			index:       0,
			fieldType:   nil,
			ownerTarget: &ns,
		}
		for i := 0; i < tl; i++ {
			paramsSlice.index = i
			err = c.copyTo(tgt.Index(i), ns.Index(i), paramsSlice)
		}
		for i := 0; i < src.Len(); i++ {
			ns = reflect.Append(ns, src.Index(i))
		}
		tgt.Set(ns)
	}
}

return
}

func copyArray(c *cpController, params *paramsPackage, from, to reflect.Value) (err error) {
	if from.IsNil() {
		return
	}

	src := c.indirect(from)
	tgt := c.indirect(to)

	//functorLog("from.type: %v, len: %v, cap: %v", src.Type().Kind(), src.Len(), src.Cap())
	//functorLog("  to.type: %v, len: %v, cap: %v", tgt.Type().Kind(), tgt.Len(), tgt.Cap())

	sl, tl := src.Cap(), tgt.Cap()
	set := src.Index(0).Type()
	if set != tgt.Index(0).Type() {
		return errors.New("cannot copy %v to %v", from.Interface(), to.Interface())
	}

	if sl > tl {
		//typ := reflect.ArrayOf(sl, set)
		//ary := reflect.New(typ)
		//ae := ary.Elem()
		//for i := 0; i < tl; i++ {
		//	ae.Index(i).Set(tgt.Index(i))
		//}
		//for i := 0; i < sl; i++ {
		//	ae.Index(i).Set(src.Index(i))
		//}
		//to.Set(ary)

		for i := 0; i < tl; i++ {
			tgt.Index(i).Set(src.Index(i))
		}
	} else {
		for i := 0; i < sl; i++ {
			tgt.Index(i).Set(src.Index(i))
		}
	}

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
		err = c.copyTo(originalValue, copyValue, params)
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
		log.Errorf("cannot Set: %v (%v) -> %v (%v)", from.Interface(), from.Kind(), to.Interface(), to.Kind())
	}
	return
}
