package deepcopy

import (
	"bytes"
	"fmt"
	"gopkg.in/hedzr/errors.v3"
	"reflect"
)

func defaultValueConverters() []ValueConverter {
	return []ValueConverter{
		&bytesBufferConverter{},
		&toStringConverter{},
		&fromStringConverter{},
	}
}

func defaultValueCopiers() []ValueCopier {
	return []ValueCopier{
		&bytesBufferConverter{},
		&toStringConverter{},
		&fromStringConverter{},
	}
}

type bytesBufferConverter struct{}

func (c *bytesBufferConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	//TO/DO implement me
	//panic("implement me")
	from := source.Interface().(bytes.Buffer)
	var to bytes.Buffer
	to.Write(from.Bytes())
	target = reflect.ValueOf(&to)
	return
}

func (c *bytesBufferConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	from, to := source.Interface().(bytes.Buffer), target.Interface().(bytes.Buffer)
	to.Reset()
	to.Write(from.Bytes())
	return
}

func (c *bytesBufferConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	//st.PkgPath() . st.Name()
	if yes = source.Kind() == reflect.Struct && source.String() == "bytes.Buffer"; yes {
		ctx = &ValueConverterContext{}
		functorLog("    src: %v, tgt: %v | Matched", source, target)
	} else {
		functorLog("    src: %v, tgt: %v", source, target)
	}
	return
}

type toStringConverter struct{}

func (c *toStringConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	if ret, e := c.Transform(ctx, source, target.Type()); e == nil {
		target.Set(ret)
		return
	}

	if source.IsValid() {
		if source.CanConvert(target.Type()) {
			nv := source.Convert(target.Type())
			target.Set(nv)
		} else {
			nv := fmt.Sprintf("%v", source.Interface())
			target.Set(reflect.ValueOf(nv))
		}
	}
	return
}

// Transform will transform source type (bool, int, ...) to target string
func (c *toStringConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	if source.IsValid() {
		switch k := source.Kind(); k {
		case reflect.Bool:
			target = rForBool(source)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			target = rForInteger(source)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			target = rForUInteger(source)

		case reflect.Uintptr:
			target = rForUIntegerHex(source.Pointer())
		case reflect.UnsafePointer:
			target = rForUIntegerHex(uintptr(source.UnsafePointer()))
		case reflect.Pointer:
			target = rForUIntegerHex(source.Pointer())

		case reflect.Float32, reflect.Float64:
			target = rForFloat(source)
		case reflect.Complex64, reflect.Complex128:
			target = rForComplex(source)

		case reflect.String:
			target = source

		//reflect.Array
		//reflect.Chan
		//reflect.Func
		//reflect.Interface
		//reflect.Map
		//reflect.Slice
		//reflect.Struct

		default:
			if source.CanConvert(targetType) {
				nv := source.Convert(targetType)
				// target.Set(nv)
				target = nv
			} else {
				nv := fmt.Sprintf("%v", source.Interface())
				// target.Set(reflect.ValueOf(nv))
				target = reflect.ValueOf(nv)
			}
		}
	}
	return
}

func (c *toStringConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	if yes = target.Kind() == reflect.String; yes {
		ctx = &ValueConverterContext{}
	}
	return
}

type fromStringConverter struct{}

func (c *fromStringConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	if ret, e := c.Transform(ctx, source, target.Type()); e == nil {
		target.Set(ret)
		return
	}

	if source.IsValid() {
		if source.CanConvert(target.Type()) {
			nv := source.Convert(target.Type())
			target.Set(nv)
		} else {
			nv := fmt.Sprintf("%v", source.Interface())
			target.Set(reflect.ValueOf(nv))
		}
	}
	return
}

// Transform will transform source string to target type (bool, int, ...)
func (c *fromStringConverter) Transform(ctx *ValueConverterContext, source reflect.Value, targetType reflect.Type) (target reflect.Value, err error) {
	if source.IsValid() {
		switch k := targetType.Kind(); k {
		case reflect.Bool:
			target = rToBool(source)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			target, err = rToInteger(source, targetType)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			target, err = rToUInteger(source, targetType)

		case reflect.Uintptr:
			target = rToUIntegerHex(source, targetType)
		case reflect.UnsafePointer:
			// target = rToUIntegerHex(source, targetType)
			err = errors.InvalidArgument
		case reflect.Pointer:
			//target = rToUIntegerHex(source, targetType)
			err = errors.InvalidArgument

		case reflect.Float32, reflect.Float64:
			target, err = rToFloat(source, targetType)
		case reflect.Complex64, reflect.Complex128:
			target, err = rToComplex(source, targetType)

		case reflect.String:
			target = source

		//reflect.Array
		//reflect.Chan
		//reflect.Func
		//reflect.Interface
		//reflect.Map
		//reflect.Slice
		//reflect.Struct

		default:
			if source.CanConvert(targetType) {
				nv := source.Convert(targetType)
				// target.Set(nv)
				target = nv
			} else {
				nv := fmt.Sprintf("%v", source.Interface())
				// target.Set(reflect.ValueOf(nv))
				target = reflect.ValueOf(nv)
			}
		}
	}
	return
}

func (c *fromStringConverter) Match(params *Params, source, target reflect.Type) (ctx *ValueConverterContext, yes bool) {
	if yes = source.Kind() == reflect.String; yes {
		ctx = &ValueConverterContext{}
	}
	return
}
