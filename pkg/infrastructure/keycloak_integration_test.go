// Copyright 2020 Red Hat, Inc. and/or its affiliates
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

package infrastructure

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	corev1 "k8s.io/api/core/v1"
	"testing"
	"time"
)

func TestDeployKeycloakWithKogitoInfra(t *testing.T) {
	type args struct {
		instance  v1alpha1.KeycloakAware
		namespace string
		cli       *client.Client
	}
	tests := []struct {
		name             string
		args             args
		wantUpdate       bool
		wantRequeueAfter time.Duration
		wantErr          bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUpdate, gotRequeueAfter, err := DeployKeycloakWithKogitoInfra(tt.args.instance, tt.args.namespace, tt.args.cli)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeployKeycloakWithKogitoInfra() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUpdate != tt.wantUpdate {
				t.Errorf("DeployKeycloakWithKogitoInfra() gotUpdate = %v, want %v", gotUpdate, tt.wantUpdate)
			}
			if gotRequeueAfter != tt.wantRequeueAfter {
				t.Errorf("DeployKeycloakWithKogitoInfra() gotRequeueAfter = %v, want %v", gotRequeueAfter, tt.wantRequeueAfter)
			}
		})
	}
}

func TestSetKeycloakVariables(t *testing.T) {
	type args struct {
		keycloakProps v1alpha1.KeycloakConnectionProperties
		clientID      string
		secret        string
		container     *corev1.Container
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}
