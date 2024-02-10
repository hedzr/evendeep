package dbglog

func init() { //nolint:gochecknoinits //no
	// stdlog.SetFlags(stdlog.LstdFlags | stdlog.Lshortfile | stdlog.Lmicroseconds)

	// enable debug level for hedzr/log and log.
	// - make log.Debugf(...) printing the text onto console;
	// - make log.VDebugf(...) printing the text onto console
	//   if build tag 'verbose' present;
	// log2.SetLevel(log2.DebugLevel)
}
