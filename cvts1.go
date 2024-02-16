package evendeep

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/hedzr/evendeep/ref"
	logz "github.com/hedzr/logg/slog"
)

func parseToSlice[T any](in string) (out []T) {
	ins := strings.TrimSpace(in)
	if ins[0] == '[' && ins[len(in)-1] == ']' {
		a := strings.Split(ins[:len(ins)-1], ",")
		for _, it := range a {
			out = append(out, parseToAny[T](it))
		}
	} else {
		out = append(out, parseToAny[T](ins))
	}
	return
}

func parseToAny[T any](in string) (out T) {
	var t1 any = &out
	switch z := t1.(type) {
	case *bool:
		*z = anyToBool(in)
	case *string:
		*z = anyToString(in)

	case *int:
		*z = int(anyToInt(in))
	case *int64:
		*z = anyToInt(in)
	case *int32:
		*z = int32(anyToInt(in))
	case *int16:
		*z = int16(anyToInt(in))
	case *int8:
		*z = int8(anyToInt(in))
	case *uint:
		*z = uint(anyToUint(in))
	case *uint64:
		*z = anyToUint(in)
	case *uint32:
		*z = uint32(anyToUint(in))
	case *uint16:
		*z = uint16(anyToUint(in))
	case *uint8:
		*z = uint8(anyToUint(in))
	case *float64:
		*z = anyToFloat[float64](in)
	case *float32:
		*z = anyToFloat[float32](in)
	case *complex128:
		*z = anyToComplex[complex128](in)
	case *complex64:
		*z = anyToComplex[complex64](in)

	case *[]int:
		*z = anyToIntSliceT[int](in)
	case *[]int64:
		*z = anyToIntSliceT[int64](in)
	case *[]int32:
		*z = anyToIntSliceT[int32](in)
	case *[]int16:
		*z = anyToIntSliceT[int16](in)
	case *[]int8:
		*z = anyToIntSliceT[int8](in)
	case *[]uint:
		*z = anyToUintSliceT[uint](in)
	case *[]uint64:
		*z = anyToUintSliceT[uint64](in)
	case *[]uint32:
		*z = anyToUintSliceT[uint32](in)
	case *[]uint16:
		*z = anyToUintSliceT[uint16](in)
	case *[]uint8:
		*z = anyToUintSliceT[uint8](in)
	case *[]float64:
		*z = anyToFloatSlice[float64](in)
	case *[]float32:
		*z = anyToFloatSlice[float32](in)
	case *[]complex128:
		*z = anyToComplexSlice[complex128](in)
	case *[]complex64:
		*z = anyToComplexSlice[complex64](in)
	}
	return
}

func anyToIntSliceT[T Integers](data any) (ret []T) {
	if data == nil {
		return
	}

	switch z := data.(type) {
	case []float64:
		return zfToIntT[float64, T](z)
	case []float32:
		return zfToIntT[float32, T](z)

	case []int:
		return zfToIntT[int, T](z)
	case []int64:
		return zfToIntT[int64, T](z)
	case []int32:
		return zfToIntT[int32, T](z)
	case []int16:
		return zfToIntT[int16, T](z)
	case []int8:
		return zfToIntT[int8, T](z)
	case []uint:
		return zfToIntT[uint, T](z)
	case []uint64:
		return zfToIntT[uint64, T](z)
	case []uint32:
		return zfToIntT[uint32, T](z)
	case []uint16:
		return zfToIntT[uint16, T](z)
	case []uint8:
		return zfToIntT[uint8, T](z)

	case []bool:
		return zfToIntT[bool, T](z)
	case []string:
		return zfToIntT[string, T](z)

	// case []fmt.Stringer:
	// 	return zfToIntT[string,T](z)
	default:
		break
	}
	return
}

func anyToUintSliceT[T Uintegers](data any) (ret []T) {
	if data == nil {
		return
	}

	switch z := data.(type) {
	case []float64:
		return zfToUintT[float64, T](z)
	case []float32:
		return zfToUintT[float32, T](z)

	case []int:
		return zfToUintT[int, T](z)
	case []int64:
		return zfToUintT[int64, T](z)
	case []int32:
		return zfToUintT[int32, T](z)
	case []int16:
		return zfToUintT[int16, T](z)
	case []int8:
		return zfToUintT[int8, T](z)
	case []uint:
		return zfToUintT[uint, T](z)
	case []uint64:
		return zfToUintT[uint64, T](z)
	case []uint32:
		return zfToUintT[uint32, T](z)
	case []uint16:
		return zfToUintT[uint16, T](z)
	case []uint8:
		return zfToUintT[uint8, T](z)

	case []bool:
		return zfToUintT[bool, T](z)
	case []string:
		return zfToUintT[string, T](z)

	// case []fmt.Stringer:
	// 	return zfToIntT[string,T](z)
	default:
		break
	}
	return
}

