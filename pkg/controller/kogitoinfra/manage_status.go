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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
)

// updateBaseStatus updates the base status for the KogitoInfra instance
func updateBaseStatus(client *client.Client, instance *v1alpha1.KogitoInfra, err *error) {
	log.Info("Updating Kogito Infra status")
	if *err != nil {
		log.Warn("Seems that an error occurred, setting failure state: ", *err)
		setResourceFailed(instance, *err)
	} else {
		setResourceSuccess(instance)
		log.Info("Kogito Infra successfully reconciled")
	}
	log.Infof("Updating kogitoInfra value with new properties : %s", instance.Name)
	if resultErr := kubernetes.ResourceC(client).UpdateStatus(instance); resultErr != nil {
		log.Errorf("reconciliationError occurs while update kogitoInfra values: %v", resultErr)
	}
	log.Info("Successfully Update Kogito Infra status")
}

// setResourceFailed sets the instance as failed
func setResourceFailed(instance *v1alpha1.KogitoInfra, err error) {
	if instance.Status.Condition.Message != err.Error() {
		log.Warn("Setting instance as failed", err)
		instance.Status.Condition.Type = v1alpha1.FailureInfraConditionType
		instance.Status.Condition.Status = corev1.ConditionFalse
		instance.Status.Condition.Message = err.Error()
		instance.Status.Condition.Reason = reasonForError(err)
		instance.Status.Condition.LastTransitionTime = metav1.Now()
	}
}

// setResourceSuccess sets the instance as success
func setResourceSuccess(instance *v1alpha1.KogitoInfra) {
	if instance.Status.Condition.Type != v1alpha1.SuccessInfraConditionType {
		instance.Status.Condition.Type = v1alpha1.SuccessInfraConditionType
		instance.Status.Condition.Status = corev1.ConditionTrue
		instance.Status.Condition.Message = ""
		instance.Status.Condition.Reason = ""
		instance.Status.Condition.LastTransitionTime = metav1.Now()
	}
}
