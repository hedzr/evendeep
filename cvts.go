package evendeep

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"gopkg.in/hedzr/errors.v3"

	"github.com/hedzr/evendeep/dbglog"
	"github.com/hedzr/evendeep/flags"
	"github.com/hedzr/evendeep/flags/cms"
	"github.com/hedzr/evendeep/internal/syscalls"
	"github.com/hedzr/evendeep/internal/times"
	"github.com/hedzr/evendeep/ref"
	"github.com/hedzr/evendeep/typ"
	logz "github.com/hedzr/logg/slog"
)

const timeConstString = "time"

// RegisterDefaultConverters registers the ValueConverter list into
// default converters registry.
//
// It takes effects on DefaultCopyController, MakeClone, DeepCopy,
// and New, ....
func RegisterDefaultConverters(ss ...ValueConverter) {
	defValueConverters = append(defValueConverters, ss...)
	lenValueConverters, lenValueCopiers = len(defValueConverters), len(defValueCopiers)
	initGlobalOperators()
}

// RegisterDefaultCopiers registers the ValueCopier list into
// default copiers registry.
//
// It takes effects on DefaultCopyController, MakeClone, DeepCopy,
// and New, ....
func RegisterDefaultCopiers(ss ...ValueCopier) {
	defValueCopiers = append(defValueCopiers, ss...)
	lenValueConverters, lenValueCopiers = len(defValueConverters), len(defValueCopiers)
	initGlobalOperators()
}

func initConverters() {
	dbglog.Log("initializing default converters and copiers ...")
	defValueConverters = ValueConverters{ // Transform()
		&fromStringConverter{}, // the final choice here
		&toStringConverter{},

		// &toFuncConverter{},
		&fromFuncConverter{},

		&toDurationConverter{},
		&fromDurationConverter{},
		&toTimeConverter{},
		&fromTimeConverter{},

		&fromBytesBufferConverter{},
		&fromSyncPkgConverter{},
		&fromMapConverter{},
	}
	defValueCopiers = ValueCopiers{ // CopyTo()
		&fromStringConverter{}, // the final choice here
		&toStringConverter{},

		&toFuncConverter{},
		&fromFuncConverter{},

		&toDurationConverter{},
		&fromDurationConverter{},
		&toTimeConverter{},
		&fromTimeConverter{},

		&fromBytesBufferConverter{},
		&fromSyncPkgConverter{},
		&fromMapConverter{},
	}

	lenValueConverters, lenValueCopiers = len(defValueConverters), len(defValueCopiers)
}

var (
	defValueConverters                  ValueConverters //nolint:gochecknoglobals //i know that
	defValueCopiers                     ValueCopiers    //nolint:gochecknoglobals //i know that
	lenValueConverters, lenValueCopiers int             //nolint:gochecknoglobals //i know that
)

func defaultValueConverters() ValueConverters { return defValueConverters }
func defaultValueCopiers() ValueCopiers       { return defValueCopiers }

// ValueConverter for internal used.
type ValueConverter interface {
	Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error)
	Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool)
}

// ValueCopier for internal used.
type ValueCopier interface {
	CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error)
	Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool)
}

// NameConverter for internal used.
type NameConverter interface {
	ToGoName(ctx *NameConverterContext, fieldName string) (goName string)
	ToFieldName(ctx *NameConverterContext, goName string) (fieldName string)
}

// ValueConverters for internal used.
type ValueConverters []ValueConverter

// ValueCopiers for internal used.
type ValueCopiers []ValueCopier

// NameConverters for internal used.
type NameConverters []NameConverter

// NameConverterContext for internal used.
type NameConverterContext struct {
	*Params
}

// ValueConverterContext for internal used.
type ValueConverterContext struct {
	*Params
}

//

type CvtV struct {
	Data any
}

func (s *CvtV) String() string {
	return anyToString(s.Data)
}

//

type Cvt struct{}

func (s *Cvt) String(data any) string               { return anyToString(data) }
func (s *Cvt) StringSlice(data any) []string        { return anyToStringSlice(data) }
func (s *Cvt) StringMap(data any) map[string]string { return anyToStringMap(data) }

func anyToStringSlice(data any) (ret []string) {
	if data == nil {
		return
	}

	switch z := data.(type) {
	case []string:
		return z

	case []float64:
		return zfToStringS(z)
	case []float32:
		return zfToStringS(z)

	case []int:
		return zfToStringS(z)
	case []int64:
		return zfToStringS(z)
	case []int32:
		return zfToStringS(z)
	case []int16:
		return zfToStringS(z)
	case []int8:
		return zfToStringS(z)
	case []uint:
		return zfToStringS(z)
	case []uint64:
		return zfToStringS(z)
	case []uint32:
		return zfToStringS(z)
	case []uint16:
		return zfToStringS(z)
	case []uint8:
		return zfToStringS(z)

	case []bool:
		return zfToStringS(z)
	case []fmt.Stringer:
		return zfToStringS(z)

	case []any:
		return zfToStringS(z)

	case string:
		return parseToSlice[string](z)
	case fmt.Stringer:
		return parseToSlice[string](z.String())
	case any:
		return parseToSlice[string](anyToString(z))

	default:
		break
	}
	return
}

func zfToStringS[T any](in []T) (out []string) {
	out = make([]string, 0, len(in))
	for _, it := range in {
		out = append(out, anyToString(it))
	}
	return
}

func zfToStringM(in map[string]any) (out map[string]string) {
	out = make(map[string]string, len(in))
	for k, it := range in {
		out[k] = anyToString(it)
	}
	return
}

func zfToStringMA(in map[any]any) (out map[string]string) {
	out = make(map[string]string, len(in))
	for k, it := range in {
		out[anyToString(k)] = anyToString(it)
	}
	return
}

func anyToStringMap(data any) (ret map[string]string) {
	if data == nil {
		return
	}

	switch z := data.(type) {
	case map[string]string:
		return z
	case map[string]any:
		return zfToStringM(z)
	case map[any]any:
		return zfToStringMA(z)

	case string:
		return parseToMap[string](z)
	case fmt.Stringer:
		return parseToMap[string](z.String())
	case any:
		return parseToMap[string](anyToString(z))

	default:
		break
	}
	return
}

//

func (s *Cvt) Bool(data any) bool { return anyToBool(data) }

func anyToBool(data any) bool {
	if data == nil {
		return false
	}

	switch z := data.(type) {
	case bool:
		return z

	case string:
		return toBool(z)
	case fmt.Stringer:
		return toBool(z.String())
	case any:
		return toBool(anyToString(z))

	default:
		rv := reflect.ValueOf(data)
		rv = ref.Rindirect(rv)
		if k := rv.Kind(); k == reflect.Bool {
			// eg: flag.stringValue (string)
			return rv.Bool()
		}
		return toBool(anyToString(data))
	}
}

func toBool(s string) bool {
	_, ok := stringToBoolMap[strings.ToLower(s)]
	if ok {
		return ok
	}
	u, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		i, err1 := strconv.ParseInt(s, 10, 64)
		if err1 != nil {
			f, err2 := strconv.ParseFloat(s, 64)
			if err2 != nil {
				return false
			}
			return f != 0
		}
		return i != 0
	}
	return u > 0
}

var stringToBoolMap = map[string]struct{}{
	"1":     {},
	"t":     {},
	"male":  {},
	"y":     {},
	"yes":   {},
	"true":  {},
	"ok":    {},
	"allow": {},
	"on":    {},
	"open":  {},
}

func (s *Cvt) BoolSlice(data any) []bool { return anyToBoolSlice(data) }

func anyToBoolSlice(data any) (ret []bool) {
	if data == nil {
		return
	}

	switch z := data.(type) {
	case []bool:
		return z

	case []float64:
		return zfToBoolS(z)
	case []float32:
		return zfToBoolS(z)

	case []int:
		return zfToBoolS(z)
	case []int64:
		return zfToBoolS(z)
	case []int32:
		return zfToBoolS(z)
	case []int16:
		return zfToBoolS(z)
	case []int8:
		return zfToBoolS(z)
	case []uint:
		return zfToBoolS(z)
	case []uint64:
		return zfToBoolS(z)
	case []uint32:
		return zfToBoolS(z)
	case []uint16:
		return zfToBoolS(z)
	case []uint8:
		return zfToBoolS(z)

	case []string:
		return zfToBoolS(z)
	case []fmt.Stringer:
		return zfToBoolS(z)

	case []any:
		return zfToBoolS(z)

	case string:
		return zfToBoolSS(z)
	case fmt.Stringer:
		return zfToBoolSS(z.String())
	case any:
		return zfToBoolSS(anyToString(z))

	default:
		break
	}
	return
}

func zfToBoolSS(in string) (out []bool) {
	ins := strings.TrimSpace(in)
	if ins[0] == '[' {
		out = parseToSlice[bool](ins)
	} else {
		out = append(out, anyToBool(ins))
	}
	return
}

func zfToBoolS[T any](in []T) (out []bool) {
	out = make([]bool, 0, len(in))
	for _, it := range in {
		out = append(out, anyToBool(it))
	}
	return
}

func (s *Cvt) BoolMap(data any) map[string]bool { return anyToBoolMap(data) }

func zfToBoolM(in map[string]any) (out map[string]bool) {
	out = make(map[string]bool, len(in))
	for k, it := range in {
		out[k] = anyToBool(it)
	}
	return
}

func anyToBoolMap(data any) (ret map[string]bool) {
	if data == nil {
		return
	}

	switch z := data.(type) {
	case map[string]bool:
		return z
	case map[string]any:
		return zfToBoolM(z)

	case string:
		return parseToMap[bool](z)
	case fmt.Stringer:
		return parseToMap[bool](z.String())
	case any:
		return parseToMap[bool](anyToString(z))

	default:
		break
	}
	return
}

// func zfToBoolMM(in string) (out map[string]bool) {
// 	ins := strings.TrimSpace(in)
// 	if ins[0] == '{' {
// 		out = parseToMap[bool](ins)
// 	}
// 	return
// }

//

func (s *Cvt) Int(data any) int64 { return anyToInt(data) }

func (s *Cvt) Int64Slice(data any) []int64 { return anyToIntSliceT[int64](data) }
func (s *Cvt) Int32Slice(data any) []int32 { return anyToIntSliceT[int32](data) }
func (s *Cvt) Int16Slice(data any) []int16 { return anyToIntSliceT[int16](data) }
func (s *Cvt) Int8Slice(data any) []int8   { return anyToIntSliceT[int8](data) }
func (s *Cvt) IntSlice(data any) []int     { return anyToIntSliceT[int](data) }

func (s *Cvt) Int64Map(data any) map[string]int64 { return anyToInt64MapT[int64](data) }
func (s *Cvt) Int32Map(data any) map[string]int32 { return anyToInt64MapT[int32](data) }
func (s *Cvt) Int16Map(data any) map[string]int16 { return anyToInt64MapT[int16](data) }
func (s *Cvt) Int8Map(data any) map[string]int8   { return anyToInt64MapT[int8](data) }
func (s *Cvt) IntMap(data any) map[string]int     { return anyToInt64MapT[int](data) }

//

func (s *Cvt) Uint(data any) uint64 { return anyToUintT[any, uint64](data) }

