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

	// index     int // struct field or slice index,
	//srcOffset int // -1, or an offset of the embedded struct fields
	//dstOffset int // -1, or an offset of the embedded struct fields

	//srcFieldType *reflect.StructField //
	//dstFieldType *reflect.StructField //
	//srcAnonymous bool                 //
	//dstAnonymous bool                 //
	//mergingMode    bool                 // base state

	sourcefields   fieldstable          //
	targetIterator structIterable       //
	accessor       *fieldaccessor       //
	srcIndex       int                  //
	field          *reflect.StructField // source field type
	fieldTags      *fieldTags           // tag of source field

	children          map[string]*Params // children of struct fields
	childrenAnonymous []*Params          // or children without name (non-struct)
	owner             *Params            //
	controller        *cpController      //
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

func withOwners(c *cpController, ownerParams *Params, ownerSource, ownerTarget, osDecoded, otDecoded *reflect.Value) paramsOpt {
	return func(p *Params) {

		p.srcOwner = ownerSource
		p.dstOwner = ownerTarget
		p.srcDecoded = osDecoded
		p.dstDecoded = otDecoded

		var st, tt reflect.Type

		if p.srcDecoded == nil && p.srcOwner != nil {
			d, _ := rdecode(*p.srcOwner)
			p.srcDecoded = &d
		}

		if p.srcDecoded != nil && p.srcDecoded.IsValid() {
			st = p.srcDecoded.Type()
			p.parseSourceStruct(ownerParams, st)
		} else if p.srcOwner != nil {
			st = p.srcOwner.Type()
			st = rindirectType(st)
			p.parseSourceStruct(ownerParams, st)
		}

		if p.dstDecoded == nil && p.dstOwner != nil {
			d, _ := rdecode(*p.dstOwner)
			p.dstDecoded = &d
		}

		if p.dstDecoded != nil && p.dstDecoded.IsValid() {
			tt = p.dstDecoded.Type()
			p.parseTargetStruct(ownerParams, tt)
		} else if p.dstOwner != nil {
			tt = p.dstOwner.Type()
			tt = rindirectType(tt)
			p.parseTargetStruct(ownerParams, tt)
		}

		//

		// p.mergingMode = c.flags.isAnyFlagsOK(SliceMerge, MapMerge) || ownerParams.isAnyFlagsOK(SliceMerge, MapMerge)

		if p.dstDecoded != nil {
			t := *p.dstDecoded
			p.targetIterator = newStructIterator(t,
				withStructPtrAutoExpand(c.autoExpandStruct),
				withStructFieldPtrAutoNew(true),
			)
		}

		if p.srcDecoded != nil {
			f := *p.srcDecoded
			p.sourcefields = p.sourcefields.getallfields(f, c.autoExpandStruct)
			p.withIteratorIndex(0)
		}

		//

		p.controller = c
		ownerParams.addChildParams(p)

	}
}

func (params *Params) withIteratorIndex(srcIndex int) (sourcefield tablerec) {
	params.srcIndex = srcIndex

	//if i < params.srcType.NumField() {
	//	t := params.srcType.Field(i)
	//	params.fieldType = &t
	//	params.fieldTags = parseFieldTags(t.Tag)
	//}

	if srcIndex < len(params.sourcefields.tablerecords) {
		sourcefield = params.sourcefields.tablerecords[srcIndex]
		params.field = sourcefield.StructField()
		params.fieldTags = parseFieldTags(params.field.Tag)
	}
	return
}

func (params *Params) nextTargetField() (ok bool) {
	if params.targetIterator != nil {
		params.accessor, ok = params.targetIterator.Next()
	}
	return
}

func (params *Params) inMergeMode() bool {
	return params.controller.flags.isAnyFlagsOK(SliceMerge, MapMerge) ||
		params.owner.isAnyFlagsOK(SliceMerge, MapMerge) ||
		(params.fieldTags != nil &&
			params.fieldTags.flags.isAnyFlagsOK(SliceMerge, MapMerge))
}

