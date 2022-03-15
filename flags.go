package deepcopy

import "strings"

// Flags is an efficient manager for a group of CopyMergeStrategy items.
type Flags map[CopyMergeStrategy]bool

func newFlags(ftf ...CopyMergeStrategy) Flags {
	lazyInitFieldTagsFlags()
	flags := make(Flags)
	flags.withFlags(ftf...)
	return flags
}

func (flags Flags) String() string {
	var sb, sbfinal strings.Builder

	for fx, ok := range flags {
		if ok {
			if sb.Len() > 0 {
				sb.WriteRune(',')
			}
			sb.WriteString(fx.String())
		}
	}

	sbfinal.WriteRune('[')
	sbfinal.WriteString(sb.String())
	sbfinal.WriteRune(']')

	return sbfinal.String()
}

func (flags Flags) withFlags(flg ...CopyMergeStrategy) Flags {
	for _, f := range flg {
		flags[f] = true
		if m, ok := mKnownFieldTagFlagsConflict[f]; ok {
			for fx := range m {
				if fx != f {
					if _, has := flags[fx]; has {
						flags[fx] = false
					}
				}
			}
		}
	}
	return flags
}

func (flags Flags) isFlagOK(ftf CopyMergeStrategy) bool {
	if flags != nil {
		return flags[ftf]
	}
	return false
}

func (flags Flags) testGroupedFlag(ftf CopyMergeStrategy) (result CopyMergeStrategy) {
	var ok, val bool
	result = InvalidStrategy

	if val, ok = flags[ftf]; ok && val {
		result = ftf
		return
	}

	if _, ok = mKnownFieldTagFlagsConflictLeaders[ftf]; ok {
		// ftf is a leader
		ok = false
		for f := range mKnownFieldTagFlagsConflict[ftf] {
			if val, ok = flags[f]; ok && val {
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
			if val, ok = flags[f]; ok && val {
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

// isGroupedFlagOK test if any of ftf is exists.
//
// If one of ftf is the leader (a.k.a. the first one) of a toggleable
// group (such as map-copy and map-merge), and, any of the group is
// not exists (either map-copy and map-merge), isGroupedFlagOK will
// report true just like map-copy was in Flags.
func (flags Flags) isGroupedFlagOK(ftf ...CopyMergeStrategy) (ok bool) {
	if flags != nil {
		for _, ff := range ftf {
			if _, ok = flags[ff]; ok {
				return
			}
		}
	}

	for _, ff := range ftf {
		if vm, ok1 := mKnownFieldTagFlagsConflict[ff]; ok1 {
			// find the default one (named as `leader` from a radio-group of flags
			leader := InvalidStrategy
			if _, ok1 = mKnownFieldTagFlagsConflictLeaders[ff]; ok1 {
				leader = ff
			}
			for f := range vm {
				if _, ok1 = mKnownFieldTagFlagsConflictLeaders[f]; ok1 {
					leader = f
				}
			}

			var found, val bool
			if flags != nil {
				for f := range vm {
					if val, found = flags[f]; found && val {
						break
					}
				}
			}

			if !found {
				// while the testing `ff` is a leader in certain a
				// radio-group, and any of the flags of the group are not
				// in flags map set, that assume the leader is exists.
				//
				// For example, when checking ftf = SliceCopy and any one
				// of SliceXXX not in flags, though the test is ok.
				if ff == leader {
					ok = true
				}
			}
		}
	}
	return
}

func (flags Flags) isAnyFlagsOK(ftf ...CopyMergeStrategy) bool {
	if flags != nil {
		for _, f := range ftf {
			if val, ok := flags[f]; val && ok {
				return true
			}
		}
	}
	return false
}

func (flags Flags) isAllFlagsOK(ftf ...CopyMergeStrategy) bool {
	if flags != nil {
		for _, f := range ftf {
			if val, ok := flags[f]; !ok || !val {
				return false
			}
		}
	}
	return true
}
