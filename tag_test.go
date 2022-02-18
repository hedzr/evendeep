package deepcopy

import (
	"reflect"
	"testing"
)

func TestFieldTags_Parse(t *testing.T) {

	type A struct {
		flags     map[fieldTagFlag]bool `copy:",cleareq"`
		converter *ValueConverter
		wouldbe   int `copy:",must,omitneq,omitsourcezero,slicecopyenh,mapmerge"`
	}

	var expects = []map[fieldTagFlag]bool{
		map[fieldTagFlag]bool{ftfDefault: true, ftfClearIfEq: true, ftfSliceCopy: true, ftfMapCopy: true, ftfOmitIfEmpty: true},
		map[fieldTagFlag]bool{ftfDefault: true, ftfSliceCopy: true, ftfMapCopy: true, ftfOmitIfEmpty: true},
		map[fieldTagFlag]bool{ftfMust: true, ftfSliceCopyEnh: true, ftfMapMerge: true, ftfOmitIfNotEq: true, ftfOmitIfSourceZero: true},
	}

	var a A

	c := newDeepCopier()

	v := reflect.ValueOf(&a)
	v = c.indirect(v)

	for i := 0; i < v.NumField(); i++ {
		fld := v.Type().Field(i)
		ft := parseFieldTags(fld.Tag)
		if !ft.isFlagOK(ftfIgnore) {
			t.Logf("%q flags: %v", fld.Tag, ft)
		} else {
			t.Logf("%q flags: %v", fld.Tag, ft)
		}
		testDeepEqual(t, ft.flags, expects[i])
	}
}
