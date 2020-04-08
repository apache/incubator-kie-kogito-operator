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
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"k8s.io/apimachinery/pkg/api/resource"
	"sort"
	"strings"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_DeployCmd_OperatorAutoInstall(t *testing.T) {
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
								--image-s2i=quay.io/namespace/myimage:latest --image-runtime=quay.io/namespace/myimage:0.2
								--limits cpu=1 --limits memory=1Gi --requests cpu=1,memory=1Gi
								--build-limits cpu=1 --build-limits memory=1Gi --build-requests cpu=1,memory=2Gi
								--enable-persistence`, ns)
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
	assert.Equal(t, *kogitoApp.Spec.KogitoServiceSpec.Resources.Limits.Cpu(), resource.MustParse("1"))
	assert.Equal(t, *kogitoApp.Spec.KogitoServiceSpec.Resources.Requests.Memory(), resource.MustParse("1Gi"))
	assert.Equal(t, *kogitoApp.Spec.Build.Resources.Limits.Cpu(), resource.MustParse("1"))
	assert.Equal(t, *kogitoApp.Spec.Build.Resources.Requests.Memory(), resource.MustParse("2Gi"))
	assert.Equal(t, "quay.io/namespace/myimage:latest", kogitoApp.Spec.Build.ImageS2ITag)
	assert.Equal(t, "quay.io/namespace/myimage:0.2", kogitoApp.Spec.Build.ImageRuntimeTag)
	assert.True(t, kogitoApp.Spec.EnablePersistence)
}

func Test_DeployCmd_CustomImage(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("deploy-service example-drools https://github.com/kiegroup/kogito-examples --native=false --context-dir drools-quarkus-example --project %s --image-s2i=quay.io/namespace/myimage:latest --image-runtime=quay.io/namespace/myimage:0.2", ns)
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

	assert.Equal(t, "quay.io/namespace/myimage:latest", instance.Spec.Build.ImageS2ITag)
	assert.Equal(t, "quay.io/namespace/myimage:0.2", instance.Spec.Build.ImageRuntimeTag)

	assert.False(t, instance.Spec.EnablePersistence)
}

func Test_DeployCmd_CustomDeploymentWithMavenMirrorURL(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf(`deploy-service example-drools https://github.com/kiegroup/kogito-examples -v --context-dir drools-quarkus-example --project %s --maven-mirror-url https://local.nexus.localhost:8081/group/public`, ns)
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
	assert.NotNil(t, kogitoApp)
	assert.Equal(t, v1alpha1.QuarkusRuntimeType, kogitoApp.Spec.Runtime)
	assert.Equal(t, "https://local.nexus.localhost:8081/group/public", kogitoApp.Spec.Build.MavenMirrorURL)
}

func Test_DeployCmd_WithInvalidMavenMirrorURL(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("deploy-service example-drools https://github.com/kiegroup/kogito-examples --project %s --maven-mirror-url invalid-url", ns)
	test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: v1alpha1.KogitoAppCRDName}})
	_, _, err := test.ExecuteCli()
	assert.Error(t, err)
}

func Test_DeployCmd_WithoutGitURL(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("deploy-service example-drools -p %s", ns)
	test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: v1alpha1.KogitoAppCRDName}})
	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "example-drools")
	assert.Contains(t, lines, "-binary")
	assert.NotContains(t, lines, "You can see the deployment status")
}

func Test_DeployCmd_WrongGitURL(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("deploy-service example-drools invalid url -p %s", ns)
	test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: v1alpha1.KogitoAppCRDName}})
	_, _, err := test.ExecuteCli()
	assert.Error(t, err)
}

