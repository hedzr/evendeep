// Code generated by "stringer -type=CopyMergeStrategy -linecomment"; DO NOT EDIT.

package deepcopy

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Default-0]
	_ = x[Ignore-1]
	_ = x[Must-2]
	_ = x[ClearIfEq-13]
	_ = x[OmitIfNotEq-14]
	_ = x[OmitIfEmpty-15]
	_ = x[OmitIfNil-16]
	_ = x[OmitIfZero-17]
	_ = x[OmitIfTargetNil-18]
	_ = x[OmitIfTargetZero-19]
	_ = x[SliceCopy-60]
	_ = x[SliceCopyAppend-61]
	_ = x[SliceMerge-62]
	_ = x[MapCopy-83]
	_ = x[MapMerge-84]
	_ = x[UnexportedToo-105]
	_ = x[ByOrdinal-106]
	_ = x[ByName-107]
	_ = x[MaxStrategy-108]
	_ = x[ftf100-119]
	_ = x[ftf110-130]
	_ = x[ftf120-141]
	_ = x[ftf130-152]
	_ = x[ftf140-163]
	_ = x[ftf150-174]
	_ = x[ftf160-185]
	_ = x[ftf170-196]
	_ = x[InvalidStrategy-197]
}

const _CopyMergeStrategy_name = "std-mustcleareqomitneqomitemptyomitnilomitzeroomitniltgtomitzerotgtslicecopyslicecopyappendslicemergemapcopymapmergeprivatebyordinalbynameMaxStrategyftf100ftf110ftf120ftf130ftf140ftf150ftf160ftf170InvalidStrategy"

var _CopyMergeStrategy_map = map[CopyMergeStrategy]string{
	0:   _CopyMergeStrategy_name[0:3],
	1:   _CopyMergeStrategy_name[3:4],
	2:   _CopyMergeStrategy_name[4:8],
	13:  _CopyMergeStrategy_name[8:15],
	14:  _CopyMergeStrategy_name[15:22],
	15:  _CopyMergeStrategy_name[22:31],
	16:  _CopyMergeStrategy_name[31:38],
	17:  _CopyMergeStrategy_name[38:46],
	18:  _CopyMergeStrategy_name[46:56],
	19:  _CopyMergeStrategy_name[56:67],
	60:  _CopyMergeStrategy_name[67:76],
	61:  _CopyMergeStrategy_name[76:91],
	62:  _CopyMergeStrategy_name[91:101],
	83:  _CopyMergeStrategy_name[101:108],
	84:  _CopyMergeStrategy_name[108:116],
	105: _CopyMergeStrategy_name[116:123],
	106: _CopyMergeStrategy_name[123:132],
	107: _CopyMergeStrategy_name[132:138],
	108: _CopyMergeStrategy_name[138:149],
	119: _CopyMergeStrategy_name[149:155],
	130: _CopyMergeStrategy_name[155:161],
	141: _CopyMergeStrategy_name[161:167],
	152: _CopyMergeStrategy_name[167:173],
	163: _CopyMergeStrategy_name[173:179],
	174: _CopyMergeStrategy_name[179:185],
	185: _CopyMergeStrategy_name[185:191],
	196: _CopyMergeStrategy_name[191:197],
	197: _CopyMergeStrategy_name[197:212],
}

func (i CopyMergeStrategy) String() string {
	if str, ok := _CopyMergeStrategy_map[i]; ok {
		return str
	}
	return "CopyMergeStrategy(" + strconv.FormatInt(int64(i), 10) + ")"
}
