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
	"github.com/kiegroup/kogito-cloud-operator/community-kogito-operator/core/kogitosupportingservice"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/community-kogito-operator/cmd/kogito/command/context"

	"github.com/kiegroup/kogito-cloud-operator/community-kogito-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/community-kogito-operator/cmd/kogito/command/test"
	"github.com/kiegroup/kogito-cloud-operator/community-kogito-operator/core/client/kubernetes"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_InstallSupportingServiceCmd_DefaultConfiguration(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("install data-index --project %s", ns)
	ctx := test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	lines, _, err := test.ExecuteCli()

	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Data Index Service successfully installed")

	// This should be created, given the command above
	dataIndex := &v1beta1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kogitosupportingservice.DefaultDataIndexName,
			Namespace: ns,
		},
	}

	exist, err := kubernetes.ResourceC(ctx.Client).Fetch(dataIndex)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.NotNil(t, dataIndex)
	assert.False(t, dataIndex.Spec.InsecureImageRegistry)
}

func Test_InstallSupportingServiceCmd_CustomConfiguration(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("install data-index --project %s --insecure-image-registry --infra kogito-kafka --infra kogito-infinispan --liveness-initial-delay 5 --readiness-initial-delay 6", ns)
	ctx := test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	lines, _, err := test.ExecuteCli()

	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Data Index Service successfully installed")

	// This should be created, given the command above
	dataIndex := &v1beta1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kogitosupportingservice.DefaultDataIndexName,
			Namespace: ns,
		},
	}

	exist, err := kubernetes.ResourceC(ctx.Client).Fetch(dataIndex)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.NotNil(t, dataIndex)
	assert.True(t, dataIndex.Spec.InsecureImageRegistry)
	assert.Contains(t, dataIndex.Spec.Infra, "kogito-kafka")
	assert.Contains(t, dataIndex.Spec.Infra, "kogito-infinispan")
	assert.Equal(t, int32(5), dataIndex.Spec.Probes.LivenessProbe.InitialDelaySeconds)
	assert.Equal(t, int32(6), dataIndex.Spec.Probes.ReadinessProbe.InitialDelaySeconds)
}
