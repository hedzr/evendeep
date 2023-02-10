//go:build !delve && !verbose
// +build !delve,!verbose

package dbglog

const LogValid bool = false //nolint:gochecknoglobals //i know that

func Log(format string, args ...interface{}) { //nolint:goprintffuncname //no

}
