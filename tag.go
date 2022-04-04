package deepcopy

import (
	"github.com/hedzr/deepcopy/flags"
	"github.com/hedzr/deepcopy/flags/cms"

	"reflect"
	"strings"
)

func parseFieldTags(tag reflect.StructTag, tagName string) *fieldTags {
	t := &fieldTags{}
	t.Parse(tag, tagName)
	return t
}

// fieldTags collect the flags and others which are parsed from a struct field tags definition.
//
//     type sample struct {
//         SomeName string `copy:"someName,omitempty"`
//         IgnoredName string `copy:"-"`
//     }
type fieldTags struct {
	flags flags.Flags `copy:"zeroIfEq"`

	converter     *ValueConverter
	copier        *ValueCopier
	nameConverter func(source string, ctx *NameConverterContext) string `yaml:"-,omitempty"`

	// targetNameRule:
	// "-"           ignore
	// "anyName"     from source field to 'anyName' field
	// "->anyName"   from source field to 'anyName' field
	targetNameRule string // first section in struct field tag, such as: "someName,must,..."
}

func (f *fieldTags) String() string {
	var a []string
	if f != nil && f.flags != nil {
		for k := range f.flags {
			a = append(a, k.String())
		}
	}
	return strings.Join(a, ", ")
}

func (f *fieldTags) isFlagExists(ftf cms.CopyMergeStrategy) bool {
	if f == nil {
		return false
	}
	return f.flags[ftf]
}

func (f *fieldTags) Parse(s reflect.StructTag, tagName string) {
	f.flags, f.targetNameRule = flags.Parse(s, tagName)
}
