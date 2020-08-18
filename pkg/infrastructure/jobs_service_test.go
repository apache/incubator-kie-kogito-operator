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

package infrastructure

import (
	appsv1 "k8s.io/api/apps/v1"
	"testing"

	"github.com/google/uuid"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func TestInjectJobsServicesURLIntoKogitoRuntime(t *testing.T) {
	URI := "http://localhost:8080"
	replicas := int32(1)
	app := &v1alpha1.KogitoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kogito-app",
			Namespace: t.Name(),
			UID:       types.UID(uuid.New().String()),
		},
	}
	jobs := &v1alpha1.KogitoJobsService{
		ObjectMeta: metav1.ObjectMeta{Name: DefaultJobsServiceName, Namespace: t.Name()},
		Spec: v1alpha1.KogitoJobsServiceSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{Replicas: &replicas},
		},
		Status: v1alpha1.KogitoJobsServiceStatus{KogitoServiceStatus: v1alpha1.KogitoServiceStatus{ExternalURI: URI}},
	}
	dc := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "dc", Namespace: t.Name(), OwnerReferences: []metav1.OwnerReference{{
			Name: app.Name,
			UID:  app.UID,
		}}},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "the-app"}}}},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{app, dc, jobs}, nil, nil)
	err := InjectJobsServicesURLIntoKogitoRuntimeServices(cli, t.Name())
	assert.NoError(t, err)
	assert.Len(t, dc.Spec.Template.Spec.Containers[0].Env, 0)

	exists, err := kubernetes.ResourceC(cli).Fetch(dc)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Len(t, dc.Spec.Template.Spec.Containers[0].Env, 1)
	assert.Contains(t, dc.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  jobsServicesHTTPURIEnv,
		Value: URI,
	})
}

func TestInjectJobsServicesURLIntoKogitoRuntimeCleanUp(t *testing.T) {
	URI := "http://localhost:8080"
	replicas := int32(1)
	app := &v1alpha1.KogitoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kogito-app",
			Namespace: t.Name(),
			UID:       types.UID(uuid.New().String()),
		},
	}
	jobs := &v1alpha1.KogitoJobsService{
		ObjectMeta: metav1.ObjectMeta{Name: DefaultJobsServiceName, Namespace: t.Name()},
		Spec: v1alpha1.KogitoJobsServiceSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{Replicas: &replicas},
		},
		Status: v1alpha1.KogitoJobsServiceStatus{KogitoServiceStatus: v1alpha1.KogitoServiceStatus{ExternalURI: URI}},
	}
	dc := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "dc", Namespace: t.Name(), OwnerReferences: []metav1.OwnerReference{{
			Name: app.Name,
			UID:  app.UID,
		}}},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "the-app"}}}},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{app, dc, jobs}, nil, nil)
	// first we inject
	err := InjectJobsServicesURLIntoKogitoRuntimeServices(cli, t.Name())
	assert.NoError(t, err)

	exists, err := kubernetes.ResourceC(cli).Fetch(dc)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Contains(t, dc.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  jobsServicesHTTPURIEnv,
		Value: URI,
	})

	err = kubernetes.ResourceC(cli).Delete(jobs)
	assert.NoError(t, err)

	// after deletion, we should have no env
	err = InjectJobsServicesURLIntoKogitoRuntimeServices(cli, t.Name())
	assert.NoError(t, err)

	dc = &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: dc.Name, Namespace: dc.Namespace}}
	exists, err = kubernetes.ResourceC(cli).Fetch(dc)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Len(t, dc.Spec.Template.Spec.Containers[0].Env, 1)
	assert.Contains(t, dc.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  jobsServicesHTTPURIEnv,
		Value: "",
	})
}
