//go:generate stringer -type=CopyMergeStrategy -linecomment

package deepcopy

import (
	"strings"
	"sync"
)

type CopyMergeStrategy int

type Flags map[CopyMergeStrategy]bool

func newFlags(ftf ...CopyMergeStrategy) Flags {
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

// type fieldTagFlag int

const (
	Default            CopyMergeStrategy = iota      // std
	Ignore                                           // -
	Must                                             // must
	ClearIfEq          CopyMergeStrategy = iota + 10 // cleareq
	OmitIfNotEq                                      // omitneq
	OmitIfEmpty                                      // omitempty
	OmitIfSourceNil                                  // omitsourcenil
	OmitIfSourceZero                                 // omitsourcezero
	SliceCopy          CopyMergeStrategy = iota + 50 // slicecopy
	SliceCopyOverwrite                               // slicecopyoverwrite
	SliceMerge                                       // slicemerge
	MapCopy            CopyMergeStrategy = iota + 70 // mapcopy
	MapMerge                                         // mapmerge

	ftf100 CopyMergeStrategy = iota + 100
	ftf110 CopyMergeStrategy = iota + 110
	ftf120 CopyMergeStrategy = iota + 120
	ftf130 CopyMergeStrategy = iota + 130
	ftf140 CopyMergeStrategy = iota + 140
	ftf150 CopyMergeStrategy = iota + 150
	ftf160 CopyMergeStrategy = iota + 160
	ftf170 CopyMergeStrategy = iota + 170
)

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

		//ftfDefault          fieldTagFlag = iota      // std
		//ftfIgnore                                    // -
		//ftfMust                                      // must
		//ftfClearIfEq        fieldTagFlag = iota + 10 // cleareq
		//ftfOmitIfNotEq                               // omitneq
		//ftfOmitIfEmpty                               // omitempty
		//ftfOmitIfSourceNil                           // omitsourcenil
		//ftfOmitIfSourceZero                          // omitsourcezero
		//ftfSliceCopy        fieldTagFlag = iota + 50 // slicecopy
		//ftfSliceCopyOverwrite                        // slicecopyoverwrite
		//ftfSliceMerge                                // slicemerge
		//ftfMapCopy          fieldTagFlag = iota + 70 // mapcopy
		//ftfMapMerge                                  // mapmerge

		conflictsAdd("omitempty,omitsourcenil,omitsourcezero")
		conflictsAdd("slicecopy,slicecopyoverwrite,slicemerge")
		conflictsAdd("mapcopy,mapmerge")
		conflictsAdd("std,-,must")
	})
}

var onceFieldTagsEquip sync.Once

//var mKnownFieldTagFlags map[fieldTagFlag]struct{}
var mKnownFieldTagFlagsConflict map[CopyMergeStrategy]map[CopyMergeStrategy]struct{}
var mKnownFieldTagFlagsConflictLeaders map[CopyMergeStrategy]struct{}
