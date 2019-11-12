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
	v1 "github.com/openshift/api/image/v1"
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
	client := test.CreateFakeClient([]runtime.Object{instance, cm, statefulset, secret}, nil, nil)

	err := ManageResources(instance, &KogitoDataIndexResources{StatefulSet: statefulset, ProtoBufConfigMap: cm}, client)
	assert.NoError(t, err)
	for _, kafkaKey := range managedKafkaKeys {
		assert.NotContains(t, statefulset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: kafkaKey, Value: serviceuri})
	}

	instance.Spec.Kafka = v1alpha1.KafkaConnectionProperties{ServiceURI: serviceuri}
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
	client := test.CreateFakeClient([]runtime.Object{instance, cm, statefulset, secret}, nil, nil)

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
	client := test.CreateFakeClient([]runtime.Object{instance, cm, statefulset}, nil, nil)

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
	isTag := &v1.ImageStreamTag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kogitoapp:latest",
			Namespace: t.Name(),
		},
		Image: v1.Image{
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
