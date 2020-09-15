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

package kogitoinfra

import (
	v12 "github.com/infinispan/infinispan-operator/pkg/apis/infinispan/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	v13 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
)

func Test_Reconcile_KafkaResource(t *testing.T) {
	kogitoInfra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: "kogito-kafka", Namespace: t.Name()},
		Spec: v1alpha1.KogitoInfraSpec{
			Resource: v1alpha1.Resource{
				APIVersion: "kafka.strimzi.io/v1beta1",
				Kind:       "Kafka",
				Name:       "kogito-kafka",
				Namespace:  t.Name(),
			},
		},
	}

	deployedKafkaInstance := &kafkabetav1.Kafka{
		ObjectMeta: v1.ObjectMeta{Name: "kogito-kafka", Namespace: t.Name()},
		Status: kafkabetav1.KafkaStatus{
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

	client := test.CreateFakeClient([]runtime.Object{
		kogitoInfra,
		deployedKafkaInstance,
	}, nil, nil)

	scheme := meta.GetRegisteredSchema()
	r := &ReconcileKogitoInfra{client: client, scheme: scheme}
	// basic checks
	test.AssertReconcile(t, r, kogitoInfra)

	exists, err := kubernetes.ResourceC(client).Fetch(kogitoInfra)
	assert.NoError(t, err)
	assert.True(t, exists)
	kafkaProperties := kogitoInfra.Status.KafkaProperties
	assert.Equal(t, "kogito-kafka:9090", kafkaProperties.ExternalURI)
}

func Test_Reconcile_Infinispan(t *testing.T) {

	kogitoInfra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: "kogito-infinispan", Namespace: t.Name()},
		Spec: v1alpha1.KogitoInfraSpec{
			Resource: v1alpha1.Resource{
				APIVersion: "infinispan.org/v1",
				Kind:       "Infinispan",
				Name:       "kogito-infinispan",
				Namespace:  t.Name(),
			},
		},
	}

	deployedInfinispan := &v12.Infinispan{
		ObjectMeta: v1.ObjectMeta{Name: "kogito-infinispan", Namespace: t.Name()},
	}

	deployedCustomSecret := &v13.Secret{
		ObjectMeta: v1.ObjectMeta{Name: "kogito-infinispan-credential", Namespace: t.Name()},
	}

	infinispanService := &v13.Service{
		ObjectMeta: v1.ObjectMeta{Name: "kogito-infinispan", Namespace: t.Name()},
		Spec: v13.ServiceSpec{
			Ports: []v13.ServicePort{
				{
					TargetPort: intstr.IntOrString{IntVal: 11222},
				},
			},
		},
	}

	client := test.CreateFakeClient([]runtime.Object{
		kogitoInfra,
		deployedInfinispan,
		deployedCustomSecret,
		infinispanService,
	}, nil, nil)

	scheme := meta.GetRegisteredSchema()
	r := &ReconcileKogitoInfra{client: client, scheme: scheme}
	// basic checks
	test.AssertReconcile(t, r, kogitoInfra)

	exists, err := kubernetes.ResourceC(client).Fetch(kogitoInfra)
	assert.NoError(t, err)
	assert.True(t, exists)
	infinispanProperties := kogitoInfra.Status.InfinispanProperties
	assert.Equal(t, "kogito-infinispan:11222", infinispanProperties.URI)
	assert.Equal(t, "kogito-infinispan-credential", infinispanProperties.Credentials.SecretName)
	assert.Equal(t, infrastructure.InfinispanSecretUsernameKey, infinispanProperties.Credentials.UsernameKey)
	assert.Equal(t, infrastructure.InfinispanSecretPasswordKey, infinispanProperties.Credentials.PasswordKey)
}
