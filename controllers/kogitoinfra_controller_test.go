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
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	"io/ioutil"
	"testing"

	ispn "github.com/infinispan/infinispan-operator/pkg/apis/infinispan/v1"
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/api/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	r := &KogitoInfraReconciler{Client: client, Scheme: scheme, Log: logger.GetLogger("kogito infra reconciler")}
	// basic checks
	test.AssertReconcileMustNotRequeue(t, r, kogitoInfra)
	exists, err := kubernetes.ResourceC(client).Fetch(kogitoInfra)
	assert.True(t, exists)
	assert.NoError(t, err)
	assert.NotEmpty(t, kogitoInfra.Status.Condition.Message)
	// we haven't created the Infinispan server and we are informing our KogitoInfra instance that it will require it :)
	assert.Equal(t, v1beta1.ResourceNotFound, kogitoInfra.Status.Condition.Reason)
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
	kafkaQuarkusAppProps := kogitoInfra.Status.RuntimeProperties[v1beta1.QuarkusRuntimeType].AppProps
	assert.Contains(t, "kogito-kafka:9090", kafkaQuarkusAppProps["kafka.bootstrap.servers"])
	kafkaSpringBootAppProps := kogitoInfra.Status.RuntimeProperties[v1beta1.SpringBootRuntimeType].AppProps
	assert.Contains(t, "kogito-kafka:9090", kafkaSpringBootAppProps["kafka.bootstrap.servers"])
}

func Test_Reconcile_Infinispan(t *testing.T) {
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
	crtFile, err := ioutil.ReadFile("./testdata/tls.crt")
	assert.NoError(t, err)
	tlsSecret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      "secret-with-truststore",
			Namespace: t.Name(),
		},
		Data: map[string][]byte{truststoreSecretKey: crtFile},
	}
	deployedInfinispan := &ispn.Infinispan{
		ObjectMeta: v1.ObjectMeta{Name: "kogito-infinispan", Namespace: t.Name()},
		Status: ispn.InfinispanStatus{
			Security: ispn.InfinispanSecurity{
				EndpointEncryption: ispn.EndpointEncryption{
					CertSecretName: tlsSecret.Name,
				},
			},
			Conditions: []ispn.InfinispanCondition{
				{
					Status: string(v1.ConditionTrue),
				},
			},
		},
	}

	deployedCustomSecret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{Name: "kogito-infinispan-credential", Namespace: t.Name()},
	}

	infinispanService := &corev1.Service{
		ObjectMeta: v1.ObjectMeta{Name: "kogito-infinispan", Namespace: t.Name()},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					TargetPort: intstr.FromInt(11222),
				},
			},
		},
	}

	client := test.NewFakeClientBuilder().
		AddK8sObjects(kogitoInfra, deployedInfinispan, deployedCustomSecret, infinispanService, tlsSecret).
		Build()

	scheme := meta.GetRegisteredSchema()
	r := &KogitoInfraReconciler{Client: client, Scheme: scheme, Log: logger.GetLogger("kogito infra reconciler")}
	// basic checks
	test.AssertReconcile(t, r, kogitoInfra)

	exists, err := kubernetes.ResourceC(client).Fetch(kogitoInfra)
	assert.NoError(t, err)
	assert.True(t, exists)
	infinispanQuarkusAppProps := kogitoInfra.Status.RuntimeProperties[v1beta1.QuarkusRuntimeType].AppProps
	assert.Equal(t, "kogito-infinispan:11222", infinispanQuarkusAppProps["quarkus.infinispan-client.server-list"])
	assert.Equal(t, "true", infinispanQuarkusAppProps["quarkus.infinispan-client.use-auth"])
	assert.Equal(t, "PLAIN", infinispanQuarkusAppProps["quarkus.infinispan-client.sasl-mechanism"])
	assert.Empty(t, infinispanQuarkusAppProps["quarkus.infinispan-client.auth-realm"])
	assert.NotEmpty(t, infinispanQuarkusAppProps["quarkus.infinispan-client.trust-store"])
	assert.Len(t, kogitoInfra.Status.Volume, 1)
}
