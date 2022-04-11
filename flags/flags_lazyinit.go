package flags

import (
	"github.com/hedzr/evendeep/flags/cms"

	"sync"
)

// lazyInitFieldTagsFlags _
func lazyInitFieldTagsFlags() {
	onceFieldTagsEquip.Do(func() {
		// nolint:gocritic //no
		// add := func(s string) { mKnownFieldTagFlags[fieldTagFlag.Parse(s)] = struct{}{} }
		conflictsAdd := func(ss ...string) {
			// nolint:gocritic //no
			// ss := strings.Split(s, ",")
			if mKnownFieldTagFlagsConflict == nil {
				mKnownFieldTagFlagsConflict = make(map[cms.CopyMergeStrategy]map[cms.CopyMergeStrategy]struct{})
			}
			if mKnownFieldTagFlagsConflictLeaders == nil {
				mKnownFieldTagFlagsConflictLeaders = make(map[cms.CopyMergeStrategy]struct{})
			}
			for i, fr := range ss {
				ftf := cms.Default.Parse(fr)
				if i == 0 {
					mKnownFieldTagFlagsConflictLeaders[ftf] = struct{}{}
				}
				if v, ok := mKnownFieldTagFlagsConflict[ftf]; !ok || (ok && v == nil) {
					mKnownFieldTagFlagsConflict[ftf] = make(map[cms.CopyMergeStrategy]struct{})
				}
				for _, to := range ss {
					if to != fr {
						mKnownFieldTagFlagsConflict[ftf][cms.Default.Parse(to)] = struct{}{}
					}
				}
			}
		}

		conflictsAdd("byordinal", "byname")

		conflictsAdd("noomit", "omitempty", "omitnil", "omitzero")
		conflictsAdd("noomittgt", "omitemptytgt", "omitniltgt", "omitzerotgt")

		conflictsAdd("slicecopy", "slicecopyappend", "slicemerge")
		conflictsAdd("mapcopy", "mapmerge")

		// nolint:gocritic //no
		// conflictsAdd("clearinvalid")
		// conflictsAdd("cleareq")
		// conflictsAdd("keepneq")

		conflictsAdd("std", "-", "must")

		mKnownStrategyGroup = []cms.CopyMergeStrategies{
			{cms.ByOrdinal, cms.ByName},
			{cms.NoOmit, cms.OmitIfEmpty, cms.OmitIfNil, cms.OmitIfZero},
			{cms.NoOmitTarget, cms.OmitIfTargetEmpty, cms.OmitIfTargetNil, cms.OmitIfTargetZero},
			{cms.SliceCopy, cms.SliceCopyAppend, cms.SliceMerge},
			{cms.MapCopy, cms.MapMerge},
			// {cms.ClearIfInvalid},
			// {cms.ClearIfEq},
			// {cms.KeepIfNotEq},
			{cms.Default, cms.Ignore, cms.Must},
		}
	})
}

var onceFieldTagsEquip sync.Once

// var mKnownFieldTagFlags map[fieldTagFlag]struct{}
var mKnownFieldTagFlagsConflict map[cms.CopyMergeStrategy]map[cms.CopyMergeStrategy]struct{}
var mKnownFieldTagFlagsConflictLeaders map[cms.CopyMergeStrategy]struct{}
var mKnownStrategyGroup []cms.CopyMergeStrategies // the toggleable radio groups
