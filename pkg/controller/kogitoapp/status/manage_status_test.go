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

package status

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
	"time"

	"encoding/json"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
	obuildv1 "github.com/openshift/api/build/v1"
	dockerv10 "github.com/openshift/api/image/docker10"
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	buildfake "github.com/openshift/client-go/build/clientset/versioned/fake"
	imgfake "github.com/openshift/client-go/image/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_createResult(t *testing.T) {
	type args struct {
		updated      bool
		requeueAfter time.Duration
		err          error
		result       *UpdateStatusResult
	}
	tests := []struct {
		name   string
		args   args
		result *UpdateStatusResult
	}{
		{
			"Update",
			args{
				true,
				755,
				fmt.Errorf("test"),
				&UpdateStatusResult{},
			},
			&UpdateStatusResult{
				true,
				755,
				fmt.Errorf("test"),
			},
		},
		{
			"NoErrorUpdate",
			args{
				false,
				755,
				fmt.Errorf("test"),
				&UpdateStatusResult{
					Err: fmt.Errorf("test1"),
				},
			},
			&UpdateStatusResult{
				false,
				755,
				fmt.Errorf("test1"),
			},
		},
		{
			"NoRequeueUpdate",
			args{
				false,
				755,
				fmt.Errorf("test"),
				&UpdateStatusResult{
					Updated:      true,
					RequeueAfter: 777,
				},
			},
			&UpdateStatusResult{
				true,
				777,
				fmt.Errorf("test"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createResult(tt.args.updated, tt.args.requeueAfter, tt.args.err, tt.args.result)
			if !reflect.DeepEqual(tt.args.result, tt.result) {
				t.Errorf("createResult() result = %v, wantResult %v", tt.args.result, tt.result)
				return
			}
		})
	}
}

