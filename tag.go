package evendeep

import (
	"github.com/hedzr/evendeep/flags"
	"github.com/hedzr/evendeep/flags/cms"
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

	converter     *ValueConverter   `yaml:"-,omitempty"`
	copier        *ValueCopier      `yaml:"-,omitempty"`
	nameConverter nameConverterFunc `yaml:"-,omitempty"`

	// nameConvertRule:
	// "-"                 ignore
	// "dstName"           from source field to 'dstName' field (thinking about name converters too)
	// "->dstName"         from source field to 'dstName' field (thinking about name converters too)
	// "srcName->dstName"  from 'srcName' to 'dstName' field
	nameConvertRule flags.NameConvertRule // first section in struct field tag, such as: "someName,must,..."
}

type nameConverterFunc func(source string, ctx *NameConverterContext) (target string, ok bool)

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
	f.flags, f.nameConvertRule = flags.Parse(s, tagName)
}

func (f *fieldTags) CalcSourceName(dstName string) (srcName string, ok bool) {
	ok = f.nameConvertRule.Valid()
	srcName = strget(f.nameConvertRule.FromName(), dstName)
	// dbglog.Log("           FromName: %v (Default to %v) | RETURN: %v", f.nameConvertRule.FromName(), dstName, srcName)
	return
}

func (f *fieldTags) CalcTargetName(srcName string, ctx *NameConverterContext) (dstName string, ok bool) {
	if f.nameConverter != nil {
		dstName, ok = f.nameConverter(srcName, ctx)
		if ok {
			return
		}
	}
	ok = f.nameConvertRule.Valid()
	dstName = strget(f.nameConvertRule.ToName(), srcName)
	return
}

func strget(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
