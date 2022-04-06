package diff

import (
	"bytes"
	"reflect"

	"github.com/hedzr/evendeep/internal/tool"
)

type bytesBufferComparer struct{}

func (c *bytesBufferComparer) Match(typ reflect.Type) bool {
	return typ.String() == "bytes.Buffer"
}

func (c *bytesBufferComparer) Equal(ctx Context, lhs, rhs reflect.Value, path Path) (equal bool) {
	a := lhs.Interface().(bytes.Buffer)
	b := rhs.Interface().(bytes.Buffer)
	if equal = c.equal(a.Bytes(), b.Bytes()); !equal {
		ctx.PutModified(ctx.PutPath(path), Update{Old: a.String(), New: b.String(), Typ: tool.Typfmtvlite(&lhs)})
	}
	return
}

func (c *bytesBufferComparer) equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
