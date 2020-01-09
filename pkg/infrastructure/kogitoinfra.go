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

	kafkaComponentType      componentType = "kafka"
	infinispanComponentType componentType = "infinispan"
)

type componentType string

type ensureComponent struct {
	infra   *v1alpha1.KogitoInfra
	created bool
	err     error
	client  *client.Client
}

// EnsureComponent interface to control how to provision an infra with a given component
type EnsureComponent interface {
	// WithInfinispan creates a new instance of KogitoInfra if not exists with an Infinispan deployed if not exists. If exists, checks if Infinispan is deployed.
	WithInfinispan() (infra *v1alpha1.KogitoInfra, created, ready bool, err error)
	// WithKafka creates a new instance of KogitoInfra if not exists with a Kafka deployed if not exists. If exists, checks if Kafka is deployed.
	WithKafka() (infra *v1alpha1.KogitoInfra, created, ready bool, err error)
}

func (k *ensureComponent) WithInfinispan() (infra *v1alpha1.KogitoInfra, created, ready bool, err error) {
	return k.withComponent(infinispanComponentType)
}

func (k *ensureComponent) WithKafka() (infra *v1alpha1.KogitoInfra, created, ready bool, err error) {
	return k.withComponent(kafkaComponentType)
}

func (k *ensureComponent) withComponent(component componentType) (*v1alpha1.KogitoInfra, bool, bool, error) {
	if k.err != nil {
		return k.infra, k.created, false, k.err
	}
	if k.created {
		return k.infra, k.created, false, nil
	}
	ready := false
	switch component {
	case kafkaComponentType:
		if !k.infra.Spec.InstallKafka {
			k.infra.Spec.InstallKafka = true
			k.err = kubernetes.ResourceC(k.client).Update(k.infra)
			return k.infra, k.created, false, nil
		}
		ready = isKafkaDeployed(k.infra)
	case infinispanComponentType:
		if !k.infra.Spec.InstallInfinispan {
			k.infra.Spec.InstallInfinispan = true
			k.err = kubernetes.ResourceC(k.client).Update(k.infra)
			return k.infra, k.created, false, nil
		}
		ready = isInfinispanDeployed(k.infra)
	}

	return k.infra, k.created, ready, nil
}

// createOrFetchInfra will fetch for any reference of KogitoInfra in the given namespace.
// If not exists, a new one with Infinispan enabled will be created and returned
func (k *ensureComponent) createOrFetchInfra(namespace string) {
	log := logger.GetLogger("infrastructure_kogitoinfra")
	log.Debug("Fetching for KogitoInfra list in namespace")
	// let's look for the deployed infra
	infras := &v1alpha1.KogitoInfraList{}
	k.created = false
	k.infra = nil
	if k.err = kubernetes.ResourceC(k.client).ListWithNamespace(namespace, infras); k.err != nil {
		return
	}
	log.Debugf("Found KogitoInfras: %s", infras.Items)
	// let's use the one we've found
	if len(infras.Items) > 0 {
		log.Debugf("Using KogitoInfra: %s", &infras.Items[0])
		k.created = false
		k.infra = &infras.Items[0]
		return
	}
	// found nothing, creating
	k.infra = &v1alpha1.KogitoInfra{
		ObjectMeta: metav1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: namespace},
		Spec:       v1alpha1.KogitoInfraSpec{InstallInfinispan: false, InstallKafka: false},
	}
	log.Debug("We don't have KogitoInfra deployed, trying to create a new one")
	if k.err = kubernetes.ResourceC(k.client).Create(k.infra); k.err != nil {
		k.created = false
		k.infra = nil
		return
	}
	k.created = true
	return
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

// EnsureKogitoInfra will create the KogitoInfra instance with default values if does not exist and return the handle to specify which component should be created
func EnsureKogitoInfra(namespace string, cli *client.Client) EnsureComponent {
	ensure := &ensureComponent{
		infra:   nil,
		created: false,
		err:     nil,
		client:  cli,
	}
	ensure.createOrFetchInfra(namespace)
	return ensure
}
