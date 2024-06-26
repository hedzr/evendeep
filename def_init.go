package evendeep

import "sync"

var (
	onceInitRoutines sync.Once                                       //nolint:gochecknoglobals //i know that
	otherRoutines    = []func(){initConverters, initGlobalOperators} //nolint:gochecknoglobals //i know that
)

func init() { //nolint:gochecknoinits //don't
	onceInitRoutines.Do(func() {
		// initConverters()
		// initGlobalOperators()
		for _, fn := range otherRoutines {
			if fn != nil {
				fn()
			}
		}
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
