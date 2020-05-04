// Copyright 2020 Red Hat, Inc. and/or its affiliates
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

package services

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

func GetRequest(namespace string) reconcile.Request {
	return reconcile.Request{NamespacedName: types.NamespacedName{Namespace: namespace}}
}

func Test_serviceDeployer_Deploy(t *testing.T) {
	replicas := int32(1)
	service := &v1alpha1.KogitoJobsService{
		ObjectMeta: v1.ObjectMeta{
			Name:      "jobs-service",
			Namespace: t.Name(),
		},
		Spec: v1alpha1.KogitoJobsServiceSpec{
			InfinispanMeta:    v1alpha1.InfinispanMeta{InfinispanProperties: v1alpha1.InfinispanConnectionProperties{UseKogitoInfra: true}},
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{Replicas: &replicas},
		},
	}
	serviceList := &v1alpha1.KogitoJobsServiceList{}
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{service}, nil, nil)
	definition := ServiceDefinition{
		DefaultImageName: infrastructure.DefaultJobsServiceImageName,
		Request:          GetRequest(t.Name()),
	}
	deployer := NewSingletonServiceDeployer(definition, serviceList, cli, meta.GetRegisteredSchema())
	reconcileAfter, err := deployer.Deploy()
	assert.NoError(t, err)
	assert.True(t, reconcileAfter > 0) // we just deployed Infinispan

	exists, err := kubernetes.ResourceC(cli).Fetch(service)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, 1, len(service.Status.Conditions))
	assert.Equal(t, int32(1), *service.Spec.Replicas)
	assert.Equal(t, v1alpha1.ProvisioningConditionType, service.Status.Conditions[0].Type)
}

func Test_serviceDeployer_deployInfinispan_dataIndex(t *testing.T) {
	replicas := int32(1)
	dataIndex := &v1alpha1.KogitoDataIndex{
		ObjectMeta: v1.ObjectMeta{Name: "data-index", Namespace: t.Name()},
		Spec: v1alpha1.KogitoDataIndexSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{Replicas: &replicas},
		},
	}
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{dataIndex}, nil, nil)
	deployer := serviceDeployer{
		definition: ServiceDefinition{
			Request:             GetRequest(dataIndex.Namespace),
			RequiresPersistence: true,
		},
		client: cli,
		scheme: meta.GetRegisteredSchema(),
	}
	requeueAfter, err := deployer.deployInfinispan(dataIndex)
	assert.NoError(t, err)
	assert.True(t, requeueAfter > 0, "Should have deployed Infinispan for us since the service requires persistence and is Infinispan aware")

	kogitoInfra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Namespace: t.Name(), Name: infrastructure.DefaultKogitoInfraName},
	}
	exists, err := kubernetes.ResourceC(cli).Fetch(kogitoInfra)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.True(t, kogitoInfra.Spec.InstallInfinispan)
}

func Test_serviceDeployer_deployInfinispan_dataIndexProvidedInfinispan(t *testing.T) {
	replicas := int32(1)
	dataIndex := &v1alpha1.KogitoDataIndex{
		ObjectMeta: v1.ObjectMeta{Name: "data-index", Namespace: t.Name()},
		Spec: v1alpha1.KogitoDataIndexSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{Replicas: &replicas},
			InfinispanMeta:    v1alpha1.InfinispanMeta{InfinispanProperties: v1alpha1.InfinispanConnectionProperties{URI: "my-infinispan-instance:5000"}},
		},
	}
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{dataIndex}, nil, nil)
	deployer := serviceDeployer{
		definition: ServiceDefinition{
			Request:             GetRequest(dataIndex.Namespace),
			RequiresPersistence: true,
		},
		client: cli,
		scheme: meta.GetRegisteredSchema(),
	}
	requeueAfter, err := deployer.deployInfinispan(dataIndex)
	assert.NoError(t, err)
	assert.True(t, requeueAfter == 0, "Should NOT have deployed Infinispan for us since the service requires persistence, but the user just pointed the URI")

	kogitoInfra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Namespace: t.Name(), Name: infrastructure.DefaultKogitoInfraName},
	}
	exists, err := kubernetes.ResourceC(cli).Fetch(kogitoInfra)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func Test_serviceDeployer_deployInfinispan_jobsService(t *testing.T) {
	replicas := int32(1)
	jobsService := &v1alpha1.KogitoJobsService{
		ObjectMeta: v1.ObjectMeta{Name: "jobs-service", Namespace: t.Name()},
		Spec: v1alpha1.KogitoJobsServiceSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{Replicas: &replicas},
		},
	}
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{jobsService}, nil, nil)
	deployer := serviceDeployer{
		definition: ServiceDefinition{
			Request:             GetRequest(jobsService.Namespace),
			RequiresPersistence: false,
		},
		client: cli,
		scheme: meta.GetRegisteredSchema(),
	}
	requeueAfter, err := deployer.deployInfinispan(jobsService)
	assert.NoError(t, err)
	assert.True(t, requeueAfter == 0, "Should NOT have deployed Infinispan for us since the service DOES NOT require persistence and is Infinispan aware")

	kogitoInfra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Namespace: t.Name(), Name: infrastructure.DefaultKogitoInfraName},
	}
	exists, err := kubernetes.ResourceC(cli).Fetch(kogitoInfra)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func Test_serviceDeployer_deployInfinispan_jobsServiceWithPersistence(t *testing.T) {
	replicas := int32(1)
	jobsService := &v1alpha1.KogitoJobsService{
		ObjectMeta: v1.ObjectMeta{Name: "jobs-service", Namespace: t.Name()},
		Spec: v1alpha1.KogitoJobsServiceSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{Replicas: &replicas},
			InfinispanMeta:    v1alpha1.InfinispanMeta{InfinispanProperties: v1alpha1.InfinispanConnectionProperties{UseKogitoInfra: true}},
		},
	}
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{jobsService}, nil, nil)
	deployer := serviceDeployer{
		definition: ServiceDefinition{
			Request: GetRequest(jobsService.Namespace),
			// does not require persistence, but if the user set in the CR, will deploy Infinispan anyway
			RequiresPersistence: false,
		},
		client: cli,
		scheme: meta.GetRegisteredSchema(),
	}
	requeueAfter, err := deployer.deployInfinispan(jobsService)
	assert.NoError(t, err)
	assert.True(t, requeueAfter > 0, "Should have deployed Infinispan for us since the service DOES NOT require persistence, is Infinispan aware and user sets to use Kogito Infra")

	kogitoInfra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Namespace: t.Name(), Name: infrastructure.DefaultKogitoInfraName},
	}
	exists, err := kubernetes.ResourceC(cli).Fetch(kogitoInfra)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.True(t, kogitoInfra.Spec.InstallInfinispan)
}

