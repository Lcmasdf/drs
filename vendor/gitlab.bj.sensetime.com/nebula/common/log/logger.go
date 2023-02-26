package log

import (
	"fmt"
	"runtime"
	"time"

	logcore "gitlab.bj.sensetime.com/nebula/common/log/core"
)

// Logger defines all methods that can be called by user within a logger.
type Logger interface {
	// With creates a child logger and adds structured context to it. Fields added
	// to the child don't affect the parent, and vice versa.
	With(fields ...Field) Logger

	// Debug logs a message at DebugLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	Debug(msg string)

	// Debugf uses fmt.Sprintf to log a templated message. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	Debugf(template string, args ...interface{})

	// Info logs a message at InfoLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	Info(msg string)

	// Infof uses fmt.Sprintf to log a templated message. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	Infof(template string, args ...interface{})

	// Warn logs a message at WarnLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	Warn(msg string)

	// Warnf uses fmt.Sprintf to log a templated message. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	Warnf(template string, args ...interface{})

	// Error logs a message at ErrorLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	Error(msg string)

	// Errorf uses fmt.Sprintf to log a templated message. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	Errorf(template string, args ...interface{})

	// DPanic logs a message at DPanicLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	DPanic(msg string)

	// DPanicf uses fmt.Sprintf to log a templated message. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	DPanicf(template string, args ...interface{})

	// Panic logs a message at PanicLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	Panic(msg string)

	// Panicf uses fmt.Sprintf to log a templated message. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	Panicf(template string, args ...interface{})

	// Fatal logs a message at FatalLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	Fatal(msg string)

	// Fatalf uses fmt.Sprintf to log a templated message. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	Fatalf(template string, args ...interface{})

	// Sync calls the underlying Core's Sync method, flushing any buffered log
	// entries. Applications should take care to call Sync before exiting.
	Sync() error

	check(lvl logcore.Level, template string, fmtArgs ...interface{}) *logcore.CheckedEntry
}

// New constructs a logger for user to use with user-defiend config.
func New(config Config) (logger Logger, err error) {
	return config.build()
}

// NewWithCore constructs a logger with a core.
func NewWithCore(core logcore.Core, config Config) (logger Logger, err error) {
	return config.buildWithCore(core)
}

type logger struct {
	core logcore.Core

	name        string
	errorOutput logcore.WriteSyncer

	addCaller  bool
	callerSkip int

	addStack    logcore.LevelEnabler
	development bool
}

// With creates a child logger and adds structured context to it. Fields added
// to the child don't affect the parent, and vice versa.
func (log *logger) With(fields ...Field) Logger {
	if len(fields) == 0 {
		return log
	}
	l := log.clone()
	l.core = l.core.With(fields)
	return l
}

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *logger) Debug(msg string) {
	if ce := log.check(DebugLevel, msg); ce != nil {
		ce.Write()
	}
}

// Debugf uses fmt.Sprintf to log a templated message. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *logger) Debugf(msg string, args ...interface{}) {
	if ce := log.check(DebugLevel, msg, args...); ce != nil {
		ce.Write()
	}
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *logger) Info(msg string) {
	if ce := log.check(InfoLevel, msg); ce != nil {
		ce.Write()
	}
}

// Infof uses fmt.Sprintf to log a templated message. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *logger) Infof(msg string, args ...interface{}) {
	if ce := log.check(InfoLevel, msg, args...); ce != nil {
		ce.Write()
	}
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *logger) Warn(msg string) {
	if ce := log.check(WarnLevel, msg); ce != nil {
		ce.Write()
	}
}

// Warnf uses fmt.Sprintf to log a templated message. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *logger) Warnf(msg string, args ...interface{}) {
	if ce := log.check(WarnLevel, msg, args...); ce != nil {
		ce.Write()
	}
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *logger) Error(msg string) {
	if ce := log.check(ErrorLevel, msg); ce != nil {
		ce.Write()
	}
}

