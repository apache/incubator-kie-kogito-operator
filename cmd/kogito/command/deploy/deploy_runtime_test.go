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
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/test"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

// kogitoRuntimeCRDName is the name of the Kogito Runtime CRD in the cluster
const kogitoRuntimeCRDName = "kogitoruntime.app.kiegroup.org"

func Test_DeployRuntimeCmd_DefaultConfigurations(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("deploy-service example-drools --image quay.io/kiegroup/drools-quarkus-example:1.0 --project %s", ns)
	ctx := test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: kogitoRuntimeCRDName}})

	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Service successfully installed")

	// This should be created, given the command above
	kogitoRuntime := &v1alpha1.KogitoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-drools",
			Namespace: ns,
		},
	}

	exist, err := kubernetes.ResourceC(ctx.Client).Fetch(kogitoRuntime)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, "quay.io", kogitoRuntime.Spec.Image.Domain)
	assert.Equal(t, "kiegroup", kogitoRuntime.Spec.Image.Namespace)
	assert.Equal(t, "drools-quarkus-example", kogitoRuntime.Spec.Image.Name)
	assert.Equal(t, "1.0", kogitoRuntime.Spec.Image.Tag)
	assert.Equal(t, v1alpha1.QuarkusRuntimeType, kogitoRuntime.Spec.Runtime)
	assert.False(t, kogitoRuntime.Spec.InfinispanMeta.InfinispanProperties.UseKogitoInfra)
	assert.False(t, kogitoRuntime.Spec.KafkaMeta.KafkaProperties.UseKogitoInfra)
	assert.False(t, kogitoRuntime.Spec.EnableIstio)
	assert.Equal(t, int32(1), *kogitoRuntime.Spec.Replicas)
	assert.Equal(t, int32(8080), kogitoRuntime.Spec.HTTPPort)
	assert.False(t, kogitoRuntime.Spec.InsecureImageRegistry)
}

func Test_DeployRuntimeCmd_CustomConfigurations(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf(`deploy-service example-drools --image quay.io/kiegroup/drools-quarkus-example:1.0 --project %s --limits cpu=1 --limits memory=1Gi --requests cpu=1,memory=1Gi --enable-istio --enable-persistence --enable-events --http-port 9090 --runtime springboot --replicas 2 --insecure-image-registry`, ns)
	ctx := test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: kogitoRuntimeCRDName}})

	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Service successfully installed")

	// This should be created, given the command above
	kogitoRuntime := &v1alpha1.KogitoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-drools",
			Namespace: ns,
		},
	}

	exist, err := kubernetes.ResourceC(ctx.Client).Fetch(kogitoRuntime)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, "quay.io", kogitoRuntime.Spec.Image.Domain)
	assert.Equal(t, "kiegroup", kogitoRuntime.Spec.Image.Namespace)
	assert.Equal(t, "drools-quarkus-example", kogitoRuntime.Spec.Image.Name)
	assert.Equal(t, "1.0", kogitoRuntime.Spec.Image.Tag)
	assert.Equal(t, v1alpha1.SpringbootRuntimeType, kogitoRuntime.Spec.Runtime)
	assert.True(t, kogitoRuntime.Spec.InfinispanMeta.InfinispanProperties.UseKogitoInfra)
	assert.True(t, kogitoRuntime.Spec.KafkaMeta.KafkaProperties.UseKogitoInfra)
	assert.True(t, kogitoRuntime.Spec.EnableIstio)
	assert.Equal(t, int32(2), *kogitoRuntime.Spec.Replicas)
	assert.Equal(t, int32(9090), kogitoRuntime.Spec.HTTPPort)
	assert.Equal(t, *kogitoRuntime.Spec.KogitoServiceSpec.Resources.Limits.Cpu(), resource.MustParse("1"))
	assert.Equal(t, *kogitoRuntime.Spec.KogitoServiceSpec.Resources.Requests.Memory(), resource.MustParse("1Gi"))
	assert.True(t, kogitoRuntime.Spec.InsecureImageRegistry)
}
