package evendeep

import (
	"reflect"
	"unsafe"

	"github.com/hedzr/evendeep/dbglog"
	"github.com/hedzr/evendeep/flags"
	"github.com/hedzr/evendeep/flags/cms"
	"github.com/hedzr/evendeep/internal/cl"
	"github.com/hedzr/evendeep/ref"
	logz "github.com/hedzr/logg/slog"
)

// Params is params package.
type Params struct {
	srcOwner   *reflect.Value // srcOwner of source slice or struct, or any others
	dstOwner   *reflect.Value // dstOwner of destination slice or struct, or any others
	srcDecoded *reflect.Value //
	dstDecoded *reflect.Value //
	srcType    reflect.Type   // = field(i+parent.srcOffset).type, or srcOwner.type for non-struct
	dstType    reflect.Type   // = field(i+parent.dstOffset).type, or dstOwner.type for non-struct

	// index     int // struct field or slice index,
	// srcOffset int // -1, or an offset of the embedded struct fields
	// dstOffset int // -1, or an offset of the embedded struct fields

	// srcFieldType *reflect.StructField //
	// dstFieldType *reflect.StructField //
	// srcAnonymous bool                 //
	// dstAnonymous bool                 //
	// mergingMode    bool                 // base state

	visited           map[visit]visiteddestination
	visiting          visit
	resultForNewSlice *reflect.Value

	targetIterator structIterable //
	accessor       accessor       //
	// srcIndex       int                  //
	// field     *reflect.StructField // source field type
	// fieldTags *fieldTags           // tag of source field

	flags flags.Flags

	children          map[string]*Params // children of struct fields
	childrenAnonymous []*Params          // or children without name (non-struct)
	owner             *Params            //
	controller        *cpController      //
}

type visit struct {
	addr1, addr2 unsafe.Pointer
	typ          reflect.Type
}

type visiteddestination struct {
	dst reflect.Value
}

type paramsOpt func(p *Params)

