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

package status

import (
	"fmt"
	v1 "github.com/infinispan/infinispan-operator/pkg/apis/infinispan/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/infinispan"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
)

var log = logger.GetLogger("status_kogitoinfra")

const maxConditionsBuffer = 5

// SetResourceFailed sets the instance as failed
func SetResourceFailed(instance *v1alpha1.KogitoInfra, cli *client.Client, err error) error {
	log.Warn("Setting instance as failed", err)
	instance.Status.Condition.Type = v1alpha1.FailureInfraConditionType
	instance.Status.Condition.Status = corev1.ConditionFalse
	instance.Status.Condition.Message = err.Error()
	instance.Status.Condition.LastTransitionTime = metav1.Now()

	if err := kubernetes.ResourceC(cli).UpdateStatus(instance); err != nil {
		return err
	}
	return nil
}

// SetResourceSuccess sets the instance as success
func SetResourceSuccess(instance *v1alpha1.KogitoInfra, cli *client.Client) error {
	instance.Status.Condition.Type = v1alpha1.SuccessInfraConditionType
	instance.Status.Condition.Status = corev1.ConditionTrue
	instance.Status.Condition.LastTransitionTime = metav1.Now()

	if err := kubernetes.ResourceC(cli).UpdateStatus(instance); err != nil {
		return err
	}
	return nil
}

// ManageDependenciesStatus will handle with the dependencies status
func ManageDependenciesStatus(instance *v1alpha1.KogitoInfra, cli *client.Client) error {
	log.Infof("Updating Kogito Infra status on namespace %s", instance.Namespace)
	updatedInfinispan := ensureInfinispan(instance, cli)
	if updatedInfinispan {
		log.Infof("Updating status of Kogito Infra instance %s ", instance.Status)
		if updatedErr := kubernetes.ResourceC(cli).UpdateStatus(instance); updatedErr != nil {
			return fmt.Errorf("Error while trying to update instance status: %s ", updatedErr)
		}
	}

	return nil
}

func ensureInfinispan(instance *v1alpha1.KogitoInfra, cli *client.Client) (update bool) {
	log.Debug("Trying to update Infinispan conditions")
	update = false
	resources, err := infinispan.GetDeployedResources(instance, cli)
	if err != nil {
		update = true
		instance.Status.Infinispan.Condition = pushFailureCondition(instance, err)
		return update
	}

	{
		log.Debug("Updating infinispan instance name conditions")
		infinispanType := reflect.TypeOf(v1.Infinispan{})
		infinispanInstances := resources[infinispanType]
		if len(infinispanInstances) == 0 {
			// we want infinispan, but we don't have it yet
			if instance.Spec.InstallInfinispan {
				update = true
				instance.Status.Infinispan.Condition =
					pushCondition(instance.Status.Infinispan.Condition,
						v1alpha1.InstallCondition{
							Type:               v1alpha1.ProvisioningInstallConditionType,
							Status:             corev1.ConditionFalse,
							LastTransitionTime: metav1.Now(),
						})
			}
			return update
		}

		if instance.Status.Infinispan.Name != resources[infinispanType][0].GetName() {
			update = true
			instance.Status.Infinispan.Name = resources[infinispanType][0].GetName()
			instance.Status.Infinispan.Condition =
				pushCondition(instance.Status.Infinispan.Condition,
					v1alpha1.InstallCondition{
						Type:               v1alpha1.SuccessInstallConditionType,
						Status:             corev1.ConditionTrue,
						LastTransitionTime: metav1.Now(),
					})
		}
	}

	{
		log.Debug("Updating Infinispan instance Secret conditions")
		secretType := reflect.TypeOf(corev1.Secret{})
		secrets := resources[secretType]
		if len(secrets) == 0 {
			return update
		}
		if instance.Status.Infinispan.CredentialSecret != resources[secretType][0].GetName() {
			update = true
			instance.Status.Infinispan.CredentialSecret = resources[secretType][0].GetName()
		}
	}

	{
		log.Debug("Updating Infinispan instance Service conditions")
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      infinispan.InstanceName,
				Namespace: instance.Namespace,
			},
		}
		if exists, err := kubernetes.ResourceC(cli).Fetch(service); err != nil {
			update = true
			instance.Status.Infinispan.Condition = pushFailureCondition(instance, err)
			return update
		} else if exists {
			if instance.Status.Infinispan.Service != service.Name {
				update = true
				instance.Status.Infinispan.Service = service.Name
			}
		}
	}

	return update
}

func pushFailureCondition(instance *v1alpha1.KogitoInfra, err error) (updated []v1alpha1.InstallCondition) {
	updated = pushCondition(instance.Status.Infinispan.Condition,
		v1alpha1.InstallCondition{
			Type:               v1alpha1.FailedInstallConditionType,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Message:            err.Error(),
		})
	return
}

func pushCondition(conditions []v1alpha1.InstallCondition, condition v1alpha1.InstallCondition) (updated []v1alpha1.InstallCondition) {
	size := len(conditions) + 1
	first := 0
	if size > maxConditionsBuffer {
		first = size - maxConditionsBuffer
	}
	updated = append(conditions, condition)[first:size]
	return
}