func Test_serviceDeployer_deployKafka_dataIndex(t *testing.T) {
	replicas := int32(1)
	dataIndex := &v1alpha1.KogitoDataIndex{
		ObjectMeta: v1.ObjectMeta{Name: "data-index", Namespace: t.Name()},
		Spec: v1alpha1.KogitoDataIndexSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{Replicas: &replicas},
		},
	}
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{dataIndex}, nil, nil)
	deployer := serviceDeployer{
		definition: ServiceDefinition{
			Request:           GetRequest(dataIndex.Namespace),
			RequiresMessaging: true,
		},
		client: cli,
		scheme: meta.GetRegisteredSchema(),
	}
	requeueAfter, err := deployer.deployKafka(dataIndex)
	assert.NoError(t, err)
	assert.True(t, requeueAfter > 0, "Should have deployed Kafka for us since the service requires messaging and it is Kafka aware")

	kogitoInfra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Namespace: t.Name(), Name: infrastructure.DefaultKogitoInfraName},
	}
	exists, err := kubernetes.ResourceC(cli).Fetch(kogitoInfra)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.True(t, kogitoInfra.Spec.InstallKafka)
}

func Test_serviceDeployer_deployKafka_dataIndexProvidedKafkaExternalURI(t *testing.T) {
	replicas := int32(1)
	dataIndex := &v1alpha1.KogitoDataIndex{
		ObjectMeta: v1.ObjectMeta{Name: "data-index", Namespace: t.Name()},
		Spec: v1alpha1.KogitoDataIndexSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{Replicas: &replicas},
			KafkaMeta:         v1alpha1.KafkaMeta{KafkaProperties: v1alpha1.KafkaConnectionProperties{ExternalURI: "my-kafka:9092"}},
		},
	}
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{dataIndex}, nil, nil)
	deployer := serviceDeployer{
		definition: ServiceDefinition{
			Request:           GetRequest(dataIndex.Namespace),
			RequiresMessaging: true,
		},
		client: cli,
		scheme: meta.GetRegisteredSchema(),
	}
	requeueAfter, err := deployer.deployKafka(dataIndex)
	assert.NoError(t, err)
	assert.True(t, requeueAfter == 0, "Should NOT have deployed Kafka for us since the service requires messaging, but the user just pointed the URI")

	kogitoInfra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Namespace: t.Name(), Name: infrastructure.DefaultKogitoInfraName},
	}
	exists, err := kubernetes.ResourceC(cli).Fetch(kogitoInfra)
	assert.NoError(t, err)
	assert.False(t, exists)
}
func Test_serviceDeployer_deployKafka_dataIndexProvidedKafkaInstance(t *testing.T) {
	replicas := int32(1)
	dataIndex := &v1alpha1.KogitoDataIndex{
		ObjectMeta: v1.ObjectMeta{Name: "data-index", Namespace: t.Name()},
		Spec: v1alpha1.KogitoDataIndexSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{Replicas: &replicas},
			KafkaMeta:         v1alpha1.KafkaMeta{KafkaProperties: v1alpha1.KafkaConnectionProperties{Instance: "my-external-kafka"}},
		},
	}
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{dataIndex}, nil, nil)
	deployer := serviceDeployer{
		definition: ServiceDefinition{
			Request:           GetRequest(dataIndex.Namespace),
			RequiresMessaging: true,
		},
		client: cli,
		scheme: meta.GetRegisteredSchema(),
	}
	requeueAfter, err := deployer.deployKafka(dataIndex)
	assert.NoError(t, err)
	assert.True(t, requeueAfter == 0, "Should NOT have deployed Kafka for us since the service requires messaging, but the user just gave the kafka instance")

	kogitoInfra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Namespace: t.Name(), Name: infrastructure.DefaultKogitoInfraName},
	}
	exists, err := kubernetes.ResourceC(cli).Fetch(kogitoInfra)
	assert.NoError(t, err)
	assert.False(t, exists)
}