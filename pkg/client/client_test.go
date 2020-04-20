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

package client

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/test"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"testing"
)

func TestGetKubeConfigFile(t *testing.T) {
	// safe backup to not jeopardize user's envs
	tempEnvConfig, rollbackEnv := test.OverrideDefaultKubeConfigEmptyContext()
	defer rollbackEnv()
	assert.Equal(t, GetKubeConfigFile(), tempEnvConfig)

	// now we can try using the home dir since the env var will be empty
	os.Setenv(clientcmd.RecommendedConfigPathEnvVar, "")
	homeKubeConfig, err := os.UserHomeDir()
	assert.NoError(t, err)
	defaultKubeConfigPath := filepath.Join(clientcmd.RecommendedHomeDir, clientcmd.RecommendedFileName)
	homeKubeConfig = filepath.Join(homeKubeConfig, defaultKubeConfigPath)
	assert.Equal(t, GetKubeConfigFile(), homeKubeConfig)
}
