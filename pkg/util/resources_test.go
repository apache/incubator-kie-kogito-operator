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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"reflect"
	"testing"
)

func Test_EnvToMap(t *testing.T) {
	type args struct {
		env []v1alpha1.Env
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{"TestEnvToMap",
			args{
				[]v1alpha1.Env{
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
			if got := EnvToMap(tt.args.env); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("envToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
