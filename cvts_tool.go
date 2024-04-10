package evendeep

import (
	"fmt"
	"log"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gopkg.in/hedzr/errors.v3"

	"github.com/hedzr/evendeep/dbglog"
	"github.com/hedzr/evendeep/internal/cl"
	"github.com/hedzr/evendeep/internal/syscalls"
	"github.com/hedzr/evendeep/ref"
)

// rForBool transform bool -> string.
func rForBool(v reflect.Value) (ret reflect.Value) {
	vs := strconv.FormatBool(v.Bool())
	ret = reflect.ValueOf(vs)
	return
}

// rToBool transform string (or anything) -> bool.
func rToBool(v reflect.Value) (ret reflect.Value) {
	var b bool

	if !v.IsValid() || ref.IsNil(v) || ref.IsZero(v) {
		return reflect.ValueOf(b)
	}

	k := v.Kind()
	switch k { //nolint:exhaustive //no need
	case reflect.Bool:
		b = v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b = v.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		b = v.Uint() != 0
	case reflect.Float32, reflect.Float64:
		b = math.Float64bits(v.Float()) != 0
	case reflect.Complex64, reflect.Complex128:
		c := v.Complex()
		b = math.Float64bits(real(c)) != 0 || math.Float64bits(imag(c)) != 0
	case reflect.Array:
		b = !ref.ArrayIsZero(v)
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		b = !ref.IsNil(v)
	case reflect.Struct:
		b = !ref.StructIsZero(v)
	case reflect.String:
		b = internalToBool(v.String())
	}

	ret = reflect.ValueOf(b)
	return
}

func internalToBool(s string) (b bool) {
	switch val := strings.ToLower(s); val {
	case "y", "t", "1", "yes", "true", "on", "ok", "m", "male":
		b = true
	}
	return
}

// rForInteger transform integer -> string.
func rForInteger(v reflect.Value) (ret reflect.Value) {
	int64typ := reflect.TypeOf((*int64)(nil)).Elem()
	if !v.IsValid() {
		v = reflect.Zero(int64typ)
	} else if k := v.Kind(); k < reflect.Int || k > reflect.Int64 {
		if ref.CanConvert(&v, int64typ) {
			v = v.Convert(int64typ)
		} else {
			v = reflect.Zero(int64typ)
		}
	}

	vs := strconv.FormatInt(v.Int(), 10)
	ret = reflect.ValueOf(vs)
	return
}

const uintSize = 32 << (^uint(0) >> 32 & 1)

//nolint:dupl //don't
func rToInteger(v reflect.Value, desiredType reflect.Type) (ret reflect.Value, err error) {
	genret := func(ival int64, desiredTypeKind reflect.Kind) (ret reflect.Value) {
		switch desiredTypeKind { //nolint:exhaustive //no need
		case reflect.Int:
			ret = reflect.ValueOf(int(ival))
		case reflect.Int32:
			ret = reflect.ValueOf(int32(ival))
		case reflect.Int16:
			ret = reflect.ValueOf(int16(ival))
		case reflect.Int8:
			ret = reflect.ValueOf(int8(ival))
		default:
			ret = reflect.ValueOf(ival)
		}
		return
	}
	ret, err = toTypeConverter(v, desiredType, 10, //nolint:gomnd //no need
		func(str string, base, bitSize int) (ret reflect.Value, err error) {
			var ival int64
			ival, err = strconv.ParseInt(str, base, bitSize)
			if err == nil {
				k := desiredType.Kind()
				ret = genret(ival, k)
			} else {
				var fval float64
				fval, err = strconv.ParseFloat(str, bitSize)
				if err == nil {
					ival = int64(math.Floor(fval + 0.5)) //nolint:gomnd //no need
					k := desiredType.Kind()
					ret = genret(ival, k)
				}
			}
			return
		})
	return
}

// rForUInteger transform uint64 -> string.
func rForUInteger(v reflect.Value) (ret reflect.Value) {
	uint64typ := reflect.TypeOf((*uint64)(nil)).Elem()
	if !v.IsValid() {
		v = reflect.Zero(uint64typ)
	} else if k := v.Kind(); k < reflect.Uint || k > reflect.Uintptr {
		if ref.CanConvert(&v, uint64typ) {
			v = v.Convert(uint64typ)
		} else {
			v = reflect.Zero(uint64typ)
		}
	}

	vs := strconv.FormatUint(v.Uint(), 10)
	ret = reflect.ValueOf(vs)
	return
}

