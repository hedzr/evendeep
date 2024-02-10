package dbglog

// a little copy from logex, so we can avoid importing it

import (
	"io"
	"testing"

	logz "github.com/hedzr/logg/slog"
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
	tl.TB.Logf(string(p))
	return len(p), nil
}

func (tl logCapturer) Release() {
	logz.WithWriter(tl.origOut)
}

// NewCaptureLog redirects logrus output to testing.Log
func NewCaptureLog(tb testing.TB) LogCapturer {
	lc := logCapturer{TB: tb, origOut: logz.GetDefaultWriter()}
	if !testing.Verbose() {
		logz.WithWriter(lc)
	}
	return &lc
}

// // NewCaptureLogOld redirects logrus output to testing.Log
// func NewCaptureLogOld(tb testing.TB) LogCapturer {
// 	lc := logCapturer{TB: tb, origOut: logrus.StandardLogger().Out}
// 	if !testing.Verbose() {
// 		logz.SetOutput(lc)
// 	}
// 	return &lc
// }

// CaptureLog redirects logrus output to testing.Log
func CaptureLog(tb testing.TB) LogCapturer {
	lc := logCapturer{TB: tb, origOut: logz.GetDefaultWriter()}
	logz.WithWriter(lc)
	return &lc
}
