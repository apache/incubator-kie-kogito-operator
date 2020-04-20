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
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/test"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUseProjectCmd_WhenTheresNoConfigAndNoNamespace(t *testing.T) {
	teardown := test.OverrideKubeConfig()
	defer teardown()
	ns := uuid.New().String()
	test.SetupCliTest(strings.Join([]string{"use-project", ns}, " "), context.CommandFactory{BuildCommands: BuildCommands})
	_, _, err := test.ExecuteCli()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), ns)
}

func TestUseProjectCmd_WhenThereIsTheNamespace(t *testing.T) {
	teardown := test.OverrideKubeConfigAndCreateDefaultContext()
	defer teardown()
	ns := uuid.New().String()
	nsObj := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
	test.SetupCliTest(strings.Join([]string{"use-project", ns}, " "), context.CommandFactory{BuildCommands: BuildCommands}, nsObj)
	o, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.NotEmpty(t, o)
	assert.Contains(t, o, ns)
}

func TestUseProjectCmd_WhenWhatIsTheNamespace_ConfigUpdated(t *testing.T) {
	teardown := test.OverrideKubeConfigAndCreateDefaultContext()
	defer teardown()
	ns := t.Name()
	nsObj := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
	// Set the project
	test.SetupCliTest(strings.Join([]string{"use-project", ns}, " "), context.CommandFactory{BuildCommands: BuildCommands}, nsObj)
	o1, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, o1, ns)
	assert.Equal(t, ns, shared.GetCurrentNamespaceFromKubeConfig())
}

func TestUseProjectCmd_WhenWhatIsTheNamespace_UseConfigNamespace(t *testing.T) {
	ns := t.Name()
	teardown := test.OverrideKubeConfigAndCreateContextInNamespace(ns)
	defer teardown()
	nsObj := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
	test.SetupCliTest("use-project", context.CommandFactory{BuildCommands: BuildCommands}, nsObj)
	o2, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.NotEmpty(t, o2)
	assert.Contains(t, o2, ns)
}
