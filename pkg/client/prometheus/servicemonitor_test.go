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

package prometheus

import (
	"reflect"
	"testing"

	monv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	monfake "github.com/coreos/prometheus-operator/pkg/client/versioned/fake"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_serviceMonitor_List(t *testing.T) {
	sm1 := monv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sm1",
			Namespace: "test",
		},
	}

	sm2 := monv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sm2",
			Namespace: "test",
		},
	}

	sm3 := monv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sm3",
			Namespace: "test1",
		},
	}

	type fields struct {
		client *client.Client
	}
	type args struct {
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *monv1.ServiceMonitorList
		wantErr bool
	}{
		{
			"ListServiceMonitor",
			fields{
				client: &client.Client{
					ControlCli:    fake.NewFakeClient(),
					PrometheusCli: monfake.NewSimpleClientset(&sm1, &sm2, &sm3).MonitoringV1(),
				},
			},
			args{
				"test",
			},
			&monv1.ServiceMonitorList{
				Items: []*monv1.ServiceMonitor{
					&sm1,
					&sm2,
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &serviceMonitor{
				client: tt.fields.client,
			}
			got, err := s.List(tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("List() got = %v, want %v", got, tt.want)
			}
		})
	}
}
