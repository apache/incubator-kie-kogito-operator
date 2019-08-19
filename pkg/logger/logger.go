package logger

import (
	"io"
	"os"
	"time"

	"github.com/kiegroup/kogito-cloud-operator/pkg/util"

	"github.com/go-logr/logr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var defaultOutputTo = os.Stderr
var outputTo io.Writer = defaultOutputTo

// Logger shared logger struct
type Logger struct {
	Logger        logr.Logger
	SugaredLogger *zap.SugaredLogger
}

// GetDefaultOutput returns the default log output
func GetDefaultOutput() io.Writer {
	return defaultOutputTo
}

// GetLoggerWithOut returns a custom named logger to out in a specific writer
func GetLoggerWithOut(name string, w io.Writer) *zap.SugaredLogger {
	if w == nil {
		outputTo = defaultOutputTo
	} else {
		outputTo = w
	}
	return getLogger(name)
}

// GetLogger returns a custom named logger
func GetLogger(name string) *zap.SugaredLogger {
	outputTo = defaultOutputTo
	return getLogger(name)
}

func getLogger(name string) *zap.SugaredLogger {
	// Set log level... override default w/ command-line variable if set.
	debugBool := util.GetBoolEnv("DEBUG") // info, debug
	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniform and structured logs.
	var logger Logger
	if debugBool {
		logger = createLogger(true)
	} else {
		logger = createLogger(false)
	}
	logger.Logger = logf.Log.WithName(name)
	return logger.SugaredLogger.Named(name)
}

func createLogger(development bool) (logger Logger) {
	log := Logger{
		Logger:        logf.ZapLogger(development),
		SugaredLogger: zapSugaredLogger(development),
	}
	defer log.SugaredLogger.Sync()

	logf.SetLogger(log.Logger)
	return log
}

// zapSugaredLogger is a Logger implementation.
// If development is true, a Zap development config will be used,
// otherwise a Zap production config will be used
// (stacktraces on errors, sampling).
func zapSugaredLogger(development bool) *zap.SugaredLogger {
	return zapSugaredLoggerTo(outputTo, development)
}

// zapSugaredLoggerTo returns a new Logger implementation using Zap which logs
// to the given destination, instead of stderr.  It otherise behaves like
// ZapLogger.
func zapSugaredLoggerTo(destWriter io.Writer, development bool) *zap.SugaredLogger {
	// this basically mimics New<type>Config, but with a custom sink
	sink := zapcore.AddSync(destWriter)

	var enc zapcore.Encoder
	var lvl zap.AtomicLevel
	var opts []zap.Option
	if development {
		encCfg := zap.NewDevelopmentEncoderConfig()
		enc = zapcore.NewConsoleEncoder(encCfg)
		lvl = zap.NewAtomicLevelAt(zap.DebugLevel)
		opts = append(opts, zap.Development(), zap.AddStacktrace(zap.ErrorLevel))
	} else {
		encCfg := zap.NewProductionEncoderConfig()
		enc = zapcore.NewJSONEncoder(encCfg)
		lvl = zap.NewAtomicLevelAt(zap.InfoLevel)
		opts = append(opts, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSampler(core, time.Second, 100, 100)
		}))
	}
	opts = append(opts, zap.AddCallerSkip(1), zap.ErrorOutput(sink))
	log := zap.New(zapcore.NewCore(&logf.KubeAwareEncoder{Encoder: enc, Verbose: development}, sink, lvl))
	log = log.WithOptions(opts...)

	return log.Sugar()
}
