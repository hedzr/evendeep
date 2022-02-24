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

func (c *bytesBufferConverter) Match(params *Params, source, target reflect.Value) (ctx *ValueConverterContext, yes bool) {
	if source.IsValid() {
		st := source.Type()
		functorLog("    src: %v (%v), tgt: %v", valfmt(&source), st, valfmt(&target))
		//st.PkgPath() . st.Name()
		if yes = st.Kind() == reflect.Struct && st.String() == "bytes.Buffer"; yes {
			ctx = &ValueConverterContext{}
		}
	}
	return
}
