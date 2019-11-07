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

package client

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func Test_getKubeConfigFile(t *testing.T) {
	// safe backup to not jeopardize user's envs
	oldEnvVar := util.GetOSEnv(envVarKubeConfig, "")
	tempEnvConfig := "/tmp/config"
	defer func() {
		os.Setenv(envVarKubeConfig, oldEnvVar)
	}()
	os.Setenv(envVarKubeConfig, tempEnvConfig)
	assert.Equal(t, getKubeConfigFile(), tempEnvConfig)

	// now we can try using the home dir since the env var will be empty
	os.Setenv(envVarKubeConfig, "")
	homeKubeConfig, err := os.UserHomeDir()
	assert.NoError(t, err)
	homeKubeConfig = filepath.Join(homeKubeConfig, defaultKubeConfigPath)
	assert.Equal(t, getKubeConfigFile(), homeKubeConfig)
}
