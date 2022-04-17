package flags

import (
	"fmt"
	"github.com/hedzr/evendeep/flags/cms"
	"io"
	"sort"

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

func (flags Flags) Clone() (n Flags) {
	n = make(Flags)
	for k, v := range flags {
		n[k] = v
	}
	return
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

func (flags Flags) StringEx() string {
	var sb, sbfinal strings.Builder
	var cache = make(Flags)

	for fx, ok := range flags {
		if ok {
			cache[fx] = ok
		}
	}

	for i := cms.Default; i < cms.MaxStrategy; i++ {
		if flags.IsGroupedFlagOK(i) {
			cache[i] = true
		}
	}

	var keys []int
	for fx, ok := range cache {
		if ok {
			keys = append(keys, int(fx))
		}
	}
	sort.Ints(keys)

	for _, fx := range keys {
		if sb.Len() > 0 {
			sb.WriteRune(',')
		}
		sb.WriteString(((cms.CopyMergeStrategy)(fx)).String())
	}

	sbfinal.WriteRune('[')
	sbfinal.WriteString(sb.String())
	sbfinal.WriteRune(']')

	return sbfinal.String()
}

func (flags *Flags) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = fmt.Fprintf(s, "%+v", flags.StringEx())
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, flags.String())
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", flags.String())
	}
}

func (flags Flags) WithFlags(flg ...cms.CopyMergeStrategy) Flags {
	for _, f := range flg {
		flags[f] = true
		toggleTheRadio(f, flags)
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
	return //nolint:nakedret //no
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

func toggleTheRadio(f cms.CopyMergeStrategy, flags Flags) {
	if m, ok := mKnownFieldTagFlagsConflict[f]; ok {
		for fx := range m {
			if fx != f {
				if _, ok = flags[fx]; ok {
					delete(flags, fx)
				}
			}
		}
	}
}

// Parse _
// use "copy" if tagName is empty.
func Parse(s reflect.StructTag, tagName string) (flags Flags, targetNameRule NameConvertRule) {
	lazyInitFieldTagsFlags()

	if flags == nil {
		flags = New()
	}

	if tagName == "" {
		tagName = "copy"
	}

	tags := s.Get(tagName)

	for i, wh := range strings.Split(tags, ",") {
		if i == 0 && wh != "-" {
			targetNameRule = NameConvertRule(wh)
			continue
		}

		ftf := cms.Default.Parse(wh)
		if ftf == cms.InvalidStrategy {
			if wh == "ignore" {
				flags[cms.Ignore] = true
			}
		} else {
			flags[ftf] = true
		}

		toggleTheRadio(ftf, flags)
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
	return //nolint:nakedret //no
}

// NameConvertRule _
type NameConvertRule string
type nameConvertRule struct {
	Valid     bool
	IsIgnored bool
	From      string
	To        string
}

func (s NameConvertRule) Valid() bool      { return s != "" && s.get().Valid }
func (s NameConvertRule) IsIgnored() bool  { return s.get().IsIgnored }
func (s NameConvertRule) FromName() string { return s.get().From }
func (s NameConvertRule) ToName() string   { return s.get().To }

func (s NameConvertRule) get() (r nameConvertRule) {
	a := strings.Split(string(s), "->")
	if len(a) > 0 {
		if a[0] == "-" {
			r.IsIgnored = true
		} else if len(a) == 1 {
			r.To = strings.TrimSpace(a[0])
			r.Valid = true
		} else {
			r.From = strings.TrimSpace(a[0])
			r.To = strings.TrimSpace(a[1])
			r.Valid = true
		}
	}
	// dbglog.Log("      nameConvertRule: %+v", r)
	return
}
