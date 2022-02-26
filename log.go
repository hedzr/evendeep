//go:build !delve && !verbose
// +build !delve,!verbose

package deepcopy

var functorLogValid bool

func functorLog(format string, args ...interface{}) {

}
