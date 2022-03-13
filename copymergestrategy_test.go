package deepcopy

import "testing"

func subtest1(t *testing.T) {
	flags := newFlags().withFlags(ByName, SliceCopyAppend, OmitIfNil, MapMerge, Ignore)

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
			if flags.isGroupedFlagOK(f) {
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
		Default.Parse("??")
	})
	t.Run("stringify nonexisted flag", func(t *testing.T) {
		println(CopyMergeStrategy(99999).String())
	})
}

func TestFlags1(t *testing.T) {

	lazyInitFieldTagsFlags()

	t.Run("normal flags", func(t *testing.T) {
		flags := newFlags(SliceMerge, MapMerge)
		flags.withFlags(SliceCopy)

		if flags.testGroupedFlag(SliceCopy) != SliceCopy {
			t.Fatalf("expect SliceCopy test ok")
		}
		if flags.testGroupedFlag(SliceCopyAppend) != SliceCopy {
			t.Fatalf("expect SliceCopy test ok 1")
		}
		if flags.testGroupedFlag(SliceMerge) != SliceCopy {
			t.Fatalf("expect SliceCopy test ok 2")
		}
		if flags.testGroupedFlag(MapCopy) != MapMerge {
			t.Fatalf("expect MapMerge test ok")
		}
		if flags.testGroupedFlag(MapMerge) != MapMerge {
			t.Fatalf("expect MapMerge test ok 1")
		}

		if !flags.isFlagOK(SliceCopy) {
			t.Fatalf("expect isFlagOK(SliceCopy) test ok")
		}
		if flags.isFlagOK(SliceMerge) {
			t.Fatalf("expect isFlagOK(SliceMerge) test failure")
		}

	})

}

func TestFlags2(t *testing.T) {

	lazyInitFieldTagsFlags()

	t.Run("normal flags", func(t *testing.T) {
		flags := newFlags(SliceMerge, MapMerge)
		flags.withFlags(SliceCopy)

		if !flags.isAllFlagsOK(SliceCopy, MapMerge) {
			t.Fatalf("expect isAllFlagsOK(SliceCopy, MapMerge) test ok")
		}
		if flags.isAllFlagsOK(SliceCopyAppend, MapMerge) {
			t.Fatalf("expect isAllFlagsOK(SliceCopyAppend, MapMerge) test failure")
		}

		if !flags.isAnyFlagsOK(SliceCopyAppend, MapMerge) {
			t.Fatalf("expect isAnyFlagsOK(SliceCopyAppend, MapMerge) test ok")
		}

		if !flags.isGroupedFlagOK(Default) {
			t.Fatalf("expect isGroupedFlagOK(Default) test ok")
		}
		if flags.isGroupedFlagOK(ByName) {
			t.Fatalf("expect isGroupedFlagOK(ByName) test failure")
		}

		if !flags.isGroupedFlagOK(OmitIfEmpty) {
			t.Fatalf("expect isGroupedFlagOK(OmitIfEmpty) test ok")
		}
		if flags.isGroupedFlagOK(OmitIfZero) {
			t.Fatalf("expect isGroupedFlagOK(OmitIfZero) test failure")
		}
	})

}
