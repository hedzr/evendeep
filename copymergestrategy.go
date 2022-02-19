//go:generate stringer -type=CopyMergeStrategy -linecomment

package deepcopy

import (
	"strings"
	"sync"
)

type CopyMergeStrategy int

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

func (i CopyMergeStrategy) Parse(s string) CopyMergeStrategy {
	for ix, str := range _CopyMergeStrategy_map {
		if s == str {
			return ix
		}
	}
	return Default
}

func onceInitFieldTagsFlags() {
	onceFieldTagsEquip.Do(func() {
		//add := func(s string) { mKnownFieldTagFlags[fieldTagFlag.Parse(s)] = struct{}{} }
		conflictsAdd := func(s string) {
			ss := strings.Split(s, ",")
			if mKnownFieldTagFlagsConflict == nil {
				mKnownFieldTagFlagsConflict = make(map[CopyMergeStrategy]map[CopyMergeStrategy]struct{})
			}
			if mKnownFieldTagFlagsConflictLeaders == nil {
				mKnownFieldTagFlagsConflictLeaders = make(map[CopyMergeStrategy]struct{})
			}
			for i, fr := range ss {
				ftf := Default.Parse(fr)
				if i == 0 {
					mKnownFieldTagFlagsConflictLeaders[ftf] = struct{}{}
				}
				if v, ok := mKnownFieldTagFlagsConflict[ftf]; !ok || (ok && v == nil) {
					mKnownFieldTagFlagsConflict[ftf] = make(map[CopyMergeStrategy]struct{})
				}
				for _, to := range ss {
					if to != fr {
						mKnownFieldTagFlagsConflict[ftf][Default.Parse(to)] = struct{}{}
					}
				}
			}
		}

		conflictsAdd("omitempty,omitsourcenil,omitsourcezero")
		conflictsAdd("slicecopy,slicecopyappend,slicemerge")
		conflictsAdd("mapcopy,mapmerge")
		conflictsAdd("std,-,must")
	})
}

const (
	Default          CopyMergeStrategy = iota      // std
	Ignore                                         // -
	Must                                           // must
	ClearIfEq        CopyMergeStrategy = iota + 10 // cleareq
	OmitIfNotEq                                    // omitneq
	OmitIfEmpty                                    // omitempty
	OmitIfSourceNil                                // omitsourcenil
	OmitIfSourceZero                               // omitsourcezero
	SliceCopy        CopyMergeStrategy = iota + 50 // slicecopy
	SliceCopyAppend                                // slicecopyappend
	SliceMerge                                     // slicemerge
	MapCopy          CopyMergeStrategy = iota + 70 // mapcopy
	MapMerge                                       // mapmerge

	MaxStrategy

	ftf100 CopyMergeStrategy = iota + 100
	ftf110 CopyMergeStrategy = iota + 110
	ftf120 CopyMergeStrategy = iota + 120
	ftf130 CopyMergeStrategy = iota + 130
	ftf140 CopyMergeStrategy = iota + 140
	ftf150 CopyMergeStrategy = iota + 150
	ftf160 CopyMergeStrategy = iota + 160
	ftf170 CopyMergeStrategy = iota + 170

	InvalidStrategy
)

var onceFieldTagsEquip sync.Once

//var mKnownFieldTagFlags map[fieldTagFlag]struct{}
var mKnownFieldTagFlagsConflict map[CopyMergeStrategy]map[CopyMergeStrategy]struct{}
var mKnownFieldTagFlagsConflictLeaders map[CopyMergeStrategy]struct{}
