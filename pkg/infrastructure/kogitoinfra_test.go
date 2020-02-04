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

package infrastructure

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func Test_EnsureKogitoInfra_NotExists(t *testing.T) {
	ns := t.Name()
	cli := test.CreateFakeClient(nil, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).Apply()
	assertDefaults(t, infra, err)
	assert.True(t, ready)
	assert.False(t, infra.Spec.InstallKafka)
	assert.False(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKeycloak)
}

func Test_EnsureKogitoInfra_NotExists_WithKafka(t *testing.T) {
	ns := t.Name()
	cli := test.CreateFakeClient(nil, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithKafka().Apply()
	assertDefaults(t, infra, err)
	assert.False(t, ready)
	assert.True(t, infra.Spec.InstallKafka)
	assert.False(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKeycloak)
}

func Test_EnsureKogitoInfra_NotExists_WithoutKafka(t *testing.T) {
	ns := t.Name()
	cli := test.CreateFakeClient(nil, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithoutKafka().Apply()
	assertDefaults(t, infra, err)
	assert.True(t, ready)
	assert.False(t, infra.Spec.InstallKafka)
	assert.False(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKeycloak)
}

func Test_EnsureKogitoInfra_NotExists_WithInfinispan(t *testing.T) {
	ns := t.Name()
	cli := test.CreateFakeClient(nil, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithInfinispan().Apply()
	assertDefaults(t, infra, err)
	assert.False(t, ready)
	assert.True(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKafka)
	assert.False(t, infra.Spec.InstallKeycloak)
}

func Test_EnsureKogitoInfra_NotExists_WithoutInfinispan(t *testing.T) {
	ns := t.Name()
	cli := test.CreateFakeClient(nil, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithoutInfinispan().Apply()
	assertDefaults(t, infra, err)
	assert.True(t, ready)
	assert.False(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKafka)
	assert.False(t, infra.Spec.InstallKeycloak)
}

func Test_EnsureKogitoInfra_NotExists_WithKeycloak(t *testing.T) {
	ns := t.Name()
	cli := test.CreateFakeClient(nil, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithKeycloak().Apply()
	assertDefaults(t, infra, err)
	assert.False(t, ready)
	assert.False(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKafka)
	assert.True(t, infra.Spec.InstallKeycloak)
}

func Test_EnsureKogitoInfra_NotExists_WithoutKeycloak(t *testing.T) {
	ns := t.Name()
	cli := test.CreateFakeClient(nil, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithoutKeycloak().Apply()
	assertDefaults(t, infra, err)
	assert.True(t, ready)
	assert.False(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKafka)
	assert.False(t, infra.Spec.InstallKeycloak)
}

func Test_EnsureKogitoInfra_NotExists_AddAllComponents(t *testing.T) {
	ns := t.Name()
	cli := test.CreateFakeClient(nil, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithInfinispan().WithKafka().WithKeycloak().Apply()
	assertDefaults(t, infra, err)
	assert.False(t, ready)
	assert.True(t, infra.Spec.InstallInfinispan)
	assert.True(t, infra.Spec.InstallKafka)
	assert.True(t, infra.Spec.InstallKeycloak)
}

func Test_EnsureKogitoInfra_NotExists_RemoveAnyComponents(t *testing.T) {
	ns := t.Name()
	cli := test.CreateFakeClient(nil, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithoutInfinispan().WithoutKafka().WithoutKeycloak().Apply()
	assertDefaults(t, infra, err)
	assert.True(t, ready)
	assert.False(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKafka)
	assert.False(t, infra.Spec.InstallKeycloak)
}

func Test_EnsureKogitoInfra_Exists(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec:       v1alpha1.KogitoInfraSpec{},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).Apply()
	assertDefaults(t, infra, err)
	assert.True(t, ready)
	assert.False(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKafka)
	assert.False(t, infra.Spec.InstallKeycloak)
}

func Test_EnsureKogitoInfra_Exists_NoChange(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallInfinispan: true,
			InstallKafka:      true,
			InstallKeycloak:   true,
		},
		Status: v1alpha1.KogitoInfraStatus{
			Infinispan: v1alpha1.InfinispanInstallStatus{
				InfraComponentInstallStatusType: fakeInstalledInfraComponentInstallStatusType(),
			},
			Kafka: fakeInstalledInfraComponentInstallStatusType(),
			Keycloak: v1alpha1.KeycloakInstallStatus{
				InfraComponentInstallStatusType: fakeInstalledInfraComponentInstallStatusType(),
				RealmStatus:                     fakeInstalledInfraComponentInstallStatusType(),
			},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).Apply()
	assertDefaults(t, infra, err)
	assert.True(t, ready)
	assert.True(t, infra.Spec.InstallInfinispan)
	assert.True(t, infra.Spec.InstallKafka)
	assert.True(t, infra.Spec.InstallKeycloak)
}

func Test_EnsureKogitoInfra_Exists_AddInfinispan(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec:       v1alpha1.KogitoInfraSpec{},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithInfinispan().Apply()
	assertDefaults(t, infra, err)
	assert.False(t, ready)
	assert.True(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKafka)
	assert.False(t, infra.Spec.InstallKeycloak)
}

func Test_EnsureKogitoInfra_Exists_RemoveInfinispan(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallInfinispan: true,
		},
		Status: v1alpha1.KogitoInfraStatus{
			Infinispan: v1alpha1.InfinispanInstallStatus{
				InfraComponentInstallStatusType: fakeInstalledInfraComponentInstallStatusType(),
			},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithoutInfinispan().Apply()
	assertDefaults(t, infra, err)
	assert.False(t, ready)
	assert.False(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKafka)
	assert.False(t, infra.Spec.InstallKeycloak)
}

func Test_EnsureKogitoInfra_Exists_AddKafka(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec:       v1alpha1.KogitoInfraSpec{},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithKafka().Apply()
	assertDefaults(t, infra, err)
	assert.False(t, ready)
	assert.True(t, infra.Spec.InstallKafka)
	assert.False(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKeycloak)
}

func Test_EnsureKogitoInfra_Exists_RemoveKafka(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallKafka: true,
		},
		Status: v1alpha1.KogitoInfraStatus{
			Kafka: fakeInstalledInfraComponentInstallStatusType(),
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithoutKafka().Apply()
	assertDefaults(t, infra, err)
	assert.False(t, ready)
	assert.False(t, infra.Spec.InstallKafka)
	assert.False(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKeycloak)
}

func Test_EnsureKogitoInfra_Exists_AddKeycloak(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec:       v1alpha1.KogitoInfraSpec{},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithKeycloak().Apply()
	assertDefaults(t, infra, err)
	assert.False(t, ready)
	assert.False(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKafka)
	assert.True(t, infra.Spec.InstallKeycloak)
}

func Test_EnsureKogitoInfra_Exists_RemoveKeycloak(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallKeycloak: true,
		},
		Status: v1alpha1.KogitoInfraStatus{
			Keycloak: v1alpha1.KeycloakInstallStatus{
				InfraComponentInstallStatusType: fakeInstalledInfraComponentInstallStatusType(),
				RealmStatus:                     fakeInstalledInfraComponentInstallStatusType(),
			},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithoutKeycloak().Apply()
	assertDefaults(t, infra, err)
	assert.False(t, ready)
	assert.False(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKafka)
	assert.False(t, infra.Spec.InstallKeycloak)
}

func Test_EnsureKogitoInfra_Exists_AddAllComponents(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec:       v1alpha1.KogitoInfraSpec{},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithKafka().WithInfinispan().WithKeycloak().Apply()
	assertDefaults(t, infra, err)
	assert.False(t, ready)
	assert.True(t, infra.Spec.InstallInfinispan)
	assert.True(t, infra.Spec.InstallKafka)
	assert.True(t, infra.Spec.InstallKeycloak)
}

func Test_EnsureKogitoInfra_Exists_RemoveAllComponents(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallInfinispan: true,
			InstallKafka:      true,
			InstallKeycloak:   true,
		},
		Status: v1alpha1.KogitoInfraStatus{
			Infinispan: v1alpha1.InfinispanInstallStatus{
				InfraComponentInstallStatusType: fakeInstalledInfraComponentInstallStatusType(),
			},
			Kafka: fakeInstalledInfraComponentInstallStatusType(),
			Keycloak: v1alpha1.KeycloakInstallStatus{
				InfraComponentInstallStatusType: fakeInstalledInfraComponentInstallStatusType(),
				RealmStatus:                     fakeInstalledInfraComponentInstallStatusType(),
			},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithoutKafka().WithoutInfinispan().WithoutKeycloak().Apply()
	assertDefaults(t, infra, err)
	assert.False(t, ready)
	assert.False(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKafka)
	assert.False(t, infra.Spec.InstallKeycloak)
}

func Test_EnsureKogitoInfra_Exists_WithInfinispanButAlreadyInstalled(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallInfinispan: true,
		},
		Status: v1alpha1.KogitoInfraStatus{
			Infinispan: v1alpha1.InfinispanInstallStatus{
				InfraComponentInstallStatusType: fakeInstalledInfraComponentInstallStatusType(),
			},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithInfinispan().Apply()
	assertDefaults(t, infra, err)
	assert.True(t, ready)
	assert.True(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKafka)
	assert.False(t, infra.Spec.InstallKeycloak)
}

func Test_EnsureKogitoInfra_Exists_WithKafkaButAlreadyInstalled(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallKafka: true,
		},
		Status: v1alpha1.KogitoInfraStatus{
			Kafka: fakeInstalledInfraComponentInstallStatusType(),
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithKafka().Apply()
	assertDefaults(t, infra, err)
	assert.True(t, ready)
	assert.True(t, infra.Spec.InstallKafka)
	assert.False(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKeycloak)
}

func Test_EnsureKogitoInfra_Exists_WithKeycloakButAlreadyInstalled(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallKeycloak: true,
		},
		Status: v1alpha1.KogitoInfraStatus{
			Keycloak: v1alpha1.KeycloakInstallStatus{
				InfraComponentInstallStatusType: fakeInstalledInfraComponentInstallStatusType(),
				RealmStatus:                     fakeInstalledInfraComponentInstallStatusType(),
			},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithKeycloak().Apply()
	assertDefaults(t, infra, err)
	assert.True(t, ready)
	assert.False(t, infra.Spec.InstallKafka)
	assert.False(t, infra.Spec.InstallInfinispan)
	assert.True(t, infra.Spec.InstallKeycloak)
}

func Test_EnsureKogitoInfra_Exists_RemoveOnlyKafka(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallInfinispan: true,
			InstallKafka:      true,
			InstallKeycloak:   true,
		},
		Status: v1alpha1.KogitoInfraStatus{
			Infinispan: v1alpha1.InfinispanInstallStatus{
				InfraComponentInstallStatusType: fakeInstalledInfraComponentInstallStatusType(),
			},
			Kafka: fakeInstalledInfraComponentInstallStatusType(),
			Keycloak: v1alpha1.KeycloakInstallStatus{
				InfraComponentInstallStatusType: fakeInstalledInfraComponentInstallStatusType(),
				RealmStatus:                     fakeInstalledInfraComponentInstallStatusType(),
			},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithoutKafka().Apply()
	assertDefaults(t, infra, err)
	assert.False(t, ready)
	assert.False(t, infra.Spec.InstallKafka)
	assert.True(t, infra.Spec.InstallInfinispan)
	assert.True(t, infra.Spec.InstallKeycloak)
}

func Test_EnsureKogitoInfra_Exists_RemoveOnlyInfinispan(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallInfinispan: true,
			InstallKafka:      true,
			InstallKeycloak:   true,
		},
		Status: v1alpha1.KogitoInfraStatus{
			Infinispan: v1alpha1.InfinispanInstallStatus{
				InfraComponentInstallStatusType: fakeInstalledInfraComponentInstallStatusType(),
			},
			Kafka: fakeInstalledInfraComponentInstallStatusType(),
			Keycloak: v1alpha1.KeycloakInstallStatus{
				InfraComponentInstallStatusType: fakeInstalledInfraComponentInstallStatusType(),
				RealmStatus:                     fakeInstalledInfraComponentInstallStatusType(),
			},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithoutInfinispan().Apply()
	assertDefaults(t, infra, err)
	assert.False(t, ready)
	assert.True(t, infra.Spec.InstallKafka)
	assert.False(t, infra.Spec.InstallInfinispan)
	assert.True(t, infra.Spec.InstallKeycloak)
}

func Test_EnsureKogitoInfra_Exists_RemoveOnlyKeycloak(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallInfinispan: true,
			InstallKafka:      true,
			InstallKeycloak:   true,
		},
		Status: v1alpha1.KogitoInfraStatus{
			Infinispan: v1alpha1.InfinispanInstallStatus{
				InfraComponentInstallStatusType: fakeInstalledInfraComponentInstallStatusType(),
			},
			Kafka: fakeInstalledInfraComponentInstallStatusType(),
			Keycloak: v1alpha1.KeycloakInstallStatus{
				InfraComponentInstallStatusType: fakeInstalledInfraComponentInstallStatusType(),
				RealmStatus:                     fakeInstalledInfraComponentInstallStatusType(),
			},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithoutKeycloak().Apply()
	assertDefaults(t, infra, err)
	assert.False(t, ready)
	assert.True(t, infra.Spec.InstallKafka)
	assert.True(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKeycloak)
}

func fakeInstalledInfraComponentInstallStatusType() v1alpha1.InfraComponentInstallStatusType {
	return v1alpha1.InfraComponentInstallStatusType{
		Service: "test",
		Condition: []v1alpha1.InstallCondition{
			{
				Type: v1alpha1.SuccessInstallConditionType,
			},
		},
	}
}

func assertDefaults(t *testing.T, infra *v1alpha1.KogitoInfra, err error) {
	assert.NoError(t, err)
	assert.NotNil(t, infra)
	assert.Equal(t, DefaultKogitoInfraName, infra.GetName())
}
