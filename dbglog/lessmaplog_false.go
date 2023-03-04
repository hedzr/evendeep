//go:build !moremaplog
// +build !moremaplog

package dbglog

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
