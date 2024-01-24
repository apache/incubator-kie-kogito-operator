// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package converter

import (
	"github.com/apache/incubator-kie-kogito-operator/cmd/kogito/command/flag"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_FromArtifactFlagsToArtifact(t *testing.T) {
	artifactFlag := &flag.ArtifactFlags{
		ProjectGroupID:    "com.company",
		ProjectArtifactID: "project",
		ProjectVersion:    "1.0-SNAPSHOT",
	}

	artifact := FromArtifactFlagsToArtifact(artifactFlag)
	assert.NotNil(t, artifact)
	assert.Equal(t, "com.company", artifact.GroupID)
	assert.Equal(t, "project", artifact.ArtifactID)
	assert.Equal(t, "1.0-SNAPSHOT", artifact.Version)
}
