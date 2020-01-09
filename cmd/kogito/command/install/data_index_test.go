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

package install

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/test"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_DeployDataIndexCmd(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("install data-index --project %s --infinispan-url myservice:11222 --kafka-url my-cluster:9092", ns)
	test.SetupCliTest(cli, context.CommandFactory{BuildCommands: BuildCommands}, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	lines, _, err := test.ExecuteCli()

	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Data Index Service successfully installed")
}

func Test_DeployDataIndexCmd_RequiredFlags(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("install data-index --project %s", ns)
	test.SetupCliTest(cli, context.CommandFactory{BuildCommands: BuildCommands}, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	lines, _, err := test.ExecuteCli()

	assert.Error(t, err)
	assert.Contains(t, lines, "Infinispan Operator is not available in the Project")
}

func Test_DeployDataIndexCmd_SuccessfulDeployWithoutKafkaURI(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("install data-index --project %s --infinispan-url myservice:11222 --kafka-instance my-cluster", ns)
	test.SetupCliTest(cli, context.CommandFactory{BuildCommands: BuildCommands}, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	lines, _, err := test.ExecuteCli()

	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Data Index Service successfully installed")
}

func Test_DeployDataIndexCmd_SuccessfulDeployWithKafkaInstance(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("install data-index --project %s --infinispan-url myservice:11222", ns)
	test.SetupCliTest(cli, context.CommandFactory{BuildCommands: BuildCommands}, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	lines, _, err := test.ExecuteCli()

	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Data Index Service successfully installed")
}

func Test_DeployDataIndexCmd_SuccessfulDeploy(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("install data-index --project %s --infinispan-url myservice:11222 --kafka-url my-cluster:9092", ns)
	test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: v1alpha1.KogitoDataIndexCRDName}})
	lines, _, err := test.ExecuteCli()

	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Data Index Service successfully installed")
}

func Test_DeployDataIndexCmd_SuccessfulDeployWithInfinispanCredentials(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("install data-index --project %s --infinispan-url myservice:11222 --kafka-url my-cluster:9092 --infinispan-user user --infinispan-password password", ns)
	test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: v1alpha1.KogitoDataIndexCRDName}})
	lines, _, err := test.ExecuteCli()

	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Data Index Service successfully installed")
}

func Test_DeployDataIndexCmd_SuccessfulDeployWithInfinispanCredentialsAndSecret(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("install data-index --project %s --infinispan-url myservice:11222 --kafka-url my-cluster:9092 --infinispan-user user --infinispan-password password", ns)
	ctx := test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: v1alpha1.KogitoDataIndexCRDName}},
		&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: defaultDataIndexInfinispanSecretName, Namespace: ns}})
	lines, _, err := test.ExecuteCli()

	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Data Index Service successfully installed")

	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
		Name:      defaultDataIndexInfinispanSecretName,
		Namespace: ns,
	}}
	exists, err := kubernetes.ResourceC(ctx.Client).Fetch(secret)
	assert.NoError(t, err)
	assert.NotNil(t, secret)
	assert.True(t, exists)
	assert.Contains(t, secret.StringData, defaultInfinispanUsernameKey, defaultInfinispanPasswordKey)
}
