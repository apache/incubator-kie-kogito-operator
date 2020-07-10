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

package service

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/test"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_InstallRuntimeService(t *testing.T) {
	ns := t.Name()
	name := "example-drools"
	runtimeFlag := flag.RuntimeFlags{
		Name: name,
		InstallFlags: flag.InstallFlags{
			Project: ns,
			ImageFlags: flag.ImageFlags{
				Image:                 "quay.io/kiegroup/drools-quarkus-example:1.0",
				InsecureImageRegistry: true,
			},
			PodResourceFlags: flag.PodResourceFlags{
				Limits:   []string{"cpu=1", "memory=1Gi"},
				Requests: []string{"cpu=1", "memory=1Gi"},
			},
			HTTPPort: int32(9090),
			Replicas: 2,
		},
		EnableIstio:       true,
		EnablePersistence: true,
		EnableEvents:      true,
		RuntimeTypeFlags: flag.RuntimeTypeFlags{
			Runtime: "springboot",
		},
	}
	client := test.SetupFakeKubeCli(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	resourceCheckService := new(shared.ResourceCheckServiceMock)
	resourceCheckService.On("CheckKogitoRuntimeNotExists", client, name, ns).Return(nil)
	runtimeService := runtimeServiceImpl{
		resourceCheckService: resourceCheckService,
	}
	err := runtimeService.InstallRuntimeService(client, &runtimeFlag)
	assert.NoError(t, err)
	// This should be created, given the command above
	kogitoRuntime := &v1alpha1.KogitoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-drools",
			Namespace: ns,
		},
	}

	exist, err := kubernetes.ResourceC(client).Fetch(kogitoRuntime)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, "quay.io", kogitoRuntime.Spec.Image.Domain)
	assert.Equal(t, "kiegroup", kogitoRuntime.Spec.Image.Namespace)
	assert.Equal(t, "drools-quarkus-example", kogitoRuntime.Spec.Image.Name)
	assert.Equal(t, "1.0", kogitoRuntime.Spec.Image.Tag)
	assert.Equal(t, v1alpha1.SpringbootRuntimeType, kogitoRuntime.Spec.Runtime)
	assert.True(t, kogitoRuntime.Spec.InfinispanMeta.InfinispanProperties.UseKogitoInfra)
	assert.True(t, kogitoRuntime.Spec.KafkaMeta.KafkaProperties.UseKogitoInfra)
	assert.True(t, kogitoRuntime.Spec.EnableIstio)
	assert.Equal(t, int32(2), *kogitoRuntime.Spec.Replicas)
	assert.Equal(t, int32(9090), kogitoRuntime.Spec.HTTPPort)
	assert.Equal(t, *kogitoRuntime.Spec.KogitoServiceSpec.Resources.Limits.Cpu(), resource.MustParse("1"))
	assert.Equal(t, *kogitoRuntime.Spec.KogitoServiceSpec.Resources.Requests.Memory(), resource.MustParse("1Gi"))
	assert.True(t, kogitoRuntime.Spec.InsecureImageRegistry)
}

func Test_DeleteRuntimeService_WhenBuildExists(t *testing.T) {
	ns := t.Name()
	name := "example-quarkus"
	client := test.SetupFakeKubeCli(&v1alpha1.KogitoRuntime{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns}})
	resourceCheckService := new(shared.ResourceCheckServiceMock)
	resourceCheckService.On("CheckKogitoRuntimeExists", client, name, ns).Return(nil)

	runtimeService := runtimeServiceImpl{
		resourceCheckService: resourceCheckService,
	}
	err := runtimeService.DeleteRuntimeService(client, name, ns)
	assert.NoError(t, err)
}

func Test_DeleteRuntimeService_WhenBuildNotExists(t *testing.T) {
	ns := t.Name()
	name := "example-quarkus"
	client := test.SetupFakeKubeCli()
	resourceCheckService := new(shared.ResourceCheckServiceMock)
	resourceCheckService.On("CheckKogitoRuntimeExists", client, name, ns).Return(fmt.Errorf(""))

	runtimeService := runtimeServiceImpl{
		resourceCheckService: resourceCheckService,
	}
	err := runtimeService.DeleteRuntimeService(client, name, ns)
	assert.Error(t, err)
}
