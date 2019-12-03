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
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/infinispan"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"

	"github.com/stretchr/testify/assert"

	imgv1 "github.com/openshift/api/image/v1"
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
	cm := newProtobufConfigMap(instance)
	secret := &corev1.Secret{}
	statefulset := newStatefulset(instance, cm, secret, serviceuri)
	client := test.CreateFakeClient([]runtime.Object{instance, cm, statefulset, secret}, nil, nil)

	err := ManageResources(instance, &KogitoDataIndexResources{StatefulSet: statefulset, ProtoBufConfigMap: cm}, client)
	assert.Error(t, err)

	instance.Spec.Kafka = v1alpha1.KafkaConnectionProperties{ExternalURI: serviceuri}
	err = ManageResources(instance, &KogitoDataIndexResources{StatefulSet: statefulset, ProtoBufConfigMap: cm}, client)
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
			Infinispan: v1alpha1.InfinispanConnectionProperties{
				Credentials: v1alpha1.SecretCredentialsType{
					SecretName:  "infinispan-secret",
					UsernameKey: "user",
					PasswordKey: "pass",
				},
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
					KafkaClusterSpec: kafkabetav1.KafkaClusterSpec{
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
	cm := newProtobufConfigMap(instance)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "infinispan-secret", Namespace: "test"},
		Data: map[string][]byte{
			"user": []byte(userBytes), "pass": []byte(passBytes),
		},
	}
	serviceuri := "kafka:9092"
	statefulset := newStatefulset(instance, cm, secret, serviceuri)
	client := test.CreateFakeClient([]runtime.Object{instance, cm, statefulset, secret, kafkaList}, nil, nil)

	valueFromUsername := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: secret.Name},
			Key:                  infinispan.SecretUsernameKey,
		},
	}
	valueFromPassword := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: secret.Name},
			Key:                  infinispan.SecretPasswordKey,
		},
	}

	// reconcile
	err := ManageResources(instance, &KogitoDataIndexResources{StatefulSet: statefulset, ProtoBufConfigMap: cm}, client)
	assert.NoError(t, err)
	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: infinispanEnvKeyUsername, ValueFrom: valueFromUsername})
	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: infinispanEnvKeyPassword, ValueFrom: valueFromPassword})
	assert.NotContains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: infinispanEnvKeyAuthRealm, Value: "default"})

	// let's change
	instance.Spec.Infinispan.AuthRealm = "default"
	instance.Spec.Infinispan.ServiceURI = "myservice:11222"
	err = ManageResources(instance, &KogitoDataIndexResources{StatefulSet: statefulset, ProtoBufConfigMap: cm}, client)
	assert.NoError(t, err)
	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: infinispanEnvKeyAuthRealm, Value: "default"})
	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: interfaceEnvKeyServiceURI, Value: "myservice:11222"})
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
			Infinispan: v1alpha1.InfinispanConnectionProperties{
				Credentials: v1alpha1.SecretCredentialsType{
					SecretName:  "infinispan-secret",
					UsernameKey: "user",
					PasswordKey: "pass",
				},
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
					KafkaClusterSpec: kafkabetav1.KafkaClusterSpec{
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
	cm := newProtobufConfigMap(instance)
	serviceuri := "kafka:9092"
	statefulset := newStatefulset(instance, cm, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: t.Name()},
		Data: map[string][]byte{
			"user": []byte(userBytes), "pass": []byte(passBytes),
		}}, serviceuri)
	client := test.CreateFakeClient([]runtime.Object{instance, cm, statefulset, kafkaList}, nil, nil)

	// make sure that defaults were inserted
	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "KOGITO_PROTOBUF_FOLDER",
		Value: defaultEnvs["KOGITO_PROTOBUF_FOLDER"],
	})

	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  infinispanEnvKeyUseAuth,
		Value: "true",
	})

	valueFromUsername := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: t.Name()},
			Key:                  infinispan.SecretUsernameKey,
		},
	}
	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:      infinispanEnvKeyUsername,
		ValueFrom: valueFromUsername,
	})

	// change the spec, adding one more key
	instance.Spec.Env = map[string]string{
		"key2":                   "value2",
		"key1":                   "value1",
		"KOGITO_PROTOBUF_FOLDER": "/any/invalid/path",
	}

	// reconcile
	err := ManageResources(instance, &KogitoDataIndexResources{StatefulSet: statefulset, ProtoBufConfigMap: cm}, client)
	assert.NoError(t, err)

	// check the result
	_, err = kubernetes.ResourceC(client).Fetch(statefulset)
	assert.NoError(t, err)
	assert.Len(t, statefulset.Spec.Template.Spec.Containers[0].Env, 2+5+2+4) //default + infinispan + custom
	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "KOGITO_PROTOBUF_FOLDER",
		Value: defaultEnvs["KOGITO_PROTOBUF_FOLDER"],
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
	err = ManageResources(instance, &KogitoDataIndexResources{StatefulSet: statefulset, ProtoBufConfigMap: cm}, client)
	assert.NoError(t, err)

	assert.Len(t, statefulset.Spec.Template.Spec.Containers[0].Env, 2+5+1+4) //default + infinispan + custom
	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "KOGITO_PROTOBUF_FOLDER",
		Value: defaultEnvs["KOGITO_PROTOBUF_FOLDER"],
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

func Test_ensureProtoBufConfigMap(t *testing.T) {
	// setup
	dataIndex := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: t.Name(),
		},
	}
	cmWithFile := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cm-with-file",
			Namespace: t.Name(),
		},
		Data: map[string]string{"file.proto": "this is a protofile"},
	}
	cli := test.CreateFakeClient([]runtime.Object{dataIndex, cmWithFile}, nil, nil)

	// sanity check
	assert.Len(t, cmWithFile.Data, 1)

	// we don't have an image available, but the configMap has a file
	// we should receive true because the map has changed with no file
	ensureProtoBufConfigMap(dataIndex, cmWithFile, cli)
	exist, err := kubernetes.ResourceC(cli).Fetch(cmWithFile)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Len(t, cmWithFile.Data, 0)

	// now we have deployed a image stream tag with a docker image that has the protobuf label with a file attached
	// we should have this file in the given configMap. The file is encoded in base64 in the docker metadata
	// see the test data directory for  the label org.kie/persistence/protobuf/onboarding(...)
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kogitoapp",
			Namespace: t.Name(),
		},
	}
	isTag := &imgv1.ImageStreamTag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kogitoapp:latest",
			Namespace: t.Name(),
		},
		Image: imgv1.Image{
			DockerImageMetadata: runtime.RawExtension{
				Raw: test.HelperLoadBytes(t, "onboarding-dockerimage-json"),
			},
		},
	}

	cli = test.CreateFakeClient([]runtime.Object{dataIndex, cmWithFile, kogitoApp}, []runtime.Object{isTag}, nil)

	ensureProtoBufConfigMap(dataIndex, cmWithFile, cli)
	exist, err = kubernetes.ResourceC(cli).Fetch(cmWithFile)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Len(t, cmWithFile.Data, 1)
	assert.Contains(t, cmWithFile.Data, "kogitoapp-onboarding.onboarding.proto")
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
			KafkaClusterSpec: kafkabetav1.KafkaClusterSpec{
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
				assert.Equal(t, kafkaTopic.Spec.Replicas, kafka.Spec.Replicas)
			}
		})
	}
}
