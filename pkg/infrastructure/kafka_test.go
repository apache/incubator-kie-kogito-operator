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

package infrastructure

import (
	"github.com/kiegroup/kogito-cloud-operator/api/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/external/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"testing"
)

func Test_getKafkaInstanceWithName(t *testing.T) {
	ns := t.Name()

	kafka := &v1beta1.Kafka{
		TypeMeta: v1.TypeMeta{Kind: "Kafka", APIVersion: "kafka.strimzi.io/v1beta1"},
		ObjectMeta: v1.ObjectMeta{
			Name:      "kafka",
			Namespace: ns,
		},
		Spec: v1beta1.KafkaSpec{
			Kafka: v1beta1.KafkaClusterSpec{
				Replicas: 1,
			},
		},
	}

	cli := test.CreateFakeClient([]runtime.Object{kafka}, nil, nil)

	type args struct {
		name      string
		namespace string
		client    *client.Client
	}
	tests := []struct {
		name    string
		args    args
		want    *v1beta1.Kafka
		wantErr bool
	}{
		{
			"KafkaInstanceExists",
			args{
				"kafka",
				ns,
				cli,
			},
			kafka,
			false,
		},
		{
			"KafkaInstanceNotExists",
			args{
				"kafka1",
				ns,
				cli,
			},
			nil,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetKafkaInstanceWithName(tt.args.name, tt.args.namespace, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("getKafkaInstanceWithName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getKafkaInstanceWithName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resolveKafkaServerReplicas(t *testing.T) {
	type args struct {
		kafka *v1beta1.Kafka
	}
	tests := []struct {
		name string
		args args
		want int32
	}{
		{
			"NoReplicas",
			args{
				nil,
			},
			0,
		},
		{
			"DefaultReplicas",
			args{
				&v1beta1.Kafka{},
			},
			1,
		},
		{
			"ResolveReplicas",
			args{
				&v1beta1.Kafka{
					Spec: v1beta1.KafkaSpec{
						Kafka: v1beta1.KafkaClusterSpec{
							Replicas: 2,
						},
					},
				},
			},
			2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResolveKafkaServerReplicas(tt.args.kafka); got != tt.want {
				t.Errorf("resolveKafkaServerReplicas() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resolveKafkaServerURI(t *testing.T) {
	type args struct {
		kafka *v1beta1.Kafka
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"ResolveKafkaServerURI",
			args{
				&v1beta1.Kafka{
					Status: v1beta1.KafkaStatus{
						Listeners: []v1beta1.ListenerStatus{
							{
								Type: "tls",
								Addresses: []v1beta1.ListenerAddress{
									{
										Host: "kafka1",
										Port: 9093,
									},
								},
							},
							{
								Type: "plain",
								Addresses: []v1beta1.ListenerAddress{
									{
										Host: "kafka",
										Port: 9092,
									},
								},
							},
						},
					},
				},
			},
			"kafka:9092",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResolveKafkaServerURI(tt.args.kafka); got != tt.want {
				t.Errorf("resolveKafkaServerURI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetReadyKafkaInstanceName(t *testing.T) {
	type args struct {
		cli   *client.Client
		infra *v1alpha1.KogitoInfra
	}
	tests := []struct {
		name      string
		args      args
		wantKafka string
		wantErr   bool
	}{
		{
			"NoStatus",
			args{
				cli:   test.CreateFakeClient(nil, nil, nil),
				infra: &v1alpha1.KogitoInfra{},
			},
			"",
			false,
		},
		{
			"NoInstance",
			args{
				cli: test.CreateFakeClient(nil, nil, nil),
				infra: &v1alpha1.KogitoInfra{
					ObjectMeta: v1.ObjectMeta{
						Namespace: "test",
					},
					Status: v1alpha1.KogitoInfraStatus{
						Kafka: v1alpha1.InfraComponentInstallStatusType{
							Name: "test",
						},
					},
				},
			},
			"",
			false,
		},
		{
			"NotReady",
			args{
				cli: test.CreateFakeClient(
					[]runtime.Object{
						&v1beta1.Kafka{
							ObjectMeta: v1.ObjectMeta{
								Name:      "test",
								Namespace: "test",
							},
							Status: v1beta1.KafkaStatus{},
						},
					},
					nil,
					nil),
				infra: &v1alpha1.KogitoInfra{
					ObjectMeta: v1.ObjectMeta{
						Namespace: "test",
					},
					Status: v1alpha1.KogitoInfraStatus{
						Kafka: v1alpha1.InfraComponentInstallStatusType{
							Name: "test",
						},
					},
				},
			},
			"",
			false,
		},
		{
			"Ready",
			args{
				cli: test.CreateFakeClient(
					[]runtime.Object{
						&v1beta1.Kafka{
							ObjectMeta: v1.ObjectMeta{
								Name:      "test",
								Namespace: "test",
							},
							Status: v1beta1.KafkaStatus{
								Listeners: []v1beta1.ListenerStatus{
									{
										Type: "tls",
										Addresses: []v1beta1.ListenerAddress{
											{
												Host: "kafka1",
												Port: 9093,
											},
										},
									},
									{
										Type: "plain",
										Addresses: []v1beta1.ListenerAddress{
											{
												Host: "kafka",
												Port: 9092,
											},
										},
									},
								},
							},
						},
					},
					nil,
					nil),
				infra: &v1alpha1.KogitoInfra{
					ObjectMeta: v1.ObjectMeta{
						Namespace: "test",
					},
					Status: v1alpha1.KogitoInfraStatus{
						Kafka: v1alpha1.InfraComponentInstallStatusType{
							Name: "test",
						},
					},
				},
			},
			"test",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKafka, err := GetReadyKafkaInstanceName(tt.args.cli, tt.args.infra)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetReadyKafkaInstanceName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotKafka != tt.wantKafka {
				t.Errorf("GetReadyKafkaInstanceName() gotKafka = %v, want %v", gotKafka, tt.wantKafka)
			}
		})
	}
}
