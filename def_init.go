package deepcopy

import "sync"

var onceInitRoutines sync.Once
var otherRoutines []func()

func init() {
	onceInitRoutines.Do(func() {

		initConverters()

		DefaultCopyController = newDeepCopier()
		DefaultCloneController = newCloner()

	})
}
