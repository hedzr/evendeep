package flags

import (
	"reflect"
	"testing"

	"github.com/hedzr/evendeep/flags/cms"
	"github.com/hedzr/evendeep/ref"
	"github.com/hedzr/evendeep/typ"
)

func subtest1(t *testing.T) {
	flags := newFlags().WithFlags(
		cms.ByName,
		cms.SliceCopyAppend,
		cms.OmitIfEmpty, cms.OmitIfTargetEmpty,
		cms.MapMerge,
		cms.Ignore)

	t.Logf("flags: %v", flags)

	for _, coll := range mKnownStrategyGroup {
		for _, f := range coll {
			if ret := flags.testGroupedFlag(f); ret != coll[1] {
				t.Fatalf("bad: ret = %v, expect %v", ret, coll[0])
			}
		}
	}
}

func subtest2(t *testing.T) {
	flags := newFlags()

	for _, coll := range mKnownStrategyGroup {
		for _, f := range coll {
			if flags.testGroupedFlag(f) != coll[0] {
				t.Fatal("bad")
			}
		}
	}
}

func subtest3(t *testing.T) { //nolint:revive
	flags := newFlags()

	for _, coll := range mKnownStrategyGroup {
		for i, f := range coll {
			if flags.IsGroupedFlagOK(f) {
				if i != 0 {
					t.Fatal("bad")
				}
			} else {
				if i == 0 {
					t.Fatal("bad")
				}
			}
		}
	}
}

func TestFlags_testGroupedFlag(t *testing.T) {
	lazyInitFieldTagsFlags()

	t.Run("dirty flags - testGroupedFlag returns the dirty flag when testing any flags of its group", subtest1)
	t.Run("cleaning flags - testGroupedFlag returns the leader in a group", subtest2)
	t.Run("cleaning flags - isGroupedFlagOK returns ok if testing a leader", subtest3)

	t.Run("parse nonexisted flag", func(t *testing.T) {
		cms.Default.Parse("??")
		t.Log()
	})
	t.Run("stringify nonexisted flag", func(t *testing.T) {
		println(cms.CopyMergeStrategy(99999).String())
		t.Log()
	})
}

func TestFlags1(t *testing.T) { //nolint:revive
	lazyInitFieldTagsFlags()

	t.Run("normal flags", func(t *testing.T) {
		flags := newFlags(cms.SliceMerge, cms.MapMerge)
		flags.WithFlags(cms.SliceCopy)

		if flags.testGroupedFlag(cms.SliceCopy) != cms.SliceCopy {
			t.Fatalf("expect SliceCopy test ok")
		}
		if flags.testGroupedFlag(cms.SliceCopyAppend) != cms.SliceCopy {
			t.Fatalf("expect SliceCopy test ok 1")
		}
		if flags.testGroupedFlag(cms.SliceMerge) != cms.SliceCopy {
			t.Fatalf("expect SliceCopy test ok 2")
		}
		if flags.testGroupedFlag(cms.MapCopy) != cms.MapMerge {
			t.Fatalf("expect MapMerge test ok")
		}
		if flags.testGroupedFlag(cms.MapMerge) != cms.MapMerge {
			t.Fatalf("expect MapMerge test ok 1")
		}

		if !flags.IsFlagOK(cms.SliceCopy) {
			t.Fatalf("expect isFlagOK(SliceCopy) test ok")
		}
		if flags.IsFlagOK(cms.SliceMerge) {
			t.Fatalf("expect isFlagOK(SliceMerge) test failure")
		}
	})
}

func TestFlags2(t *testing.T) { //nolint:revive
	lazyInitFieldTagsFlags()

	t.Run("normal flags", func(t *testing.T) {
		flags := newFlags(cms.SliceMerge, cms.MapMerge)
		flags.WithFlags(cms.SliceCopy)

		if !flags.IsAllFlagsOK(cms.SliceCopy, cms.MapMerge) {
			t.Fatalf("expect isAllFlagsOK(SliceCopy, MapMerge) test ok")
		}
		if flags.IsAllFlagsOK(cms.SliceCopyAppend, cms.MapMerge) {
			t.Fatalf("expect isAllFlagsOK(SliceCopyAppend, MapMerge) test failure")
		}

		if !flags.IsAnyFlagsOK(cms.SliceCopyAppend, cms.MapMerge) {
			t.Fatalf("expect isAnyFlagsOK(SliceCopyAppend, MapMerge) test ok")
		}

		if !flags.IsGroupedFlagOK(cms.Default) {
			t.Fatalf("expect isGroupedFlagOK(Default) test ok")
		}
		if flags.IsGroupedFlagOK(cms.ByName) {
			t.Fatalf("expect isGroupedFlagOK(ByName) test failure")
		}

		if !flags.IsGroupedFlagOK(cms.NoOmit) {
			t.Fatalf("expect isGroupedFlagOK(NoOmit) test ok")
		}
		if flags.IsGroupedFlagOK(cms.OmitIfZero) {
			t.Fatalf("expect isGroupedFlagOK(OmitIfZero) test failure")
		}

		if !flags.IsGroupedFlagOK(cms.NoOmitTarget) {
			t.Fatalf("expect isGroupedFlagOK(NoOmitTarget) test ok")
		}
	})
}

