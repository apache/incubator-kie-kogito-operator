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
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"reflect"
	"testing"

	monv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"

	dockerv10 "github.com/openshift/api/image/docker10"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	discfake "k8s.io/client-go/discovery/fake"
	clienttesting "k8s.io/client-go/testing"
)

func TestNewServiceMonitor(t *testing.T) {
	port := intstr.FromInt(8080)

	type args struct {
		kogitoApp   *v1alpha1.KogitoApp
		dockerImage *dockerv10.DockerImage
		service     *corev1.Service
		client      *client.Client
	}
	tests := []struct {
		name    string
		args    args
		want    *monv1.ServiceMonitor
		wantErr bool
	}{
		{
			"TestNewServiceMonitor",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
				},
				&dockerv10.DockerImage{
					Config: &dockerv10.DockerConfig{
						Labels: map[string]string{
							framework.LabelPrometheusScrape: "true",
							framework.LabelPrometheusPath:   "/ms",
							framework.LabelPrometheusPort:   "8080",
							framework.LabelPrometheusScheme: "https",
						},
					},
				},
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"test": "test",
						},
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(),
					Discovery:  test.CreateFakeDiscoveryClient(false),
				},
			},
			&monv1.ServiceMonitor{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ServiceMonitor",
					APIVersion: "monitoring.coreos.com/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
					Labels: map[string]string{
						"app": "test",
					},
					Annotations: defaultAnnotations,
				},
				Spec: monv1.ServiceMonitorSpec{
					NamespaceSelector: monv1.NamespaceSelector{
						MatchNames: []string{
							"test",
						},
					},
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"test": "test",
						},
					},
					Endpoints: []monv1.Endpoint{
						{
							TargetPort: &port,
							Path:       "/ms",
							Scheme:     "https",
						},
					},
				},
			},
			false,
		},

		{
			"TestNewServiceMonitorDefault",
			args{
				&v1alpha1.KogitoApp{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
				},
				&dockerv10.DockerImage{
					Config: &dockerv10.DockerConfig{
						Labels: map[string]string{
							framework.LabelPrometheusScrape: "true",
						},
					},
				},
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"test": "test",
						},
					},
				},
				&client.Client{
					ControlCli: fake.NewFakeClient(),
					Discovery:  test.CreateFakeDiscoveryClient(false),
				},
			},
			&monv1.ServiceMonitor{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ServiceMonitor",
					APIVersion: "monitoring.coreos.com/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
					Labels: map[string]string{
						"app": "test",
					},
					Annotations: defaultAnnotations,
				},
				Spec: monv1.ServiceMonitorSpec{
					NamespaceSelector: monv1.NamespaceSelector{
						MatchNames: []string{
							"test",
						},
					},
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"test": "test",
						},
					},
					Endpoints: []monv1.Endpoint{
						{
							Port:   "http",
							Path:   "/metrics",
							Scheme: "http",
						},
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newServiceMonitor(tt.args.kogitoApp, tt.args.dockerImage, tt.args.service, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("newServiceMonitor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newServiceMonitor() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isPrometheusOperatorReady(t *testing.T) {
	type args struct {
		client *client.Client
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"PrometheusOperatorReady",
			args{
				&client.Client{
					ControlCli: fake.NewFakeClient(),
					Discovery:  test.CreateFakeDiscoveryClient(false),
				},
			},
			true,
		},
		{
			"PrometheusOperatorNotReady",
			args{
				&client.Client{
					ControlCli: fake.NewFakeClient(),
					Discovery: &discfake.FakeDiscovery{
						Fake: &clienttesting.Fake{},
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPrometheusOperatorReady(tt.args.client)
			if got != tt.want {
				t.Errorf("isPrometheusOperatorReady() got = %v, want %v", got, tt.want)
			}
		})
	}
}
