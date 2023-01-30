package diff

import (
	"reflect"

	"github.com/hedzr/evendeep/typ"
)

// Comparer interface.
type Comparer interface {
	Match(typ reflect.Type) bool
	Equal(ctx Context, lhs, rhs reflect.Value, path Path) (equal bool)
}

// Context interface.
type Context interface {
	PutAdded(k string, v typ.Any)
	PutRemoved(k string, v typ.Any)
	PutModified(k string, v Update)
	PutPath(path Path, parts ...PathPart) string
}
