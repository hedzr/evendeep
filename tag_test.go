package deepcopy

import (
	"reflect"
	"testing"
)

func TestFieldTags_Parse(t *testing.T) {

	type A struct {
		flags     Flags `copy:",cleareq"`
		converter *ValueConverter
		wouldbe   int `copy:",must,omitneq,omitsourcezero,slicecopyoverwrite,mapmerge"`
	}

	var expects = []Flags{
		{Default: true, ClearIfEq: true, SliceCopy: true, MapCopy: true, OmitIfEmpty: true},
		{Default: true, SliceCopy: true, MapCopy: true, OmitIfEmpty: true},
		{Must: true, SliceCopyOverwrite: true, MapMerge: true, OmitIfNotEq: true, OmitIfSourceZero: true},
	}

	var a A

	c := newCopier()

	v := reflect.ValueOf(&a)
	v = c.indirect(v)

	for i := 0; i < v.NumField(); i++ {
		fld := v.Type().Field(i)
		ft := parseFieldTags(fld.Tag)
		if !ft.isFlagOK(Ignore) {
			t.Logf("%q flags: %v", fld.Tag, ft)
		} else {
			t.Logf("%q flags: %v", fld.Tag, ft)
		}
		testDeepEqual(t, ft.flags, expects[i])
	}
}
