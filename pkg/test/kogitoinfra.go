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
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateFakeKogitoKafka create fake kogito infra instance for kafka
func CreateFakeKogitoKafka(namespace string) *v1beta1.KogitoInfra {
	return &v1beta1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{
			Name:      "kogito-kafka",
			Namespace: namespace,
		},
		Spec: v1beta1.KogitoInfraSpec{
			Resource: v1beta1.Resource{
				Kind:       "kafka.strimzi.io/v1beta1",
				APIVersion: "Kafka",
			},
		},
		Status: v1beta1.KogitoInfraStatus{
			RuntimeProperties: map[v1beta1.RuntimeType]v1beta1.RuntimeProperties{
				v1beta1.QuarkusRuntimeType: {
					AppProps: map[string]string{
						"kafka.bootstrap.servers": "kogito-kafka-kafka-bootstrap.test.svc:9092",
					},
					Env: []corev1.EnvVar{
						{
							Name:  "ENABLE_PERSISTENCE",
							Value: "true",
						},
					},
				},
			},
		},
	}
}

// CreateFakeKogitoInfinispan create fake kogito infra instance for Infinispan
func CreateFakeKogitoInfinispan(namespace string) *v1beta1.KogitoInfra {
	return &v1beta1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{
			Name:      "kogito-Infinispan",
			Namespace: namespace,
		},
		Spec: v1beta1.KogitoInfraSpec{
			Resource: v1beta1.Resource{
				Kind:       "infinispan.org/v1",
				APIVersion: "Infinispan",
			},
		},
		Status: v1beta1.KogitoInfraStatus{
			RuntimeProperties: map[v1beta1.RuntimeType]v1beta1.RuntimeProperties{
				v1beta1.QuarkusRuntimeType: {
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
			Volume: []v1beta1.KogitoInfraVolume{
				{
					Mount: corev1.VolumeMount{
						Name:      "tls-configuration",
						ReadOnly:  true,
						MountPath: "/home/kogito/certs",
						SubPath:   "truststore.p12",
					},
					NamedVolume: v1beta1.ConfigVolume{
						Name: "tls-configuration",
						ConfigVolumeSource: v1beta1.ConfigVolumeSource{
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
		},
	}
}
