package inmemory

import (
	"fmt"
	"log"
)

const loggingPrefix = "local_inmem_test"

type StdDebugLogger struct{}

func (l StdDebugLogger) Fatalf(format string, v ...interface{}) {
	log.Fatalf(fmt.Sprintf("%s FATAL: %s", loggingPrefix, format), v...)
}

func (l StdDebugLogger) Errorf(format string, v ...interface{}) {
	log.Printf(fmt.Sprintf("%s ERROR: %s", loggingPrefix, format), v...)
}

func (l StdDebugLogger) Warnf(format string, v ...interface{}) {
	log.Printf(fmt.Sprintf("%s WARN: %s", loggingPrefix, format), v...)
}

func (l StdDebugLogger) Infof(format string, v ...interface{}) {
	log.Printf(fmt.Sprintf("%s INFO: %s", loggingPrefix, format), v...)
}

func (l StdDebugLogger) Debugf(format string, v ...interface{}) {
	log.Printf(fmt.Sprintf("%s DEBUG: %s", loggingPrefix, format), v...)
}

func (l StdDebugLogger) Tracef(_ string, _ ...interface{}) {}

type EmptyLogger struct{}

func (l EmptyLogger) Fatalf(_ string, _ ...interface{}) {}

func (l EmptyLogger) Errorf(_ string, _ ...interface{}) {}

func (l EmptyLogger) Warnf(_ string, _ ...interface{}) {}

func (l EmptyLogger) Infof(_ string, _ ...interface{}) {}

func (l EmptyLogger) Debugf(_ string, _ ...interface{}) {}

func (l EmptyLogger) Tracef(_ string, _ ...interface{}) {}
