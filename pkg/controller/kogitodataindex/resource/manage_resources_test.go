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
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"

	"github.com/stretchr/testify/assert"
)

func Test_ManageResources_WhenKafkaURIIsChanged(t *testing.T) {
	serviceuri := "myserviceuri:9092"
	instance := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-data-index",
			Namespace: "test",
		},
		Spec: v1alpha1.KogitoDataIndexSpec{
			Replicas: 1,
		},
	}
	cli := test.CreateFakeClient(nil, nil, nil)
	secret := &corev1.Secret{}
	statefulset, err := newStatefulset(instance, secret, serviceuri, cli)
	assert.NoError(t, err)
	cli = test.CreateFakeClient([]runtime.Object{instance, statefulset, secret}, nil, nil)

	err = ManageResources(instance, &KogitoDataIndexResources{StatefulSet: statefulset}, cli)
	assert.Error(t, err)

	instance.Spec.Kafka = v1alpha1.KafkaConnectionProperties{ExternalURI: serviceuri}
	err = ManageResources(instance, &KogitoDataIndexResources{StatefulSet: statefulset}, cli)
	assert.NoError(t, err)
	for _, kafkaKey := range managedKafkaKeys {
		assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: kafkaKey, Value: serviceuri})
	}
}

func Test_ManageResources_WhenWeChangeInfinispanVars(t *testing.T) {
	instance := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-data-index",
			Namespace: "test",
		},
		Spec: v1alpha1.KogitoDataIndexSpec{
			Replicas: 1,
			InfinispanMeta: v1alpha1.InfinispanMeta{InfinispanProperties: v1alpha1.InfinispanConnectionProperties{
				Credentials: v1alpha1.SecretCredentialsType{
					SecretName:  "infinispan-secret",
					UsernameKey: "user",
					PasswordKey: "pass",
				}},
			},
		},
	}
	kafkaList := &kafkabetav1.KafkaList{
		Items: []kafkabetav1.Kafka{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kafka",
					Namespace: "test",
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
			},
		},
	}
	userBytes := []byte("developer")
	passBytes := []byte("developer")
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "infinispan-secret", Namespace: "test"},
		Data: map[string][]byte{
			"user": userBytes, "pass": passBytes,
		},
	}
	serviceuri := "kafka:9092"
	statefulset, err := newStatefulset(instance, secret, serviceuri, test.CreateFakeClient(nil, nil, nil))
	assert.NoError(t, err)
	cli := test.CreateFakeClient([]runtime.Object{instance, statefulset, secret, kafkaList}, nil, nil)

	valueFromUsername := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: secret.Name},
			Key:                  infrastructure.InfinispanSecretUsernameKey,
		},
	}
	valueFromPassword := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: secret.Name},
			Key:                  infrastructure.InfinispanSecretPasswordKey,
		},
	}

	// reconcile
	err = ManageResources(instance, &KogitoDataIndexResources{StatefulSet: statefulset}, cli)
	assert.NoError(t, err)
	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: "INFINISPAN_USERNAME", ValueFrom: valueFromUsername})
	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: "INFINISPAN_PASSWORD", ValueFrom: valueFromPassword})
	assert.NotContains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: "INFINISPAN_AUTHREALM", Value: "default"})

	// let's change
	instance.Spec.InfinispanProperties.AuthRealm = "default"
	instance.Spec.InfinispanProperties.URI = "myservice:11222"
	err = ManageResources(instance, &KogitoDataIndexResources{StatefulSet: statefulset}, cli)
	assert.NoError(t, err)
	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: "INFINISPAN_AUTHREALM", Value: "default"})
	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: "INFINISPAN_CLIENT_SERVER_LIST", Value: "myservice:11222"})
}

func Test_ManageResources_WhenTheresAMixOnEnvs(t *testing.T) {
	instance := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-data-index",
			Namespace: "test",
		},
		Spec: v1alpha1.KogitoDataIndexSpec{
			Replicas: 1,
			Env: map[string]string{
				"key1":                   "value1",
				"KOGITO_PROTOBUF_FOLDER": "/any/invalid/path",
			},
			InfinispanMeta: v1alpha1.InfinispanMeta{InfinispanProperties: v1alpha1.InfinispanConnectionProperties{
				Credentials: v1alpha1.SecretCredentialsType{
					SecretName:  "infinispan-secret",
					UsernameKey: "user",
					PasswordKey: "pass",
				},
			}},
		},
	}
	kafkaList := &kafkabetav1.KafkaList{
		Items: []kafkabetav1.Kafka{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kafka",
					Namespace: "test",
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
			},
		},
	}
	userBytes := []byte("developer")
	passBytes := []byte("developer")
	serviceuri := "kafka:9092"
	statefulset, err := newStatefulset(instance, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: t.Name()},
		Data: map[string][]byte{
			"user": userBytes, "pass": passBytes,
		}}, serviceuri, test.CreateFakeClient(nil, nil, nil))
	assert.NoError(t, err)
	cli := test.CreateFakeClient([]runtime.Object{instance, statefulset, kafkaList}, nil, nil)

	// make sure that defaults were inserted
	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "KOGITO_PROTOBUF_FOLDER",
		Value: "",
	})

	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "INFINISPAN_USEAUTH",
		Value: "true",
	})

	valueFromUsername := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: t.Name()},
			Key:                  infrastructure.InfinispanSecretUsernameKey,
		},
	}
	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:      "INFINISPAN_USERNAME",
		ValueFrom: valueFromUsername,
	})

	// change the spec, adding one more key
	instance.Spec.Env = map[string]string{
		"key2":                   "value2",
		"key1":                   "value1",
		"KOGITO_PROTOBUF_FOLDER": "/any/invalid/path",
	}

	// reconcile
	err = ManageResources(instance, &KogitoDataIndexResources{StatefulSet: statefulset}, cli)
	assert.NoError(t, err)

	// check the result
	_, err = kubernetes.ResourceC(cli).Fetch(statefulset)
	assert.NoError(t, err)
	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "KOGITO_PROTOBUF_FOLDER",
		Value: "",
	})
	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "key1",
		Value: "value1",
	})
	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "key2",
		Value: "value2",
	})

	// change the spec, removing  one key
	instance.Spec.Env = map[string]string{
		"key1": "value1",
	}
	err = ManageResources(instance, &KogitoDataIndexResources{StatefulSet: statefulset}, cli)
	assert.NoError(t, err)

	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "KOGITO_PROTOBUF_FOLDER",
		Value: "", //we don't have a configMap associated to it
	})
	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "key1",
		Value: "value1",
	})
	assert.NotContains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "key2",
		Value: "value2",
	})
}

