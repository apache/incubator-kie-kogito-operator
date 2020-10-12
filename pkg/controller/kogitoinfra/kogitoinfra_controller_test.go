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
	"testing"

	v12 "github.com/infinispan/infinispan-operator/pkg/apis/infinispan/v1"
	v14 "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/grafana"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/infinispan"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/kafka"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	v13 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Test_Reconcile_KafkaResource(t *testing.T) {
	kogitoInfra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: "kogito-kafka", Namespace: t.Name()},
		Spec: v1alpha1.KogitoInfraSpec{
			Resource: v1alpha1.Resource{
				APIVersion: kafka.APIVersion,
				Kind:       kafka.Kind,
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
	kafkaAppProps := kogitoInfra.Status.AppProps
	assert.Contains(t, "kogito-kafka:9090", kafkaAppProps["kafka.bootstrap.servers"])
}

func Test_Reconcile_Infinispan(t *testing.T) {

	kogitoInfra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: "kogito-infinispan", Namespace: t.Name()},
		Spec: v1alpha1.KogitoInfraSpec{
			Resource: v1alpha1.Resource{
				APIVersion: infinispan.APIVersion,
				Kind:       infinispan.Kind,
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
					TargetPort: intstr.FromInt(11222),
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
	infinispanAppProps := kogitoInfra.Status.AppProps
	assert.Equal(t, "kogito-infinispan:11222", infinispanAppProps["quarkus.infinispan-client.server-list"])
	assert.Equal(t, "true", infinispanAppProps["quarkus.infinispan-client.use-auth"])
	assert.Equal(t, "PLAIN", infinispanAppProps["quarkus.infinispan-client.sasl-mechanism"])
	assert.Empty(t, infinispanAppProps["quarkus.infinispan-client.auth-realm"])
}

func Test_Reconcile_Grafana(t *testing.T) {

	kogitoInfra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: "kogito-grafana", Namespace: t.Name()},
		Spec: v1alpha1.KogitoInfraSpec{
			Resource: v1alpha1.Resource{
				APIVersion: grafana.APIVersion,
				Kind:       grafana.Kind,
				Name:       "kogito-grafana",
				Namespace:  t.Name(),
			},
		},
	}

	deployedGrafana := &v14.Grafana{
		ObjectMeta: v1.ObjectMeta{Name: "kogito-grafana", Namespace: t.Name()},
	}

	grafanaService := &v13.Service{
		ObjectMeta: v1.ObjectMeta{Name: "kogito-grafana", Namespace: t.Name()},
		Spec: v13.ServiceSpec{
			Ports: []v13.ServicePort{
				{
					TargetPort: intstr.FromInt(11222),
				},
			},
		},
	}

	client := test.CreateFakeClient([]runtime.Object{
		kogitoInfra,
		deployedGrafana,
		grafanaService,
	}, nil, nil)

	scheme := meta.GetRegisteredSchema()
	r := &ReconcileKogitoInfra{client: client, scheme: scheme}
	// basic checks
	test.AssertReconcile(t, r, kogitoInfra)

	exists, err := kubernetes.ResourceC(client).Fetch(kogitoInfra)
	assert.NoError(t, err)
	assert.True(t, exists)
}