func Test_updateObj(t *testing.T) {
	s := scheme.Scheme
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.KogitoApp{})

	type args struct {
		obj    meta.ResourceObject
		client *client.Client
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			"UpdateObjError",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(),
				},
			},
			false,
			true,
		},
		{
			"UpdateObj",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&v1alpha1.KogitoApp{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test",
							},
						},
					),
				},
			},
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := updateObj(tt.args.obj, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("updateObj() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("updateObj() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_statusUpdateForResources(t *testing.T) {
	type args struct {
		instance *v1alpha1.KogitoApp
		result   *UpdateResourcesResult
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			"Error",
			args{
				&v1alpha1.KogitoApp{},
				&UpdateResourcesResult{
					Updated:     false,
					Err:         fmt.Errorf("test"),
					ErrorReason: v1alpha1.BuildRuntimeFailedReason,
				},
			},
			true,
			true,
		},
		{
			"Update",
			args{
				&v1alpha1.KogitoApp{},
				&UpdateResourcesResult{
					Updated:     true,
					Err:         nil,
					ErrorReason: v1alpha1.ReasonType(""),
				},
			},
			true,
			false,
		},
		{
			"Nothing",
			args{
				&v1alpha1.KogitoApp{},
				&UpdateResourcesResult{
					Updated:     false,
					Err:         nil,
					ErrorReason: v1alpha1.ReasonType(""),
				},
			},
			false,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := statusUpdateForResources(tt.args.instance, tt.args.result)
			if (err != nil) != tt.wantErr {
				t.Errorf("statusUpdateForResources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("statusUpdateForResources() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_statusUpdateForDeployment(t *testing.T) {
	s := scheme.Scheme
	s.AddKnownTypes(appsv1.GroupVersion, &appsv1.DeploymentConfig{})

	type args struct {
		instance  *v1alpha1.KogitoApp
		resources *resource.KogitoAppResources
		client    *client.Client
	}
	tests := []struct {
		name           string
		args           args
		want           bool
		wantErr        bool
		wantDeployment v1alpha1.Deployments
	}{
		{
			"Stopped",
			args{
				&v1alpha1.KogitoApp{},
				&resource.KogitoAppResources{
					DeploymentConfig: &appsv1.DeploymentConfig{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "test",
							Name:      "test",
						},
					},
				},
				&client.Client{ControlCli: fake.NewFakeClient(
					&appsv1.DeploymentConfig{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "test",
							Name:      "test",
						},
						Spec: appsv1.DeploymentConfigSpec{
							Replicas: 0,
						},
						Status: appsv1.DeploymentConfigStatus{
							Replicas: 0,
						},
					},
				),
				},
			},
			true,
			false,
			v1alpha1.Deployments{
				Stopped: []string{"test"},
			},
		},
		{
			"Starting",
			args{
				&v1alpha1.KogitoApp{},
				&resource.KogitoAppResources{
					DeploymentConfig: &appsv1.DeploymentConfig{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "test",
							Name:      "test",
						},
					},
				},
				&client.Client{ControlCli: fake.NewFakeClient(
					&appsv1.DeploymentConfig{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "test",
							Name:      "test",
						},
						Spec: appsv1.DeploymentConfigSpec{
							Replicas: 1,
						},
						Status: appsv1.DeploymentConfigStatus{
							Replicas:      1,
							ReadyReplicas: 0,
						},
					},
				),
				},
			},
			true,
			false,
			v1alpha1.Deployments{
				Starting: []string{"test"},
			},
		},
		{
			"Ready",
			args{
				&v1alpha1.KogitoApp{},
				&resource.KogitoAppResources{
					DeploymentConfig: &appsv1.DeploymentConfig{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "test",
							Name:      "test",
						},
					},
				},
				&client.Client{ControlCli: fake.NewFakeClient(
					&appsv1.DeploymentConfig{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "test",
							Name:      "test",
						},
						Spec: appsv1.DeploymentConfigSpec{
							Replicas: 1,
						},
						Status: appsv1.DeploymentConfigStatus{
							Replicas:      1,
							ReadyReplicas: 1,
						},
					},
				),
				},
			},
			true,
			false,
			v1alpha1.Deployments{
				Ready: []string{"test"},
			},
		},
		{
			"NoUpdate",
			args{
				&v1alpha1.KogitoApp{
					Status: v1alpha1.KogitoAppStatus{
						Deployments: v1alpha1.Deployments{
							Ready: []string{"test"},
						},
					},
				},
				&resource.KogitoAppResources{
					DeploymentConfig: &appsv1.DeploymentConfig{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "test",
							Name:      "test",
						},
					},
				},
				&client.Client{ControlCli: fake.NewFakeClient(
					&appsv1.DeploymentConfig{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "test",
							Name:      "test",
						},
						Spec: appsv1.DeploymentConfigSpec{
							Replicas: 1,
						},
						Status: appsv1.DeploymentConfigStatus{
							Replicas:      1,
							ReadyReplicas: 1,
						},
					},
				),
				},
			},
			false,
			false,
			v1alpha1.Deployments{
				Ready: []string{"test"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := statusUpdateForDeployment(tt.args.instance, tt.args.resources, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("statusUpdateForDeployment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("statusUpdateForDeployment() got = %v, want %v", got, tt.want)
				return
			}
			if !reflect.DeepEqual(tt.wantDeployment, tt.args.instance.Status.Deployments) {
				t.Errorf("statusUpdateForDeployment() got = %v, want %v", tt.args.instance.Status.Deployments, tt.wantDeployment)
			}
		})
	}
}

func Test_statusUpdateForRoute(t *testing.T) {
	s := scheme.Scheme
	s.AddKnownTypes(routev1.GroupVersion, &routev1.Route{})

	type args struct {
		instance  *v1alpha1.KogitoApp
		resources *resource.KogitoAppResources
		client    *client.Client
	}
	tests := []struct {
		name        string
		args        args
		wantRequeue time.Duration
		wantUpdated bool
		wantErr     bool
	}{
		{
			"Requeue",
			args{
				&v1alpha1.KogitoApp{
					Status: v1alpha1.KogitoAppStatus{
						Route: "test",
					},
				},
				&resource.KogitoAppResources{
					Route: &routev1.Route{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "test",
							Name:      "test",
						},
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(),
				},
			},
			time.Duration(500) * time.Millisecond,
			false,
			false,
		},
		{
			"UpdateRoute",
			args{
				&v1alpha1.KogitoApp{
					Status: v1alpha1.KogitoAppStatus{
						Route: "test",
					},
				},
				&resource.KogitoAppResources{
					Route: &routev1.Route{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "test",
							Name:      "test",
						},
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&routev1.Route{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "test",
								Name:      "test",
							},
							Spec: routev1.RouteSpec{
								Host: "test",
							},
						},
					),
				},
			},
			0,
			true,
			false,
		},
		{
			"NoUpdateRoute",
			args{
				&v1alpha1.KogitoApp{
					Status: v1alpha1.KogitoAppStatus{
						Route: "http://test",
					},
				},
				&resource.KogitoAppResources{
					Route: &routev1.Route{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "test",
							Name:      "test",
						},
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&routev1.Route{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "test",
								Name:      "test",
							},
							Spec: routev1.RouteSpec{
								Host: "test",
							},
						},
					),
				},
			},
			0,
			false,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRequeue, gotUpdated, err := statusUpdateForRoute(tt.args.instance, tt.args.resources, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("statusUpdateForRoute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRequeue != tt.wantRequeue {
				t.Errorf("statusUpdateForRoute() gotRequeue = %v, want %v", gotRequeue, tt.wantRequeue)
			}
			if gotUpdated != tt.wantUpdated {
				t.Errorf("statusUpdateForRoute() gotUpdated = %v, want %v", gotUpdated, tt.wantUpdated)
			}
		})
	}
}

