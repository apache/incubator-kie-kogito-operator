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

package converter

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_FromStringArrayToEnvs(t *testing.T) {
	keyValuePair := []string{
		"key1=value1",
		"key2=value2",
	}

	envVars := FromStringArrayToEnvs(keyValuePair)
	assert.NotNil(t, envVars)
	assert.Equal(t, 2, len(envVars))
	for _, envVar := range envVars {
		if envVar.Name == "key1" {
			assert.Equal(t, "value1", envVar.Value)
		} else if envVar.Name == "key2" {
			assert.Equal(t, "value2", envVar.Value)
		} else {
			assert.Fail(t, "Invalid EnvVar")
		}
	}
}
