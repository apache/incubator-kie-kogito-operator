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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// DefaultKogitoInfraName is the default name given to the Kogito Infra.
	// We're not attached to this name, but since we're going to create it automagically, it's better to have a standard one.
	DefaultKogitoInfraName = "kogito-infra"
)

// EnsureInfinispanWithKogitoInfra creates a new instance of KogitoInfra if not exists with an Infinispan deployed if not exists. If exists, checks if Infinispan is deployed.
func EnsureInfinispanWithKogitoInfra(namespace string, cli *client.Client) (infra *v1alpha1.KogitoInfra, created, ready bool, err error) {
	ready = false
	infra, created, err = createOrFetchInfra(namespace, cli)
	if err != nil {
		return
	}
	if created {
		// since we just created a new Infra instance, let's wait for it to provision everything before proceeding
		return
	}
	if !infra.Spec.InstallInfinispan {
		infra.Spec.InstallInfinispan = true
		err = kubernetes.ResourceC(cli).Update(infra)
		ready = false
		return
	}
	ready = isInfinispanDeployed(infra)
	return
}

// createOrFetchInfra will fetch for any reference of KogitoInfra in the given namespace.
// If not exists, a new one with Infinispan enabled will be created and returned
func createOrFetchInfra(namespace string, cli *client.Client) (infra *v1alpha1.KogitoInfra, created bool, err error) {
	log := logger.GetLogger("infrastructure_kogitoinfra")
	log.Debug("Fetching for KogitoInfra list in namespace")
	// let's look for the deployed infra
	infras := &v1alpha1.KogitoInfraList{}
	if err := kubernetes.ResourceC(cli).ListWithNamespace(namespace, infras); err != nil {
		return nil, false, err
	}
	log.Debugf("Found KogitoInfras: %s", infras.Items)
	// let's use the one we've found
	if len(infras.Items) > 0 {
		log.Debugf("Using KogitoInfra: %s", &infras.Items[0])
		return &infras.Items[0], false, nil
	}
	// found nothing, creating
	infra = &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: namespace},
		Spec:       v1alpha1.KogitoInfraSpec{InstallInfinispan: true},
	}
	log.Debug("We don't have KogitoInfra deployed, trying to create a new one")
	if err := kubernetes.ResourceC(cli).Create(infra); err != nil {
		return nil, false, err
	}
	return infra, true, nil
}

// isInfinispanDeployed will verify if the given KogitoInfra has Infinispan deployed
func isInfinispanDeployed(infra *v1alpha1.KogitoInfra) bool {
	if &infra.Status != nil &&
		&infra.Status.Infinispan != nil &&
		len(infra.Status.Infinispan.Condition) > 0 {

		return len(infra.Status.Infinispan.Service) > 0 &&
			infra.Status.Infinispan.Condition[len(infra.Status.Infinispan.Condition)-1].Type == v1alpha1.SuccessInstallConditionType
	}
	return false
}