//see: https://issues.redhat.com/browse/KOGITO-1431
func Test_DeployCmd_CommasOnSlicesParameters(t *testing.T) {
	ns := t.Name()
	cli :=
		fmt.Sprintf("deploy-service process-example https://github.com/kiegroup/kogito-examples --context-dir process-quarkus-example --build-env MAVEN_ARGS_APPEND=-Ppersistence,events --env JAVA_OPTIONS=-Dvalue1=test1,value2=test2 --project %s", ns)
	ctx := test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: v1alpha1.KogitoAppCRDName}})

	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "process-example")
	assert.Contains(t, lines, "successfully created")

	kogitoApp := &v1alpha1.KogitoApp{ObjectMeta: metav1.ObjectMeta{Name: "process-example", Namespace: ns}}
	exists, err := kubernetes.ResourceC(ctx.Client).Fetch(kogitoApp)
	assert.True(t, exists)
	assert.NoError(t, err)
	assert.Len(t, kogitoApp.Spec.Build.Envs, 1)
	assert.Equal(t, "MAVEN_ARGS_APPEND", kogitoApp.Spec.Build.Envs[0].Name)
	assert.Equal(t, "-Ppersistence,events", kogitoApp.Spec.Build.Envs[0].Value)
	assert.Len(t, kogitoApp.Spec.Envs, 1)
	assert.Equal(t, "JAVA_OPTIONS", kogitoApp.Spec.Envs[0].Name)
	assert.Equal(t, "-Dvalue1=test1,value2=test2", kogitoApp.Spec.Envs[0].Value)
}

//see: https://issues.redhat.com/browse/KOGITO-1431
func Test_DeployCmd_CommasOnMultipleSlicesParameters(t *testing.T) {
	ns := t.Name()
	cli :=
		fmt.Sprintf("deploy-service process-example https://github.com/kiegroup/kogito-examples --context-dir process-quarkus-example --build-env MAVEN_ARGS_APPEND=-Ppersistence,events --build-env DEBUG=true --project %s", ns)
	ctx := test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: v1alpha1.KogitoAppCRDName}})

	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "process-example")
	assert.Contains(t, lines, "successfully created")

	kogitoApp := &v1alpha1.KogitoApp{ObjectMeta: metav1.ObjectMeta{Name: "process-example", Namespace: ns}}
	exists, err := kubernetes.ResourceC(ctx.Client).Fetch(kogitoApp)
	assert.True(t, exists)
	assert.NoError(t, err)
	assert.Len(t, kogitoApp.Spec.Build.Envs, 2)

	sort.Slice(kogitoApp.Spec.Build.Envs, func(i, j int) bool {
		return kogitoApp.Spec.Build.Envs[i].Name < kogitoApp.Spec.Build.Envs[j].Name
	})

	assert.Equal(t, "DEBUG", kogitoApp.Spec.Build.Envs[0].Name)
	assert.Equal(t, "true", kogitoApp.Spec.Build.Envs[0].Value)
	assert.Equal(t, "MAVEN_ARGS_APPEND", kogitoApp.Spec.Build.Envs[1].Name)
	assert.Equal(t, "-Ppersistence,events", kogitoApp.Spec.Build.Envs[1].Value)
}

func Test_NoManagementConsole(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("deploy-service example-drools https://github.com/kiegroup/kogito-examples --context-dir drools-quarkus-example --project %s", ns)
	test.SetupCliTest(cli, context.CommandFactory{BuildCommands: BuildCommands}, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})

	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "example-drools")
	assert.Contains(t, lines, "successfully created")
	assert.Contains(t, lines, "install Data Index Service and Process Instance Management")
}

func Test_WithManagementConsole(t *testing.T) {
	ns := t.Name()
	mgmtConsoleURI := "http://mgmtconsole.com"
	cli := fmt.Sprintf("deploy-service example-drools https://github.com/kiegroup/kogito-examples --context-dir drools-quarkus-example --project %s", ns)
	test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&v1alpha1.KogitoMgmtConsole{
			ObjectMeta: metav1.ObjectMeta{Name: infrastructure.DefaultMgmtConsoleName, Namespace: ns},
			Status:     v1alpha1.KogitoMgmtConsoleStatus{KogitoServiceStatus: v1alpha1.KogitoServiceStatus{ExternalURI: mgmtConsoleURI}}})

	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "example-drools")
	assert.Contains(t, lines, "successfully created")
	assert.Contains(t, lines, mgmtConsoleURI)
}
