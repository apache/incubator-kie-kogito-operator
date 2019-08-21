package shared

import (
	"testing"

	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
)

var (
	cpuResource    = v1alpha1.ResourceCPU
	memoryResource = v1alpha1.ResourceMemory
)

func TestEnvOverride(t *testing.T) {
	src := []corev1.EnvVar{
		{
			Name:  "test1",
			Value: "value1",
		},
		{
			Name:  "test2",
			Value: "value2",
		},
	}
	dst := []corev1.EnvVar{
		{
			Name:  "test1",
			Value: "valueX",
		},
		{
			Name:  "test3",
			Value: "value3",
		},
	}
	result := EnvOverride(src, dst)
	assert.Equal(t, 3, len(result))
	assert.Equal(t, result[0], dst[0])
	assert.Equal(t, result[1], src[1])
	assert.Equal(t, result[2], dst[1])
}

func TestGetEnvVar(t *testing.T) {
	vars := []corev1.EnvVar{
		{
			Name:  "test1",
			Value: "value1",
		},
		{
			Name:  "test2",
			Value: "value2",
		},
	}
	pos := GetEnvVar("test1", vars)
	assert.Equal(t, 0, pos)

	pos = GetEnvVar("other", vars)
	assert.Equal(t, -1, pos)
}

func TestEnvVarCheck(t *testing.T) {
	empty := []corev1.EnvVar{}
	a := []corev1.EnvVar{
		{Name: "A", Value: "1"},
	}
	b := []corev1.EnvVar{
		{Name: "A", Value: "2"},
	}
	c := []corev1.EnvVar{
		{Name: "A", Value: "1"},
		{Name: "B", Value: "1"},
	}

	assert.True(t, EnvVarCheck(empty, empty))
	assert.True(t, EnvVarCheck(a, a))

	assert.False(t, EnvVarCheck(empty, a))
	assert.False(t, EnvVarCheck(a, empty))

	assert.False(t, EnvVarCheck(a, b))
	assert.False(t, EnvVarCheck(b, a))

	assert.False(t, EnvVarCheck(a, c))
	assert.False(t, EnvVarCheck(c, b))
}

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