//nolint:dupl //don't
func rToUInteger(v reflect.Value, desiredType reflect.Type) (ret reflect.Value, err error) {
	genret := func(ival uint64, desiredTypeKind reflect.Kind) (ret reflect.Value) {
		switch desiredTypeKind { //nolint:exhaustive //no need
		case reflect.Uint:
			ret = reflect.ValueOf(uint(ival))
		case reflect.Uint32:
			ret = reflect.ValueOf(uint32(ival))
		case reflect.Uint16:
			ret = reflect.ValueOf(uint16(ival))
		case reflect.Uint8:
			ret = reflect.ValueOf(uint8(ival))
		default:
			ret = reflect.ValueOf(ival)
		}
		return
	}
	ret, err = toTypeConverter(v, desiredType, 10, //nolint:gomnd //no need
		func(str string, base, bitSize int) (ret reflect.Value, err error) {
			var ival uint64
			ival, err = strconv.ParseUint(str, base, bitSize)
			if err == nil {
				k := desiredType.Kind()
				ret = genret(ival, k)
			} else {
				var fval float64
				fval, err = strconv.ParseFloat(str, bitSize)
				if err == nil {
					ival = uint64(math.Floor(fval + 0.5)) //nolint:gomnd //no need
					k := desiredType.Kind()
					ret = genret(ival, k)
				}
			}
			return
		})
	return
}

// rForUIntegerHex transform uintptr/... -> string.
func rForUIntegerHex(u uintptr) (ret reflect.Value) {
	vs := syscalls.UintptrToString(u)
	ret = reflect.ValueOf(vs)
	return
}

func rToUIntegerHex(s reflect.Value, desiredType reflect.Type) (ret reflect.Value) {
	vs := syscalls.UintptrFromString(s.String())
	log.Printf("vs : %v, k: %v\n", vs, desiredType.Kind())
	switch k := desiredType.Kind(); k { //nolint:exhaustive //no need
	case reflect.Uintptr:
		ret = reflect.ValueOf(vs)
	case reflect.UnsafePointer, reflect.Ptr:
		ret = reflect.ValueOf(vs)
		// u := getPointerAsUintptr(ret)
	}
	return
}

func getPointerAsUintptr(v reflect.Value) uintptr { //nolint:unused // intergrality
	var p uintptr
	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct {
		p = v.Pointer()
	}
	return p
}

// rForFloat transform float -> string.
func rForFloat(v reflect.Value) (ret reflect.Value) {
	float64typ := reflect.TypeOf((*float64)(nil)).Elem()
	if !v.IsValid() {
		v = reflect.Zero(float64typ)
	} else if k := v.Kind(); k < reflect.Float32 || k > reflect.Float64 {
		if ref.CanConvert(&v, float64typ) {
			v = v.Convert(float64typ)
		} else {
			v = reflect.Zero(float64typ)
		}
	}

	vs := strconv.FormatFloat(v.Float(), 'g', -1, 64)
	ret = reflect.ValueOf(vs)
	return
}

func rToFloat(v reflect.Value, desiredType reflect.Type) (ret reflect.Value, err error) {
	toFloat := func(fval float64, desiredType reflect.Type) (ret reflect.Value) {
		if desiredType.Kind() == reflect.Float64 {
			ret = reflect.ValueOf(fval)
		} else {
			ret = reflect.ValueOf(float32(fval))
		}
		return
	}
	ret, err = toTypeConverter(v, desiredType, 10, //nolint:gomnd //no need
		func(str string, base, bitSize int) (ret reflect.Value, err error) {
			var fval float64
			fval, err = strconv.ParseFloat(str, bitSize)
			if err == nil {
				return toFloat(fval, desiredType), nil
			}

			var ival int64
			ival, err = strconv.ParseInt(str, 10, bitSize)
			if err == nil {
				return toFloat(float64(ival), desiredType), nil
			}

			var uval uint64
			uval, err = strconv.ParseUint(str, 10, bitSize)
			if err == nil {
				return toFloat(float64(uval), desiredType), nil
			}

			var cval complex128
			cval, err = cl.ParseComplex(str)
			if err == nil {
				fval = real(cval)
				ret = toFloat(fval, desiredType)
			}
			return
		})
	return
}

