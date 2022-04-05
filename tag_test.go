package evendeep

import (
	"github.com/hedzr/evendeep/flags"
	"github.com/hedzr/evendeep/flags/cms"
	"github.com/hedzr/evendeep/internal/tool"
	"reflect"
	"testing"
)

func TestFieldTags_Parse(t *testing.T) {

	t.Run("test fieldTags parse", subtestParse)
	t.Run("test fieldTags flags tests", subtestFlagTests)

}

type AFT struct {
	flags     flags.Flags `copy:",cleareq"`
	converter *ValueConverter
	wouldbe   int `copy:",must,keepneq,omitzero,slicecopyappend,mapmerge"`
}

func prepareAFT() (a AFT, expects []flags.Flags) {
	expects = []flags.Flags{
		{cms.Default: true, cms.ClearIfEq: true, cms.SliceCopy: true, cms.MapCopy: true, cms.NoOmitTarget: true, cms.NoOmit: true, cms.ByOrdinal: true},
		{cms.Default: true, cms.SliceCopy: true, cms.MapCopy: true, cms.NoOmitTarget: true, cms.NoOmit: true, cms.ByOrdinal: true},
		{cms.Must: true, cms.KeepIfNotEq: true, cms.SliceCopyAppend: true, cms.MapMerge: true, cms.NoOmitTarget: true, cms.OmitIfZero: true, cms.ByOrdinal: true},
		{cms.ByOrdinal: true, cms.ByName: true},
	}

	return
}

func subtestParse(t *testing.T) {

	a, expects := prepareAFT()

	// c := newCopier()

	v := reflect.ValueOf(&a)
	v = tool.Rindirect(v)

	for i := 0; i < v.NumField(); i++ {
		fld := v.Type().Field(i)
		ft := parseFieldTags(fld.Tag, "")
		if !ft.isFlagExists(cms.Ignore) {
			t.Logf("%q flags: %v", fld.Tag, ft)
		} else {
			t.Logf("%q flags: %v", fld.Tag, ft)
		}
		testDeepEqual(t.Errorf, ft.flags, expects[i])
	}
}

func subtestFlagTests(t *testing.T) {
	type AFS1 struct {
		flags     flags.Flags     `copy:",cleareq,must"`
		converter *ValueConverter `copy:",ignore"`
		wouldbe   int             `copy:",must,keepneq,omitzero,slicecopyappend,mapmerge"`
	}
	var a AFS1
	v := reflect.ValueOf(&a)
	v = tool.Rindirect(v)
	sf, _ := v.Type().FieldByName("wouldbe")
	sf0, _ := v.Type().FieldByName("flags")
	sf1, _ := v.Type().FieldByName("converter")

	var ft fieldTags
	ft.Parse(sf.Tag, "")
	ft.Parse(sf0.Tag, "") // entering 'continue' branch
	ft.Parse(sf1.Tag, "") // entering 'delete' branch

	var z *fieldTags
	z.isFlagExists(cms.SliceCopy)

	v = reflect.ValueOf(&z)
	tool.Rwant(v, reflect.Struct)
	ve := v.Elem()
	t.Logf("z: %v, nil: %v", tool.Valfmt(&ve), tool.Valfmt(nil))

	var nilArray = [1]*int{(*int)(nil)}
	v = reflect.ValueOf(nilArray)
	t.Logf("nilArray: %v, nil: %v", tool.Valfmt(&v), tool.Valfmt(nil))

	v = reflect.ValueOf(&fieldTags{
		flags:          nil,
		converter:      nil,
		copier:         nil,
		nameConverter:  nil,
		targetNameRule: "",
	})
	tool.Rwant(v, reflect.Struct)

	var ss1 = []int{8, 9}
	var ss2 = []int64{}
	var ss3 = []int{}
	var ss4 = [4]int{}
	var vv1 = reflect.ValueOf(ss1)
	var tt3 = reflect.TypeOf(ss3)
	var tp4 = reflect.TypeOf(&ss4)
	t.Logf("ss1.type: %v", tool.Typfmtv(&vv1))
	t.Log(tool.CanConvertHelper(reflect.ValueOf(&ss1), reflect.TypeOf(&ss2)))
	t.Log(tool.CanConvertHelper(vv1, reflect.TypeOf(ss2)))
	t.Log(tool.CanConvertHelper(vv1, tt3))
	t.Log(tool.CanConvertHelper(vv1, tp4))
}
