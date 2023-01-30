#!/bin/bash

ver="1.19.5"
osarch="darwin_amd64"
srcdir="$HOME/go/go$ver/pkg/$osarch"
srcdir="$HOME/go/go$ver/src"

cat <<"EOT"
package evendeep

func packageisreserved(packagename string) (shouldIgnored bool) {
	onceinitignoredpackages.Do(func() {
		_ignoredpackageprefixes = ignoredpackageprefixes{
			"github.com/golang",
			"golang.org/",
			"google.golang.org/",
		}
		// the following name list comes with go1.18beta1 src/.
		// Perhaps it would need to be updated in the future.
		_ignoredpackages = ignoredpackages{
EOT

for p in $(ls -b "$srcdir" | sort); do
	[ -d "$srcdir/$p" ] && printf "			%-15s true,\n" "\"$p\":"
done

cat <<"EOT"
		}
	})

	shouldIgnored = packagename != "" && (_ignoredpackages.contains(packagename) ||
		_ignoredpackageprefixes.contains(packagename))
	return
}
EOT
