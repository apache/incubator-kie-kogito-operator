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
	"time"

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
		if statusErr := setResourceFailed(instance, client, *err); statusErr != nil {
			err = &statusErr
			log.Errorf("Error in setting status failes: %v", *err)
		}
	} else {
		log.Info("Kogito Infra successfully reconciled")
		if statusErr := setResourceSuccess(instance, client); statusErr != nil {
			err = &statusErr
			log.Errorf("Error in setting status failes: %v", *err)
		}
	}
}

// setResourceFailed sets the instance as failed
func setResourceFailed(instance *v1alpha1.KogitoInfra, cli *client.Client, err error) error {
	if instance.Status.Condition.Message != err.Error() {
		log.Warn("Setting instance as failed", err)
		instance.Status.Condition.Type = v1alpha1.FailureInfraConditionType
		instance.Status.Condition.Status = corev1.ConditionFalse
		instance.Status.Condition.Message = err.Error()
		instance.Status.Condition.LastTransitionTime = metav1.Now().Format(time.RFC3339)

		if err := kubernetes.ResourceC(cli).Update(instance); err != nil {
			return err
		}
	}

	return nil
}

// setResourceSuccess sets the instance as success
func setResourceSuccess(instance *v1alpha1.KogitoInfra, cli *client.Client) error {
	if instance.Status.Condition.Type != v1alpha1.SuccessInfraConditionType {
		instance.Status.Condition.Type = v1alpha1.SuccessInfraConditionType
		instance.Status.Condition.Status = corev1.ConditionTrue
		instance.Status.Condition.Message = ""
		instance.Status.Condition.LastTransitionTime = metav1.Now().Format(time.RFC3339)

		if err := kubernetes.ResourceC(cli).Update(instance); err != nil {
			return err
		}
	}
	return nil
}
