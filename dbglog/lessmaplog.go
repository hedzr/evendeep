//go:build moremaplog
// +build moremaplog

package dbglog

const MoreMapLog = true

func DisableLog() func()    { return func() {} }
func ChildLogEnabled() bool { return true }
