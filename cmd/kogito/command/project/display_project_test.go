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
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/test"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDisplayProjectCmd_WhenTheresNoConfigAndNoNamespace(t *testing.T) {
	teardown := test.OverrideKubeConfig()
	defer teardown()

	test.SetupCliTest(strings.Join([]string{"project"}, " "), context.CommandFactory{BuildCommands: BuildCommands})
	o, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, o, "No project configured")
}

func TestDisplayProjectCmd_WhenThereIsTheNamespace(t *testing.T) {
	ns := uuid.New().String()
	teardown := test.OverrideKubeConfigAndCreateDefaultContext()
	defer teardown()

	nsObj := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
	test.SetupCliTest(strings.Join([]string{"project", ns}, " "), context.CommandFactory{BuildCommands: BuildCommands}, nsObj)
	o, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.NotEmpty(t, o)
	assert.Contains(t, o, "default")
}

func TestDisplayProjectCmd_WithJsonOutputFormat(t *testing.T) {
	ns := "default"
	teardown := test.OverrideKubeConfigAndCreateDefaultContext()
	defer teardown()

	nsObj := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
	test.SetupCliTest(strings.Join([]string{"project", "-o", "json"}, " "), context.CommandFactory{BuildCommands: BuildCommands}, nsObj)
	o, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.NotEmpty(t, o)
	assert.Contains(t, o, "\"level\":\"INFO\"")
	assert.Contains(t, o, "\"name\":\"kogito-cli\"")
	assert.Contains(t, o, "\"message\":\"Using project '"+ns+"'\"")
}
