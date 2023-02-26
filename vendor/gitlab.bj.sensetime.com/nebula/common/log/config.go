package log

import (
	"strings"

	logcore "gitlab.bj.sensetime.com/nebula/common/log/core"
	"gitlab.bj.sensetime.com/nebula/common/log/internal/lumberjack"
)

const (
	defaultLevel            = "info"
	defaultStackEnableLevel = "dpanic"
	defaultEncoding         = "json"
	defaultOutputPath       = "stderr"
	defaultErrorOutputPath  = "stderr"
)

// Config defines all declarative config of logger.
type Config struct {
	// Level is the minimum enabled logging level.
	Level string `json:"level" toml:"level" yaml:"level"`

	// Development puts the logger in development mode, which changes the
	// behavior of DPanicLevel and takes stacktraces more liberally.
	Development bool `json:"development" toml:"development" yaml:"development"`

	CallerSkip int `json:"caller_skip" toml:"caller_skip" yaml:"caller_skip"`

	// DisableCaller stops annotating logs with the calling function's file
	// name and line number. By default, all logs are annotated.
	DisableCaller bool `json:"disable_caller" toml:"disable_caller" yaml:"disable_caller"`

	// StackEnableLevel is minimum enable logging level which will contain code stack.
	StackEnableLevel string `json:"stack_enable_level" toml:"stack_enable_level" yaml:"stack_enable_level"`

	// Encoding sets the logger's encoding. Valid values are "json" and
	// "console", as well as any third-party encodings registered via
	// RegisterEncoder.
	Encoding string `json:"encoding" toml:"encoding" yaml:"encoding"`

	// OutputPath is an URL or file paths to write logging output to.
	// See Open for details.
	OutputPath string `json:"output_path" toml:"output_path" yaml:"output_path"`

	// ErrorOutputPath is an URL to write internal logger errors to.
	// The default is standard error.
	//
	// Note that this setting only affects internal errors; for sample code that
	// sends error-level logs to a different location from info- and debug-level
	// logs, see the package-level AdvancedConfiguration example.
	ErrorOutputPath string `json:"error_output_path" toml:"error_output_path" yaml:"error_output_path"`

	// MaxSize set the log rotation
	// log rotation with lumberjack.
	// exported the log config from lumberjack
	// is the maximum size in megabytes of the log file before it gets
	// rotated. It defaults to 100 megabytes
	MaxSize int `json:"max_size" toml:"max_size" yaml:"max_size"`

	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.  Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc. The default is not to remove old log files
	// based on age. exported from lumberjack
	MaxAge int `json:"max_age" toml:"max_age" yaml:"max_age"`

	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files (though MaxAge may still cause them to get
	// deleted.)
	MaxBackups int `json:"max_backups" toml:"max_backups" yaml:"max_backups"`
}

func (cfg *Config) build() (lo Logger, err error) {
	if cfg.Level == "" {
		cfg.Level = defaultLevel
	}
	if cfg.StackEnableLevel == "" {
		cfg.StackEnableLevel = defaultStackEnableLevel
	}
	if cfg.Encoding == "" {
		cfg.Encoding = defaultEncoding
	}
	if cfg.ErrorOutputPath == "" {
		cfg.ErrorOutputPath = defaultErrorOutputPath
	}
	if cfg.OutputPath == "" {
		cfg.OutputPath = defaultOutputPath
	}

	enc, err := cfg.buildEncoder()
	if err != nil {
		return
	}

	sink, errSink, err := cfg.openSinks()
	if err != nil {
		return
	}

	core := logcore.NewCore(enc, sink, cfg.level(cfg.Level))

	l := &logger{
		core:        core,
		errorOutput: errSink,
		addStack:    cfg.level(cfg.StackEnableLevel),
		development: cfg.Development,
		addCaller:   true,
	}
	if cfg.DisableCaller {
		l.addCaller = false
	}
	if cfg.CallerSkip > 0 {
		l.callerSkip = cfg.CallerSkip
	}

	lo = l

	return
}

func (cfg *Config) buildWithCore(core logcore.Core) (lo Logger, err error) {
	if cfg.Level == "" {
		cfg.Level = defaultLevel
	}
	if cfg.StackEnableLevel == "" {
		cfg.StackEnableLevel = defaultStackEnableLevel
	}
	if cfg.ErrorOutputPath == "" {
		cfg.ErrorOutputPath = defaultErrorOutputPath
	}
	if cfg.OutputPath == "" {
		cfg.OutputPath = defaultOutputPath
	}

	_, errSink, err := cfg.openSinks()
	if err != nil {
		return
	}

	l := &logger{
		core:        core,
		errorOutput: errSink,
		addStack:    cfg.level(cfg.StackEnableLevel),
		development: cfg.Development,
	}
	if cfg.DisableCaller {
		l.addCaller = false
	}
	if cfg.CallerSkip > 0 {
		l.callerSkip = cfg.CallerSkip
	}

	lo = l
	return
}

func (cfg *Config) level(level string) (lv logcore.Level) {
	switch strings.ToUpper(level) {
	case "DEBUG":
		lv = DebugLevel
	case "INFO":
		lv = InfoLevel
	case "WARN":
		lv = WarnLevel
	case "ERROR":
		lv = ErrorLevel
	case "DPANIC":
		lv = DPanicLevel
	case "PANIC":
		lv = PanicLevel
	case "FATAL":
		lv = FatalLevel
	default:
		lv = WarnLevel
	}
	return
}

func (cfg *Config) buildEncoder() (logcore.Encoder, error) {
	encConfig := logcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     logcore.DefaultLineEnding,
		EncodeLevel:    logcore.CapitalLevelEncoder,
		EncodeTime:     logcore.ISO8601TimeEncoder,
		EncodeDuration: logcore.SecondsDurationEncoder,
		EncodeCaller:   logcore.ShortCallerEncoder,
	}

	return newEncoder(cfg.Encoding, encConfig)
}

func (cfg *Config) openSinks() (logcore.WriteSyncer, logcore.WriteSyncer, error) {
	var sink, errSink logcore.WriteSyncer
	var closeOut func()
	var err error
	// if path != std, it will be use lumberjack logger
	if cfg.OutputPath != defaultOutputPath {
		sink = logcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.OutputPath,
			MaxSize:    cfg.MaxSize,
			MaxAge:     cfg.MaxAge,
			MaxBackups: cfg.MaxBackups,
		})
	} else {
		sink, closeOut, err = Open(cfg.OutputPath)
		if err != nil {
			return nil, nil, err
		}
	}

	// if path != std, it will be use lumberjack logger
	if cfg.ErrorOutputPath != defaultErrorOutputPath {
		errSink = logcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.OutputPath,
			MaxSize:    cfg.MaxSize,
			MaxAge:     cfg.MaxAge,
			MaxBackups: cfg.MaxBackups,
		})
	} else {
		errSink, _, err = Open(cfg.ErrorOutputPath)
		if err != nil {
			if closeOut != nil {
				closeOut()
			}
			return nil, nil, err
		}
	}
	return sink, errSink, nil

}
