package dbglog

import (
	log2 "github.com/hedzr/log"
	"log"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)

	// enable debug level for hedzr/log and log.
	// - make log.Debugf(...) printing the text onto console;
	// - make log.VDebugf(...) printing the text onto console
	//   if build tag 'verbose' present;
	log2.SetLevel(log2.DebugLevel)
}