// rForComplex transform complex -> string.
func rForComplex(v reflect.Value) (ret reflect.Value) {
	complex128typ := reflect.TypeOf((*complex128)(nil)).Elem()
	if !v.IsValid() {
		v = reflect.Zero(complex128typ)
	} else if k := v.Kind(); k < reflect.Complex64 || k > reflect.Complex128 {
		// if canConvert(&v, complex128typ) {
		//	v = v.Convert(complex128typ)
		// } else {
		switch k { //nolint:exhaustive //no need
		case reflect.Float64, reflect.Float32:
			v = reflect.ValueOf(complex(v.Float(), 0.0))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v = reflect.ValueOf(complex(float64(v.Int()), 0.0))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			v = reflect.ValueOf(complex(float64(v.Uint()), 0.0))
		default:
			v = reflect.Zero(complex128typ)
		}
		// }
	}

	vs := cl.FormatComplex(v.Complex(), 'g', -1, 128) //nolint:gomnd //no need
	ret = reflect.ValueOf(vs)
	return
}

func rToComplex(v reflect.Value, desiredType reflect.Type) (ret reflect.Value, err error) {
	toComplex := func(cval complex128, desiredType reflect.Type) (ret reflect.Value) {
		if desiredType.Kind() == reflect.Complex128 {
			ret = reflect.ValueOf(cval)
		} else {
			ret = reflect.ValueOf(complex64(cval))
		}
		return
	}
	ret, err = toTypeConverter(v, desiredType, 10, //nolint:gomnd //no need
		func(str string, base, bitSize int) (ret reflect.Value, err error) {
			var cval complex128
			if str[0] != '(' {
				str = "(" + str
			}
			if lastch := str[len(str)-1]; lastch != ')' {
				if lastch != 'i' {
					str += "+0i"
				}
				str += ")"
			}
			cval, err = cl.ParseComplex(str)
			if err == nil {
				ret = toComplex(cval, desiredType)
			}
			return
		})
	return
}

func toTypeConverter(v reflect.Value, desiredType reflect.Type, base int, //nolint:unparam
	converter func(str string, int, bitSize int) (ret reflect.Value, err error),
) (ret reflect.Value, err error) {
	if !v.IsValid() || ref.IsNil(v) || ref.IsZero(v) {
		ret = reflect.Zero(desiredType)
		return
	}

	if ref.CanConvert(&v, desiredType) {
		ret = v.Convert(desiredType)
		return
	}

	val := v.String()
	bitSize := 64
	switch k := desiredType.Kind(); k { //nolint:exhaustive //no need
	case reflect.Int, reflect.Uint:
		bitSize = uintSize
	case reflect.Int32, reflect.Uint32, reflect.Float32:
		bitSize = 32
	case reflect.Int16, reflect.Uint16:
		bitSize = 16
	case reflect.Int8, reflect.Uint8:
		bitSize = 8
	case reflect.Complex128:
		bitSize = 128
		// case reflect.Float64, reflect.Complex64:
		//	bitSize = 64
	}
	ret, err = converter(val, base, bitSize)
	return
}

func tryStringerIt(source reflect.Value, desiredType reflect.Type) (target reflect.Value, processed bool, err error) { //nolint:unparam
	val := source.Interface()
	if ss, ok := val.(interface{ String() string }); ok {
		nv := ss.String()
		target = reflect.ValueOf(nv)
		processed = true
		return
	}

	if ref.CanConvert(&source, desiredType) {
		nv := source.Convert(desiredType)
		// target.Set(nv)
		target = nv
		processed = true
	} else {
		nv := fmt.Sprintf("%v", val)
		// target.Set(reflect.ValueOf(nv))
		target = reflect.ValueOf(nv)
		processed = true
	}
	return
}

