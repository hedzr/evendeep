package diff

// Opt the options functor for New().
type Opt func(*info)

// WithIgnoredFields specifies the struct field whom should be ignored
// in comparing.
func WithIgnoredFields(names ...string) Opt {
	return func(i *info) {
		for _, name := range names {
			if name != "" {
				i.ignoredFields[name] = true
			}
		}
	}
}

// WithSliceOrderedComparison allows which one algorithm in comparing
// two slice.
//
// 1. false (default), each element will be compared one by one.
//
// 2. true, the elements in slice will be compared without ordered
// insensitive. In this case, [9, 5] and [5, 9] are equal.
func WithSliceOrderedComparison(b bool) Opt {
	return func(i *info) {
		i.sliceNoOrder = b
	}
}

// WithComparer registers your customized Comparer into internal structure.
func WithComparer(comparer ...Comparer) Opt {
	return func(i *info) {
		for _, c := range comparer {
			if c != nil {
				i.compares = append(i.compares, c)
			}
		}
	}
}

func WithSliceNoOrder(b bool) Opt {
	return func(i *info) {
		i.sliceNoOrder = b
	}
}

// WithStripPointerAtFirst set the flag which allow finding the real
// targets of the input objects.
//
// Typically, the two struct pointers will be compared with field
// by field rather than comparing its pointer addresses.
//
// For example, when you diff.Diff(&b1, &b2, diff.WithStripPointerAtFirst(true)),
// we compare the fields content of b1 and b2, we don't compare its
// pointer addresses at this time.
//
// The another implicit thing is, this feature also strips `interface{}`/`any`
// variable out of to its underlying typed object: a `var lhs any = int64(0)`
// will be decoded as `var lhsReal int64 = 0`.
func WithStripPointerAtFirst(b bool) Opt {
	return func(i *info) {
		i.stripPtr1st = b
	}
}

// WithTreatEmptyStructPtrAsNilPtr set the flag which allow a field
// with nil pointer to a struct is treated as equal to the pointer
// to this field to pointed to an empty struct.
//
// For example,
//
//	struct A{I int}
//	struct B( A *A,}
//	b1, b2 := B{}, B{ &A{}}
//	diffs := diff.Diff(b1, b2, diff. diff.WithTreatEmptyStructPtrAsNilPtr(true))
//	println(diffs)
//
// And the result is the two struct are equal. the nil pointer `b1.A`
// and the empty struct pointer `b2.A` are treated as equivalent.
func WithTreatEmptyStructPtrAsNilPtr(b bool) Opt {
	return func(i *info) {
		i.treatEmptyStructPtrAsNil = b
	}
}

// WithCompareDifferentTypeStructs gives a way to compare two different
// type structs with their fields one by one.
//
// By default, the unmatched fields will be ignored. But you can
// disable the feature by calling WithIgnoreUnmatchedFields(false).
func WithCompareDifferentTypeStructs(b bool) Opt {
	return func(i *info) {
		i.differentTypeStructs = b
		i.ignoreUnmatchedFields = true
	}
}

// WithIgnoreUnmatchedFields takes effect except in
// WithCompareDifferentTypeStructs(true) mode. It allows those unmatched
// fields don't stop the fields comparing processing.
//
// So, the two different type structs are equivalent even if some
// fields are unmatched each others, so long as the matched fields
// are equivalent.
func WithIgnoreUnmatchedFields(b bool) Opt {
	return func(i *info) {
		i.ignoreUnmatchedFields = b
	}
}

// WithCompareDifferentSizeArrays supports a feature to treat these two array is equivalent: `[2]string{"1","2"}` and `[3]string{"1","2",<empty>}`
func WithCompareDifferentSizeArrays(b bool) Opt {
	return func(i *info) {
		i.differentSizeArrays = b
	}
}
