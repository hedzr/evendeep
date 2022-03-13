package deepcopy

import "gopkg.in/hedzr/errors.v3"

var (
	// ErrUnknownState error
	ErrUnknownState = errors.New("unknown state, cannot copy to")

	// ErrCannotSet error
	ErrCannotSet = errors.New("cannot set: %v (%v) -> %v (%v)")

	// ErrCannotConvertTo error
	ErrCannotConvertTo = errors.New("cannot convert/set: %v (%v) -> %v (%v)")
)
