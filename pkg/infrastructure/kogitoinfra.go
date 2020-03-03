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

package infrastructure

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// DefaultKogitoInfraName is the default name given to the Kogito Infra.
	// We're not attached to this name, but since we're going to create it automagically, it's better to have a standard one.
	DefaultKogitoInfraName = "kogito-infra"

	noActionComponentState componentState = "noaction"
	installComponentState  componentState = "install"
	removeComponentState   componentState = "remove"
)

type componentState string

type ensureComponent struct {
	namespace string
	client    *client.Client

	infinispan componentState
	kafka      componentState
	keycloak   componentState
}

// EnsureComponent interface to control how to provision an infra with a given component
type EnsureComponent interface {
	// WithInfinispan creates a new instance of KogitoInfra if not exists with an Infinispan deployed if not exists. If exists, checks if Infinispan is deployed.
	WithInfinispan() EnsureComponent
	// WithKafka creates a new instance of KogitoInfra if not exists with a Kafka deployed if not exists. If exists, checks if Kafka is deployed.
	WithKafka() EnsureComponent
	// WithKeycloak creates a new instance of KogitoInfra if not exists with a Keycloak deployed if not exists. If exists, checks if Keycloak is deployed.
	WithKeycloak() EnsureComponent

	// WithoutInfinispan deletes instance of Infinispan if exists.
	WithoutInfinispan() EnsureComponent
	// WithoutKafka deletes instance of Infinispan if exists.
	WithoutKafka() EnsureComponent
	// WithoutKeycloak deletes instance of Keycloak if exists.
	WithoutKeycloak() EnsureComponent

	Apply() (infra *v1alpha1.KogitoInfra, ready bool, err error)
}

func (k *ensureComponent) WithInfinispan() EnsureComponent {
	k.infinispan = installComponentState
	return k
}

func (k *ensureComponent) WithKafka() EnsureComponent {
	k.kafka = installComponentState
	return k
}

func (k *ensureComponent) WithKeycloak() EnsureComponent {
	k.keycloak = installComponentState
	return k
}

func (k *ensureComponent) WithoutInfinispan() EnsureComponent {
	k.infinispan = removeComponentState
	return k
}

func (k *ensureComponent) WithoutKafka() EnsureComponent {
	k.kafka = removeComponentState
	return k
}

func (k *ensureComponent) WithoutKeycloak() EnsureComponent {
	k.keycloak = removeComponentState
	return k
}

func (k *ensureComponent) Apply() (*v1alpha1.KogitoInfra, bool, error) {
	// Create or update kogitoinfra
	infra, err := k.createOrUpdateInfra()
	if err != nil {
		return nil, false, err
	}

	// Check if ready
	ready := true
	if deployed := isKafkaDeployed(infra); k.kafka == installComponentState && !deployed {
		ready = false
	} else if k.kafka == removeComponentState && deployed {
		ready = false
	}
	if deployed := isInfinispanDeployed(infra); k.infinispan == installComponentState && !deployed {
		ready = false
	} else if k.infinispan == removeComponentState && deployed {
		ready = false
	}
	if deployed := isKeycloakDeployed(infra); k.keycloak == installComponentState && !deployed {
		ready = false
	} else if k.keycloak == removeComponentState && deployed {
		ready = false
	}

	return infra, ready, nil
}

// createOrUpdateInfra will fetch for any reference of KogitoInfra in the given namespace.
// If not exists, a new one with Infinispan enabled will be created and returned
func (k *ensureComponent) createOrUpdateInfra() (*v1alpha1.KogitoInfra, error) {
	log := logger.GetLogger("infrastructure_kogitoinfra")
	log.Debug("Fetching for KogitoInfra list in namespace")
	// let's look for the deployed infra
	infras := &v1alpha1.KogitoInfraList{}
	if err := kubernetes.ResourceC(k.client).ListWithNamespace(k.namespace, infras); err != nil {
		return nil, err
	}
	log.Debugf("Found KogitoInfras: %s", infras.Items)
	var infra *v1alpha1.KogitoInfra
	// let's use the one we've found
	if len(infras.Items) > 0 {
		log.Debugf("Using and updating KogitoInfra: %s", &infras.Items[0])
		infra = &infras.Items[0]

		// Update only if any change
		infraChanged := false
		infraChanged = updateInstallValueIfNeeded(k.infinispan, &infra.Spec.InstallInfinispan) || infraChanged
		infraChanged = updateInstallValueIfNeeded(k.kafka, &infra.Spec.InstallKafka) || infraChanged
		infraChanged = updateInstallValueIfNeeded(k.keycloak, &infra.Spec.InstallKeycloak) || infraChanged
		if infraChanged {
			if err := kubernetes.ResourceC(k.client).Update(infra); err != nil {
				return nil, err
			}
		}
	} else {
		// found nothing, creating
		infra = &v1alpha1.KogitoInfra{
			ObjectMeta: metav1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: k.namespace},
			Spec: v1alpha1.KogitoInfraSpec{
				InstallInfinispan: getInstallValue(k.infinispan),
				InstallKafka:      getInstallValue(k.kafka),
				InstallKeycloak:   getInstallValue(k.keycloak),
			},
		}
		log.Debug("We don't have KogitoInfra deployed, trying to create a new one")
		if err := kubernetes.ResourceC(k.client).Create(infra); err != nil {
			return nil, err
		}
	}
	return infra, nil
}

// isInfraComponentDeployed verifies if the given component is available in the infrastructure
func isInfraComponentDeployed(status *v1alpha1.InfraComponentInstallStatusType) bool {
	if &status != nil &&
		len(status.Condition) > 0 {
		return len(status.Service) > 0 &&
			status.Condition[len(status.Condition)-1].Type == v1alpha1.SuccessInstallConditionType
	}
	return false
}

func isInfinispanDeployed(infra *v1alpha1.KogitoInfra) bool {
	if &infra.Status != nil && &infra.Status.Infinispan != nil {
		return isInfraComponentDeployed(&infra.Status.Infinispan.InfraComponentInstallStatusType)
	}
	return false
}

func isKafkaDeployed(infra *v1alpha1.KogitoInfra) bool {
	if &infra.Status != nil && &infra.Status.Kafka != nil {
		return isInfraComponentDeployed(&infra.Status.Kafka)
	}
	return false
}

func isKeycloakDeployed(infra *v1alpha1.KogitoInfra) bool {
	if &infra.Status != nil && &infra.Status.Keycloak != nil {
		return isInfraComponentDeployed(&infra.Status.Keycloak)
	}
	return false
}

// EnsureKogitoInfra will create the KogitoInfra instance with default values if does not exist and return the handle to specify which component should be created
func EnsureKogitoInfra(namespace string, cli *client.Client) EnsureComponent {
	ensure := &ensureComponent{
		namespace: namespace,
		client:    cli,

		infinispan: noActionComponentState,
		kafka:      noActionComponentState,
		keycloak:   noActionComponentState,
	}
	return ensure
}

func updateInstallValueIfNeeded(state componentState, installValueToChange *bool) (changed bool) {
	if state != noActionComponentState {
		newValue := getInstallValue(state)
		if *installValueToChange != newValue {
			*installValueToChange = newValue
			return true
		}
	}
	return false
}

func getInstallValue(state componentState) bool {
	return state == installComponentState
}
