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
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDisplayProjectCmd_WhenTheresNoConfigAndNoNamespace(t *testing.T) {
	path := test.GetTestConfigFilePath()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		err := os.Remove(path)
		assert.NoError(t, err)
	} else {
		err := os.MkdirAll(test.GetTestConfigPath(), os.ModePerm)
		assert.NoError(t, err)
	}

	// open
	file, err := os.Create(path)
	defer file.Close()
	assert.NoError(t, err)
	test.SetupCliTest(strings.Join([]string{"project"}, " "), context.CommandFactory{BuildCommands: BuildCommands})
	o, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, o, "No project configured")
}

func TestDisplayProjectCmd_WhenThereIsTheNamespace(t *testing.T) {
	config := context.ReadConfig()
	ns := uuid.New().String()
	config.Namespace = ns
	config.Save()

	nsObj := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
	test.SetupCliTest(strings.Join([]string{"project", ns}, " "), context.CommandFactory{BuildCommands: BuildCommands}, nsObj)
	o, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.NotEmpty(t, o)
	assert.Contains(t, o, ns)
}

func TestDisplayProjectCmd_WithJsonOutputFormat(t *testing.T) {
	config := context.ReadConfig()
	ns := uuid.New().String()
	config.Namespace = ns
	config.Save()

	nsObj := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
	test.SetupCliTest(strings.Join([]string{"project", ns, "-o", "json"}, " "), context.CommandFactory{BuildCommands: BuildCommands}, nsObj)
	o, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.NotEmpty(t, o)
	assert.Contains(t, o, "\"level\":\"INFO\"")
	assert.Contains(t, o, "\"name\":\"kogito-cli\"")
	assert.Contains(t, o, "\"message\":\"Using project '"+ns+"'\"")
}
