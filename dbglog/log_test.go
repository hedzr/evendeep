//go:build delve || verbose
// +build delve verbose

package dbglog

import (
	"testing"

	logz "github.com/hedzr/logg/slog"
)

func TestFLog(t *testing.T) {
	// config := log.NewLoggerConfigWith(true, "logrus", "trace")
	// logger := logrus.NewWithConfig(config)
	logz.Printf("hello")
	logz.Infof("hello info")
	logz.Warnf("hello warn")
	logz.Errorf("hello error")
	logz.Debugf("hello debug")
	logz.Tracef("hello trace")

	Log("but again")

	Log("child-enabled: %v", ChildLogEnabled())
}
