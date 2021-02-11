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

package test

import (
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/client"
	"github.com/kiegroup/kogito-cloud-operator/core/client/kubernetes"
	testapi "github.com/kiegroup/kogito-cloud-operator/core/test/api"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type fakeKogitoInfraHandler struct {
	client *client.Client
}

// CreateFakeKogitoInfraHandler ...
func CreateFakeKogitoInfraHandler(client *client.Client) api.KogitoInfraHandler {
	return &fakeKogitoInfraHandler{
		client: client,
	}
}

func (k *fakeKogitoInfraHandler) FetchKogitoInfraInstance(key types.NamespacedName) (api.KogitoInfraInterface, error) {
	instance := &testapi.KogitoInfraTest{}
	if exists, resultErr := kubernetes.ResourceC(k.client).FetchWithKey(key, instance); resultErr != nil {
		return nil, resultErr
	} else if !exists {
		return nil, nil
	} else {
		return instance, nil
	}
}

// CreateFakeKogitoKafka create fake kogito infra instance for kafka
func CreateFakeKogitoKafka(namespace string) api.KogitoInfraInterface {
	return &testapi.KogitoInfraTest{
		ObjectMeta: v1.ObjectMeta{
			Name:      "kogito-kafka",
			Namespace: namespace,
		},
		Spec: api.KogitoInfraSpec{
			Resource: api.Resource{
				Kind:       "Kafka",
				APIVersion: "kafka.strimzi.io/v1beta1",
			},
		},
		Status: api.KogitoInfraStatus{
			RuntimeProperties: map[api.RuntimeType]api.RuntimeProperties{
				api.QuarkusRuntimeType: {
					AppProps: map[string]string{
						"kafka.bootstrap.servers": "kogito-kafka-kafka-bootstrap.test.svc:9092",
					},
					Env: []corev1.EnvVar{
						{
							Name:  "ENABLE_EVENTS",
							Value: "true",
						},
					},
				},
			},
			Condition: api.KogitoInfraCondition{
				Type:   api.SuccessInfraConditionType,
				Status: v1.StatusSuccess,
				Reason: "",
			},
		},
	}
}

// CreateFakeKogitoInfinispan create fake kogito infra instance for Infinispan
func CreateFakeKogitoInfinispan(namespace string) api.KogitoInfraInterface {
	return &testapi.KogitoInfraTest{
		ObjectMeta: v1.ObjectMeta{
			Name:      "kogito-Infinispan",
			Namespace: namespace,
		},
		Spec: api.KogitoInfraSpec{
			Resource: api.Resource{
				Kind:       "Infinispan",
				APIVersion: "infinispan.org/v1",
			},
		},
		Status: api.KogitoInfraStatus{
			RuntimeProperties: map[api.RuntimeType]api.RuntimeProperties{
				api.QuarkusRuntimeType: {
					AppProps: map[string]string{
						"quarkus.infinispan-client.server-list": "infinispanInstance:11222",
					},
					Env: []corev1.EnvVar{
						{
							Name:  "ENABLE_PERSISTENCE",
							Value: "true",
						},
					},
				},
			},
			Volumes: []api.KogitoInfraVolume{
				{
					Mount: corev1.VolumeMount{
						Name:      "tls-configuration",
						ReadOnly:  true,
						MountPath: "/home/kogito/certs",
						SubPath:   "truststore.p12",
					},
					NamedVolume: api.ConfigVolume{
						Name: "tls-configuration",
						ConfigVolumeSource: api.ConfigVolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "infinispan-secret",
								Items: []corev1.KeyToPath{
									{
										Key:  "tls.crt",
										Path: "tls.crt",
									},
								},
							},
						},
					},
				},
			},
			Condition: api.KogitoInfraCondition{
				Type:   api.SuccessInfraConditionType,
				Status: v1.StatusSuccess,
				Reason: "",
			},
		},
	}
}
