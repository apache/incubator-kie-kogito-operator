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
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/test"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"

	v1 "k8s.io/api/core/v1"
)

func TestEnsureProject(t *testing.T) {
	ns := t.Name()
	kubeCli := test.SetupFakeKubeCli(&v1.Namespace{
		ObjectMeta: v12.ObjectMeta{Name: ns},
	})
	type args struct {
		kubeCli *client.Client
		project string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"With project name",
			args{
				kubeCli: kubeCli,
				project: ns,
			},
			ns,
			false,
		},
		{
			"Without project name",
			args{
				kubeCli: kubeCli,
				project: "",
			},
			ns,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EnsureProject(tt.args.kubeCli, tt.args.project)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsureProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EnsureProject() got = %v, want %v", got, tt.want)
			}
		})
	}
}
