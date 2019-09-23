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

package resource

import (
	"errors"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	routev1 "github.com/openshift/api/route/v1"
	buildfake "github.com/openshift/client-go/build/clientset/versioned/fake"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestManageResources(t *testing.T) {
	s := scheme.Scheme
	s.AddKnownTypes(appsv1.SchemeGroupVersion, &buildv1.BuildConfig{}, &appsv1.DeploymentConfig{}, &routev1.Route{})

	turi := "test-uri"
	replicas := int32(1)
	incremental := true

	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test1",
		},
		Spec: v1alpha1.KogitoAppSpec{
			Runtime:  v1alpha1.QuarkusRuntimeType,
			Replicas: &replicas,
			Resources: v1alpha1.Resources{
				Limits: []v1alpha1.ResourceMap{
					{
						Resource: v1alpha1.ResourceCPU,
						Value:    "500m",
					},
					{
						Resource: v1alpha1.ResourceMemory,
						Value:    "128Mi",
					},
				},
				Requests: []v1alpha1.ResourceMap{
					{
						Resource: v1alpha1.ResourceCPU,
						Value:    "250m",
					},
					{
						Resource: v1alpha1.ResourceMemory,
						Value:    "64Mi",
					},
				},
			},
			Env: []v1alpha1.Env{
				{
					Name:  "test1",
					Value: "test1",
				},
				{
					Name:  "test2",
					Value: "test2",
				},
			},
			Build: &v1alpha1.KogitoAppBuildObject{
				ImageS2I: v1alpha1.Image{
					ImageStreamName:      "kogito-quarkus-ubi8-s2i",
					ImageStreamNamespace: ImageStreamNamespace,
					ImageStreamTag:       ImageStreamTag,
				},
				ImageRuntime: v1alpha1.Image{
					ImageStreamName:      "kogito-quarkus-ubi8",
					ImageStreamNamespace: ImageStreamNamespace,
					ImageStreamTag:       ImageStreamTag,
				},
				Native: true,
				Env: []v1alpha1.Env{
					{Name: "test1", Value: "test1"},
				},
				Incremental: true,
				GitSource: &v1alpha1.GitSource{
					URI:        &turi,
					Reference:  "test-ref",
					ContextDir: "test-ctx",
				},
				Resources: v1alpha1.Resources{
					Limits: []v1alpha1.ResourceMap{
						{
							Resource: v1alpha1.ResourceCPU,
							Value:    "500m",
						},
						{
							Resource: v1alpha1.ResourceMemory,
							Value:    "128Mi",
						},
					},
					Requests: []v1alpha1.ResourceMap{
						{
							Resource: v1alpha1.ResourceCPU,
							Value:    "250m",
						},
						{
							Resource: v1alpha1.ResourceMemory,
							Value:    "64Mi",
						},
					},
				},
			},
			Service: v1alpha1.KogitoAppServiceObject{
				Labels: map[string]string{
					"test2": "test2",
					"test3": "test3",
				},
			},
		},
	}

	kogitoAppResources := &KogitoAppResources{
		KogitoAppResourcesStatus: KogitoAppResourcesStatus{},
		BuildConfigS2I: &buildv1.BuildConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-bc-s2i",
			},
		},
		BuildConfigRuntime: &buildv1.BuildConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-bc-runtime",
			},
		},
		DeploymentConfig: &appsv1.DeploymentConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-dc",
			},
		},
		Route: &routev1.Route{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-route",
			},
		},
		Service: &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-service",
			},
		},
	}

	cl := &client.Client{
		ControlCli: fake.NewFakeClient(
			&buildv1.BuildConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-bc-runtime",
					Labels: map[string]string{
						LabelKeyBuildType: string(BuildTypeRuntime),
					},
				},
				Spec: buildv1.BuildConfigSpec{
					CommonSpec: buildv1.CommonSpec{
						Strategy: buildv1.BuildStrategy{
							SourceStrategy: &buildv1.SourceBuildStrategy{
								From: corev1.ObjectReference{
									Name:      "kogito-quarkus-ubi8:" + ImageStreamTag,
									Namespace: ImageStreamNamespace,
								},
							},
						},
					},
				},
			},
			&buildv1.BuildConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-bc-s2i",
				},
				Spec: buildv1.BuildConfigSpec{
					CommonSpec: buildv1.CommonSpec{
						Source: buildv1.BuildSource{
							ContextDir: "test-ctx",
							Git: &buildv1.GitBuildSource{
								URI: turi,
								Ref: "test-ref",
							},
						},
						Strategy: buildv1.BuildStrategy{
							SourceStrategy: &buildv1.SourceBuildStrategy{
								From: corev1.ObjectReference{
									Name:      "kogito-quarkus-ubi8-s2i:" + ImageStreamTag,
									Namespace: ImageStreamNamespace,
								},
								Env: []corev1.EnvVar{
									{Name: "test1", Value: "test1"},
									{Name: nativeBuildEnvVarKey, Value: "true"},
									{Name: buildS2IlimitCPUEnvVarKey, Value: "500"},
									{Name: buildS2IlimitMemoryEnvVarKey, Value: "134217728"},
								},
								Incremental: &incremental,
							},
						},
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								"cpu":    resource.MustParse("500m"),
								"memory": resource.MustParse("128Mi"),
							},
							Requests: corev1.ResourceList{
								"cpu":    resource.MustParse("250m"),
								"memory": resource.MustParse("64Mi"),
							},
						},
					},
				},
			},
			&appsv1.DeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-dc",
				},
				Spec: appsv1.DeploymentConfigSpec{
					Replicas: replicas,
					Template: &corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Env: []corev1.EnvVar{
										{Name: "test1", Value: "test1"},
										{Name: "test2", Value: "test2"},
									},
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											"cpu":    resource.MustParse("500m"),
											"memory": resource.MustParse("128Mi"),
										},
										Requests: corev1.ResourceList{
											"cpu":    resource.MustParse("250m"),
											"memory": resource.MustParse("64Mi"),
										},
									},
								},
							},
						},
					},
				},
			},
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-service",
					Labels: map[string]string{
						LabelKeyAppName: "test1",
						"test2":         "test2",
						"test3":         "test3",
					},
				},
			},
			&routev1.Route{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-route",
					Labels: map[string]string{
						LabelKeyAppName: "test1",
						"test2":         "test2",
						"test3":         "test3",
					},
				},
			},
		),
		BuildCli: buildfake.NewSimpleClientset().BuildV1(),
	}

	type args struct {
		instance  *v1alpha1.KogitoApp
		resources *KogitoAppResources
		client    *client.Client
	}

	tests := []struct {
		name string
		args args
		want *UpdateResourcesResult
	}{
		{
			"NoUpdate",
			args{
				kogitoApp,
				kogitoAppResources,
				cl,
			},
			&UpdateResourcesResult{
				Updated:     false,
				Err:         nil,
				ErrorReason: v1alpha1.ReasonType(""),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ManageResources(tt.args.instance, tt.args.resources, tt.args.client); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ManageResources() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createResult(t *testing.T) {
	type args struct {
		updated     bool
		err         error
		errorReason v1alpha1.ReasonType
		result      *UpdateResourcesResult
	}
	tests := []struct {
		name            string
		args            args
		wantUpdated     bool
		wantErr         error
		wantErrorReason v1alpha1.ReasonType
	}{
		{
			"Updated_NoError",
			args{
				true,
				nil,
				"",
				&UpdateResourcesResult{
					Updated:     false,
					ErrorReason: "",
					Err:         nil,
				},
			},
			true,
			nil,
			"",
		},
		{
			"NoUpdate_Error",
			args{
				false,
				errors.New("test"),
				"test",
				&UpdateResourcesResult{
					Updated:     true,
					ErrorReason: "",
					Err:         nil,
				},
			},
			true,
			errors.New("test"),
			"test",
		},
		{
			"Update_NoError",
			args{
				true,
				errors.New("test"),
				"test",
				&UpdateResourcesResult{
					Updated:     true,
					ErrorReason: "test1",
					Err:         errors.New("test1"),
				},
			},
			true,
			errors.New("test1"),
			"test1",
		},
		{
			"NoUpdate_NoError",
			args{
				false,
				nil,
				"",
				&UpdateResourcesResult{
					Updated:     false,
					ErrorReason: "test1",
					Err:         errors.New("test1"),
				},
			},
			false,
			errors.New("test1"),
			"test1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createResult(tt.args.updated, tt.args.err, tt.args.errorReason, tt.args.result)
			if tt.args.result.Updated != tt.wantUpdated {
				t.Errorf("createResult() updated = %v, wantUpdated %v", tt.args.result.Updated, tt.wantUpdated)
				return
			}
			if !reflect.DeepEqual(tt.args.result.Err, tt.wantErr) {
				t.Errorf("createResult() error = %v, wantErr %v", tt.args.result.Err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.args.result.ErrorReason, tt.wantErrorReason) {
				t.Errorf("createResult() ErrorReason = %v, want %v", tt.args.result.ErrorReason, tt.wantErrorReason)
			}
		})
	}
}

func Test_ensureBuildConfigRuntime(t *testing.T) {
	s := scheme.Scheme
	s.AddKnownTypes(appsv1.SchemeGroupVersion, &buildv1.BuildConfig{})

	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test1",
		},
		Spec: v1alpha1.KogitoAppSpec{
			Runtime: v1alpha1.QuarkusRuntimeType,
			Build: &v1alpha1.KogitoAppBuildObject{
				ImageRuntime: v1alpha1.Image{
					ImageStreamName:      "kogito-quarkus-ubi8",
					ImageStreamNamespace: ImageStreamNamespace,
					ImageStreamTag:       ImageStreamTag,
				},
				Native: true,
			},
		},
	}

	type args struct {
		instance           *v1alpha1.KogitoApp
		buildConfigRuntime *buildv1.BuildConfig
		client             *client.Client
	}

	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			"NoUpdate",
			args{
				kogitoApp,
				&buildv1.BuildConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-bc",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&buildv1.BuildConfig{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-bc",
								Labels: map[string]string{
									LabelKeyBuildType: string(BuildTypeRuntime),
								},
							},
							Spec: buildv1.BuildConfigSpec{
								CommonSpec: buildv1.CommonSpec{
									Strategy: buildv1.BuildStrategy{
										SourceStrategy: &buildv1.SourceBuildStrategy{
											From: corev1.ObjectReference{
												Name:      "kogito-quarkus-ubi8:" + ImageStreamTag,
												Namespace: ImageStreamNamespace,
											},
										},
									},
								},
							},
						},
					),
					BuildCli: buildfake.NewSimpleClientset().BuildV1(),
				},
			},
			false,
			false,
		},
		{
			"Update_Label",
			args{
				kogitoApp,
				&buildv1.BuildConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-bc",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&buildv1.BuildConfig{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-bc",
								Labels: map[string]string{
									LabelKeyBuildType: string(BuildTypeRuntimeJvm),
								},
							},
							Spec: buildv1.BuildConfigSpec{
								CommonSpec: buildv1.CommonSpec{
									Strategy: buildv1.BuildStrategy{
										SourceStrategy: &buildv1.SourceBuildStrategy{
											From: corev1.ObjectReference{
												Name:      "kogito-quarkus-ubi8:" + ImageStreamTag,
												Namespace: ImageStreamNamespace,
											},
										},
									},
								},
							},
						},
					),
					BuildCli: buildfake.NewSimpleClientset().BuildV1(),
				},
			},
			true,
			false,
		},
		{
			"Update_ImageName",
			args{
				kogitoApp,
				&buildv1.BuildConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-bc",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&buildv1.BuildConfig{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-bc",
								Labels: map[string]string{
									LabelKeyBuildType: string(BuildTypeRuntimeJvm),
								},
							},
							Spec: buildv1.BuildConfigSpec{
								CommonSpec: buildv1.CommonSpec{
									Strategy: buildv1.BuildStrategy{
										SourceStrategy: &buildv1.SourceBuildStrategy{
											From: corev1.ObjectReference{
												Name:      "kogito-springboot-ubi8:" + ImageStreamTag,
												Namespace: ImageStreamNamespace,
											},
										},
									},
								},
							},
						},
					),
					BuildCli: buildfake.NewSimpleClientset().BuildV1(),
				},
			},
			true,
			false,
		},
		{
			"Update_ImageNamespace",
			args{
				kogitoApp,
				&buildv1.BuildConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-bc",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&buildv1.BuildConfig{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-bc",
								Labels: map[string]string{
									LabelKeyBuildType: string(BuildTypeRuntimeJvm),
								},
							},
							Spec: buildv1.BuildConfigSpec{
								CommonSpec: buildv1.CommonSpec{
									Strategy: buildv1.BuildStrategy{
										SourceStrategy: &buildv1.SourceBuildStrategy{
											From: corev1.ObjectReference{
												Name:      "kogito-quarkus-ubi8:" + ImageStreamTag,
												Namespace: ImageStreamNamespace + "a",
											},
										},
									},
								},
							},
						},
					),
					BuildCli: buildfake.NewSimpleClientset().BuildV1(),
				},
			},
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ensureBuildConfigRuntime(tt.args.instance, tt.args.buildConfigRuntime, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("ensureBuildConfigRuntime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ensureBuildConfigRuntime() got = %v, want %v", got, tt.want)
				return
			}

			if tt.args.buildConfigRuntime.Labels[LabelKeyBuildType] != string(BuildTypeRuntime) {
				t.Errorf("ensureBuildConfigRuntime() runtimeType = %v, wantRuntimeType %v",
					tt.args.buildConfigRuntime.Labels[LabelKeyBuildType], string(BuildTypeRuntime))
				return
			}
			if tt.args.buildConfigRuntime.Spec.Strategy.SourceStrategy.From.Name != ("kogito-quarkus-ubi8:" + ImageStreamTag) {
				t.Errorf("ensureBuildConfigRuntime() imageName = %v, wantImageName %v",
					tt.args.buildConfigRuntime.Spec.Strategy.SourceStrategy.From.Name, "kogito-quarkus-ubi8:"+ImageStreamTag)
				return
			}
			if tt.args.buildConfigRuntime.Spec.Strategy.SourceStrategy.From.Namespace != ImageStreamNamespace {
				t.Errorf("ensureBuildConfigRuntime() imageName = %v, wantImageName %v",
					tt.args.buildConfigRuntime.Spec.Strategy.SourceStrategy.From.Namespace, ImageStreamNamespace)
			}
		})
	}
}

