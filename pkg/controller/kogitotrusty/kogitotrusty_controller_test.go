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

package kogitotrusty

import (
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"

	routev1 "github.com/openshift/api/route/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/stretchr/testify/assert"
)

func TestReconcileKogitoTrusty_Reconcile(t *testing.T) {
	ns := t.Name()
	instance := &v1alpha1.KogitoTrusty{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-trusty",
			Namespace: ns,
		},
		// We don't need to specify that we need Infinispan, it will figure out that alone :)
		Spec: v1alpha1.KogitoTrustySpec{},
	}
	kafkaList := &kafkabetav1.KafkaList{
		Items: []kafkabetav1.Kafka{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "kafka", Namespace: ns},
				Spec:       kafkabetav1.KafkaSpec{Kafka: kafkabetav1.KafkaClusterSpec{Replicas: 1}},
				Status: kafkabetav1.KafkaStatus{
					Listeners: []kafkabetav1.ListenerStatus{
						{
							Type:      "plain",
							Addresses: []kafkabetav1.ListenerAddress{{Host: "kafka", Port: 9092}},
						},
					},
				},
			},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{instance, kafkaList}, nil, nil)
	r := &ReconcileKogitoTrusty{
		client: cli,
		scheme: meta.GetRegisteredSchema(),
	}

	// basic checks
	test.AssertReconcile(t, r, instance)
}

func TestReconcileKogitoTrusty_UpdateHTTPPort(t *testing.T) {
	ns := t.Name()
	instance := &v1alpha1.KogitoTrusty{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-trusty",
			Namespace: ns,
		},
		Spec: v1alpha1.KogitoTrustySpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{
				HTTPPort: 9090,
			},
		},
	}
	is, tag := test.GetImageStreams(infrastructure.DefaultTrustyImageName, instance.Namespace, instance.Name, infrastructure.GetKogitoImageVersion())
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{instance, is}, []runtime.Object{tag}, nil)
	r := &ReconcileKogitoTrusty{
		client: cli,
		scheme: meta.GetRegisteredSchema(),
	}

	test.AssertReconcile(t, r, instance)

	// make sure HTTPPort env was added on the deployment
	deployment := &appsv1.Deployment{}
	test.AssertFetchWithKeyMustExist(t, cli, deployment, instance)

	// make sure that the http port was correctly added.
	assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  services.HTTPPortEnvKey,
		Value: "9090",
	})

	assert.Equal(t, int32(9090), deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort)
	assert.Equal(t, int32(9090), deployment.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet.Port.IntVal)
	assert.Equal(t, int32(9090), deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Port.IntVal)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: instance.Name, Namespace: instance.Namespace},
	}
	test.AssertFetchMustExist(t, cli, service)
	assert.Equal(t, int32(9090), service.Spec.Ports[0].TargetPort.IntVal)

	// update the route
	// reconcile and test
	// compare the route http port
	routeFromResource := &routev1.Route{}
	test.AssertFetchWithKeyMustExist(t, cli, routeFromResource, instance)

	// update http port on the given route
	routeFromResource.Spec.Port.TargetPort.IntVal = 4000
	err := kubernetes.ResourceC(cli).Update(routeFromResource)
	assert.NoError(t, err)

	// reconcile
	test.AssertReconcile(t, r, instance)

	// get the route after reconcile
	routeAfterReconcile := &routev1.Route{}
	test.AssertFetchWithKeyMustExist(t, cli, routeAfterReconcile, instance)
	assert.Equal(t, intstr.IntOrString{Type: 0, IntVal: 9090, StrVal: ""}, routeAfterReconcile.Spec.Port.TargetPort)

	// update the service
	// reconcile and test
	// compare the service http and target port
	serviceFromResource := &corev1.Service{}
	test.AssertFetchWithKeyMustExist(t, cli, serviceFromResource, instance)

	// update ports
	serviceFromResource.Spec.Ports[0].Port = 4000
	serviceFromResource.Spec.Ports[0].TargetPort = intstr.FromString("4000")
	err = kubernetes.ResourceC(cli).Update(serviceFromResource)
	assert.NoError(t, err)

	// reconcile
	test.AssertReconcile(t, r, instance)

	// get the service after reconcile
	serviceAfterReconcile := &corev1.Service{}
	test.AssertFetchWithKeyMustExist(t, cli, serviceAfterReconcile, instance)

	// compare again if the port was updated after reconcile
	assert.Equal(t, int32(9090), serviceAfterReconcile.Spec.Ports[0].Port)
	assert.Equal(t, intstr.FromInt(9090), serviceAfterReconcile.Spec.Ports[0].TargetPort)
}
