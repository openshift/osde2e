package logging

import (
	"log"
	"os"
)

// CreateNewStdLoggerOrUseExistingLogger will create a new logger that is instantiated just like the default standard
// logger in the log package. This is necessary because there's apparently no way to get the standard logger from
// the log package.
func CreateNewStdLoggerOrUseExistingLogger(logger *log.Logger) *log.Logger {
	if logger == nil {
		return log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)
	}

	return logger
}
