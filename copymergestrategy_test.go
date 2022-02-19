package deepcopy

import "testing"

func subtest1(t *testing.T) {
	flags := newFlags().withFlags(SliceCopyAppend)

	for _, coll := range [][]CopyMergeStrategy{
		{SliceCopy, SliceCopyAppend, SliceMerge},
		{OmitIfEmpty, OmitIfSourceNil, OmitIfSourceZero},
		{MapCopy, MapMerge},
		{Default, Ignore, Must},
	} {
		for _, f := range coll {
			if ret := flags.testGroupedFlag(f); ret != coll[0] {
				t.Fatal("bad")
			}
		}
	}
}

func subtest2(t *testing.T) {
	flags := newFlags()

	for _, coll := range [][]CopyMergeStrategy{
		{SliceCopy, SliceCopyAppend, SliceMerge},
		{OmitIfEmpty, OmitIfSourceNil, OmitIfSourceZero},
		{MapCopy, MapMerge},
		{Default, Ignore, Must},
	} {
		for _, f := range coll {
			if flags.testGroupedFlag(f) != coll[0] {
				t.Fatal("bad")
			}
		}
	}
}

func subtest3(t *testing.T) {
	flags := newFlags()

	for _, coll := range [][]CopyMergeStrategy{
		{SliceCopy, SliceCopyAppend, SliceMerge},
		{OmitIfEmpty, OmitIfSourceNil, OmitIfSourceZero},
		{MapCopy, MapMerge},
		{Default, Ignore, Must},
	} {
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

	onceInitFieldTagsFlags()

	t.Run("dirty flags - testGroupedFlag returns the dirty flag when testing any flags of its group", subtest1)
	t.Run("cleaning flags - testGroupedFlag returns the leader in a group", subtest2)
	t.Run("cleaning flags - isGroupedFlagOK returns ok if testing a leader", subtest3)

}
