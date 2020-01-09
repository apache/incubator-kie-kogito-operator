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
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"

	"github.com/stretchr/testify/assert"
)

func Test_createKafkaTopic(t *testing.T) {
	ns := t.Name()

	instance := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-data-index",
			Namespace: ns,
		},
	}

	kafka := kafkabetav1.Kafka{
		ObjectMeta: metav1.ObjectMeta{
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
		Items: []kafkabetav1.Kafka{
			kafka,
		},
	}

	client := test.CreateFakeClient([]runtime.Object{instance, kafkaList}, nil, nil)

	factory := &kogitoDataIndexResourcesFactory{
		Factory: framework.Factory{
			Context: &framework.FactoryContext{
				Client: client,
			},
		},
		Resources:       &KogitoDataIndexResources{},
		KogitoDataIndex: instance,
	}

	type args struct {
		f *kogitoDataIndexResourcesFactory
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"CreateKafkaTopics",
			args{
				factory,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createKafkaTopic(tt.args.f)

			assert.NoError(t, got.Error)
			assert.True(t, got.Resources.KafkaTopicStatus.New)
			assert.Equal(t, len(got.Resources.KafkaTopics), len(kafkaTopicNames))

			for _, kafkaTopicName := range kafkaTopicNames {
				for _, kafkaTopic := range got.Resources.KafkaTopics {
					if kafkaTopic.Name == kafkaTopicName {
						assert.Equal(t, kafkaTopic.Namespace, instance.Namespace)
						assert.Equal(t, kafkaTopic.Labels[kafkaClusterLabel], kafka.Name)
						assert.Equal(t, kafkaTopic.Spec.Replicas, kafka.Spec.Kafka.Replicas)
						break
					}
				}
			}
		})
	}
}
