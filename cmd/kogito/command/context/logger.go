// Copyright 2019 Red Hat, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package context

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	"go.uber.org/zap"
	"io"
)

var (
	commandOutput io.Writer
	logVerbose    bool
)

// GetDefaultLogger retrieves the default logger
func GetDefaultLogger() *zap.SugaredLogger {
	return getDefaultLoggerWithOut(logVerbose, commandOutput)
}

// GetDefaultLoggerWithOut returns a logger set to a given output
func getDefaultLoggerWithOut(verbose bool, commandOutput io.Writer) *zap.SugaredLogger {
	return logger.GetLoggerWithOptions("kogito-cli", &logger.Opts{Output: commandOutput, Verbose: verbose, Console: true})
}
