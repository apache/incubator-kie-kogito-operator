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
