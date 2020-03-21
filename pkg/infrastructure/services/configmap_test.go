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

package services

import (
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"testing"
)

func TestExcludeAppPropConfigMapFromResource(t *testing.T) {
	type args struct {
		name        string
		resourceMap map[reflect.Type][]resource.KubernetesResource
	}
	tests := []struct {
		name   string
		args   args
		length int
	}{
		{
			"No Element",
			args{
				name:        "test",
				resourceMap: map[reflect.Type][]resource.KubernetesResource{},
			},
			0,
		},
		{
			"One Element",
			args{
				name: "test",
				resourceMap: map[reflect.Type][]resource.KubernetesResource{
					reflect.TypeOf(corev1.ConfigMap{}): {
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test" + appPropConfigMapSuffix,
							},
						},
					},
				},
			},
			0,
		},
		{
			"First Element",
			args{
				name: "test",
				resourceMap: map[reflect.Type][]resource.KubernetesResource{
					reflect.TypeOf(corev1.ConfigMap{}): {
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test" + appPropConfigMapSuffix,
							},
						},
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test",
							},
						},
					},
				},
			},
			1,
		},
		{
			"Last Element",
			args{
				name: "test",
				resourceMap: map[reflect.Type][]resource.KubernetesResource{
					reflect.TypeOf(corev1.ConfigMap{}): {
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test",
							},
						},
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test" + appPropConfigMapSuffix,
							},
						},
					},
				},
			},
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ExcludeAppPropConfigMapFromResource(tt.args.name, tt.args.resourceMap)
			if tt.length != len(tt.args.resourceMap[reflect.TypeOf(corev1.ConfigMap{})]) {
				t.Errorf("ExcludeAppPropConfigMapFromResource() error = %v, wantErr %v", tt.args.resourceMap, tt.length)
				return
			}
		})
	}
}

func TestGetAppPropConfigMapContentHash(t *testing.T) {
	serviceName := "test"
	serviceNamespace := "test"
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName + appPropConfigMapSuffix,
			Namespace: serviceNamespace,
		},
		Data: map[string]string{
			appPropFileName: defaultAppPropContent,
		},
	}

	type args struct {
		name      string
		namespace string
		cli       *client.Client
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   *corev1.ConfigMap
		wantErr bool
	}{
		{
			"New ConfigMap",
			args{
				serviceName,
				serviceNamespace,
				test.CreateFakeClientOnOpenShift([]runtime.Object{}, nil, nil),
			},
			"d41d8cd98f00b204e9800998ecf8427e",
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      serviceName + appPropConfigMapSuffix,
					Namespace: serviceNamespace,
				},
				Data: map[string]string{
					appPropFileName: defaultAppPropContent,
				},
			},
			false,
		},
		{
			"No New ConfigMap",
			args{
				serviceName,
				serviceNamespace,
				test.CreateFakeClientOnOpenShift([]runtime.Object{cm}, nil, nil),
			},
			"d41d8cd98f00b204e9800998ecf8427e",
			nil,
			false,
		},
		{
			"No Data",
			args{
				serviceName,
				serviceNamespace,
				test.CreateFakeClientOnOpenShift([]runtime.Object{&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      serviceName + appPropConfigMapSuffix,
						Namespace: serviceNamespace,
					},
					Data: map[string]string{},
				}}, nil, nil),
			},
			"d41d8cd98f00b204e9800998ecf8427e",
			cm,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := GetAppPropConfigMapContentHash(tt.args.name, tt.args.namespace, tt.args.cli)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAppPropConfigMapContentHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetAppPropConfigMapContentHash() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("GetAppPropConfigMapContentHash() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
