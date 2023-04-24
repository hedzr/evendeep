//go:build !moremaplog
// +build !moremaplog

package dbglog

// MoreMapLog shows the buildtag 'moremaplog' is enabled or not.
//
// This flag is used for hedzr/evendeep package so that it can print
// more debug logging messages when copying map struct.
//
// You may borrow the mechanism to dump the more verbose messages for
// debugging purpose.
const MoreMapLog = false

var logValid = true

// DisableLog can be used to disable dbglog.Log at runtime.
//
// It detects and prevent log output if buildtag 'moremaplog' present.
//
// To query the active state by calling ChildLogEnabled.
//
// The best practise for DisableLog is:
//
//    defer dbglog.DisableLog()()
//    evendeep.CopyTo(...) // the verbose logging will be prevent even if buildtag 'verbose' defined.
//
func DisableLog() func()    { var sav = logValid; logValid = MoreMapLog; return func() { logValid = sav } }
func ChildLogEnabled() bool { return logValid }
