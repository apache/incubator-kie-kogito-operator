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

package kogitoinfra

import (
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
)

// StatusHandler ...
type StatusHandler interface {
	UpdateBaseStatus(instance api.KogitoInfraInterface, err *error)
}

type statusHandler struct {
	client *client.Client
	log    logger.Logger
}

// NewStatusHandler ...
func NewStatusHandler(client *client.Client, log logger.Logger) StatusHandler {
	return &statusHandler{
		client: client,
		log:    log,
	}
}

// updateBaseStatus updates the base status for the KogitoInfra instance
func (r *statusHandler) UpdateBaseStatus(instance api.KogitoInfraInterface, err *error) {
	r.log.Info("Updating Kogito Infra status")
	if *err != nil {
		if reasonForError(*err) == api.ReconciliationFailure {
			r.log.Info("Seems that an error occurred, setting", "failure state", *err)
		}
		r.setResourceFailed(instance, *err)
	} else {
		r.setResourceSuccess(instance)
		r.log.Info("Kogito Infra successfully reconciled")
	}
	r.log.Info("Updating kogitoInfra value with new properties.")
	if resultErr := kubernetes.ResourceC(r.client).UpdateStatus(instance); resultErr != nil {
		r.log.Error(resultErr, "reconciliationError occurs while update kogitoInfra values")
	}
	r.log.Info("Successfully Update Kogito Infra status")
}

// setResourceFailed sets the instance as failed
func (r *statusHandler) setResourceFailed(instance api.KogitoInfraInterface, err error) {
	infraCondition := instance.GetStatus().GetCondition()
	if infraCondition.Message != err.Error() {
		r.log.Warn("Setting instance", "Failed", err)
		infraCondition.Type = api.FailureInfraConditionType
		infraCondition.Status = corev1.ConditionFalse
		infraCondition.Message = err.Error()
		infraCondition.Reason = reasonForError(err)
		infraCondition.LastTransitionTime = metav1.Now()
	}
}

// setResourceSuccess sets the instance as success
func (r *statusHandler) setResourceSuccess(instance api.KogitoInfraInterface) {
	if instance.GetStatus().GetCondition().Type != api.SuccessInfraConditionType {
		infraCondition := instance.GetStatus().GetCondition()
		infraCondition.Type = api.SuccessInfraConditionType
		infraCondition.Status = corev1.ConditionTrue
		infraCondition.Message = ""
		infraCondition.Reason = ""
		infraCondition.LastTransitionTime = metav1.Now()
	}
}

// setRuntimeProperties sets the instance status' runtime properties
func setRuntimeProperties(instance api.KogitoInfraInterface, runtime api.RuntimeType, runtimeProps api.RuntimeProperties) {
	if instance.GetStatus().GetRuntimeProperties() == nil {
		instance.GetStatus().SetRuntimeProperties(api.RuntimePropertiesMap{})
	}
	instance.GetStatus().GetRuntimeProperties()[runtime] = runtimeProps
}
