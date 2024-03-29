// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package context

import (
	"github.com/go-logr/logr"
	"github.com/kiegroup/kogito-operator/core/framework/util"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	logzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"time"
)

var (
	commandOutput io.Writer
	outputFormat  string
	logVerbose    bool
)

// GetDefaultLogger retrieves the default logger
func GetDefaultLogger() *zap.SugaredLogger {
	return getDefaultLoggerWithOut(logVerbose, outputFormat, commandOutput)
}

// GetDefaultLoggerWithOut returns a logger set to a given output
func getDefaultLoggerWithOut(verbose bool, outputFormat string, commandOutput io.Writer) *zap.SugaredLogger {
	var badOutputFormatMsg string
	if len(outputFormat) > 0 && outputFormat != "json" {
		badOutputFormatMsg = "'" + outputFormat + "' is not a supported output format"
		outputFormat = ""
	}
	if commandOutput == nil {
		commandOutput = os.Stdout
	}
	log := GetLoggerWithOptions("kogito-cli", &Opts{
		Output:       commandOutput,
		OutputFormat: outputFormat,
		Verbose:      verbose,
		Console:      true,
	})
	if len(badOutputFormatMsg) > 0 {
		log.Warn(badOutputFormatMsg)
	}
	return log
}

var (
	defaultOutput = os.Stdout
)

// Logger shared logger struct
type Logger struct {
	Logger        logr.Logger
	SugaredLogger *zap.SugaredLogger
}

// Opts describe logger options
type Opts struct {
	// Verbose will increase logging
	Verbose bool
	// Output specifies where to log
	Output io.Writer
	// Output format
	OutputFormat string
	// Console logging doesn't display level nor timestamp and should be readable by humans
	Console bool
}

// GetLoggerWithOptions returns a custom named logger with given options
func GetLoggerWithOptions(name string, options *Opts) *zap.SugaredLogger {
	if options == nil {
		options = getDefaultOpts()
	} else if options.Output == nil {
		options.Output = defaultOutput
	}
	return getLogger(name, options)
}

func getDefaultOpts() *Opts {
	return &Opts{
		Verbose: util.GetBoolOSEnv("DEBUG"),
		Output:  defaultOutput,
		Console: false,
	}
}

func getLogger(name string, options *Opts) *zap.SugaredLogger {
	// Set log level... override default w/ command-line variable if set.
	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniform and structured logs.
	logger := createLogger(options)
	//logger.Logger = logf.Log.WithName(name)
	return logger.SugaredLogger.Named(name)
}

func createLogger(options *Opts) (logger Logger) {
	log := Logger{
		Logger: logzap.New(func(opts *logzap.Options) {
			opts.Development = options.Verbose
		}),
		SugaredLogger: zapSugaredLogger(options),
	}

	logf.SetLogger(log.Logger)
	return log
}

// zapSugaredLogger is a Logger implementation.
// If development is true, a Zap development config will be used,
// otherwise a Zap production config will be used
// (stacktraces on errors, sampling).
func zapSugaredLogger(options *Opts) *zap.SugaredLogger {
	return zapSugaredLoggerTo(options)
}

// zapSugaredLoggerTo returns a new Logger implementation using Zap which logs
// to the given destination, instead of stderr.  It otherise behaves like
// ZapLogger.
func zapSugaredLoggerTo(options *Opts) *zap.SugaredLogger {
	// this basically mimics New<type>Config, but with a custom sink
	sink := zapcore.AddSync(options.Output)

	var enc zapcore.Encoder
	var lvl zap.AtomicLevel
	var opts []zap.Option

	if options.Console {
		encCfg := zap.NewDevelopmentEncoderConfig()
		if options.OutputFormat == "json" {
			encCfg.CallerKey = util.GetOSEnv("KOGITO_LOGGER_CALLER_KEY", "caller")
			encCfg.LevelKey = util.GetOSEnv("KOGITO_LOGGER_LEVEL_KEY", "level")
			encCfg.MessageKey = util.GetOSEnv("KOGITO_LOGGER_MESSAGE_KEY", "message")
			encCfg.NameKey = util.GetOSEnv("KOGITO_LOGGER_NAME_KEY", "name")
			encCfg.StacktraceKey = util.GetOSEnv("KOGITO_LOGGER_STACKTRACE_KEY", "stacktrace")
			encCfg.TimeKey = util.GetOSEnv("KOGITO_LOGGER_TIME_KEY", "time")
			enc = zapcore.NewJSONEncoder(encCfg)
		} else {
			encCfg.CallerKey = util.GetOSEnv("KOGITO_LOGGER_CALLER_KEY", "")
			encCfg.LevelKey = util.GetOSEnv("KOGITO_LOGGER_LEVEL_KEY", "")
			encCfg.MessageKey = util.GetOSEnv("KOGITO_LOGGER_MESSAGE_KEY", encCfg.MessageKey)
			encCfg.NameKey = util.GetOSEnv("KOGITO_LOGGER_NAME_KEY", "")
			encCfg.StacktraceKey = util.GetOSEnv("KOGITO_LOGGER_STACKTRACE_KEY", encCfg.StacktraceKey)
			encCfg.TimeKey = util.GetOSEnv("KOGITO_LOGGER_TIME_KEY", "")
			enc = zapcore.NewConsoleEncoder(encCfg)
		}
		if options.Verbose {
			lvl = zap.NewAtomicLevelAt(zap.DebugLevel)
		} else {
			lvl = zap.NewAtomicLevelAt(zap.InfoLevel)
		}
		opts = append(opts, zap.Development(), zap.AddStacktrace(zap.ErrorLevel))
	} else {
		if options.Verbose {
			encCfg := zap.NewDevelopmentEncoderConfig()
			enc = zapcore.NewConsoleEncoder(encCfg)
			lvl = zap.NewAtomicLevelAt(zap.DebugLevel)
			opts = append(opts, zap.Development(), zap.AddStacktrace(zap.ErrorLevel))
		} else {
			encCfg := zap.NewProductionEncoderConfig()
			encCfg.TimeKey = "T"
			encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
			enc = zapcore.NewJSONEncoder(encCfg)
			lvl = zap.NewAtomicLevelAt(zap.InfoLevel)
			opts = append(opts, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
				return zapcore.NewSamplerWithOptions(core, time.Second, 100, 100)
			}))
		}
	}

	opts = append(opts, zap.AddCallerSkip(1), zap.ErrorOutput(sink))
	log := zap.New(zapcore.NewCore(&logzap.KubeAwareEncoder{Encoder: enc, Verbose: options.Verbose}, sink, lvl))
	log = log.WithOptions(opts...)

	return log.Sugar()
}
