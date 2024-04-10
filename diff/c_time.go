package diff

import (
	"reflect"
	"time"

	"github.com/hedzr/evendeep/ref"
)

type timeComparer struct{}

func (c *timeComparer) Match(typ reflect.Type) bool {
	return typ.String() == "time.Time"
}

func (c *timeComparer) Equal(ctx Context, lhs, rhs reflect.Value, path Path) (equal bool) {
	var aTime, bTime time.Time
	var ok bool
	if aTime, ok = lhs.Interface().(time.Time); !ok {
		return
	}
	if bTime, ok = rhs.Interface().(time.Time); !ok {
		return
	}
	if equal = aTime.Equal(bTime); !equal {
		ctx.PutModified(ctx.PutPath(path), Update{Old: aTime.String(), New: bTime.String(), Typ: ref.Typfmtvlite(&lhs)})
	}
	return
}
