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

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
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
	result := EnvOverride(dst, src...)
	assert.Equal(t, 3, len(result))
	assert.Equal(t, result[0], dst[0])
	assert.Equal(t, result[1], dst[1])
	assert.Equal(t, result[2], src[1])
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
