package diff

import (
	"github.com/hedzr/evendeep/internal/tool"

	"reflect"
	"time"
)

type timeComparer struct{}

func (c *timeComparer) Match(typ reflect.Type) bool {
	return typ.String() == "time.Time"
}

func (c *timeComparer) Equal(ctx Context, lhs, rhs reflect.Value, path Path) (equal bool) {
	aTime := lhs.Interface().(time.Time) //nolint:errcheck //yes
	bTime := rhs.Interface().(time.Time) //nolint:errcheck //yes
	if equal = aTime.Equal(bTime); !equal {
		ctx.PutModified(ctx.PutPath(path), Update{Old: aTime.String(), New: bTime.String(), Typ: tool.Typfmtvlite(&lhs)})
	}
	return
}
