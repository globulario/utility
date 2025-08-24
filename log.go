// utility/log.go
package Utility

import (
	"fmt"
	"log"
	"os"
)

// Logging globals
var (
	logChannel = make(chan string)
	logFct     func()
)

// Log writes messages both to stdout and to a logfile named after the running binary.
// It launches a background goroutine the first time it's called.
func Log(infos ...interface{}) {
	// if the channel is nil that's mean no processing function is running,
	// so I will create it once.
	if logFct == nil {
		logFct = func() {
			for msg := range logChannel {
				// Open the log file.
				f, err := os.OpenFile(os.Args[0]+".log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
				if err == nil {
					logger := log.New(f, "", log.LstdFlags)
					logger.Println(msg)
					f.Close()
				}
			}
		}
		go logFct()
	}

	// also display in the command prompt.
	logChannel <- fmt.Sprintln(infos...)
}

