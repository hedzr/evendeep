package diff

import (
	"reflect"
	"time"
)

type timeComparer struct{}

func (c *timeComparer) Match(typ reflect.Type) bool {
	return typ.String() == "time.Time"
}

func (c *timeComparer) Equal(ctx Context, lhs, rhs reflect.Value, path Path) (equal bool) {
	aTime := lhs.Interface().(time.Time)
	bTime := rhs.Interface().(time.Time)
	if equal = aTime.Equal(bTime); !equal {
		ctx.PutModified(ctx.PutPath(path), Update{Old: aTime.String(), New: bTime.String(), Typ: typfmtlite(&lhs)})
	}
	return
}
