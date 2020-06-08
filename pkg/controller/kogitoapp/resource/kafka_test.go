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

package resource

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
)

func createKafkaTestService(kogitoInfra *v1alpha1.KogitoInfra, port int) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: kogitoInfra.Status.Kafka.Service, Namespace: kogitoInfra.Namespace},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					TargetPort: intstr.FromInt(port),
				},
			},
		},
	}
}

func Test_CreateKafkaProperties_QuarkusRuntime(t *testing.T) {
	kogitoApp := createTestKogitoApp(v1alpha1.QuarkusRuntimeType)
	kogitoInfra := createTestKogitoInfra()
	service := createKafkaTestService(kogitoInfra, 9092)

	objs := []runtime.Object{service}
	fakeClient := test.CreateFakeClient(objs, nil, []runtime.Object{})
	envs, appProps, err := CreateKafkaProperties(fakeClient, kogitoInfra, kogitoApp)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(envs))
	assert.Equal(t, "TEST_BOOTSTRAP_SERVERS", envs[0].Name)
	assert.Equal(t, "test:9092", envs[0].Value)

	assert.Equal(t, 1, len(appProps))
	_, ok := appProps[propertiesKafkaQuarkus[appPropKafkaBootstrapServerList]]
	assert.True(t, ok)
}

func Test_CreateKafkaProperties_SpringBootRuntime(t *testing.T) {
	kogitoApp := createTestKogitoApp(v1alpha1.SpringbootRuntimeType)
	kogitoInfra := createTestKogitoInfra()
	service := createKafkaTestService(kogitoInfra, 9092)

	objs := []runtime.Object{service}
	fakeClient := test.CreateFakeClient(objs, nil, []runtime.Object{})
	envs, appProps, err := CreateKafkaProperties(fakeClient, kogitoInfra, kogitoApp)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(envs))
	assert.Equal(t, "TEST_BOOTSTRAP_SERVERS", envs[0].Name)
	assert.Equal(t, "test:9092", envs[0].Value)

	assert.Equal(t, 1, len(appProps))
	_, ok := appProps[propertiesKafkaSpring[appPropKafkaBootstrapServerList]]
	assert.True(t, ok)
}
