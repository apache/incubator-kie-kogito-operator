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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
	"time"
)

func TestDeployKafkaWithKogitoInfra(t *testing.T) {
	type args struct {
		instance  v1alpha1.KafkaAware
		namespace string
		cli       *client.Client
	}
	tests := []struct {
		name             string
		args             args
		wantUpdate       bool
		wantRequeueAfter time.Duration
		wantErr          bool
	}{
		{
			"NoInstance",
			args{
				nil,
				"test",
				test.CreateFakeClient(nil, nil, nil),
			},
			false,
			0,
			false,
		},
		{
			"NoKafkaProperties",
			args{
				&v1alpha1.KafkaMeta{
					KafkaProperties: v1alpha1.KafkaConnectionProperties{},
				},
				"test",
				test.CreateFakeClient(nil, nil, nil),
			},
			true,
			0,
			false,
		},
		{
			"WithKafkaInstance",
			args{
				&v1alpha1.KafkaMeta{
					KafkaProperties: v1alpha1.KafkaConnectionProperties{
						Instance: "test",
					},
				},
				"test",
				test.CreateFakeClient(nil, nil, nil),
			},
			false,
			0,
			false,
		},
		{
			"WithKafkaURI",
			args{
				&v1alpha1.KafkaMeta{
					KafkaProperties: v1alpha1.KafkaConnectionProperties{
						ExternalURI: "test:8080",
					},
				},
				"test",
				test.CreateFakeClient(nil, nil, nil),
			},
			false,
			0,
			false,
		},
		{
			"KafkaCreate",
			args{
				&v1alpha1.KafkaMeta{
					KafkaProperties: v1alpha1.KafkaConnectionProperties{
						UseKogitoInfra: true,
					},
				},
				"test",
				test.CreateFakeClient(nil, nil, nil),
			},
			false,
			time.Second * 10,
			false,
		},
		{
			"KafkaStatusNotReady",
			args{
				&v1alpha1.KafkaMeta{
					KafkaProperties: v1alpha1.KafkaConnectionProperties{
						UseKogitoInfra: true,
					},
				},
				"test",
				test.CreateFakeClient([]runtime.Object{
					&v1beta1.Kafka{
						ObjectMeta: v1.ObjectMeta{
							Name:      "test",
							Namespace: "test",
						},
						Status: v1beta1.KafkaStatus{},
					},
					&v1alpha1.KogitoInfra{
						ObjectMeta: v1.ObjectMeta{
							Namespace: "test",
						},
						Spec: v1alpha1.KogitoInfraSpec{
							InstallKafka: true,
						},
						Status: v1alpha1.KogitoInfraStatus{
							Kafka: v1alpha1.KafkaInstallStatus{
								InfraComponentInstallStatusType: v1alpha1.InfraComponentInstallStatusType{
									Name: "test",
									Condition: []v1alpha1.InstallCondition{
										{
											Type: v1alpha1.ProvisioningInstallConditionType,
										},
									},
								},
							},
						},
					},
				},
					nil,
					nil),
			},
			false,
			time.Second * 10,
			false,
		},
		{
			"KafkaServiceNotReady",
			args{
				&v1alpha1.KafkaMeta{
					KafkaProperties: v1alpha1.KafkaConnectionProperties{
						UseKogitoInfra: true,
					},
				},
				"test",
				test.CreateFakeClient([]runtime.Object{
					&v1beta1.Kafka{
						ObjectMeta: v1.ObjectMeta{
							Name:      "test",
							Namespace: "test",
						},
						Status: v1beta1.KafkaStatus{},
					},
					&v1alpha1.KogitoInfra{
						ObjectMeta: v1.ObjectMeta{
							Namespace: "test",
						},
						Spec: v1alpha1.KogitoInfraSpec{
							InstallKafka: true,
						},
						Status: v1alpha1.KogitoInfraStatus{
							Kafka: v1alpha1.KafkaInstallStatus{
								InfraComponentInstallStatusType: v1alpha1.InfraComponentInstallStatusType{
									Service: "test",
									Name:    "test",
									Condition: []v1alpha1.InstallCondition{
										{
											Type: v1alpha1.SuccessInstallConditionType,
										},
									},
								},
							},
						},
					},
				},
					nil,
					nil),
			},
			false,
			time.Second * 10,
			false,
		},
		{
			"KafkaAlreadyUsed",
			args{
				&v1alpha1.KafkaMeta{
					KafkaProperties: v1alpha1.KafkaConnectionProperties{
						UseKogitoInfra: true,
						Instance:       "test",
					},
				},
				"test",
				test.CreateFakeClient([]runtime.Object{
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
					&v1alpha1.KogitoInfra{
						ObjectMeta: v1.ObjectMeta{
							Namespace: "test",
						},
						Spec: v1alpha1.KogitoInfraSpec{
							InstallKafka: true,
						},
						Status: v1alpha1.KogitoInfraStatus{
							Kafka: v1alpha1.KafkaInstallStatus{
								InfraComponentInstallStatusType: v1alpha1.InfraComponentInstallStatusType{
									Service: "test",
									Name:    "test",
									Condition: []v1alpha1.InstallCondition{
										{
											Type: v1alpha1.SuccessInstallConditionType,
										},
									},
								},
							},
						},
					},
				},
					nil,
					nil),
			},
			false,
			0,
			false,
		},
		{
			"KafkaUse",
			args{
				&v1alpha1.KafkaMeta{
					KafkaProperties: v1alpha1.KafkaConnectionProperties{
						UseKogitoInfra: true,
					},
				},
				"test",
				test.CreateFakeClient([]runtime.Object{
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
					&v1alpha1.KogitoInfra{
						ObjectMeta: v1.ObjectMeta{
							Namespace: "test",
						},
						Spec: v1alpha1.KogitoInfraSpec{
							InstallKafka: true,
						},
						Status: v1alpha1.KogitoInfraStatus{
							Kafka: v1alpha1.KafkaInstallStatus{
								InfraComponentInstallStatusType: v1alpha1.InfraComponentInstallStatusType{
									Service: "test",
									Name:    "test",
									Condition: []v1alpha1.InstallCondition{
										{
											Type: v1alpha1.SuccessInstallConditionType,
										},
									},
								},
							},
						},
					},
				},
					nil,
					nil),
			},
			true,
			0,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUpdate, gotRequeueAfter, err := DeployKafkaWithKogitoInfra(tt.args.instance, tt.args.namespace, tt.args.cli)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeployKafkaWithKogitoInfra() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUpdate != tt.wantUpdate {
				t.Errorf("DeployKafkaWithKogitoInfra() gotUpdate = %v, want %v", gotUpdate, tt.wantUpdate)
			}
			if gotRequeueAfter != tt.wantRequeueAfter {
				t.Errorf("DeployKafkaWithKogitoInfra() gotRequeueAfter = %v, want %v", gotRequeueAfter, tt.wantRequeueAfter)
			}
		})
	}
}
