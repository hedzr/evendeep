package diff

import (
	"fmt"
	"reflect"

	"github.com/hedzr/evendeep/typ"
)

type Update struct {
	Old, New typ.Any // string
	Typ      string
}

func (n Update) String() string {
	if n.Old == nil {
		return fmt.Sprintf("%#v", n.New)
	}
	if n.New == nil {
		return fmt.Sprintf("%#v -> nil", n.Old)
	}

	a, b := reflect.ValueOf(n.Old), reflect.ValueOf(n.New)
	if a.Kind() != b.Kind() {
		return fmt.Sprintf("%#v (%v) -> %#v (%v)", n.Old, a.Kind(), n.New, b.Kind())
	}

	if n.Typ != "" {
		return fmt.Sprintf("%#v -> %#v (typ = %q)", n.Old, n.New, n.Typ)
	}

	return fmt.Sprintf("%#v -> %#v", n.Old, n.New)
}
