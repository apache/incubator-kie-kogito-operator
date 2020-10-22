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
	"testing"
	"time"

	grafanav1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func newReconcileRequest(namespace string) reconcile.Request {
	return reconcile.Request{NamespacedName: types.NamespacedName{Namespace: namespace}}
}

func Test_serviceDeployer_DataIndex(t *testing.T) {
	replicas := int32(1)
	requiredTopic := "dataindex-required-topic"
	infraKafka := createSuccessfulInfinispanInfra(t.Name())
	infraInfinispan := createSuccessfulKafkaInfra(t.Name())
	dataIndex := &v1alpha1.KogitoDataIndex{
		ObjectMeta: v1.ObjectMeta{
			Name:      "data-index",
			Namespace: t.Name(),
		},
		Spec: v1alpha1.KogitoDataIndexSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{
				Replicas: &replicas,
				Infra: []string{
					infraKafka.Name, infraInfinispan.Name,
				},
			},
		},
	}
	cli := test.NewFakeClientBuilder().AddK8sObjects(infraInfinispan, infraKafka, dataIndex).Build()
	definition := ServiceDefinition{
		DefaultImageName: infrastructure.DefaultDataIndexImageName,
		Request:          newReconcileRequest(t.Name()),
		KafkaTopics:      []string{requiredTopic},
	}
	deployer := NewSingletonServiceDeployer(definition, &v1alpha1.KogitoDataIndexList{}, cli, meta.GetRegisteredSchema())
	reconcileAfter, err := deployer.Deploy()
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), reconcileAfter)

	topic := &v1beta1.KafkaTopic{
		ObjectMeta: v1.ObjectMeta{
			Name:      requiredTopic,
			Namespace: t.Name(),
		},
	}
	test.AssertFetchMustExist(t, cli, topic)
}

func Test_serviceDeployer_Deploy(t *testing.T) {
	replicas := int32(1)
	service := &v1alpha1.KogitoJobsService{
		ObjectMeta: v1.ObjectMeta{
			Name:      "jobs-service",
			Namespace: t.Name(),
		},
		Spec: v1alpha1.KogitoJobsServiceSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{Replicas: &replicas},
		},
	}
	cli := test.NewFakeClientBuilder().AddK8sObjects(service).OnOpenShift().Build()
	definition := ServiceDefinition{
		DefaultImageName: infrastructure.DefaultJobsServiceImageName,
		Request:          newReconcileRequest(t.Name()),
	}
	deployer := NewSingletonServiceDeployer(definition, &v1alpha1.KogitoJobsServiceList{}, cli, meta.GetRegisteredSchema())
	reconcileAfter, err := deployer.Deploy()
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), reconcileAfter)

	exists, err := kubernetes.ResourceC(cli).Fetch(service)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, 1, len(service.Status.Conditions))
	assert.Equal(t, int32(1), *service.Spec.Replicas)
	assert.Equal(t, v1alpha1.ProvisioningConditionType, service.Status.Conditions[0].Type)
}

func createSuccessfulKafkaInfra(namespace string) *v1alpha1.KogitoInfra {
	return &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: "kafka-infra", Namespace: namespace},
		Spec: v1alpha1.KogitoInfraSpec{
			Resource: v1alpha1.Resource{
				APIVersion: infrastructure.KafkaAPIVersion,
				Kind:       infrastructure.KafkaKind,
				Namespace:  namespace,
				Name:       "kogito-kafka",
			},
		},
		Status: v1alpha1.KogitoInfraStatus{
			Condition: v1alpha1.KogitoInfraCondition{
				Type:   v1alpha1.SuccessInfraConditionType,
				Status: v1.StatusSuccess,
				Reason: "",
			},
			AppProps: map[string]string{QuarkusKafkaBootstrapAppProp: "kafka:1101"},
		},
	}
}

func createSuccessfulInfinispanInfra(namespace string) *v1alpha1.KogitoInfra {
	return &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: "infinispan-infra", Namespace: namespace},
		Spec: v1alpha1.KogitoInfraSpec{
			Resource: v1alpha1.Resource{
				APIVersion: infrastructure.InfinispanAPIVersion,
				Kind:       infrastructure.InfinispanKind,
				Namespace:  namespace,
				Name:       "kogito-infinispan",
			},
		},
		Status: v1alpha1.KogitoInfraStatus{
			Condition: v1alpha1.KogitoInfraCondition{
				Type:   v1alpha1.SuccessInfraConditionType,
				Status: v1.StatusSuccess,
				Reason: "",
			},
		},
	}
}

func Test_serviceDeployer_DeployGrafanaDashboards(t *testing.T) {
	replicas := int32(1)
	service := &v1alpha1.KogitoRuntime{
		ObjectMeta: v1.ObjectMeta{
			Name:      "my-kogito-runtime",
			Namespace: t.Name(),
		},
		Spec: v1alpha1.KogitoRuntimeSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{Replicas: &replicas},
		},
	}
	cli := test.NewFakeClientBuilder().AddK8sObjects(service).OnOpenShift().Build()

	dashboards := []GrafanaDashboard{
		{
			Name:             "mydashboard",
			RawJSONDashboard: "[]",
		},
		{
			Name:             "myseconddashboard",
			RawJSONDashboard: "[]",
		},
	}

	reconcileAfter, err := deployGrafanaDashboards(dashboards, cli, t.Name())
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), reconcileAfter)

	dashboard := &grafanav1.GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mydashboard",
			Namespace: t.Name(),
		},
	}
	test.AssertFetchMustExist(t, cli, dashboard)

	dashboard = &grafanav1.GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myseconddashboard",
			Namespace: t.Name(),
		},
	}
	test.AssertFetchMustExist(t, cli, dashboard)
}