func zfToIntT[In any, Out Integers](in []In) (out []Out) {
	out = make([]Out, 0, len(in))
	for _, it := range in {
		out = append(out, Out(anyToInt(it)))
	}
	return
}

func zfToUintT[In any, Out Uintegers](in []In) (out []Out) {
	out = make([]Out, 0, len(in))
	for _, it := range in {
		out = append(out, Out(anyToUint(it)))
	}
	return
}

func anyToIntT[In any, Out Integers](in In) (out Out) {
	out = Out(anyToInt(in))
	return
}

func anyToUintT[In any, Out Uintegers](in In) (out Out) {
	out = Out(anyToUint(in))
	return
}

func anyToInt(data any) int64 {
	if data == nil {
		return 0
	}

	switch z := data.(type) {
	case int:
		return int64(z)
	case int8:
		return int64(z)
	case int16:
		return int64(z)
	case int32:
		return int64(z)
	case int64:
		return z

	case uint:
		return int64(z)
	case uint8:
		return int64(z)
	case uint16:
		return int64(z)
	case uint32:
		return int64(z)
	case uint64:
		if z <= uint64(math.MaxInt64) {
			return int64(z)
		}
		break

	case float32:
		return int64(z)
	case float64:
		return int64(z)

	case complex64:
		return int64(real(z))
	case complex128:
		return int64(real(z))

	case string:
		return atoi(z)

	case time.Duration:
		return int64(z)
	case time.Time:
		return z.UnixNano()

	default:
		return atoi(fmt.Sprint(data))
	}

	// reflect approach
	rv := reflect.ValueOf(data)
	logz.Warn("[anyToInt]: unrecognized data type",
		"typ", ref.Typfmtv(&rv),
		"val", ref.Valfmt(&rv),
	)
	return 0
}

func anyToUint(data any) uint64 {
	if data == nil {
		return 0
	}

	switch z := data.(type) {
	case int:
		return uint64(z)
	case int8:
		return uint64(z)
	case int16:
		return uint64(z)
	case int32:
		return uint64(z)
	case int64:
		return uint64(z)

	case uint:
		return uint64(z)
	case uint8:
		return uint64(z)
	case uint16:
		return uint64(z)
	case uint32:
		return uint64(z)
	case uint64:
		return z

	case float32:
		return uint64(z)
	case float64:
		return uint64(z)

	case complex64:
		return uint64(real(z))
	case complex128:
		return uint64(real(z))

	case string:
		return atou(z)

	case time.Duration:
		return uint64(z)
	case time.Time:
		return uint64(z.UnixNano())

	default:
		return atou(fmt.Sprint(data))
	}
}