func Test_statusUpdateForImageBuild(t *testing.T) {
	s := scheme.Scheme
	s.AddKnownTypes(obuildv1.GroupVersion, &obuildv1.BuildConfig{}, &obuildv1.Build{})
	s.AddKnownTypes(imgv1.GroupVersion, &imgv1.ImageStreamTag{})

	dockerImageRaw, _ := json.Marshal(&dockerv10.DockerImage{
		Config: &dockerv10.DockerConfig{
			Labels: map[string]string{
				openshift.ImageLabelForExposeServices: "8080:http",
			},
		},
	})

	s2iBuildConfig := &obuildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "tests2i",
			Labels: map[string]string{
				"app":       "test",
				"buildtype": "s2i",
			},
		},
	}

	runtimeBuildConfig := &obuildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "testruntime",
			Labels: map[string]string{
				"app":       "test",
				"buildtype": "runtime",
			},
		},
	}

	s2iImage := &imgv1.ImageStreamTag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tests2i:latest",
			Namespace: "test",
		},
		Image: imgv1.Image{
			DockerImageMetadata: runtime.RawExtension{
				Raw: dockerImageRaw,
			},
		},
	}

	runtimeImage := &imgv1.ImageStreamTag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testruntime:latest",
			Namespace: "test",
		},
		Image: imgv1.Image{
			DockerImageMetadata: runtime.RawExtension{
				Raw: dockerImageRaw,
			},
		},
	}

	s2iBuild := &obuildv1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "tests2ibuild",
			Labels: map[string]string{
				"app":       "test",
				"buildtype": "s2i",
			},
		},
		Status: obuildv1.BuildStatus{
			Phase: obuildv1.BuildPhaseNew,
		},
	}

	runtimeBuild := &obuildv1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "testruntimebuild",
			Labels: map[string]string{
				"app":       "test",
				"buildtype": "runtime",
			},
		},
		Status: obuildv1.BuildStatus{
			Phase: obuildv1.BuildPhaseRunning,
		},
	}

	s2iBuildFail := &obuildv1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "tests2ibuildfail",
			Labels: map[string]string{
				"app":       "test",
				"buildtype": "s2i",
			},
		},
		Status: obuildv1.BuildStatus{
			Phase: obuildv1.BuildPhaseFailed,
		},
	}

	runtimeBuildFail := &obuildv1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "testruntimebuildfail",
			Labels: map[string]string{
				"app":       "test",
				"buildtype": "runtime",
			},
		},
		Status: obuildv1.BuildStatus{
			Phase: obuildv1.BuildPhaseError,
		},
	}

	s2iBuildComplete := &obuildv1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "tests2ibuildcomplete",
			Labels: map[string]string{
				"app":       "test",
				"buildtype": "s2i",
			},
		},
		Status: obuildv1.BuildStatus{
			Phase: obuildv1.BuildPhaseComplete,
		},
	}

	runtimeBuildComplete := &obuildv1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "testruntimebuildcomplete",
			Labels: map[string]string{
				"app":       "test",
				"buildtype": "runtime",
			},
		},
		Status: obuildv1.BuildStatus{
			Phase: obuildv1.BuildPhaseComplete,
		},
	}

	type args struct {
		instance  *v1alpha1.KogitoApp
		resources *resource.KogitoAppResources
		client    *client.Client
	}
	tests := []struct {
		name        string
		args        args
		wantRequeue time.Duration
		wantUpdated bool
		wantErr     bool
	}{
		{
			"Running",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Status: v1alpha1.KogitoAppStatus{
						Builds: v1alpha1.Builds{
							New:     []string{"tests2ibuild"},
							Running: []string{"testruntimebuild"},
						},
					},
				},
				&resource.KogitoAppResources{
					BuildConfigS2I: &obuildv1.BuildConfig{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "test",
							Name:      "tests2i",
							Labels: map[string]string{
								"app":       "test",
								"buildtype": "s2i",
							},
						},
					},
					BuildConfigRuntime: &obuildv1.BuildConfig{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "test",
							Name:      "testruntime",
							Labels: map[string]string{
								"app":       "test",
								"buildtype": "runtime",
							},
						},
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(),
					ImageCli:   imgfake.NewSimpleClientset(s2iImage, runtimeImage).ImageV1(),
					BuildCli:   buildfake.NewSimpleClientset(s2iBuildConfig, runtimeBuildConfig, s2iBuild, runtimeBuild).BuildV1(),
				},
			},
			time.Duration(50) * time.Second,
			true,
			false,
		},
		{
			"Fail",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
				},
				&resource.KogitoAppResources{
					BuildConfigS2I: &obuildv1.BuildConfig{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "test",
							Name:      "tests2i",
							Labels: map[string]string{
								"app":       "test",
								"buildtype": "s2i",
							},
						},
					},
					BuildConfigRuntime: &obuildv1.BuildConfig{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "test",
							Name:      "testruntime",
							Labels: map[string]string{
								"app":       "test",
								"buildtype": "runtime",
							},
						},
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(),
					ImageCli:   imgfake.NewSimpleClientset(s2iImage, runtimeImage).ImageV1(),
					BuildCli:   buildfake.NewSimpleClientset(s2iBuildConfig, runtimeBuildConfig, s2iBuildFail, runtimeBuildFail).BuildV1(),
				},
			},
			0,
			true,
			false,
		},
		{
			"NoImage",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Status: v1alpha1.KogitoAppStatus{
						Builds: v1alpha1.Builds{
							Complete: []string{"tests2ibuildcomplete", "testruntimebuildcomplete"},
						},
					},
				},
				&resource.KogitoAppResources{
					BuildConfigS2I: &obuildv1.BuildConfig{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "test",
							Name:      "tests2i",
							Labels: map[string]string{
								"app":       "test",
								"buildtype": "s2i",
							},
						},
					},
					BuildConfigRuntime: &obuildv1.BuildConfig{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "test",
							Name:      "testruntime",
							Labels: map[string]string{
								"app":       "test",
								"buildtype": "runtime",
							},
						},
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(),
					ImageCli:   imgfake.NewSimpleClientset().ImageV1(),
					BuildCli:   buildfake.NewSimpleClientset(s2iBuildConfig, runtimeBuildConfig, s2iBuildComplete, runtimeBuildComplete).BuildV1(),
				},
			},
			time.Duration(50) * time.Second,
			false,
			false,
		},
		{
			"Update",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
				},
				&resource.KogitoAppResources{
					BuildConfigS2I: &obuildv1.BuildConfig{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "test",
							Name:      "tests2i",
							Labels: map[string]string{
								"app":       "test",
								"buildtype": "s2i",
							},
						},
					},
					BuildConfigRuntime: &obuildv1.BuildConfig{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "test",
							Name:      "testruntime",
							Labels: map[string]string{
								"app":       "test",
								"buildtype": "runtime",
							},
						},
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(),
					ImageCli:   imgfake.NewSimpleClientset(s2iImage, runtimeImage).ImageV1(),
					BuildCli:   buildfake.NewSimpleClientset(s2iBuildConfig, runtimeBuildConfig, s2iBuildComplete, runtimeBuildComplete).BuildV1(),
				},
			},
			0,
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRequeue, gotUpdated, err := statusUpdateForImageBuild(tt.args.instance, tt.args.resources, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("statusUpdateForImageBuild() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRequeue != tt.wantRequeue {
				t.Errorf("statusUpdateForImageBuild() gotRequeue = %v, want %v", gotRequeue, tt.wantRequeue)
			}
			if gotUpdated != tt.wantUpdated {
				t.Errorf("statusUpdateForImageBuild() gotUpdated = %v, want %v", gotUpdated, tt.wantUpdated)
			}
		})
	}
}

