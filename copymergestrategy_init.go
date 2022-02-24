package deepcopy

import (
	"strings"
	"sync"
)

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

		conflictsAdd("byordinal,byname")

		conflictsAdd("omitempty,omitnil,omitzero,omitniltgt,omitzerotgt")
		conflictsAdd("slicecopy,slicecopyappend,slicemerge")
		conflictsAdd("mapcopy,mapmerge")
		conflictsAdd("std,-,must")

		mKnownStrategyGroup = []CopyMergeStrategies{
			{SliceCopy, SliceCopyAppend, SliceMerge},
			{OmitIfEmpty, OmitIfNil, OmitIfZero, OmitIfTargetNil, OmitIfTargetZero},
			{MapCopy, MapMerge},
			{Default, Ignore, Must},
		}
	})
}

var onceFieldTagsEquip sync.Once

//var mKnownFieldTagFlags map[fieldTagFlag]struct{}
var mKnownFieldTagFlagsConflict map[CopyMergeStrategy]map[CopyMergeStrategy]struct{}
var mKnownFieldTagFlagsConflictLeaders map[CopyMergeStrategy]struct{}
var mKnownStrategyGroup []CopyMergeStrategies // the toggleable radio groups
