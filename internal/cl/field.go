package cl

import (
	"reflect"
	"unsafe"
)

// Field wraps a struct field with its type and value pair
type Field struct {
	Type  reflect.StructField
	Value reflect.Value
}

// GetUnexportedField returns the value of the unexported field
func (field Field) GetUnexportedField() interface{} {
	return GetUnexportedField(field.Value).Interface()
}

// SetUnexportedField puts a new value into the unexported field
func (field Field) SetUnexportedField(value interface{}) {
	SetUnexportedField(field.Value, reflect.ValueOf(value))
}

// // GetUnexportedField return the value of the unexported field
// func GetUnexportedField(field reflect.Value) interface{} {
// 	return getUnexportedField(field).Interface()
// }

// GetUnexportedField return the value of the unexported field
func GetUnexportedField(field reflect.Value) reflect.Value {
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
}

// // SetUnexportedField puts a new value into the unexported field
// func SetUnexportedField(field reflect.Value, value interface{}) {
// 	setUnexportedField(field, reflect.ValueOf(value))
// }

// SetUnexportedField puts a new value into the unexported field
func SetUnexportedField(field, value reflect.Value) {
	ptr := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr()))
	dat := ptr.Elem()
	dat.Set(value)
}
