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
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mitchellh/go-homedir"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUseProjectCmd_WhenTheresNoConfigAndNoNamespace(t *testing.T) {
	home, _ := homedir.Dir()
	ns := uuid.New().String()
	path := filepath.Join(home, context.DefaultConfigPath, context.DefaultConfigFinalName)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		err := os.Remove(path)
		assert.NoError(t, err)
	} else {
		err := os.MkdirAll(filepath.Join(home, context.DefaultConfigPath), os.ModePerm)
		assert.NoError(t, err)
	}

	// open
	file, err := os.Create(path)
	defer file.Close()
	assert.NoError(t, err)
	test.SetupCliTest(strings.Join([]string{"use-project", ns}, " "), context.CommandFactory{BuildCommands: BuildCommands})
	_, _, err = test.ExecuteCli()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), ns)
}

func TestUseProjectCmd_WhenThereIsTheNamespace(t *testing.T) {
	ns := uuid.New().String()
	nsObj := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
	test.SetupCliTest(strings.Join([]string{"use-project", ns}, " "), context.CommandFactory{BuildCommands: BuildCommands}, nsObj)
	o, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.NotEmpty(t, o)
	assert.Contains(t, o, ns)
}

func TestUseProjectCmd_WhenWhatIsTheNamespace(t *testing.T) {
	config := context.ReadConfig()
	config.Namespace = ""
	config.Save()
	ns := "config-project"
	nsObj := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
	// set the project
	test.SetupCliTest(strings.Join([]string{"use-project", ns}, " "), context.CommandFactory{BuildCommands: BuildCommands}, nsObj)
	o1, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, o1, ns)

	test.SetupCliTest("use-project", context.CommandFactory{BuildCommands: BuildCommands}, nsObj)
	o2, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.NotEmpty(t, o2)
	assert.Contains(t, o2, ns)
}
