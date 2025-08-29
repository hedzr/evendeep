//go:build go1.18
// +build go1.18

// for go1.18+

package cl

// // setUnexportedField puts a new value into the unexported field
// func setUnexportedField(field, value reflect.Value) {
//	ptr := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr()))
//	dat := ptr.Elem()
//	dat.Set(value)
// }
