//go:build !delve && !verbose
// +build !delve,!verbose

package dbglog

import (
	"github.com/hedzr/log/color"
	"gopkg.in/hedzr/errors.v3"
)

// LogValid shows dbglog.Log is enabled or abandoned
const LogValid bool = false //nolint:gochecknoglobals //i know that

func SetLogEnabled()  {} // enables dbglog.Log at runtime
func SetLogDisabled() {} // disables dbglog.Log at runtime

// DeferVisit moves errors in container ec, and log its via dbglog.Log
func DeferVisit(ec errors.Error, err *error) {
	ec.Defer(err)
}

// Log will print formatted message while build-tags `delve` or `verbose` present.
//
// The flag dbglog.LogValid identify that state.
func Log(format string, args ...interface{}) { //nolint:goprintffuncname //no

}

func Err(format string, args ...interface{}) { //nolint:goprintffuncname //no

}

func Wrn(format string, args ...interface{}) { //nolint:goprintffuncname //so what

}

func Colored(clr color.Color, format string, args ...interface{}) { //nolint:goprintffuncname //so what

}
