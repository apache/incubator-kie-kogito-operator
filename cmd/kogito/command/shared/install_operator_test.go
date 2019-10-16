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

package shared

import (
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/test"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/resource"
	"github.com/kiegroup/kogito-cloud-operator/pkg/operator"

	"github.com/stretchr/testify/assert"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"testing"
)

func Test_InstallOperatorWithYaml(t *testing.T) {
	ns := t.Name()
	client := test.SetupFakeKubeCli(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kogito"}})
	image := "docker.io/myrepo/custom-operator:1.0"

	err := installOperatorWithYamlFiles(image, ns, client)
	assert.NoError(t, err)

	serviceAccount := v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resource.ServiceAccountName,
			Namespace: ns,
		},
	}

	_, err = kubernetes.ResourceC(client).Fetch(&serviceAccount)
	assert.NoError(t, err)
	assert.Equal(t, resource.ServiceAccountName, serviceAccount.Name)

	serviceAccount = v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      operator.Name,
			Namespace: ns,
		},
	}

	_, err = kubernetes.ResourceC(client).Fetch(&serviceAccount)
	assert.NoError(t, err)
	assert.Equal(t, operator.Name, serviceAccount.Name)

	deployment := &apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      operator.Name,
			Namespace: ns,
		},
	}
	_, err = kubernetes.ResourceC(client).Fetch(deployment)
	assert.NoError(t, err)
	assert.Equal(t, operator.Name, deployment.Name)
	assert.Equal(t, image, deployment.Spec.Template.Spec.Containers[0].Image)
}