func TestFlagsNew(t *testing.T) {
	f := New()
	t.Logf("f: %v", f)

	f = New(cms.SliceCopy)
	t.Logf("f: %v / %v", f, f.StringEx())

	f2 := f.Clone()
	t.Logf("f2: %v / %v", f2, f2.StringEx())
}

//

func TestFieldTags_Parse(t *testing.T) {
	t.Run("test fieldTags parse", subtestParse)
	// t.Run("test fieldTags flags tests", subtestFlagTests)
}

type AFT struct {
	flat01 *int `copy:",flat"` //nolint:revive,unused

	shallow01 *int `copy:",shallow"`

	flags   Flags `copy:",cleareq"`                                        //nolint:revive,unused,structcheck //test only
	wouldBe int   `copy:",must,keepneq,omitzero,slicecopyappend,mapmerge"` //nolint:revive,unused,structcheck //test only

	ignored01 int `copy:"-"` //nolint:revive,unused
}

func prepareAFT() (a AFT, expects []Flags) { //nolint:revive,unparam
	expects = []Flags{
		// flat01
		{cms.Flat: true, cms.Default: true, cms.SliceCopy: true, cms.MapCopy: true, cms.NoOmitTarget: true, cms.NoOmit: true, cms.ByOrdinal: true},

		// shallow01
		{cms.Shallow: true, cms.Default: true, cms.SliceCopy: true, cms.MapCopy: true, cms.NoOmitTarget: true, cms.NoOmit: true, cms.ByOrdinal: true},

		{cms.Default: true, cms.ClearIfEq: true, cms.SliceCopy: true, cms.MapCopy: true, cms.NoOmitTarget: true, cms.NoOmit: true, cms.ByOrdinal: true},

		// {cms.Default: true, cms.SliceCopy: true, cms.MapCopy: true,
		//	cms.NoOmitTarget: true, cms.NoOmit: true, cms.ByOrdinal: true},

		{cms.Must: true, cms.KeepIfNotEq: true, cms.SliceCopyAppend: true, cms.MapMerge: true, cms.NoOmitTarget: true, cms.OmitIfZero: true, cms.ByOrdinal: true}, //nolint:revive,lll

		// ignored01
		{cms.Ignore: true, cms.SliceCopy: true, cms.MapCopy: true, cms.NoOmitTarget: true, cms.NoOmit: true, cms.ByOrdinal: true},

		{cms.ByOrdinal: true, cms.ByName: true},
	}

	return
}

func subtestParse(t *testing.T) {
	a, expects := prepareAFT()

	// c := newCopier()

	v := reflect.ValueOf(&a)
	v = ref.Rindirect(v)

	for i := 0; i < v.NumField(); i++ {
		fld := v.Type().Field(i)

		f, nameConvertRules := Parse(fld.Tag, CopyTagName)

		if !isFlagExists(f, cms.Ignore) {
			t.Logf("%q flags: %v [without ignore] | %v: %v, %v -> %v", fld.Tag,
				nameConvertRules, nameConvertRules.Valid(), nameConvertRules.IsIgnored(),
				nameConvertRules.FromName(), nameConvertRules.ToName(),
			)
		} else {
			t.Logf("%q flags: %v [ignore] | %v: %v, %v -> %v", fld.Tag,
				nameConvertRules, nameConvertRules.Valid(), nameConvertRules.IsIgnored(),
				nameConvertRules.FromName(), nameConvertRules.ToName(),
			)
		}
		testDeepEqual(t.Errorf, f, expects[i])
	}
}

func isFlagExists(f Flags, ftf cms.CopyMergeStrategy) bool {
	if f == nil {
		return false
	}
	return f[ftf]
}

func testDeepEqual(printer func(msg string, args ...interface{}), got, expect typ.Any) { //nolint:revive
	// a,b:=reflect.ValueOf(got),reflect.ValueOf(expect)
	// switch kind:=a.Kind();kind {
	// case reflect.Map:
	// case reflect.Slice:
	// }

	if !reflect.DeepEqual(got, expect) {
		printer("FAIL: expecting %v but got %v", expect, got)
	}
}