func newParams(opts ...paramsOpt) *Params {
	p := &Params{}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func withFlags(flagsList ...cms.CopyMergeStrategy) paramsOpt { //nolint:unused //future code
	return func(p *Params) {
		p.flags = flags.New(flagsList...)
	}
}

func withOwnersSimple(c *cpController, ownerParams *Params) paramsOpt { //nolint:revive,unparam //future code
	return func(p *Params) {
		p.controller = c
		if ownerParams != nil {
			ownerParams.addChildParams(p)
		}
	}
}

func withOwners(c *cpController, ownerParams *Params, ownerSource, ownerTarget, osDecoded, otDecoded *reflect.Value) paramsOpt { //nolint:revive,lll,gocognit //future
	return func(p *Params) {
		p.srcOwner = ownerSource
		p.dstOwner = ownerTarget
		p.srcDecoded = osDecoded
		p.dstDecoded = otDecoded

		var st, tt reflect.Type

		if p.srcDecoded == nil && p.srcOwner != nil {
			d, _ := ref.Rdecode(*p.srcOwner)
			p.srcDecoded = &d
			if p.srcOwner.IsValid() {
				p.srcType = p.srcOwner.Type()
			}
		}

		if p.srcDecoded != nil && p.srcDecoded.IsValid() {
			st = p.srcDecoded.Type()
			p.parseSourceStruct(ownerParams, st)
			p.dstType = p.dstOwner.Type()
		} else if p.srcOwner != nil {
			st = p.srcOwner.Type()
			st = ref.RindirectType(st)
			p.parseSourceStruct(ownerParams, st)
			p.dstType = st
		}

		if p.dstDecoded == nil && p.dstOwner != nil {
			d, _ := ref.Rdecode(*p.dstOwner)
			p.dstDecoded = &d
		}

		if p.dstDecoded != nil && p.dstDecoded.IsValid() {
			tt = p.dstDecoded.Type()
			p.parseTargetStruct(ownerParams, tt)
		} else if p.dstOwner != nil {
			tt = p.dstOwner.Type()
			tt = ref.RindirectType(tt)
			p.parseTargetStruct(ownerParams, tt)
		}

		//

		// p.mergingMode = c.flags.isAnyFlagsOK(SliceMerge, MapMerge) || ownerParams.isAnyFlagsOK(SliceMerge, MapMerge)

		if p.dstDecoded != nil {
			t := *p.dstDecoded
			p.targetIterator = newStructIterator(t,
				withStructPtrAutoExpand(c.autoExpandStruct),
				withStructFieldPtrAutoNew(c.autoNewStruct),
				withStructSource(p.srcDecoded, c.autoExpandStruct),
			)
		}

		//

		p.controller = c
		ownerParams.addChildParams(p)
	}
}

// func (params *Params) withIteratorIndex(srcIndex int) (sourcefield tableRecT) {
//	params.srcIndex = srcIndex
//
//	//if i < params.srcType.NumField() {
//	//	t := params.srcType.Field(i)
//	//	params.fieldType = &t
//	//	params.fieldTags = parseFieldTags(t.Tag)
//	//}
//
//	if srcIndex < len(params.sourcefields.tableRecordsT) {
//		sourcefield = params.sourcefields.tableRecordsT[srcIndex]
//		params.field = sourcefield.StructField()
//		params.fieldTags = parseFieldTags(params.field.Tag)
//	}
//	return
// }

func (params *Params) sourceFieldShouldBeIgnored() (yes bool) {
	if params.targetIterator != nil {
		yes = params.targetIterator.SourceFieldShouldBeIgnored(params.controller.ignoreNames)
	}
	return
}

func (params *Params) shouldBeIgnored(name string) (yes bool) {
	if name == "" {
		return true
	}
	if params.targetIterator != nil {
		yes = params.targetIterator.ShouldBeIgnored(name, params.controller.ignoreNames)
	}
	return
}

func (params *Params) nextTargetField() (sourceField *tableRecT, ok bool) {
	if params.targetIterator != nil {
		byName := params.isGroupedFlagOKDeeply(cms.ByName)
		params.accessor, ok = params.targetIterator.Next(params, byName)
		if ok {
			sourceField = params.accessor.SourceField()
			_, isIgnored := params.parseFieldTags(sourceField.structField.Tag)
			ok = !isIgnored
		}
	}
	return
}

func (params *Params) nextTargetFieldLite() (ok bool) {
	if params.targetIterator != nil {
		byName := params.isGroupedFlagOKDeeply(cms.ByName)
		params.accessor, ok = params.targetIterator.Next(params, byName)
	}
	return
}

func (params *Params) inMergeMode() bool {
	return params.controller.flags.IsAnyFlagsOK(cms.SliceMerge, cms.MapMerge) ||
		params.owner.isAnyFlagsOK(cms.SliceMerge, cms.MapMerge) ||
		params.flags.IsAnyFlagsOK(cms.SliceMerge, cms.MapMerge)
}

func (params *Params) dstFieldIsExportedR() (copyUnexportedFields, isExported bool) {
	k := reflect.Invalid
	if params != nil && params.dstDecoded != nil {
		k = params.dstDecoded.Kind()
	}
	if k != reflect.Struct || params.accessor == nil {
		return false, true // non-struct-field target, treat it as exported
	}

	isExported, copyUnexportedFields = ref.IsExported(params.accessor.StructField()), params.controller.copyUnexportedFields
	if isExported && params.owner != nil {
		copyUnexportedFields, isExported = params.owner.dstFieldIsExportedR()
	}
	return
}

// processUnexportedField try to set newval into target if it's an unexported field.
func (params *Params) processUnexportedField(target, newval reflect.Value) (processed bool) {
	if params == nil || params.controller == nil || params.accessor == nil {
		return
	}
	if !params.controller.copyUnexportedFields {
		return
	}
	if fld := params.accessor.StructField(); fld != nil {
		// in a struct
		if !ref.IsExported(fld) {
			dbglog.Log("    unexported field %q (typ: %v): old(%v) -> new(%v)",
				fld.Name, ref.Typfmt(fld.Type), ref.Valfmt(&target), ref.Valfmt(&newval))
			cl.SetUnexportedField(target, newval)
			processed = true
		}
	}
	return
}

// func (params *Params) parseSourceFieldTag(i int) {
//	params.index = i
// }

func (params *Params) parseSourceStruct(ownerParams *Params, st reflect.Type) {
	params.srcType = st
	_ = ownerParams
	// if kind := st.Kind(); kind == reflect.Struct {
	// 	// idx := index
	// 	// if ownerParams != nil {
	// 	//	idx += ownerParams.srcOffset
	// 	// }
	// 	// params.withIteratorIndex(idx)
	// 	// //if idx < st.NumField() {
	// 	// //	t := st.Field(idx)
	// 	// //	p.srcFieldType = &t
	// 	// //	p.fieldTags = parseFieldTags(t.Tag)
	// 	// //	p.srcType = t.Type
	// 	// //}
	// 	// //if ownerParams != nil {
	// 	// //	if oft := ownerParams.srcFieldType; oft != nil && oft.Anonymous && oft.Type.Kind() == reflect.Struct {
	// 	// //		p.srcAnonymous = true
	// 	// //		p.srcOffset = p.index
	// 	// //	}
	// 	// //}
	// }
}

func (params *Params) parseTargetStruct(ownerParams *Params, tt reflect.Type) {
	params.dstType = tt
	_ = ownerParams
	// if kind := tt.Kind(); kind == reflect.Struct {
	// 	// idx := index
	// 	// if ownerParams != nil {
	// 	//	idx += ownerParams.dstOffset
	// 	// }
	// 	// //if idx < tt.NumField() {
	// 	// //	t := tt.Field(idx)
	// 	// //	p.dstFieldType = &t
	// 	// //	p.dstType = t.Type
	// 	// //}
	// 	// //if ownerParams != nil {
	// 	// //	if oft := ownerParams.dstFieldType; oft != nil && oft.Anonymous && oft.Type.Kind() == reflect.Struct {
	// 	// //		p.dstAnonymous = true
	// 	// //		p.dstOffset = p.index
	// 	// //	}
	// 	// //}
	// 	// //} else if ownerParams != nil && ownerParams.dstFieldType != nil {
	// }
}

// addChildParams does link this params into parent params.
func (params *Params) addChildParams(ppChild *Params) {
	if params == nil || ppChild == nil {
		return
	}

	// if struct
	if ppChild.accessor != nil && ppChild.accessor.StructField() != nil {
		fieldName := ppChild.accessor.StructFieldName()

		if ppChild.children == nil {
			ppChild.children = make(map[string]*Params)
		}
		if params.children == nil {
			params.children = make(map[string]*Params)
		}
		if _, ok := ppChild.children[fieldName]; ok {
			logz.Panic("field exists, cannot iterate another field on the same name", "field", fieldName)
		}
		// if ppChild == nil {
		//	logz.Panic("setting nil Params for field, r u kidding me?", "field", fieldName)
		// }

		params.children[fieldName] = ppChild
	} else {
		params.childrenAnonymous = append(params.childrenAnonymous, ppChild)
	}

	ppChild.owner = params
}

// revoke does revoke itself from parent params if necessary.
func (params *Params) revoke() {
	if pp := params.owner; pp != nil {
		if pp.accessor != nil && pp.accessor.StructField() != nil {
			fieldName := pp.accessor.StructFieldName()
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

// // ValueOfSource _
// func (params *Params) ValueOfSource() reflect.Value {
//	if params.srcFieldType != nil {
//		return params.srcDecoded.Field(params.index + params.srcOffset)
//	}
//	return *params.srcOwner
// }
//
// // ValueOfDestination _
// func (params *Params) ValueOfDestination() reflect.Value {
//	if params.dstFieldType != nil {
//		return params.dstDecoded.Field(params.index + params.dstOffset)
//	}
//	return *params.dstOwner
// }

func (params *Params) isStruct() bool { //nolint:unused //future
	if params == nil || params.accessor == nil || params.dstOwner == nil {
		return false
	}
	return params.accessor.IsStruct()
}

func (params *Params) parseFieldTags(tag reflect.StructTag) (flagsInTag *fieldTags, isIgnored bool) {
	var tagName string
	if params.controller != nil {
		tagName = params.controller.tagKeyName
	}
	flagsInTag = parseFieldTags(tag, tagName)
	isIgnored = flagsInTag.isFlagIgnored()
	return
}

func (params *Params) isFlagExists(ftf cms.CopyMergeStrategy) (ret bool) {
	if params == nil {
		return
	}
	if params.controller != nil {
		ret = params.controller.flags.IsFlagOK(ftf)
	}
	if !ret && params.flags != nil {
		ret = params.flags.IsFlagOK(ftf)
	}
	if !ret && params.accessor != nil {
		ret = params.accessor.IsFlagOK(ftf)
	}
	return
}

// isGroupedFlagOK tests if the given flag is exists or valid.
//
// Different with isGroupedFlagOKDeeply is, isGroupedFlagOK will return
// false simply while Params.fieldTags is empty or unset.
//
// When Params.fieldTags is valid, the actual testing will be forwarded
// to Params.fieldTags.flags.isGroupedFlagOK().
func (params *Params) isGroupedFlagOK(ftf ...cms.CopyMergeStrategy) (ret bool) { //nolint:unused //future
	if params == nil {
		return flags.New().IsGroupedFlagOK(ftf...)
	}
	if params.controller != nil {
		ret = params.controller.flags.IsGroupedFlagOK(ftf...)
	}
	if !ret && params.flags != nil {
		ret = params.flags.IsGroupedFlagOK(ftf...)
	}
	if !ret && params.accessor != nil {
		ret = params.accessor.IsGroupedFlagOK(ftf...)
	}
	return
}

// isGroupedFlagOKDeeply tests if the given flag is exists or valid.
//
// Different with isGroupedFlagOK is, isGroupedFlagOKDeeply will check
// whether the given flag is a leader (i.e. default choice) in a group
// or not, even if Params.fieldTags is empty or unset. And too, we run
// the same logical on these sources:
//
//  1. params.controller.flags
//  2. params.flags
//  3. params.accessor.fieldTags.flags if present
//
// When Params.fieldTags is valid, the actual testing will be forwarded
// to Params.fieldTags.flags.isGroupedFlagOK().
func (params *Params) isGroupedFlagOKDeeply(ftf ...cms.CopyMergeStrategy) (ret bool) {
	if params == nil {
		return flags.New().IsGroupedFlagOK(ftf...)
	}
	if params.controller != nil {
		ret = params.controller.flags.IsGroupedFlagOK(ftf...)
	}
	if !ret && params.flags != nil {
		ret = params.flags.IsGroupedFlagOK(ftf...)
	}
	if !ret && params.accessor != nil {
		ret = params.accessor.IsGroupedFlagOK(ftf...)
	}
	return
}

func (params *Params) isAnyFlagsOK(ftf ...cms.CopyMergeStrategy) (ret bool) {
	if params == nil {
		return
	}
	if params.controller != nil {
		ret = params.controller.flags.IsAnyFlagsOK(ftf...)
	}
	if !ret && params.flags != nil {
		ret = params.flags.IsAnyFlagsOK(ftf...)
	}
	if !ret && params.accessor != nil {
		ret = params.accessor.IsAnyFlagsOK(ftf...)
	}
	return
}

func (params *Params) isAllFlagsOK(ftf ...cms.CopyMergeStrategy) (ret bool) { //nolint:unused //future
	if params == nil {
		return
	}
	if params.controller != nil {
		ret = params.controller.flags.IsAllFlagsOK(ftf...)
	}
	if !ret && params.flags != nil {
		ret = params.flags.IsAllFlagsOK(ftf...)
	}
	if !ret && params.accessor != nil {
		ret = params.accessor.IsAllFlagsOK(ftf...)
	}
	return
}

func (params *Params) depth() (depth int) {
	p := params
	for p != nil {
		depth++
		p = p.owner
	}
	return
}
