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

package project

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/test"
	"testing"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_DeleteProjectCmd_WhenWeSuccessfullyDelete(t *testing.T) {
	teardown := test.OverrideKubeConfigAndCreateDefaultContext()
	defer teardown()
	ns := t.Name()
	cli := fmt.Sprintf("delete-project %s", ns)
	test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, fmt.Sprintf("Successfully deleted Kogito Project %s", ns))
}

func Test_DeleteProjectCmd_WhenProjectDoesNotExist(t *testing.T) {
	teardown := test.OverrideKubeConfigAndCreateDefaultContext()
	defer teardown()
	ns := t.Name()
	cli := fmt.Sprintf("delete-project %s", ns)
	test.SetupCliTest(cli, context.CommandFactory{BuildCommands: BuildCommands})
	_, errLines, err := test.ExecuteCli()
	assert.Error(t, err)
	assert.Contains(t, errLines, fmt.Sprintf("Project %s not found", ns))
}
