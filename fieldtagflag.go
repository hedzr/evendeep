//go:generate stringer -type=fieldTagFlag -linecomment

package deepcopy

import (
	"strings"
	"sync"
)

type fieldTagFlag int

const (
	ftfDefault          fieldTagFlag = iota      // std
	ftfIgnore                                    // -
	ftfMust                                      // must
	ftfClearIfEq        fieldTagFlag = iota + 10 // cleareq
	ftfOmitIfNotEq                               // omitneq
	ftfOmitIfEmpty                               // omitempty
	ftfOmitIfSourceNil                           // omitsourcenil
	ftfOmitIfSourceZero                          // omitsourcezero
	ftfSliceCopy        fieldTagFlag = iota + 50 // slicecopy
	ftfSliceCopyEnh                              // slicecopyenh
	ftfSliceMerge                                // slicemerge
	ftfMapCopy          fieldTagFlag = iota + 70 // mapcopy
	ftfMapMerge                                  // mapmerge

	ftf100 fieldTagFlag = iota + 100
	ftf110 fieldTagFlag = iota + 110
	ftf120 fieldTagFlag = iota + 120
	ftf130 fieldTagFlag = iota + 130
	ftf140 fieldTagFlag = iota + 140
	ftf150 fieldTagFlag = iota + 150
	ftf160 fieldTagFlag = iota + 160
	ftf170 fieldTagFlag = iota + 170
)

func (i fieldTagFlag) Parse(s string) fieldTagFlag {
	for ix, str := range _fieldTagFlag_map {
		if s == str {
			return ix
		}
	}
	return ftfDefault
}

func onceInitFieldTagsFlags() {
	onceFieldTagsEquip.Do(func() {
		//add := func(s string) { mKnownFieldTagFlags[fieldTagFlag.Parse(s)] = struct{}{} }
		conflictsAdd := func(s string) {
			ss := strings.Split(s, ",")
			if mKnownFieldTagFlagsConflict == nil {
				mKnownFieldTagFlagsConflict = make(map[fieldTagFlag]map[fieldTagFlag]struct{})
			}
			if mKnownFieldTagFlagsConflictLeaders == nil {
				mKnownFieldTagFlagsConflictLeaders = make(map[fieldTagFlag]struct{})
			}
			for i, fr := range ss {
				ftf := ftfDefault.Parse(fr)
				if i == 0 {
					mKnownFieldTagFlagsConflictLeaders[ftf] = struct{}{}
				}
				if v, ok := mKnownFieldTagFlagsConflict[ftf]; !ok || (ok && v == nil) {
					mKnownFieldTagFlagsConflict[ftf] = make(map[fieldTagFlag]struct{})
				}
				for _, to := range ss {
					if to != fr {
						mKnownFieldTagFlagsConflict[ftf][ftfDefault.Parse(to)] = struct{}{}
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
		//ftfSliceCopyEnh                              // slicecopyenh
		//ftfSliceMerge                                // slicemerge
		//ftfMapCopy          fieldTagFlag = iota + 70 // mapcopy
		//ftfMapMerge                                  // mapmerge

		conflictsAdd("omitempty,omitsourcenil,omitsourcezero")
		conflictsAdd("slicecopy,slicecopyenh,slicemerge")
		conflictsAdd("mapcopy,mapmerge")
		conflictsAdd("std,-,must")
	})
}

var onceFieldTagsEquip sync.Once

//var mKnownFieldTagFlags map[fieldTagFlag]struct{}
var mKnownFieldTagFlagsConflict map[fieldTagFlag]map[fieldTagFlag]struct{}
var mKnownFieldTagFlagsConflictLeaders map[fieldTagFlag]struct{}
