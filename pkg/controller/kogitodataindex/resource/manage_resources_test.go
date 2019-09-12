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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"

	"github.com/stretchr/testify/assert"
)

func Test_ManageResources_WhenKafkaURIIsChanged(t *testing.T) {
	instance := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-data-index",
			Namespace: "test",
		},
		Spec: v1alpha1.KogitoDataIndexSpec{
			Replicas: 1,
		},
	}
	serviceuri := "myserviceuri:9092"
	cm := newProtobufConfigMap(instance)
	secret := &corev1.Secret{}
	statefulset := newStatefulset(instance, cm, *secret)
	client, _ := test.CreateFakeClient([]runtime.Object{instance, cm, statefulset, secret}, nil, nil)

	err := ManageResources(instance, &KogitoDataIndexResources{StatefulSet: statefulset, ProtoBufConfigMap: cm}, client)
	assert.NoError(t, err)
	assert.NotContains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: kafkaEnvKeyServiceURI, Value: serviceuri})

	instance.Spec.Kafka = v1alpha1.KafkaConnectionProperties{ServiceURI: serviceuri}
	err = ManageResources(instance, &KogitoDataIndexResources{StatefulSet: statefulset, ProtoBufConfigMap: cm}, client)
	assert.NoError(t, err)
	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: kafkaEnvKeyServiceURI, Value: serviceuri})
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
	userBytes := []byte("developer")
	passBytes := []byte("developer")
	cm := newProtobufConfigMap(instance)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "infinispan-secret", Namespace: "test"},
		Data: map[string][]byte{
			"user": []byte(userBytes), "pass": []byte(passBytes),
		},
	}
	statefulset := newStatefulset(instance, cm, *secret)
	client, _ := test.CreateFakeClient([]runtime.Object{instance, cm, statefulset, secret}, nil, nil)

	// reconcile
	err := ManageResources(instance, &KogitoDataIndexResources{StatefulSet: statefulset, ProtoBufConfigMap: cm}, client)
	assert.NoError(t, err)
	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: string(infinispanEnvKeyUsername), Value: "developer"})
	assert.NotContains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: string(infinispanEnvKeyAuthRealm), Value: "default"})

	// let's change
	instance.Spec.Infinispan.AuthRealm = "default"
	instance.Spec.Infinispan.ServiceURI = "myservice:11222"
	err = ManageResources(instance, &KogitoDataIndexResources{StatefulSet: statefulset, ProtoBufConfigMap: cm}, client)
	assert.NoError(t, err)
	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: string(infinispanEnvKeyAuthRealm), Value: "default"})
	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: string(interfaceEnvKeyServiceURI), Value: "myservice:11222"})
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
	userBytes := []byte("developer")
	passBytes := []byte("developer")
	cm := newProtobufConfigMap(instance)
	statefulset := newStatefulset(instance, cm, corev1.Secret{
		Data: map[string][]byte{
			"user": []byte(userBytes), "pass": []byte(passBytes),
		}})
	client, _ := test.CreateFakeClient([]runtime.Object{instance, cm, statefulset}, nil, nil)

	// make sure that defaults were inserted
	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "KOGITO_PROTOBUF_FOLDER",
		Value: defaultEnvs["KOGITO_PROTOBUF_FOLDER"],
	})

	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  infinispanEnvKeyUseAuth,
		Value: "true",
	})

	assert.Contains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  infinispanEnvKeyUsername,
		Value: "developer",
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
	assert.Len(t, statefulset.Spec.Template.Spec.Containers[0].Env, 2+4+2) //default + infinispan + custom
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

	assert.Len(t, statefulset.Spec.Template.Spec.Containers[0].Env, 2+4+1) //default + infinispan + custom
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
