// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package project

import (
	"fmt"
	"github.com/apache/incubator-kie-kogito-operator/cmd/kogito/command/context"
	"github.com/apache/incubator-kie-kogito-operator/cmd/kogito/command/test"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
)

func TestNewProject_WhenNewProjectExist(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("new-project --project %s", ns)
	ctx := test.SetupCliTest(cli, context.CommandFactory{BuildCommands: BuildCommands}, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	lines, _, err := ctx.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "exists")
	assert.Contains(t, lines, ns)
}

func TestNewProject_WhenTheresNoNamedFlag(t *testing.T) {
	teardown := test.OverrideKubeConfigAndCreateDefaultContext()
	defer teardown()
	ns := t.Name()
	cli := fmt.Sprintf("new-project %s", ns)
	ctx := test.SetupCliTest(cli, context.CommandFactory{BuildCommands: BuildCommands})
	lines, _, err := ctx.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "created successfully")
	assert.Contains(t, lines, ns)
}

func TestNewProject_WhenTheresNoName(t *testing.T) {
	cli := "new-project"
	ctx := test.SetupCliTest(cli, context.CommandFactory{BuildCommands: BuildCommands})
	_, errLines, err := ctx.ExecuteCli()
	assert.Error(t, err)
	assert.Contains(t, errLines, "Please set a project for new-project")
}
