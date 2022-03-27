package deepcopy

import (
	"encoding/json"
	"github.com/hedzr/deepcopy/flags"
	"github.com/hedzr/deepcopy/flags/cms"
	"github.com/hedzr/log/dir"
)

// Opt _
type Opt func(c *cpController)

// WithValueConverters gives a set of ValueConverter.
// The value converters will be applied on its Match returns ok.
func WithValueConverters(cvt ...ValueConverter) Opt {
	return func(c *cpController) {
		c.valueConverters = append(c.valueConverters, cvt...)
	}
}

// WithValueCopiers gives a set of ValueCopier.
// The value copiers will be applied on its Match returns ok.
func WithValueCopiers(cvt ...ValueCopier) Opt {
	return func(c *cpController) {
		c.valueCopiers = append(c.valueCopiers, cvt...)
	}
}

// WithTryApplyConverterAtFirst specifies which is first when
// trying/applying ValueConverters and ValueCopiers.
func WithTryApplyConverterAtFirst(b bool) Opt {
	return func(c *cpController) {
		c.tryApplyConverterAtFirst = b
	}
}

// WithTryApplyConverterAtFirstOpt is shortcut of WithTryApplyConverterAtFirst(true)
var WithTryApplyConverterAtFirstOpt = WithTryApplyConverterAtFirst(true)

// WithCloneStyle sets the cpController to clone mode.
// In this mode, source object will be cloned to a new
// object and returned as new target object.
func WithCloneStyle() Opt {
	return func(c *cpController) {
		c.makeNewClone = true
	}
}

// WithCopyStyle sets the cpController to copier mode.
// In this mode, source object will be deepcopied to
// target object.
func WithCopyStyle() Opt {
	return func(c *cpController) {
		c.makeNewClone = false
	}
}

// WithStrategies appends more flags into *cpController
func WithStrategies(flagsList ...cms.CopyMergeStrategy) Opt {
	return func(c *cpController) {
		if c.flags == nil {
			c.flags = flags.New(flagsList...)
		} else {
			for _, f := range flagsList {
				c.flags[f] = true
			}
		}
	}
}

// WithCopyStrategyOpt is synonym of cms.SliceCopy + cms.MapCopy
var WithCopyStrategyOpt = WithStrategies(cms.SliceCopy, cms.MapCopy)

// WithMergeStrategyOpt is synonym of cms.SliceMerge + cms.MapMerge
var WithMergeStrategyOpt = WithStrategies(cms.SliceMerge, cms.MapMerge)

// WithORMDiffOpt is synonym of cms.ClearIfEq + cms.KeepIfNotEq + cms.ClearIfInvalid
var WithORMDiffOpt = WithStrategies(cms.ClearIfEq, cms.KeepIfNotEq, cms.ClearIfInvalid)

// WithStrategiesReset clears the exists flags in a *cpController.
// So that you can append new ones (with WithStrategies(flags...)).
//
// In generally, WithStrategiesReset is synonym of cms.SliceCopy +
// cms.MapCopy, since all strategies are cleared. A nothing Flags
// means that a set of default strategies will be applied,
// in other words, its include:
//
//    cms.Default, cms.OmitIfEmpty, cms.SliceCopy,
//    cms.MapCopy, cms.ByOrdinal.
//
//
func WithStrategiesReset(flagsList ...cms.CopyMergeStrategy) Opt {
	return func(c *cpController) {
		c.flags = flags.New()
		for _, f := range flagsList {
			if _, ok := c.flags[f]; ok {
				c.flags[f] = false
			}
		}
	}
}

// WithAutoExpandForInnerStruct does copy fields with flat struct.
func WithAutoExpandForInnerStruct(autoexpand bool) Opt {
	return func(c *cpController) {
		c.autoExpandStruct = autoexpand
	}
}

// WithAutoExpandStructOpt is synonym of WithAutoExpandForInnerStruct(true)
var WithAutoExpandStructOpt = WithAutoExpandForInnerStruct(true)

// WithAutoNewForStructField does create new instance on ptr field of a struct.
func WithAutoNewForStructField(autoNew bool) Opt {
	return func(c *cpController) {
		c.autoNewStruct = autoNew
	}
}

// WithAutoNewForStructFieldOpt is synonym of WithAutoNewForStructField(true)
var WithAutoNewForStructFieldOpt = WithAutoNewForStructField(true)

// WithCopyUnexportedField try to copy the unexported fields
// with special way.
func WithCopyUnexportedField(b bool) Opt {
	return func(c *cpController) {
		c.copyUnexportedFields = b
	}
}

// WithCopyUnexportedFieldOpt is shortcut of WithCopyUnexportedField
var WithCopyUnexportedFieldOpt = WithCopyUnexportedField(true)

// WithCopyFunctionResultToTarget invoke source function member and
// pass the result to the responsible target field.
//
// It just works when target field is acceptable.
func WithCopyFunctionResultToTarget(b bool) Opt {
	return func(c *cpController) {
		c.copyFunctionResultToTarget = b
	}
}

// WithCopyFunctionResultToTargetOpt is shortcut of WithCopyFunctionResultToTarget
var WithCopyFunctionResultToTargetOpt = WithCopyFunctionResultToTarget(true)

// WithPassSourceToTargetFunction invoke target function member and
// pass the source as its input parameters.
func WithPassSourceToTargetFunction(b bool) Opt {
	return func(c *cpController) {
		c.passSourceAsFunctionInArgs = b
	}
}

// WithPassSourceToTargetFunctionOpt is shortcut of WithPassSourceToTargetFunction
var WithPassSourceToTargetFunctionOpt = WithPassSourceToTargetFunction(true)

// WithIgnoreNames does specify the ignored field names list.
//
// Use the filename wildcard match characters (aka. '*' and '?', and '**')
// as your advantages, the algor is isWildMatch() and dir.IsWildMatch.
//
// These patterns will only be tested on struct fields.
func WithIgnoreNames(names ...string) Opt {
	return func(c *cpController) {
		c.ignoreNames = append(c.ignoreNames, names...)
	}
}

// WithIgnoreNamesReset clear the ignored name list set.
func WithIgnoreNamesReset() Opt {
	return func(c *cpController) {
		c.ignoreNames = nil
	}
}

// isWildMatch provides a filename wildcard matching algorithm
// by dir.IsWildMatch.
func isWildMatch(s, pattern string) bool {
	return dir.IsWildMatch(s, pattern)
}

// WithoutPanic disable panic() call internally
func WithoutPanic() Opt {
	return func(c *cpController) {
		c.rethrow = false
	}
}

// WithStringMarshaller provides a string marshaller which will
// be applied when a map is going to be copied to string.
//
// Default is json marshaller.
//
// If BinaryMarshaler has been implemented, the source.Marshal() will
// be applied.
func WithStringMarshaller(m TextMarshaller) Opt {
	return func(c *cpController) {
		if m == nil {
			m = json.Marshal
		}
		textMarshaller = m
	}
}

// TextMarshaller for string
type TextMarshaller func(v interface{}) ([]byte, error)
