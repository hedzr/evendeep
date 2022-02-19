package deepcopy

import "testing"

func TestFlags_testGroupedFlag(t *testing.T) {

	onceInitFieldTagsFlags()

	flags := newFlags().withFlags(SliceCopyAppend)

	for _, f := range []CopyMergeStrategy{SliceCopy, SliceCopyAppend, SliceMerge} {
		if ret := flags.testGroupedFlag(f); ret != SliceCopyAppend {
			t.Fatal("bad")
		}
	}

	flags = newFlags()

	for _, f := range []CopyMergeStrategy{SliceCopy, SliceCopyAppend, SliceMerge} {
		if flags.testGroupedFlag(f) != SliceCopy {
			t.Fatal("bad")
		}
	}

	for _, f := range []CopyMergeStrategy{OmitIfEmpty, OmitIfSourceNil, OmitIfSourceZero} {
		if flags.testGroupedFlag(f) != OmitIfEmpty {
			t.Fatal("bad")
		}
	}

	for _, f := range []CopyMergeStrategy{MapCopy, MapMerge} {
		if flags.testGroupedFlag(f) != MapCopy {
			t.Fatal("bad")
		}
	}

	for _, f := range []CopyMergeStrategy{Default, Ignore, Must} {
		if flags.testGroupedFlag(f) != Default {
			t.Fatal("bad")
		}
	}

	//

	for i, f := range []CopyMergeStrategy{SliceCopy, SliceCopyAppend, SliceMerge} {
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

	for i, f := range []CopyMergeStrategy{OmitIfEmpty, OmitIfSourceNil, OmitIfSourceZero} {
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

	for i, f := range []CopyMergeStrategy{MapCopy, MapMerge} {
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

	for i, f := range []CopyMergeStrategy{Default, Ignore, Must} {
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