func (s *Cvt) Uint64Slice(data any) []uint64 { return anyToUintSliceT[uint64](data) }
func (s *Cvt) Uint32Slice(data any) []uint32 { return anyToUintSliceT[uint32](data) }
func (s *Cvt) Uint16Slice(data any) []uint16 { return anyToUintSliceT[uint16](data) }
func (s *Cvt) Uint8Slice(data any) []uint8   { return anyToUintSliceT[uint8](data) }
func (s *Cvt) UintSlice(data any) []uint     { return anyToUintSliceT[uint](data) }

func (s *Cvt) Uint64Map(data any) map[string]uint64 { return anyToUint64MapT[uint64](data) }
func (s *Cvt) Uint32Map(data any) map[string]uint32 { return anyToUint64MapT[uint32](data) }
func (s *Cvt) Uint16Map(data any) map[string]uint16 { return anyToUint64MapT[uint16](data) }
func (s *Cvt) Uint8Map(data any) map[string]uint8   { return anyToUint64MapT[uint8](data) }
func (s *Cvt) UintMap(data any) map[string]uint     { return anyToUint64MapT[uint](data) }

//

//

func anyToString(data any) string {
	if data == nil {
		return "<nil>"
	}

	// rv := reflect.ValueOf(data)
	// if ref.IsZero(rv) {
	// 	return "<zero>"
	// }

	switch z := data.(type) {
	case string:
		return z

	case time.Duration:
		return durationToString(z)
	case time.Time:
		return timeToString(z)
	case []time.Time:
		return timeSliceToString(z)
	case []time.Duration:
		return durationSliceToString(z)

	case error:
		return z.Error()

	case fmt.Stringer:
		return z.String()

	case bool:
		return boolToString(z)

	case []byte:
		return bytesToString(z)

	case []string:
		return stringSliceToString(z)

	case []bool:
		return boolSliceToString(z)

	case []int:
		return intSliceToString(z)
	case []int8:
		return intSliceToString(z)
	case []int16:
		return intSliceToString(z)
	case []int32:
		return intSliceToString(z)
	case []int64:
		return intSliceToString(z)

	case int:
		return intToString(z)
	case int8:
		return intToString(z)
	case int16:
		return intToString(z)
	case int32:
		return intToString(z)
	case int64:
		return intToString(z)

	case []uint:
		return uintSliceToString(z)
	// case []uint8: // = []byte
	case []uint16:
		return uintSliceToString(z)
	case []uint32:
		return uintSliceToString(z)
	case []uint64:
		return uintSliceToString(z)

	case uint:
		return uintToString(z)
	case uint8:
		return uintToString(z)
	case uint16:
		return uintToString(z)
	case uint32:
		return uintToString(z)
	case uint64:
		return uintToString(z)

	case []float32:
		return floatSliceToString(z)
	case []float64:
		return floatSliceToString(z)

	case float32:
		return floatToString(z)
	case float64:
		return floatToString(z)

	case []complex64:
		return complexSliceToString(z)
	case []complex128:
		return complexSliceToString(z)

	case complex64:
		return complexToString(z)
	case complex128:
		return complexToString(z)

	case []any:
		return fmt.Sprint(data)

	case any:
		return fmt.Sprint(data)

	default:
		rv := reflect.ValueOf(data)
		rv = ref.Rindirect(rv) // *string -> string
		if k := rv.Kind(); k == reflect.String {
			// eg: flag.stringValue (string)
			return rv.String()
		}

		// reflect approach
		logz.Warn("[anyToString]: unrecognized data type",
			"typ", ref.Typfmtv(&rv),
			"val", ref.Valfmt(&rv),
		)
		return ref.Valfmt(&rv)
	}
}

//

func (s *Cvt) Float64(data any) float64 { return anyToFloat[float64](data) }
func (s *Cvt) Float32(data any) float32 { return anyToFloat[float32](data) }

func (s *Cvt) Float64Slice(data any) []float64 { return anyToFloatSlice[float64](data) }
func (s *Cvt) Float32Slice(data any) []float32 { return anyToFloatSlice[float32](data) }

func (s *Cvt) Float64Map(data any) map[string]float64 { return anyToFloat64Map[float64](data) }
func (s *Cvt) Float32Map(data any) map[string]float32 { return anyToFloat64Map[float32](data) }

//

func (s *Cvt) Complex128(data any) complex128 { return anyToComplex[complex128](data) }
func (s *Cvt) Complex64(data any) complex64   { return anyToComplex[complex64](data) }

func anyToComplex[R Complexes](data any) R {
	if data == nil {
		return 0
	}

	switch z := data.(type) {
	case complex128:
		return R(z)
	case complex64:
		return R(z)

	case int:
		return R(complex(float64(z), 0))
	case int64:
		return R(complex(float64(z), 0))
	case int32:
		return R(complex(float64(z), 0))
	case int16:
		return R(complex(float64(z), 0))
	case int8:
		return R(complex(float64(z), 0))
	case uint:
		return R(complex(float64(z), 0))
	case uint64:
		return R(complex(float64(z), 0))
	case uint32:
		return R(complex(float64(z), 0))
	case uint16:
		return R(complex(float64(z), 0))
	case uint8:
		return R(complex(float64(z), 0))

	case float64:
		return R(complex(z, 0))
	case float32:
		return R(complex(z, 0))

	case string:
		return R(mustParseComplex(z))
	case fmt.Stringer:
		return R(mustParseComplex(z.String()))

	case any:
		return R(mustParseComplex(anyToString(z)))

	default:
		rv := reflect.ValueOf(data)
		rv = ref.Rindirect(rv) // *string -> string
		if k := rv.Kind(); k == reflect.Complex128 || k == reflect.Complex64 {
			// eg: flag.complexValue (complex)
			return R(rv.Complex())
		}
		str := fmt.Sprintf("%v", data)
		return R(mustParseComplex(str))
	}
}

func mustParseComplex(s string) (ret complex128) {
	ret, _ = strconv.ParseComplex(s, 64)
	return
}

func (s *Cvt) Complex128Slice(data any) []complex128 { return anyToComplexSlice[complex128](data) }
func (s *Cvt) Complex64Slice(data any) []complex64   { return anyToComplexSlice[complex64](data) }

func zfToComplexS[T, R Complexes](in []T) (out []R) {
	out = make([]R, 0, len(in))
	for _, it := range in {
		out = append(out, R(it))
	}
	return
}

func anyToComplexSlice[R Complexes](data any) (ret []R) {
	if data == nil {
		return
	}

	switch z := data.(type) {
	case []complex128:
		return zfToComplexS[complex128, R](z)
	case []complex64:
		return zfToComplexS[complex64, R](z)

	case []float64:
		return zsToComplexS[float64, R](z)
	case []float32:
		return zsToComplexS[float32, R](z)

	case []int:
		return zsToComplexS[int, R](z)
	case []int64:
		return zsToComplexS[int64, R](z)
	case []int32:
		return zsToComplexS[int32, R](z)
	case []int16:
		return zsToComplexS[int16, R](z)
	case []int8:
		return zsToComplexS[int8, R](z)
	case []uint:
		return zsToComplexS[uint, R](z)
	case []uint64:
		return zsToComplexS[uint64, R](z)
	case []uint32:
		return zsToComplexS[uint32, R](z)
	case []uint16:
		return zsToComplexS[uint16, R](z)
	case []uint8:
		return zsToComplexS[uint8, R](z)

	case []string:
		ret = make([]R, 0, len(z))
		for _, it := range z {
			ret = append(ret, R(mustParseComplex(it)))
		}
		return
	case []fmt.Stringer:
		ret = make([]R, 0, len(z))
		for _, it := range z {
			ret = append(ret, R(mustParseComplex(it.String())))
		}
		return

	case string:
		return parseToSlice[R](z)
	case fmt.Stringer:
		return parseToSlice[R](z.String())
	case any:
		return parseToSlice[R](anyToString(z))

	default:
		break
	}
	return
}

func zsToComplexS[T Integers | Uintegers | Floats, R Complexes](z []T) (ret []R) {
	ret = make([]R, 0, len(z))
	for _, it := range z {
		ret = append(ret, R(complex(float64(it), 0)))
	}
	return
}

func (s *Cvt) Complex128Map(data any) map[string]complex128 { return anyToComplexMap[complex128](data) }
func (s *Cvt) Complex64Map(data any) map[string]complex64   { return anyToComplexMap[complex64](data) }

func anyToComplexMap[R Complexes](data any) (ret map[string]R) {
	if data == nil {
		return
	}

	switch z := data.(type) {
	case map[string]any:
		return zfToComplex128M[R](z)

	case map[string]bool:
		return zfToComplex128MN[bool, R](z)
	case map[string]string:
		return zfToComplex128MN[string, R](z)

	case map[string]complex128:
		return zfToComplex128MN[complex128, R](z)
	case map[string]complex64:
		return zfToComplex128MN[complex64, R](z)
	case map[string]float64:
		return zfToComplex128MN[float64, R](z)
	case map[string]float32:
		return zfToComplex128MN[float32, R](z)

	case map[string]int64:
		return zfToComplex128MN[int64, R](z)
	case map[string]int32:
		return zfToComplex128MN[int32, R](z)
	case map[string]int16:
		return zfToComplex128MN[int16, R](z)
	case map[string]int8:
		return zfToComplex128MN[int8, R](z)
	case map[string]int:
		return zfToComplex128MN[int, R](z)
	case map[string]uint64:
		return zfToComplex128MN[uint64, R](z)
	case map[string]uint32:
		return zfToComplex128MN[uint32, R](z)
	case map[string]uint16:
		return zfToComplex128MN[uint16, R](z)
	case map[string]uint8:
		return zfToComplex128MN[uint8, R](z)
	case map[string]uint:
		return zfToComplex128MN[uint, R](z)

	case string:
		return parseToMap[R](z)
	case fmt.Stringer:
		return parseToMap[R](z.String())
	case any:
		return parseToMap[R](anyToString(z))

	default:
		break
	}
	return
}

func zfToComplex128M[R Complexes](in map[string]any) (out map[string]R) {
	out = make(map[string]R, len(in))
	for k, it := range in {
		out[k] = anyToComplex[R](it)
	}
	return
}

func zfToComplex128MN[T Numerics | string | bool, Out Complexes](in map[string]T) (out map[string]Out) {
	out = make(map[string]Out, len(in))
	for k, it := range in {
		out[k] = anyToComplex[Out](it)
	}
	return
}

//

func (s *Cvt) Duration(data any) time.Duration { return anyToDuration(data) }

func anyToDuration(data any) time.Duration {
	if data == nil {
		return 0
	}

	switch z := data.(type) {
	case time.Duration:
		return z

	case int:
		return time.Duration(int64(z))
	case int64:
		return time.Duration(int64(z)) //nolint:unconvert
	case int32:
		return time.Duration(int64(z))
	case int16:
		return time.Duration(int64(z))
	case int8:
		return time.Duration(int64(z))
	case uint:
		return time.Duration(int64(z))
	case uint64:
		return time.Duration(int64(z))
	case uint32:
		return time.Duration(int64(z))
	case uint16:
		return time.Duration(int64(z))
	case uint8:
		return time.Duration(int64(z))

	case string:
		return mustParseDuration(z)
	case fmt.Stringer:
		return mustParseDuration(z.String())
	case any:
		return mustParseDuration(anyToString(z))

	default:
		str := fmt.Sprintf("%v", data)
		return mustParseDuration(str)
	}
}

func (s *Cvt) DurationSlice(data any) []time.Duration { return anyToDurationSlice(data) }

