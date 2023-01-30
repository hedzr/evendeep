package flags

import (
	"github.com/hedzr/evendeep/flags/cms"

	"sync"
)

// lazyInitFieldTagsFlags initialize something.
func lazyInitFieldTagsFlags() {
	onceFieldTagsEquip.Do(func() {
		// add := func(s string) { mKnownFieldTagFlags[fieldTagFlag.Parse(s)] = struct{}{} }

		conflictsAdd("byordinal", "byname")

		conflictsAdd("noomit", "omitempty", "omitnil", "omitzero")
		conflictsAdd("noomittgt", "omitemptytgt", "omitniltgt", "omitzerotgt")

		conflictsAdd("slicecopy", "slicecopyappend", "slicemerge")
		conflictsAdd("mapcopy", "mapmerge")

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
			// {cms.Flat},
		}
	})
}

func conflictsAdd(ss ...string) {
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

var onceFieldTagsEquip sync.Once //nolint:gochecknoglobals //i know that

// var mKnownFieldTagFlags map[fieldTagFlag]struct{}

var mKnownFieldTagFlagsConflict map[cms.CopyMergeStrategy]map[cms.CopyMergeStrategy]struct{} //nolint:lll,gochecknoglobals //i know that
var mKnownFieldTagFlagsConflictLeaders map[cms.CopyMergeStrategy]struct{}                    //nolint:lll,gochecknoglobals //i know that
var mKnownStrategyGroup []cms.CopyMergeStrategies                                            //nolint:lll,unused,gochecknoglobals //i know that
// the toggleable radio groups
