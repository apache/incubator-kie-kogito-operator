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

package infrastructure

import (
	"testing"

	"github.com/google/uuid"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	oappsv1 "github.com/openshift/api/apps/v1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func TestInjectTrustyURLIntoKogitoApps(t *testing.T) {
	ns := t.Name()
	name := "my-kogito-app"
	expectedRoute := "http://trusty-route.com"
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			UID:       types.UID(uuid.New().String()),
		},
	}
	dc := &oappsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{Name: kogitoApp.Name, Namespace: kogitoApp.Namespace, OwnerReferences: []metav1.OwnerReference{{UID: kogitoApp.UID}}},
		Spec: oappsv1.DeploymentConfigSpec{
			Template: &v1.PodTemplateSpec{
				Spec: v1.PodSpec{Containers: []v1.Container{{Name: "test"}}},
			},
		},
	}
	trustyServices := &v1alpha1.KogitoTrustyList{
		Items: []v1alpha1.KogitoTrusty{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      DefaultTrustyName,
					Namespace: ns,
				},
				Status: v1alpha1.KogitoTrustyStatus{
					KogitoServiceStatus: v1alpha1.KogitoServiceStatus{ExternalURI: expectedRoute},
				},
			},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{kogitoApp, trustyServices, dc}, nil, nil)

	err := InjectTrustyURLIntoKogitoApps(cli, ns)
	assert.NoError(t, err)

	exist, err := kubernetes.ResourceC(cli).Fetch(dc)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Contains(t, dc.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{Name: trustyHTTPRouteEnv, Value: expectedRoute})
}
