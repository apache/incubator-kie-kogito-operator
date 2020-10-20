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

package explainability

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
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
)

func TestReconcileKogitoSupportingServiceExplainability_Reconcile(t *testing.T) {
	ns := t.Name()
	kogitoKafka := test.CreateFakeKogitoKafka(t.Name())
	instance := &v1alpha1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "expl",
			Namespace: ns,
		},
		Spec: v1alpha1.KogitoSupportingServiceSpec{
			ServiceType: v1alpha1.Explainablity,
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{
				Infra: []string{
					kogitoKafka.Name,
				},
			},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{instance, kogitoKafka}, nil, nil)
	r := &SupportingServiceResource{}

	// basic checks
	requeue, err := r.Reconcile(cli, instance, meta.GetRegisteredSchema())
	assert.NoError(t, err)
	assert.False(t, requeue)
}

func TestReconcileKogitoSupportingServiceExplainability_UpdateHTTPPort(t *testing.T) {
	ns := t.Name()
	instance := &v1alpha1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "expl",
			Namespace: ns,
		},
		Spec: v1alpha1.KogitoSupportingServiceSpec{
			ServiceType: v1alpha1.Explainablity,
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{
				HTTPPort: 9090,
			},
		},
	}
	is, tag := test.GetImageStreams(infrastructure.DefaultExplainabilityImageName, instance.Namespace, instance.Name, infrastructure.GetKogitoImageVersion())
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{instance, is}, []runtime.Object{tag}, nil)
	r := &SupportingServiceResource{}

	requeue, err := r.Reconcile(cli, instance, meta.GetRegisteredSchema())
	assert.NoError(t, err)
	assert.False(t, requeue)

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
	err = kubernetes.ResourceC(cli).Update(routeFromResource)
	assert.NoError(t, err)

	// reconcile
	_, err = r.Reconcile(cli, instance, meta.GetRegisteredSchema())
	assert.NoError(t, err)

	// get the route after reconcile
	routeAfterReconcile := &routev1.Route{}
	test.AssertFetchWithKeyMustExist(t, cli, routeAfterReconcile, instance)
	assert.Equal(t, "http", routeAfterReconcile.Spec.Port.TargetPort.StrVal)

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
	_, err = r.Reconcile(cli, instance, meta.GetRegisteredSchema())
	assert.NoError(t, err)

	// get the service after reconcile
	serviceAfterReconcile := &corev1.Service{}
	test.AssertFetchWithKeyMustExist(t, cli, serviceAfterReconcile, instance)

	// compare again if the port was updated after reconcile
	assert.Equal(t, int32(80), serviceAfterReconcile.Spec.Ports[0].Port)
	assert.Equal(t, intstr.FromInt(9090), serviceAfterReconcile.Spec.Ports[0].TargetPort)
}
