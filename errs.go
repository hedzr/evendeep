package evendeep

import "gopkg.in/hedzr/errors.v3"

var (
	// ErrUnknownState error
	ErrUnknownState = errors.New("unknown state, cannot copy to")

	// ErrCannotSet error
	ErrCannotSet = errors.New("cannot set: %v (%v) -> %v (%v)")

	// ErrCannotCopy error
	ErrCannotCopy = errors.New("cannot copy: %v (%v) -> %v (%v)")

	// ErrCannotConvertTo error
	ErrCannotConvertTo = errors.New("cannot convert/set: %v (%v) -> %v (%v)")

	// ErrShouldFallback tells the caller please continue its
	// internal process.
	// The error would be used in your callback function. For
	// instance, you could return it in a target-setter (see
	// also WithTargetValueSetter()) to ask the Copier do a
	// standard processing, typically that will set the field
	// with reflection.
	ErrShouldFallback = errors.New("fallback to evendeep internals")
)
