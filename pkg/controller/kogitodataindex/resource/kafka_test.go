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
	"reflect"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestIsKafkaServerURIResolved(t *testing.T) {
	ns := t.Name()

	kafka := kafkabetav1.Kafka{
		ObjectMeta: v1.ObjectMeta{
			Name:      "kafka",
			Namespace: ns,
		},
		Spec: kafkabetav1.KafkaSpec{
			Kafka: kafkabetav1.KafkaClusterSpec{
				Replicas: 1,
			},
		},
	}

	kafkaList := &kafkabetav1.KafkaList{
		Items: []kafkabetav1.Kafka{kafka},
	}

	cli := test.CreateFakeClient([]runtime.Object{kafkaList}, nil, nil)

	type args struct {
		instance *v1alpha1.KogitoDataIndex
		client   *client.Client
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			"URI",
			args{
				&v1alpha1.KogitoDataIndex{
					Spec: v1alpha1.KogitoDataIndexSpec{
						KafkaMeta: v1alpha1.KafkaMeta{
							KafkaProperties: v1alpha1.KafkaConnectionProperties{
								ExternalURI: "kafka:9092",
							},
						},
					},
				},
				cli,
			},
			true,
			false,
		},
		{
			"InstanceURI",
			args{
				&v1alpha1.KogitoDataIndex{
					Spec: v1alpha1.KogitoDataIndexSpec{
						KafkaMeta: v1alpha1.KafkaMeta{
							KafkaProperties: v1alpha1.KafkaConnectionProperties{
								Instance: "kafka",
							},
						},
					},
				},
				cli,
			},
			true,
			false,
		},
		{
			"NoMatchingInstance",
			args{
				&v1alpha1.KogitoDataIndex{
					Spec: v1alpha1.KogitoDataIndexSpec{
						KafkaMeta: v1alpha1.KafkaMeta{
							KafkaProperties: v1alpha1.KafkaConnectionProperties{
								Instance: "kafka1",
							},
						},
					},
				},
				cli,
			},
			false,
			false,
		},
		{
			"NoServer",
			args{
				&v1alpha1.KogitoDataIndex{},
				test.CreateFakeClient(nil, nil, nil),
			},
			false,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsKafkaServerURIResolved(tt.args.instance, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsKafkaServerURIResolved() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsKafkaServerURIResolved() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getKafkaInstance(t *testing.T) {
	ns := t.Name()

	kafka := kafkabetav1.Kafka{
		TypeMeta: v1.TypeMeta{Kind: "Kafka", APIVersion: "kafka.strimzi.io/v1beta1"},
		ObjectMeta: v1.ObjectMeta{
			Name:      "kafka",
			Namespace: ns,
		},
		Spec: kafkabetav1.KafkaSpec{
			Kafka: kafkabetav1.KafkaClusterSpec{
				Replicas: 1,
			},
		},
	}

	kafkaList := &kafkabetav1.KafkaList{
		Items: []kafkabetav1.Kafka{kafka},
	}

	cli := test.CreateFakeClient([]runtime.Object{kafkaList}, nil, nil)

	type args struct {
		kafka     v1alpha1.KafkaConnectionProperties
		namespace string
		client    *client.Client
	}
	tests := []struct {
		name    string
		args    args
		want    *kafkabetav1.Kafka
		wantErr bool
	}{
		{
			"WithInstanceName",
			args{
				v1alpha1.KafkaConnectionProperties{
					Instance: "kafka",
				},
				ns,
				cli,
			},
			&kafka,
			false,
		},
		{
			"WithoutInstanceName",
			args{
				v1alpha1.KafkaConnectionProperties{},
				ns,
				cli,
			},
			nil,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getKafkaInstance(tt.args.kafka, tt.args.namespace, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("getKafkaInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getKafkaInstance() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getKafkaServerReplicas(t *testing.T) {
	ns := t.Name()

	kafka := kafkabetav1.Kafka{
		ObjectMeta: v1.ObjectMeta{
			Name:      "kafka",
			Namespace: ns,
		},
		Spec: kafkabetav1.KafkaSpec{
			Kafka: kafkabetav1.KafkaClusterSpec{
				Replicas: 1,
			},
		},
	}

	kafkaList := &kafkabetav1.KafkaList{
		Items: []kafkabetav1.Kafka{kafka},
	}

	cli := test.CreateFakeClient([]runtime.Object{kafkaList}, nil, nil)

	type args struct {
		kafkaProp v1alpha1.KafkaConnectionProperties
		namespace string
		client    *client.Client
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   int32
		wantErr bool
	}{
		{
			"WithInstance",
			args{
				v1alpha1.KafkaConnectionProperties{
					Instance: "kafka",
				},
				ns,
				cli,
			},
			"kafka",
			1,
			false,
		},
		{
			"AnyInstance",
			args{
				v1alpha1.KafkaConnectionProperties{},
				ns,
				cli,
			},
			"",
			0,
			true,
		},
		{
			"NoInstance",
			args{
				v1alpha1.KafkaConnectionProperties{},
				ns,
				test.CreateFakeClient(nil, nil, nil),
			},
			"",
			0,
			true,
		},
		{
			"WithExternalURI",
			args{
				v1alpha1.KafkaConnectionProperties{
					ExternalURI: "kafka:9092",
				},
				ns,
				test.CreateFakeClient(nil, nil, nil),
			},
			"",
			0,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := getKafkaServerReplicas(tt.args.kafkaProp, tt.args.namespace, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("getKafkaServerReplicas() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getKafkaServerReplicas() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getKafkaServerReplicas() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_getKafkaServerURI(t *testing.T) {
	ns := t.Name()

	kafka := kafkabetav1.Kafka{
		ObjectMeta: v1.ObjectMeta{
			Name:      "kafka",
			Namespace: ns,
		},
		Spec: kafkabetav1.KafkaSpec{
			Kafka: kafkabetav1.KafkaClusterSpec{
				Replicas: 1,
			},
		},
		Status: kafkabetav1.KafkaStatus{
			Listeners: []kafkabetav1.ListenerStatus{
				{
					Type: "plain",
					Addresses: []kafkabetav1.ListenerAddress{
						{
							Host: "kafka",
							Port: 9092,
						},
					},
				},
			},
		},
	}

	kafkaList := &kafkabetav1.KafkaList{
		Items: []kafkabetav1.Kafka{kafka},
	}

	cli := test.CreateFakeClient([]runtime.Object{kafkaList}, nil, nil)

	type args struct {
		kafkaProp v1alpha1.KafkaConnectionProperties
		namespace string
		client    *client.Client
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"WithExternalURI",
			args{
				v1alpha1.KafkaConnectionProperties{
					ExternalURI: "kafka:9092",
				},
				ns,
				cli,
			},
			"kafka:9092",
			false,
		},
		{
			"WithInstance",
			args{
				v1alpha1.KafkaConnectionProperties{
					Instance: "kafka",
				},
				ns,
				cli,
			},
			"kafka:9092",
			false,
		},
		{
			"AnyInstance",
			args{
				v1alpha1.KafkaConnectionProperties{},
				ns,
				cli,
			},
			"",
			true,
		},
		{
			"NoInstance",
			args{
				v1alpha1.KafkaConnectionProperties{},
				ns,
				test.CreateFakeClient(nil, nil, nil),
			},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getKafkaServerURI(tt.args.kafkaProp, tt.args.namespace, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("getKafkaServerURI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getKafkaServerURI() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newKafkaTopic(t *testing.T) {
	type args struct {
		topicName     string
		kafkaName     string
		kafkaReplicas int32
		namespace     string
	}
	tests := []struct {
		name string
		args args
		want *kafkabetav1.KafkaTopic
	}{
		{
			"NewKafkaTopic",
			args{
				topicName:     "topic1",
				kafkaName:     "kafka1",
				kafkaReplicas: 3,
				namespace:     "test",
			},
			&kafkabetav1.KafkaTopic{
				ObjectMeta: v1.ObjectMeta{
					Name:      "topic1",
					Namespace: "test",
					Labels: map[string]string{
						kafkaClusterLabel: "kafka1",
					},
				},
				Spec: kafkabetav1.KafkaTopicSpec{
					Replicas:   3,
					Partitions: 10,
					Config: map[string]string{
						kafkaTopicConfigRetentionKey: kafkaTopicConfigRetentionValue,
						kafkaTopicConfigSegmentKey:   kafkaTopicConfigSegmentValue,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newKafkaTopic(tt.args.topicName, tt.args.kafkaName, tt.args.kafkaReplicas, tt.args.namespace); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newKafkaTopic() = %v, want %v", got, tt.want)
			}
		})
	}
}
