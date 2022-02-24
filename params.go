package deepcopy

import (
	"github.com/hedzr/log"
	"reflect"
)

// Params is params package
type Params struct {
	srcOwner   *reflect.Value // srcOwner of source slice or struct, or any others
	dstOwner   *reflect.Value // dstOwner of destination slice or struct, or any others
	srcDecoded *reflect.Value //
	dstDecoded *reflect.Value //
	srcType    reflect.Type   // = field(i+parent.srcOffset).type, or srcOwner.type for non-struct
	dstType    reflect.Type   // = field(i+parent.dstOffset).type, or dstOwner.type for non-struct

	index     int // struct field or slice index,
	srcOffset int // -1, or an offset of the embedded struct fields
	dstOffset int // -1, or an offset of the embedded struct fields

	srcFieldType *reflect.StructField //
	dstFieldType *reflect.StructField //
	fieldTags    *fieldTags           // tag of source field
	srcAnonymous bool                 //
	dstAnonymous bool                 //

	children          map[string]*Params // children of struct fields
	childrenAnonymous []*Params          // or children without name (non-struct)
	owner             *Params            //
}

type paramsOpt func(p *Params)

func newParams(opts ...paramsOpt) *Params {
	p := &Params{}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func withFlags(flags ...CopyMergeStrategy) paramsOpt {
	return func(p *Params) {
		p.fieldTags = &fieldTags{
			flags:          newFlags(flags...),
			converter:      nil,
			copier:         nil,
			nameConverter:  nil,
			targetNameRule: "",
		}
	}
}

func withOwners(ownerParams *Params, ownerSource, ownerTarget, osDecoded, otDecoded *reflect.Value, index int) paramsOpt {
	return func(p *Params) {

		p.srcOwner = ownerSource
		p.dstOwner = ownerTarget
		p.srcDecoded = osDecoded
		p.dstDecoded = otDecoded
		if p.srcDecoded == nil {
			d, _ := rdecode(*p.srcOwner)
			p.srcDecoded = &d
		}
		if p.dstDecoded == nil {
			d, _ := rdecode(*p.dstOwner)
			p.dstDecoded = &d
		}

		p.index = index

		st := p.srcDecoded.Type()
		p.srcType = st
		if kind := st.Kind(); kind == reflect.Struct {
			idx := index
			if ownerParams != nil {
				idx += ownerParams.srcOffset
			}
			t := st.Field(idx)
			p.srcFieldType = &t
			p.fieldTags = parseFieldTags(t.Tag)
			p.srcType = t.Type
			if ownerParams != nil {
				if oft := ownerParams.srcFieldType; oft != nil && oft.Anonymous && oft.Type.Kind() == reflect.Struct {
					p.srcAnonymous = true
					p.srcOffset = p.index
				}
			}
		}

		tt := p.dstDecoded.Type()
		p.dstType = tt
		if kind := tt.Kind(); kind == reflect.Struct {
			idx := index
			if ownerParams != nil {
				idx += ownerParams.dstOffset
			}
			t := tt.Field(idx)
			p.dstFieldType = &t
			p.dstType = t.Type
			if ownerParams != nil {
				if oft := ownerParams.dstFieldType; oft != nil && oft.Anonymous && oft.Type.Kind() == reflect.Struct {
					p.dstAnonymous = true
					p.dstOffset = p.index
				}
			}
		} else if ownerParams != nil && ownerParams.dstFieldType != nil {

		}

		ownerParams.addChildParams(p)

	}
}

// addChildParams does link this params into parent params
func (params *Params) addChildParams(pp *Params) {
	if params == nil {
		return
	}

	// if struct
	if pp.srcFieldType != nil {
		fieldName := pp.srcFieldType.Name

		if pp.children == nil {
			pp.children = make(map[string]*Params)
		}
		if _, ok := pp.children[fieldName]; ok {
			log.Panicf("field %q exists, cannot iterate another field on the same name", fieldName)
		}
		if pp == nil {
			log.Panicf("setting nil Params for field %q, r u kidding me?", fieldName)
		}

		pp.children[fieldName] = pp
	} else {
		pp.childrenAnonymous = append(params.childrenAnonymous, params)
	}

	pp.owner = params
}

// revoke does revoke itself from parent params if necessary
func (params *Params) revoke() {
	if pp := params.owner; pp != nil {
		if pp.srcFieldType != nil {
			fieldName := pp.srcFieldType.Name
			delete(pp.children, fieldName)
		} else {
			for i := 0; i < len(pp.childrenAnonymous); i++ {
				if child := pp.childrenAnonymous[i]; child == params {
					pp.childrenAnonymous = append(pp.childrenAnonymous[0:i], pp.childrenAnonymous[i+1:]...)
					break
				}
			}
		}
	}
}

func (params *Params) ValueOfSource() reflect.Value {
	if params.srcFieldType != nil {
		return params.srcDecoded.Field(params.index + params.srcOffset)
	}
	return *params.srcOwner
}

func (params *Params) ValueOfDestination() reflect.Value {
	if params.dstFieldType != nil {
		return params.dstDecoded.Field(params.index + params.dstOffset)
	}
	return *params.dstOwner
}

func (params *Params) isStruct() bool { return params != nil && params.fieldTags != nil }

func (params *Params) isFlagExists(ftf CopyMergeStrategy) bool {
	if params == nil || params.fieldTags == nil {
		return false
	}
	return params.fieldTags.flags.isFlagOK(ftf)
}

func (params *Params) isGroupedFlagOK(ftf CopyMergeStrategy) bool {
	if params == nil || params.fieldTags == nil {
		return newFlags().isGroupedFlagOK(ftf)
	}
	return params.fieldTags.flags.isGroupedFlagOK(ftf)
}

func (params *Params) isAnyFlagsOK(ftf ...CopyMergeStrategy) bool {
	if params == nil || params.fieldTags == nil {
		return false
	}
	return params.fieldTags.flags.isAnyFlagsOK(ftf...)
}

func (params *Params) isAllFlagsOK(ftf ...CopyMergeStrategy) bool {
	if params == nil || params.fieldTags == nil {
		return false
	}
	return params.fieldTags.flags.isAllFlagsOK(ftf...)
}

func (params *Params) depth() (depth int) {
	p := params
	for p != nil {
		depth++
		p = p.owner
	}
	return
}
