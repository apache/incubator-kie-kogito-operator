package log

import (
	"io"
	"os"

	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
	logrus "github.com/sirupsen/logrus"
)

var logger = logrus.New()

func init() {
	logLevel := logrus.InfoLevel
	// Set log level... override default w/ command-line variable if set.
	if debugBool := util.GetBoolEnv("DEBUG"); debugBool {
		logLevel = logrus.DebugLevel
	}
	logger.SetLevel(logLevel)
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.TextFormatter{})
}

// TODO: create a logger specific for cli
// TODO: create a logger specific for controller (deprecate the controller/kogitoapp/log one). This sould be the default

// SetOutput sets the standard logger output.
func SetOutput(writer io.Writer) {
	logger.SetOutput(writer)
}

// Debug logs a message at level Debug on the standard logger.
func Debug(args ...interface{}) {
	logger.Debug(args...)
}

// Debugf logs a message at level Debug on the standard logger.
func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

// Info logs a message at level Info on the standard logger.
func Info(args ...interface{}) {
	logger.Info(args...)
}

// Infof logs a message at level Info on the standard logger.
func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

// Warn logs a message at level Warn on the standard logger.
func Warn(args ...interface{}) {
	logger.Warn(args...)
}

// Error logs a message at level Error on the standard logger.
func Error(args ...interface{}) {
	logger.Error(args...)
}
