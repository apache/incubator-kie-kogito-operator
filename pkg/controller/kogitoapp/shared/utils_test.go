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

package shared

import (
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	cpuResource    = v1alpha1.ResourceCPU
	memoryResource = v1alpha1.ResourceMemory
)

func TestFromEnvToEnvVar(t *testing.T) {
	a := []v1alpha1.Env{
		{
			Name: "A", Value: "1",
		},
	}
	b := FromEnvToEnvVar(a)

	assert.True(t, a[0].Name == b[0].Name)
	assert.True(t, a[0].Value == b[0].Value)
}

func TestFromResourcesToResourcesRequirements(t *testing.T) {
	cpuValue := "1"
	memValue := "250Mi"
	a := v1alpha1.Resources{
		Limits: []v1alpha1.ResourceMap{
			{
				Resource: cpuResource,
				Value:    cpuValue,
			},
			{
				Resource: memoryResource,
				Value:    memValue,
			},
		},
		Requests: []v1alpha1.ResourceMap{
			{
				Resource: cpuResource,
				Value:    cpuValue,
			},
			{
				Resource: memoryResource,
				Value:    memValue,
			},
		},
	}
	b := FromResourcesToResourcesRequirements(a)

	assert.True(t, *b.Limits.Cpu() == resource.MustParse(a.Limits[0].Value))
	assert.True(t, *b.Requests.Cpu() == resource.MustParse(a.Requests[0].Value))
	assert.True(t, *b.Limits.Memory() == resource.MustParse(a.Limits[1].Value))
	assert.True(t, *b.Requests.Memory() == resource.MustParse(a.Requests[1].Value))
}

func TestFromEmptyResourcesToResourcesRequirements(t *testing.T) {
	a := FromResourcesToResourcesRequirements(v1alpha1.Resources{})
	assert.True(t, len(a.Requests) == 0)
}

func TestFromResourcesToResourcesRequirementsOnlyLimit(t *testing.T) {
	value := "1"
	a := FromResourcesToResourcesRequirements(v1alpha1.Resources{
		Limits: []v1alpha1.ResourceMap{
			{
				Resource: cpuResource,
				Value:    value,
			},
		},
	})
	assert.True(t, len(a.Requests) == 0)
	assert.True(t, len(a.Limits) == 1)
}

func TestFromResourcesToResourcesRequirementsHalfLimit(t *testing.T) {
	value := "500m"
	a := FromResourcesToResourcesRequirements(v1alpha1.Resources{
		Limits: []v1alpha1.ResourceMap{
			{
				Resource: cpuResource,
				Value:    value,
			},
		},
	})
	assert.True(t, len(a.Requests) == 0)
	assert.True(t, len(a.Limits) == 1)
	assert.Equal(t, "500m", a.Limits.Cpu().String())
}