func Test_ensureKafka(t *testing.T) {
	ns := t.Name()

	instance := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-data-index",
			Namespace: ns,
		},
		Spec: v1alpha1.KogitoDataIndexSpec{
			Kafka: v1alpha1.KafkaConnectionProperties{
				ExternalURI: "kafka:9092",
			},
		},
	}

	statefulSet := &appsv1.StatefulSet{
		Spec: appsv1.StatefulSetSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Env: []v1.EnvVar{
								{
									Name:  kafkaEnvKeyProcessInstancesServer,
									Value: "kafka1:9092",
								},
								{
									Name:  kafkaEnvKeyUserTaskInstanceServer,
									Value: "kafka2:9092",
								},
								{
									Name:  kafkaEnvKeyProcessDomainServer,
									Value: "kafka3:9092",
								},
								{
									Name:  kafkaEnvKeyUserTaskDomainServer,
									Value: "kafka4:9092",
								},
							},
						},
					},
				},
			},
		},
	}

	cli := test.CreateFakeClient([]runtime.Object{instance, statefulSet}, nil, nil)

	type args struct {
		instance    *v1alpha1.KogitoDataIndex
		statefulset *appsv1.StatefulSet
		client      *client.Client
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			"EnsureKafka",
			args{
				instance,
				statefulSet,
				cli,
			},
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ensureKafka(tt.args.instance, tt.args.statefulset, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("ensureKafka() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ensureKafka() got = %v, want %v", got, tt.want)
				return
			}
			assert.Equal(t, len(statefulSet.Spec.Template.Spec.Containers[0].Env), 4)
			for _, kafkaEnv := range managedKafkaKeys {
				for _, env := range statefulSet.Spec.Template.Spec.Containers[0].Env {
					if kafkaEnv == env.Name {
						assert.Equal(t, env.Value, instance.Spec.Kafka.ExternalURI)
						break
					}
				}
			}
		})
	}
}

func Test_ensureKafkaTopics(t *testing.T) {
	ns := t.Name()

	kafka := &kafkabetav1.Kafka{
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

	kafkaTopic1 := kafkabetav1.KafkaTopic{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kafkaTopic1",
			Namespace: ns,
			Labels: map[string]string{
				kafkaClusterLabel: "kafka1",
			},
		},
		Spec: kafkabetav1.KafkaTopicSpec{
			Replicas: 2,
		},
	}

	kafkaTopic2 := kafkabetav1.KafkaTopic{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kafkaTopic2",
			Namespace: ns,
			Labels: map[string]string{
				kafkaClusterLabel: "kafka2",
			},
		},
		Spec: kafkabetav1.KafkaTopicSpec{
			Replicas: 3,
		},
	}

	instance := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-data-index",
			Namespace: ns,
		},
		Spec: v1alpha1.KogitoDataIndexSpec{
			Kafka: v1alpha1.KafkaConnectionProperties{
				Instance: kafka.Name,
			},
		},
	}

	cli := test.CreateFakeClient([]runtime.Object{instance, kafka, &kafkaTopic1, &kafkaTopic2}, nil, nil)

	type args struct {
		instance    *v1alpha1.KogitoDataIndex
		kafkaTopics []kafkabetav1.KafkaTopic
		client      *client.Client
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"EnsureKafkaTopics",
			args{
				instance,
				[]kafkabetav1.KafkaTopic{
					kafkaTopic1,
					kafkaTopic2,
				},
				cli,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ensureKafkaTopics(tt.args.instance, tt.args.kafkaTopics, tt.args.client); (err != nil) != tt.wantErr {
				t.Errorf("ensureKafkaTopics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, kafkaTopic := range tt.args.kafkaTopics {
				exists, err := kubernetes.ResourceC(cli).Fetch(&kafkaTopic)
				assert.True(t, exists)
				assert.NoError(t, err)
				assert.Equal(t, kafkaTopic.Labels[kafkaClusterLabel], kafka.Name)
				assert.Equal(t, kafkaTopic.Spec.Replicas, kafka.Spec.Kafka.Replicas)
			}
		})
	}
}
