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

package controllers

import (
	"github.com/kiegroup/kogito-operator/api"
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	"github.com/kiegroup/kogito-operator/core/logger"
	"github.com/kiegroup/kogito-operator/core/test"
	"github.com/kiegroup/kogito-operator/meta"
	meta2 "k8s.io/apimachinery/pkg/api/meta"
	"testing"

	"github.com/kiegroup/kogito-operator/api/v1beta1"
	"github.com/kiegroup/kogito-operator/core/client/kubernetes"
	kafkabetav1 "github.com/kiegroup/kogito-operator/core/infrastructure/kafka/v1beta1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_Reconcile_ResourceNotFound(t *testing.T) {
	kogitoInfra := &v1beta1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: "kogito-infinispan", Namespace: t.Name()},
		Spec: v1beta1.KogitoInfraSpec{
			Resource: v1beta1.Resource{
				APIVersion: infrastructure.InfinispanAPIVersion,
				Kind:       infrastructure.InfinispanKind,
				Name:       "kogito-infinispan",
				Namespace:  t.Name(),
			},
		},
	}
	client := test.NewFakeClientBuilder().AddK8sObjects(kogitoInfra).Build()
	scheme := meta.GetRegisteredSchema()
	r := &KogitoInfraReconciler{Client: client, Scheme: scheme, Log: test.TestLogger}
	// basic checks
	test.AssertReconcileMustNotRequeue(t, r, kogitoInfra)
	exists, err := kubernetes.ResourceC(client).Fetch(kogitoInfra)
	assert.True(t, exists)
	assert.NoError(t, err)
	failureCondition := meta2.FindStatusCondition(*kogitoInfra.Status.Conditions, string(api.KogitoInfraConfigured))
	assert.NotNil(t, failureCondition)
	assert.NotEmpty(t, failureCondition.Message)
	// we haven't created the Infinispan server and we are informing our KogitoInfra instance that it will require it :)
	assert.Equal(t, string(api.ResourceNotFound), failureCondition.Reason)
}

func Test_Reconcile_KafkaResource(t *testing.T) {
	kogitoInfra := &v1beta1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: "kogito-kafka", Namespace: t.Name()},
		Spec: v1beta1.KogitoInfraSpec{
			Resource: v1beta1.Resource{
				APIVersion: infrastructure.KafkaAPIVersion,
				Kind:       infrastructure.KafkaKind,
				Name:       "kogito-kafka",
				Namespace:  t.Name(),
			},
		},
	}

	deployedKafkaInstance := &kafkabetav1.Kafka{
		ObjectMeta: v1.ObjectMeta{Name: "kogito-kafka", Namespace: t.Name()},
		Status: kafkabetav1.KafkaStatus{
			Conditions: []kafkabetav1.KafkaCondition{
				{Type: kafkabetav1.KafkaConditionTypeReady},
			},
			Listeners: []kafkabetav1.ListenerStatus{
				{
					Type: "plain",
					Addresses: []kafkabetav1.ListenerAddress{
						{
							Host: "kogito-kafka",
							Port: 9090,
						},
					},
				},
			},
		},
	}

	client := test.NewFakeClientBuilder().AddK8sObjects(kogitoInfra, deployedKafkaInstance).Build()

	scheme := meta.GetRegisteredSchema()
	r := &KogitoInfraReconciler{Client: client, Scheme: scheme, Log: logger.GetLogger("kogito infra reconciler")}
	// basic checks
	test.AssertReconcile(t, r, kogitoInfra)

	exists, err := kubernetes.ResourceC(client).Fetch(kogitoInfra)
	assert.NoError(t, err)
	assert.True(t, exists)
	kafkaQuarkusAppProps := kogitoInfra.Status.RuntimeProperties[api.QuarkusRuntimeType].AppProps
	assert.Contains(t, "kogito-kafka:9090", kafkaQuarkusAppProps["kafka.bootstrap.servers"])
	kafkaSpringBootAppProps := kogitoInfra.Status.RuntimeProperties[api.SpringBootRuntimeType].AppProps
	assert.Contains(t, "kogito-kafka:9090", kafkaSpringBootAppProps["kafka.bootstrap.servers"])
}
