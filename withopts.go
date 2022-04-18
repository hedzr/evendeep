package evendeep

import (
	"encoding/json"
	"github.com/hedzr/evendeep/flags"
	"github.com/hedzr/evendeep/flags/cms"
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

// WithSourceValueExtractor specify a source field value extractor,
// which will be applied on each field being copied to target.
//
// Just work for non-nested struct.
//
// For instance:
//
//      c := context.WithValue(context.TODO(), "Data", map[string]typ.Any{
//      	"A": 12,
//      })
//
//      tgt := struct {
//      	A int
//      }{}
//
//      evendeep.DeepCopy(c, &tgt,
//        evendeep.WithSourceValueExtractor(func(targetName string) typ.Any {
//      	if m, ok := c.Value("Data").(map[string]typ.Any); ok {
//      		return m[targetName]
//      	}
//      	return nil
//      }))
//
//      if tgt.A != 12 {
//      	t.FailNow()
//      }
//
func WithSourceValueExtractor(e SourceValueExtractor) Opt {
	return func(c *cpController) {
		c.sourceExtractor = e
	}
}

// WithTargetValueSetter _
//
// In the TargetValueSetter you could return evendeep.ErrShouldFallback to
// call the evendeep standard processing.
//
// TargetValueSetter can work for struct and map.
//
// NOTE that the sourceNames[0] is current field name, and the whole
// sourceNames slice includes the path of the nested struct(s),
// in reversal order.
//
// For instance:
//
//    type srcS struct {
//      A int
//      B bool
//      C string
//    }
//
//    src := &srcS{
//      A: 5,
//      B: true,
//      C: helloString,
//    }
//    tgt := map[string]typ.Any{
//      "Z": "str",
//    }
//
//    err := evendeep.New().CopyTo(src, &tgt,
//      evendeep.WithTargetValueSetter(
//      func(value *reflect.Value, sourceNames ...string) (err error) {
//        if value != nil {
//          name := "Mo" + strings.Join(sourceNames, ".")
//          tgt[name] = value.Interface()
//        }
//        return // ErrShouldFallback to call the evendeep standard processing
//      }),
//    )
//
//    if err != nil || tgt["MoA"] != 5 || tgt["MoB"] != true || tgt["MoC"] != helloString || tgt["Z"] != "str" {
//      t.Errorf("err: %v, tgt: %v", err, tgt)
//      t.FailNow()
//    }
func WithTargetValueSetter(e TargetValueSetter) Opt {
	return func(c *cpController) {
		c.targetSetter = e
	}
}

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

// WithByNameStrategyOpt is synonym of cms.ByName
var WithByNameStrategyOpt = WithStrategies(cms.ByName)

// WithByOrdinalStrategyOpt is synonym of cms.ByOrdinal
var WithByOrdinalStrategyOpt = WithStrategies(cms.ByOrdinal)

// WithCopyStrategyOpt is synonym of cms.SliceCopy + cms.MapCopy
var WithCopyStrategyOpt = WithStrategies(cms.SliceCopy, cms.MapCopy)

// WithMergeStrategyOpt is synonym of cms.SliceMerge + cms.MapMerge
var WithMergeStrategyOpt = WithStrategies(cms.SliceMerge, cms.MapMerge)

// WithORMDiffOpt is synonym of cms.ClearIfEq + cms.KeepIfNotEq + cms.ClearIfInvalid
var WithORMDiffOpt = WithStrategies(cms.ClearIfEq, cms.KeepIfNotEq, cms.ClearIfInvalid)

// WithOmitEmptyOpt is synonym of cms.OmitIfEmpty
var WithOmitEmptyOpt = WithStrategies(cms.OmitIfEmpty)

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
		for _, fx := range flagsList {
			if _, ok := c.flags[fx]; ok {
				c.flags[fx] = false
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

// WithStructTagName set the name which is used for retrieve the struct tag pieces.
//
// Default is "copy", the corresponding struct with tag looks like:
//
//    type AFT struct {
//        flags     flags.Flags `copy:",cleareq"`
//        converter *ValueConverter
//        wouldbe   int `copy:",must,keepneq,omitzero,mapmerge"`
//        ignored1  int `copy:"-"`
//        ignored2  int `copy:",-"`
//    }
//
func WithStructTagName(name string) Opt {
	return func(c *cpController) {
		c.tagName = name
	}
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
		RegisterDefaultStringMarshaller(m)
	}
}

// RegisterDefaultStringMarshaller provides a string marshaller which will
// be applied when a map is going to be copied to string.
//
// Default is json marshaller (json.MarshalIndent).
//
// If encoding.TextMarshaler/json.Marshaler have been implemented, the
// source.MarshalText/MarshalJSON() will be applied.
func RegisterDefaultStringMarshaller(m TextMarshaller) {
	if m == nil {
		m = json.Marshal
	}
	textMarshaller = m
}

// TextMarshaller for string
type TextMarshaller func(v interface{}) ([]byte, error)
