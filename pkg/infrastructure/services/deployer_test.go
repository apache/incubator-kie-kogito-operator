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
	corev1 "k8s.io/api/core/v1"
	"testing"
	"time"

	kafkav1beta1 "github.com/kiegroup/kogito-cloud-operator/api/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func newReconcileRequest(namespace string) reconcile.Request {
	return reconcile.Request{NamespacedName: types.NamespacedName{Namespace: namespace}}
}

func Test_serviceDeployer_DataIndex_InfraNotReady(t *testing.T) {
	replicas := int32(1)
	infraKafka := newSuccessfulInfinispanInfra(t.Name())
	infraInfinispan := newSuccessfulKafkaInfra(t.Name())
	dataIndex := &v1beta1.KogitoSupportingService{
		ObjectMeta: v1.ObjectMeta{
			Name:      "data-index",
			Namespace: t.Name(),
		},
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType: v1beta1.DataIndex,
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{
				Replicas: &replicas,
				Infra: []string{
					infraKafka.Name, infraInfinispan.Name,
				},
			},
		},
	}
	cli := test.NewFakeClientBuilder().AddK8sObjects(dataIndex).Build()
	definition := ServiceDefinition{
		DefaultImageName: infrastructure.DefaultDataIndexImageName,
		Request:          newReconcileRequest(t.Name()),
		KafkaTopics:      []string{"mytopic"},
	}
	deployer := NewServiceDeployer(definition, dataIndex, cli, meta.GetRegisteredSchema())
	reconcileAfter, err := deployer.Deploy()
	assert.Error(t, err)
	assert.Equal(t, time.Duration(0), reconcileAfter)

	test.AssertFetchMustExist(t, cli, dataIndex)
	assert.NotNil(t, dataIndex.Status)
	assert.Len(t, dataIndex.Status.Conditions, 1)
	assert.Equal(t, dataIndex.Status.Conditions[0].Reason, v1beta1.ServiceReconciliationFailure)

	// Infinispan is not ready :)
	infraInfinispan.Status.Condition.Message = "Headaches"
	infraInfinispan.Status.Condition.Status = corev1.ConditionFalse
	infraInfinispan.Status.Condition.Reason = v1beta1.ResourceNotReady
	infraInfinispan.Status.Condition.Type = v1beta1.FailureInfraConditionType

	test.AssertCreate(t, cli, infraInfinispan)
	test.AssertCreate(t, cli, infraKafka)

	reconcileAfter, err = deployer.Deploy()
	assert.NoError(t, err)
	assert.Equal(t, reconcileAfter, reconciliationIntervalAfterInfraError)
	test.AssertFetchMustExist(t, cli, dataIndex)
	assert.NotNil(t, dataIndex.Status)
	assert.Len(t, dataIndex.Status.Conditions, 2)
	for _, condition := range dataIndex.Status.Conditions {
		assert.Equal(t, condition.Type, v1beta1.FailedConditionType)
		assert.Equal(t, condition.Status, corev1.ConditionFalse)
	}
}

func Test_serviceDeployer_DataIndex(t *testing.T) {
	replicas := int32(1)
	requiredTopic := "dataindex-required-topic"
	infraKafka := newSuccessfulInfinispanInfra(t.Name())
	infraInfinispan := newSuccessfulKafkaInfra(t.Name())
	dataIndex := &v1beta1.KogitoSupportingService{
		ObjectMeta: v1.ObjectMeta{
			Name:      "data-index",
			Namespace: t.Name(),
		},
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType: v1beta1.DataIndex,
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{
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
	deployer := NewServiceDeployer(definition, dataIndex, cli, meta.GetRegisteredSchema())
	reconcileAfter, err := deployer.Deploy()
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), reconcileAfter)

	topic := &kafkav1beta1.KafkaTopic{
		ObjectMeta: v1.ObjectMeta{
			Name:      requiredTopic,
			Namespace: t.Name(),
		},
	}
	test.AssertFetchMustExist(t, cli, topic)
}

func Test_serviceDeployer_Deploy(t *testing.T) {
	replicas := int32(1)
	service := &v1beta1.KogitoSupportingService{
		ObjectMeta: v1.ObjectMeta{
			Name:      "jobs-service",
			Namespace: t.Name(),
		},
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType:       v1beta1.JobsService,
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{Replicas: &replicas},
		},
	}
	cli := test.NewFakeClientBuilder().AddK8sObjects(service).OnOpenShift().Build()
	definition := ServiceDefinition{
		DefaultImageName: infrastructure.DefaultJobsServiceImageName,
		Request:          newReconcileRequest(t.Name()),
	}
	deployer := NewServiceDeployer(definition, service, cli, meta.GetRegisteredSchema())
	requeueAfter, err := deployer.Deploy()
	assert.NoError(t, err)
	assert.True(t, requeueAfter == 0)

	exists, err := kubernetes.ResourceC(cli).Fetch(service)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, 1, len(service.Status.Conditions))
	assert.Equal(t, int32(1), *service.Spec.Replicas)
	assert.Equal(t, v1beta1.ProvisioningConditionType, service.Status.Conditions[0].Type)
}

func newSuccessfulKafkaInfra(namespace string) *v1beta1.KogitoInfra {
	return &v1beta1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: "kafka-infra", Namespace: namespace},
		Spec: v1beta1.KogitoInfraSpec{
			Resource: v1beta1.Resource{
				APIVersion: infrastructure.KafkaAPIVersion,
				Kind:       infrastructure.KafkaKind,
				Namespace:  namespace,
				Name:       "kogito-kafka",
			},
		},
		Status: v1beta1.KogitoInfraStatus{
			Condition: v1beta1.KogitoInfraCondition{
				Type:   v1beta1.SuccessInfraConditionType,
				Status: v1.StatusSuccess,
				Reason: "",
			},
			RuntimeProperties: map[v1beta1.RuntimeType]v1beta1.RuntimeProperties{
				v1beta1.QuarkusRuntimeType: {
					AppProps: map[string]string{QuarkusKafkaBootstrapAppProp: "kafka:1101"},
				},
			},
		},
	}
}

func newSuccessfulInfinispanInfra(namespace string) *v1beta1.KogitoInfra {
	return &v1beta1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: "infinispan-infra", Namespace: namespace},
		Spec: v1beta1.KogitoInfraSpec{
			Resource: v1beta1.Resource{
				APIVersion: infrastructure.InfinispanAPIVersion,
				Kind:       infrastructure.InfinispanKind,
				Namespace:  namespace,
				Name:       "kogito-infinispan",
			},
		},
		Status: v1beta1.KogitoInfraStatus{
			Condition: v1beta1.KogitoInfraCondition{
				Type:   v1beta1.SuccessInfraConditionType,
				Status: v1.StatusSuccess,
				Reason: "",
			},
		},
	}
}
