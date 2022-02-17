//go:generate stringer -type=fieldTagFlag

package deepcopy

import (
	"reflect"
	"strings"
	"sync"
)

func parseFieldTags(tag reflect.StructTag) *fieldTags {
	t := &fieldTags{}
	t.Parse(tag)
	return t
}

type fieldTags struct {
	flags         map[string]struct{} `copy:"zeroIfEq"`
	converter     func(ctx *ValueConverterContext, source reflect.Value) (target reflect.Value, err error)
	copier        func(ctx *ValueConverterContext, source, target reflect.Value) (err error)
	nameConverter func(source string, ctx *NameConverterContext) string
}

type NameConverterContext struct {
	// todo
}

type ValueConverterContext struct {
	// todo
}

var onceFieldTagsEquip sync.Once
var mKnownFieldTagFlags map[string]struct{}
var mKnownFieldTagFlagsConflict map[string]map[string]struct{}

type fieldTagFlag int

const (
	ftfDefault          fieldTagFlag = iota //
	ftIgnore                                // -
	ftMust                                  // must
	ftfZeroIfEq                             // zeroIfEq
	ftfKeepIfSourceNil                      // keepIfSourceNil
	ftfKeepIfSourceZero                     // keepIfSourceZero
	ftfKeepIfNotEq                          // keepIfNotEq
	ftfSliceCopy                            // sliceCopy
	ftfSliceMerge                           // sliceMerge
	ftfMapCopy                              // mapCopy
	ftfMapMerge                             // mapMerge
)

func (f *fieldTags) Parse(s reflect.StructTag) {
	onceFieldTagsEquip.Do(func() {
		add := func(s string) { mKnownFieldTagFlags[s] = struct{}{} }
		conflictsAdd := func(s string) {
			ss := strings.Split(s, ",")
			if mKnownFieldTagFlagsConflict == nil {
				mKnownFieldTagFlagsConflict = make(map[string]map[string]struct{})
			}
			for _, fr := range ss {
				if v, ok := mKnownFieldTagFlagsConflict[fr]; !ok || (ok && v == nil) {
					mKnownFieldTagFlagsConflict[fr] = make(map[string]struct{})
				}
				for _, to := range ss {
					if to != fr {
						mKnownFieldTagFlagsConflict[fr][to] = struct{}{}
					}
				}
			}
		}
		mKnownFieldTagFlags = map[string]struct{}{}
		for _, wh := range strings.Split("-,must,zeroIfEq,keepIfSourceNil,keepIfSourceZero,keepIfNotEq,sliceCopy,sliceMerge,mapCopy,mapMerge", ",") {
			add(wh)
		}
		conflictsAdd("keepIfSourceNil,keepIfSourceZero")
		conflictsAdd("sliceCopy,sliceMerge")
		conflictsAdd("mapCopy,mapMerge")
		conflictsAdd("-,must")
	})

	tags := s.Get("copy")
	for _, wh := range strings.Split(tags, ",") {
		if _, ok := mKnownFieldTagFlags[wh]; ok {
			f.flags[wh] = struct{}{}
		}
	}
}
