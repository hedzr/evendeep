package deepcopy

import "github.com/hedzr/log/dir"

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

// WithAutoExpandInnerStruct does copy fields with flat struct.
func WithAutoExpandInnerStruct(autoexpand bool) Opt {
	return func(c *cpController) {
		c.autoExpandStuct = autoexpand
	}
}

// WithStrategies appends more flags into *cpController
func WithStrategies(flags ...CopyMergeStrategy) Opt {
	return func(c *cpController) {
		if c.flags == nil {
			c.flags = newFlags(flags...)
		} else {
			for _, f := range flags {
				c.flags[f] = true
			}
		}
	}
}

// WithCopyStrategy is synonym of SliceCopy + MapCopy
var WithCopyStrategy = WithStrategies(SliceCopy, MapCopy)

// WithMergeStrategy is synonym of SliceMerge + MapMerge
var WithMergeStrategy = WithStrategies(SliceMerge, MapMerge)

// WithStrategiesReset clears the exists flags in a *cpController.
// So that you can append new ones (with WithStrategies(flags...)).
//
// In generally, WithStrategiesReset is synonym of SliceCopy +
// MapCopy, since all strategies are cleared. A nothing Flags
// means that a set of default strategies will be applied,
// in another words, its include:
//
//    Default, OmitIfEmpty, SliceCopy,
//    MapCopy, ByOrdinal.
//
//
func WithStrategiesReset() Opt {
	return func(c *cpController) {
		c.flags = newFlags()
	}
}

// WithIgnoreNames does specify the ignored field names list.
//
// Use the filename wildcard match characters (aka. '*' and '?', and '**')
// as your advantages, the algor is isWildMatch() and dir.IsWildMatch.
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