func Test_ensureBuildConfigS2I(t *testing.T) {
	s := scheme.Scheme
	s.AddKnownTypes(appsv1.SchemeGroupVersion, &buildv1.BuildConfig{})

	turi := "test-uri"
	incremental := true
	notIncremental := false
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test1",
		},
		Spec: v1alpha1.KogitoAppSpec{
			Runtime: v1alpha1.QuarkusRuntimeType,
			Build: &v1alpha1.KogitoAppBuildObject{
				ImageS2I: v1alpha1.Image{
					ImageStreamName:      "kogito-quarkus-ubi8-s2i",
					ImageStreamNamespace: ImageStreamNamespace,
					ImageStreamTag:       ImageStreamTag,
				},
				Native: true,
				Env: []v1alpha1.Env{
					{Name: "test1", Value: "test1"},
				},
				Incremental: true,
				GitSource: &v1alpha1.GitSource{
					URI:        &turi,
					Reference:  "test-ref",
					ContextDir: "test-ctx",
				},
				Resources: v1alpha1.Resources{
					Limits: []v1alpha1.ResourceMap{
						{
							Resource: v1alpha1.ResourceCPU,
							Value:    "500m",
						},
						{
							Resource: v1alpha1.ResourceMemory,
							Value:    "128Mi",
						},
					},
					Requests: []v1alpha1.ResourceMap{
						{
							Resource: v1alpha1.ResourceCPU,
							Value:    "250m",
						},
						{
							Resource: v1alpha1.ResourceMemory,
							Value:    "64Mi",
						},
					},
				},
			},
		},
	}
	buildSource := buildv1.BuildSource{
		ContextDir: "test-ctx",
		Git: &buildv1.GitBuildSource{
			URI: turi,
			Ref: "test-ref",
		},
	}
	var envs = []corev1.EnvVar{
		{Name: "test1", Value: "test1"},
		{Name: nativeBuildEnvVarKey, Value: "true"},
		{Name: buildS2IlimitCPUEnvVarKey, Value: "500"},
		{Name: buildS2IlimitMemoryEnvVarKey, Value: "134217728"},
	}
	resources := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			"cpu":    resource.MustParse("500m"),
			"memory": resource.MustParse("128Mi"),
		},
		Requests: corev1.ResourceList{
			"cpu":    resource.MustParse("250m"),
			"memory": resource.MustParse("64Mi"),
		},
	}

	type args struct {
		instance       *v1alpha1.KogitoApp
		buildConfigS2I *buildv1.BuildConfig
		client         *client.Client
	}

	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			"NoUpdate",
			args{
				kogitoApp,
				&buildv1.BuildConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-bc",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&buildv1.BuildConfig{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-bc",
							},
							Spec: buildv1.BuildConfigSpec{
								CommonSpec: buildv1.CommonSpec{
									Source: buildSource,
									Strategy: buildv1.BuildStrategy{
										SourceStrategy: &buildv1.SourceBuildStrategy{
											From: corev1.ObjectReference{
												Name:      "kogito-quarkus-ubi8-s2i:" + ImageStreamTag,
												Namespace: ImageStreamNamespace,
											},
											Env:         envs,
											Incremental: &incremental,
										},
									},
									Resources: resources,
								},
							},
						},
					),
					BuildCli: buildfake.NewSimpleClientset().BuildV1(),
				},
			},
			false,
			false,
		},
		{
			"Update_Resources",
			args{
				kogitoApp,
				&buildv1.BuildConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-bc",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&buildv1.BuildConfig{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-bc",
							},
							Spec: buildv1.BuildConfigSpec{
								CommonSpec: buildv1.CommonSpec{
									Source: buildSource,
									Strategy: buildv1.BuildStrategy{
										SourceStrategy: &buildv1.SourceBuildStrategy{
											From: corev1.ObjectReference{
												Name:      "kogito-quarkus-ubi8-s2i:" + ImageStreamTag,
												Namespace: ImageStreamNamespace,
											},
											Env:         envs,
											Incremental: &incremental,
										},
									},
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											"cpu":    resource.MustParse("250m"),
											"memory": resource.MustParse("64Mi"),
										},
										Requests: corev1.ResourceList{
											"cpu":    resource.MustParse("125m"),
											"memory": resource.MustParse("32Mi"),
										},
									},
								},
							},
						},
					),
					BuildCli: buildfake.NewSimpleClientset().BuildV1(),
				},
			},
			true,
			false,
		},
		{
			"Update_Incremental",
			args{
				kogitoApp,
				&buildv1.BuildConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-bc",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&buildv1.BuildConfig{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-bc",
							},
							Spec: buildv1.BuildConfigSpec{
								CommonSpec: buildv1.CommonSpec{
									Source: buildSource,
									Strategy: buildv1.BuildStrategy{
										SourceStrategy: &buildv1.SourceBuildStrategy{
											From: corev1.ObjectReference{
												Name:      "kogito-quarkus-ubi8-s2i:" + ImageStreamTag,
												Namespace: ImageStreamNamespace,
											},
											Env:         envs,
											Incremental: &notIncremental,
										},
									},
									Resources: resources,
								},
							},
						},
					),
					BuildCli: buildfake.NewSimpleClientset().BuildV1(),
				},
			},
			true,
			false,
		},
		{
			"Update_ImageName",
			args{
				kogitoApp,
				&buildv1.BuildConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-bc",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&buildv1.BuildConfig{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-bc",
							},
							Spec: buildv1.BuildConfigSpec{
								CommonSpec: buildv1.CommonSpec{
									Source: buildSource,
									Strategy: buildv1.BuildStrategy{
										SourceStrategy: &buildv1.SourceBuildStrategy{
											From: corev1.ObjectReference{
												Name:      "kogito-springboot-ubi8-s2i:" + ImageStreamTag,
												Namespace: ImageStreamNamespace,
											},
											Env:         envs,
											Incremental: &incremental,
										},
									},
									Resources: resources,
								},
							},
						},
					),
					BuildCli: buildfake.NewSimpleClientset().BuildV1(),
				},
			},
			true,
			false,
		},
		{
			"Update-ImageNamespace",
			args{
				kogitoApp,
				&buildv1.BuildConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-bc",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&buildv1.BuildConfig{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-bc",
							},
							Spec: buildv1.BuildConfigSpec{
								CommonSpec: buildv1.CommonSpec{
									Source: buildSource,
									Strategy: buildv1.BuildStrategy{
										SourceStrategy: &buildv1.SourceBuildStrategy{
											From: corev1.ObjectReference{
												Name:      "kogito-quarkus-ubi8-s2i:" + ImageStreamTag,
												Namespace: ImageStreamNamespace + "a",
											},
											Env:         envs,
											Incremental: &incremental,
										},
									},
									Resources: resources,
								},
							},
						},
					),
					BuildCli: buildfake.NewSimpleClientset().BuildV1(),
				},
			},
			true,
			false,
		},
		{
			"Update_Env",
			args{
				kogitoApp,
				&buildv1.BuildConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-bc",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&buildv1.BuildConfig{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-bc",
							},
							Spec: buildv1.BuildConfigSpec{
								CommonSpec: buildv1.CommonSpec{
									Source: buildSource,
									Strategy: buildv1.BuildStrategy{
										SourceStrategy: &buildv1.SourceBuildStrategy{
											From: corev1.ObjectReference{
												Name:      "kogito-quarkus-ubi8-s2i:" + ImageStreamTag,
												Namespace: ImageStreamNamespace,
											},
											Env: []corev1.EnvVar{
												{Name: "test1", Value: "test2"},
											},
											Incremental: &incremental,
										},
									},
									Resources: resources,
								},
							},
						},
					),
					BuildCli: buildfake.NewSimpleClientset().BuildV1(),
				},
			},
			true,
			false,
		},
		{
			"Update_SourceURI",
			args{
				kogitoApp,
				&buildv1.BuildConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-bc",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&buildv1.BuildConfig{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-bc",
							},
							Spec: buildv1.BuildConfigSpec{
								CommonSpec: buildv1.CommonSpec{
									Source: buildv1.BuildSource{
										ContextDir: "test-ctx",
										Git: &buildv1.GitBuildSource{
											URI: turi + "-1",
											Ref: "test-ref",
										},
									},
									Strategy: buildv1.BuildStrategy{
										SourceStrategy: &buildv1.SourceBuildStrategy{
											From: corev1.ObjectReference{
												Name:      "kogito-quarkus-ubi8-s2i:" + ImageStreamTag,
												Namespace: ImageStreamNamespace,
											},
											Env:         envs,
											Incremental: &incremental,
										},
									},
									Resources: resources,
								},
							},
						},
					),
					BuildCli: buildfake.NewSimpleClientset().BuildV1(),
				},
			},
			true,
			false,
		},
		{
			"Update_SourceRef",
			args{
				kogitoApp,
				&buildv1.BuildConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-bc",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&buildv1.BuildConfig{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-bc",
							},
							Spec: buildv1.BuildConfigSpec{
								CommonSpec: buildv1.CommonSpec{
									Source: buildv1.BuildSource{
										ContextDir: "test-ctx",
										Git: &buildv1.GitBuildSource{
											URI: turi,
											Ref: "test-ref-1",
										},
									},
									Strategy: buildv1.BuildStrategy{
										SourceStrategy: &buildv1.SourceBuildStrategy{
											From: corev1.ObjectReference{
												Name:      "kogito-quarkus-ubi8-s2i:" + ImageStreamTag,
												Namespace: ImageStreamNamespace,
											},
											Env:         envs,
											Incremental: &incremental,
										},
									},
									Resources: resources,
								},
							},
						},
					),
					BuildCli: buildfake.NewSimpleClientset().BuildV1(),
				},
			},
			true,
			false,
		},
		{
			"Update_SourceContext",
			args{
				kogitoApp,
				&buildv1.BuildConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-bc",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&buildv1.BuildConfig{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-bc",
							},
							Spec: buildv1.BuildConfigSpec{
								CommonSpec: buildv1.CommonSpec{
									Source: buildv1.BuildSource{
										ContextDir: "test-ctx-1",
										Git: &buildv1.GitBuildSource{
											URI: turi,
											Ref: "test-ref",
										},
									},
									Strategy: buildv1.BuildStrategy{
										SourceStrategy: &buildv1.SourceBuildStrategy{
											From: corev1.ObjectReference{
												Name:      "kogito-quarkus-ubi8-s2i:" + ImageStreamTag,
												Namespace: ImageStreamNamespace,
											},
											Env:         envs,
											Incremental: &incremental,
										},
									},
									Resources: resources,
								},
							},
						},
					),
					BuildCli: buildfake.NewSimpleClientset().BuildV1(),
				},
			},
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ensureBuildConfigS2I(tt.args.instance, tt.args.buildConfigS2I, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("ensureBuildConfigS2I() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ensureBuildConfigS2I() got = %v, want %v", got, tt.want)
				return
			}

			if tt.args.buildConfigS2I.Spec.Strategy.SourceStrategy.From.Name != ("kogito-quarkus-ubi8-s2i:" + ImageStreamTag) {
				t.Errorf("ensureBuildConfigS2I() imageName = %v, wantImageName %v",
					tt.args.buildConfigS2I.Spec.Strategy.SourceStrategy.From.Name, "kogito-quarkus-ubi8-s2i:"+ImageStreamTag)
				return
			}
			if tt.args.buildConfigS2I.Spec.Strategy.SourceStrategy.From.Namespace != ImageStreamNamespace {
				t.Errorf("ensureBuildConfigS2I() imageName = %v, wantImageName %v",
					tt.args.buildConfigS2I.Spec.Strategy.SourceStrategy.From.Namespace, ImageStreamNamespace)
				return
			}
			if tt.args.buildConfigS2I.Spec.Strategy.SourceStrategy.From.Namespace != ImageStreamNamespace {
				t.Errorf("ensureBuildConfigS2I() imageName = %v, wantImageName %v",
					tt.args.buildConfigS2I.Spec.Strategy.SourceStrategy.From.Namespace, ImageStreamNamespace)
				return
			}
			if !util.EnvVarArrayEquals(tt.args.buildConfigS2I.Spec.Strategy.SourceStrategy.Env, envs) {
				t.Errorf("ensureBuildConfigS2I() envs = %v, wantEnvs %v", tt.args.buildConfigS2I.Spec.Strategy.SourceStrategy.Env, envs)
				return
			}
			if !reflect.DeepEqual(tt.args.buildConfigS2I.Spec.Strategy.SourceStrategy.Incremental, &incremental) {
				t.Errorf("ensureBuildConfigS2I() incremental = %v, wantIncremental %v",
					tt.args.buildConfigS2I.Spec.Strategy.SourceStrategy.Incremental, &incremental)
				return
			}
			if !reflect.DeepEqual(tt.args.buildConfigS2I.Spec.Source, buildSource) {
				t.Errorf("ensureBuildConfigS2I() source = %v, wantSource %v",
					tt.args.buildConfigS2I.Spec.Source, buildSource)
				return
			}
		})
	}
}

