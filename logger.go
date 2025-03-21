package inmemory

import (
	"fmt"
	"log"

	inmemLogger "github.com/mopo3ula/inmempubsub/internal/logger"
)

type StdLogger inmemLogger.Logger

const loggingPrefix = "local_inmem"

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
