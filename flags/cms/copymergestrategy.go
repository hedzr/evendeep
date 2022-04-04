// go: generate stringer -type=CopyMergeStrategy -linecomment

package cms

// CopyMergeStrategy _
type CopyMergeStrategy int

// CopyMergeStrategies an array of CopyMergeStrategy
type CopyMergeStrategies []CopyMergeStrategy

// Parse decodes the given string and return the matched CopyMergeStrategy value.
func (i CopyMergeStrategy) Parse(s string) CopyMergeStrategy {
	for ix, str := range _CopyMergeStrategy_map {
		if s == str {
			return ix
		}
	}
	return InvalidStrategy
}

const (
	// Default the public fields will be copied
	Default CopyMergeStrategy = iota // std
	// Ignore the ignored fields will be ignored in all scenes
	Ignore // -
	// Must the must-be-copied fields will always be copied to the target
	Must // must

	// ClearIfEq the target field will be reset/clear to zero if it equals to the source.
	// Just for struct fields
	ClearIfEq CopyMergeStrategy = iota + 10 // cleareq

	// KeepIfNotEq the source field will not be copied if it does not equal to the target
	// Just for struct fields
	KeepIfNotEq // keepneq

	// ClearIfInvalid the target field will be reset/cleart to zero if source is invalid.
	// default is ON.
	ClearIfInvalid // clearinvalid

	// NoOmit never omit any source fields
	NoOmit CopyMergeStrategy = iota + 20 - 5 // noomit
	// OmitIfEmpty is both OmitIfSourceNil + OmitIfSourceZero
	OmitIfEmpty // omitempty
	// OmitIfNil the target field will be kept if source is nil
	OmitIfNil // omitnil
	// OmitIfZero the target field will be kept if source is zero
	OmitIfZero // omitzero

	// NoOmitTarget never omit any target fields
	NoOmitTarget CopyMergeStrategy = iota + 30 - 9 // noomittgt
	// OmitIfTargetEmpty is both OmitIfTargetNil + OmitIfTargetZero
	OmitIfTargetEmpty // omitemptytgt
	// OmitIfTargetNil keeps the target field if it is nil
	OmitIfTargetNil // omitniltgt
	// OmitIfTargetZero keeps the target field if it is zero
	OmitIfTargetZero // omitzerotgt

	// SliceCopy the source slice will be set or duplicated to the target.
	// the target slice will be lost.
	SliceCopy CopyMergeStrategy = iota + 50 - 13 // slicecopy
	// SliceCopyAppend the source slice will be appended into the target.
	// The original value in the target will be kept
	SliceCopyAppend // slicecopyappend
	// SliceMerge the source slice will be appended into the target
	// if anyone of them is not exists inside the target slice.
	//
	// The duplicated items in the target original slice have no changes.
	//
	// The uniqueness checking is only applied to each source slice items.
	SliceMerge // slicemerge

	// MapCopy do copy source map to the target
	MapCopy CopyMergeStrategy = iota + 70 - 16 // mapcopy
	// MapMerge try to merge each fields inside source map recursively,
	// even if it's a slice, a pointer, another sub-map, and so on.
	MapMerge // mapmerge

	//
	// // --- Globally settings ---
	//

	// The following constants are reserved for the future purpose.
	// All of them should NOT be used in your user-side codes.

	// UnexportedToo _
	UnexportedToo CopyMergeStrategy = iota + 90 - 18 // private

	// ByOrdinal will be applied to struct, map and slice.
	// As to slice, it is standard and unique choice.
	ByOrdinal // byordinal
	// ByName will be applied to struct or map.
	ByName // byname

	// MaxStrategy is a mark to indicate the max value of all available CopyMergeStrategies
	MaxStrategy

	// reserved:

	ftf100 CopyMergeStrategy = iota + 100
	ftf110 CopyMergeStrategy = iota + 110
	ftf120 CopyMergeStrategy = iota + 120
	ftf130 CopyMergeStrategy = iota + 130
	ftf140 CopyMergeStrategy = iota + 140
	ftf150 CopyMergeStrategy = iota + 150
	ftf160 CopyMergeStrategy = iota + 160
	ftf170 CopyMergeStrategy = iota + 170

	// InvalidStrategy for algorithm purpose
	InvalidStrategy CopyMergeStrategy = CopyMergeStrategy(MaxInt)
)

// Limit values of implementation-specific int type.
const (

	// https://stackoverflow.com/questions/6878590/the-maximum-value-for-an-int-type-in-go
	// https://github.com/golang/go/blob/master/src/math/const.go#L39
	// intSize = 32 << (^uint(0) >> 63) // 32 or 64
	// MaxInt  = 1<<(intSize-1) - 1
	// MinInt  = -1 << (intSize - 1)

	// MaxInt = int.max (2^63-1 or 2^31-1 for CPU Bit Size = 32bits)
	MaxInt = int(^uint(0) >> 1)
	// MinInt = int.max (-2^63)
	MinInt = -MaxInt - 1
)
