package log

import (
	"log"
	"sync"
)

var (
	globalLogger Logger

	createdLoggersMu sync.Mutex
)

func init() {
	globalLogger, _ = New(Config{})
}

// Init recreates a new globalogger and close old logger.
// It's better for user to call this method only once.
func Init(config Config) (err error) {
	if err != nil {
		log.Fatal(err)
	}
	createdLoggersMu.Lock()
	defer createdLoggersMu.Unlock()

	newLogger, err := New(config)
	if err != nil {
		return
	}

	oldLogger := globalLogger
	globalLogger = newLogger

	oldCloseChan := make(chan struct{})
	go func() {
		oldLogger.Sync()
		close(oldCloseChan)
	}()

	<-oldCloseChan

	return
}

// With creates a child logger and adds structured context to it. Fields added
// to the child don't affect the parent, and vice versa.
func With(fields ...Field) Logger {
	return globalLogger.With(fields...)
}

// Debug logs a message at DebugLevel.
func Debug(msg string) {
	if ce := globalLogger.check(DebugLevel, msg); ce != nil {
		ce.Write()
	}
}

// Debugf uses fmt.Sprintf to log a templated message.
func Debugf(template string, args ...interface{}) {
	if ce := globalLogger.check(DebugLevel, template, args...); ce != nil {
		ce.Write()
	}
}

// Info logs a message at InfoLevel.
func Info(msg string) {
	if ce := globalLogger.check(InfoLevel, msg); ce != nil {
		ce.Write()
	}
}

// Infof uses fmt.Sprintf to log a templated message.
func Infof(template string, args ...interface{}) {
	if ce := globalLogger.check(InfoLevel, template, args...); ce != nil {
		ce.Write()
	}
}

// Warn logs a message at WarnLevel.
func Warn(msg string) {
	if ce := globalLogger.check(WarnLevel, msg); ce != nil {
		ce.Write()
	}
}

// Warnf uses fmt.Sprintf to log a templated message.
func Warnf(template string, args ...interface{}) {
	if ce := globalLogger.check(WarnLevel, template, args...); ce != nil {
		ce.Write()
	}
}

// Error logs a message at ErrorLevel.
func Error(msg string) {
	if ce := globalLogger.check(ErrorLevel, msg); ce != nil {
		ce.Write()
	}
}

// Errorf uses fmt.Sprintf to log a templated message.
func Errorf(template string, args ...interface{}) {
	if ce := globalLogger.check(ErrorLevel, template, args...); ce != nil {
		ce.Write()
	}
}

// DPanic logs a message at DPanicLevel.
func DPanic(msg string) {
	if ce := globalLogger.check(DPanicLevel, msg); ce != nil {
		ce.Write()
	}
}

// DPanicf uses fmt.Sprintf to log a templated message.
func DPanicf(template string, args ...interface{}) {
	if ce := globalLogger.check(DPanicLevel, template, args...); ce != nil {
		ce.Write()
	}
}

// Panic logs a message at PanicLevel.
func Panic(msg string) {
	if ce := globalLogger.check(PanicLevel, msg); ce != nil {
		ce.Write()
	}
}

// Panicf uses fmt.Sprintf to log a templated message.
func Panicf(template string, args ...interface{}) {
	if ce := globalLogger.check(PanicLevel, template, args...); ce != nil {
		ce.Write()
	}
}

// Fatal logs a message at FatalLevel.
func Fatal(msg string) {
	if ce := globalLogger.check(FatalLevel, msg); ce != nil {
		ce.Write()
	}
}

// Fatalf uses fmt.Sprintf to log a templated message.
func Fatalf(template string, args ...interface{}) {
	if ce := globalLogger.check(FatalLevel, template, args...); ce != nil {
		ce.Write()
	}
}

// Sync calls the underlying Core's Sync method, flushing any buffered log
// entries. Applications should take care to call Sync before exiting.
func Sync() error {
	return globalLogger.Sync()
}