func atoi(v string) int64 {
	if i, err := strconv.ParseInt(v, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(v, 64); err == nil {
		return int64(f)
	}
	if u, err := strconv.ParseUint(v, 10, 64); err != nil {
		return int64(u)
	}
	return 0
}

func atou(v string) uint64 {
	if u, err := strconv.ParseUint(v, 10, 64); err != nil {
		return u
	}
	if i, err := strconv.ParseInt(v, 10, 64); err == nil {
		return uint64(i)
	}
	if f, err := strconv.ParseFloat(v, 64); err == nil {
		return uint64(f)
	}
	return 0
}

//

//

func zfToInt64MNT[T any, Out Integers](in map[string]T) (out map[string]Out) {
	out = make(map[string]Out, len(in))
	for k, it := range in {
		out[k] = anyToIntT[T, Out](it)
	}
	return
}

func zfToUint64MNT[T any, Out Uintegers](in map[string]T) (out map[string]Out) {
	out = make(map[string]Out, len(in))
	for k, it := range in {
		out[k] = anyToUintT[T, Out](it)
	}
	return
}

func anyToInt64MapT[Out Integers](data any) (ret map[string]Out) {
	if data == nil {
		return
	}

	switch z := data.(type) {
	case map[string]string:
		return zfToInt64MNT[string, Out](z)
	case map[string]bool:
		return zfToInt64MNT[bool, Out](z)

	case map[string]int:
		return zfToInt64MNT[int, Out](z)
	case map[string]int64:
		return zfToInt64MNT[int64, Out](z)
	case map[string]int32:
		return zfToInt64MNT[int32, Out](z)
	case map[string]int16:
		return zfToInt64MNT[int16, Out](z)
	case map[string]int8:
		return zfToInt64MNT[int8, Out](z)
	case map[string]uint:
		return zfToInt64MNT[uint, Out](z)
	case map[string]uint64:
		return zfToInt64MNT[uint64, Out](z)
	case map[string]uint32:
		return zfToInt64MNT[uint32, Out](z)
	case map[string]uint16:
		return zfToInt64MNT[uint16, Out](z)
	case map[string]uint8:
		return zfToInt64MNT[uint8, Out](z)
	case map[string]float64:
		return zfToInt64MNT[float64, Out](z)
	case map[string]float32:
		return zfToInt64MNT[float32, Out](z)
	case map[string]complex128:
		return zfToInt64MNT[complex128, Out](z)
	case map[string]complex64:
		return zfToInt64MNT[complex64, Out](z)

	case map[string]any:
		return zfToInt64MNT[any, Out](z)
	default:
		break
	}
	return
}

func anyToUint64MapT[Out Uintegers](data any) (ret map[string]Out) {
	if data == nil {
		return
	}

	switch z := data.(type) {
	case map[string]string:
		return zfToUint64MNT[string, Out](z)
	case map[string]bool:
		return zfToUint64MNT[bool, Out](z)

	case map[string]int:
		return zfToUint64MNT[int, Out](z)
	case map[string]int64:
		return zfToUint64MNT[int64, Out](z)
	case map[string]int32:
		return zfToUint64MNT[int32, Out](z)
	case map[string]int16:
		return zfToUint64MNT[int16, Out](z)
	case map[string]int8:
		return zfToUint64MNT[int8, Out](z)
	case map[string]uint:
		return zfToUint64MNT[uint, Out](z)
	case map[string]uint64:
		return zfToUint64MNT[uint64, Out](z)
	case map[string]uint32:
		return zfToUint64MNT[uint32, Out](z)
	case map[string]uint16:
		return zfToUint64MNT[uint16, Out](z)
	case map[string]uint8:
		return zfToUint64MNT[uint8, Out](z)
	case map[string]float64:
		return zfToUint64MNT[float64, Out](z)
	case map[string]float32:
		return zfToUint64MNT[float32, Out](z)
	case map[string]complex128:
		return zfToUint64MNT[complex128, Out](z)
	case map[string]complex64:
		return zfToUint64MNT[complex64, Out](z)

	case map[string]any:
		return zfToUint64MNT[any, Out](z)
	default:
		break
	}
	return
}

//

//

func anyToFloat[R Floats](data any) R {
	if data == nil {
		return 0
	}

	switch z := data.(type) {
	case float64:
		return R(z)
	case float32:
		return R(z)

	case int:
		return R(z)
	case int64:
		return R(z)
	case int32:
		return R(z)
	case int16:
		return R(z)
	case int8:
		return R(z)
	case uint:
		return R(z)
	case uint64:
		return R(z)
	case uint32:
		return R(z)
	case uint16:
		return R(z)
	case uint8:
		return R(z)

	case string:
		return R(mustParseFloat(z))
	case fmt.Stringer:
		return R(mustParseFloat(z.String()))

	default:
		str := fmt.Sprintf("%v", data)
		return R(mustParseFloat(str))
	}
}

func mustParseFloat(s string) (ret float64) {
	ret, _ = strconv.ParseFloat(s, 64)
	return
}

//

func zfToFloatS[T Floats, R Floats](in []T) (out []R) {
	out = make([]R, 0, len(in))
	for _, it := range in {
		out = append(out, R(it))
	}
	return
}

func anyToFloatSlice[R Floats](data any) (ret []R) {
	if data == nil {
		return
	}

	switch z := data.(type) {
	case []float64:
		return zfToFloatS[float64, R](z)
	case []float32:
		return zfToFloatS[float32, R](z)

	case []int:
		return zsToFloatS[int, R](z)
	case []int64:
		return zsToFloatS[int64, R](z)
	case []int32:
		return zsToFloatS[int32, R](z)
	case []int16:
		return zsToFloatS[int16, R](z)
	case []int8:
		return zsToFloatS[int8, R](z)
	case []uint:
		return zsToFloatS[uint, R](z)
	case []uint64:
		return zsToFloatS[uint64, R](z)
	case []uint32:
		return zsToFloatS[uint32, R](z)
	case []uint16:
		return zsToFloatS[uint16, R](z)
	case []uint8:
		return zsToFloatS[uint8, R](z)

	case []string:
		ret = make([]R, 0, len(z))
		for _, it := range z {
			ret = append(ret, R(mustParseFloat(it)))
		}
		return
	case []fmt.Stringer:
		ret = make([]R, 0, len(z))
		for _, it := range z {
			ret = append(ret, R(mustParseFloat(it.String())))
		}
		return

	default:
		break
	}
	return
}

func zsToFloatS[T Integers | Uintegers, R Floats](z []T) (ret []R) {
	ret = make([]R, 0, len(z))
	for _, it := range z {
		ret = append(ret, R(int64(it)))
	}
	return
}

//

//

func anyToFloat64Map(data any) (ret map[string]float64) {
	if data == nil {
		return
	}

	switch z := data.(type) {
	case map[string]any:
		return zfToFloat64M(z)

	case map[string]bool:
		return zfToFloat64MN(z)
	case map[string]string:
		return zfToFloat64MN(z)

	case map[string]complex128:
		return zfToFloat64MN(z)
	case map[string]complex64:
		return zfToFloat64MN(z)
	case map[string]float64:
		return z // zfToFloat64MN(z)
	case map[string]float32:
		return zfToFloat64MN(z)

	case map[string]int64:
		return zfToFloat64MN(z)
	case map[string]int32:
		return zfToFloat64MN(z)
	case map[string]int16:
		return zfToFloat64MN(z)
	case map[string]int8:
		return zfToFloat64MN(z)
	case map[string]int:
		return zfToFloat64MN(z)
	case map[string]uint64:
		return zfToFloat64MN(z)
	case map[string]uint32:
		return zfToFloat64MN(z)
	case map[string]uint16:
		return zfToFloat64MN(z)
	case map[string]uint8:
		return zfToFloat64MN(z)
	case map[string]uint:
		return zfToFloat64MN(z)

	default:
		break
	}
	return
}

func anyToFloat32Map(data any) (ret map[string]float32) {
	if data == nil {
		return
	}

	switch z := data.(type) {
	case map[string]any:
		return zfToFloat32M(z)

	case map[string]bool:
		return zfToFloat32MN(z)
	case map[string]string:
		return zfToFloat32MN(z)

	case map[string]complex128:
		return zfToFloat32MN(z)
	case map[string]complex64:
		return zfToFloat32MN(z)
	case map[string]float64:
		return zfToFloat32MN(z)
	case map[string]float32:
		return z // zfToFloat32MN(z)

	case map[string]int64:
		return zfToFloat32MN(z)
	case map[string]int32:
		return zfToFloat32MN(z)
	case map[string]int16:
		return zfToFloat32MN(z)
	case map[string]int8:
		return zfToFloat32MN(z)
	case map[string]int:
		return zfToFloat32MN(z)
	case map[string]uint64:
		return zfToFloat32MN(z)
	case map[string]uint32:
		return zfToFloat32MN(z)
	case map[string]uint16:
		return zfToFloat32MN(z)
	case map[string]uint8:
		return zfToFloat32MN(z)
	case map[string]uint:
		return zfToFloat32MN(z)

	default:
		break
	}
	return
}

func zfToFloat64M(in map[string]any) (out map[string]float64) {
	out = make(map[string]float64, len(in))
	for k, it := range in {
		out[k] = anyToFloat[float64](it)
	}
	return
}

func zfToFloat64MN[T Numerics | string | bool](in map[string]T) (out map[string]float64) {
	out = make(map[string]float64, len(in))
	for k, it := range in {
		out[k] = anyToFloat[float64](it)
	}
	return
}

func zfToFloat32M(in map[string]any) (out map[string]float32) {
	out = make(map[string]float32, len(in))
	for k, it := range in {
		out[k] = anyToFloat[float32](it)
	}
	return
}

func zfToFloat32MN[T Numerics | string | bool](in map[string]T) (out map[string]float32) {
	out = make(map[string]float32, len(in))
	for k, it := range in {
		out[k] = anyToFloat[float32](it)
	}
	return
}
