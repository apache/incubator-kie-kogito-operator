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
	"fmt"
	v12 "github.com/infinispan/infinispan-operator/pkg/apis/infinispan/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/infinispan"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/kafka"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/keycloak"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v13 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

func TestReconcileKogitoInfra_Reconcile_AllInstalled(t *testing.T) {
	kogitoInfra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: infrastructure.DefaultKogitoInfraName, Namespace: t.Name()},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallInfinispan: true,
			InstallKafka:      true,
			InstallKeycloak:   true,
		},
	}
	client := test.CreateFakeClient([]runtime.Object{
		kogitoInfra,
		createInfinispanOperatorDeployment(t.Name()),
		createStrimziOperatorDeployment(t.Name()),
	},
		nil, nil)
	scheme := meta.GetRegisteredSchema()
	request := reconcile.Request{NamespacedName: types.NamespacedName{Name: kogitoInfra.Name, Namespace: kogitoInfra.Namespace}}

	r := ReconcileKogitoInfra{client: client, scheme: scheme}

	res, err := r.Reconcile(request)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	// we shouldn't have services for kafka nor infinispan, so requeue to give time for the 3rd party operators to create them
	assert.True(t, res.Requeue)

	exists, err := kubernetes.ResourceC(client).Fetch(kogitoInfra)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, infinispan.InstanceName, kogitoInfra.Status.Infinispan.Name)
	assert.Equal(t, kafka.InstanceName, kogitoInfra.Status.Kafka.Name)
	assert.Equal(t, keycloak.InstanceName, kogitoInfra.Status.Keycloak.Name)
	assert.Empty(t, kogitoInfra.Status.Kafka.Service)
	assert.Empty(t, kogitoInfra.Status.Infinispan.Service)
	assert.Empty(t, kogitoInfra.Status.Keycloak.Service)
}

func TestReconcileKogitoInfra_Reconcile_Keycloak(t *testing.T) {
	kogitoInfra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: infrastructure.DefaultKogitoInfraName, Namespace: t.Name()},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallKeycloak: true,
		},
	}
	client := test.CreateFakeClient([]runtime.Object{kogitoInfra}, nil, nil)
	scheme := meta.GetRegisteredSchema()
	request := reconcile.Request{NamespacedName: types.NamespacedName{Name: kogitoInfra.Name, Namespace: kogitoInfra.Namespace}}

	r := ReconcileKogitoInfra{client: client, scheme: scheme}

	res, err := r.Reconcile(request)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	// we shouldn't have services for kafka nor infinispan, so requeue to give time for the 3rd party operators to create them
	assert.True(t, res.Requeue)

	exists, err := kubernetes.ResourceC(client).Fetch(kogitoInfra)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Empty(t, kogitoInfra.Status.Infinispan.Name)
	assert.Empty(t, kogitoInfra.Status.Kafka.Name)
	assert.Equal(t, keycloak.InstanceName, kogitoInfra.Status.Keycloak.Name)
	assert.Empty(t, kogitoInfra.Status.Kafka.Service)
	assert.Empty(t, kogitoInfra.Status.Infinispan.Service)
	assert.Empty(t, kogitoInfra.Status.Keycloak.Service)
}

func createInfinispanOperatorDeployment(namespace string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Namespace: namespace,
			Name:      infrastructure.InfinispanOperatorName,
		},
	}
}

func createStrimziOperatorDeployment(namespace string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Namespace: namespace,
			Name:      fmt.Sprintf("%s-version-whatever", infrastructure.StrimziOperatorName),
			OwnerReferences: []v1.OwnerReference{
				{
					Name: fmt.Sprintf("%s-version-whatever", infrastructure.StrimziOperatorName),
				},
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &v1.LabelSelector{MatchLabels: map[string]string{"name": infrastructure.StrimziOperatorName}},
		},
	}
}

func Test_Reconcile_KafkaResource(t *testing.T) {
	kogitoInfra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: infrastructure.DefaultKogitoInfraName, Namespace: t.Name()},
		Spec: v1alpha1.KogitoInfraSpec{
			Resource: v1alpha1.Resource{
				Kind:      "kafka.strimzi.io/v1beta1",
				Name:      "kogito-kafka",
				Namespace: t.Name(),
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
	request := reconcile.Request{NamespacedName: types.NamespacedName{Name: kogitoInfra.Name, Namespace: kogitoInfra.Namespace}}
	r := ReconcileKogitoInfra{client: client, scheme: scheme}
	_, err := r.Reconcile(request)
	assert.NoError(t, err)

	exists, err := kubernetes.ResourceC(client).Fetch(kogitoInfra)
	assert.NoError(t, err)
	assert.True(t, exists)
	kafkaProperties := kogitoInfra.Status.Kafka.KafkaProperties
	assert.Equal(t, "kogito-kafka:9090", kafkaProperties.ExternalURI)
}

func Test_Reconcile_Infinispan(t *testing.T) {

	kogitoInfra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: infrastructure.DefaultKogitoInfraName, Namespace: t.Name()},
		Spec: v1alpha1.KogitoInfraSpec{
			Resource: v1alpha1.Resource{
				Kind:      "infinispan.org/v1",
				Name:      "kogito-infinispan",
				Namespace: t.Name(),
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
	request := reconcile.Request{NamespacedName: types.NamespacedName{Name: kogitoInfra.Name, Namespace: kogitoInfra.Namespace}}
	r := ReconcileKogitoInfra{client: client, scheme: scheme}
	_, err := r.Reconcile(request)
	assert.NoError(t, err)

	exists, err := kubernetes.ResourceC(client).Fetch(kogitoInfra)
	assert.NoError(t, err)
	assert.True(t, exists)
	infinispanProperties := kogitoInfra.Status.Infinispan.InfinispanProperties
	assert.Equal(t, "kogito-infinispan:11222", infinispanProperties.URI)
	assert.Equal(t, "kogito-infinispan-credential", infinispanProperties.Credentials.SecretName)
	assert.Equal(t, infrastructure.InfinispanSecretUsernameKey, infinispanProperties.Credentials.UsernameKey)
	assert.Equal(t, infrastructure.InfinispanSecretPasswordKey, infinispanProperties.Credentials.PasswordKey)
}
