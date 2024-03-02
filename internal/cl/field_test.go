package cl

import (
	"reflect"
	"testing"
)

type T1 struct {
	test int
}

func TestGetUnexportedField(t *testing.T) {
	t1 := &T1{1}
	rv := reflect.ValueOf(t1)
	rt := reflect.TypeOf(t1)
	rs := rt.Elem().Field(0)

	fld := &Field{
		Type:  rs,
		Value: rv.Elem().Field(0),
	}

	t.Log(fld.GetUnexportedField())
	fld.SetUnexportedField(2)
	t.Log(fld.GetUnexportedField())
	if t1.test != 2 {
		t.Fail()
	}

	tri := 3
	SetUnexportedFieldIfMap(fld.Value, rv, reflect.ValueOf(tri))
	t.Log(fld.GetUnexportedField())
	if t1.test != 3 {
		t.Fail()
	}
}

func TestFormatComplex(t *testing.T) {
	for _, c := range []struct {
		src  complex128
		fmt  byte
		prec int
		tgt  string
	}{
		{3 + 4i, 'g', -1, "(3+4i)"},
	} {
		actual := FormatComplex(c.src, c.fmt, c.prec, 128)
		if actual != c.tgt {
			t.Fatalf("expect %q, but got %q", c.tgt, actual)
		}

		src, err := ParseComplex(actual)
		if src != c.src || err != nil {
			t.Fatalf("expect %v, but got %v | err=%v", c.src, src, err)
		}
	}
}