func rToString(source reflect.Value, desiredType reflect.Type) (target reflect.Value, err error) {
	if source.IsValid() {
		switch k := source.Kind(); k { //nolint:exhaustive //no need
		case reflect.Bool:
			target = rForBool(source)
		case reflect.Int64:
			var processed bool
			if target, processed, err = tryStringerIt(source, desiredType); processed {
				return
			}
			fallthrough
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
			target = rForInteger(source)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			target = rForUInteger(source)

		case reflect.Uintptr:
			target = rForUIntegerHex(uintptr(source.Uint()))
		// case reflect.UnsafePointer:
		//	target = rForUIntegerHex(uintptr(source.Uint()))
		// case reflect.Ptr:
		//	target = rForUIntegerHex(source.Pointer())

		case reflect.Float32, reflect.Float64:
			target = rForFloat(source)
		case reflect.Complex64, reflect.Complex128:
			target = rForComplex(source)

		case reflect.String:
			target = reflect.ValueOf(source.String())

		// reflect.Array
		// reflect.Chan
		// reflect.Func
		// reflect.Interface
		// reflect.Map
		// reflect.Slice
		// reflect.Struct

		default:
			target, _, err = tryStringerIt(source, desiredType)
		}
	} else {
		target = reflect.Zero(ref.StringType)
	}
	return
}

//nolint:unparam,unused,deadcode,lll //reserved
func rToArray(ctx *ValueConverterContext, sources reflect.Value, desiredType reflect.Type, targetLength int) (target reflect.Value, err error) {
	eltyp := desiredType.Elem() // length := desiredType.Len()
	dbglog.Log("  desiredType: %v, el.type: %v", ref.Typfmt(desiredType), ref.Typfmt(eltyp))

	count, length := sources.Len(), targetLength
	if length <= 0 {
		length = count
	}
	if count > length {
		count = length
	}

	target = reflect.New(reflect.ArrayOf(length, eltyp)).Elem()

	for ix := 0; ix < count; ix++ {
		src := sources.Index(ix)
		if ix < length && src.IsValid() {
			if src.Type().AssignableTo(eltyp) {
				target.Index(ix).Set(src)
			} else if src.Type().ConvertibleTo(eltyp) {
				target.Index(ix).Set(src.Convert(eltyp))
			}
		}
	}
	_ = ctx
	return
}

//nolint:unparam,unused,deadcode,lll //reserved
func rToSlice(ctx *ValueConverterContext, sources reflect.Value, desiredType reflect.Type, targetLength int) (target reflect.Value, err error) {
	eltyp := desiredType.Elem() // length := desiredType.Len()
	dbglog.Log("  desiredType: %v, el.type: %v", ref.Typfmt(desiredType), ref.Typfmt(eltyp))

	count, length := sources.Len(), targetLength
	if length <= 0 {
		length = count
	}
	if count > length {
		count = length
	}

	target = reflect.MakeSlice(desiredType, length, length)
	for ix := 0; ix < count; ix++ {
		src := sources.Index(ix)
		if ix < length && src.IsValid() {
			if src.Type().AssignableTo(eltyp) {
				target.Index(ix).Set(src)
			} else if src.Type().ConvertibleTo(eltyp) {
				target.Index(ix).Set(src.Convert(eltyp))
			}
		}
	}
	_ = ctx
	return
}

//nolint:unparam,unused,deadcode,lll //reserved
func rToMap(ctx *ValueConverterContext, source reflect.Value, fromFuncType, desiredType reflect.Type) (target reflect.Value, err error) {
	ec := errors.New("cannot transform item into map")
	defer ec.Defer(&err)

	dtyp := desiredType.Elem()
	styp := source.Type()
	// srceltyp := styp.Elem()
	target = reflect.MakeMap(desiredType)

	if source.IsValid() {
		for ix, key := range source.MapKeys() {
			val := source.MapIndex(key)
			if err = rSetMapValue(ix, target, key, val, styp, dtyp); err != nil {
				ec.Attach(err)
				continue
			}
		}
	}

	_, _ = ctx, fromFuncType

	// for i := 0; i < fromFuncType.NumOut(); i++ {
	//	if i >= len(sources) {
	//		continue
	//	}
	//
	//	styp := fromFuncType.Out(i)
	//	sname := styp.Name()
	//	var key reflect.Value
	//	if key, err = nameToMapKey(sname, desiredType); err != nil {
	//		ec.Attach(err)
	//		continue
	//	}
	//
	//	if err = rSetMapValue(target, key, sources[i], styp, dtyp); err != nil {
	//		ec.Attach(err)
	//		continue
	//	}
	// }
	return
}

