package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

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
