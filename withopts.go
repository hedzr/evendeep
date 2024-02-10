package evendeep

import (
	"encoding/json"

	"github.com/hedzr/evendeep/flags"
	"github.com/hedzr/evendeep/flags/cms"
)

// Opt options functor.
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

// WithTryApplyConverterAtFirstOpt is shortcut of WithTryApplyConverterAtFirst(true).
var WithTryApplyConverterAtFirstOpt = WithTryApplyConverterAtFirst(true) //nolint:gochecknoglobals //i know that

// WithSourceValueExtractor specify a source field value extractor,
// which will be applied on each field being copied to target.
//
// Just work for non-nested struct.
//
// For instance:
//
//	c := context.WithValue(context.TODO(), "Data", map[string]typ.Any{
//		"A": 12,
//	})
//
//	tgt := struct {
//		A int
//	}{}
//
//	evendeep.DeepCopy(c, &tgt,
//	  evendeep.WithSourceValueExtractor(func(targetName string) typ.Any {
//		if m, ok := c.Value("Data").(map[string]typ.Any); ok {
//			return m[targetName]
//		}
//		return nil
//	}))
//
//	if tgt.A != 12 {
//		t.FailNow()
//	}
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
//	type srcS struct {
//	  A int
//	  B bool
//	  C string
//	}
//
//	src := &srcS{
//	  A: 5,
//	  B: true,
//	  C: helloString,
//	}
//	tgt := map[string]typ.Any{
//	  "Z": "str",
//	}
//
//	err := evendeep.New().CopyTo(src, &tgt,
//	  evendeep.WithTargetValueSetter(
//	  func(value *reflect.Value, sourceNames ...string) (err error) {
//	    if value != nil {
//	      name := "Mo" + strings.Join(sourceNames, ".")
//	      tgt[name] = value.Interface()
//	    }
//	    return // ErrShouldFallback to call the evendeep standard processing
//	  }),
//	)
//
//	if err != nil || tgt["MoA"] != 5 || tgt["MoB"] != true || tgt["MoC"] != helloString || tgt["Z"] != "str" {
//	  t.Errorf("err: %v, tgt: %v", err, tgt)
//	  t.FailNow()
//	}
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

// WithStrategies appends more flags into *cpController.
//
// For example:
//
//	WithStrategies(cms.OmitIfZero, cms.OmitIfNil, cms.OmitIfEmpty, cms.NoOmit)
//	WithStrategies(cms.ClearIfMissed, cms.ClearIfInvalid)
//	WithStrategies(cms.KeepIfNotEq, cms.ClearIfEq)
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

// WithCleanStrategies set the given flags into *cpController, older
// flags will be clear at first.
func WithCleanStrategies(flagsList ...cms.CopyMergeStrategy) Opt {
	return func(c *cpController) {
		c.flags = flags.New(flagsList...)
	}
}

// WithByNameStrategyOpt is synonym of cms.ByName by calling WithCleanStrategies.
//
// If you're using WithByNameStrategyOpt and WithStrategies(...) at same time, please notes it
// will clear any existent flags before setting.
var WithByNameStrategyOpt = WithCleanStrategies(cms.ByName) //nolint:gochecknoglobals //i know that

// WithByOrdinalStrategyOpt is synonym of cms.ByOrdinal by calling WithCleanStrategies.
//
// If you're using WithByOrdinalStrategyOpt and WithStrategies(...) at same time, please notes it
// will clear any existent flags before setting.
var WithByOrdinalStrategyOpt = WithCleanStrategies(cms.ByOrdinal) //nolint:gochecknoglobals //i know that

// WithCopyStrategyOpt is synonym of cms.SliceCopy + cms.MapCopy by calling WithCleanStrategies.
//
// If you're using WithCopyStrategyOpt and WithStrategies(...) at same time, please notes it
// will clear any existent flags before setting.
var WithCopyStrategyOpt = WithCleanStrategies(cms.SliceCopy, cms.MapCopy) //nolint:gochecknoglobals //i know that

// WithMergeStrategyOpt is synonym of cms.SliceMerge + cms.MapMerge by calling WithCleanStrategies.
//
// If you're using WithMergeStrategyOpt and WithStrategies(...) at same time, please notes it
// will clear any existent flags before setting.
var WithMergeStrategyOpt = WithCleanStrategies(cms.SliceMerge, cms.MapMerge) //nolint:gochecknoglobals //i know that

