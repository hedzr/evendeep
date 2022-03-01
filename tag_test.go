package deepcopy

import (
	"reflect"
	"testing"
)

func TestFieldTags_Parse(t *testing.T) {

	t.Run("test fieldTags parse", subtestParse)
	t.Run("test fieldTags flags tests", subtestFlagTests)

}

type AFT struct {
	flags     Flags `copy:",cleareq"`
	converter *ValueConverter
	wouldbe   int `copy:",must,omitneq,omitzero,slicecopyappend,mapmerge"`
}

func prepareAFT() (a AFT, expects []Flags) {
	expects = []Flags{
		{Default: true, ClearIfEq: true, SliceCopy: true, MapCopy: true, OmitIfEmpty: true, ByOrdinal: true},
		{Default: true, SliceCopy: true, MapCopy: true, OmitIfEmpty: true, ByOrdinal: true},
		{Must: true, SliceCopyAppend: true, MapMerge: true, OmitIfNotEq: true, OmitIfZero: true, ByOrdinal: true},
		{ByOrdinal: true, ByName: true},
	}

	return
}

func subtestParse(t *testing.T) {

	a, expects := prepareAFT()

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

func subtestFlagTests(t *testing.T) {
	type AFS1 struct {
		flags     Flags           `copy:",cleareq,must"`
		converter *ValueConverter `copy:",ignore"`
		wouldbe   int             `copy:",must,omitneq,omitzero,slicecopyappend,mapmerge"`
	}
	var a AFS1
	v := reflect.ValueOf(&a)
	v = rindirect(v)
	sf, _ := v.Type().FieldByName("wouldbe")
	sf0, _ := v.Type().FieldByName("flags")
	sf1, _ := v.Type().FieldByName("converter")

	var ft fieldTags
	ft.Parse(sf.Tag)
	ft.Parse(sf0.Tag) // entering 'continue' branch
	ft.Parse(sf1.Tag) // entering 'delete' branch

	var z *fieldTags
	z.isFlagExists(SliceCopy)

	v = reflect.ValueOf(&z)
	rwant(v, reflect.Struct)
	ve := v.Elem()
	t.Logf("z: %v, nil: %v", valfmt(&ve), valfmt(nil))

	var nilArray = [1]*int{(*int)(nil)}
	v = reflect.ValueOf(nilArray)
	t.Logf("nilArray: %v, nil: %v", valfmt(&v), valfmt(nil))

	v = reflect.ValueOf(&fieldTags{
		flags:          nil,
		converter:      nil,
		copier:         nil,
		nameConverter:  nil,
		targetNameRule: "",
	})
	rwant(v, reflect.Struct)

	var ss1 = []int{8, 9}
	var ss2 = []int64{}
	var ss3 = []int{}
	var ss4 = [4]int{}
	var vv1 = reflect.ValueOf(ss1)
	var tt3 = reflect.TypeOf(ss3)
	var tp4 = reflect.TypeOf(&ss4)
	t.Logf("ss1.type: %v", typfmtv(&vv1))
	t.Log(canConvertHelper(reflect.ValueOf(&ss1), reflect.TypeOf(&ss2)))
	t.Log(canConvertHelper(vv1, reflect.TypeOf(ss2)))
	t.Log(canConvertHelper(vv1, tt3))
	t.Log(canConvertHelper(vv1, tp4))
}
