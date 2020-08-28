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

package controllers

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sort"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/api/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/external/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	routev1 "github.com/openshift/api/route/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/stretchr/testify/assert"
)

func TestReconcileKogitoDataIndex_Reconcile(t *testing.T) {
	ns := t.Name()
	instance := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-data-index",
			Namespace: ns,
		},
		// We don't need to specify that we need Infinispan, it will figure out that alone :)
		Spec: v1alpha1.KogitoDataIndexSpec{},
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
	r := &KogitoDataIndexReconciler{
		Client: cli,
		Scheme: meta.GetRegisteredSchema(),
	}
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
	}

	// basic checks
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}

	// check infra
	infra, ready, err := infrastructure.EnsureKogitoInfra(ns, cli).WithInfinispan().Apply()
	assert.NoError(t, err)
	assert.False(t, ready)  // we don't have status defined since the KogitoInfra controller is not running
	assert.NotNil(t, infra) // we have a infra instance created during reconciliation phase
	assert.Equal(t, infrastructure.DefaultKogitoInfraName, infra.GetName())
	assert.True(t, infra.Spec.InstallInfinispan)
}

func TestReconcileKogitoDataIndex_UpdateHTTPPort(t *testing.T) {
	ns := t.Name()
	instance := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-data-index",
			Namespace: ns,
		},
		Spec: v1alpha1.KogitoDataIndexSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{
				HTTPPort: 9090,
			},
			KafkaMeta: v1alpha1.KafkaMeta{
				KafkaProperties: v1alpha1.KafkaConnectionProperties{
					UseKogitoInfra: false,
					ExternalURI:    "my-uri:9022",
				},
			},
			InfinispanMeta: v1alpha1.InfinispanMeta{
				InfinispanProperties: v1alpha1.InfinispanConnectionProperties{
					UseKogitoInfra: false,
					URI:            "another-uri:11222",
				},
			},
		},
	}
	is, tag := test.GetImageStreams(infrastructure.DefaultDataIndexImageName, instance.Namespace, instance.Name, infrastructure.GetKogitoImageVersion())
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{instance, is}, []runtime.Object{tag}, nil)
	r := &KogitoDataIndexReconciler{
		Client: cli,
		Scheme: meta.GetRegisteredSchema(),
	}
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
	}

	// first reconcile
	_, err := r.Reconcile(req)
	assert.NoError(t, err)

	// make sure HTTPPort env was added on the deployment
	deployment := &appsv1.Deployment{}
	exists, err := kubernetes.ResourceC(cli).FetchWithKey(types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, deployment)
	assert.True(t, exists)
	assert.NoError(t, err)

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
	exists, err = kubernetes.ResourceC(cli).Fetch(service)
	assert.True(t, exists)
	assert.NoError(t, err)
	assert.Equal(t, int32(9090), service.Spec.Ports[0].TargetPort.IntVal)

	// update the route
	// reconcile and test
	// compare the route http port
	routeFromResource := &routev1.Route{}
	routeFound, err := kubernetes.ResourceC(cli).FetchWithKey(types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, routeFromResource)
	assert.NoError(t, err)
	assert.True(t, routeFound)
	// update http port on the given route
	routeFromResource.Spec.Port.TargetPort.IntVal = 4000
	err = kubernetes.ResourceC(cli).Update(routeFromResource)
	assert.NoError(t, err)
	// reconcile
	_, err = r.Reconcile(req)
	assert.NoError(t, err)
	// get the route after reconcile
	routeAfterReconcile := &routev1.Route{}
	routeAfterReconcileFound, err := kubernetes.ResourceC(cli).FetchWithKey(types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, routeAfterReconcile)
	assert.True(t, routeAfterReconcileFound)
	assert.NoError(t, err)
	assert.True(t, routeAfterReconcileFound)
	assert.Equal(t, intstr.IntOrString{Type: 0, IntVal: 9090, StrVal: ""}, routeAfterReconcile.Spec.Port.TargetPort)

	// update the service
	// reconcile and test
	// compare the service http and target port
	serviceFromResource := &corev1.Service{}
	serviceFound, err := kubernetes.ResourceC(cli).FetchWithKey(types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, serviceFromResource)
	assert.True(t, serviceFound)
	assert.NoError(t, err)
	// update ports
	serviceFromResource.Spec.Ports[0].Port = 4000
	serviceFromResource.Spec.Ports[0].TargetPort = intstr.FromString("4000")
	err = kubernetes.ResourceC(cli).Update(serviceFromResource)
	assert.NoError(t, err)
	// reconcile
	_, err = r.Reconcile(req)
	assert.NoError(t, err)
	// get the service after reconcile
	serviceAfterReconcile := &corev1.Service{}
	serviceAfterReconcileFound, err := kubernetes.ResourceC(cli).FetchWithKey(types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, serviceAfterReconcile)
	assert.True(t, serviceAfterReconcileFound)
	assert.NoError(t, err)
	// compare again if the port was updated after reconcile
	assert.Equal(t, int32(9090), serviceAfterReconcile.Spec.Ports[0].Port)
	assert.Equal(t, intstr.FromInt(9090), serviceAfterReconcile.Spec.Ports[0].TargetPort)
}

