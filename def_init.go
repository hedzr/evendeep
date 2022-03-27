package deepcopy

import "sync"

var onceInitRoutines sync.Once
var otherRoutines []func()

func init() {
	onceInitRoutines.Do(func() {

		initConverters()

		initGlobalOperators()

	})
}

func initGlobalOperators() {
	DefaultCopyController = newDeepCopier()
	defaultCloneController = newCloner()
}
