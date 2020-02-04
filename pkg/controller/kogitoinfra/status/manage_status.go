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
	keycloakv1alpha1 "github.com/keycloak/keycloak-operator/pkg/apis/keycloak/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/keycloak"
	"reflect"
	"time"

	"github.com/RHsyseng/operator-utils/pkg/resource"

	infinispanv1 "github.com/infinispan/infinispan-operator/pkg/apis/infinispan/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/infinispan"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/kafka"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
)

var log = logger.GetLogger("status_kogitoinfra")

const (
	maxConditionsBuffer     = 5
	kafkaBootstrapSvcSuffix = "-kafka-bootstrap"
)

// SetResourceFailed sets the instance as failed
func SetResourceFailed(instance *v1alpha1.KogitoInfra, cli *client.Client, err error) error {
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

// SetResourceSuccess sets the instance as success
func SetResourceSuccess(instance *v1alpha1.KogitoInfra, cli *client.Client) error {
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

// ManageDependenciesStatus will handle with the dependencies status
func ManageDependenciesStatus(instance *v1alpha1.KogitoInfra, cli *client.Client) (requeue bool, err error) {
	log.Infof("Updating Kogito Infra status on namespace %s", instance.Namespace)
	updatedInfinispan, requeueInfinispan := ensureInfinispan(instance, cli)
	updatedKafka, requeueKafka := ensureKafka(instance, cli)
	updatedKeycloak, requeueKeycloak := ensureKeycloak(instance, cli)

	if updatedInfinispan || updatedKafka || updatedKeycloak {
		log.Infof("Updating status of Kogito Infra instance %s ", instance.Status)
		if updatedErr := kubernetes.ResourceC(cli).Update(instance); updatedErr != nil {
			return false, fmt.Errorf("Error while trying to update instance status: %s ", updatedErr)
		}
	}
	return requeueInfinispan || requeueKafka || requeueKeycloak, nil
}

func ensureKeycloak(instance *v1alpha1.KogitoInfra, cli *client.Client) (update, requeue bool) {
	update = false
	requeue = false
	log.Debug("Trying to update Keycloak conditions")
	if !instance.Spec.InstallKeycloak {
		if &instance.Status.Keycloak != nil && len(instance.Status.Keycloak.Condition) > 0 {
			instance.Status.Keycloak = v1alpha1.KeycloakInstallStatus{}
			update = true
		}
		return
	}

	resources, err := keycloak.GetDeployedResources(instance, cli)
	if err != nil {
		update = true
		instance.Status.Keycloak.Condition = pushFailureCondition(instance.Status.Keycloak.Condition, err)
		return
	}

	updateKeycloak, requeueKeycloak := updateNameStatus(instance.Spec.InstallKeycloak, reflect.TypeOf(keycloakv1alpha1.Keycloak{}), &instance.Status.Keycloak.InfraComponentInstallStatusType, resources)
	updateKeycloakRealm, requeueKeycloakRealm := updateNameStatus(instance.Spec.InstallKeycloak, reflect.TypeOf(keycloakv1alpha1.KeycloakRealm{}), &instance.Status.Keycloak.RealmStatus, resources)
	update = updateKeycloak || updateKeycloakRealm
	requeue = requeueKeycloak || requeueKeycloakRealm
	if requeue {
		return
	}

	keycloakRes := resources[reflect.TypeOf(keycloakv1alpha1.Keycloak{})]
	updateSvc := false
	if len(keycloakRes) > 0 {
		requeue = true
		for _, svc := range keycloakRes[0].(*keycloakv1alpha1.Keycloak).Status.SecondaryResources["Service"] {
			if svc == "keycloak" {
				updateSvc, requeue =
					updateServiceStatus(instance, cli, &instance.Status.Keycloak.InfraComponentInstallStatusType, svc)
				break
			}
		}
	}

	update = updateSvc || update

	return
}

func ensureKafka(instance *v1alpha1.KogitoInfra, cli *client.Client) (update, requeue bool) {
	update = false
	requeue = false
	log.Debug("Trying to update Kafka conditions")
	if !instance.Spec.InstallKafka {
		if &instance.Status.Kafka != nil && len(instance.Status.Kafka.Condition) > 0 {
			instance.Status.Kafka = v1alpha1.InfraComponentInstallStatusType{}
			update = true
		}
		return
	}

	resources, err := kafka.GetDeployedResources(instance, cli)
	if err != nil {
		update = true
		instance.Status.Kafka.Condition = pushFailureCondition(instance.Status.Kafka.Condition, err)
		return
	}

	if update, requeue = updateNameStatus(instance.Spec.InstallKafka, reflect.TypeOf(kafkabetav1.Kafka{}), &instance.Status.Kafka, resources); requeue {
		return
	}

	updateSvc, requeue :=
		updateServiceStatus(instance, cli, &instance.Status.Kafka, fmt.Sprintf("%s%s", kafka.InstanceName, kafkaBootstrapSvcSuffix))
	update = updateSvc || update

	return
}

func ensureInfinispan(instance *v1alpha1.KogitoInfra, cli *client.Client) (update, requeue bool) {
	update = false
	log.Debug("Trying to update Infinispan conditions")
	if !instance.Spec.InstallInfinispan {
		if &instance.Status.Infinispan != nil && len(instance.Status.Infinispan.Condition) > 0 {
			instance.Status.Infinispan = v1alpha1.InfinispanInstallStatus{}
			update = true
		}
		return
	}

	resources, err := infinispan.GetDeployedResources(instance, cli)
	if err != nil {
		update = true
		instance.Status.Infinispan.Condition = pushFailureCondition(instance.Status.Infinispan.Condition, err)
		return
	}

	update, requeue = updateNameStatus(instance.Spec.InstallInfinispan, reflect.TypeOf(infinispanv1.Infinispan{}), &instance.Status.Infinispan.InfraComponentInstallStatusType, resources)
	if requeue {
		return
	}

	{
		log.Debug("Updating Infinispan instance Secret conditions")
		secretType := reflect.TypeOf(corev1.Secret{})
		secrets := resources[secretType]
		if len(secrets) == 0 {
			requeue = true
			return
		}
		if instance.Status.Infinispan.CredentialSecret != resources[secretType][0].GetName() {
			update = true
			instance.Status.Infinispan.CredentialSecret = resources[secretType][0].GetName()
		}
	}

	updateSvc, requeue := updateServiceStatus(instance, cli, &instance.Status.Infinispan.InfraComponentInstallStatusType, infinispan.InstanceName)
	update = updateSvc || update

	return
}

// updateNameStatus updates the name status for the infrastructure actor
func updateNameStatus(installRequired bool, resType reflect.Type, infraStatus *v1alpha1.InfraComponentInstallStatusType, resources map[reflect.Type][]resource.KubernetesResource) (update, requeue bool) {
	update = false
	requeue = false
	log.Debug("Updating infra actor instance name conditions")
	instances := resources[resType]
	if len(instances) == 0 {
		// we want it, but we don't have it yet
		if installRequired {
			update = true
			requeue = true
			infraStatus.Condition =
				pushCondition(infraStatus.Condition,
					v1alpha1.InstallCondition{
						Type:               v1alpha1.ProvisioningInstallConditionType,
						Status:             corev1.ConditionFalse,
						LastTransitionTime: metav1.Now(),
					})
		}
		return
	}

	if infraStatus.Name != resources[resType][0].GetName() {
		update = true
		infraStatus.Name = resources[resType][0].GetName()
		infraStatus.Condition =
			pushCondition(infraStatus.Condition,
				v1alpha1.InstallCondition{
					Type:               v1alpha1.SuccessInstallConditionType,
					Status:             corev1.ConditionTrue,
					LastTransitionTime: metav1.Now()})
	}

	return
}

// updateServiceStatus updates the service status for the given infrastructure actor
func updateServiceStatus(instance *v1alpha1.KogitoInfra, cli *client.Client, infraStatus *v1alpha1.InfraComponentInstallStatusType, name string) (update, requeue bool) {
	update = false
	requeue = false
	log.Debug("Updating infra actor instance Service conditions")
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: instance.Namespace,
		},
	}
	if exists, err := kubernetes.ResourceC(cli).Fetch(service); err != nil {
		update = true
		infraStatus.Condition = pushFailureCondition(infraStatus.Condition, err)
		return
	} else if exists {
		if infraStatus.Service != service.Name {
			update = true
			infraStatus.Service = service.Name
		}
		infraStatus.Condition = pushCondition(infraStatus.Condition, v1alpha1.InstallCondition{
			Type:               v1alpha1.SuccessInstallConditionType,
			Status:             corev1.ConditionTrue,
			LastTransitionTime: metav1.Now()})
	} else if !exists {
		update = true
		requeue = true
		infraStatus.Condition = pushCondition(infraStatus.Condition, v1alpha1.InstallCondition{
			Type:               v1alpha1.ProvisioningInstallConditionType,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now()})
	}
	return
}

func pushFailureCondition(conditions []v1alpha1.InstallCondition, err error) (updated []v1alpha1.InstallCondition) {
	updated = pushCondition(conditions,
		v1alpha1.InstallCondition{
			Type:               v1alpha1.FailedInstallConditionType,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Message:            err.Error(),
		})
	return
}

func pushCondition(conditions []v1alpha1.InstallCondition, condition v1alpha1.InstallCondition) (updated []v1alpha1.InstallCondition) {
	length := len(conditions)
	if length > 0 && conditions[length-1].Type == condition.Type {
		return conditions
	}
	size := length + 1
	first := 0
	if size > maxConditionsBuffer {
		first = size - maxConditionsBuffer
	}
	updated = append(conditions, condition)[first:size]
	return
}
