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

package controllers

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
)

// updateBaseStatus updates the base status for the KogitoInfra instance
func (r *KogitoInfraReconciler) updateBaseStatus(client *client.Client, instance *v1beta1.KogitoInfra, err *error) {
	r.Log.Info("Updating Kogito Infra status")
	if *err != nil {
		if reasonForError(*err) == v1beta1.ReconciliationFailure {
			r.Log.Info("Seems that an error occurred, setting", "failure state", *err)
		}
		r.setResourceFailed(instance, *err)
	} else {
		setResourceSuccess(instance)
		r.Log.Info("Kogito Infra successfully reconciled")
	}
	r.Log.Info("Updating kogitoInfra value with new properties in", "Instance", instance.Name)
	if resultErr := kubernetes.ResourceC(client).UpdateStatus(instance); resultErr != nil {
		r.Log.Error(resultErr, "reconciliationError occurs while update kogitoInfra values")
	}
	r.Log.Info("Successfully Update Kogito Infra status")
}

// setResourceFailed sets the instance as failed
func (r *KogitoInfraReconciler) setResourceFailed(instance *v1beta1.KogitoInfra, err error) {
	if instance.Status.Condition.Message != err.Error() {
		r.Log.Warn("Setting instance", "Failed", err)
		instance.Status.Condition.Type = v1beta1.FailureInfraConditionType
		instance.Status.Condition.Status = corev1.ConditionFalse
		instance.Status.Condition.Message = err.Error()
		instance.Status.Condition.Reason = reasonForError(err)
		instance.Status.Condition.LastTransitionTime = metav1.Now()
	}
}

// setResourceSuccess sets the instance as success
func setResourceSuccess(instance *v1beta1.KogitoInfra) {
	if instance.Status.Condition.Type != v1beta1.SuccessInfraConditionType {
		instance.Status.Condition.Type = v1beta1.SuccessInfraConditionType
		instance.Status.Condition.Status = corev1.ConditionTrue
		instance.Status.Condition.Message = ""
		instance.Status.Condition.Reason = ""
		instance.Status.Condition.LastTransitionTime = metav1.Now()
	}
}

// setRuntimeProperties sets the instance status' runtime properties
func setRuntimeProperties(instance *v1beta1.KogitoInfra, runtime v1beta1.RuntimeType, runtimeProps v1beta1.RuntimeProperties) {
	if instance.Status.RuntimeProperties == nil {
		instance.Status.RuntimeProperties = v1beta1.RuntimePropertiesMap{}
	}
	instance.Status.RuntimeProperties[runtime] = runtimeProps
}
