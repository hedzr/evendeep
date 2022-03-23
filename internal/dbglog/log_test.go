//go:build delve || verbose
// +build delve verbose

package dbglog

import (
	"github.com/hedzr/log"
	"testing"
)

func TestFLog(t *testing.T) {
	// config := log.NewLoggerConfigWith(true, "logrus", "trace")
	// logger := logrus.NewWithConfig(config)
	log.Printf("hello")
	log.Infof("hello info")
	log.Warnf("hello warn")
	log.Errorf("hello error")
	log.Debugf("hello debug")
	log.Tracef("hello trace")

	Log("but again")
}