// Errorf uses fmt.Sprintf to log a templated message. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *logger) Errorf(msg string, args ...interface{}) {
	if ce := log.check(ErrorLevel, msg, args...); ce != nil {
		ce.Write()
	}
}

// DPanic logs a message at DPanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *logger) DPanic(msg string) {
	if ce := log.check(DPanicLevel, msg); ce != nil {
		ce.Write()
	}
}

// DPanicf uses fmt.Sprintf to log a templated message. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *logger) DPanicf(msg string, args ...interface{}) {
	if ce := log.check(DPanicLevel, msg, args...); ce != nil {
		ce.Write()
	}
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *logger) Panic(msg string) {
	if ce := log.check(PanicLevel, msg); ce != nil {
		ce.Write()
	}
}

// Panicf uses fmt.Sprintf to log a templated message. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *logger) Panicf(msg string, args ...interface{}) {
	if ce := log.check(PanicLevel, msg, args...); ce != nil {
		ce.Write()
	}
}

// Fatal logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *logger) Fatal(msg string) {
	if ce := log.check(FatalLevel, msg); ce != nil {
		ce.Write()
	}
}

// Fatalf uses fmt.Sprintf to log a templated message. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *logger) Fatalf(msg string, args ...interface{}) {
	if ce := log.check(FatalLevel, msg, args...); ce != nil {
		ce.Write()
	}
}

// Sync calls the underlying Core's Sync method, flushing any buffered log
// entries. Applications should take care to call Sync before exiting.
func (log *logger) Sync() error {
	return log.core.Sync()
}

func (log *logger) clone() *logger {
	copy := *log
	return &copy
}

func (log *logger) check(lvl logcore.Level, template string, fmtArgs ...interface{}) *logcore.CheckedEntry {
	// check must always be called directly by a method in the Logger interface
	// (e.g., Check, Info, Fatal).
	const callerSkipOffset = 2

	// Check the level first to reduce the cost of disabled log calls.
	// Since Panic and higher may exit, we skip the optimization for those levels.
	if lvl < logcore.DPanicLevel && !log.core.Enabled(lvl) {
		return nil
	}

	// Format with Sprint, Sprintf, or neither.
	msg := template
	if msg == "" && len(fmtArgs) > 0 {
		msg = fmt.Sprint(fmtArgs...)
	} else if msg != "" && len(fmtArgs) > 0 {
		msg = fmt.Sprintf(template, fmtArgs...)
	}

	// Create basic checked entry thru the core; this will be non-nil if the
	// log message will actually be written somewhere.
	ent := logcore.Entry{
		LoggerName: log.name,
		Time:       time.Now(),
		Level:      lvl,
		Message:    msg,
	}
	ce := log.core.Check(ent, nil)
	willWrite := ce != nil

	// Set up any required terminal behavior.
	switch ent.Level {
	case logcore.PanicLevel:
		ce = ce.Should(ent, logcore.WriteThenPanic)
	case logcore.FatalLevel:
		ce = ce.Should(ent, logcore.WriteThenFatal)
	case logcore.DPanicLevel:
		if log.development {
			ce = ce.Should(ent, logcore.WriteThenPanic)
		}
	}

	// Only do further annotation if we're going to write this message; checked
	// entries that exist only for terminal behavior don't benefit from
	// annotation.
	if !willWrite {
		return ce
	}

	// Thread the error output through to the CheckedEntry.
	ce.ErrorOutput = log.errorOutput
	if log.addCaller {
		ce.Entry.Caller = logcore.NewEntryCaller(runtime.Caller(log.callerSkip + callerSkipOffset))
		if !ce.Entry.Caller.Defined {
			fmt.Fprintf(log.errorOutput, "%v Logger.check error: failed to get caller\n", time.Now().UTC())
			log.errorOutput.Sync()
		}
	}
	if log.addStack.Enabled(ce.Entry.Level) {
		ce.Entry.Stack = Stack("").String
	}

	return ce
}
