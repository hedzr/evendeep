package deepcopy

import (
	"github.com/hedzr/log"
	"reflect"
)

// paramsPackage is params package
type paramsPackage struct {
	ownerSource *reflect.Value // ownerSource of source slice or struct, or any others
	ownerTarget *reflect.Value

	index int // struct field or slice index,

	fieldTypeSource *reflect.StructField
	fieldTypeTarget *reflect.StructField
	fieldTags       *fieldTags

	children          map[string]*paramsPackage // children of struct fields
	childrenAnonymous []*paramsPackage          // or children without name (non-struct)
	owner             *paramsPackage
}

type paramsOpt func(p *paramsPackage)

func newParams(opts ...paramsOpt) *paramsPackage {
	p := &paramsPackage{}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func withOwners(ownerSource, ownerTarget *reflect.Value, index int) paramsOpt {
	return func(p *paramsPackage) {
		p.ownerSource = ownerSource
		p.ownerTarget = ownerTarget

		p.index = index
		st := ownerSource.Type()
		if st.Kind() == reflect.Struct {
			t := ownerSource.Type().Field(index)
			p.fieldTypeSource = &t
			p.fieldTags = parseFieldTags(t.Tag)
		}

		tt := ownerTarget.Type()
		if tt.Kind() == reflect.Struct {
			t := ownerTarget.Type().Field(index)
			p.fieldTypeTarget = &t
		}

	}
}

func withFlags(flags ...CopyMergeStrategy) paramsOpt {
	return func(p *paramsPackage) {
		p.fieldTags = &fieldTags{
			flags:          make(map[CopyMergeStrategy]bool),
			converter:      nil,
			copier:         nil,
			nameConverter:  nil,
			targetNameRule: "",
		}
		for _, f := range flags {
			p.fieldTags.flags[f] = true
		}
	}
}

func (params *paramsPackage) isStruct() bool { return params.fieldTags != nil }

func (params *paramsPackage) isFlagOK(ftf CopyMergeStrategy) bool {
	if params == nil {
		return false
	}
	if params.fieldTags == nil {
		return false
	}
	return params.fieldTags.flags.isFlagOK(ftf)
}

func (params *paramsPackage) isAnyFlagsOK(ftf ...CopyMergeStrategy) bool {
	if params == nil {
		return false
	}
	if params.fieldTags == nil {
		return false
	}
	return params.fieldTags.flags.isAnyFlagsOK(ftf...)
}

func (params *paramsPackage) isAllFlagsOK(ftf ...CopyMergeStrategy) bool {
	if params == nil {
		return false
	}
	if params.fieldTags == nil {
		return false
	}
	return params.fieldTags.flags.isAllFlagsOK(ftf...)
}

func (params *paramsPackage) depth() (depth int) {
	p := params
	for p != nil {
		depth++
		p = p.owner
	}
	return
}

func (params *paramsPackage) addChildField(pp *paramsPackage) {
	if params == nil {
		return
	}

	// if struct
	if pp.fieldTypeSource != nil {
		fieldName := pp.fieldTypeSource.Name

		if params.children == nil {
			params.children = make(map[string]*paramsPackage)
		}
		if _, ok := params.children[fieldName]; ok {
			log.Panicf("field %q exists, cannot iterate another field on the same name", fieldName)
		}
		if pp == nil {
			log.Panicf("setting nil paramsPackage for field %q, r u kidding me?", fieldName)
		}

		params.children[fieldName] = pp
	} else {
		params.childrenAnonymous = append(params.childrenAnonymous, pp)
	}

	pp.owner = params
}