func anyToDurationSlice(data any) (ret []time.Duration) {
	if data == nil {
		return
	}

	switch z := data.(type) {
	case []time.Duration:
		return z

	case []int:
		return zsToDurationS(z)
	case []int64:
		return zsToDurationS(z)
	case []int32:
		return zsToDurationS(z)
	case []int16:
		return zsToDurationS(z)
	case []int8:
		return zsToDurationS(z)
	case []uint:
		return zsToDurationS(z)
	case []uint64:
		return zsToDurationS(z)
	case []uint32:
		return zsToDurationS(z)
	case []uint16:
		return zsToDurationS(z)
	case []uint8:
		return zsToDurationS(z)

	case []bool:
		return zsToDurationSB(z)

	case []float32:
		return zsToDurationSB(z)
	case []float64:
		return zsToDurationSB(z)
	case []complex64:
		return zsToDurationSB(z)
	case []complex128:
		return zsToDurationSB(z)

	case []string:
		ret = make([]time.Duration, 0, len(z))
		for _, it := range z {
			ret = append(ret, mustParseDuration(it))
		}
		return
	case []fmt.Stringer:
		ret = make([]time.Duration, 0, len(z))
		for _, it := range z {
			ret = append(ret, mustParseDuration(it.String()))
		}
		return
	case []any:
		ret = make([]time.Duration, 0, len(z))
		for _, it := range z {
			ret = append(ret, mustParseDuration(anyToString(it)))
		}
		return

	case string:
		ret = parseToSlice[time.Duration](z)
	case fmt.Stringer:
		ret = parseToSlice[time.Duration](z.String())
	case any:
		return parseToSlice[time.Duration](anyToString(z))
	default:
		break
	}
	return
}

func mustParseDuration(s string) (dur time.Duration) {
	if v, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Duration(v)
	}
	dur, _ = times.ParseDuration(s)
	return
}

func zsToDurationS[T Integers | Uintegers](z []T) (ret []time.Duration) {
	ret = make([]time.Duration, 0, len(z))
	for _, it := range z {
		ret = append(ret, time.Duration(int64(it)))
	}
	return
}

func zsToDurationSB[T bool | Floats | Complexes](z []T) (ret []time.Duration) {
	ret = make([]time.Duration, 0, len(z))
	for _, it := range z {
		ret = append(ret, time.Duration(anyToInt(it)))
	}
	return
}

func (s *Cvt) DurationMap(data any) map[string]time.Duration { return anyToDurationMap(data) }

func anyToDurationMap(data any) (ret map[string]time.Duration) {
	if data == nil {
		return
	}

	switch z := data.(type) {

	case map[string]time.Duration:
		return z
	case map[string]string:
		ret = make(map[string]time.Duration, len(z))
		for k, v := range z {
			ret[k] = mustParseDuration(v)
		}
		return
	case map[string]fmt.Stringer:
		ret = make(map[string]time.Duration, len(z))
		for k, v := range z {
			ret[k] = mustParseDuration(v.String())
		}
		return
	case map[string]any:
		ret = make(map[string]time.Duration, len(z))
		for k, v := range z {
			ret[k] = anyToDuration(v)
		}
		return
	case map[any]any:
		ret = make(map[string]time.Duration, len(z))
		for k, v := range z {
			ret[anyToString(k)] = anyToDuration(v)
		}
		return

	case string:
		ret = parseToMap[time.Duration](z)
	case fmt.Stringer:
		ret = parseToMap[time.Duration](z.String())
	case any:
		return parseToMap[time.Duration](anyToString(z))

	default:
		break
	}
	return
}

//

func (s *Cvt) Time(data any) time.Time { return anyToTime(data) }

func anyToTime(data any) (tm time.Time) {
	if data == nil {
		return
	}

	switch z := data.(type) {
	case time.Time:
		return z
	case *time.Time:
		return *z

	case int:
		return time.Unix(int64(z), 0)
	case int64:
		return time.Unix(int64(z), 0) //nolint:unconvert
	case int32:
		return time.Unix(int64(z), 0)
	case int16:
		return time.Unix(int64(z), 0)
	case int8:
		return time.Unix(int64(z), 0)
	case uint:
		return time.Unix(int64(z), 0)
	case uint64:
		return time.Unix(int64(z), 0)
	case uint32:
		return time.Unix(int64(z), 0)
	case uint16:
		return time.Unix(int64(z), 0)
	case uint8:
		return time.Unix(int64(z), 0)

	case bool:
		return time.Unix(anyToInt(z), 0)
	case float64:
		return TimestampFromFloat64(z)
	case float32:
		return TimestampFromFloat64(float64(z))
	case complex128:
		return TimestampFromFloat64(real(z))
	case complex64:
		return TimestampFromFloat64(float64(real(z)))

	case string:
		return mustSmartParseTime(z)
	case fmt.Stringer:
		return mustSmartParseTime(z.String())
	case any:
		return mustSmartParseTime(anyToString(z))

	default:
		str := fmt.Sprintf("%v", data)
		return mustSmartParseTime(str)
	}
}

// // readFiletime reads a Windows Filetime struct and converts it to a
// // time.Time value with a UTC timezone.
// func readFiletime(reader io.Reader) (*time.Time, error) {
// 	var filetime syscall.Filetime
// 	err := binary.Read(reader, binary.LittleEndian, &filetime.LowDateTime)
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = binary.Read(reader, binary.LittleEndian, &filetime.HighDateTime)
// 	if err != nil {
// 		return nil, err
// 	}
// 	t := time.Unix(0, filetime.Nanoseconds()).UTC()
// 	return &t, nil
// }

func timeToFloat(t time.Time) float64 { //nolint:unused
	// If time.Time is the empty value, UnixNano will return
	// the farthest back timestamp a float can represent,
	// which is some large negative value.
	// We compromise and call it zero.
	if t.IsZero() {
		return 0
	}
	return float64(t.UnixNano()) / 1e9
}

// TimestampFromFloat64 returns a Timestamp equal to the given
// float64, assuming it too is an unix timestamp.
//
// The float64 is interpreted as number of seconds, with everything
// after the decimal indicating milliseconds, microseconds, and
// nanoseconds
func TimestampFromFloat64(ts float64) time.Time {
	sec, dec := math.Modf(ts)
	return time.Unix(int64(sec), int64(dec*(1e9)))
	// secs := int64(ts)
	// nsecs := int64((ts - float64(secs)) * 1e9)
	// return time.Unix(secs, nsecs)
}

// mustSmartParseTime parses a formatted string and returns the time value it represents.
func mustSmartParseTime(str string) (tm time.Time) {
	tm, _ = smartParseTime(str)
	return
}

func smartParseTime(str string) (tm time.Time, err error) {
	for _, layout := range onceInitTimeFormats() {
		if tm, err = time.Parse(layout, str); err == nil {
			break
		}
	}
	return
}

var (
	knownDateTimeFormats []string
	onceFormats          sync.Once
)

func onceInitTimeFormats() []string {
	onceFormats.Do(func() {
		knownDateTimeFormats = []string{
			"2006-01-02 15:04:05.999999999 -0700",
			"2006-01-02 15:04:05.999999999Z07:00",
			"2006-01-02 15:04:05.999999999",
			"2006-01-02 15:04:05.999",
			"2006-01-02 15:04:05",
			"2006-01-02",
			"2006/01/02",
			"01/02/2006",
			"01-02",

			"2006-1-2 15:4:5.999999999 -0700",
			"2006-1-2 15:4:5.999999999Z07:00",
			"2006-1-2 15:4:5.999999999",
			"2006-1-2 15:4:5.999",
			"2006-1-2 15:4:5",
			"2006-1-2",
			"2006/1/2",
			"1/2/2006",
			"1-2",

			"15:04:05.999999999",
			"15:04.999999999",
			"15:04:05.999",
			"15:04.999",
			"15:04:05",
			"15:04",

			"15:4:5.999999999",
			"15:4.999999999",
			"15:4:5.999",
			"15:4.999",
			"15:4:5",
			"15:4",

			time.RFC3339,
			time.RFC3339Nano,
			time.RFC1123Z,
			time.RFC1123,
			time.RFC850,
			time.RFC822Z,
			time.RFC822,
			time.RubyDate,
			time.UnixDate,
			time.ANSIC,
		}
	})
	return knownDateTimeFormats
}

func (s *Cvt) TimeSlice(data any) []time.Time { return anyToTimeSlice(data) }

func anyToTimeSlice(data any) (ret []time.Time) {
	if data == nil {
		return
	}

	switch z := data.(type) {
	case []time.Time:
		return z
	case []*time.Time:
		break // todo convert []*time.Time to []time.Time?

	case []int:
		return zsToTimeS(z)
	case []int64:
		return zsToTimeS(z)
	case []int32:
		return zsToTimeS(z)
	case []int16:
		return zsToTimeS(z)
	case []int8:
		return zsToTimeS(z)
	case []uint:
		return zsToTimeS(z)
	case []uint64:
		return zsToTimeS(z)
	case []uint32:
		return zsToTimeS(z)
	case []uint16:
		return zsToTimeS(z)
	case []uint8:
		return zsToTimeS(z)

	case []string:
		ret = make([]time.Time, 0, len(z))
		for _, it := range z {
			ret = append(ret, mustSmartParseTime(it))
		}
		return
	case []fmt.Stringer:
		ret = make([]time.Time, 0, len(z))
		for _, it := range z {
			ret = append(ret, mustSmartParseTime(it.String()))
		}
		return

	case []any:
		ret = make([]time.Time, 0, len(z))
		for _, it := range z {
			ret = append(ret, anyToTime(it))
		}
		return

	case string:
		return parseToSlice[time.Time](z)
	case fmt.Stringer:
		return parseToSlice[time.Time](z.String())
	case any:
		return parseToSlice[time.Time](anyToString(z))

	default:
		break
	}
	return
}

func zsToTimeS[T Integers | Uintegers](z []T) (ret []time.Time) {
	ret = make([]time.Time, 0, len(z))
	for _, it := range z {
		ret = append(ret, time.Unix(int64(it), 0))
	}
	return
}

func (s *Cvt) TimeMap(data any) map[string]time.Time { return anyToTimeMap(data) }

func anyToTimeMap(data any) (ret map[string]time.Time) {
	if data == nil {
		return
	}

	switch z := data.(type) {
	case map[string]time.Time:
		return z

	case map[string]string:
		ret = make(map[string]time.Time, len(z))
		for k, v := range z {
			ret[k] = mustSmartParseTime(v)
		}
		return
	case map[string]fmt.Stringer:
		ret = make(map[string]time.Time, len(z))
		for k, v := range z {
			ret[k] = mustSmartParseTime(v.String())
		}
		return
	case map[string]any:
		ret = make(map[string]time.Time, len(z))
		for k, v := range z {
			ret[k] = anyToTime(v)
		}
		return
	case map[any]any:
		ret = make(map[string]time.Time, len(z))
		for k, v := range z {
			ret[anyToString(k)] = anyToTime(v)
		}
		return

	case string:
		return parseToMap[time.Time](z)
	case fmt.Stringer:
		return parseToMap[time.Time](z.String())
	case any:
		return parseToMap[time.Time](anyToString(z))

	default:
		break
	}
	return
}

//

//

//

//

//

