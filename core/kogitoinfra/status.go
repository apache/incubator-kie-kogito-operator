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
	"github.com/kiegroup/kogito-cloud-operator/api"
	"github.com/kiegroup/kogito-cloud-operator/core/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kiegroup/kogito-cloud-operator/core/client/kubernetes"
)

// StatusHandler ...
type StatusHandler interface {
	UpdateBaseStatus(instance api.KogitoInfraInterface, err *error)
}

type statusHandler struct {
	*operator.Context
}

// NewStatusHandler ...
func NewStatusHandler(context *operator.Context) StatusHandler {
	return &statusHandler{
		context,
	}
}

// updateBaseStatus updates the base status for the KogitoInfra instance
func (r *statusHandler) UpdateBaseStatus(instance api.KogitoInfraInterface, err *error) {
	r.Log.Info("Updating Kogito Infra status")
	updateStatus := false
	if *err != nil {
		if reasonForError(*err) == api.ReconciliationFailure {
			r.Log.Info("Seems that an error occurred, setting", "failure state", *err)
		}
		updateStatus = r.setResourceFailed(instance, *err)
	} else {
		updateStatus = r.setResourceSuccess(instance)
		r.Log.Info("Kogito Infra successfully reconciled")
	}
	if updateStatus {
		r.Log.Info("Updating kogitoInfra value with new properties.")
		if resultErr := kubernetes.ResourceC(r.Client).UpdateStatus(instance); resultErr != nil {
			r.Log.Error(resultErr, "reconciliationError occurs while update kogitoInfra values")
		}
		r.Log.Info("Successfully Update Kogito Infra status")
	}
}

// setResourceFailed sets the instance as failed
func (r *statusHandler) setResourceFailed(instance api.KogitoInfraInterface, err error) bool {
	infraCondition := instance.GetStatus().GetCondition()
	if infraCondition.Message != err.Error() {
		r.Log.Warn("Setting instance Failed", "reason", err.Error())
		infraCondition.Type = string(api.FailureInfraConditionType)
		infraCondition.Status = metav1.ConditionFalse
		infraCondition.Message = err.Error()
		infraCondition.Reason = string(reasonForError(err))
		infraCondition.LastTransitionTime = metav1.Now()
		instance.GetStatus().SetCondition(infraCondition)
		return true
	}
	return false
}

// setResourceSuccess sets the instance as success
func (r *statusHandler) setResourceSuccess(instance api.KogitoInfraInterface) bool {
	infraCondition := instance.GetStatus().GetCondition()
	if infraCondition.Type != string(api.SuccessInfraConditionType) {
		r.Log.Warn("Setting instance Successes")
		infraCondition.Type = string(api.SuccessInfraConditionType)
		infraCondition.Status = metav1.ConditionTrue
		infraCondition.Message = ""
		infraCondition.Reason = ""
		infraCondition.LastTransitionTime = metav1.Now()
		instance.GetStatus().SetCondition(infraCondition)
		return true
	}
	return false
}

// setRuntimeProperties sets the instance status' runtime properties
func setRuntimeProperties(instance api.KogitoInfraInterface, runtime api.RuntimeType, runtimeProps api.RuntimePropertiesInterface) {
	instance.GetStatus().AddRuntimeProperties(runtime, runtimeProps)
}
