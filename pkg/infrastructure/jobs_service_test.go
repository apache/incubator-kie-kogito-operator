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
	"github.com/google/uuid"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	oappsv1 "github.com/openshift/api/apps/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

func TestInjectJobsServicesURLIntoKogitoApps(t *testing.T) {
	URI := "http://localhost:8080"
	app := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kogito-app",
			Namespace: t.Name(),
			UID:       types.UID(uuid.New().String()),
		},
	}
	jobs := &v1alpha1.KogitoJobsService{
		ObjectMeta: metav1.ObjectMeta{Name: "jobs-service", Namespace: t.Name()},
		Spec:       v1alpha1.KogitoJobsServiceSpec{Replicas: 1},
		Status:     v1alpha1.KogitoJobsServiceStatus{ExternalURI: URI},
	}
	dc := &oappsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "dc", Namespace: t.Name(), OwnerReferences: []metav1.OwnerReference{{
			Name: app.Name,
			UID:  app.UID,
		}}},
		Spec: oappsv1.DeploymentConfigSpec{
			Template: &corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "the-app"}}}},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{app, dc, jobs}, nil, nil)
	err := InjectJobsServicesURLIntoKogitoApps(cli, t.Name())
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

func TestInjectJobsServicesURLIntoKogitoAppsCleanUp(t *testing.T) {
	URI := "http://localhost:8080"
	app := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kogito-app",
			Namespace: t.Name(),
			UID:       types.UID(uuid.New().String()),
		},
	}
	jobs := &v1alpha1.KogitoJobsService{
		ObjectMeta: metav1.ObjectMeta{Name: "jobs-service", Namespace: t.Name()},
		Spec:       v1alpha1.KogitoJobsServiceSpec{Replicas: 1},
		Status:     v1alpha1.KogitoJobsServiceStatus{ExternalURI: URI},
	}
	dc := &oappsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "dc", Namespace: t.Name(), OwnerReferences: []metav1.OwnerReference{{
			Name: app.Name,
			UID:  app.UID,
		}}},
		Spec: oappsv1.DeploymentConfigSpec{
			Template: &corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "the-app"}}}},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{app, dc, jobs}, nil, nil)
	// first we inject
	err := InjectJobsServicesURLIntoKogitoApps(cli, t.Name())
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
	err = InjectJobsServicesURLIntoKogitoApps(cli, t.Name())
	assert.NoError(t, err)

	dc = &oappsv1.DeploymentConfig{ObjectMeta: metav1.ObjectMeta{Name: dc.Name, Namespace: dc.Namespace}}
	exists, err = kubernetes.ResourceC(cli).Fetch(dc)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Len(t, dc.Spec.Template.Spec.Containers[0].Env, 1)
	assert.Contains(t, dc.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  jobsServicesHTTPURIEnv,
		Value: "",
	})
}