func Test_ensureApplicationImageExists(t *testing.T) {
	s := scheme.Scheme
	s.AddKnownTypes(obuildv1.GroupVersion, &obuildv1.BuildConfig{}, &obuildv1.Build{})
	s.AddKnownTypes(imgv1.GroupVersion, &imgv1.ImageStreamTag{})

	dockerImageRaw, _ := json.Marshal(&dockerv10.DockerImage{
		Config: &dockerv10.DockerConfig{
			Labels: map[string]string{
				openshift.ImageLabelForExposeServices: "8080:http",
			},
		},
	})

	s2iBuildConfig := &obuildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "tests2i",
			Labels: map[string]string{
				"app":       "test",
				"buildtype": "s2i",
			},
		},
	}

	runtimeBuildConfig := &obuildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "testruntime",
			Labels: map[string]string{
				"app":       "test",
				"buildtype": "runtime",
			},
		},
	}

	s2iImage := &imgv1.ImageStreamTag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tests2i:latest",
			Namespace: "test",
		},
		Image: imgv1.Image{
			DockerImageMetadata: runtime.RawExtension{
				Raw: dockerImageRaw,
			},
		},
	}

	runtimeImage := &imgv1.ImageStreamTag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testruntime:latest",
			Namespace: "test",
		},
		Image: imgv1.Image{
			DockerImageMetadata: runtime.RawExtension{
				Raw: dockerImageRaw,
			},
		},
	}

	s2iBuild := &obuildv1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "tests2ibuild",
			Labels: map[string]string{
				"app":       "test",
				"buildtype": "s2i",
			},
		},
		Status: obuildv1.BuildStatus{
			Phase: obuildv1.BuildPhaseNew,
		},
	}

	runtimeBuild := &obuildv1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "testruntimebuild",
			Labels: map[string]string{
				"app":       "test",
				"buildtype": "runtime",
			},
		},
		Status: obuildv1.BuildStatus{
			Phase: obuildv1.BuildPhaseRunning,
		},
	}

	s2iBuildFail := &obuildv1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "tests2ibuildfail",
			Labels: map[string]string{
				"app":       "test",
				"buildtype": "s2i",
			},
		},
		Status: obuildv1.BuildStatus{
			Phase: obuildv1.BuildPhaseFailed,
		},
	}

	runtimeBuildFail := &obuildv1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "testruntimebuildfail",
			Labels: map[string]string{
				"app":       "test",
				"buildtype": "runtime",
			},
		},
		Status: obuildv1.BuildStatus{
			Phase: obuildv1.BuildPhaseError,
		},
	}
	type args struct {
		instance  *v1alpha1.KogitoApp
		resources *resource.KogitoAppResources
		client    *client.Client
	}
	tests := []struct {
		name              string
		args              args
		wantExists        bool
		wantRunning       bool
		wantUpdated       bool
		wantRuntimeFailed bool
		wantS2iFailed     bool
		wantErr           bool
	}{
		{
			"ImageExists",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
				},
				&resource.KogitoAppResources{
					BuildConfigS2I: &obuildv1.BuildConfig{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "test",
							Name:      "tests2i",
							Labels: map[string]string{
								"app":       "test",
								"buildtype": "s2i",
							},
						},
					},
					BuildConfigRuntime: &obuildv1.BuildConfig{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "test",
							Name:      "testruntime",
							Labels: map[string]string{
								"app":       "test",
								"buildtype": "runtime",
							},
						},
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(),
					ImageCli:   imgfake.NewSimpleClientset(s2iImage, runtimeImage).ImageV1(),
					BuildCli: buildfake.NewSimpleClientset(s2iBuildConfig, runtimeBuildConfig,
						s2iBuild, runtimeBuild, s2iBuildFail, runtimeBuildFail).BuildV1(),
				},
			},
			true,
			true,
			true,
			true,
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotExists, gotRunning, gotUpdated, gotRuntimeFailed, gotS2iFailed, err := ensureApplicationImageExists(tt.args.instance, tt.args.resources, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("ensureApplicationImageExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotExists != tt.wantExists {
				t.Errorf("ensureApplicationImageExists() gotExists = %v, want %v", gotExists, tt.wantExists)
			}
			if gotRunning != tt.wantRunning {
				t.Errorf("ensureApplicationImageExists() gotRunning = %v, want %v", gotRunning, tt.wantRunning)
			}
			if gotUpdated != tt.wantUpdated {
				t.Errorf("ensureApplicationImageExists() gotUpdated = %v, want %v", gotUpdated, tt.wantUpdated)
			}
			if gotRuntimeFailed != tt.wantRuntimeFailed {
				t.Errorf("ensureApplicationImageExists() gotRuntimeFailed = %v, want %v", gotRuntimeFailed, tt.wantRuntimeFailed)
			}
			if gotS2iFailed != tt.wantS2iFailed {
				t.Errorf("ensureApplicationImageExists() gotS2iFailed = %v, want %v", gotS2iFailed, tt.wantS2iFailed)
			}
		})
	}
}