func Test_ensureDeploymentConfig(t *testing.T) {
	s := scheme.Scheme
	s.AddKnownTypes(appsv1.SchemeGroupVersion, &appsv1.DeploymentConfig{})

	replicas := int32(1)
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test1",
		},
		Spec: v1alpha1.KogitoAppSpec{
			Replicas: &replicas,
			Resources: v1alpha1.Resources{
				Limits: []v1alpha1.ResourceMap{
					{
						Resource: v1alpha1.ResourceCPU,
						Value:    "500m",
					},
					{
						Resource: v1alpha1.ResourceMemory,
						Value:    "128Mi",
					},
				},
				Requests: []v1alpha1.ResourceMap{
					{
						Resource: v1alpha1.ResourceCPU,
						Value:    "250m",
					},
					{
						Resource: v1alpha1.ResourceMemory,
						Value:    "64Mi",
					},
				},
			},
			Env: []v1alpha1.Env{
				{
					Name:  "test1",
					Value: "test1",
				},
				{
					Name:  "test2",
					Value: "test2",
				},
			},
		},
	}
	resources := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			"cpu":    resource.MustParse("500m"),
			"memory": resource.MustParse("128Mi"),
		},
		Requests: corev1.ResourceList{
			"cpu":    resource.MustParse("250m"),
			"memory": resource.MustParse("64Mi"),
		},
	}
	envs := []corev1.EnvVar{
		{Name: "test1", Value: "test1"},
		{Name: "test2", Value: "test2"},
	}

	type args struct {
		instance  *v1alpha1.KogitoApp
		depConfig *appsv1.DeploymentConfig
		client    *client.Client
	}

	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			"NoUpdate",
			args{
				kogitoApp,
				&appsv1.DeploymentConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-dc",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&appsv1.DeploymentConfig{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-dc",
							},
							Spec: appsv1.DeploymentConfigSpec{
								Replicas: replicas,
								Template: &corev1.PodTemplateSpec{
									Spec: corev1.PodSpec{
										Containers: []corev1.Container{
											{
												Env:       envs,
												Resources: resources,
											},
										},
									},
								},
							},
						},
					),
				},
			},
			false,
			false,
		},
		{
			"Update_Replicas",
			args{
				kogitoApp,
				&appsv1.DeploymentConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-dc",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&appsv1.DeploymentConfig{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-dc",
							},
							Spec: appsv1.DeploymentConfigSpec{
								Replicas: int32(2),
								Template: &corev1.PodTemplateSpec{
									Spec: corev1.PodSpec{
										Containers: []corev1.Container{
											{
												Env:       envs,
												Resources: resources,
											},
										},
									},
								},
							},
						},
					),
				},
			},
			true,
			false,
		},
		{
			"Update_Env",
			args{
				kogitoApp,
				&appsv1.DeploymentConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-dc",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&appsv1.DeploymentConfig{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-dc",
							},
							Spec: appsv1.DeploymentConfigSpec{
								Replicas: replicas,
								Template: &corev1.PodTemplateSpec{
									Spec: corev1.PodSpec{
										Containers: []corev1.Container{
											{
												Env: []corev1.EnvVar{
													{Name: "test1", Value: "test2"},
													{Name: "test2", Value: "test1"},
												},
												Resources: resources,
											},
										},
									},
								},
							},
						},
					),
				},
			},
			true,
			false,
		},
		{
			"Update_Resources",
			args{
				kogitoApp,
				&appsv1.DeploymentConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-dc",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&appsv1.DeploymentConfig{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-dc",
							},
							Spec: appsv1.DeploymentConfigSpec{
								Replicas: replicas,
								Template: &corev1.PodTemplateSpec{
									Spec: corev1.PodSpec{
										Containers: []corev1.Container{
											{
												Env: envs,
												Resources: corev1.ResourceRequirements{
													Limits: corev1.ResourceList{
														"cpu":    resource.MustParse("250m"),
														"memory": resource.MustParse("64Mi"),
													},
													Requests: corev1.ResourceList{
														"cpu":    resource.MustParse("125m"),
														"memory": resource.MustParse("32Mi"),
													},
												},
											},
										},
									},
								},
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
			got, err := ensureDeploymentConfig(tt.args.instance, tt.args.depConfig, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("ensureDeploymentConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ensureDeploymentConfig() got = %v, want %v", got, tt.want)
				return
			}

			if tt.args.depConfig.Spec.Replicas != replicas {
				t.Errorf("ensureDeploymentConfig() replicas = %v, wantReplicas %v", tt.args.depConfig.Spec.Replicas, replicas)
				return
			}
			if !reflect.DeepEqual(tt.args.depConfig.Spec.Template.Spec.Containers[0].Resources, resources) {
				t.Errorf("ensureDeploymentConfig() resources = %v, wantResources %v", tt.args.depConfig.Spec.Template.Spec.Containers[0].Resources, resources)
				return
			}
			if !util.EnvVarArrayEquals(tt.args.depConfig.Spec.Template.Spec.Containers[0].Env, envs) {
				t.Errorf("ensureDeploymentConfig() envs = %v, wantEnvs %v", tt.args.depConfig.Spec.Template.Spec.Containers[0].Env, envs)
			}
		})
	}
}

