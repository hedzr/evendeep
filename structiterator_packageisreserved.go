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
			"archive":   true,
			"bufio":     true,
			"builtin":   true,
			"bytes":     true,
			"cmd":       true,
			"compress":  true,
			"container": true,
			"context":   true,
			"crypto":    true,
			"database":  true,
			"debug":     true,
			"embed":     true,
			"encoding":  true,
			"errors":    true,
			"expvar":    true,
			"flag":      true,
			"fmt":       true,
			"go":        true,
			"hash":      true,
			"html":      true,
			"image":     true,
			"index":     true,
			"internal":  true,
			"io":        true,
			"log":       true,
			"math":      true,
			"mime":      true,
			"net":       true,
			"os":        true,
			"path":      true,
			"plugin":    true,
			"reflect":   true,
			"regexp":    true,
			"runtime":   true,
			"sort":      true,
			"strconv":   true,
			"strings":   true,
			"sync":      true,
			"syscall":   true,
			"testdata":  true,
			"testing":   true,
			"text":      true,
			"time":      true,
			"unicode":   true,
			"unsafe":    true,
			"vendor":    true,
		}
	})

	shouldIgnored = packagename != "" && (_ignoredpackages.contains(packagename) ||
		_ignoredpackageprefixes.contains(packagename))
	return
}