//func (params *Params) parseSourceFieldTag(i int) {
//	params.index = i
//}

func (params *Params) parseSourceStruct(ownerParams *Params, st reflect.Type) {
	params.srcType = st
	if kind := st.Kind(); kind == reflect.Struct {
		//idx := index
		//if ownerParams != nil {
		//	idx += ownerParams.srcOffset
		//}
		//params.withIteratorIndex(idx)
		////if idx < st.NumField() {
		////	t := st.Field(idx)
		////	p.srcFieldType = &t
		////	p.fieldTags = parseFieldTags(t.Tag)
		////	p.srcType = t.Type
		////}
		////if ownerParams != nil {
		////	if oft := ownerParams.srcFieldType; oft != nil && oft.Anonymous && oft.Type.Kind() == reflect.Struct {
		////		p.srcAnonymous = true
		////		p.srcOffset = p.index
		////	}
		////}
	}
}

func (params *Params) parseTargetStruct(ownerParams *Params, tt reflect.Type) {
	params.dstType = tt
	if kind := tt.Kind(); kind == reflect.Struct {
		//idx := index
		//if ownerParams != nil {
		//	idx += ownerParams.dstOffset
		//}
		////if idx < tt.NumField() {
		////	t := tt.Field(idx)
		////	p.dstFieldType = &t
		////	p.dstType = t.Type
		////}
		////if ownerParams != nil {
		////	if oft := ownerParams.dstFieldType; oft != nil && oft.Anonymous && oft.Type.Kind() == reflect.Struct {
		////		p.dstAnonymous = true
		////		p.dstOffset = p.index
		////	}
		////}
		////} else if ownerParams != nil && ownerParams.dstFieldType != nil {
	}
}

// addChildParams does link this params into parent params
func (params *Params) addChildParams(pp *Params) {
	if params == nil {
		return
	}

	// if struct
	if pp.field != nil {
		fieldName := pp.field.Name

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
		if pp.field != nil {
			fieldName := pp.field.Name
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

//// ValueOfSource _
//func (params *Params) ValueOfSource() reflect.Value {
//	if params.srcFieldType != nil {
//		return params.srcDecoded.Field(params.index + params.srcOffset)
//	}
//	return *params.srcOwner
//}
//
//// ValueOfDestination _
//func (params *Params) ValueOfDestination() reflect.Value {
//	if params.dstFieldType != nil {
//		return params.dstDecoded.Field(params.index + params.dstOffset)
//	}
//	return *params.dstOwner
//}

func (params *Params) isStruct() bool { return params != nil && params.fieldTags != nil }

func (params *Params) isFlagExists(ftf CopyMergeStrategy) bool {
	if params == nil || params.fieldTags == nil {
		return false
	}
	return params.fieldTags.flags.isFlagOK(ftf)
}

// isGroupedFlagOK tests if the given flag is exists or valid.
//
// Different with isGroupedFlagOKDeeply is, isGroupedFlagOK will return
// false simply while Params.fieldTags is empty or unset.
//
// When Params.fieldTags is valid, the actual testing will be forwarded
// to Params.fieldTags.flags.isGroupedFlagOK().
func (params *Params) isGroupedFlagOK(ftf CopyMergeStrategy) bool {
	if params == nil /* || params.fieldTags == nil */ {
		return newFlags().isGroupedFlagOK(ftf)
	}
	if params.fieldTags == nil {
		return false
	}
	return params.fieldTags.flags.isGroupedFlagOK(ftf)
}

// isGroupedFlagOKDeeply tests if the given flag is exists or valid.
//
// Different with isGroupedFlagOK is, isGroupedFlagOKDeeply will check
// whether the given flag is a leader (i.e. default choice) in a group
// or not, even if Params.fieldTags is empty or unset.
//
// When Params.fieldTags is valid, the actual testing will be forwarded
// to Params.fieldTags.flags.isGroupedFlagOK().
func (params *Params) isGroupedFlagOKDeeply(ftf CopyMergeStrategy) bool {
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
