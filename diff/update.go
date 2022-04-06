package diff

import (
	"fmt"

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
	return fmt.Sprintf("%#v -> %#v", n.Old, n.New)
}
