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

package controllers

import (
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/core/framework"
	"github.com/kiegroup/kogito-cloud-operator/core/test"
	"github.com/kiegroup/kogito-cloud-operator/internal"
	imagev1 "github.com/openshift/api/image/v1"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestReconcileKogitoRuntime_Reconcile(t *testing.T) {
	replicas := int32(1)

	kogitoKafka := newSuccessfulKafkaInfra(t.Name())
	kogitoInfinispan := newSuccessfulInfinispanInfra(t.Name())

	instance := &v1beta1.KogitoRuntime{
		ObjectMeta: v1.ObjectMeta{Name: "example-quarkus", Namespace: t.Name()},
		Spec: v1beta1.KogitoRuntimeSpec{
			KogitoServiceSpec: api.KogitoServiceSpec{
				Replicas:      &replicas,
				ServiceLabels: map[string]string{"process": "example-quarkus"},
				Infra: []string{
					kogitoKafka.GetName(),
					kogitoInfinispan.GetName(),
				},
			},
		},
	}

	cli := test.NewFakeClientBuilder().UseScheme(internal.GetRegisteredSchema()).AddK8sObjects(instance, kogitoKafka, kogitoInfinispan).Build()
	r := KogitoRuntimeReconciler{Client: cli, Scheme: internal.GetRegisteredSchema(), Log: test.TestLogger}

	// first reconciliation
	test.AssertReconcileMustNotRequeue(t, &r, instance)
	// second time
	test.AssertReconcileMustNotRequeue(t, &r, instance)

	_, err := kubernetes.ResourceC(cli).Fetch(instance)
	assert.NoError(t, err)
	assert.NotNil(t, instance.Status)
	assert.Len(t, instance.Status.Conditions, 1)

	// svc discovery
	svc := &corev1.Service{ObjectMeta: v1.ObjectMeta{Name: instance.Name, Namespace: instance.Namespace}}
	exists, err := kubernetes.ResourceC(cli).Fetch(svc)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.True(t, svc.Labels["process"] == instance.Name)

	// sa, namespace env var, volume count and protobuf
	deployment := &appsv1.Deployment{ObjectMeta: v1.ObjectMeta{Name: instance.Name, Namespace: instance.Namespace}}
	exists, err = kubernetes.ResourceC(cli).Fetch(deployment)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.True(t, framework.GetEnvVarFromContainer("NAMESPACE", &deployment.Spec.Template.Spec.Containers[0]) == instance.Namespace)
	assert.Equal(t, "kogito-service-viewer", deployment.Spec.Template.Spec.ServiceAccountName)
	assert.Len(t, deployment.Spec.Template.Spec.Volumes, 2) // #1 for property, #2 for tls
	// command to register protobuf does not exist anymore
	assert.Nil(t, deployment.Spec.Template.Spec.Containers[0].Lifecycle)

	configMap := &corev1.ConfigMap{ObjectMeta: v1.ObjectMeta{Name: getProtoBufConfigMapName(instance.Name), Namespace: instance.Namespace}}
	exists, err = kubernetes.ResourceC(cli).Fetch(configMap)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, getProtoBufConfigMapName(instance.Name), configMap.Name)
}

// see https://issues.redhat.com/browse/KOGITO-2535
func TestReconcileKogitoRuntime_CustomImage(t *testing.T) {
	replicas := int32(1)
	instance := &v1beta1.KogitoRuntime{
		ObjectMeta: v1.ObjectMeta{Name: "process-springboot-example", Namespace: t.Name()},
		Spec: v1beta1.KogitoRuntimeSpec{
			Runtime: api.SpringBootRuntimeType,
			KogitoServiceSpec: api.KogitoServiceSpec{
				Replicas: &replicas,
				Image:    "quay.io/custom/process-springboot-example-default:latest",
			},
		},
	}
	cli := test.NewFakeClientBuilder().UseScheme(internal.GetRegisteredSchema()).AddK8sObjects(instance).OnOpenShift().Build()

	test.AssertReconcileMustNotRequeue(t, &KogitoRuntimeReconciler{Client: cli, Scheme: internal.GetRegisteredSchema(), Log: test.TestLogger}, instance)

	_, err := kubernetes.ResourceC(cli).Fetch(instance)
	assert.NoError(t, err)
	assert.NotNil(t, instance.Status)
	assert.Len(t, instance.Status.Conditions, 1)

	// image stream
	is := imagev1.ImageStream{
		ObjectMeta: v1.ObjectMeta{Name: instance.Name, Namespace: instance.Namespace},
	}
	exists, err := kubernetes.ResourceC(cli).Fetch(&is)
	assert.True(t, exists)
	assert.NoError(t, err)
	assert.Len(t, is.Spec.Tags, 1)
	assert.Equal(t, "latest", is.Spec.Tags[0].Name)
	assert.Equal(t, "quay.io/custom/process-springboot-example-default:latest", is.Spec.Tags[0].From.Name)
}

func TestReconcileKogitoRuntime_CustomConfigMap(t *testing.T) {
	replicas := int32(1)
	cm := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{Namespace: t.Name(), Name: "mysuper-cm"},
	}
	instance := &v1beta1.KogitoRuntime{
		ObjectMeta: v1.ObjectMeta{Name: "process-springboot-example", Namespace: t.Name()},
		Spec: v1beta1.KogitoRuntimeSpec{
			Runtime: api.SpringBootRuntimeType,
			KogitoServiceSpec: api.KogitoServiceSpec{
				Replicas:            &replicas,
				PropertiesConfigMap: "mysuper-cm",
			},
		},
	}
	cli := test.NewFakeClientBuilder().UseScheme(internal.GetRegisteredSchema()).AddK8sObjects(instance, cm).Build()
	// we take the ownership of the custom cm
	test.AssertReconcileMustRequeue(t, &KogitoRuntimeReconciler{Client: cli, Scheme: internal.GetRegisteredSchema(), Log: test.TestLogger}, instance)
	// we requeue..
	test.AssertReconcileMustNotRequeue(t, &KogitoRuntimeReconciler{Client: cli, Scheme: internal.GetRegisteredSchema(), Log: test.TestLogger}, instance)
	_, err := kubernetes.ResourceC(cli).Fetch(cm)
	assert.NoError(t, err)
	assert.NotEmpty(t, cm.OwnerReferences)
}

// newSuccessfulKafkaInfra create kogito infra instance for kafka
func newSuccessfulKafkaInfra(namespace string) api.KogitoInfraInterface {
	return &v1beta1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{
			Name:      "kogito-kafka",
			Namespace: namespace,
		},
		Spec: api.KogitoInfraSpec{
			Resource: api.Resource{
				Kind:       "kafka.strimzi.io/v1beta1",
				APIVersion: "Kafka",
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

// newSuccessfulInfinispanInfra create kogito infra instance for Infinispan
func newSuccessfulInfinispanInfra(namespace string) api.KogitoInfraInterface {
	return &v1beta1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{
			Name:      "kogito-Infinispan",
			Namespace: namespace,
		},
		Spec: api.KogitoInfraSpec{
			Resource: api.Resource{
				Kind:       "infinispan.org/v1",
				APIVersion: "Infinispan",
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
