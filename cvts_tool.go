package deepcopy

import (
	"fmt"
	"github.com/hedzr/deepcopy/cl"
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

// rToBool transform string (or anything) -> bool
func rToBool(v reflect.Value) (ret reflect.Value) {
	var b bool

	if !v.IsValid() || isNil(v) || isZero(v) {
		return reflect.ValueOf(b)
	}

	k := v.Kind()
	switch k {
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
		b = !arrayIsZero(v)
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		b = !isNil(v)
	case reflect.Struct:
		b = !structIsZero(v)
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

// rForInteger transform integer -> string
func rForInteger(v reflect.Value) (ret reflect.Value) {
	int64typ := reflect.TypeOf((*int64)(nil)).Elem()
	if !v.IsValid() {
		v = reflect.Zero(int64typ)
	} else if k := v.Kind(); k < reflect.Int || k > reflect.Int64 {
		if canConvert(&v, int64typ) {
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

// rForUInteger transform uint64 -> string
func rForUInteger(v reflect.Value) (ret reflect.Value) {
	uint64typ := reflect.TypeOf((*uint64)(nil)).Elem()
	if !v.IsValid() {
		v = reflect.Zero(uint64typ)
	} else if k := v.Kind(); k < reflect.Uint || k > reflect.Uintptr {
		if canConvert(&v, uint64typ) {
			v = v.Convert(uint64typ)
		} else {
			v = reflect.Zero(uint64typ)
		}
	}

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

// rForUIntegerHex transform uintptr/... -> string
func rForUIntegerHex(u uintptr) (ret reflect.Value) {
	vs := syscalls.UintptrToString(u)
	ret = reflect.ValueOf(vs)
	return
}

func rToUIntegerHex(s reflect.Value, desiredType reflect.Type) (ret reflect.Value) {
	vs := syscalls.UintptrFromString(s.String())
	fmt.Printf("vs : %v, k: %v\n", vs, desiredType.Kind())
	switch k := desiredType.Kind(); k {
	case reflect.Uintptr:
		ret = reflect.ValueOf(vs)
	case reflect.UnsafePointer, reflect.Ptr:
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

// rForFloat transform float -> string
func rForFloat(v reflect.Value) (ret reflect.Value) {
	float64typ := reflect.TypeOf((*float64)(nil)).Elem()
	if !v.IsValid() {
		v = reflect.Zero(float64typ)
	} else if k := v.Kind(); k < reflect.Float32 || k > reflect.Float64 {
		if canConvert(&v, float64typ) {
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
	ret, err = toTypeConverter(v, desiredType, 10,
		func(str string, base int, bitSize int) (ret reflect.Value, err error) {
			var fval float64
			fval, err = strconv.ParseFloat(str, bitSize)
			if err == nil {
				ret = toFloat(fval, desiredType)
			} else {
				var ival int64
				ival, err = strconv.ParseInt(str, 10, bitSize)
				if err == nil {
					ret = toFloat(float64(ival), desiredType)
				} else {
					var uval uint64
					uval, err = strconv.ParseUint(str, 10, bitSize)
					if err == nil {
						ret = toFloat(float64(uval), desiredType)
					} else {
						var cval complex128
						cval, err = cl.ParseComplex(str)
						if err == nil {
							fval = real(cval)
							ret = toFloat(fval, desiredType)
						}
					}
				}
			}
			return
		})
	return
}

// rForComplex transform complex -> string
func rForComplex(v reflect.Value) (ret reflect.Value) {
	complex128typ := reflect.TypeOf((*complex128)(nil)).Elem()
	if !v.IsValid() {
		v = reflect.Zero(complex128typ)
	} else if k := v.Kind(); k < reflect.Complex64 || k > reflect.Complex128 {
		//if canConvert(&v, complex128typ) {
		//	v = v.Convert(complex128typ)
		//} else {
		switch k {
		case reflect.Float64, reflect.Float32:
			v = reflect.ValueOf(complex(v.Float(), 0.0))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v = reflect.ValueOf(complex(float64(v.Int()), 0.0))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			v = reflect.ValueOf(complex(float64(v.Uint()), 0.0))
		default:
			v = reflect.Zero(complex128typ)
		}
		//}
	}

	vs := cl.FormatComplex(v.Complex(), 'g', -1, 128)
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
	ret, err = toTypeConverter(v, desiredType, 10, func(str string, base int, bitSize int) (ret reflect.Value, err error) {
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

func toTypeConverter(v reflect.Value, desiredType reflect.Type, base int,
	converter func(str string, base int, bitSize int) (ret reflect.Value, err error),
) (ret reflect.Value, err error) {
	if !v.IsValid() || isNil(v) || isZero(v) {
		ret = reflect.Zero(desiredType)
	}

	if canConvert(&v, desiredType) {
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
