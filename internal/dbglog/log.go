//go:build !delve && !verbose
// +build !delve,!verbose

package dbglog

import (
	"github.com/hedzr/log/color"
)

const LogValid bool = false //nolint:gochecknoglobals //i know that

// Log will print formatted message while build-tags `delve` or `verbose` present.
//
// The flag dbglog.LogValid identify that state.
func Log(format string, args ...interface{}) { //nolint:goprintffuncname //no

}

func Colored(clr color.Color, format string, args ...interface{}) { //nolint:goprintffuncname //so what

}