func TestReconcileKogitoDataIndex_mountProtoBufConfigMaps(t *testing.T) {
	fileName1 := "mydomain.proto"
	fileName2 := "mydomain2.proto"
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: t.Name(),
			Name:      "my-domain-protobufs",
			Labels:    map[string]string{infrastructure.ConfigMapProtoBufEnabledLabelKey: "true"},
		},
		Data: map[string]string{
			fileName1: "This is a protobuf file",
			fileName2: "This is another file",
		},
	}
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{cm}, nil, nil)
	reconcileDataIndex := KogitoDataIndexReconciler{
		Client: cli,
		Scheme: meta.GetRegisteredSchema(),
	}
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Namespace: t.Name(), Name: infrastructure.DefaultDataIndexName},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "my-container",
						},
					},
				},
			},
		},
	}
	err := reconcileDataIndex.mountProtoBufConfigMaps(deployment)
	assert.NoError(t, err)
	assert.Len(t, deployment.Spec.Template.Spec.Volumes, 1)
	assert.Contains(t, deployment.Spec.Template.Spec.Volumes[0].Name, cm.Name)
	assert.Len(t, deployment.Spec.Template.Spec.Containers[0].VolumeMounts, 2)

	// we need to have them ordered to be able to do the appropriate comparision.
	sort.Slice(deployment.Spec.Template.Spec.Containers[0].VolumeMounts, func(i, j int) bool {
		return deployment.Spec.Template.Spec.Containers[0].VolumeMounts[i].SubPath < deployment.Spec.Template.Spec.Containers[0].VolumeMounts[j].SubPath
	})
	assert.Equal(t, deployment.Spec.Template.Spec.Containers[0].VolumeMounts[0].Name, cm.Name)
	assert.Equal(t, deployment.Spec.Template.Spec.Containers[0].VolumeMounts[1].Name, cm.Name)
	assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].VolumeMounts[0].SubPath, fileName1)
	assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath, fileName1)
}

func TestReconcileKogitoDataIndex_MultipleProtoBufCMs(t *testing.T) {
	fileName1 := "mydomain.proto"
	fileName2 := "mydomain2.proto"
	instance := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{Namespace: t.Name(), Name: infrastructure.DefaultDataIndexName},
		Spec: v1alpha1.KogitoDataIndexSpec{
			InfinispanMeta: v1alpha1.InfinispanMeta{InfinispanProperties: v1alpha1.InfinispanConnectionProperties{UseKogitoInfra: false, URI: "infinispan:20220"}},
			KafkaMeta:      v1alpha1.KafkaMeta{KafkaProperties: v1alpha1.KafkaConnectionProperties{UseKogitoInfra: false, ExternalURI: "kafka:9900"}},
		},
	}
	cm1 := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: t.Name(),
			Name:      "my-domain-protobufs1",
			Labels:    map[string]string{infrastructure.ConfigMapProtoBufEnabledLabelKey: "true"},
		},
		Data: map[string]string{fileName1: "This is a protobuf file"},
	}
	cm2 := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: t.Name(),
			Name:      "my-domain-protobufs2",
			Labels:    map[string]string{infrastructure.ConfigMapProtoBufEnabledLabelKey: "true"},
		},
		Data: map[string]string{fileName2: "This is a protobuf file"},
	}
	cli := test.CreateFakeClient([]runtime.Object{instance, cm1, cm2}, nil, nil)
	r := &KogitoDataIndexReconciler{
		Client: cli,
		Scheme: meta.GetRegisteredSchema(),
	}
	test.AssertReconcileMustNotRequeue(t, r, instance)
	test.AssertReconcileMustNotRequeue(t, r, instance)
}
