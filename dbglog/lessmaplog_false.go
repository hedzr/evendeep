//go:build !moremaplog
// +build !moremaplog

package dbglog

const MoreMapLog = false

var logValid = true

func DisableLog() func()    { var sav = logValid; logValid = MoreMapLog; return func() { logValid = sav } }
func ChildLogEnabled() bool { return logValid }
