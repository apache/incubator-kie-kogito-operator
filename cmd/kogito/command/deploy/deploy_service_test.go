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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_DeployServiceCmd_DefaultConfigurations(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("deploy-service example-drools --project %s --image quay.io/kiegroup/drools-quarkus-example:1.0", ns)
	ctx := test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})

	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "Image details are provided, skipping to install kogito build")
	assert.Contains(t, lines, "Kogito Service successfully installed in the Project")

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
