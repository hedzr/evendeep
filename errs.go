package deepcopy

import "gopkg.in/hedzr/errors.v3"

var (
	// ErrUnknownState error
	ErrUnknownState = errors.New("unknown state, cannot copy to")

	// ErrCannotSet error
	ErrCannotSet = errors.New("cannot set: %v (%v) -> %v (%v)")

	errCannotConvertTo = errors.New("cannot convert: %v -> %v")
)