//nolint:lll //no why
func (valueConverters ValueConverters) findConverters(params *Params, from, to reflect.Type, userDefinedOnly bool) (converter ValueConverter, ctx *ValueConverterContext) {
	var yes bool
	var minV int
	if userDefinedOnly {
		minV = lenValueConverters
	}
	for i := len(valueConverters) - 1; i >= minV; i-- {
		// FILO: the last added converter has the first priority
		cvt := valueConverters[i]
		if cvt != nil {
			if ctx, yes = cvt.Match(params, from, to); yes {
				converter = cvt
				break
			}
		}
	}
	return
}

//nolint:lll //no why
func (valueCopiers ValueCopiers) findCopiers(params *Params, from, to reflect.Type, userDefinedOnly bool) (copier ValueCopier, ctx *ValueConverterContext) {
	var yes bool
	var minV int
	if userDefinedOnly {
		minV = lenValueCopiers
	}
	for i := len(valueCopiers) - 1; i >= minV; i-- {
		// FILO: the last added converter has the first priority
		cpr := valueCopiers[i]
		if cpr != nil {
			if ctx, yes = cpr.Match(params, from, to); yes {
				copier = cpr
				break
			}
		}
	}
	return
}

//

// IsCopyFunctionResultToTarget does SAFELY test if copyFunctionResultToTarget is true or not.
func (ctx *ValueConverterContext) IsCopyFunctionResultToTarget() bool {
	if ctx == nil || ctx.Params == nil || ctx.controller == nil {
		return false
	}
	return ctx.controller.copyFunctionResultToTarget
}

// IsPassSourceToTargetFunction does SAFELY test if passSourceAsFunctionInArgs is true or not.
func (ctx *ValueConverterContext) IsPassSourceToTargetFunction() bool {
	if ctx == nil || ctx.Params == nil || ctx.controller == nil {
		return false
	}
	return ctx.controller.passSourceAsFunctionInArgs
}

// Preprocess find out a converter to transform source to target.
// If no comfortable converter found, the return processed is false.
//
//nolint:lll //no why
func (ctx *ValueConverterContext) Preprocess(source reflect.Value, targetType reflect.Type, cvtOuter ValueConverter) (processed bool, target reflect.Value, err error) {
	if ctx != nil && ctx.Params != nil && ctx.Params.controller != nil {
		sourceType := source.Type()
		if cvt, ctxCvt := ctx.controller.valueConverters.findConverters(ctx.Params, sourceType, targetType, false); cvt != nil && cvt != cvtOuter {
			target, err = cvt.Transform(ctxCvt, source, targetType)
			processed = true
			return
		}
	}
	return
}

//nolint:unused //future
func (ctx *ValueConverterContext) isStruct() bool {
	if ctx == nil {
		return false
	}
	return ctx.Params.isStruct()
}

//nolint:unused //future
func (ctx *ValueConverterContext) isFlagExists(ftf cms.CopyMergeStrategy) (ret bool) {
	if ctx == nil {
		return
	}
	return ctx.Params.isFlagExists(ftf)
}

// isGroupedFlagOK tests if the given flag is exists or valid.
//
// Different with isGroupedFlagOKDeeply is, isGroupedFlagOK will return
// false simply while Params.fieldTags is empty or unset.
//
// When Params.fieldTags is valid, the actual testing will be forwarded
// to Params.fieldTags.flags.isGroupedFlagOK().
func (ctx *ValueConverterContext) isGroupedFlagOK(ftf ...cms.CopyMergeStrategy) (ret bool) { //nolint:unused //future
	if ctx == nil {
		return flags.New().IsGroupedFlagOK(ftf...)
	}
	return ctx.Params.isGroupedFlagOK(ftf...)
}

// isGroupedFlagOKDeeply tests if the given flag is exists or valid.
//
// Different with isGroupedFlagOK is, isGroupedFlagOKDeeply will check
// whether the given flag is a leader (i.e. default choice) in a group
// or not, even if Params.fieldTags is empty or unset.
//
// When Params.fieldTags is valid, the actual testing will be forwarded
// to Params.fieldTags.flags.isGroupedFlagOK().
func (ctx *ValueConverterContext) isGroupedFlagOKDeeply(ftf ...cms.CopyMergeStrategy) (ret bool) {
	if ctx == nil {
		return flags.New().IsGroupedFlagOK(ftf...)
	}
	return ctx.Params.isGroupedFlagOKDeeply(ftf...)
}

//nolint:unused //future
func (ctx *ValueConverterContext) isAnyFlagsOK(ftf ...cms.CopyMergeStrategy) (ret bool) {
	if ctx == nil {
		return flags.New().IsAnyFlagsOK(ftf...)
	}
	return ctx.Params.isAnyFlagsOK(ftf...)
}

//nolint:unused //future
func (ctx *ValueConverterContext) isAllFlagsOK(ftf ...cms.CopyMergeStrategy) (ret bool) {
	if ctx == nil {
		return flags.New().IsAllFlagsOK(ftf...)
	}
	return ctx.Params.isAllFlagsOK(ftf...)
}

//

//

//

type cvtbase struct{}

func (c *cvtbase) safeType(tgt, tgtptr reflect.Value) reflect.Type {
	if tgt.IsValid() {
		return tgt.Type()
	}
	if tgtptr.IsValid() {
		if tgtptr.Kind() == reflect.Interface {
			return tgtptr.Type()
		}
		return tgtptr.Type().Elem()
	}
	logz.Panic("niltyp !! CANNOT fetch type:", "tgt", ref.Typfmtv(&tgt), "tgtptr", ref.Typfmtv(&tgtptr))
	return ref.Niltyp
}

// processUnexportedField try to set newval into target if it's an unexported field.
func (c *cvtbase) processUnexportedField(ctx *ValueConverterContext, target, newval reflect.Value) (processed bool) {
	if ctx == nil || ctx.Params == nil {
		return
	}
	processed = ctx.Params.processUnexportedField(target, newval)
	return
}

//nolint:lll //no why
func (c *cvtbase) checkSource(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, processed bool) {
	if ctx == nil {
		return
	}

	if processed = ctx.isGroupedFlagOKDeeply(cms.Ignore); processed {
		return
	}
	if processed = ref.IsNil(source) && ctx.isGroupedFlagOKDeeply(cms.OmitIfNil, cms.OmitIfEmpty); processed {
		target = reflect.Zero(targetType)
		return
	}
	if processed = ref.IsZero(source) && ctx.isGroupedFlagOKDeeply(cms.OmitIfZero, cms.OmitIfEmpty); processed {
		target = reflect.Zero(targetType)
	}
	return
}

//nolint:lll //no why
func (c *cvtbase) checkTarget(ctx *ValueConverterContext, target reflect.Value, targetType reflect.Type) (processed bool) {
	// if processed = !target.IsValid(); processed {
	// 	return
	// }

	if processed = c.checkTargetLite(ctx, target, targetType); processed {
		return
	}

	if processed = !ref.IsValid(target); processed {
		return
	}

	return
}

//nolint:lll //no why
func (c *cvtbase) checkTargetLite(ctx *ValueConverterContext, target reflect.Value, targetType reflect.Type) (processed bool) {
	// if processed = !target.IsValid(); processed {
	// 	return
	// }

	if processed = ref.IsNil(target) && ctx.isGroupedFlagOKDeeply(cms.OmitIfTargetNil); processed {
		return
	}
	processed = ref.IsZero(target) && ctx.isGroupedFlagOKDeeply(cms.OmitIfTargetZero)
	_ = targetType
	return
}

//

type toConverterBase struct{ cvtbase }

func (c *toConverterBase) fallback(target reflect.Value) (err error) { //nolint:unparam
	tgtType := reflect.TypeOf((*time.Duration)(nil)).Elem()
	ref.Rindirect(target).Set(reflect.Zero(tgtType))
	return
}

//

type fromConverterBase struct{ cvtbase }

func (c *fromConverterBase) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	panic("not impl")
}

func (c *fromConverterBase) Transform(ctx *ValueConverterContext,
	source reflect.Value,
	targetType reflect.Type,
) (target reflect.Value, err error) {
	panic("not impl")
}

func (c *fromConverterBase) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	panic("not impl")
}

//nolint:unused,lll //future
func (c *fromConverterBase) preprocess(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (processed bool, target reflect.Value, err error) {
	if !(ctx != nil && ctx.Params != nil && ctx.Params.controller != nil) {
		return
	}

	sourceType := source.Type()
	if cvt, ctxCvt := ctx.controller.valueConverters.findConverters(ctx.Params, sourceType, targetType, false); cvt != nil {
		if cvt == c {
			return
		}
		if cc, ok := cvt.(*fromConverterBase); ok && cc == c {
			return
		}

		target, err = cvt.Transform(ctxCvt, source, targetType)
		processed = true
		return
	}
	return
}

func (c *fromConverterBase) postCopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	// if source.IsValid() {
	//	if canConvert(&source, target.Type()) {
	//		nv := source.Convert(target.Type())
	//		target.Set(nv)
	//		return
	//		//} else {
	//		//	nv := fmt.Sprintf("%v", source.Interface())
	//		//	target.Set(reflect.ValueOf(nv))
	//	}
	// }
	//
	// target = reflect.Zero(target.Type())
	// return
	var nv reflect.Value
	nv, err = c.convertToOrZeroTarget(ctx, source, target.Type())
	if err == nil {
		if target.CanSet() {
			dbglog.Log("    postCopyTo: set nv(%v) into target (%v)", ref.Valfmt(&nv), ref.Valfmt(&target))
			target.Set(nv)
		} else {
			err = ErrCannotSet.FormatWith(ref.Valfmt(&target), ref.Typfmtv(&target), ref.Valfmt(&nv), ref.Typfmtv(&nv))
		}
	}
	return
}

func (c *fromConverterBase) convertToOrZeroTarget(ctx *ValueConverterContext,
	source reflect.Value,
	targetType reflect.Type,
) (target reflect.Value, err error) { //nolint:unparam
	if ref.CanConvert(&source, targetType) {
		nv := source.Convert(targetType)
		target = nv
	} else if ctx.isGroupedFlagOKDeeply(cms.ClearIfInvalid) {
		target = reflect.Zero(targetType)
	}
	return
}

//

//

//

type toStringConverter struct{ toConverterBase }

func (c *toStringConverter) postCopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) { //nolint:unparam
	if source.IsValid() {
		if ref.CanConvert(&source, target.Type()) {
			nv := source.Convert(target.Type())
			if c.processUnexportedField(ctx, target, nv) {
				return
			}
			target.Set(nv)
			return
		}

		newVal := fmt.Sprintf("%v", source.Interface())
		nv := reflect.ValueOf(newVal)
		if c.processUnexportedField(ctx, target, nv) {
			return
		}
		target.Set(nv)
		return
	}

	// target = reflect.Zero(target.Type())
	return //nolint:nakedret //i do
}

func (c *toStringConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	tgt, tgtptr := ref.Rdecode(target)
	tgtType := c.safeType(tgt, tgtptr) // because tgt might be invalid, so we fetch tgt type via its pointer
	dbglog.Log("     target: %v (%v), tgtptr: %v, tgt: %v, tgttyp: %v",
		ref.Typfmtv(&target), ref.Typfmt(target.Type()), ref.Typfmtv(&tgtptr),
		ref.Typfmtv(&tgt), ref.Typfmt(tgtType))

	if processed := c.checkTargetLite(ctx, tgt, tgtType); processed {
		return
	}

	if ret, e := c.Transform(ctx, source, tgtType); e == nil {
		if c.processUnexportedField(ctx, target, ret) {
			return
		}
		dbglog.Log("     set: %v (%v) <- %v", ref.Valfmt(&target), ref.Typfmtv(&target), ref.Valfmt(&ret))
		tgtptr.Set(ret)
	} else {
		err = c.postCopyTo(ctx, source, target)
	}
	return
}

