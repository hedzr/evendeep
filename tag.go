package deepcopy

import (
	"reflect"
	"strings"
)

func parseFieldTags(tag reflect.StructTag) *fieldTags {
	t := &fieldTags{}
	t.Parse(tag)
	return t
}

// fieldTags collect the flags and others which are parsed from a struct field tags definition.
//
//     type sample struct {
//         SomeName string `copy:"someName,omitempty"`
//         IgnoredName string `copy:"-"`
//     }
type fieldTags struct {
	flags Flags `copy:"zeroIfEq"`

	converter     *ValueConverter
	copier        *ValueCopier
	nameConverter func(source string, ctx *NameConverterContext) string `yaml:"-,omitempty"`

	// targetNameRule:
	// "-"           ignore
	// "anyName"     from source field to 'anyName' field
	// "->anyName"   from source field to 'anyName' field
	targetNameRule string // first section in struct field tag, such as: "someName,must,..."
}

type ValueConverter interface {
	Transform(ctx *ValueConverterContext, source reflect.Value) (target reflect.Value, err error)
	Match(params *paramsPackage, source, target reflect.Value) (ctx *ValueConverterContext, yes bool)
}

type ValueCopier interface {
	CopyTo(ctx *ValueConverterContext, source, target reflect.Value) (err error)
	Match(params *paramsPackage, source, target reflect.Value) (ctx *ValueConverterContext, yes bool)
}

type NameConverterContext struct {
	*paramsPackage
}

type ValueConverterContext struct {
	*paramsPackage
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

func (f *fieldTags) isFlagOK(ftf CopyMergeStrategy) bool {
	if f == nil {
		return false
	}
	return f.flags[ftf]
}

func (f *fieldTags) Parse(s reflect.StructTag) {
	onceInitFieldTagsFlags()

	if f.flags == nil {
		f.flags = newFlags()
	}

	tags := s.Get("copy")

	for i, wh := range strings.Split(tags, ",") {
		if i == 0 && wh != "-" {
			f.targetNameRule = wh
			continue
		}

		ftf := Default.Parse(wh)
		f.flags[ftf] = true

		if vm, ok := mKnownFieldTagFlagsConflict[ftf]; ok {
			for k1 := range vm {
				if _, ok = f.flags[k1]; ok {
					delete(f.flags, k1)
				}
			}
		}
	}

	for k := range mKnownFieldTagFlagsConflictLeaders {
		var ok bool
		if _, ok = f.flags[k]; ok {
			continue
		}
		for k1 := range mKnownFieldTagFlagsConflict[k] {
			if _, ok = f.flags[k1]; ok {
				break
			}
		}

		if !ok {
			// set default mode
			f.flags[k] = true
		}
	}
}
