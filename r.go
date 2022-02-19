package deepcopy

import "reflect"

//
// r.go - the routines about reflect operations
//

// rdecode decodes an interface{} or an pointer to something to its
// underlying data type.
//
// Suppose we have an interface{} pointer Value which stored a
// pointer to an integer, rdecode will extract or retrieve the Value
// of that integer.
//
// See our TestRdecode() in r_test.go
//
//    var b = 11
//    var i interface{} = &b
//    var v = reflect.ValueOf(&i)
//    var n = rdecode(v)
//    println(n.Type())    // = int
//
// `prev` returns the previous Value before we arrived at the
// final `ret` Value.
// In another word, the value of `prev` Value is a pointer which
// points to the value of `ret` Value. Or, it's a interface{}
// wrapped about `ret`.
func rdecode(reflectValue reflect.Value) (ret, prev reflect.Value) {
	return rskip(reflectValue, reflect.Pointer, reflect.Interface)
}

func rskip(reflectValue reflect.Value, kinds ...reflect.Kind) (ret, prev reflect.Value) {
	ret, prev = reflectValue, reflectValue
retry:
	k := ret.Kind()
	for _, kk := range kinds {
		if k == kk {
			prev = ret
			ret = ret.Elem()
			goto retry
		}
	}
	return
}

func rindirect(reflectValue reflect.Value) reflect.Value {
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	return reflectValue
}

func rindirectType(reflectType reflect.Type) reflect.Type {
	for reflectType.Kind() == reflect.Ptr || reflectType.Kind() == reflect.Slice {
		reflectType = reflectType.Elem()
	}
	return reflectType
}

func rwant(reflectValue reflect.Value, kinds ...reflect.Kind) reflect.Value {
	k := reflectValue.Kind()
retry:
	for _, kk := range kinds {
		if k == kk {
			return reflectValue
		}
	}

	if k == reflect.Interface || k == reflect.Ptr {
		reflectValue = reflectValue.Elem()
		k = reflectValue.Kind()
		goto retry
	}

	return reflectValue
}
