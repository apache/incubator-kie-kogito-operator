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
