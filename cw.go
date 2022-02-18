package deepcopy

import "github.com/hedzr/log/dir"

// Opt _
type Opt func(c *cpController)

func WithValueConverters(cvt ...ValueConverter) Opt {
	return func(c *cpController) {
		c.valueConverters = append(c.valueConverters, cvt...)
	}
}

func WithValueCopiers(cvt ...ValueCopier) Opt {
	return func(c *cpController) {
		c.valueCopiers = append(c.valueCopiers, cvt...)
	}
}

func WithCloneStyle() Opt {
	return func(c *cpController) {
		c.makeNewClone = true
	}
}

func WithCopyStyle() Opt {
	return func(c *cpController) {
		c.makeNewClone = false
	}
}

func WithStrategies(flags ...CopyMergeStrategy) Opt {
	return func(c *cpController) {
		if c.flags == nil {
			c.flags = make(map[CopyMergeStrategy]bool)
		}
		for _, f := range flags {
			c.flags[f] = true
		}
	}
}

func WithStrategiesReset() Opt {
	return func(c *cpController) {
		c.flags = make(map[CopyMergeStrategy]bool)
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
