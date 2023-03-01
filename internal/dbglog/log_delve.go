//go:build delve || verbose
// +build delve verbose

package dbglog

import (
	"github.com/hedzr/log"
	"github.com/hedzr/log/color"
)

const LogValid bool = true

// Log will print formatted message while build-tags `delve` or `verbose` present.
//
// The flag dbglog.LogValid identify that state.
func Log(format string, args ...interface{}) { //nolint:goprintffuncname //no
	log.Skip(0).Infof("%v", color.ToDim(format, args...)) // is there a `log` bug? so Skip(0) is a must-have rather than Skip(1), because stdLogger will detect how many frames should be skipped
	// color.Dim(format, args...)
}

func Colored(clr color.Color, format string, args ...interface{}) { //nolint:goprintffuncname //so what
	log.Skip(0).Infof("%v", color.ToColor(clr, format, args...))
}
