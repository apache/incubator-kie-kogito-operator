// Copyright 2021 Red Hat, Inc. and/or its affiliates
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

package v1beta1

import (
	"github.com/kiegroup/kogito-operator/api"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestKogitoInfra_Spec(t *testing.T) {
	instance := &KogitoInfra{
		Spec: KogitoInfraSpec{
			Resource: InfraResource{
				APIVersion: "infinispan.org/v1",
				Kind:       "Infinispan",
				Name:       "test-infinispan",
				Namespace:  t.Name(),
			},
			InfraProperties: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}

	spec := instance.GetSpec()
	assert.Equal(t, "test-infinispan", spec.GetResource().GetName())
	assert.Equal(t, "infinispan.org/v1", spec.GetResource().GetAPIVersion())
	assert.Equal(t, "Infinispan", spec.GetResource().GetKind())
	assert.Equal(t, t.Name(), spec.GetResource().GetNamespace())
	assert.Equal(t, 2, len(spec.GetInfraProperties()))
	assert.Equal(t, "value1", spec.GetInfraProperties()["key1"])
	assert.Equal(t, "value2", spec.GetInfraProperties()["key2"])
}

func TestKogitoInfra_Status(t *testing.T) {
	instance1 := &KogitoInfra{
		Status: KogitoInfraStatus{
			Conditions: &[]metav1.Condition{
				{
					Type:    string(api.KogitoInfraConfigured),
					Status:  metav1.ConditionTrue,
					Reason:  string(api.ReconciliationFailure),
					Message: "Infra success",
				},
			},
		},
	}
	status := instance1.GetStatus()
	quarkusRuntimeProperties := RuntimeProperties{
		AppProps: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		Env: []corev1.EnvVar{
			{
				Name: "name1",
			},
			{
				Name: "name2",
			},
		},
	}
	springbootRuntimeProperties := RuntimeProperties{
		AppProps: map[string]string{
			"key3": "value3",
			"key4": "value4",
		},
		Env: []corev1.EnvVar{
			{
				Name: "name3",
			},
			{
				Name: "name4",
			},
		},
	}
	status.AddRuntimeProperties(api.QuarkusRuntimeType, quarkusRuntimeProperties)
	status.AddRuntimeProperties(api.SpringBootRuntimeType, springbootRuntimeProperties)

	var volumes []api.KogitoInfraVolumeInterface
	volume1 := KogitoInfraVolume{
		Mount: corev1.VolumeMount{
			Name: "volumeMount1",
		},
		NamedVolume: ConfigVolume{
			Name: "configVolume1",
		},
	}

	volume2 := KogitoInfraVolume{
		Mount: corev1.VolumeMount{
			Name: "volumeMount2",
		},
		NamedVolume: ConfigVolume{
			Name: "configVolume2",
		},
	}
	volumes = append(volumes, volume1, volume2)
	status.SetVolumes(volumes)

	conditions := *status.GetConditions()
	assert.Equal(t, string(api.KogitoInfraConfigured), conditions[0].Type)
	assert.Equal(t, metav1.ConditionTrue, conditions[0].Status)
	assert.Equal(t, string(api.ReconciliationFailure), conditions[0].Reason)
	assert.Equal(t, "Infra success", conditions[0].Message)
	assert.Equal(t, 2, len(status.GetRuntimeProperties()))
	assert.Equal(t, 2, len(status.GetRuntimeProperties()[api.QuarkusRuntimeType].GetAppProps()))
	assert.Equal(t, "value1", status.GetRuntimeProperties()[api.QuarkusRuntimeType].GetAppProps()["key1"])
	assert.Equal(t, "value2", status.GetRuntimeProperties()[api.QuarkusRuntimeType].GetAppProps()["key2"])
	assert.Equal(t, 2, len(status.GetRuntimeProperties()[api.QuarkusRuntimeType].GetEnv()))
	assert.Equal(t, "name1", status.GetRuntimeProperties()[api.QuarkusRuntimeType].GetEnv()[0].Name)
	assert.Equal(t, "name2", status.GetRuntimeProperties()[api.QuarkusRuntimeType].GetEnv()[1].Name)
	assert.Equal(t, 2, len(status.GetRuntimeProperties()[api.SpringBootRuntimeType].GetAppProps()))
	assert.Equal(t, "value3", status.GetRuntimeProperties()[api.SpringBootRuntimeType].GetAppProps()["key3"])
	assert.Equal(t, "value4", status.GetRuntimeProperties()[api.SpringBootRuntimeType].GetAppProps()["key4"])
	assert.Equal(t, 2, len(status.GetRuntimeProperties()[api.SpringBootRuntimeType].GetEnv()))
	assert.Equal(t, "name3", status.GetRuntimeProperties()[api.SpringBootRuntimeType].GetEnv()[0].Name)
	assert.Equal(t, "name4", status.GetRuntimeProperties()[api.SpringBootRuntimeType].GetEnv()[1].Name)
	assert.Equal(t, 2, len(status.GetVolumes()))
	assert.Equal(t, "volumeMount1", status.GetVolumes()[0].GetMount().Name)
	assert.Equal(t, "configVolume1", status.GetVolumes()[0].GetNamedVolume().GetName())
	assert.Equal(t, "volumeMount2", status.GetVolumes()[1].GetMount().Name)
	assert.Equal(t, "configVolume2", status.GetVolumes()[1].GetNamedVolume().GetName())
}
