package evendeep

import (
	"github.com/hedzr/evendeep/flags"
	"github.com/hedzr/evendeep/flags/cms"
	"github.com/hedzr/evendeep/ref"

	"reflect"
	"testing"
)

func TestFieldTags_Parse(t *testing.T) {
	t.Run("test fieldTags parse", subtestParse)
	t.Run("test fieldTags flags tests", subtestFlagTests)
}

type AFT struct {
	flat01 *int `copy:",flat"`

	flags     flags.Flags     `copy:",cleareq"` //nolint:unused,structcheck //test only
	converter *ValueConverter //nolint:unused,structcheck //test only
	wouldBe   int             `copy:",must,keepneq,omitzero,slicecopyappend,mapmerge"` //nolint:unused,structcheck //test only

	ignored01 int `copy:"-"`
}

func prepareAFT() (a AFT, expects []flags.Flags) {
	expects = []flags.Flags{
		// flat01
		{cms.Flat: true, cms.Default: true, cms.SliceCopy: true, cms.MapCopy: true, cms.NoOmitTarget: true, cms.NoOmit: true, cms.ByOrdinal: true},

		{cms.Default: true, cms.ClearIfEq: true, cms.SliceCopy: true, cms.MapCopy: true, cms.NoOmitTarget: true, cms.NoOmit: true, cms.ByOrdinal: true},
		{cms.Default: true, cms.SliceCopy: true, cms.MapCopy: true, cms.NoOmitTarget: true, cms.NoOmit: true, cms.ByOrdinal: true},
		{cms.Must: true, cms.KeepIfNotEq: true, cms.SliceCopyAppend: true, cms.MapMerge: true, cms.NoOmitTarget: true, cms.OmitIfZero: true, cms.ByOrdinal: true},

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
		ft := parseFieldTags(fld.Tag, "")
		if !ft.isFlagIgnored() {
			t.Logf("%q flags: %v [without ignore]", fld.Tag, ft)
		} else {
			t.Logf("%q flags: %v [ignore]", fld.Tag, ft)
		}
		testDeepEqual(t.Errorf, ft.flags, expects[i])
	}
}

func subtestFlagTests(t *testing.T) {
	type AFS1 struct {
		flags     flags.Flags     `copy:",cleareq,must"`                                   //nolint:unused,structcheck //test
		converter *ValueConverter `copy:",ignore"`                                         //nolint:unused,structcheck //test
		wouldBe   int             `copy:",must,keepneq,omitzero,slicecopyappend,mapmerge"` //nolint:unused,structcheck //test
	}
	var a AFS1

	v := reflect.ValueOf(&a)
	v = ref.Rindirect(v)
	sf, _ := v.Type().FieldByName("wouldBe")
	sf0, _ := v.Type().FieldByName("flags")
	sf1, _ := v.Type().FieldByName("converter")

	var ft fieldTags
	ft.Parse(sf.Tag, "")
	ft.Parse(sf0.Tag, "copy") // entering 'continue' branch
	ft.Parse(sf1.Tag, "")     // entering 'delete' branch

	var z *fieldTags
	z.isFlagExists(cms.SliceCopy)

	v = reflect.ValueOf(&z)
	ref.Rwant(v, reflect.Struct)
	ve := v.Elem()
	t.Logf("z: %v, nil: %v", ref.Valfmt(&ve), ref.Valfmt(nil))

	var nilArray = [1]*int{(*int)(nil)}
	v = reflect.ValueOf(nilArray)
	t.Logf("nilArray: %v, nil: %v", ref.Valfmt(&v), ref.Valfmt(nil))

	v = reflect.ValueOf(&fieldTags{
		flags:           nil,
		converter:       nil,
		copier:          nil,
		nameConverter:   nil,
		nameConvertRule: "",
	})
	ref.Rwant(v, reflect.Struct)

	var ss1 = []int{8, 9}
	var ss2 = []int64{}
	var ss3 = []int{}
	var ss4 = [4]int{}
	var vv1 = reflect.ValueOf(ss1)
	var tt3 = reflect.TypeOf(ss3)
	var tp4 = reflect.TypeOf(&ss4)
	t.Logf("ss1.type: %v", ref.Typfmtv(&vv1))
	t.Log(ref.CanConvertHelper(reflect.ValueOf(&ss1), reflect.TypeOf(&ss2)))
	t.Log(ref.CanConvertHelper(vv1, reflect.TypeOf(ss2)))
	t.Log(ref.CanConvertHelper(vv1, tt3))
	t.Log(ref.CanConvertHelper(vv1, tp4))
}

func TestFieldTags_CalcTargetName(t *testing.T) {
	ft := fieldTags{
		flags: flags.Flags{
			cms.Default: true,
		},
		converter: nil,
		copier:    nil,
		nameConverter: func(source string, ctx *NameConverterContext) (target string, ok bool) {
			if source == "" {
				target, ok = "hehe", true
			}
			return
		},
		nameConvertRule: "",
	}

	if tn, ok := ft.CalcTargetName("", nil); !ok || tn != "hehe" {
		t.FailNow()
	}
}