func Test_checkBuildsStatus(t *testing.T) {
	type args struct {
		state     *v1alpha1.Builds
		lastState *v1alpha1.Builds
	}
	tests := []struct {
		name          string
		args          args
		wantUpdated   bool
		wantRunning   bool
		wantNewFailed bool
	}{
		{
			"Running",
			args{
				&v1alpha1.Builds{
					New:       []string{"build1"},
					Pending:   []string{"build2"},
					Running:   []string{"build3"},
					Complete:  nil,
					Failed:    nil,
					Error:     nil,
					Cancelled: nil,
				},
				&v1alpha1.Builds{
					New:       []string{"build1"},
					Pending:   []string{"build2"},
					Running:   []string{"build3"},
					Complete:  nil,
					Failed:    nil,
					Error:     nil,
					Cancelled: nil,
				},
			},
			false,
			true,
			false,
		},
		{
			"Updating",
			args{
				&v1alpha1.Builds{
					New:       nil,
					Pending:   nil,
					Running:   nil,
					Complete:  []string{"build1", "build2"},
					Failed:    nil,
					Error:     nil,
					Cancelled: nil,
				},
				&v1alpha1.Builds{
					New:       nil,
					Pending:   nil,
					Running:   nil,
					Complete:  []string{"build1"},
					Failed:    nil,
					Error:     nil,
					Cancelled: nil,
				},
			},
			true,
			false,
			false,
		},
		{
			"Error",
			args{
				&v1alpha1.Builds{
					New:       nil,
					Pending:   nil,
					Running:   nil,
					Complete:  nil,
					Failed:    []string{"build1", "build4"},
					Error:     []string{"build2", "build5"},
					Cancelled: []string{"build3", "build6"},
				},
				&v1alpha1.Builds{
					New:       nil,
					Pending:   nil,
					Running:   nil,
					Complete:  nil,
					Failed:    []string{"build1"},
					Error:     []string{"build2"},
					Cancelled: []string{"build3"},
				},
			},
			true,
			false,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUpdated, gotRunning, gotNewFailed := checkBuildsStatus(tt.args.state, tt.args.lastState)
			if gotUpdated != tt.wantUpdated {
				t.Errorf("checkBuildsStatus() gotUpdated = %v, want %v", gotUpdated, tt.wantUpdated)
			}
			if gotRunning != tt.wantRunning {
				t.Errorf("checkBuildsStatus() gotRunning = %v, want %v", gotRunning, tt.wantRunning)
			}
			if gotNewFailed != tt.wantNewFailed {
				t.Errorf("checkBuildsStatus() gotNewFailed = %v, want %v", gotNewFailed, tt.wantNewFailed)
			}
		})
	}
}

