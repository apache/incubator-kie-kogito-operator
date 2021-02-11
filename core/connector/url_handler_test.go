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

package connector

import (
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/client"
	"github.com/kiegroup/kogito-cloud-operator/core/operator"
	"github.com/kiegroup/kogito-cloud-operator/core/test"
	api2 "github.com/kiegroup/kogito-cloud-operator/core/test/api"
	"testing"

	"github.com/google/uuid"
	"github.com/kiegroup/kogito-cloud-operator/core/client/kubernetes"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestInjectDataIndexURLIntoKogitoRuntime(t *testing.T) {
	ns := t.Name()
	name := "my-kogito-app"
	expectedRoute := "http://dataindex-route.com"
	kogitoRuntime := &api2.KogitoRuntimeTest{
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
	dataIndex := &api2.KogitoSupportingServiceTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "data-index",
			Namespace: ns,
		},
		Spec: api2.KogitoSupportingServiceSpecTest{
			ServiceType: api.DataIndex,
		},
		Status: api2.KogitoSupportingServiceStatusTest{
			KogitoServiceStatus: api.KogitoServiceStatus{ExternalURI: expectedRoute},
		},
	}

	cli := test.NewFakeClientBuilder().AddK8sObjects(dc, kogitoRuntime, dataIndex).Build()
	runtimeHandler := test.CreateFakeKogitoRuntimeHandler(cli)
	supportingServiceHandler := test.CreateFakeKogitoSupportingServiceHandler(cli)

	context := &operator.Context{
		Client: cli,
		Log:    test.TestLogger,
		Scheme: test.GetRegisteredSchema(),
	}
	urlHandler := NewURLHandler(context, runtimeHandler, supportingServiceHandler)
	err := urlHandler.InjectDataIndexURLIntoKogitoRuntimeServices(ns)
	assert.NoError(t, err)

	exist, err := kubernetes.ResourceC(cli).Fetch(dc)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Contains(t, dc.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{Name: dataIndexHTTPRouteEnv, Value: expectedRoute})
}

func TestInjectJobsServicesURLIntoKogitoRuntime(t *testing.T) {
	URI := "http://localhost:8080"
	app := &api2.KogitoRuntimeTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kogito-app",
			Namespace: t.Name(),
			UID:       types.UID(uuid.New().String()),
		},
	}
	jobs := &api2.KogitoSupportingServiceTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "jobs-service",
			Namespace: t.Name(),
		},
		Spec: api2.KogitoSupportingServiceSpecTest{
			ServiceType: api.JobsService,
		},
		Status: api2.KogitoSupportingServiceStatusTest{
			KogitoServiceStatus: api.KogitoServiceStatus{
				ExternalURI: URI,
			},
		},
	}
	dc := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "dc", Namespace: t.Name(), OwnerReferences: []metav1.OwnerReference{{
			Name: app.Name,
			UID:  app.UID,
		}}},
		Spec: appsv1.DeploymentSpec{
			Template: v1.PodTemplateSpec{Spec: v1.PodSpec{Containers: []v1.Container{{Name: "the-app"}}}},
		},
	}
	cli := test.NewFakeClientBuilder().AddK8sObjects(app, dc, jobs).Build()
	runtimeHandler := test.CreateFakeKogitoRuntimeHandler(cli)
	supportingServiceHandler := test.CreateFakeKogitoSupportingServiceHandler(cli)
	context := &operator.Context{
		Client: cli,
		Log:    test.TestLogger,
		Scheme: test.GetRegisteredSchema(),
	}
	urlHandler := NewURLHandler(context, runtimeHandler, supportingServiceHandler)
	err := urlHandler.InjectJobsServicesURLIntoKogitoRuntimeServices(t.Name())
	assert.NoError(t, err)
	assert.Len(t, dc.Spec.Template.Spec.Containers[0].Env, 0)

	exists, err := kubernetes.ResourceC(cli).Fetch(dc)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Len(t, dc.Spec.Template.Spec.Containers[0].Env, 1)
	assert.Contains(t, dc.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{
		Name:  jobsServicesHTTPRouteEnv,
		Value: URI,
	})
}

func TestInjectJobsServicesURLIntoKogitoRuntimeCleanUp(t *testing.T) {
	URI := "http://localhost:8080"
	app := &api2.KogitoRuntimeTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kogito-app",
			Namespace: t.Name(),
			UID:       types.UID(uuid.New().String()),
		},
	}
	jobs := &api2.KogitoSupportingServiceTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "jobs-service",
			Namespace: t.Name(),
		},
		Spec: api2.KogitoSupportingServiceSpecTest{
			ServiceType: api.JobsService,
		},
		Status: api2.KogitoSupportingServiceStatusTest{KogitoServiceStatus: api.KogitoServiceStatus{ExternalURI: URI}},
	}
	dc := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "dc", Namespace: t.Name(), OwnerReferences: []metav1.OwnerReference{{
			Name: app.Name,
			UID:  app.UID,
		}}},
		Spec: appsv1.DeploymentSpec{
			Template: v1.PodTemplateSpec{Spec: v1.PodSpec{Containers: []v1.Container{{Name: "the-app"}}}},
		},
	}
	cli := test.NewFakeClientBuilder().AddK8sObjects(dc, app, jobs).Build()
	runtimeHandler := test.CreateFakeKogitoRuntimeHandler(cli)
	supportingServiceHandler := test.CreateFakeKogitoSupportingServiceHandler(cli)
	context := &operator.Context{
		Client: cli,
		Log:    test.TestLogger,
		Scheme: test.GetRegisteredSchema(),
	}
	urlHandler := NewURLHandler(context, runtimeHandler, supportingServiceHandler)
	// first we inject
	err := urlHandler.InjectJobsServicesURLIntoKogitoRuntimeServices(t.Name())
	assert.NoError(t, err)

	exists, err := kubernetes.ResourceC(cli).Fetch(dc)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Contains(t, dc.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{
		Name:  jobsServicesHTTPRouteEnv,
		Value: URI,
	})
}

