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

package framework

import (
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"reflect"
	"testing"
)

func Test_EnvVarToMap(t *testing.T) {
	type args struct {
		env []corev1.EnvVar
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{"TestEnvVarToMap",
			args{
				[]corev1.EnvVar{
					{Name: "test1", Value: "test1"},
					{Name: "test2", Value: "test2"},
				}},
			map[string]string{
				"test1": "test1",
				"test2": "test2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EnvVarToMap(tt.args.env); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("envVarToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_MapToEnvVar(t *testing.T) {
	type args struct {
		env map[string]string
	}
	tests := []struct {
		name          string
		args          args
		wantEnvVars   []corev1.EnvVar
		orWantEnvVars []corev1.EnvVar
	}{
		{
			"TestMapToEnv",
			args{
				map[string]string{
					"test1": "test1",
					"test2": "test2",
				},
			},
			[]corev1.EnvVar{
				{Name: "test1", Value: "test1"},
				{Name: "test2", Value: "test2"},
			},
			[]corev1.EnvVar{
				{Name: "test2", Value: "test2"},
				{Name: "test1", Value: "test1"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotEnvVars := MapToEnvVar(tt.args.env); !reflect.DeepEqual(gotEnvVars, tt.wantEnvVars) && !reflect.DeepEqual(gotEnvVars, tt.orWantEnvVars) {
				t.Errorf("mapToEnv() = %v, want %v or %v", gotEnvVars, tt.wantEnvVars, tt.orWantEnvVars)
			}
		})
	}
}

func TestEnvVarArrayEquals(t *testing.T) {
	type args struct {
		array1 []corev1.EnvVar
		array2 []corev1.EnvVar
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"Equals",
			args{
				[]corev1.EnvVar{
					{Name: "test1", Value: "test1"},
					{Name: "test2", Value: "test2"},
					{Name: "test3", Value: "test3"},
				},
				[]corev1.EnvVar{
					{Name: "test3", Value: "test3"},
					{Name: "test1", Value: "test1"},
					{Name: "test2", Value: "test2"},
				},
			},
			true,
		},
		{
			"NotEqual",
			args{
				[]corev1.EnvVar{
					{Name: "test1", Value: "test2"},
					{Name: "test2", Value: "test1"},
					{Name: "test3", Value: "test3"},
				},
				[]corev1.EnvVar{
					{Name: "test3", Value: "test3"},
					{Name: "test1", Value: "test1"},
					{Name: "test2", Value: "test2"},
				},
			},
			false,
		},
		{
			"NotEqualLength",
			args{
				[]corev1.EnvVar{
					{Name: "test1", Value: "test1"},
					{Name: "test2", Value: "test2"},
					{Name: "test3", Value: "test3"},
					{Name: "test1", Value: "test1"},
				},
				[]corev1.EnvVar{
					{Name: "test3", Value: "test3"},
					{Name: "test1", Value: "test1"},
					{Name: "test2", Value: "test2"},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EnvVarArrayEquals(tt.args.array1, tt.args.array2); got != tt.want {
				t.Errorf("EnvVarArrayEquals() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