// WithORMDiffOpt is synonym of cms.ClearIfEq + cms.KeepIfNotEq + cms.ClearIfInvalid by calling WithCleanStrategies.
//
// If you're using WithORMDiffOpt and WithStrategies(...) at same time, please notes it
// will clear any existent flags before setting.
var WithORMDiffOpt = WithCleanStrategies(cms.ClearIfEq, cms.KeepIfNotEq, cms.ClearIfInvalid) //nolint:gochecknoglobals,lll //i know that

// WithOmitEmptyOpt is synonym of cms.OmitIfEmpty by calling Clean.
//
// If you're using WithOmitEmptyOpt and WithStrategies(...) at same time, please notes it
// will clear any existent flags before setting.
var WithOmitEmptyOpt = WithCleanStrategies(cms.OmitIfEmpty) //nolint:gochecknoglobals //i know that

// WithStrategiesReset clears the exists flags in a *cpController.
// So that you can append new ones (with WithStrategies(flags...)).
//
// In generally, WithStrategiesReset is synonym of cms.SliceCopy +
// cms.MapCopy, since all strategies are cleared. A nothing Flags
// means that a set of default strategies will be applied,
// in other words, its include:
//
//	cms.Default, cms.NoOmit, cms.NoOmitTarget,
//	cms.SliceCopy, cms.MapCopy,
//	cms.ByOrdinal,
//
// If a flagsList supplied, WithStrategiesReset will add them and
// set the state to false.
func WithStrategiesReset(flagsList ...cms.CopyMergeStrategy) Opt {
	return func(c *cpController) {
		c.flags = flags.New(flagsList...)
		for _, fx := range flagsList {
			if _, ok := c.flags[fx]; ok {
				c.flags[fx] = false
			}
		}
	}
}

// WithAutoExpandForInnerStruct does copy fields with flat struct.
// When autoExpandForInnerStruct is enabled, the iterator will go into
// any embedded struct and traverse its fields with a flatten mode.
//
// For instance, the iteration on struct:
//
//	type A struct {
//	   F1 string
//	   F2 int
//	}
//	type B struct {
//	   F1 bool
//	   F2 A
//	   F3 float32
//	}
//
// will produce the sequences:
//
//	B.F1, B.F2, B.F2 - A.F1, B.F2 - A.F2, B.F3
//
// Default is true.
func WithAutoExpandForInnerStruct(autoExpand bool) Opt {
	return func(c *cpController) {
		c.autoExpandStruct = autoExpand
	}
}

// WithAutoExpandStructOpt is synonym of WithAutoExpandForInnerStruct(true).
var WithAutoExpandStructOpt = WithAutoExpandForInnerStruct(true) //nolint:gochecknoglobals //i know that

// WithAutoNewForStructField does create new instance on ptr field of a struct.
//
// When cloning to a new target object, it might be helpful.
//
// Default is true.
func WithAutoNewForStructField(autoNew bool) Opt {
	return func(c *cpController) {
		c.autoNewStruct = autoNew
	}
}

// WithAutoNewForStructFieldOpt is synonym of WithAutoNewForStructField(true).
var WithAutoNewForStructFieldOpt = WithAutoNewForStructField(true) //nolint:gochecknoglobals //i know that

// WithCopyUnexportedField try to copy the unexported fields
// with special way.
//
// This feature needs unsafe package present.
//
// Default is true.
func WithCopyUnexportedField(b bool) Opt {
	return func(c *cpController) {
		c.copyUnexportedFields = b
	}
}

// WithCopyUnexportedFieldOpt is shortcut of WithCopyUnexportedField.
var WithCopyUnexportedFieldOpt = WithCopyUnexportedField(true) //nolint:gochecknoglobals //i know that

// WithCopyFunctionResultToTarget invoke source function member and
// pass the result to the responsible target field.
//
// It just works when target field is acceptable.
//
// Default is true.
func WithCopyFunctionResultToTarget(b bool) Opt {
	return func(c *cpController) {
		c.copyFunctionResultToTarget = b
	}
}

// WithCopyFunctionResultToTargetOpt is shortcut of WithCopyFunctionResultToTarget.
var WithCopyFunctionResultToTargetOpt = WithCopyFunctionResultToTarget(true) //nolint:gochecknoglobals //i know that

// WithPassSourceToTargetFunction invoke target function member and
// pass the source as its input parameters.
//
// Default is true.
func WithPassSourceToTargetFunction(b bool) Opt {
	return func(c *cpController) {
		c.passSourceAsFunctionInArgs = b
	}
}

