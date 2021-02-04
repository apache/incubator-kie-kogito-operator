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

package deploy

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/test"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	test2 "github.com/kiegroup/kogito-cloud-operator/pkg/test"
	v1 "github.com/openshift/api/build/v1"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"testing"
)

func Test_DeployServiceCmd_DefaultConfigurations(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("deploy-service example-drools --project %s --image quay.io/kiegroup/drools-quarkus-example:1.0 --env myvar1=myvalue1 --secret-env myvar2=mysecretName2#mysecretKey2", ns)
	ctx := test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})

	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "Image details are provided, skipping to install kogito build")
	assert.Contains(t, lines, "Kogito Service successfully installed in the Project")

	// This should be created, given the command above
	kogitoRuntime := &v1beta1.KogitoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-drools",
			Namespace: ns,
		},
	}

	exist, err := kubernetes.ResourceC(ctx.Client).Fetch(kogitoRuntime)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, "quay.io/kiegroup/drools-quarkus-example:1.0", kogitoRuntime.Spec.Image)
	assert.Equal(t, v1beta1.QuarkusRuntimeType, kogitoRuntime.Spec.Runtime)
	assert.False(t, kogitoRuntime.Spec.EnableIstio)
	assert.Equal(t, int32(1), *kogitoRuntime.Spec.Replicas)
	assert.False(t, kogitoRuntime.Spec.InsecureImageRegistry)
	assert.Equal(t, 2, len(kogitoRuntime.Spec.Env))
}

func Test_DeployCmd_WithCustomImage(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf(`deploy-service process-business-rules-quarkus --image localhost:5000/kiegroup/process-business-rules-quarkus --project %s`, ns)
	ctx := test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "Image details are provided, skipping to install kogito build")
	assert.Contains(t, lines, "Kogito Service successfully installed in the Project")

	// This should be created, given the command above
	kogitoRuntime := &v1beta1.KogitoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "process-business-rules-quarkus",
			Namespace: ns,
		},
	}

	exist, err := kubernetes.ResourceC(ctx.Client).Fetch(kogitoRuntime)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, "localhost:5000/kiegroup/process-business-rules-quarkus", kogitoRuntime.Spec.Image)
	assert.Equal(t, v1beta1.QuarkusRuntimeType, kogitoRuntime.Spec.Runtime)
	assert.False(t, kogitoRuntime.Spec.EnableIstio)
	assert.Equal(t, int32(1), *kogitoRuntime.Spec.Replicas)
	assert.False(t, kogitoRuntime.Spec.InsecureImageRegistry)
	assert.Equal(t, 0, len(kogitoRuntime.Spec.Env))
}

func Test_DeployCmd_WithCustomConfig(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "application.properties")
	assert.NoError(t, err)
	properties := `
quarkus.log.level=DEBUG
my.nice.property=socool
`
	err = ioutil.WriteFile(tempFile.Name(), []byte(properties), 0644)
	assert.NoError(t, err)

	ns := t.Name()
	cli := fmt.Sprintf(`deploy-service process-business-rules-quarkus --image docker.io/ns/mycoolimage --config-file %s --project %s`, tempFile.Name(), ns)
	ctx := test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Service successfully installed in the Project")

	// This should be created, given the command above
	kogitoRuntime := &v1beta1.KogitoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "process-business-rules-quarkus",
			Namespace: ns,
		},
	}

	exist, err := kubernetes.ResourceC(ctx.Client).Fetch(kogitoRuntime)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, v1beta1.QuarkusRuntimeType, kogitoRuntime.Spec.Runtime)
	assert.False(t, kogitoRuntime.Spec.EnableIstio)
	assert.Equal(t, int32(1), *kogitoRuntime.Spec.Replicas)
	assert.False(t, kogitoRuntime.Spec.InsecureImageRegistry)
	assert.Equal(t, 0, len(kogitoRuntime.Spec.Env))
	assert.NotEmpty(t, kogitoRuntime.Spec.PropertiesConfigMap)

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Namespace: t.Name(), Name: kogitoRuntime.Spec.PropertiesConfigMap},
	}
	exists, err := kubernetes.ResourceC(ctx.Client).Fetch(cm)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Contains(t, cm.Data[services.ConfigMapApplicationPropertyKey], "quarkus.log.level")
}

func Test_DeployCmd_SWFile(t *testing.T) {
	ns := t.Name()
	swFile, err := os.Open("testdata/greetings.sw.json")
	assert.NoError(t, err)
	assert.NotNil(t, swFile)

	cli := fmt.Sprintf(`deploy-service serverless-workflow-greeting-quarkus %s --project %s`, swFile.Name(), ns)
	ctx := test.SetupCliTestWithKubeClient(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		test2.NewFakeClientBuilder().
			OnOpenShift().
			AddK8sObjects(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}).
			AddBuildObjects(&v1.BuildConfig{ObjectMeta: metav1.ObjectMeta{Name: "serverless-workflow-greeting-quarkus-builder", Namespace: ns}}).
			Build())

	lines, errLines, err := test.ExecuteCli()
	assert.Error(t, err)
	assert.Contains(t, lines, "Kogito Build Service successfully installed in the Project")
	assert.Contains(t, lines, "File(s) found: testdata/greetings.sw.json")
	assert.Contains(t, lines, "Triggering the new build")
	// error from fake build is ok, we don't have a server to upload the binaries here.
	assert.Contains(t, errLines, "v1.BinaryBuildRequestOptions is not suitable for converting")

	kogitoBuild := &v1beta1.KogitoBuild{
		ObjectMeta: metav1.ObjectMeta{Name: "serverless-workflow-greeting-quarkus", Namespace: ns},
	}
	exists, err := kubernetes.ResourceC(ctx.Client).Fetch(kogitoBuild)
	assert.NoError(t, err)
	assert.True(t, exists)
}
