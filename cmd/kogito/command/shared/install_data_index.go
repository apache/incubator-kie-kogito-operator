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

package shared

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/message"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	dataindex "github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitodataindex/resource"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultDataIndexReplicas = 1
)

// installDefaultDataIndex shortcut to install the default Data Index in the given namespace.
// Useful for quick starts, since the Data Index can be updated later to fulfill users needs.
func installDefaultDataIndex(cli *client.Client, namespace string) error {
	log := context.GetDefaultLogger()

	// TODO: add Kafka to be auto provisioned when implementing KOGITO-614, then we can delete this kogitoInfra instance
	kogitoInfra := v1alpha1.KogitoInfra{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: infrastructure.DefaultKogitoInfraName},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallInfinispan: true,
			InstallKafka:      true,
		},
	}
	if err := kubernetes.ResourceC(cli).Create(&kogitoInfra); err != nil {
		return fmt.Errorf(message.KogitoInfraErrCreating, err)
	}

	kogitoDataIndex := v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{Name: dataindex.DefaultDataIndexName, Namespace: namespace},
		Spec: v1alpha1.KogitoDataIndexSpec{
			Replicas:       defaultDataIndexReplicas,
			Image:          dataindex.DefaultImage,
			InfinispanMeta: v1alpha1.InfinispanMeta{InfinispanProperties: v1alpha1.InfinispanConnectionProperties{UseKogitoInfra: true}},
		},
		Status: v1alpha1.KogitoDataIndexStatus{
			Conditions:         []v1alpha1.DataIndexCondition{},
			DependenciesStatus: []v1alpha1.DataIndexDependenciesStatus{},
		},
	}

	if err := kubernetes.ResourceC(cli).Create(&kogitoDataIndex); err != nil {
		return fmt.Errorf(message.DataIndexErrCreating, err)
	}

	log.Infof(message.DataIndexSuccessfulInstalled, namespace)
	log.Infof(message.DataIndexCheckStatus, kogitoDataIndex.Name, namespace)

	return nil
}