// Transform will transform source type (bool, int, ...) to target string.
func (c *toStringConverter) Transform(ctx *ValueConverterContext,
	source reflect.Value,
	targetType reflect.Type,
) (target reflect.Value, err error) {
	if source.IsValid() {
		// var processed bool
		// if processed, target, err = ctx.Preprocess(source, targetType, c); processed {
		//	return
		// }

		var processed bool
		if target, processed = c.checkSource(ctx, source, targetType); processed {
			return
		}

		var str string
		if str, processed, err = tryMarshalling(source); processed {
			if err == nil {
				target = reflect.ValueOf(str)
			}
			return
		}

		target, err = rToString(source, targetType)
		return
	}

	if ctx == nil || ctx.isGroupedFlagOKDeeply(cms.ClearIfInvalid) {
		target = reflect.Zero(reflect.TypeOf((*string)(nil)).Elem())
	}
	return
}

func (c *toStringConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	if yes = target.Kind() == reflect.String; yes {
		ctx = &ValueConverterContext{params}
	}
	return
}

//

var marshallableTypes = map[string]reflect.Type{ //nolint:gochecknoglobals //no
	// "MarshalBinary": reflect.TypeOf((*encoding.BinaryMarshaler)(nil)).Elem(),
	"MarshalText": reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem(),
	"MarshalJSON": reflect.TypeOf((*json.Marshaler)(nil)).Elem(),
}

var textMarshaller = TextMarshaller(func(v interface{}) ([]byte, error) { //nolint:gochecknoglobals //no
	return json.MarshalIndent(v, "", "  ")
})

func canMarshalling(source reflect.Value) (mtd reflect.Value, yes bool) {
	st := source.Type()
	for fnn, t := range marshallableTypes {
		if st.Implements(t) {
			yes, mtd = true, source.MethodByName(fnn)
			break
		}
	}
	return
}

// FallbackToBuiltinStringMarshalling exposes the builtin string
// marshaling mechanism for your customized ValueConverter or
// ValueCopier.
func FallbackToBuiltinStringMarshalling(source reflect.Value) (str string, err error) {
	return doMarshalling(source)
}

func doMarshalling(source reflect.Value) (str string, err error) {
	var data []byte
	if mtd, yes := canMarshalling(source); yes {
		ret := mtd.Call(nil)
		if err, yes = (ret[1].Interface()).(error); err == nil && yes {
			data = ret[0].Interface().([]byte) //nolint:errcheck //no need
		}
	} else {
		data, err = textMarshaller(source.Interface())
	}
	if err == nil {
		str = string(data)
	}
	return
}

func tryMarshalling(source reflect.Value) (str string, processed bool, err error) {
	var data []byte
	var mtd reflect.Value
	if mtd, processed = canMarshalling(source); processed {
		ret := mtd.Call(nil)
		if err, _ = (ret[1].Interface()).(error); err == nil {
			data = ret[0].Interface().([]byte) //nolint:errcheck //no need
		}
	}
	if err == nil {
		str = string(data)
	}
	return
}

//

type fromStringConverter struct{ fromConverterBase }

func (c *fromStringConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	tgt, tgtptr := ref.Rdecode(target)
	tgttyp := c.safeType(tgt, tgtptr) // because tgt might be invalid, so we fetch tgt type via its pointer
	dbglog.Log("  target: %v (%v), tgtptr: %v, tgt: %v, tgttyp: %v",
		ref.Typfmtv(&target), ref.Typfmt(target.Type()), ref.Typfmtv(&tgtptr),
		ref.Typfmtv(&tgt), ref.Typfmt(tgttyp))

	if processed := c.checkTargetLite(ctx, tgt, tgttyp); processed {
		// target.Set(ret)
		return
	}

	var ret reflect.Value
	var e error
	if ret, e = c.Transform(ctx, source, tgttyp); e == nil {
		if tgtptr.Kind() == reflect.Interface { //nolint:gocritic // no need to switch to 'switch' clause
			tgtptr.Set(ret)
		} else if tgtptr.Kind() == reflect.Ptr {
			tgtptr.Elem().Set(ret)
		} else if tgt.CanSet() {
			tgt.Set(ret)
		} else {
			err = ErrCannotSet.FormatWith(ref.Valfmt(&tgt), ref.Typfmtv(&tgt), ref.Valfmt(&ret), ref.Typfmtv(&ret))
		}
		dbglog.Log("  tgt / ret transformed: %v / %v", ref.Valfmt(&tgt), ref.Valfmt(&ret))
		return
	}

	if !errors.Is(e, &strconv.NumError{Err: strconv.ErrSyntax}) && !errors.IsAnyOf(e, strconv.ErrSyntax, strconv.ErrRange) {
		dbglog.Log("  Transform() failed: %v", e)
		dbglog.Log("  try running postCopyTo()")
		err = c.postCopyTo(ctx, source, target)
	}
	return
}

// Transform will transform source string to target type (bool, int, ...)
func (c *fromStringConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	if !source.IsValid() {
		target, err = c.convertToOrZeroTarget(ctx, source, targetType)
		return
	}

	// var processed bool
	// if processed, target, err = c.preprocess(ctx, source, targetType); processed {
	//	return
	// }

	var processed bool
	if target, processed = c.checkSource(ctx, source, targetType); processed {
		return
	}

	switch k := targetType.Kind(); k { //nolint:exhaustive //no need
	case reflect.Bool:
		target = rToBool(source)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		target, err = rToInteger(source, targetType)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		target, err = rToUInteger(source, targetType)

	case reflect.Uintptr:
		target = rToUIntegerHex(source, targetType)
	// case reflect.UnsafePointer:
	//	// target = rToUIntegerHex(source, targetType)
	//	err = errors.InvalidArgument
	// case reflect.Ptr:
	//	//target = rToUIntegerHex(source, targetType)
	//	err = errors.InvalidArgument

	case reflect.Float32, reflect.Float64:
		target, err = rToFloat(source, targetType)
	case reflect.Complex64, reflect.Complex128:
		target, err = rToComplex(source, targetType)

	case reflect.String:
		target = source

	// reflect.Array
	// reflect.Chan
	// reflect.Func
	// reflect.Interface
	// reflect.Map
	// reflect.Slice
	// reflect.Struct

	default:
		target, err = c.convertToOrZeroTarget(ctx, source, targetType)
	}
	return
}

//nolint:lll //keep it
func (c *fromStringConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	if yes = source.Kind() == reflect.String; yes {
		ctx = &ValueConverterContext{params}
	}
	return
}

//

//

//

// fromMapConverter transforms a map to other types (esp string, slice, struct).
type fromMapConverter struct{ fromConverterBase }

func (c *fromMapConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	tgt, tgtptr := ref.Rdecode(target)
	tgttyp := c.safeType(tgt, tgtptr) // because tgt might be invalid, so we fetch tgt type via its pointer
	dbglog.Log("  target: %v (%v), tgtptr: %v, tgt: %v, tgttyp: %v",
		ref.Typfmtv(&target), ref.Typfmt(target.Type()), ref.Typfmtv(&tgtptr),
		ref.Typfmtv(&tgt), ref.Typfmt(tgttyp))

	if processed := c.checkTargetLite(ctx, tgt, tgttyp); processed {
		// target.Set(ret)
		return
	}

	if ctx.controller.targetSetter != nil {
		if tgttyp.Kind() == reflect.Struct {
			// DON'T get into c.Transform(), because Transform will modify
			// a temporary new instance and return it to caller, and
			// the new instance will be set to 'target'.
			// When target setter is valid, we assume the setter will
			// modify the real target directly rather than on a temporary
			// object.
			err = c.toStructDirectly(ctx, source, target, tgttyp)
			return
		}
	}

	if ret, e := c.Transform(ctx, source, tgttyp); e == nil { //nolint:nestif //keep it
		if k := tgtptr.Kind(); k == reflect.Interface { //nolint:gocritic // no need to switch to 'switch' clause
			tgtptr.Set(ret)
		} else if k == reflect.Ptr {
			tgtptr.Elem().Set(ret)
			// } else if tool.IsZero(tgt) {
		} else if tgt.CanSet() {
			tgt.Set(ret)
		} else {
			err = ErrCannotSet.FormatWith(ref.Valfmt(&tgt), ref.Typfmtv(&tgt), ref.Valfmt(&ret), ref.Typfmtv(&ret))
		}
		dbglog.Log("  tgt: %v (ret = %v)", ref.Valfmt(&tgt), ref.Valfmt(&ret))
	} else if !errors.Is(e, strconv.ErrSyntax) && !errors.Is(e, strconv.ErrRange) {
		dbglog.Log("  Transform() failed: %v", e)
		dbglog.Log("  try running postCopyTo()")
		err = c.postCopyTo(ctx, source, target)
	}
	return
}

// Transform will transform source string to target type (bool, int, ...)
//
//nolint:lll //keep it
func (c *fromMapConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	if source.IsValid() {
		var processed bool
		if target, processed = c.checkSource(ctx, source, targetType); processed {
			return
		}

		switch k := targetType.Kind(); k { //nolint:exhaustive //no need
		case reflect.String:
			var str string
			if str, err = doMarshalling(source); err == nil {
				target = reflect.ValueOf(str)
			}

		case reflect.Struct:
			target, err = c.toStruct(ctx, source, targetType)

		// case reflect.Slice:
		// case reflect.Array:

		default:
			target, err = c.convertToOrZeroTarget(ctx, source, targetType)
		}
	} else {
		target, err = c.convertToOrZeroTarget(ctx, source, targetType)
	}
	return
}

//nolint:lll,gocognit //keep it
func (c *fromMapConverter) toStructDirectly(ctx *ValueConverterContext, source, target reflect.Value, targetType reflect.Type) (err error) { //nolint:unparam
	cc, _, _ := ctx.controller, target, targetType

	preSetter := func(value reflect.Value, names ...string) (processed bool, err error) {
		if cc.targetSetter != nil {
			if err = cc.targetSetter(&value, names...); err == nil {
				processed = true
			} else {
				if err != ErrShouldFallback { //nolint:errorlint //want it exactly
					return // has error
				}
				err, processed = nil, false
			}
		}
		return
	}

	ec := errors.New("map -> struct errors")
	defer ec.Defer(&err)

	keys := source.MapKeys()
	for _, key := range keys {
		src := source.MapIndex(key)
		st := src.Kind()
		if st == reflect.Interface {
			src = src.Elem()
		}

		// convert map key to string type
		key, err = rToString(key, ref.StringType)
		if err != nil {
			continue // ignore non-string key
		}
		ks := key.String()
		dbglog.Log("  key %q, src: %v (%v)", ks, ref.Valfmt(&src), ref.Typfmtv(&src))

		if cc.targetSetter != nil {
			newtyp := src.Type()
			val := reflect.New(newtyp).Elem()
			err = ctx.controller.copyTo(ctx.Params, src, val)
			dbglog.Log("  nv.%q: %v (%v) ", ks, ref.Valfmt(&val), ref.Typfmtv(&val))
			var processed bool
			if processed, err = preSetter(val, ks); err != nil || processed {
				ec.Attach(err)
				continue
			}
		}
	}
	return
}

