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
	"github.com/kiegroup/kogito-operator/cmd/kogito/command/flag"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_FromConfigFlagsToMap(t *testing.T) {
	configFlag := &flag.ConfigFlags{
		Config: []string{
			"key1=value1",
			"key2=value2",
		},
	}

	configMap := FromConfigFlagsToMap(configFlag)
	assert.NotNil(t, configMap)
	assert.Equal(t, "value1", configMap["key1"])
	assert.Equal(t, "value2", configMap["key2"])
}
