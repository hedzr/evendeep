//go:build !go1.18beta1
// +build !go1.18beta1

// for lower than go1.18

package cl

//// setUnexportedField puts a new value into the unexported field
//func setUnexportedField(field, value reflect.Value) {
//	ptr := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr()))
//	dat := ptr.Elem()
//	dat.Set(value)
//}