func Test_ensureRoute(t *testing.T) {
	type args struct {
		instance *v1alpha1.KogitoApp
		route    *routev1.Route
		client   *client.Client
	}

	s := scheme.Scheme
	s.AddKnownTypes(appsv1.SchemeGroupVersion, &routev1.Route{})

	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			"TestEnsureRoute_NoUpdate",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
					Spec: v1alpha1.KogitoAppSpec{
						Service: v1alpha1.KogitoAppServiceObject{
							Labels: map[string]string{
								"test2": "test2",
								"test3": "test3",
							},
						},
					},
				},
				&routev1.Route{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-route",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&routev1.Route{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-route",
								Labels: map[string]string{
									LabelKeyAppName: "test1",
									"test2":         "test2",
									"test3":         "test3",
								},
							},
						},
					),
				},
			},
			false,
			false,
		},
		{
			"TestEnsureRoute_Updated_Name",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
					Spec: v1alpha1.KogitoAppSpec{
						Service: v1alpha1.KogitoAppServiceObject{
							Labels: map[string]string{
								"test2": "test2",
								"test3": "test3",
							},
						},
					},
				},
				&routev1.Route{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-route",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&routev1.Route{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-route",
								Labels: map[string]string{
									LabelKeyAppName: "test2",
									"test2":         "test2",
									"test3":         "test3",
								},
							},
						},
					),
				},
			},
			true,
			false,
		},
		{
			"TestEnsureRoute_Updated_Labels",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
					Spec: v1alpha1.KogitoAppSpec{
						Service: v1alpha1.KogitoAppServiceObject{
							Labels: map[string]string{
								"test2": "test2",
								"test3": "test3",
							},
						},
					},
				},
				&routev1.Route{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-route",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&routev1.Route{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-route",
								Labels: map[string]string{
									LabelKeyAppName: "test1",
									"test2":         "test3",
									"test3":         "test2",
								},
							},
						},
					),
				},
			},
			true,
			false,
		},
		{
			"TestEnsureRoute_NoUpdate_Default",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
				},
				&routev1.Route{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-route",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&routev1.Route{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-route",
								Labels: map[string]string{
									LabelKeyAppName:     "test1",
									LabelKeyServiceName: "test1",
								},
							},
						},
					),
				},
			},
			false,
			false,
		},
		{
			"TestEnsureRoute_Updated_Default",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
				},
				&routev1.Route{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-route",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&routev1.Route{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-route",
								Labels: map[string]string{
									LabelKeyAppName: "test1",
								},
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
			got, err := ensureRoute(tt.args.instance, tt.args.route, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("ensureRoute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ensureRoute() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ensureService(t *testing.T) {
	type args struct {
		instance *v1alpha1.KogitoApp
		service  *corev1.Service
		client   *client.Client
	}

	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			"TestEnsureService_NoUpdate",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
					Spec: v1alpha1.KogitoAppSpec{
						Service: v1alpha1.KogitoAppServiceObject{
							Labels: map[string]string{
								"test2": "test2",
								"test3": "test3",
							},
						},
					},
				},
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-service",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&corev1.Service{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-service",
								Labels: map[string]string{
									LabelKeyAppName: "test1",
									"test2":         "test2",
									"test3":         "test3",
								},
							},
						},
					),
				},
			},
			false,
			false,
		},
		{
			"TestEnsureService_Updated_Name",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
					Spec: v1alpha1.KogitoAppSpec{
						Service: v1alpha1.KogitoAppServiceObject{
							Labels: map[string]string{
								"test2": "test2",
								"test3": "test3",
							},
						},
					},
				},
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-service",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&corev1.Service{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-service",
								Labels: map[string]string{
									LabelKeyAppName: "test2",
									"test2":         "test2",
									"test3":         "test3",
								},
							},
						},
					),
				},
			},
			true,
			false,
		},
		{
			"TestEnsureService_Updated_Labels",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
					Spec: v1alpha1.KogitoAppSpec{
						Service: v1alpha1.KogitoAppServiceObject{
							Labels: map[string]string{
								"test2": "test2",
								"test3": "test3",
							},
						},
					},
				},
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-service",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&corev1.Service{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-service",
								Labels: map[string]string{
									LabelKeyAppName: "test1",
									"test2":         "test3",
									"test3":         "test2",
								},
							},
						},
					),
				},
			},
			true,
			false,
		},
		{
			"TestEnsureService_NoUpdate_Default",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
				},
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-service",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&corev1.Service{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-service",
								Labels: map[string]string{
									LabelKeyAppName:     "test1",
									LabelKeyServiceName: "test1",
								},
							},
						},
					),
				},
			},
			false,
			false,
		},
		{
			"TestEnsureService_Updated_Default",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
				},
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-service",
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(
						&corev1.Service{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-service",
								Labels: map[string]string{
									LabelKeyAppName: "test1",
								},
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
			got, err := ensureService(tt.args.instance, tt.args.service, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("ensureService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ensureService() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ensureServiceLabels(t *testing.T) {
	type args struct {
		instance      *v1alpha1.KogitoApp
		serviceLabels map[string]string
		log           *zap.SugaredLogger
	}

	log := log.With("Test", "ensureServiceLabels")

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"TestEnsureServiceLabels_NoUpdate",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
					Spec: v1alpha1.KogitoAppSpec{
						Service: v1alpha1.KogitoAppServiceObject{
							Labels: map[string]string{
								"test2": "test2",
								"test3": "test3",
							},
						},
					},
				},
				map[string]string{
					LabelKeyAppName: "test1",
					"test2":         "test2",
					"test3":         "test3",
				},
				log,
			},
			false,
		},
		{
			"TestEnsureServiceLabels_Updated_Name",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
					Spec: v1alpha1.KogitoAppSpec{
						Service: v1alpha1.KogitoAppServiceObject{
							Labels: map[string]string{
								"test2": "test2",
								"test3": "test3",
							},
						},
					},
				},
				map[string]string{
					LabelKeyAppName: "test2",
					"test2":         "test2",
					"test3":         "test3",
				},
				log,
			},
			true,
		},
		{
			"TestEnsureServiceLabels_Updated_Labels",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
					Spec: v1alpha1.KogitoAppSpec{
						Service: v1alpha1.KogitoAppServiceObject{
							Labels: map[string]string{
								"test2": "test2",
								"test3": "test3",
							},
						},
					},
				},
				map[string]string{
					LabelKeyAppName: "test1",
					"test2":         "test3",
					"test3":         "test2",
				},
				log,
			},
			true,
		},
		{
			"TestEnsureServiceLabels_NoUpdate_Default",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
				},
				map[string]string{
					LabelKeyAppName:     "test1",
					LabelKeyServiceName: "test1",
				},
				log,
			},
			false,
		},
		{
			"TestEnsureServiceLabels_Updated_Default",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
				},
				map[string]string{
					LabelKeyAppName: "test1",
				},
				log,
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ensureServiceLabels(tt.args.instance, tt.args.serviceLabels, tt.args.log); got != tt.want {
				t.Errorf("ensureServiceLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}
