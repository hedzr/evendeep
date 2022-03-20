//go:build delve || verbose
// +build delve verbose

package dbglog

import "github.com/hedzr/log"

var LogValid bool = true

func Log(format string, args ...interface{}) {
	log.Skip(0).Infof(format, args...) // i had a `log` bug? so Skip(0) is a must-have rather than Skip(1), because stdLogger will detect how many frames should be skipped
}
