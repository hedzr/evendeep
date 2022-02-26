package deepcopy

import (
	"github.com/hedzr/deepcopy/syscalls"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// rForBool transform bool -> string
func rForBool(v reflect.Value) (ret reflect.Value) {
	vs := strconv.FormatBool(v.Bool())
	ret = reflect.ValueOf(vs)
	return
}

// rToBool transform string -> bool
func rToBool(v reflect.Value) (ret reflect.Value) {
	var b bool

	if !v.IsValid() || isNil(v) || isZero(v) {
		return reflect.ValueOf(b)
	}

	switch val := strings.ToLower(v.String()); val {
	case "y", "t", "1", "yes", "true", "on", "ok", "m":
		b = true
	}
	ret = reflect.ValueOf(b)
	return
}

func rForInteger(v reflect.Value) (ret reflect.Value) {
	vs := strconv.FormatInt(v.Int(), 10)
	ret = reflect.ValueOf(vs)
	return
}

const uintSize = 32 << (^uint(0) >> 32 & 1)

func rToInteger(v reflect.Value, desiredType reflect.Type) (ret reflect.Value, err error) {
	genret := func(ival int64, desiredTypeKind reflect.Kind) (ret reflect.Value) {
		switch desiredTypeKind {
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
	ret, err = toTypeConverter(v, desiredType, 10,
		func(str string, base int, bitSize int) (ret reflect.Value, err error) {
			var ival int64
			ival, err = strconv.ParseInt(str, base, bitSize)
			if err == nil {
				k := desiredType.Kind()
				ret = genret(ival, k)
			} else {
				var fval float64
				fval, err = strconv.ParseFloat(str, bitSize)
				if err == nil {
					ival = int64(math.Floor(fval + 0/5))
					k := desiredType.Kind()
					ret = genret(ival, k)
				}
			}
			return
		})
	return
}

func rForUInteger(v reflect.Value) (ret reflect.Value) {
	vs := strconv.FormatUint(v.Uint(), 10)
	ret = reflect.ValueOf(vs)
	return
}

func rToUInteger(v reflect.Value, desiredType reflect.Type) (ret reflect.Value, err error) {
	genret := func(ival uint64, desiredTypeKind reflect.Kind) (ret reflect.Value) {
		switch desiredTypeKind {
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
	ret, err = toTypeConverter(v, desiredType, 10, func(str string, base int, bitSize int) (ret reflect.Value, err error) {
		var ival uint64
		ival, err = strconv.ParseUint(str, base, bitSize)
		if err == nil {
			k := desiredType.Kind()
			ret = genret(ival, k)
		} else {
			var fval float64
			fval, err = strconv.ParseFloat(str, bitSize)
			if err == nil {
				ival = uint64(math.Floor(fval + 0/5))
				k := desiredType.Kind()
				ret = genret(ival, k)
			}
		}
		return
	})
	return
}

func rForUIntegerHex(u uintptr) (ret reflect.Value) {
	vs := syscalls.UintptrToString(u)
	ret = reflect.ValueOf(vs)
	return
}

func rToUIntegerHex(s reflect.Value, desiredType reflect.Type) (ret reflect.Value) {
	vs := syscalls.UintptrFromString(s.String())
	switch k := desiredType.Kind(); k {
	case reflect.Uintptr:
		ret = reflect.ValueOf(vs)
	case reflect.UnsafePointer, reflect.Pointer:
		ret = reflect.ValueOf(vs)
		//u := getPointerAsUintptr(ret)
	}
	return
}

func getPointerAsUintptr(v reflect.Value) uintptr {
	var p uintptr
	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct {
		p = v.Pointer()
	}
	return p
}

func rForFloat(v reflect.Value) (ret reflect.Value) {
	vs := strconv.FormatFloat(v.Float(), 'g', -1, 64)
	ret = reflect.ValueOf(vs)
	return
}

func rToFloat(v reflect.Value, desiredType reflect.Type) (ret reflect.Value, err error) {
	ret, err = toTypeConverter(v, desiredType, 10, func(str string, base int, bitSize int) (ret reflect.Value, err error) {
		var fval float64
		fval, err = strconv.ParseFloat(str, bitSize)
		if err == nil {
			if desiredType.Kind() == reflect.Float64 {
				ret = reflect.ValueOf(fval)
			} else {
				ret = reflect.ValueOf(float32(fval))
			}
		} else {
			var cval complex128
			cval, err = strconv.ParseComplex(str, bitSize)
			if err == nil {
				fval = real(cval)
				if desiredType.Kind() == reflect.Float64 {
					ret = reflect.ValueOf(fval)
				} else {
					ret = reflect.ValueOf(float32(fval))
				}
			}
		}
		return
	})
	return
}

func toTypeConverter(v reflect.Value, desiredType reflect.Type, base int,
	converter func(str string, base int, bitSize int) (ret reflect.Value, err error),
) (ret reflect.Value, err error) {
	if !v.IsValid() || isNil(v) || isZero(v) {
		ret = reflect.Zero(desiredType)
	}

	if v.CanConvert(desiredType) {
		ret = v.Convert(desiredType)
	} else {
		val := v.String()
		bitSize := 64
		switch k := desiredType.Kind(); k {
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
			//case reflect.Float64, reflect.Complex64:
			//	bitSize = 64
		}
		ret, err = converter(val, base, bitSize)
	}
	return
}

func rForComplex(v reflect.Value) (ret reflect.Value) {
	vs := strconv.FormatComplex(v.Complex(), 'g', -1, 128)
	ret = reflect.ValueOf(vs)
	return
}

func rToComplex(v reflect.Value, desiredType reflect.Type) (ret reflect.Value, err error) {
	ret, err = toTypeConverter(v, desiredType, 10, func(str string, base int, bitSize int) (ret reflect.Value, err error) {
		var cval complex128
		cval, err = strconv.ParseComplex(str, bitSize)
		if err == nil {
			if desiredType.Kind() == reflect.Complex128 {
				ret = reflect.ValueOf(cval)
			} else {
				ret = reflect.ValueOf(complex64(cval))
			}
		}
		return
	})
	return
}