func Test_getBCLabelsAsUniqueSelectors(t *testing.T) {
	type args struct {
		bc *obuildv1.BuildConfig
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"GetBCLabels",
			args{
				&obuildv1.BuildConfig{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app":       "test",
							"buildtype": "s2i",
						},
					},
				},
			},
			"app=test,buildtype=s2i",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getBCLabelsAsUniqueSelectors(tt.args.bc); got != tt.want {
				t.Errorf("getBCLabelsAsUniqueSelectors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_statusUpdateForKogitoApp(t *testing.T) {
	type args struct {
		instance       *v1alpha1.KogitoApp
		cachedInstance *v1alpha1.KogitoApp
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"UpdateSpec",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						ResourceVersion: "v1",
					},
					Spec: v1alpha1.KogitoAppSpec{
						Build: &v1alpha1.KogitoAppBuildObject{
							Native: false,
						},
					},
					Status: v1alpha1.KogitoAppStatus{
						Route: "route1",
					},
				},
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						ResourceVersion: "v1",
					},
					Spec: v1alpha1.KogitoAppSpec{
						Build: &v1alpha1.KogitoAppBuildObject{
							Native: true,
						},
					},
					Status: v1alpha1.KogitoAppStatus{
						Route: "route1",
					},
				},
			},
			true,
		},
		{
			"NoUpdateSpec",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						ResourceVersion: "v1",
					},
					Spec: v1alpha1.KogitoAppSpec{
						Build: &v1alpha1.KogitoAppBuildObject{
							Native: false,
						},
					},
					Status: v1alpha1.KogitoAppStatus{
						Route: "route1",
					},
				},
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						ResourceVersion: "v1",
					},
					Spec: v1alpha1.KogitoAppSpec{
						Build: &v1alpha1.KogitoAppBuildObject{
							Native: false,
						},
					},
					Status: v1alpha1.KogitoAppStatus{
						Route: "route1",
					},
				},
			},
			false,
		},
		{
			"NoUpdateRevisionSpec",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						ResourceVersion: "v1",
					},
					Spec: v1alpha1.KogitoAppSpec{
						Build: &v1alpha1.KogitoAppBuildObject{
							Native: false,
						},
					},
					Status: v1alpha1.KogitoAppStatus{
						Route: "route1",
					},
				},
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						ResourceVersion: "v2",
					},
					Spec: v1alpha1.KogitoAppSpec{
						Build: &v1alpha1.KogitoAppBuildObject{
							Native: true,
						},
					},
					Status: v1alpha1.KogitoAppStatus{
						Route: "route1",
					},
				},
			},
			false,
		},
		{
			"UpdateStatus",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						ResourceVersion: "v1",
					},
					Spec: v1alpha1.KogitoAppSpec{
						Build: &v1alpha1.KogitoAppBuildObject{
							Native: false,
						},
					},
					Status: v1alpha1.KogitoAppStatus{
						Route: "route1",
					},
				},
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						ResourceVersion: "v1",
					},
					Spec: v1alpha1.KogitoAppSpec{
						Build: &v1alpha1.KogitoAppBuildObject{
							Native: false,
						},
					},
					Status: v1alpha1.KogitoAppStatus{
						Route: "route2",
					},
				},
			},
			true,
		},
		{
			"NoUpdateStatus",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						ResourceVersion: "v1",
					},
					Spec: v1alpha1.KogitoAppSpec{
						Build: &v1alpha1.KogitoAppBuildObject{
							Native: false,
						},
					},
					Status: v1alpha1.KogitoAppStatus{
						Route: "route1",
					},
				},
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						ResourceVersion: "v1",
					},
					Spec: v1alpha1.KogitoAppSpec{
						Build: &v1alpha1.KogitoAppBuildObject{
							Native: false,
						},
					},
					Status: v1alpha1.KogitoAppStatus{
						Route: "route1",
					},
				},
			},
			false,
		},
		{
			"NoUpdateRevisionStatus",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						ResourceVersion: "v1",
					},
					Spec: v1alpha1.KogitoAppSpec{
						Build: &v1alpha1.KogitoAppBuildObject{
							Native: false,
						},
					},
					Status: v1alpha1.KogitoAppStatus{
						Route: "route1",
					},
				},
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						ResourceVersion: "v2",
					},
					Spec: v1alpha1.KogitoAppSpec{
						Build: &v1alpha1.KogitoAppBuildObject{
							Native: false,
						},
					},
					Status: v1alpha1.KogitoAppStatus{
						Route: "route2",
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := statusUpdateForKogitoApp(tt.args.instance, tt.args.cachedInstance); got != tt.want {
				t.Errorf("statusUpdateForKogitoApp() = %v, want %v", got, tt.want)
			}
		})
	}
}
