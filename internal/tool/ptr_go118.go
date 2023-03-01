//go:build go1.18
// +build go1.18

package tool

import "reflect"

// PointerTo returns the pointer type of element t.
// For example, if t represents type Foo, PointerTo(t) represents *Foo.
func PointerTo(t reflect.Type) reflect.Type {
	return reflect.PointerTo(t)
}
