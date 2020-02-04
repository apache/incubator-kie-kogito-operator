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

package resource

import (
	keycloakv1alpha1 "github.com/keycloak/keycloak-operator/pkg/apis/keycloak/v1alpha1"
	"reflect"
	"testing"
)

func Test_newKeycloakClient(t *testing.T) {
	type args struct {
		clientName  string
		namespace   string
		realmLabels map[string]string
	}
	tests := []struct {
		name string
		args args
		want *keycloakv1alpha1.KeycloakClient
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newKeycloakClient(tt.args.namespace, tt.args.realmLabels); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newKeycloakClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newKeycloakUser(t *testing.T) {
	type args struct {
		userName    string
		namespace   string
		realmLabels map[string]string
	}
	tests := []struct {
		name string
		args args
		want *keycloakv1alpha1.KeycloakUser
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newKeycloakUser(tt.args.namespace, tt.args.realmLabels); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newKeycloakUser() = %v, want %v", got, tt.want)
			}
		})
	}
}
