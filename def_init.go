package deepcopy

import "sync"

var onceInitRoutines sync.Once
var otherRoutines []func()

func init() { //nolint:gochecknoinits
	onceInitRoutines.Do(func() {
		initConverters()
		initGlobalOperators()
	})
}

func initGlobalOperators() {
	DefaultCopyController = newDeepCopier()
	defaultCloneController = newCloner()
}

// ResetDefaultCopyController discards the changes for DefaultCopyController and more.
func ResetDefaultCopyController() {
	initGlobalOperators()
}
