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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"

	"reflect"
	"testing"

	monv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	monfake "github.com/coreos/prometheus-operator/pkg/client/versioned/fake"

	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetDeployedResources(t *testing.T) {
	s := scheme.Scheme
	s.AddKnownTypes(buildv1.GroupVersion, &buildv1.BuildConfig{}, &buildv1.BuildConfigList{})
	s.AddKnownTypes(appsv1.GroupVersion, &appsv1.DeploymentConfig{}, &appsv1.DeploymentConfigList{})
	s.AddKnownTypes(imgv1.GroupVersion, &imgv1.ImageStream{}, &imgv1.ImageStreamList{})
	s.AddKnownTypes(routev1.GroupVersion, &routev1.Route{}, &routev1.RouteList{})
	s.AddKnownTypes(monv1.SchemeGroupVersion, &monv1.ServiceMonitor{}, &monv1.ServiceMonitorList{})

	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			UID:       "1234567",
			Namespace: "test",
			Name:      "kogito",
		},
	}

	bc1 := buildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "bc1",
			OwnerReferences: []metav1.OwnerReference{
				{
					UID: "1234567",
				},
			},
		},
	}

	bc2 := buildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "bc2",
			OwnerReferences: []metav1.OwnerReference{
				{
					UID: "1234567890",
				},
			},
		},
	}

	dc1 := appsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "dc1",
			OwnerReferences: []metav1.OwnerReference{
				{
					UID: "1234567",
				},
			},
		},
	}

	dc2 := appsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "dc2",
			OwnerReferences: []metav1.OwnerReference{
				{
					UID: "1234567890",
				},
			},
		},
	}

	is1 := imgv1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "is1",
			OwnerReferences: []metav1.OwnerReference{
				{
					UID: "1234567",
				},
			},
		},
	}

	is2 := imgv1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "is2",
			OwnerReferences: []metav1.OwnerReference{
				{
					UID: "1234567890",
				},
			},
		},
	}

	rt1 := routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "is1",
			OwnerReferences: []metav1.OwnerReference{
				{
					UID: "1234567",
				},
			},
		},
	}

	rt2 := routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "is2",
			OwnerReferences: []metav1.OwnerReference{
				{
					UID: "1234567890",
				},
			},
		},
	}

	svc1 := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "svc1",
			Namespace: "test",
			OwnerReferences: []metav1.OwnerReference{
				{
					UID: "1234567",
				},
			},
		},
	}

	svc2 := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "svc2",
			Namespace: "test",
			OwnerReferences: []metav1.OwnerReference{
				{
					UID: "1234567890",
				},
			},
		},
	}

	sm1 := monv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sm1",
			Namespace: "test",
			OwnerReferences: []metav1.OwnerReference{
				{
					UID: "1234567",
				},
			},
		},
	}

	sm2 := monv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sm2",
			Namespace: "test",
			OwnerReferences: []metav1.OwnerReference{
				{
					UID: "1234567890",
				},
			},
		},
	}

	type args struct {
		instance *v1alpha1.KogitoApp
		client   *client.Client
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"GetDeployedResources",
			args{
				kogitoApp,
				&client.Client{
					ControlCli:    fake.NewFakeClient(&bc1, &bc2, &dc1, &dc2, &is1, &is2, &rt1, &rt2, &svc1, &svc2),
					PrometheusCli: monfake.NewSimpleClientset(&sm1, &sm2).MonitoringV1(),
					Discovery:     test.CreateFakeDiscoveryClient(false),
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDeployedResources(tt.args.instance, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDeployedResources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != 6 {
				t.Errorf("GetDeployedResources() got = %v, want %v", got, "6 types of resources")
			}
			if len(got[reflect.TypeOf(buildv1.BuildConfig{})]) != 1 ||
				got[reflect.TypeOf(buildv1.BuildConfig{})][0].GetName() != bc1.GetName() {
				t.Errorf("getBuildConfig() gotBC = %v, want %v", got, bc1)
			}
			if len(got[reflect.TypeOf(appsv1.DeploymentConfig{})]) != 1 ||
				got[reflect.TypeOf(appsv1.DeploymentConfig{})][0].GetName() != dc1.GetName() {
				t.Errorf("getDeploymentConfig() gotDC = %v, want %v", got, dc1)
			}
			if len(got[reflect.TypeOf(imgv1.ImageStream{})]) != 1 ||
				got[reflect.TypeOf(imgv1.ImageStream{})][0].GetName() != is1.GetName() {
				t.Errorf("getImageStream() gotIS = %v, want %v", got, is1)
			}
			if len(got[reflect.TypeOf(routev1.Route{})]) != 1 ||
				got[reflect.TypeOf(routev1.Route{})][0].GetName() != rt1.GetName() {
				t.Errorf("getRoute() gotRoute = %v, want %v", got, rt1)
			}
			if len(got[reflect.TypeOf(v1.Service{})]) != 1 ||
				got[reflect.TypeOf(v1.Service{})][0].GetName() != svc1.GetName() {
				t.Errorf("getService() gotService = %v, want %v", got, svc1)
			}
			if len(got[reflect.TypeOf(monv1.ServiceMonitor{})]) != 1 ||
				got[reflect.TypeOf(monv1.ServiceMonitor{})][0].GetName() != sm1.GetName() {
				t.Errorf("getServiceMonitor() gotServiceMonitor = %v, want %v", got, sm1)
			}
		})
	}
}
