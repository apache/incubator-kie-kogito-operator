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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
)

//context.CommandFactory{BuildCommands:BuildCommands},

func TestNewProject_WhenNewProjectDoesNotExist(t *testing.T) {
	cli := fmt.Sprintf("new-project --project test")
	test.SetupCliTest(cli, context.CommandFactory{BuildCommands: BuildCommands})
	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "created")
}

func TestNewProject_WhenNewProjectExist(t *testing.T) {
	cli := fmt.Sprintf("new-project --project test")
	test.SetupCliTest(cli, context.CommandFactory{BuildCommands: BuildCommands}, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test"}})
	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "exists")
}

func TestNewProject_WhenTheresNoNamedFlag(t *testing.T) {
	cli := fmt.Sprintf("new-project test1")
	test.SetupCliTest(cli, context.CommandFactory{BuildCommands: BuildCommands})
	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "created")
}

func TestNewProject_WhenTheresNoName(t *testing.T) {
	cli := fmt.Sprintf("new-project")
	test.SetupCliTest(cli, context.CommandFactory{BuildCommands: BuildCommands})
	lines, _, err := test.ExecuteCli()
	assert.Error(t, err)
	assert.Contains(t, lines, "Please set a project for new-project")
}