// WithPassSourceToTargetFunctionOpt is shortcut of WithPassSourceToTargetFunction.
var WithPassSourceToTargetFunctionOpt = WithPassSourceToTargetFunction(true) //nolint:gochecknoglobals //i know that

// WithSyncAdvancing decides how to advance to next field especially
// a source field had been ignored.
// By default, (false), the target field won't be advanced while the
// source field had been ignored.
// For sync-advanced flag is true, the target field step to next.
//
// Just for cms.ByOrdinal mode.
func WithSyncAdvancing(syncAdvancing bool) Opt {
	return func(c *cpController) {
		c.advanceTargetFieldPointerEvenIfSourceIgnored = syncAdvancing
	}
}

// WithSyncAdvancingOpt is synonym of WithAutoExpandForInnerStruct(true).
var WithSyncAdvancingOpt = WithSyncAdvancing(true) //nolint:gochecknoglobals //i know that

// WithWipeTargetSliceFirst enables the option which assumes the target
// Slice or Map will be wipe out at first before copying/merging from
// source field.
func WithWipeTargetSliceFirst(wipe bool) Opt {
	return func(c *cpController) {
		c.wipeSlice1st = wipe
	}
}

// WithWipeTargetSliceFirstOpt is synonym of WithWipeTargetSliceFirst(true).
var WithWipeTargetSliceFirstOpt = WithWipeTargetSliceFirst(true) //nolint:gochecknoglobals //i know that

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
	runeInputArray := []rune(s)
	runePatternArray := []rune(pattern)
	if len(runeInputArray) > 0 && len(runePatternArray) > 0 {
		if runePatternArray[len(runePatternArray)-1] != '*' && runePatternArray[len(runePatternArray)-1] != '?' && runeInputArray[len(runeInputArray)-1] != runePatternArray[len(runePatternArray)-1] {
			return false
		}
	}
	return isMatchUtil([]rune(s), []rune(pattern), 0, 0, len([]rune(s)), len([]rune(pattern)))
}

func isMatchUtil(input, pattern []rune, inputIndex, patternIndex int, inputLength, patternLength int) bool {

	if inputIndex == inputLength && patternIndex == patternLength {
		return true
	} else if patternIndex == patternLength {
		return false
	} else if inputIndex == inputLength {
		if pattern[patternIndex] == '*' && restPatternStar(pattern, patternIndex+1, patternLength) {
			return true
		} else {
			return false
		}
	}

	if pattern[patternIndex] == '*' {
		return isMatchUtil(input, pattern, inputIndex, patternIndex+1, inputLength, patternLength) ||
			isMatchUtil(input, pattern, inputIndex+1, patternIndex, inputLength, patternLength)
	}

	if pattern[patternIndex] == '?' {
		return isMatchUtil(input, pattern, inputIndex+1, patternIndex+1, inputLength, patternLength)
	}

	if inputIndex < inputLength {
		if input[inputIndex] == pattern[patternIndex] {
			return isMatchUtil(input, pattern, inputIndex+1, patternIndex+1, inputLength, patternLength)
		} else {
			return false
		}
	}

	return false
}

func restPatternStar(pattern []rune, patternIndex int, patternLength int) bool {
	for patternIndex < patternLength {
		if pattern[patternIndex] != '*' {
			return false
		}
		patternIndex++
	}
	return true
}

// WithStructTagName set the name which is used for retrieve the struct tag pieces.
//
// Default is "copy", the corresponding struct with tag looks like:
//
//	type AFT struct {
//	    flags     flags.Flags `copy:",cleareq"`
//	    converter *ValueConverter
//	    wouldbe   int `copy:",must,keepneq,omitzero,mapmerge"`
//	    ignored1  int `copy:"-"`
//	    ignored2  int `copy:",-"`
//	}
func WithStructTagName(name string) Opt {
	return func(c *cpController) {
		c.tagKeyName = name
	}
}

// WithoutPanic disable panic() call internally.
//
// Default is true.
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
//
// It's synonym of RegisterDefaultStringMarshaller.
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
//
// It's synonym of WithStringMarshaller.
func RegisterDefaultStringMarshaller(m TextMarshaller) {
	if m == nil {
		m = json.Marshal
	}
	textMarshaller = m
}

// TextMarshaller for string marshalling.
type TextMarshaller func(v interface{}) ([]byte, error)
