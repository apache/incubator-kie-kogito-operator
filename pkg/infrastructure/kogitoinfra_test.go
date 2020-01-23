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
	assert.NoError(t, err)
	assert.NotNil(t, infra)
	assert.True(t, ready)
	assert.Equal(t, DefaultKogitoInfraName, infra.GetName())
}

func Test_EnsureKogitoInfra_NotExists_WithKafka(t *testing.T) {
	ns := t.Name()
	cli := test.CreateFakeClient(nil, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithKafka().Apply()
	assert.NoError(t, err)
	assert.NotNil(t, infra)
	assert.False(t, ready)
	assert.True(t, infra.Spec.InstallKafka)
	assert.False(t, infra.Spec.InstallInfinispan)
	assert.Equal(t, DefaultKogitoInfraName, infra.GetName())
}

func Test_EnsureKogitoInfra_NotExists_WithoutKafka(t *testing.T) {
	ns := t.Name()
	cli := test.CreateFakeClient(nil, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithoutKafka().Apply()
	assert.NoError(t, err)
	assert.NotNil(t, infra)
	assert.True(t, ready)
	assert.False(t, infra.Spec.InstallKafka)
	assert.False(t, infra.Spec.InstallInfinispan)
	assert.Equal(t, DefaultKogitoInfraName, infra.GetName())
}

func Test_EnsureKogitoInfra_NotExists_WithInfinispan(t *testing.T) {
	ns := t.Name()
	cli := test.CreateFakeClient(nil, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithInfinispan().Apply()
	assert.NoError(t, err)
	assert.NotNil(t, infra)
	assert.False(t, ready)
	assert.True(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKafka)
	assert.Equal(t, DefaultKogitoInfraName, infra.GetName())
}

func Test_EnsureKogitoInfra_NotExists_WithoutInfinispan(t *testing.T) {
	ns := t.Name()
	cli := test.CreateFakeClient(nil, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithoutInfinispan().Apply()
	assert.NoError(t, err)
	assert.NotNil(t, infra)
	assert.True(t, ready)
	assert.False(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKafka)
	assert.Equal(t, DefaultKogitoInfraName, infra.GetName())
}

func Test_EnsureKogitoInfra_NotExists_WithInfinispanAndKafka(t *testing.T) {
	ns := t.Name()
	cli := test.CreateFakeClient(nil, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithInfinispan().WithKafka().Apply()
	assert.NoError(t, err)
	assert.NotNil(t, infra)
	assert.False(t, ready)
	assert.True(t, infra.Spec.InstallInfinispan)
	assert.True(t, infra.Spec.InstallKafka)
	assert.Equal(t, DefaultKogitoInfraName, infra.GetName())
}

func Test_EnsureKogitoInfra_NotExists_WithoutInfinispanAndKafka(t *testing.T) {
	ns := t.Name()
	cli := test.CreateFakeClient(nil, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithoutInfinispan().WithoutKafka().Apply()
	assert.NoError(t, err)
	assert.NotNil(t, infra)
	assert.True(t, ready)
	assert.False(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKafka)
	assert.Equal(t, DefaultKogitoInfraName, infra.GetName())
}

func Test_EnsureKogitoInfra_Exists(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallInfinispan: false,
			InstallKafka:      false,
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).Apply()
	assert.NoError(t, err)
	assert.NotNil(t, infra)
	assert.True(t, ready)
	assert.Equal(t, DefaultKogitoInfraName, infra.GetName())
	assert.False(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKafka)
}

func Test_EnsureKogitoInfra_Exists_AddInfinispan(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallInfinispan: false,
			InstallKafka:      false,
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithInfinispan().Apply()
	assert.NoError(t, err)
	assert.NotNil(t, infra)
	assert.False(t, ready)
	assert.Equal(t, DefaultKogitoInfraName, infra.GetName())
	assert.True(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKafka)
}

func Test_EnsureKogitoInfra_Exists_RemoveInfinispan(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallInfinispan: true,
			InstallKafka:      false,
		},
		Status: v1alpha1.KogitoInfraStatus{
			Infinispan: v1alpha1.InfinispanInstallStatus{
				InfraComponentInstallStatusType: fakeInstalledInfraComponentInstallStatusType(),
			},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithoutInfinispan().Apply()
	assert.NoError(t, err)
	assert.NotNil(t, infra)
	assert.False(t, ready)
	assert.Equal(t, DefaultKogitoInfraName, infra.GetName())
	assert.False(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKafka)
}

func Test_EnsureKogitoInfra_Exists_AddKafka(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallInfinispan: false,
			InstallKafka:      false,
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithKafka().Apply()
	assert.NoError(t, err)
	assert.NotNil(t, infra)
	assert.False(t, ready)
	assert.Equal(t, DefaultKogitoInfraName, infra.GetName())
	assert.True(t, infra.Spec.InstallKafka)
	assert.False(t, infra.Spec.InstallInfinispan)
}

func Test_EnsureKogitoInfra_Exists_RemoveKafka(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallInfinispan: false,
			InstallKafka:      true,
		},
		Status: v1alpha1.KogitoInfraStatus{
			Kafka: fakeInstalledInfraComponentInstallStatusType(),
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithoutKafka().Apply()
	assert.NoError(t, err)
	assert.NotNil(t, infra)
	assert.False(t, ready)
	assert.Equal(t, DefaultKogitoInfraName, infra.GetName())
	assert.False(t, infra.Spec.InstallKafka)
	assert.False(t, infra.Spec.InstallInfinispan)
}

func Test_EnsureKogitoInfra_Exists_AddInfinispanAndKafka(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallInfinispan: false,
			InstallKafka:      false,
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithKafka().WithInfinispan().Apply()
	assert.NoError(t, err)
	assert.NotNil(t, infra)
	assert.False(t, ready)
	assert.Equal(t, DefaultKogitoInfraName, infra.GetName())
	assert.True(t, infra.Spec.InstallInfinispan)
	assert.True(t, infra.Spec.InstallKafka)
}

func Test_EnsureKogitoInfra_Exists_RemoveInfinispanAndKafka(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallInfinispan: true,
			InstallKafka:      true,
		},
		Status: v1alpha1.KogitoInfraStatus{
			Infinispan: v1alpha1.InfinispanInstallStatus{
				InfraComponentInstallStatusType: fakeInstalledInfraComponentInstallStatusType(),
			},
			Kafka: fakeInstalledInfraComponentInstallStatusType(),
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithoutKafka().WithoutInfinispan().Apply()
	assert.NoError(t, err)
	assert.NotNil(t, infra)
	assert.False(t, ready)
	assert.Equal(t, DefaultKogitoInfraName, infra.GetName())
	assert.False(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKafka)
}

func Test_EnsureKogitoInfra_Exists_WithInfinispanButAlreadyInstalled(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallInfinispan: true,
			InstallKafka:      false,
		},
		Status: v1alpha1.KogitoInfraStatus{
			Infinispan: v1alpha1.InfinispanInstallStatus{
				InfraComponentInstallStatusType: fakeInstalledInfraComponentInstallStatusType(),
			},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithInfinispan().Apply()
	assert.NoError(t, err)
	assert.NotNil(t, infra)
	assert.True(t, ready)
	assert.Equal(t, DefaultKogitoInfraName, infra.GetName())
	assert.True(t, infra.Spec.InstallInfinispan)
	assert.False(t, infra.Spec.InstallKafka)
}

func Test_EnsureKogitoInfra_Exists_WithKafkaButAlreadyInstalled(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallInfinispan: false,
			InstallKafka:      true,
		},
		Status: v1alpha1.KogitoInfraStatus{
			Kafka: fakeInstalledInfraComponentInstallStatusType(),
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithKafka().Apply()
	assert.NoError(t, err)
	assert.NotNil(t, infra)
	assert.True(t, ready)
	assert.Equal(t, DefaultKogitoInfraName, infra.GetName())
	assert.True(t, infra.Spec.InstallKafka)
	assert.False(t, infra.Spec.InstallInfinispan)
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
