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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	imgv1 "github.com/openshift/api/image/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"

	routev1 "github.com/openshift/api/route/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"testing"
)

func TestGetDeployedResources(t *testing.T) {
	dataIndex := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{
			UID:       "1234567",
			Namespace: "test",
			Name:      "kogitoDataIndex",
		},
	}

	is1 := imgv1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "imagestream1",
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
			Name:      "imagestream2",
			OwnerReferences: []metav1.OwnerReference{
				{
					UID: "1234567890",
				},
			},
		},
	}

	dep1 := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "dep1",
			OwnerReferences: []metav1.OwnerReference{
				{
					UID: "1234567",
				},
			},
		},
	}

	dep2 := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "dep2",
			OwnerReferences: []metav1.OwnerReference{
				{
					UID: "1234567890",
				},
			},
		},
	}

	svc1 := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "svc1",
			OwnerReferences: []metav1.OwnerReference{
				{
					UID: "1234567",
				},
			},
		},
	}

	svc2 := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "svc2",
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
			Name:      "rt1",
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
			Name:      "rt2",
			OwnerReferences: []metav1.OwnerReference{
				{
					UID: "1234567890",
				},
			},
		},
	}

	kt1 := kafkabetav1.KafkaTopic{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "kt1",
			OwnerReferences: []metav1.OwnerReference{
				{
					UID: "1234567",
				},
			},
		},
	}

	kt2 := kafkabetav1.KafkaTopic{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "kt2",
			OwnerReferences: []metav1.OwnerReference{
				{
					UID: "1234567890",
				},
			},
		},
	}

	ssList := &appsv1.DeploymentList{
		Items: []appsv1.Deployment{dep1, dep2},
	}

	svcList := &corev1.ServiceList{
		Items: []corev1.Service{svc1, svc2},
	}

	rtList := &routev1.RouteList{
		Items: []routev1.Route{rt1, rt2},
	}

	ktList := &kafkabetav1.KafkaTopicList{
		Items: []kafkabetav1.KafkaTopic{kt1, kt2},
	}

	isList := &imgv1.ImageStreamList{
		Items: []imgv1.ImageStream{is1, is2},
	}

	type args struct {
		instance *v1alpha1.KogitoDataIndex
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
				dataIndex,
				test.CreateFakeClientOnOpenShift([]runtime.Object{ssList, svcList, rtList, ktList, isList}, nil, nil),
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
			if len(got) != 5 {
				t.Errorf("GetDeployedResources() got = %v, want %v", got, "5 types of resources")
			}
			if len(got[reflect.TypeOf(appsv1.Deployment{})]) != 1 ||
				got[reflect.TypeOf(appsv1.Deployment{})][0].GetName() != dep1.GetName() {
				t.Errorf("getDeployment() gotDeployment = %v, want %v", got, dep1)
			}
			if len(got[reflect.TypeOf(corev1.Service{})]) != 1 ||
				got[reflect.TypeOf(corev1.Service{})][0].GetName() != svc1.GetName() {
				t.Errorf("getService() gotService = %v, want %v", got, svc1)
			}
			if len(got[reflect.TypeOf(routev1.Route{})]) != 1 ||
				got[reflect.TypeOf(routev1.Route{})][0].GetName() != rt1.GetName() {
				t.Errorf("getRoute() gotRoute = %v, want %v", got, rt1)
			}
			if len(got[reflect.TypeOf(kafkabetav1.KafkaTopic{})]) != 1 ||
				got[reflect.TypeOf(kafkabetav1.KafkaTopic{})][0].GetName() != kt1.GetName() {
				t.Errorf("getKafkaTopic() gotKafkaTopic = %v, want %v", got, kt1)
			}
			if len(got[reflect.TypeOf(imgv1.ImageStream{})]) != 1 ||
				got[reflect.TypeOf(imgv1.ImageStream{})][0].GetName() != is1.GetName() {
				t.Errorf("getImageStream() getImageStream = %v, want %v", got, is1)
			}
		})
	}
}
