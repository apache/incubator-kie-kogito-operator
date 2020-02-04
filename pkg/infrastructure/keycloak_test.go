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
	keycloakv1alpha1 "github.com/keycloak/keycloak-operator/pkg/apis/keycloak/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"reflect"
	"testing"
)

func TestGetKeycloakClientInstance(t *testing.T) {
	type args struct {
		name      string
		namespace string
		client    *client.Client
	}
	tests := []struct {
		name    string
		args    args
		want    *keycloakv1alpha1.KeycloakClient
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetKeycloakClientInstance(tt.args.name, tt.args.namespace, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetKeycloakClientInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetKeycloakClientInstance() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetKeycloakInstance(t *testing.T) {
	type args struct {
		name      string
		namespace string
		client    *client.Client
	}
	tests := []struct {
		name    string
		args    args
		want    *keycloakv1alpha1.Keycloak
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetKeycloakInstance(tt.args.name, tt.args.namespace, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetKeycloakInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetKeycloakInstance() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetKeycloakProperties(t *testing.T) {
	type args struct {
		cli   *client.Client
		infra *v1alpha1.KogitoInfra
	}
	tests := []struct {
		name     string
		args     args
		wantName string
		wantURL  string
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotURL, err := GetKeycloakProperties(tt.args.cli, tt.args.infra)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetKeycloakProperties() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotName != tt.wantName {
				t.Errorf("GetKeycloakProperties() gotName = %v, want %v", gotName, tt.wantName)
			}
			if gotURL != tt.wantURL {
				t.Errorf("GetKeycloakProperties() gotURL = %v, want %v", gotURL, tt.wantURL)
			}
		})
	}
}

func TestGetKeycloakRealmInstance(t *testing.T) {
	type args struct {
		name      string
		namespace string
		client    *client.Client
	}
	tests := []struct {
		name    string
		args    args
		want    *keycloakv1alpha1.KeycloakRealm
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetKeycloakRealmInstance(tt.args.name, tt.args.namespace, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetKeycloakRealmInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetKeycloakRealmInstance() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetKeycloakRealmProperties(t *testing.T) {
	type args struct {
		cli   *client.Client
		infra *v1alpha1.KogitoInfra
	}
	tests := []struct {
		name          string
		args          args
		wantName      string
		wantRealmName string
		wantLabels    map[string]string
		wantErr       bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotRealmName, gotLabels, err := GetKeycloakRealmProperties(tt.args.cli, tt.args.infra)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetKeycloakRealmProperties() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotName != tt.wantName {
				t.Errorf("GetKeycloakRealmProperties() gotName = %v, want %v", gotName, tt.wantName)
			}
			if gotRealmName != tt.wantRealmName {
				t.Errorf("GetKeycloakRealmProperties() gotRealmName = %v, want %v", gotRealmName, tt.wantRealmName)
			}
			if !reflect.DeepEqual(gotLabels, tt.wantLabels) {
				t.Errorf("GetKeycloakRealmProperties() gotLabels = %v, want %v", gotLabels, tt.wantLabels)
			}
		})
	}
}

func TestGetKeycloakUserInstance(t *testing.T) {
	type args struct {
		name      string
		namespace string
		client    *client.Client
	}
	tests := []struct {
		name    string
		args    args
		want    *keycloakv1alpha1.KeycloakUser
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetKeycloakUserInstance(tt.args.name, tt.args.namespace, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetKeycloakUserInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetKeycloakUserInstance() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsKeycloakAvailable(t *testing.T) {
	type args struct {
		client *client.Client
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsKeycloakAvailable(tt.args.client); got != tt.want {
				t.Errorf("IsKeycloakAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}
