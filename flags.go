package deepcopy

// Flags is an efficient manager for a group of CopyMergeStrategy items.
type Flags map[CopyMergeStrategy]bool

func newFlags(ftf ...CopyMergeStrategy) Flags {
	onceInitFieldTagsFlags()
	flags := make(map[CopyMergeStrategy]bool)
	for _, f := range ftf {
		flags[f] = true
	}
	return flags
}

func (flags Flags) withFlags(flg ...CopyMergeStrategy) Flags {
	for _, f := range flg {
		flags[f] = true
	}
	return flags
}

func (flags Flags) isFlagOK(ftf CopyMergeStrategy) bool {
	return flags[ftf]
}

func (flags Flags) testGroupedFlag(ftf CopyMergeStrategy) (result CopyMergeStrategy) {
	var ok bool
	result = InvalidStrategy

	if _, ok = flags[ftf]; ok {
		result = ftf
		return
	}

	if _, ok = mKnownFieldTagFlagsConflictLeaders[ftf]; ok {
		// ftf is a leader
		ok = false
		for f := range mKnownFieldTagFlagsConflict[ftf] {
			if _, ok = flags[f]; ok {
				result = f
				break
			}
		}
		if !ok {
			result = ftf
		}
	} else {
		leader := InvalidStrategy
		found := false
		for f := range mKnownFieldTagFlagsConflict[ftf] {
			if _, ok = mKnownFieldTagFlagsConflictLeaders[f]; ok {
				leader = f
			}
			if _, ok = flags[f]; ok {
				result = f
				found = true
				break
			}
		}
		if !found {
			result = leader
		}
	}
	return
}

// isGroupedFlagOK _
func (flags Flags) isGroupedFlagOK(ftf CopyMergeStrategy) (ok bool) {
	if flags == nil {
		return
	}

	if _, ok = flags[ftf]; ok {
		return
	}

	//var vm map[CopyMergeStrategy]struct{}
	if vm, ok1 := mKnownFieldTagFlagsConflict[ftf]; ok1 {
		// find the default one (named as `leader` from a radio-group of flags
		leader := InvalidStrategy
		if _, ok1 = mKnownFieldTagFlagsConflictLeaders[ftf]; ok1 {
			leader = ftf
		}

		for f := range vm {
			if _, ok1 = mKnownFieldTagFlagsConflictLeaders[f]; ok1 {
				leader = f
			}
		}
		found := false
		for f := range vm {
			if _, found = flags[f]; found {
				break
			}
		}
		if !found {
			// while the testing `ftf` is a leader in certain a
			// radio-group, and any of the flags of the group are not
			// in flags map set, that assume the leader is exists.
			//
			// For example, when checking ftf = SliceCopy and any one
			// of SliceXXX not in flags, though the test is ok.
			if ftf == leader {
				ok = true
			}
		}
	}
	return
}

func (flags Flags) isAnyFlagsOK(ftf ...CopyMergeStrategy) bool {
	for _, f := range ftf {
		if flags[f] {
			return true
		}
	}
	return false
}

func (flags Flags) isAllFlagsOK(ftf ...CopyMergeStrategy) bool {
	for _, f := range ftf {
		if !flags[f] {
			return false
		}
	}
	return true
}