func toExportedName(s string) string {
	if s != "" {
		a := wordSplitter(s)
		for i, word := range a {
			a[i] = makeCapitalize1st(word)
		}
		var r []rune
		for _, word := range a {
			r = append(r, word...)
		}
		return string(r)
	}
	return s
}

func wordSplitter(s string) (result [][]rune) {
	runes := []rune(s)
	var word []rune
	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, word)
			}
			word = nil
		} else if r == '-' || r == '_' {
			if i > 0 {
				result = append(result, word)
			}
			word = nil
			continue
		}
		word = append(word, r)
	}
	if len(word) > 0 {
		result = append(result, word)
	}
	return
}

func makeCapitalize1st(r []rune) (ret []rune) {
	if len(r) > 0 {
		ret = append(ret, unicode.ToUpper(r[0]))
		ret = append(ret, r[1:]...)
		return
	}
	return r
}

//nolint:lll,gocognit //keep it
func (c *fromMapConverter) toStruct(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	cc := ctx.controller

	preSetter := func(value reflect.Value, names ...string) (processed bool, err error) {
		if cc.targetSetter != nil {
			if err = cc.targetSetter(&value, names...); err == nil {
				processed = true
			} else {
				if err != ErrShouldFallback { //nolint:errorlint //want it exactly
					return // has error
				}
				err, processed = nil, false
			}
		}
		return
	}

	ec := errors.New("map -> struct errors")
	defer ec.Defer(&err)

	target = reflect.New(targetType).Elem()
	keys := source.MapKeys()
	for _, key := range keys {
		src := source.MapIndex(key)
		st := src.Kind()
		if st == reflect.Interface {
			src = src.Elem()
		}

		// convert map key to string type
		key, err = rToString(key, ref.StringType)
		if err != nil {
			continue // ignore non-string key
		}
		ks := key.String()
		dbglog.Log("  key %q, src: %v (%v)", ks, ref.Valfmt(&src), ref.Typfmtv(&src))

		if cc.targetSetter != nil {
			newtyp := src.Type()
			val := reflect.New(newtyp).Elem()
			err = ctx.controller.copyTo(ctx.Params, src, val)
			dbglog.Log("  nv.%q: %v (%v) ", ks, ref.Valfmt(&val), ref.Typfmtv(&val))
			var processed bool
			if processed, err = preSetter(val, ks); err != nil || processed {
				ec.Attach(err)
				continue
			}
		}

		const tryForExportedFieldName = true

		// use the key.(string) as the target struct field name
		tsf, ok := targetType.FieldByName(ks)
		if !ok {
			var Ks string
			if tryForExportedFieldName {
				if Ks = toExportedName(ks); Ks != ks {
					tsf, ok = targetType.FieldByName(Ks)
				}
			}
			if !ok {
				continue
			}
			ks = Ks
		}

		fld := target.FieldByName(ks)
		// dbglog.Log("  fld %q: ", ks)
		tsft := tsf.Type
		tsfk := tsft.Kind()
		if tsfk == reflect.Interface {
			// tsft = tsft.Elem()
			fld = fld.Elem()
		} else if tsfk == reflect.Ptr {
			dbglog.Log("  fld.%q: %v (%v)", ks, ref.Valfmt(&fld), ref.Typfmtv(&fld))
			if fld.IsNil() {
				n := reflect.New(fld.Type().Elem())
				target.FieldByName(ks).Set(n)
				fld = target.FieldByName(ks)
			}
			// tsft = tsft.Elem()
			fld = fld.Elem()
			dbglog.Log("  fld.%q: %v (%v)", ks, ref.Valfmt(&fld), ref.Typfmtv(&fld))
		}

		err = ctx.controller.copyTo(ctx.Params, src, fld)
		dbglog.Log("  nv.%q: %v (%v) ", ks, ref.Valfmt(&fld), ref.Typfmtv(&fld))
		ec.Attach(err)

		// var nv reflect.Value
		// nv, err = c.fromConverterBase.convertToOrZeroTarget(ctx, src, tsft)
		// dbglog.Log("  nv.%q: %v (%v) ", ks, valfmt(&nv), typfmtv(&nv))
		// if err == nil {
		//	if fld.CanSet() {
		//		if tsfk == reflect.Ptr {
		//			n := reflect.New(fld.Type().Elem())
		//			n.Elem().Set(nv)
		//			nv = n
		//		}
		//		fld.Set(nv)
		//	} else {
		//		err = ErrCannotSet.FormatWith(valfmt(&fld), typfmtv(&fld), valfmt(&nv), typfmtv(&nv))
		//	}
		// }
	}
	dbglog.Log("  target: %v (%v) ", ref.Valfmt(&target), ref.Typfmtv(&target))
	return
}

func (c *fromMapConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	if yes = source.Kind() == reflect.Map && target.Kind() != reflect.Map; yes {
		ctx = &ValueConverterContext{params}
	}
	return
}

//

//

//

// fromSyncPkgConverter provides default actions for all entities
// in sync package, such as sync.Pool, sync.RWMutex, and so on.
//
// By default, these entities should NOT be copied from one to another
// one. So our default actions are empty.
type fromSyncPkgConverter struct{ fromConverterBase } //nolint:unused

//nolint:lll //keep it
func (c *fromSyncPkgConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	// st.PkgPath() . st.Name()
	if yes = source.Kind() == reflect.Struct && strings.HasPrefix(source.String(), "sync."); yes {
		ctx = &ValueConverterContext{params}
		dbglog.Log("    src: %v, tgt: %v | Matched", source, target)
		// } else {
		// dbglog.Log("    src: %v, tgt: %v", source, target)
	}
	return
}

func (c *fromSyncPkgConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	_, _, _ = ctx, source, target
	return
}

func (c *fromSyncPkgConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	_, _, _ = ctx, source, target
	return
}

//

type fromBytesBufferConverter struct{ fromConverterBase }

func (c *fromBytesBufferConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	tgt, tgtptr := ref.Rdecode(target)
	tgtType := c.safeType(tgt, tgtptr) // because tgt might be invalid, so we fetch tgt type via its pointer
	// tgtType := target.Type()
	dbglog.Log(" target: %v (%v), tgtptr: %v, tgt: %v, tgttyp: %v",
		ref.Typfmtv(&target), ref.Typfmt(target.Type()),
		ref.Typfmtv(&tgtptr), ref.Typfmtv(&tgt), ref.Typfmt(tgtType))

	if processed := c.checkTarget(ctx, tgt, tgtType); processed {
		// target.Set(ret)
		return
	}

	from := source.Interface().(bytes.Buffer) //nolint:errcheck //no need
	tv := tgtptr.Interface()
	switch to := tv.(type) {
	case bytes.Buffer:
		to.Reset()
		to.Write(from.Bytes())
		// dbglog.Log("     to: %v", to.String())
	case *bytes.Buffer:
		to.Reset()
		to.Write(from.Bytes())
		// dbglog.Log("    *to: %v", to.String())
	case *[]byte:
		tgtptr.Elem().Set(reflect.ValueOf(from.Bytes()))
	case []byte:
		tgtptr.Elem().Set(reflect.ValueOf(from.Bytes()))
	}
	return
}

//nolint:lll //keep it
func (c *fromBytesBufferConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	var processed bool
	if target, processed = c.checkSource(ctx, source, targetType); processed {
		return
	}

	// TO/DO implement me
	// panic("implement me")
	from := source.Interface().(bytes.Buffer) //nolint:errcheck //no need
	var to bytes.Buffer
	_, err = to.Write(from.Bytes())
	target = reflect.ValueOf(to)
	return
}

//nolint:lll //keep it
func (c *fromBytesBufferConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	// st.PkgPath() . st.Name()
	if yes = source.Kind() == reflect.Struct && source.String() == "bytes.Buffer"; yes {
		ctx = &ValueConverterContext{params}
		dbglog.Log("    src: %v, tgt: %v | Matched", source, target)
		// } else {
		// dbglog.Log("    src: %v, tgt: %v", source, target)
	}
	_, _ = source, target
	return
}

//

//

//

type fromTimeConverter struct{ fromConverterBase }

func (c *fromTimeConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	tgt, tgtptr := ref.Rdecode(target)
	tgtType := c.safeType(tgt, tgtptr) // because tgt might be invalid, so we fetch tgt type via its pointer
	dbglog.Log("  target: %v (%v), tgtptr: %v, tgt: %v, tgttyp: %v",
		ref.Typfmtv(&target), ref.Typfmt(target.Type()), ref.Typfmtv(&tgtptr),
		ref.Typfmtv(&tgt), ref.Typfmt(tgtType))

	if processed := c.checkTargetLite(ctx, tgt, tgtType); processed {
		// tgtptr.Set(ret)
		return
	}

	var ret reflect.Value
	var e error
	if ret, e = c.Transform(ctx, source, tgtType); e == nil {
		if k := tgtptr.Kind(); k == reflect.Interface { //nolint:gocritic // no need to switch to 'switch' clause
			tgtptr.Set(ret)
		} else if k == reflect.Ptr {
			tgtptr.Elem().Set(ret)
		} else if tgt.CanSet() {
			tgt.Set(ret)
		} else {
			err = ErrCannotSet.FormatWith(ref.Valfmt(&tgt), ref.Typfmtv(&tgt), ref.Valfmt(&ret), ref.Typfmtv(&ret))
		}
		dbglog.Log("  tgt: %v (ret = %v)", ref.Valfmt(&tgt), ref.Valfmt(&ret))
		return
	}

	dbglog.Log("  Transform() failed: %v", e)
	dbglog.Log("              trying to postCopyTo()")
	err = c.postCopyTo(ctx, source, target)
	return
}

//nolint:lll //keep it
func (c *fromTimeConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	if source.IsValid() {
		var processed bool
		if target, processed = c.checkSource(ctx, source, targetType); processed {
			return
		}

		switch k := targetType.Kind(); k { //nolint:exhaustive //no need
		case reflect.Bool:
			b := ref.IsNil(source) || ref.IsZero(source)
			target = reflect.ValueOf(b)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			tm := source.Interface().(time.Time) //nolint:errcheck //no need
			t := reflect.ValueOf(tm.Unix())
			target, err = rToInteger(t, targetType)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tm := source.Interface().(time.Time) //nolint:errcheck //no need
			t := reflect.ValueOf(tm.Unix())
			target, err = rToUInteger(t, targetType)

		case reflect.Float32, reflect.Float64:
			tm := source.Interface().(time.Time) //nolint:errcheck //no need
			f := float64(tm.UnixNano()) / 1e9    //nolint:gomnd //simple case
			t := reflect.ValueOf(f)
			target, err = rToFloat(t, targetType)
		case reflect.Complex64, reflect.Complex128:
			tm := source.Interface().(time.Time) //nolint:errcheck //no need
			f := float64(tm.UnixNano()) / 1e9    //nolint:gomnd //simple case
			t := reflect.ValueOf(f)
			target, err = rToComplex(t, targetType)

		case reflect.String:
			tm := source.Interface().(time.Time) //nolint:errcheck //no need
			str := tm.Format(time.RFC3339)
			t := reflect.ValueOf(str)
			target, err = rToString(t, targetType)

		default:
			target, err = c.convertToOrZeroTarget(ctx, source, targetType)
		}
	} else {
		target, err = c.convertToOrZeroTarget(ctx, source, targetType)
	}
	return
}

