//go:build !delve && !verbose
// +build !delve,!verbose

package dbglog

var LogValid bool //nolint:gochecknoglobals //i know that

func Log(format string, args ...interface{}) { //nolint:goprintffuncname //no

}
