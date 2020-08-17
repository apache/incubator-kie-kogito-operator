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
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

func TestInjectDataIndexURLIntoKogitoRuntime(t *testing.T) {
	ns := t.Name()
	name := "my-kogito-app"
	expectedRoute := "http://dataindex-route.com"
	kogitoRuntime := &v1alpha1.KogitoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			UID:       types.UID(uuid.New().String()),
		},
	}
	dc := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: kogitoRuntime.Name, Namespace: kogitoRuntime.Namespace, OwnerReferences: []metav1.OwnerReference{{UID: kogitoRuntime.UID}}},
		Spec: appsv1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{Containers: []v1.Container{{Name: "test"}}},
			},
		},
	}
	dataIndexes := &v1alpha1.KogitoDataIndexList{
		Items: []v1alpha1.KogitoDataIndex{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      DefaultDataIndexName,
					Namespace: ns,
				},
				Status: v1alpha1.KogitoDataIndexStatus{
					KogitoServiceStatus: v1alpha1.KogitoServiceStatus{ExternalURI: expectedRoute},
				},
			},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{kogitoRuntime, dataIndexes, dc}, nil, nil)

	err := InjectDataIndexURLIntoKogitoRuntimeServices(cli, ns)
	assert.NoError(t, err)

	exist, err := kubernetes.ResourceC(cli).Fetch(dc)
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
	insecureDI := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{Name: DefaultDataIndexName, Namespace: ns},
		Status:     v1alpha1.KogitoDataIndexStatus{KogitoServiceStatus: v1alpha1.KogitoServiceStatus{ExternalURI: expectedHTTPURL}},
	}
	secureDI := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{Name: DefaultDataIndexName, Namespace: ns},
		Status:     v1alpha1.KogitoDataIndexStatus{KogitoServiceStatus: v1alpha1.KogitoServiceStatus{ExternalURI: expectedHTTPSURL}},
	}
	cliInsecure := test.CreateFakeClient([]runtime.Object{insecureDI}, nil, nil)
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
			name: "With insecure route",
			args: args{
				client:    cliInsecure,
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
			gotDataIndexEndpoints, err := GetDataIndexEndpoints(tt.args.client, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDataIndexEndpoints() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (len(gotDataIndexEndpoints.HTTPRouteEnv) > 0 ||
				len(gotDataIndexEndpoints.WSRouteEnv) > 0 ||
				len(gotDataIndexEndpoints.WSRouteURI) > 0) &&
				gotDataIndexEndpoints.HTTPRouteURI != tt.wantHTTPURL {
				t.Errorf("GetDataIndexEndpoints() gotHTTPURL = %v, want %v", gotDataIndexEndpoints.HTTPRouteURI, tt.wantHTTPURL)
			}
			if (len(gotDataIndexEndpoints.HTTPRouteEnv) > 0 ||
				len(gotDataIndexEndpoints.WSRouteEnv) > 0 ||
				len(gotDataIndexEndpoints.HTTPRouteURI) > 0) &&
				gotDataIndexEndpoints.WSRouteURI != tt.wantWSURL {
				t.Errorf("GetDataIndexEndpoints() gotWSURL = %v, want %v", gotDataIndexEndpoints.WSRouteURI, tt.wantWSURL)
			}
		})
	}
}
