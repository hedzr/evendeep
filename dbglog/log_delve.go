//go:build delve || verbose
// +build delve verbose

package dbglog

import (
	"github.com/hedzr/is/term/color"
	logz "github.com/hedzr/logg/slog"

	"gopkg.in/hedzr/errors.v3"
)

// LogValid shows dbglog.Log is enabled or abandoned
const LogValid bool = true

func SetLogEnabled()  { logValid = true }  // enables dbglog.Log at runtime
func SetLogDisabled() { logValid = false } // disables dbglog.Log at runtime

// DeferVisit moves errors in container ec, and log its via dbglog.Log
func DeferVisit(ec errors.Error, err *error) {
	ec.Defer(err)
	if *err != nil {
		logz.WithSkip(1).Error("FAILED", "error", color.ToColor(color.FgRed, "%+v", *err))
	}
}

// Log will print formatted message while build-tags `delve` or `verbose` present.
//
// The flag dbglog.LogValid identify that state.
func Log(format string, args ...interface{}) { //nolint:goprintffuncname //no
	if logValid {
		logz.WithSkip(1).Info(color.ToDim(format, args...)) // is there a `log` bug? so Skip(0) is a must-have rather than Skip(1), because stdLogger will detect how many frames should be skipped
		// color.Dim(format, args...)
	}
}

func Err(format string, args ...interface{}) { //nolint:goprintffuncname //so what
	logz.WithSkip(1).Error(color.ToColor(color.FgRed, format, args...))
}

func Wrn(format string, args ...interface{}) { //nolint:goprintffuncname //so what
	logz.WithSkip(1).Warn(color.ToColor(color.FgYellow, format, args...))
}

func Colored(clr color.Color, format string, args ...interface{}) { //nolint:goprintffuncname //so what
	logz.WithSkip(1).Success(color.ToColor(clr, format, args...))
}
