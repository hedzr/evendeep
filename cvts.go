package deepcopy

import (
	"bytes"
	"reflect"
)

func defaultValueConverters() []ValueConverter {
	return []ValueConverter{
		&bytesBufferConverter{},
	}
}

func defaultValueCopiers() []ValueCopier {
	return []ValueCopier{
		&bytesBufferConverter{},
	}
}

type bytesBufferConverter struct{}

func (c *bytesBufferConverter) Transform(ctx *ValueConverterContext, source reflect.Value) (target reflect.Value, err error) {
	//TODO implement me
	panic("implement me")
}

func (c *bytesBufferConverter) CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error) {
	from, to := source.Interface().(bytes.Buffer), target.Interface().(bytes.Buffer)
	to.Reset()
	to.Write(from.Bytes())
	return
}

func (c *bytesBufferConverter) Match(params *paramsPackage, source, target reflect.Value) (ctx *ValueConverterContext, yes bool) {
	st := source.Type()
	if yes = st.Kind() == reflect.Struct && st.String() == "bytes.Buffer"; yes {
		ctx = &ValueConverterContext{}
	}
	return
}