func TestInjectTrustyURLIntoKogitoApps(t *testing.T) {
	ns := t.Name()
	name := "my-kogito-app"
	expectedRoute := "http://trusty-route.com"
	kogitoRuntime := &api2.KogitoRuntimeTest{
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
	trustyService := &api2.KogitoSupportingServiceTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "trusty",
			Namespace: ns,
		},
		Spec: api2.KogitoSupportingServiceSpecTest{
			ServiceType: api.TrustyAI,
		},
		Status: api2.KogitoSupportingServiceStatusTest{
			KogitoServiceStatus: api.KogitoServiceStatus{ExternalURI: expectedRoute},
		},
	}
	cli := test.NewFakeClientBuilder().AddK8sObjects(kogitoRuntime, dc, trustyService).Build()
	runtimeHandler := test.CreateFakeKogitoRuntimeHandler(cli)
	supportingServiceHandler := test.CreateFakeKogitoSupportingServiceHandler(cli)
	context := &operator.Context{
		Client: cli,
		Log:    test.TestLogger,
		Scheme: test.GetRegisteredSchema(),
	}
	urlHandler := NewURLHandler(context, runtimeHandler, supportingServiceHandler)
	err := urlHandler.InjectTrustyURLIntoKogitoRuntimeServices(ns)
	assert.NoError(t, err)

	exist, err := kubernetes.ResourceC(cli).Fetch(dc)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Contains(t, dc.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{Name: trustyHTTPRouteEnv, Value: expectedRoute})
}

func Test_getKogitoDataIndexURLs(t *testing.T) {
	ns := t.Name()
	hostname := "dataindex-route.com"
	expectedHTTPURL := "http://" + hostname
	expectedWSURL := "ws://" + hostname
	expectedHTTPSURL := "https://" + hostname
	expectedWSSURL := "wss://" + hostname
	insecureDI := &api2.KogitoSupportingServiceTest{
		Spec: api2.KogitoSupportingServiceSpecTest{
			ServiceType: api.DataIndex,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "data-index",
			Namespace: ns,
		},
		Status: api2.KogitoSupportingServiceStatusTest{
			KogitoServiceStatus: api.KogitoServiceStatus{
				ExternalURI: expectedHTTPURL,
			},
		},
	}
	secureDI := &api2.KogitoSupportingServiceTest{
		Spec: api2.KogitoSupportingServiceSpecTest{
			ServiceType: api.DataIndex,
		},
		ObjectMeta: metav1.ObjectMeta{Name: "data-index", Namespace: ns},
		Status:     api2.KogitoSupportingServiceStatusTest{KogitoServiceStatus: api.KogitoServiceStatus{ExternalURI: expectedHTTPSURL}},
	}

	cliInsecure := test.NewFakeClientBuilder().AddK8sObjects(insecureDI).Build()
	cliSecure := test.NewFakeClientBuilder().AddK8sObjects(secureDI).Build()
	inSecureSupportingServiceHandler := test.CreateFakeKogitoSupportingServiceHandler(cliInsecure)
	secureSupportingServiceHandler := test.CreateFakeKogitoSupportingServiceHandler(cliSecure)
	type args struct {
		client                   *client.Client
		namespace                string
		supportingServiceHandler api.KogitoSupportingServiceHandler
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
				client:                   cliInsecure,
				namespace:                ns,
				supportingServiceHandler: inSecureSupportingServiceHandler,
			},
			wantHTTPURL: expectedHTTPURL,
			wantWSURL:   expectedWSURL,
			wantErr:     false,
		},
		{
			name: "With secure route",
			args: args{
				client:                   cliSecure,
				namespace:                ns,
				supportingServiceHandler: secureSupportingServiceHandler,
			},
			wantHTTPURL: expectedHTTPSURL,
			wantWSURL:   expectedWSSURL,
			wantErr:     false,
		},
		{
			name: "With blank route",
			args: args{
				client:                   test.NewFakeClientBuilder().Build(),
				namespace:                ns,
				supportingServiceHandler: test.CreateFakeKogitoSupportingServiceHandler(test.NewFakeClientBuilder().Build()),
			},
			wantHTTPURL: "",
			wantWSURL:   "",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urlHandler := &urlHandler{
				Context: &operator.Context{
					Client: tt.args.client,
					Log:    test.TestLogger,
					Scheme: test.GetRegisteredSchema(),
				},
				runtimeHandler:           test.CreateFakeKogitoRuntimeHandler(tt.args.client),
				supportingServiceHandler: tt.args.supportingServiceHandler,
			}
			gotDataIndexEndpoints, err := urlHandler.getSupportingServiceEndpoints(tt.args.namespace, tt.wantHTTPURL, tt.wantWSURL, api.DataIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDataIndexEndpoints() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotDataIndexEndpoints != nil &&
				gotDataIndexEndpoints.HTTPRouteURI != tt.wantHTTPURL {
				t.Errorf("GetDataIndexEndpoints() gotHTTPURL = %v, want %v", gotDataIndexEndpoints.HTTPRouteURI, tt.wantHTTPURL)
			}
			if gotDataIndexEndpoints != nil &&
				gotDataIndexEndpoints.WSRouteURI != tt.wantWSURL {
				t.Errorf("GetDataIndexEndpoints() gotWSURL = %v, want %v", gotDataIndexEndpoints.WSRouteURI, tt.wantWSURL)
			}
		})
	}
}
