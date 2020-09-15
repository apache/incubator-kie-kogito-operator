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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"testing"
)

func TestGetAppPropConfigMapContentHash(t *testing.T) {
	service := &v1alpha1.KogitoRuntime{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
	}
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      service.Name + appPropConfigMapSuffix,
			Namespace: service.Namespace,
		},
		Data: map[string]string{
			appPropFileName: defaultAppPropContent,
		},
	}

	type args struct {
		instance v1alpha1.KogitoService
		appProps map[string]string
		cli      *client.Client
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   *corev1.ConfigMap
		wantErr bool
	}{
		{
			"No ConfigMap, empty appProps",
			args{
				service,
				map[string]string{},
				test.CreateFakeClientOnOpenShift([]runtime.Object{}, nil, nil),
			},
			"d41d8cd98f00b204e9800998ecf8427e",
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      service.Name + appPropConfigMapSuffix,
					Namespace: service.Namespace,
				},
				Data: map[string]string{
					appPropFileName: defaultAppPropContent,
				},
			},
			false,
		},
		{
			"No ConfigMap, nil appProps",
			args{
				service,
				nil,
				test.CreateFakeClientOnOpenShift([]runtime.Object{}, nil, nil),
			},
			"d41d8cd98f00b204e9800998ecf8427e",
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      service.Name + appPropConfigMapSuffix,
					Namespace: service.Namespace,
				},
				Data: map[string]string{
					appPropFileName: defaultAppPropContent,
				},
			},
			false,
		},
		{
			"Default ConfigMap, empty appProps",
			args{
				service,
				map[string]string{},
				test.CreateFakeClientOnOpenShift([]runtime.Object{cm}, nil, nil),
			},
			"d41d8cd98f00b204e9800998ecf8427e",
			cm,
			false,
		},
		{
			"Empty ConfigMap, empty appProps",
			args{
				service,
				map[string]string{},
				test.CreateFakeClientOnOpenShift([]runtime.Object{&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      service.Name + appPropConfigMapSuffix,
						Namespace: service.Namespace,
					},
					Data: map[string]string{},
				}}, nil, nil),
			},
			"d41d8cd98f00b204e9800998ecf8427e",
			cm,
			false,
		},
		{
			"No ConfigMap, appProps with data",
			args{
				service,
				map[string]string{
					"test1": "abc",
					"test2": "def",
					"test3": "ghi",
				},
				test.CreateFakeClientOnOpenShift([]runtime.Object{}, nil, nil),
			},
			"bb2bea2d5b08e3d93142da5b17ed2af0",
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      service.Name + appPropConfigMapSuffix,
					Namespace: service.Namespace,
				},
				Data: map[string]string{
					appPropFileName: "\ntest1=abc\ntest2=def\ntest3=ghi",
				},
			},
			false,
		},
		{
			"ConfigMap with data, empty appProps",
			args{
				service,
				map[string]string{},
				test.CreateFakeClientOnOpenShift([]runtime.Object{&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      service.Name + appPropConfigMapSuffix,
						Namespace: service.Namespace,
					},
					Data: map[string]string{
						appPropFileName: "\ntest1=123\ntest2=456\ntest3=789\ntest4=012\ntest5=345",
					},
				}}, nil, nil),
			},
			"be33decf24f9ecf68c0feaeffc3f76f9",
			&corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      service.Name + appPropConfigMapSuffix,
					Namespace: service.Namespace,
				},
				Data: map[string]string{
					appPropFileName: "\ntest1=123\ntest2=456\ntest3=789\ntest4=012\ntest5=345",
				},
			},
			false,
		},
		{
			"ConfigMap with data, appProps with data",
			args{
				service,
				map[string]string{
					"test1": "abc",
					"test2": "def",
					"test3": "ghi",
					"test7": "jkl",
				},
				test.CreateFakeClientOnOpenShift([]runtime.Object{&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      service.Name + appPropConfigMapSuffix,
						Namespace: service.Namespace,
					},
					Data: map[string]string{
						appPropFileName: "\ntest1=123\ntest2=456\ntest3=789\ntest4=012\ntest5=345",
					},
				}}, nil, nil),
			},
			"b641b11398d67b6a80eee92cfdf8fc8c",
			&corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      service.Name + appPropConfigMapSuffix,
					Namespace: service.Namespace,
				},
				Data: map[string]string{
					appPropFileName: "\ntest1=abc\ntest2=def\ntest3=ghi\ntest4=012\ntest5=345\ntest7=jkl",
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := getAppPropConfigMapContentHash(tt.args.instance, tt.args.appProps, tt.args.cli)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAppPropConfigMapContentHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getAppPropConfigMapContentHash() got = %v, want %v", got, tt.want)
				return
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("getAppPropConfigMapContentHash() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_getAppPropsFromConfigMap(t *testing.T) {
	type args struct {
		configMap *corev1.ConfigMap
		exist     bool
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			"Not exist",
			args{
				&corev1.ConfigMap{},
				false,
			},
			map[string]string{},
		},
		{
			"Empty ConfigMap",
			args{
				&corev1.ConfigMap{},
				true,
			},
			map[string]string{},
		},
		{
			"Empty Property",
			args{
				&corev1.ConfigMap{
					Data: map[string]string{
						appPropContentKey: "",
					},
				},
				true,
			},
			map[string]string{},
		},
		{
			"With Properties",
			args{
				&corev1.ConfigMap{
					Data: map[string]string{
						appPropContentKey: "\ntest1=test1\ntest2=test2\ntest3=test3",
					},
				},
				true,
			},
			map[string]string{
				"test1": "test1",
				"test2": "test2",
				"test3": "test3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAppPropsFromConfigMap(tt.args.configMap, tt.args.exist); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getAppPropsFromConfigMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
