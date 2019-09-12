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

package main

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
)

func TestNewProject_WhenNewProjectDoesNotExist(t *testing.T) {
	cli := fmt.Sprintf("new-project --name test")
	lines, _, err := executeCli(cli)
	assert.NoError(t, err)
	assert.Contains(t, lines, "created")
}

func TestNewProject_WhenNewProjectExist(t *testing.T) {
	cli := fmt.Sprintf("new-project --name test")
	lines, _, err := executeCli(cli, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test"}})
	assert.NoError(t, err)
	assert.Contains(t, lines, "exists")
}

func TestNewProject_WhenTheresNoNamedFlag(t *testing.T) {
	cli := fmt.Sprintf("new-project test1")
	lines, _, err := executeCli(cli)
	assert.NoError(t, err)
	assert.Contains(t, lines, "created")
}

func TestNewProject_WhenTheresNoName(t *testing.T) {
	cli := fmt.Sprintf("new-project")
	lines, _, err := executeCli(cli)
	assert.Error(t, err)
	assert.Contains(t, lines, "Please set a name for new-project")
}
