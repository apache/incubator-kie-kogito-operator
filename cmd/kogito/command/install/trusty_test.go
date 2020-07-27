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
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"

	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"

	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/test"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_DeployTrustyCmd_DefaultConfiguration(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("install trusty --project %s", ns)
	ctx := test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: v1alpha1.KogitoTrustyCRDName}})
	lines, _, err := test.ExecuteCli()

	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Trusty Service successfully installed")

	// This should be created, given the command above
	trusty := &v1alpha1.KogitoTrusty{
		ObjectMeta: metav1.ObjectMeta{
			Name:      infrastructure.DefaultTrustyName,
			Namespace: ns,
		},
	}

	exist, err := kubernetes.ResourceC(ctx.Client).Fetch(trusty)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.NotNil(t, trusty)
	assert.False(t, trusty.Spec.InsecureImageRegistry)
	assert.True(t, trusty.Spec.InfinispanProperties.UseKogitoInfra)
	assert.True(t, trusty.Spec.KafkaProperties.UseKogitoInfra)
}

func Test_DeployTrustyCmd_CustomConfiguration(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("install data-index --project %s --infinispan-url myservice:11222 --kafka-url my-cluster:9092 --infinispan-user user --infinispan-password password --insecure-image-registry --http-port 9090", ns)
	ctx := test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: v1alpha1.KogitoTrustyCRDName}})
	lines, _, err := test.ExecuteCli()

	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Trusty Service successfully installed")

	// This should be created, given the command above
	trusty := &v1alpha1.KogitoTrusty{
		ObjectMeta: metav1.ObjectMeta{
			Name:      infrastructure.DefaultTrustyName,
			Namespace: ns,
		},
	}

	exist, err := kubernetes.ResourceC(ctx.Client).Fetch(trusty)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.NotNil(t, trusty)
	assert.True(t, trusty.Spec.InsecureImageRegistry)
	assert.False(t, trusty.Spec.InfinispanProperties.UseKogitoInfra)
	assert.False(t, trusty.Spec.KafkaProperties.UseKogitoInfra)
	assert.Equal(t, int32(9090), trusty.Spec.HTTPPort)
}
