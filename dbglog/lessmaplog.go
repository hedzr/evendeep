//go:build moremaplog
// +build moremaplog

package dbglog

const MoreMapLog = true

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
// If you want to enable/disable dbglog.Log manually, try SetLogEnabled and SetLogDisabled.
func DisableLog() func()    { return func() {} }
func ChildLogEnabled() bool { return true }
