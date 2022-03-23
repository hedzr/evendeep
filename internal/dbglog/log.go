//go:build !delve && !verbose
// +build !delve,!verbose

package dbglog

var LogValid bool

func Log(format string, args ...interface{}) {

}
