package deepcopy_test

// a little copy from logex, so we can avoid importing it

import (
	"github.com/hedzr/log"
	"io"
	"testing"
)

// LogCapturer reroutes testing.T log output
type LogCapturer interface {
	Release()
}

type logCapturer struct {
	testing.TB
	origOut io.Writer
}

func (tl logCapturer) Write(p []byte) (n int, err error) {
	tl.Logf((string)(p))
	return len(p), nil
}

func (tl logCapturer) Release() {
	log.SetOutput(tl.origOut)
}

// newCaptureLog redirects logrus output to testing.Log
func newCaptureLog(tb testing.TB) LogCapturer {
	lc := logCapturer{TB: tb, origOut: log.GetOutput()}
	if !testing.Verbose() {
		log.SetOutput(lc)
	}
	return &lc
}
