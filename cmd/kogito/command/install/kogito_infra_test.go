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

package install

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/core/infrastructure"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/test"
	"github.com/kiegroup/kogito-cloud-operator/core/client/kubernetes"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_InstallInfraServiceCmd_DefaultConfiguration(t *testing.T) {
	name := "kogito-infinispan-infra"
	ns := t.Name()
	cli := fmt.Sprintf("install infra %s --project %s --apiVersion %s --kind %s --resource-name %s", name, ns, infrastructure.InfinispanAPIVersion, infrastructure.InfinispanKind, "my-infinispan")
	ctx := test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&v1beta1.KogitoInfra{ObjectMeta: metav1.ObjectMeta{Name: "kogito-infra", Namespace: ns}})
	lines, _, err := test.ExecuteCli()

	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Infra Service successfully installed")

	// This should be created, given the command above
	kogitoInfra := &v1beta1.KogitoInfra{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}

	exist, err := kubernetes.ResourceC(ctx.Client).Fetch(kogitoInfra)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.NotNil(t, kogitoInfra)
	assert.Equal(t, name, kogitoInfra.Name)
	assert.Equal(t, ns, kogitoInfra.Namespace)
	assert.Equal(t, infrastructure.InfinispanAPIVersion, kogitoInfra.Spec.Resource.APIVersion)
	assert.Equal(t, infrastructure.InfinispanKind, kogitoInfra.Spec.Resource.Kind)
	assert.Equal(t, "my-infinispan", kogitoInfra.Spec.Resource.Name)
}
