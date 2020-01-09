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
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	appsv1 "github.com/openshift/api/apps/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
)

func Test_InjectEnvironmentVarsFromExternalServices(t *testing.T) {
	ns := t.Name()
	name := "my-kogito-app"
	expectedRoute := "http://dataindex-route.com"

	kogitoApp := &v1alpha1.KogitoApp{ObjectMeta: v1.ObjectMeta{Name: name, Namespace: ns}}
	dc := &appsv1.DeploymentConfig{
		ObjectMeta: v1.ObjectMeta{Name: name, Namespace: ns},
		Spec:       appsv1.DeploymentConfigSpec{Template: &corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: name}}}}},
	}
	dataIndexes := &v1alpha1.KogitoDataIndexList{Items: []v1alpha1.KogitoDataIndex{{
		ObjectMeta: v1.ObjectMeta{Name: "kogito-data-index", Namespace: ns},
		Status:     v1alpha1.KogitoDataIndexStatus{Route: expectedRoute}},
	}}

	client := test.CreateFakeClient([]runtime.Object{dc, dataIndexes}, nil, nil)
	err := InjectEnvVarsFromExternalServices(kogitoApp, &dc.Spec.Template.Spec.Containers[0], client)
	assert.NoError(t, err)
	assert.Contains(t, dc.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: dataIndexHTTPRouteEnv, Value: expectedRoute})
}

func Test_InjectEnvironmentVarsFromExternalServices_NewRoute(t *testing.T) {
	ns := t.Name()
	name := "my-kogito-app"
	oldRoute := "http://dataindex-route.com"
	expectedRoute := "http://dataindex-route2.com"

	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: v1.ObjectMeta{Name: name, Namespace: ns},
		Spec: v1alpha1.KogitoAppSpec{
			Env: []v1alpha1.Env{
				{
					Name:  dataIndexHTTPRouteEnv,
					Value: oldRoute,
				},
			},
		},
	}
	dc := &appsv1.DeploymentConfig{
		ObjectMeta: v1.ObjectMeta{Name: name, Namespace: ns},
		Spec:       appsv1.DeploymentConfigSpec{Template: &corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: name}}}}},
	}
	dataIndexes := &v1alpha1.KogitoDataIndexList{
		Items: []v1alpha1.KogitoDataIndex{
			{
				ObjectMeta: v1.ObjectMeta{
					Name:      "kogito-data-index",
					Namespace: ns,
				},
				Status: v1alpha1.KogitoDataIndexStatus{
					Route: expectedRoute,
				},
			},
		},
	}
	client := test.CreateFakeClient([]runtime.Object{kogitoApp, dataIndexes}, nil, nil)
	err := InjectEnvVarsFromExternalServices(kogitoApp, &dc.Spec.Template.Spec.Containers[0], client)
	assert.NoError(t, err)
	assert.Contains(t, dc.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: dataIndexHTTPRouteEnv, Value: expectedRoute})
}
