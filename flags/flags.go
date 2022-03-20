package flags

import (
	"github.com/hedzr/deepcopy/flags/cms"
	"reflect"
	"strings"
)

// Flags is an efficient manager for a group of CopyMergeStrategy items.
type Flags map[cms.CopyMergeStrategy]bool

// New creates a new instance for Flags
func New(ftf ...cms.CopyMergeStrategy) Flags {
	return newFlags(ftf...)
}

func newFlags(ftf ...cms.CopyMergeStrategy) Flags {
	lazyInitFieldTagsFlags()
	flags := make(Flags)
	flags.WithFlags(ftf...)
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

func (flags Flags) WithFlags(flg ...cms.CopyMergeStrategy) Flags {
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

func (flags Flags) IsFlagOK(ftf cms.CopyMergeStrategy) bool {
	if flags != nil {
		return flags[ftf]
	}
	return false
}

func (flags Flags) testGroupedFlag(ftf cms.CopyMergeStrategy) (result cms.CopyMergeStrategy) {
	var ok, val bool
	result = cms.InvalidStrategy

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
		leader := cms.InvalidStrategy
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

func (flags Flags) leader(ff cms.CopyMergeStrategy, vm map[cms.CopyMergeStrategy]struct{}) (leader cms.CopyMergeStrategy) {
	leader = cms.InvalidStrategy
	if _, ok1 := mKnownFieldTagFlagsConflictLeaders[ff]; ok1 {
		leader = ff
	}
	for f := range vm {
		if _, ok1 := mKnownFieldTagFlagsConflictLeaders[f]; ok1 {
			leader = f
		}
	}
	return
}

// IsGroupedFlagOK test if any of ftf is exists.
//
// If one of ftf is the leader (a.k.a. the first one) of a toggleable
// group (such as map-copy and map-merge), and, any of the group is
// not exists (either map-copy and map-merge), IsGroupedFlagOK will
// report true just like map-copy was in Flags.
func (flags Flags) IsGroupedFlagOK(ftf ...cms.CopyMergeStrategy) (ok bool) {
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
			leader := flags.leader(ff, vm)

			var found, val bool
			for f := range vm {
				if val, found = flags[f]; found && val {
					break
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

func (flags Flags) IsAnyFlagsOK(ftf ...cms.CopyMergeStrategy) bool {
	if flags != nil {
		for _, f := range ftf {
			if val, ok := flags[f]; val && ok {
				return true
			}
		}
	}
	return false
}

func (flags Flags) IsAllFlagsOK(ftf ...cms.CopyMergeStrategy) bool {
	if flags != nil {
		for _, f := range ftf {
			if val, ok := flags[f]; !ok || !val {
				return false
			}
		}
	}
	return true
}

// Parse _
func Parse(s reflect.StructTag) (flags Flags, targetNameRule string) {
	lazyInitFieldTagsFlags()

	if flags == nil {
		flags = New()
	}

	tags := s.Get("copy")

	for i, wh := range strings.Split(tags, ",") {
		if i == 0 && wh != "-" {
			targetNameRule = wh
			continue
		}

		ftf := cms.Default.Parse(wh)
		flags[ftf] = true

		if vm, ok := mKnownFieldTagFlagsConflict[ftf]; ok {
			for k1 := range vm {
				if _, ok = flags[k1]; ok {
					delete(flags, k1)
				}
			}
		}
	}

	for k := range mKnownFieldTagFlagsConflictLeaders {
		var ok bool
		if _, ok = flags[k]; ok {
			continue
		}
		for k1 := range mKnownFieldTagFlagsConflict[k] {
			if _, ok = flags[k1]; ok {
				break
			}
		}

		if !ok {
			// set default mode
			flags[k] = true
		}
	}
	return
}