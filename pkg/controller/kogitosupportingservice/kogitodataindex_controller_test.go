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

package kogitosupportingservice

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sort"
	"testing"
)

func TestKogitoSupportingServiceDataIndex_Reconcile(t *testing.T) {
	ns := t.Name()
	kogitoKafka := test.CreateFakeKogitoKafka(t.Name())
	kogitoInfinispan := test.CreateFakeKogitoInfinispan(t.Name())

	instance := &v1alpha1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-data-index",
			Namespace: ns,
		},
		// We don't need to specify that we need Infinispan, it will figure out that alone :)
		Spec: v1alpha1.KogitoSupportingServiceSpec{
			ServiceType: v1alpha1.DataIndex,
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{
				Infra: []string{
					kogitoKafka.Name,
					kogitoInfinispan.Name,
				},
			},
		},
	}

	cli := test.CreateFakeClient([]runtime.Object{instance, kogitoKafka, kogitoInfinispan}, nil, nil)
	r := &DataIndexSupportingServiceResource{}
	// basic checks
	_, err := r.Reconcile(cli, instance, meta.GetRegisteredSchema())
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
}

func TestReconcileKogitoSupportingServiceDataIndex_UpdateHTTPPort(t *testing.T) {
	ns := t.Name()
	instance := &v1alpha1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-data-index",
			Namespace: ns,
		},
		Spec: v1alpha1.KogitoSupportingServiceSpec{
			ServiceType: v1alpha1.DataIndex,
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{
				HTTPPort: 9090,
			},
		},
	}
	is, tag := test.GetImageStreams(infrastructure.DefaultDataIndexImageName, instance.Namespace, instance.Name, infrastructure.GetKogitoImageVersion())
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{instance, is}, []runtime.Object{tag}, nil)
	r := &DataIndexSupportingServiceResource{}

	// first reconcile
	_, err := r.Reconcile(cli, instance, meta.GetRegisteredSchema())
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
	_, err = r.Reconcile(cli, instance, meta.GetRegisteredSchema())
	assert.NoError(t, err)
	// get the route after reconcile
	routeAfterReconcile := &routev1.Route{}
	routeAfterReconcileFound, err := kubernetes.ResourceC(cli).FetchWithKey(types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, routeAfterReconcile)
	assert.True(t, routeAfterReconcileFound)
	assert.NoError(t, err)
	assert.True(t, routeAfterReconcileFound)
	assert.Equal(t, "http", routeAfterReconcile.Spec.Port.TargetPort.StrVal)

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
	_, err = r.Reconcile(cli, instance, meta.GetRegisteredSchema())
	assert.NoError(t, err)
	// get the service after reconcile
	serviceAfterReconcile := &corev1.Service{}
	serviceAfterReconcileFound, err := kubernetes.ResourceC(cli).FetchWithKey(types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, serviceAfterReconcile)
	assert.True(t, serviceAfterReconcileFound)
	assert.NoError(t, err)
	// compare again if the port was updated after reconcile
	assert.Equal(t, int32(80), serviceAfterReconcile.Spec.Ports[0].Port)
	assert.Equal(t, intstr.FromInt(9090), serviceAfterReconcile.Spec.Ports[0].TargetPort)
}

func TestReconcileKogitoSupportingServiceDataIndex_mountProtoBufConfigMaps(t *testing.T) {
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
	err := mountProtoBufConfigMaps(deployment, cli)
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

func TestReconcileKogitoSupportingServiceDataIndex_MultipleProtoBufCMs(t *testing.T) {
	fileName1 := "mydomain.proto"
	fileName2 := "mydomain2.proto"
	instance := &v1alpha1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{Namespace: t.Name(), Name: infrastructure.DefaultDataIndexName},
		Spec: v1alpha1.KogitoSupportingServiceSpec{
			ServiceType: v1alpha1.DataIndex,
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
	r := &DataIndexSupportingServiceResource{}
	requeue, err := r.Reconcile(cli, instance, meta.GetRegisteredSchema())
	assert.False(t, requeue)
	assert.NoError(t, err)
}
