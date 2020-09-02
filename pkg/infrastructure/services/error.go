// Copyright 2020 Red Hat, Inc. and/or its affiliates
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

package services

import (
	"fmt"
	"time"
)

const (
	serviceNotReadyMessage          = "%s service not found in namespace %s and it's required for %s. Replicas will be set to 0."
	serviceDependencyReconcileAfter = time.Minute * 2
)

// requiresReconciliationError indicates that something went wrong and the controller needs a new reconciliation after a given time
type requiresReconciliationError interface {
	error
	// GetReconcileAfter gets the time duration for a new reconciliation
	GetReconcileAfter() time.Duration
}

func isRequiresReconciliationError(e error) bool {
	if _, ok := e.(requiresReconciliationError); ok {
		return true
	}
	return false
}

type requiredServiceNotReadyError struct {
	message        string
	reconcileAfter time.Duration
}

func (r requiredServiceNotReadyError) Error() string {
	return r.message
}

func (r requiredServiceNotReadyError) GetReconcileAfter() time.Duration {
	return r.reconcileAfter
}

func newKogitoServiceNotReadyError(namespace, serviceName string, requiredServiceName string) requiredServiceNotReadyError {
	return requiredServiceNotReadyError{
		message:        fmt.Sprintf(serviceNotReadyMessage, requiredServiceName, namespace, serviceName),
		reconcileAfter: serviceDependencyReconcileAfter,
	}
}
