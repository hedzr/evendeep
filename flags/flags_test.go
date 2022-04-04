package flags

import (
	"github.com/hedzr/evendeep/flags/cms"
	"testing"
)

func subtest1(t *testing.T) {
	flags := newFlags().WithFlags(cms.ByName, cms.SliceCopyAppend, cms.OmitIfNil, cms.MapMerge, cms.Ignore)

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

func subtest3(t *testing.T) {
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
	})
	t.Run("stringify nonexisted flag", func(t *testing.T) {
		println(cms.CopyMergeStrategy(99999).String())
	})
}

func TestFlags1(t *testing.T) {

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

func TestFlags2(t *testing.T) {

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