func (c *fromTimeConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	if sk := source.Kind(); sk == reflect.Struct {
		if yes = source.Name() == "Time" && source.PkgPath() == timeConstString; yes {
			ctx = &ValueConverterContext{params}
		}
	}
	return
}

//

//nolint:gochecknoglobals //i know that
var knownTimeLayouts = []string{
	"2006-01-02 15:04:05.000000000",
	"2006-01-02 15:04:05.000000",
	"2006-01-02 15:04:05.000",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02",

	"2006-01-02 15:04:05.999999999Z07:00",
	"2006-01-02 15:04:05.999999999",
	"2006-01-02 15:04:05Z07:00",
	"2006-01-02 15:04:05",

	time.RFC3339,

	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC3339Nano,
	time.Kitchen,
	time.Stamp,
	time.StampMilli,
	time.StampMicro,
	time.StampNano,

	"01/02/2006 15:04:05.000000000",
	"01/02/2006 15:04:05.000000",
	"01/02/2006 15:04:05.000",
	"01/02/2006 15:04:05",
	"01/02/2006 15:04",
	"01/02/2006",
}

type toTimeConverter struct{ toConverterBase }

// func (c *toTimeConverter) fallback(target reflect.Value) (err error) {
//	var timeTimeTyp = reflect.TypeOf((*time.Time)(nil)).Elem()
//	rindirect(target).Set(reflect.Zero(timeTimeTyp))
//	return
// }

func (c *toTimeConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	// tgtType := target.Type()
	tgt, tgtptr := ref.Rdecode(target)
	tgtType := c.safeType(tgt, tgtptr) // because tgt might be invalid, so we fetch tgt type via its pointer
	dbglog.Log("  target: %v (%v), tgtptr: %v, tgt: %v, tgttyp: %v",
		ref.Typfmtv(&target), ref.Typfmt(target.Type()), ref.Typfmtv(&tgtptr),
		ref.Typfmtv(&tgt), ref.Typfmt(tgtType))

	if processed := c.checkTargetLite(ctx, tgt, tgtType); processed {
		// target.Set(ret)
		return
	}

	if ret, e := c.Transform(ctx, source, tgtType); e == nil {
		target.Set(ret)
	} else if ctx.isGroupedFlagOKDeeply(cms.ClearIfInvalid) {
		err = c.fallback(target)
	}
	return
}

func tryParseTime(s string) (tm time.Time) {
	var err error
	for _, layout := range knownTimeLayouts {
		tm, err = time.Parse(layout, s)
		if err == nil {
			return
		}
	}
	return
}

//nolint:lll //keep it
func (c *toTimeConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	if source.IsValid() { //nolint:gocritic // no need to switch to 'switch' clause
		var processed bool
		if target, processed = c.checkSource(ctx, source, targetType); processed {
			return
		}

		switch k := source.Kind(); k { //nolint:exhaustive //no need
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			tm := time.Unix(source.Int(), 0)
			target = reflect.ValueOf(tm)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tm := time.Unix(int64(source.Uint()), 0)
			target = reflect.ValueOf(tm)

		case reflect.Float32, reflect.Float64:
			sec, dec := math.Modf(source.Float())
			tm := time.Unix(int64(sec), int64(dec*(1e9)))
			target = reflect.ValueOf(tm)
		case reflect.Complex64, reflect.Complex128:
			sec, dec := math.Modf(real(source.Complex()))
			tm := time.Unix(int64(sec), int64(dec*(1e9)))
			target = reflect.ValueOf(tm)

		case reflect.String:
			tm := tryParseTime(source.String())
			target = reflect.ValueOf(tm)

		default:
			err = ErrCannotConvertTo.FormatWith(source, ref.Typfmtv(&source), targetType, targetType.Kind())
		}
	} else if ctx.isGroupedFlagOKDeeply(cms.ClearIfInvalid) {
		target = reflect.Zero(targetType)
	} else {
		err = errors.New("source (%v) is invalid", ref.Valfmt(&source))
	}
	return
}

func (c *toTimeConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	if tk := target.Kind(); tk == reflect.Struct {
		if yes = target.Name() == "Time" && target.PkgPath() == timeConstString; yes {
			ctx = &ValueConverterContext{params}
		}
	}
	return
}

//

//

//

type fromDurationConverter struct{ fromConverterBase }

func (c *fromDurationConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	tgt, tgtptr := ref.Rdecode(target)
	tgttyp := c.safeType(tgt, tgtptr) // because tgt might be invalid, so we fetch tgt type via its pointer
	dbglog.Log("  target: %v (%v), tgtptr: %v, tgt: %v, tgttyp: %v",
		ref.Typfmtv(&target), ref.Typfmt(target.Type()), ref.Typfmtv(&tgtptr),
		ref.Typfmtv(&tgt), ref.Typfmt(tgttyp))

	var processed bool
	if target, processed = c.checkSource(ctx, source, tgttyp); processed {
		return
	}

	var ret reflect.Value
	var e error
	if ret, e = c.Transform(ctx, source, tgttyp); e == nil {
		if tgtptr.Kind() == reflect.Interface { //nolint:gocritic // no need to switch to 'switch' clause
			tgtptr.Set(ret)
		} else if tgtptr.Kind() == reflect.Ptr {
			tgtptr.Elem().Set(ret)
		} else if tgt.CanSet() {
			tgt.Set(ret)
		} else {
			err = ErrCannotSet.FormatWith(ref.Valfmt(&tgt), ref.Typfmtv(&tgt), ref.Valfmt(&ret), ref.Typfmtv(&ret))
		}
		dbglog.Log("  tgt: %v (ret = %v)", ref.Valfmt(&tgt), ref.Valfmt(&ret))
		return
	}

	dbglog.Log("  Transform() failed: %v", e)
	dbglog.Log("              trying to postCopyTo()")
	err = c.postCopyTo(ctx, source, target)
	return
}

//nolint:lll //keep it
func (c *fromDurationConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	if source.IsValid() {
		// var processed bool
		// if processed, target, err = c.preprocess(ctx, source, targetType); processed {
		//	return
		// }

		var processed bool
		if target, processed = c.checkSource(ctx, source, targetType); processed {
			return
		}

		switch k := targetType.Kind(); k { //nolint:exhaustive //no need
		case reflect.Bool:
			target = rToBool(source)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			target, err = rToInteger(source, targetType)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			target, err = rToUInteger(source, targetType)

		case reflect.Uintptr:
			target = rToUIntegerHex(source, targetType)
		// case reflect.UnsafePointer:
		//	// target = rToUIntegerHex(source, targetType)
		//	err = errors.InvalidArgument
		// case reflect.Ptr:
		//	//target = rToUIntegerHex(source, targetType)
		//	err = errors.InvalidArgument

		case reflect.Float32, reflect.Float64:
			target, err = rToFloat(source, targetType)
		case reflect.Complex64, reflect.Complex128:
			target, err = rToComplex(source, targetType)

		case reflect.String:
			target, err = rToString(source, targetType)

		// reflect.Array
		// reflect.Chan
		// reflect.Func
		// reflect.Interface
		// reflect.Map
		// reflect.Slice
		// reflect.Struct

		default:
			target, err = c.convertToOrZeroTarget(ctx, source, targetType)
		}
	} else {
		target, err = c.convertToOrZeroTarget(ctx, source, targetType)
	}
	return
}

//nolint:lll //keep it
func (c *fromDurationConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	if sk := source.Kind(); sk == reflect.Int64 {
		if yes = source.Name() == "Duration" && source.PkgPath() == timeConstString; yes {
			ctx = &ValueConverterContext{params}
		}
	}
	return
}

//

//

//

type toDurationConverter struct{ toConverterBase }

func (c *toDurationConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	// tgtType := target.Type()
	tgt, tgtptr := ref.Rdecode(target)
	tgtType := c.safeType(tgt, tgtptr) // because tgt might be invalid, so we fetch tgt type via its pointer
	dbglog.Log("  target: %v (%v), tgtptr: %v, tgt: %v, tgttyp: %v",
		ref.Typfmtv(&target), ref.Typfmt(target.Type()), ref.Typfmtv(&tgtptr),
		ref.Typfmtv(&tgt), ref.Typfmt(tgtType))

	if processed := c.checkTargetLite(ctx, tgt, tgtType); processed {
		// tgtptr.Set(ret)
		return
	}

	if ret, e := c.Transform(ctx, source, tgtType); e == nil {
		target.Set(ret)
	} else if ctx.isGroupedFlagOKDeeply(cms.ClearIfInvalid) {
		err = c.fallback(target)
	}
	return
}

