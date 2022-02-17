//go:build delve || verbose
// +build delve verbose

package deepcopy

import "github.com/hedzr/log"

func functorLog(format string, args ...interface{}) {
	log.Skip(0).Infof(format, args...) //
}