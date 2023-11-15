/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package logger

import (
	"context"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Logger custom logger to define logs in kogito application with levels
type Logger struct {
	logr.Logger
}

// Debug alternative for info format with DEBUG named and correct log level
func (l *Logger) Debug(message string, keysAndValues ...interface{}) {
	l.Logger.WithName("DEBUG").V(1).Info(message, keysAndValues...)
}

// Warn alternative for info format with sprintf and WARN named.
func (l *Logger) Warn(message string, keysAndValues ...interface{}) {
	l.Logger.WithName("WARNING").V(0).Info(message, keysAndValues...)
}

// FromContext returns a logger with predefined values from a context.Context.
func FromContext(ctx context.Context) Logger {
	logger := log.FromContext(ctx)
	return Logger{Logger: logger}
}

// GetLogger returns a custom named logger
func GetLogger(name string) Logger {
	logger := log.Log.WithName(name)
	return Logger{Logger: logger}
}

// WithValues adds some key-value pairs of context to a logger.
func (l *Logger) WithValues(keysAndValues ...interface{}) Logger {
	return Logger{l.Logger.WithValues(keysAndValues...)}
}
