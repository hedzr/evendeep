package cl

import (
	"reflect"
	"unsafe"

	"github.com/hedzr/evendeep/typ"
)

// Field wraps a struct field with its type and value pair.
type Field struct {
	Type  reflect.StructField
	Value reflect.Value
}

// GetUnexportedField returns the value of the unexported field.
func (field *Field) GetUnexportedField() typ.Any {
	return GetUnexportedField(field.Value).Interface()
}

// SetUnexportedField puts a new value into the unexported field.
func (field *Field) SetUnexportedField(value typ.Any) {
	SetUnexportedField(field.Value, reflect.ValueOf(value))
}

// // GetUnexportedField return the value of the unexported field
// func GetUnexportedField(field reflect.Value) typ.Any {
// 	return getUnexportedField(field).Interface()
// }

// GetUnexportedField return the value of the unexported field.
func GetUnexportedField(field reflect.Value) reflect.Value {
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
}

// // SetUnexportedField puts a new value into the unexported field
// func SetUnexportedField(field reflect.Value, value typ.Any) {
// 	setUnexportedField(field, reflect.ValueOf(value))
// }

// SetUnexportedField puts a new value into the unexported field.
func SetUnexportedField(field, value reflect.Value) {
	ptr := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr()))
	dat := ptr.Elem()
	dat.Set(value)
}
