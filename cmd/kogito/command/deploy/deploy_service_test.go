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

package deploy

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/test"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"strings"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_DeployCmd_OperatorAutoInstal(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("deploy-service example-drools https://github.com/kiegroup/kogito-examples --context-dir drools-quarkus-example --project %s", ns)
	test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: v1alpha1.KogitoAppCRDName}})

	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "example-drools")
	assert.Contains(t, lines, "successfully created")
}

func Test_DeployCmd_CustomDeployment(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf(`deploy-service example-drools https://github.com/kiegroup/kogito-examples
								-v --context-dir drools-quarkus-example --project %s
								--image-s2i=myimage --image-runtime=myimage:0.2
								--limits cpu=1 --limits memory=1Gi --requests cpu=1,memory=1Gi
								--build-limits cpu=1 --build-limits memory=1Gi --build-requests cpu=1,memory=2Gi
								--install-infinispan Always`, ns)
	// Clean up after the command above
	cli = strings.Join(strings.Fields(cli), " ")
	ctx := test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: v1alpha1.KogitoAppCRDName}})
	// Start the test
	_, _, err := test.ExecuteCli()
	assert.NoError(t, err)

	// This should be created, given the command above
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-drools",
			Namespace: ns,
		},
	}

	exist, err := kubernetes.ResourceC(ctx.Client).Fetch(kogitoApp)
	assert.NoError(t, err)
	assert.True(t, exist)

	assert.NoError(t, err)
	assert.NotNil(t, kogitoApp)
	assert.Equal(t, v1alpha1.QuarkusRuntimeType, kogitoApp.Spec.Runtime)
	assert.Contains(t, kogitoApp.Spec.Resources.Limits, v1alpha1.ResourceMap{Resource: v1alpha1.ResourceCPU, Value: "1"})
	assert.Contains(t, kogitoApp.Spec.Resources.Requests, v1alpha1.ResourceMap{Resource: v1alpha1.ResourceMemory, Value: "1Gi"})
	assert.Contains(t, kogitoApp.Spec.Build.Resources.Limits, v1alpha1.ResourceMap{Resource: v1alpha1.ResourceCPU, Value: "1"})
	assert.Contains(t, kogitoApp.Spec.Build.Resources.Requests, v1alpha1.ResourceMap{Resource: v1alpha1.ResourceMemory, Value: "2Gi"})
	assert.Equal(t, kogitoApp.Spec.Build.ImageS2I.ImageStreamName, "myimage")
	assert.Equal(t, kogitoApp.Spec.Build.ImageRuntime.ImageStreamName, "myimage")
	assert.Equal(t, kogitoApp.Spec.Build.ImageRuntime.ImageStreamTag, "0.2")
	assert.Equal(t, v1alpha1.KogitoAppInfraInstallInfinispanAlways, kogitoApp.Spec.Infra.InstallInfinispan)
}

func Test_DeployCmd_CustomImage(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("deploy-service example-drools https://github.com/kiegroup/kogito-examples --native=false --context-dir drools-quarkus-example --project %s --image-s2i=openshift/myimage --image-runtime=openshift/myimage:0.2", ns)
	ctx := test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: v1alpha1.KogitoAppCRDName}})
	_, _, err := test.ExecuteCli()
	assert.NoError(t, err)

	instance := v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-drools",
			Namespace: ns,
		},
	}

	exists, err := kubernetes.ResourceC(ctx.Client).Fetch(&instance)
	assert.NoError(t, err)
	assert.True(t, exists)

	assert.Equal(t, "openshift", instance.Spec.Build.ImageS2I.ImageStreamNamespace)
	assert.Equal(t, "myimage", instance.Spec.Build.ImageS2I.ImageStreamName)

	assert.Equal(t, "openshift", instance.Spec.Build.ImageRuntime.ImageStreamNamespace)
	assert.Equal(t, "myimage", instance.Spec.Build.ImageRuntime.ImageStreamName)
	assert.Equal(t, "0.2", instance.Spec.Build.ImageRuntime.ImageStreamTag)

	assert.Equal(t, v1alpha1.KogitoAppInfraInstallInfinispanAuto, instance.Spec.Infra.InstallInfinispan)
}
