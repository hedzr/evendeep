package deepcopy

import (
	"reflect"
	"testing"
)

func TestFieldTags_Parse(t *testing.T) {

	type A struct {
		flags     Flags `copy:",cleareq"`
		converter *ValueConverter
		wouldbe   int `copy:",must,omitneq,omitzero,slicecopyappend,mapmerge"`
	}

	var expects = []Flags{
		{Default: true, ClearIfEq: true, SliceCopy: true, MapCopy: true, OmitIfEmpty: true, ByOrdinal: true},
		{Default: true, SliceCopy: true, MapCopy: true, OmitIfEmpty: true, ByOrdinal: true},
		{Must: true, SliceCopyAppend: true, MapMerge: true, OmitIfNotEq: true, OmitIfZero: true, ByOrdinal: true},
		{ByOrdinal: true, ByName: true},
	}

	var a A

	// c := newCopier()

	v := reflect.ValueOf(&a)
	v = rindirect(v)

	for i := 0; i < v.NumField(); i++ {
		fld := v.Type().Field(i)
		ft := parseFieldTags(fld.Tag)
		if !ft.isFlagExists(Ignore) {
			t.Logf("%q flags: %v", fld.Tag, ft)
		} else {
			t.Logf("%q flags: %v", fld.Tag, ft)
		}
		testDeepEqual(t, ft.flags, expects[i])
	}
}
