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
	keyPairStrings := []string{
		"VAR1=key1",
	}

	secretKeyPairStrings := []string{
		"VAR2=secretName2#secretKey2",
	}

	envVars := FromStringArrayToEnvs(keyPairStrings, secretKeyPairStrings)
	assert.NotNil(t, envVars)
	assert.Equal(t, 2, len(envVars))
	generalEnv := envVars[0]
	secretEnvVar := envVars[1]
	assert.Equal(t, "VAR1", generalEnv.Name)
	assert.Equal(t, "key1", generalEnv.Value)
	assert.Equal(t, "VAR2", secretEnvVar.Name)
	assert.Equal(t, "secretKey2", secretEnvVar.ValueFrom.SecretKeyRef.Key)
	assert.Equal(t, "secretName2", secretEnvVar.ValueFrom.SecretKeyRef.LocalObjectReference.Name)
}
