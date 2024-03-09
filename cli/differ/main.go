package main

import (
	"log"
	"os"
)

var (
	// body         = flag.String("body", "foobar", "Body of message")
	// continuous   = flag.Bool("continuous", false, "Keep publishing messages at a 1msg/sec rate")

	WarnLog = log.New(os.Stderr, "[WARNING] ", log.LstdFlags|log.Lmsgprefix)
	ErrLog  = log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lmsgprefix)
	Log     = log.New(os.Stdout, "[INFO] ", log.LstdFlags|log.Lmsgprefix)
)

// func init() {
// 	flag.Parse()
// }
//
// func main() {
// 	if flag.NArg() > 1 {
// 		f1, f2 := flag.Arg(0), flag.Arg(1)
// 		println(f1, '\n', f2, '\n')
// 	}
// }

func main() {}
