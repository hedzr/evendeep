package diff

import (
	"bytes"
	"reflect"

	"github.com/hedzr/evendeep/ref"
)

type bytesBufferComparer struct{}

func (c *bytesBufferComparer) Match(typ reflect.Type) bool {
	return typ.String() == "bytes.Buffer"
}

func (c *bytesBufferComparer) Equal(ctx Context, lhs, rhs reflect.Value, path Path) (equal bool) {
	var a, b bytes.Buffer
	var ok bool
	if a, ok = lhs.Interface().(bytes.Buffer); !ok {
		return
	}
	if b, ok = rhs.Interface().(bytes.Buffer); !ok {
		return
	}
	if equal = c.equal(a.Bytes(), b.Bytes()); !equal {
		ctx.PutModified(ctx.PutPath(path), Update{Old: a.String(), New: b.String(), Typ: ref.Typfmtvlite(&lhs)})
	}
	return
}

func (c *bytesBufferComparer) equal(a, b []byte) bool { //nolint:revive
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
