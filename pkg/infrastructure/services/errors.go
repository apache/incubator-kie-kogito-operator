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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1beta1"
	"time"
)

const (
	reconciliationIntervalAfterInfraError                = time.Minute
	reconciliationIntervalAfterMessagingError            = time.Second * 30
	reconciliationIntervalMonitoringEndpointNotAvailable = time.Second * 10
	reconciliationIntervalAfterDashboardsError           = time.Second * 30
	reconciliationIntervalAfterFetchService              = time.Minute * 3
)

type reconciliationError struct {
	reason                 v1beta1.KogitoServiceConditionReason
	reconciliationInterval time.Duration
	innerError             error
}

// String stringer implementation
func (e reconciliationError) String() string {
	return e.innerError.Error()
}

// Error error implementation
func (e reconciliationError) Error() string {
	return e.innerError.Error()
}

func errorForInfraNotReady(service v1beta1.KogitoService, infra *v1beta1.KogitoInfra) reconciliationError {
	return reconciliationError{
		reconciliationInterval: reconciliationIntervalAfterInfraError,
		reason:                 v1beta1.KogitoInfraNotReadyReason,
		innerError: fmt.Errorf("KogitoService '%s' is waiting for infra dependency; skipping deployment; KogitoInfra not ready: %s; Status: %s",
			service.GetName(), infra.Name, infra.Status.Condition.Reason),
	}
}

func errorForMessaging(err error) reconciliationError {
	return reconciliationError{
		reconciliationInterval: reconciliationIntervalAfterMessagingError,
		reason:                 v1beta1.MessagingIntegrationFailureReason,
		innerError:             err,
	}
}

func errorForMonitoring(err error) reconciliationError {
	return reconciliationError{
		reconciliationInterval: reconciliationIntervalMonitoringEndpointNotAvailable,
		reason:                 v1beta1.MonitoringIntegrationFailureReason,
		innerError:             err,
	}
}

func errorForDashboards(err error) reconciliationError {
	return reconciliationError{
		reconciliationInterval: reconciliationIntervalAfterDashboardsError,
		reason:                 v1beta1.MonitoringIntegrationFailureReason,
		innerError:             err,
	}
}

func errorForServiceNotReachable(statusCode int, requestURL string, method string) reconciliationError {
	return reconciliationError{
		reason:                 v1beta1.InternalServiceNotReachable,
		reconciliationInterval: reconciliationIntervalAfterFetchService,
		innerError:             fmt.Errorf("Received NOT expected status code %d while making a %s request to %s ", statusCode, method, requestURL),
	}
}

func reasonForError(err error) v1beta1.KogitoServiceConditionReason {
	if err == nil {
		return ""
	}
	switch t := err.(type) {
	case reconciliationError:
		return t.reason
	}
	return v1beta1.ServiceReconciliationFailure
}

func reconciliationIntervalForError(err error) time.Duration {
	switch t := err.(type) {
	case reconciliationError:
		return t.reconciliationInterval
	}
	return 0
}
