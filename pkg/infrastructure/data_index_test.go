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

package infrastructure

import (
	"github.com/google/uuid"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	oappsv1 "github.com/openshift/api/apps/v1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

func Test_getKogitoDataIndexRoute(t *testing.T) {
	ns := t.Name()
	expectedRoute := "http://dataindex-route.com"
	dataIndexes := &v1alpha1.KogitoDataIndexList{
		Items: []v1alpha1.KogitoDataIndex{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kogito-data-index",
					Namespace: ns,
				},
				Status: v1alpha1.KogitoDataIndexStatus{
					KogitoServiceStatus: v1alpha1.KogitoServiceStatus{ExternalURI: expectedRoute},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kogito-data-index2",
					Namespace: ns,
				},
				Status: v1alpha1.KogitoDataIndexStatus{
					KogitoServiceStatus: v1alpha1.KogitoServiceStatus{ExternalURI: ""},
				},
			},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{dataIndexes}, nil, nil)

	route, err := getKogitoDataIndexRoute(cli, ns)
	assert.NoError(t, err)
	assert.Equal(t, expectedRoute, route)
}

func Test_getKogitoDataIndexRoute_NoDataIndex(t *testing.T) {
	cli := test.CreateFakeClient(nil, nil, nil)
	route, err := getKogitoDataIndexRoute(cli, t.Name())
	assert.NoError(t, err)
	assert.Empty(t, route)
}

func TestInjectDataIndexURLIntoKogitoApps(t *testing.T) {
	ns := t.Name()
	name := "my-kogito-app"
	expectedRoute := "http://dataindex-route.com"
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			UID:       types.UID(uuid.New().String()),
		},
	}
	dc := &oappsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{Name: kogitoApp.Name, Namespace: kogitoApp.Namespace, OwnerReferences: []metav1.OwnerReference{{UID: kogitoApp.UID}}},
		Spec: oappsv1.DeploymentConfigSpec{
			Template: &v1.PodTemplateSpec{
				Spec: v1.PodSpec{Containers: []v1.Container{{Name: "test"}}},
			},
		},
	}
	dataIndexes := &v1alpha1.KogitoDataIndexList{
		Items: []v1alpha1.KogitoDataIndex{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kogito-data-index",
					Namespace: ns,
				},
				Status: v1alpha1.KogitoDataIndexStatus{
					KogitoServiceStatus: v1alpha1.KogitoServiceStatus{ExternalURI: expectedRoute},
				},
			},
		},
	}
	client := test.CreateFakeClient([]runtime.Object{kogitoApp, dataIndexes, dc}, nil, nil)

	err := InjectDataIndexURLIntoKogitoApps(client, ns)
	assert.NoError(t, err)

	exist, err := kubernetes.ResourceC(client).Fetch(dc)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Contains(t, dc.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{Name: dataIndexHTTPRouteEnv, Value: expectedRoute})
}

func Test_getKogitoDataIndexURLs(t *testing.T) {
	ns := t.Name()
	hostname := "dataindex-route.com"
	expectedHTTPURL := "http://" + hostname
	expectedWSURL := "ws://" + hostname
	expectedHTTPSURL := "https://" + hostname
	expectedWSSURL := "wss://" + hostname
	unsecureDI := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{Name: "kogito-data-index", Namespace: ns},
		Status:     v1alpha1.KogitoDataIndexStatus{KogitoServiceStatus: v1alpha1.KogitoServiceStatus{ExternalURI: expectedHTTPURL}},
	}
	secureDI := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{Name: "kogito-data-index", Namespace: ns},
		Status:     v1alpha1.KogitoDataIndexStatus{KogitoServiceStatus: v1alpha1.KogitoServiceStatus{ExternalURI: expectedHTTPSURL}},
	}
	cliUnsecure := test.CreateFakeClient([]runtime.Object{unsecureDI}, nil, nil)
	cliSecure := test.CreateFakeClient([]runtime.Object{secureDI}, nil, nil)
	type args struct {
		client    *client.Client
		namespace string
	}
	tests := []struct {
		name        string
		args        args
		wantHTTPURL string
		wantWSURL   string
		wantErr     bool
	}{
		{
			name: "With unsecure route",
			args: args{
				client:    cliUnsecure,
				namespace: ns,
			},
			wantHTTPURL: expectedHTTPURL,
			wantWSURL:   expectedWSURL,
			wantErr:     false,
		},
		{
			name: "With secure route",
			args: args{
				client:    cliSecure,
				namespace: ns,
			},
			wantHTTPURL: expectedHTTPSURL,
			wantWSURL:   expectedWSSURL,
			wantErr:     false,
		},
		{
			name: "With blank route",
			args: args{
				client:    test.CreateFakeClient(nil, nil, nil),
				namespace: ns,
			},
			wantHTTPURL: "",
			wantWSURL:   "",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHTTPURL, gotWSURL, err := getKogitoDataIndexURLs(tt.args.client, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("getKogitoDataIndexURLs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotHTTPURL != tt.wantHTTPURL {
				t.Errorf("getKogitoDataIndexURLs() gotHTTPURL = %v, want %v", gotHTTPURL, tt.wantHTTPURL)
			}
			if gotWSURL != tt.wantWSURL {
				t.Errorf("getKogitoDataIndexURLs() gotWSURL = %v, want %v", gotWSURL, tt.wantWSURL)
			}
		})
	}
}