//nolint:unused //future
func rSetMapValue(ix int, target, key, srcVal reflect.Value, sTyp, dTyp reflect.Type) (err error) {
	if sTyp.AssignableTo(dTyp) { //nolint:gocritic // no need to switch to 'switch' clause
		target.SetMapIndex(key, srcVal)
	} else if sTyp.ConvertibleTo(dTyp) {
		target.SetMapIndex(key, srcVal.Convert(dTyp))
	} else {
		dstval := target.MapIndex(key)
		err = errors.New("cannot set map[%v] since transforming/converting failed: %v -> %v",
			ref.Valfmt(&key), ref.Valfmt(&srcVal), ref.Valfmt(&dstval))
	}
	_ = ix
	return
}

//nolint:unused //future
func nameToMapKey(name string, mapType reflect.Type) (key reflect.Value, err error) {
	nameval := reflect.ValueOf(name)
	nametyp := nameval.Type()

	if keytyp := mapType.Key(); nametyp.AssignableTo(keytyp) { //nolint:gocritic // no need to switch to 'switch' clause
		key = nameval
	} else if nametyp.ConvertibleTo(keytyp) {
		key = nameval.Convert(keytyp)
	} else {
		cvt := fromStringConverter{}
		key, err = cvt.Transform(nil, nameval, keytyp)
	}
	return
}

//nolint:unused,deadcode,lll //reserved
func rToStruct(ctx *ValueConverterContext, source reflect.Value, fromFuncType, desiredType reflect.Type) (target reflect.Value, err error) {
	// result (source) -> struct (target)
	_, _, _, _ = ctx, source, fromFuncType, desiredType
	return
}

//nolint:unused,deadcode,lll //reserved
func rToFunc(ctx *ValueConverterContext, source reflect.Value, fromFuncType, desiredType reflect.Type) (target reflect.Value, err error) {
	_, _, _, _ = ctx, source, fromFuncType, desiredType
	return
}

//

//

//

func intToString[T Integers](val T) string {
	return intToStringEx(val, 10)
}

func intToStringEx[T Integers](val T, base int) string {
	return strconv.FormatInt(int64(val), base)
}

func uintToString[T Uintegers](val T) string {
	return uintToStringEx(val, 10)
}

func uintToStringEx[T Uintegers](val T, base int) string {
	return strconv.FormatUint(uint64(val), base)
}

func floatToString[T Floats](val T) string {
	return floatToStringEx(float64(val), 'f', -1, 64)
}

func floatToStringEx[T Floats](val T, format byte, prec, bitSize int) string {
	return strconv.FormatFloat(float64(val), format, prec, bitSize)
}

func complexToString[T Complexes](val T) string {
	return complexToStringEx(val, 'f', -1, 128)
}

func complexToStringEx[T Complexes](val T, format byte, prec, bitSize int) string {
	return strconv.FormatComplex(complex128(val), format, prec, bitSize)
}

func boolToString(b bool) string {
	return strconv.FormatBool(b)
}

func timeToString(tm time.Time) string {
	const layout = time.RFC3339Nano
	return tm.Format(layout)
}

func durationToString(dur time.Duration) string {
	return dur.String()
}

//

func bytesToString(data []byte) string {
	b := make([]byte, 0, 3+len(data)*4)
	b = append(b, data...)
	b = append(b, []byte(" [")...)
	for i := 0; i < len(data); {
		if i > 0 {
			b = append(b, ' ')
		}
		strconv.AppendInt(b, int64(data[i]), 16)
	}
	b = append(b, []byte("]")...)
	return string(b)
}

func intSliceToString[T Integers](val IntSlice[T]) string {
	b := make([]byte, 0, len(val)*8) // 8: assume integer need 8 runes
	b = append(b, []byte("[")...)
	for i := range val {
		if i > 0 {
			b = append(b, []byte(",")...)
		}
		b = strconv.AppendInt(b, int64(val[i]), 10)
	}
	b = append(b, []byte("]")...)
	return string(b)
}

func uintSliceToString[T Uintegers](val UintSlice[T]) string {
	b := make([]byte, 0, len(val)*8) // 8: assume unsigned integer need 8 runes
	b = append(b, []byte("[")...)
	for i := range val {
		if i > 0 {
			b = append(b, []byte(",")...)
		}
		b = strconv.AppendUint(b, uint64(val[i]), 10)
	}
	b = append(b, []byte("]")...)
	return string(b)
}

