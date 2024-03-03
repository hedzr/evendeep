package diff

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/hedzr/evendeep/typ"
)

func TestBytesBufferComparer_Equal(t *testing.T) {
	var bbc bytesBufferComparer
	var ctx ctxS
	var p Path

	rv := func(a []byte) reflect.Value {
		bb := bytes.NewBuffer(a)
		return reflect.ValueOf(bb).Elem()
	}

	if bbc.Equal(&ctx, rv([]byte("hello")), rv([]byte("Hello")), p) {
		t.Fail()
	}
	if bbc.Equal(&ctx, rv([]byte("hello")), rv([]byte("G")), p) {
		t.Fail()
	}
	if !bbc.Equal(&ctx, rv([]byte("hello")), rv([]byte("hello")), p) {
		t.Fail()
	}
}

type ctxS struct{}

func (c *ctxS) PutAdded(k string, v typ.Any) {}

func (c *ctxS) PutRemoved(k string, v typ.Any) {}

func (c *ctxS) PutModified(k string, v Update) {}

func (c *ctxS) PutPath(path Path, parts ...PathPart) string { return "" }

func TestIsEmptyObject(t *testing.T) {
	var a int
	rv := reflect.ValueOf(a)
	t.Logf("Result: %v", isEmptyObject(rv))

	a = 1
	rv = reflect.ValueOf(a)
	t.Logf("Result: %v", isEmptyObject(rv))

	var b *int
	rv = reflect.ValueOf(b)
	t.Logf("Result: %v", isEmptyObject(rv))

	b = &a
	rv = reflect.ValueOf(b)
	t.Logf("Result: %v", isEmptyObject(rv))

	var c []int
	rv = reflect.ValueOf(c)
	t.Logf("Result: %v", isEmptyObject(rv))

	c = []int{1, 2, 3}
	rv = reflect.ValueOf(c)
	t.Logf("Result: %v", isEmptyObject(rv))

	c = c[:0]
	rv = reflect.ValueOf(c)
	t.Logf("Result: %v", isEmptyObject(rv))

	var d map[int]bool
	rv = reflect.ValueOf(d)
	t.Logf("Result: %v", isEmptyObject(rv))
}

func TestIsEmptyStruct(t *testing.T) {
	type aS struct {
		a int
	}

	var a aS
	rv := reflect.ValueOf(a)
	t.Logf("Result: %v / %v", isEmptyStruct(rv), isEmptyStructDeeply(rv))

	var b *aS
	rv = reflect.ValueOf(b)
	t.Logf("Result: %v / %v", isEmptyStruct(rv), isEmptyStructDeeply(rv))
}
