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

package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_FromKafkaFlagsToKafkaProperties_EnableEventsWithUserDefineProperties(t *testing.T) {
	kafkaFlags := &KafkaFlags{
		ExternalURI: "ExternalURI",
		Instance:    "Instance",
	}
	kafkaProperties := FromKafkaFlagsToKafkaProperties(kafkaFlags, true)
	assert.NotNil(t, kafkaProperties)
	assert.Equal(t, "ExternalURI", kafkaProperties.ExternalURI)
	assert.Equal(t, "Instance", kafkaProperties.Instance)
	assert.False(t, kafkaProperties.UseKogitoInfra)
}

func Test_FromKafkaFlagsToKafkaProperties_EnableEventsWithDefaultProperties(t *testing.T) {
	kafkaFlags := &KafkaFlags{}
	kafkaProperties := FromKafkaFlagsToKafkaProperties(kafkaFlags, true)
	assert.NotNil(t, kafkaProperties)
	assert.True(t, kafkaProperties.UseKogitoInfra)
}

func Test_FromKafkaFlagsToKafkaProperties_DisableEvents(t *testing.T) {
	kafkaFlags := &KafkaFlags{}
	kafkaProperties := FromKafkaFlagsToKafkaProperties(kafkaFlags, false)
	assert.NotNil(t, kafkaProperties)
	assert.False(t, kafkaProperties.UseKogitoInfra)
}
