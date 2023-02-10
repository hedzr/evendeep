//go:build delve || verbose
// +build delve verbose

package dbglog

import "github.com/hedzr/log"

const LogValid bool = true

func Log(format string, args ...interface{}) { //nolint:goprintffuncname //no
	log.Skip(0).Infof(format, args...) // is there a `log` bug? so Skip(0) is a must-have rather than Skip(1), because stdLogger will detect how many frames should be skipped
}
