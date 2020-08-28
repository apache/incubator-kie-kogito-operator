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
	"github.com/kiegroup/kogito-cloud-operator/api/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_FromWebHookFlagsToWebHookSecret(t *testing.T) {
	flags := &flag.WebHookFlags{
		WebHook: []string{
			"GitHub=53537568546353",
		},
	}

	webHookSecrets := FromWebHookFlagsToWebHookSecret(flags)
	assert.NotNil(t, webHookSecrets)
	assert.Equal(t, 1, len(webHookSecrets))
	webHookSecret := webHookSecrets[0]
	assert.Equal(t, v1alpha1.GitHubWebHook, webHookSecret.Type)
	assert.Equal(t, "53537568546353", webHookSecret.Secret)
}