//nolint:lll //keep it
func (c *toDurationConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	if source.IsValid() { //nolint:nestif,gocritic // no need to switch to 'switch' clause
		var processed bool
		if target, processed = c.checkSource(ctx, source, targetType); processed {
			return
		}

		switch k := source.Kind(); k { //nolint:exhaustive //no need
		case reflect.Bool:
			if source.Bool() {
				target = reflect.ValueOf(1 * time.Nanosecond)
			} else {
				target = reflect.ValueOf(0 * time.Second)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			target = reflect.ValueOf(time.Duration(source.Int()))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			target = reflect.ValueOf(time.Duration(int64(source.Uint())))

		case reflect.Uintptr:
			target = reflect.ValueOf(time.Duration(int64(syscalls.UintptrToUint(source.Pointer()))))
		// case reflect.UnsafePointer:
		//	// target = rToUIntegerHex(source, targetType)
		//	err = errors.InvalidArgument
		// case reflect.Ptr:
		//	//target = rToUIntegerHex(source, targetType)
		//	err = errors.InvalidArgument

		case reflect.Float32, reflect.Float64:
			target = reflect.ValueOf(time.Duration(int64(source.Float())))
		case reflect.Complex64, reflect.Complex128:
			target = reflect.ValueOf(time.Duration(int64(real(source.Complex()))))

		case reflect.String:
			var dur time.Duration
			dur, err = times.ParseDuration(source.String())
			if err == nil {
				target = reflect.ValueOf(dur)
			}

		// reflect.Array
		// reflect.Chan
		// reflect.Func
		// reflect.Interface
		// reflect.Map
		// reflect.Slice
		// reflect.Struct

		default:
			err = ErrCannotConvertTo.FormatWith(source, ref.Typfmtv(&source), targetType, targetType.Kind())
		}
	} else if ctx.isGroupedFlagOKDeeply(cms.ClearIfInvalid) {
		target = reflect.Zero(targetType)
	} else {
		err = errors.New("source (%v) is invalid", ref.Valfmt(&source))
	}
	return
}

//nolint:lll //keep it
func (c *toDurationConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	if tk := target.Kind(); tk == reflect.Int64 {
		if yes = target.Name() == "Duration" && target.PkgPath() == timeConstString; yes {
			ctx = &ValueConverterContext{params}
		}
	}
	return
}

//

//

type toFuncConverter struct{ fromConverterBase }

func copyToFuncImpl(controller *cpController, source, target reflect.Value, targetType reflect.Type) (err error) {
	var presets []typ.Any
	if controller != nil {
		presets = controller.funcInputs
	}
	if targetType.NumIn() == len(presets)+1 {
		var args []reflect.Value
		for _, in := range presets {
			args = append(args, reflect.ValueOf(in))
		}
		args = append(args, source)

		res := target.Call(args)
		if len(res) > 0 {
			last := res[len(res)-1]
			if ref.Iserrortype(targetType.Out(len(res)-1)) && !ref.IsNil(last) {
				err = last.Interface().(error) //nolint:errcheck //no need
			}
		}
	}
	return
}

// processUnexportedField try to set newval into target if it's an unexported field.
//
//nolint:lll //keep it
func (c *toFuncConverter) processUnexportedField(ctx *ValueConverterContext, target, newval reflect.Value) (processed bool) {
	if ctx == nil || ctx.Params == nil {
		return
	}
	processed = ctx.Params.processUnexportedField(target, newval)
	return
}

func (c *toFuncConverter) copyTo(ctx *ValueConverterContext, source, src, tgt, tsetter reflect.Value) (err error) {
	if ctx.isGroupedFlagOKDeeply(cms.Ignore) {
		return
	}

	tgttyp := tgt.Type()
	dbglog.Log("  copyTo: src: %v, tgt: %v,", ref.Typfmtv(&src), ref.Typfmt(tgttyp))

	if k := src.Kind(); k != reflect.Func && ctx.IsPassSourceToTargetFunction() {
		var controller *cpController
		if ctx.Params != nil && ctx.controller != nil {
			controller = ctx.controller
		}
		err = copyToFuncImpl(controller, source, tgt, tgttyp)
	} else if k == reflect.Func {
		if !c.processUnexportedField(ctx, tgt, src) {
			tsetter.Set(src)
		}
		dbglog.Log("    function pointer copied: %v (%v) -> %v", source.Kind(), source.Interface(), tgt.Kind())
	}
	return
}

func (c *toFuncConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	src := ref.Rdecodesimple(source)
	tgt, tgtptr := ref.Rdecode(target)
	tgtType := c.safeType(tgt, tgtptr) // because tgt might be invalid, so we fetch tgt type via its pointer
	// Log("  CopyTo: src: %v, tgt: %v,", typfmtv(&src), typfmt(tgtType))
	dbglog.Log("  target: %v (%v), tgtptr: %v, tgt: %v, tgttyp: %v",
		ref.Typfmtv(&target), ref.Typfmt(target.Type()), ref.Typfmtv(&tgtptr),
		ref.Typfmtv(&tgt), ref.Typfmt(tgtType))

	if processed := c.checkTargetLite(ctx, tgt, tgtType); processed {
		// tgtptr.Set(ret)
		return
	}

	err = c.copyTo(ctx, source, src, tgt, tgtptr)
	return
}

// func (c *toFuncConverter) Transform(ctx *ValueConverterContext, source reflect.Value,
//	 targetType reflect.Type) (target reflect.Value, err error) {
//
//	target = reflect.New(targetType).Elem()
//
//	src := rdecodesimple(source)
//	tgt, tgtptr := rdecode(target)
//
//	var processed bool
//	if target, processed = c.checkSource(ctx, source, targetType); processed {
//		return
//	}
//
//	err = c.copyTo(ctx, source, src, tgt, tgtptr)
//	return
// }

func (c *toFuncConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	if tk := target.Kind(); tk == reflect.Func {
		yes, ctx = true, &ValueConverterContext{params}
	}
	return
}

//

type fromFuncConverter struct{ fromConverterBase }

func (c *fromFuncConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	// tsetter might not be equal to tgt when:
	//    target represents -> (ptr - interface{} - bool)
	// such as:
	//    var a interface{} = true
	//    var target = reflect.ValueOf(&a)
	//    tgt, tsetter := rdecodesimple(target), rindirect(target)
	//    assertNotEqual(tgt, tsetter)
	//    // in this case, tsetter represents 'a' and tgt represents
	//    // 'decoded bool(a)'.
	//
	src := ref.Rdecodesimple(source)
	tgt, tgtptr := ref.Rdecode(target)
	tgtType := c.safeType(tgt, tgtptr)
	// dbglog.Log("  CopyTo: src: %v, tgt: %v, tsetter: %v", typfmtv(&src), typfmt(tgttyp), typfmtv(&tsetter))
	dbglog.Log("  target: %v (%v), tgtptr: %v, tgt: %v, tgttyp: %v",
		ref.Typfmtv(&target), ref.Typfmt(target.Type()), ref.Typfmtv(&tgtptr),
		ref.Typfmtv(&tgt), ref.Typfmt(tgtType))

	if processed := c.checkTargetLite(ctx, tgt, tgtType); processed {
		// target.Set(ret)
		return
	}

	if k := tgtType.Kind(); k != reflect.Func && ctx.IsCopyFunctionResultToTarget() {
		err = c.funcResultToTarget(ctx, src, target)
		return
	} else if k == reflect.Func {
		if !c.processUnexportedField(ctx, tgt, src) {
			tgtptr.Set(src)
		}
		dbglog.Log("    function pointer copied: %v (%v) -> %v", source.Kind(), source.Interface(), target.Kind())
	}

	// if ret, e := c.Transform(ctx, src, tgttyp); e == nil {
	//	if !target.IsValid() || isZero(target) {
	//		return errors.New("cannot set to zero or invalid target")
	//	}
	//	if canConvert(&ret, tgttyp) {
	//		nv := ret.Convert(tgttyp)
	//		if c.processUnexportedField(ctx, tgt, nv) {
	//			return
	//		}
	//		tsetter.Set(nv)
	//	}
	// }
	return
}

//nolint:lll,gocognit //keep it
func (c *fromFuncConverter) funcResultToTarget(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	sourceType := source.Type()
	var presetInArgsLen int
	var ok bool
	controllerIsValid := ctx != nil && ctx.Params != nil && ctx.Params.controller != nil
	if controllerIsValid {
		presetInArgsLen = len(ctx.controller.funcInputs)
	}
	if sourceType.NumIn() == presetInArgsLen { //nolint:nestif //keep it
		numOutSrc := sourceType.NumOut()
		if numOutSrc > 0 {
			srcResults := source.Call([]reflect.Value{})

			results := srcResults
			lastoutargtype := sourceType.Out(sourceType.NumOut() - 1)
			ok = ref.Iserrortype(lastoutargtype)
			if ok {
				v := results[len(results)-1].Interface()
				err, _ = v.(error)
				if err != nil {
					return
				}
				results = results[:len(results)-1]
			}

			if len(results) > 0 {
				if controllerIsValid {
					// if tk := target.Kind(); tk == reflect.Ptr && isNil(target) {}
					err = ctx.controller.copyTo(ctx.Params, results[0], target)
					return
				}

				// target, err = c.expandResults(ctx, sourceType, targetType, results)
				err = errors.New("expecting ctx.Params.controller is valid object ptr")
				return
			}
		}
	}
	//nolint:lll //keep it
	err = errors.New("unmatched number of function return and preset input args: function needs %v params but preset %v input args", sourceType.NumIn(), presetInArgsLen)
	return
}

// // processUnexportedField try to set newval into target if it's an unexported field
// func (c *fromFuncConverter) processUnexportedField(ctx *ValueConverterContext, target,
//	newval reflect.Value) (processed bool) {
//	if ctx == nil || ctx.Params == nil {
//		return
//	}
//	processed = ctx.Params.processUnexportedField(target, newval)
//	return
// }

//nolint:lll //keep it
func (c *fromFuncConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	var processed bool
	if target, processed = c.checkSource(ctx, source, targetType); processed {
		return
	}

	target = reflect.New(targetType).Elem()
	err = c.CopyTo(ctx, source, target)

	// src, tgt, tgttyp := rdecodesimple(source), rdecodesimple(target), rdecodetypesimple(targetType)
	// Log("  Transform: src: %v, tgt: %v", typfmtv(&src), typfmt(tgttyp))
	// if k := tgttyp.Kind(); k != reflect.Func && ctx.IsCopyFunctionResultToTarget() {
	//	target, err = c.funcResultToField(ctx, src, tgttyp)
	//	return
	//
	// } else if k == reflect.Func {
	//
	//	if c.processUnexportedField(ctx, tgt, src) {
	//		ptr := source.Pointer()
	//		target.SetPointer(unsafe.Pointer(ptr))
	//	}
	//	Log("    function pointer copied: %v (%v) -> %v", source.Kind(), source.Interface(), target.Kind())
	// }
	return
}

// func (c *fromFuncConverter) funcResultToField(ctx *ValueConverterContext, source reflect.Value,
// 	targetType reflect.Type) (target reflect.Value, err error) {
//	sourceType := source.Type()
//	var presetInArgsLen int
//	var ok bool
//	var controllerIsValid = ctx != nil && ctx.Params != nil && ctx.Params.controller != nil
//	if controllerIsValid {
//		presetInArgsLen = len(ctx.controller.funcInputs)
//	}
//	if sourceType.NumIn() == presetInArgsLen {
//		numOutSrc := sourceType.NumOut()
//		if numOutSrc > 0 {
//			srcResults := source.Call([]reflect.Value{})
//
//			results := srcResults
//			lastoutargtype := sourceType.Out(sourceType.NumOut() - 1)
//			ok = iserrortype(lastoutargtype)
//			if ok {
//				err, ok = results[len(results)-1].Interface().(error)
//				if err != nil {
//					return
//				}
//				results = results[:len(results)-1]
//			}
//
//			if len(results) > 0 {
//				// slice,map,struct
//				// scalar
//
//				target, err = c.expandResults(ctx, sourceType, targetType, results)
//			}
//		}
//	} else {
//		err = errors.New("unmatched number of function return and preset input args: function needs
// %v params but preset %v input args", sourceType.NumIn(), presetInArgsLen)
//	}
//	return
// }
//
// func (c *fromFuncConverter) expandResults(ctx *ValueConverterContext, sourceType, targetType
// reflect.Type, results []reflect.Value) (target reflect.Value, err error) {
//	//srclen := len(results)
//	switch kind := targetType.Kind(); kind {
//	case reflect.Bool:
//		target = rToBool(results[0])
//	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
//		target, err = rToInteger(results[0], targetType)
//	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
//		target, err = rToUInteger(results[0], targetType)
//	case reflect.Float32, reflect.Float64:
//		target, err = rToFloat(results[0], targetType)
//	case reflect.Complex64, reflect.Complex128:
//		target, err = rToComplex(results[0], targetType)
//	case reflect.String:
//		{
//			var processed bool
//			if processed, target, err = ctx.Preprocess(results[0], targetType, c); processed {
//				return
//			}
//		}
//		target, err = rToString(results[0], targetType)
//
//	case reflect.Array:
//		target, err = rToArray(ctx, results[0], targetType, -1)
//	case reflect.Slice:
//		target, err = rToSlice(ctx, results[0], targetType, -1)
//	case reflect.Map:
//		target, err = rToMap(ctx, results[0], sourceType, targetType)
//	case reflect.Struct:
//		target, err = rToStruct(ctx, results[0], sourceType, targetType)
//	case reflect.Func:
//		target, err = rToFunc(ctx, results[0], sourceType, targetType)
//
//	case reflect.Interface, reflect.Ptr, reflect.Chan:
//		if results[0].Type().ConvertibleTo(targetType) {
//			target = results[0].Convert(targetType)
//		} else {
//			err = errCannotConvertTo.FormatWith(typfmt(results[0].Type()), typfmt(targetType))
//		}
//
//	case reflect.UnsafePointer:
//		err = errCannotConvertTo.FormatWith(typfmt(results[0].Type()), typfmt(targetType))
//	}
//	return
// }

func (c *fromFuncConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	if sk := source.Kind(); sk == reflect.Func {
		yes, ctx = true, &ValueConverterContext{params}
	}
	return
}