func floatSliceToString[T Floats](val FloatSlice[T]) string {
	b := make([]byte, 0, len(val)*16+2) // 8: assume floats need 16 runes
	b = append(b, []byte("[")...)
	for i := range val {
		if i > 0 {
			b = append(b, []byte(",")...)
		}
		b = strconv.AppendFloat(b, float64(val[i]), 'f', -1, 64)
	}
	b = append(b, []byte("]")...)
	return string(b)
}

func complexSliceToString[T Complexes](val ComplexSlice[T]) string {
	b := make([]byte, 0, len(val)*32+2) // 8: assume complex need 32 runes
	b = append(b, []byte("[")...)
	for i := range val {
		if i > 0 {
			b = append(b, []byte(",")...)
		}
		num := strconv.FormatComplex(complex128(val[i]), 'f', -1, 128)
		b = append(b, []byte(num)...)
	}
	b = append(b, []byte("]")...)
	return string(b)
}

// func complexSliceToString[T Complexes](val ComplexSlice[T]) string {
// 	var b = make([]byte, 0, len(val)*32+2) // 8: assume complex need 32 runes
// 	b = append(b, []byte("[")...)
// 	for i := range val {
// 		if i > 0 {
// 			b = append(b, []byte(",")...)
// 		}
// 		b=strconv.AppendComplex(b, complex128(val[i]), 'f', -1, 128)
// 	}
// 	b = append(b, []byte("]")...)
// 	return string(b)
// }

func stringSliceToString(val []string) string {
	b := make([]byte, 0, len(val)*32+2) // 8: assume integer need 32 runes
	b = append(b, []byte("[")...)
	for i := range val {
		if i > 0 {
			b = append(b, []byte(",")...)
		}
		b = strconv.AppendQuote(b, val[i])
	}
	b = append(b, []byte("]")...)
	return string(b)
}

func boolSliceToString(val []bool) string {
	b := make([]byte, 0, len(val)*8) // 8: assume bool need 5 runes
	b = append(b, []byte("[")...)
	for i := range val {
		if i > 0 {
			b = append(b, []byte(",")...)
		}
		b = strconv.AppendBool(b, val[i])
	}
	b = append(b, []byte("]")...)
	return string(b)
}

func timeSliceToString(val []time.Time) string {
	b := make([]byte, 0, len(val)*32+2) // 8: assume time need 24 runes
	b = append(b, []byte("[")...)
	for i := range val {
		if i > 0 {
			b = append(b, []byte(",")...)
		}
		b = strconv.AppendQuote(b, val[i].Format(time.RFC3339Nano))
	}
	b = append(b, []byte("]")...)
	return string(b)
}

func durationSliceToString(val []time.Duration) string {
	b := make([]byte, 0, len(val)*16+2) // 8: assume duration need 16 runes
	b = append(b, []byte("[")...)
	for i := range val {
		if i > 0 {
			b = append(b, []byte(",")...)
		}
		b = strconv.AppendQuote(b, val[i].String())
	}
	b = append(b, []byte("]")...)
	return string(b)
}

//

type Stringer interface {
	String() string
}

type ToString interface {
	ToString(args ...any) string
}

type Integers interface {
	int | int8 | int16 | int32 | int64
}

type Uintegers interface {
	uint | uint8 | uint16 | uint32 | uint64
}

type Floats interface {
	float32 | float64
}

type Complexes interface {
	complex64 | complex128
}

type Numerics interface {
	Integers | Uintegers | Floats | Complexes
}

type IntSlice[T Integers] []T

type UintSlice[T Uintegers] []T

type FloatSlice[T Floats] []T

type ComplexSlice[T Complexes] []T

type StringSlice[T string] []T

type BoolSlice[T bool] []T

type Slice[T Integers | Uintegers | Floats] []T

// const (
// 	TimeNoNano      = "15:04:05Z07:00"
// 	TimeNano        = "15:04:05.000000Z07:00"
// 	DateTime        = "2006-01-0215:04:05Z07:00"
// 	RFC3339Nano     = "2006-01-02T15:04:05.000000Z07:00"
// 	RFC3339NanoOrig = "2006-01-02T15:04:05.999999999Z07:00"
// )
