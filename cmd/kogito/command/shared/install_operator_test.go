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
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/operator"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	operators "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"

	"github.com/stretchr/testify/assert"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"testing"
)

// ServiceAccountName is the name of service account used by Kogito Services Runtimes
const serviceAccountName = "kogito-service-viewer"

func Test_InstallOperatorWithYaml(t *testing.T) {
	ns := t.Name()
	client := test.NewFakeClientBuilder().AddK8sObjects(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}).Build()
	image := "docker.io/myrepo/custom-operator:1.0"

	err := installOperatorWithYamlFiles(image, ns, false, client)
	assert.NoError(t, err)

	serviceAccount := v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: ns,
		},
	}

	_, err = kubernetes.ResourceC(client).Fetch(&serviceAccount)
	assert.NoError(t, err)
	assert.Equal(t, serviceAccountName, serviceAccount.Name)

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

	// checks CRD
	crds := &apiextensionsv1beta1.CustomResourceDefinitionList{}
	err = kubernetes.ResourceC(client).ListWithNamespace(ns, crds)
	assert.NoError(t, err)
	assert.Len(t, crds.Items, 4)
	assert.Contains(t, crds.Items[0].Name, "app.kiegroup.org")
	assert.Contains(t, crds.Items[1].Name, "app.kiegroup.org")
	assert.Contains(t, crds.Items[2].Name, "app.kiegroup.org")
}

func TestMustInstallOperatorIfNotExists_WithOperatorHubOnOpenshift(t *testing.T) {
	ns := operatorOpenshiftMarketplaceNamespace
	packageManifest := &operators.PackageManifest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultOperatorPackageName,
			Namespace: openshiftGlobalOperatorNamespace,
		},
	}
	client := test.NewFakeClientBuilder().AddK8sObjects(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: openshiftGlobalOperatorNamespace}}, packageManifest).OnOpenShift().SupportOLM().Build()
	// Operator is there in the hub and not exists in the given namespace, let's check if there's no error
	err := InstallOperatorIfNotExists(ns, defaultOperatorImageName, client, false, false, GetDefaultChannel(), false)
	assert.NoError(t, err)

	subs := &v1alpha1.Subscription{ObjectMeta: metav1.ObjectMeta{Name: defaultOperatorPackageName, Namespace: openshiftGlobalOperatorNamespace}}
	exists, err := kubernetes.ResourceC(client).Fetch(subs)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestMustInstallOperatorIfNotExists_WithOperatorHubOnKubernetes(t *testing.T) {
	ns := operatorOpenshiftMarketplaceNamespace
	packageManifest := &operators.PackageManifest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultOperatorPackageName,
			Namespace: kubernetesGlobalOperatorNamespace,
		},
	}
	client := test.NewFakeClientBuilder().AddK8sObjects(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: kubernetesGlobalOperatorNamespace}}, packageManifest).SupportOLM().Build()
	// Operator is there in the hub and not exists in the given namespace, let's check if there's no error
	err := InstallOperatorIfNotExists(ns, defaultOperatorImageName, client, false, false, GetDefaultChannel(), false)
	assert.NoError(t, err)

	subs := &v1alpha1.Subscription{ObjectMeta: metav1.ObjectMeta{Name: defaultOperatorPackageName, Namespace: kubernetesGlobalOperatorNamespace}}
	exists, err := kubernetes.ResourceC(client).Fetch(subs)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestMustInstallOperatorIfNotExists_WithoutOperatorHub(t *testing.T) {
	ns := t.Name()

	client := test.NewFakeClientBuilder().AddK8sObjects(
		&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: operatorNamespace}}).Build()

	// Operator is not in the hub. Install with yaml files.
	err := InstallOperatorIfNotExists(ns, defaultOperatorImageName, client, false, false, GetDefaultChannel(), false)
	assert.NoError(t, err)
	// Operator is now in the hub, but no pods are running because this is a controlled test environment
	exist, err := infrastructure.CheckKogitoOperatorExists(client, operatorNamespace)
	assert.Error(t, err)
	assert.True(t, exist)
	assert.Contains(t, err.Error(), "kogito-operator Operator seems to be created in the namespace")
}
