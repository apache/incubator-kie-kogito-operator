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

package infinispan

import (
	"fmt"
	infinispanv1 "github.com/infinispan/infinispan-operator/pkg/apis/infinispan/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"testing"
)

func Test_CreateRequiredInfinispanResources_NewResources(t *testing.T) {
	kogitoInfra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{
			Name:      "kogito-infra",
			Namespace: t.Name(),
		},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallInfinispan: true,
		},
	}
	infinispan := &infinispanv1.Infinispan{
		ObjectMeta: v1.ObjectMeta{
			Namespace: t.Name(),
			Name:      InstanceName,
		},
		Spec: infinispanv1.InfinispanSpec{
			Replicas: 1,
			Security: infinispanv1.InfinispanSecurity{
				EndpointSecretName: fmt.Sprintf("my-infinispan-%s", InstanceName),
			},
		},
	}
	infinispanSecret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      fmt.Sprintf("my-infinispan-%s", InstanceName),
			Namespace: t.Name(),
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{kogitoInfra, infinispanSecret, infinispan}, nil, nil)
	resources, err := CreateRequiredResources(kogitoInfra, cli)

	assert.NoError(t, err)
	assert.Len(t, resources, 2)
	assert.Equal(t, secretName, resources[reflect.TypeOf(corev1.Secret{})][0].GetName())
	assert.Equal(t, InstanceName, resources[reflect.TypeOf(infinispanv1.Infinispan{})][0].GetName())
}

func Test_CreateRequiredInfinispanResources_HaveGeneratedSecret(t *testing.T) {
	secreteMap := make(map[string][]byte)
	secreteMap[infrastructure.InfinispanIdentityFileName] =
		[]byte(`
credentials:
- username: testuser
  password: testpassword`)

	kogitoInfra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{
			Name:      "kogito-infra",
			Namespace: t.Name(),
		},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallInfinispan: true,
		},
		Status: v1alpha1.KogitoInfraStatus{
			Infinispan: v1alpha1.InfinispanInstallStatus{
				InfraComponentInstallStatusType: v1alpha1.InfraComponentInstallStatusType{
					Name: InstanceName,
				},
			},
		},
	}
	infinispan := &infinispanv1.Infinispan{
		ObjectMeta: v1.ObjectMeta{
			Namespace: t.Name(),
			Name:      InstanceName,
		},
		Spec: infinispanv1.InfinispanSpec{
			Replicas: 1,
			Security: infinispanv1.InfinispanSecurity{
				EndpointSecretName: fmt.Sprintf("my-infinispan-%s", InstanceName),
			},
		},
	}
	infinispanSecret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      fmt.Sprintf("my-infinispan-%s", InstanceName),
			Namespace: t.Name(),
		},
		Data: secreteMap,
	}
	cli := test.CreateFakeClient([]runtime.Object{kogitoInfra, infinispanSecret, infinispan}, nil, nil)
	secret, err := newInfinispanLinkedSecret(kogitoInfra, cli)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", secret.StringData[infrastructure.InfinispanSecretUsernameKey])
	assert.Equal(t, "testpassword", secret.StringData[infrastructure.InfinispanSecretPasswordKey])

	resources, err := CreateRequiredResources(kogitoInfra, cli)
	assert.NoError(t, err)
	assert.Len(t, resources, 2)
	assert.Equal(t, secretName, resources[reflect.TypeOf(corev1.Secret{})][0].GetName())
	assert.Equal(t, InstanceName, resources[reflect.TypeOf(infinispanv1.Infinispan{})][0].GetName())
}
