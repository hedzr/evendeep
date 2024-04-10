package diff

import (
	"testing"
)

func TestUpdate_String(t *testing.T) {
	u := Update{
		Old: nil,
		New: nil,
		Typ: "",
	}
	t.Logf("u: %v", u)

	u = Update{
		Old: 1,
		New: nil,
		Typ: "",
	}
	t.Logf("u: %v", u)

	u = Update{
		Old: 1,
		New: 2,
		Typ: "",
	}
	t.Logf("u: %v", u)

	u = Update{
		Old: 1,
		New: int64(2),
		Typ: "int64",
	}
	t.Logf("u: %v", u)

	u = Update{
		Old: 1,
		New: 2,
		Typ: "int",
	}
	t.Logf("u: %v", u)
}
