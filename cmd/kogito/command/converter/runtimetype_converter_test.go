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
	"github.com/kiegroup/kogito-cloud-operator/api"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/test"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_FromRuntimeFlagsToRuntimeType_SpringBoot(t *testing.T) {
	flags := &flag.RuntimeTypeFlags{
		Runtime: "springboot",
	}

	runtimeType := FromRuntimeFlagsToRuntimeType(flags)
	assert.Equal(t, api.SpringBootRuntimeType, runtimeType)
}

func Test_FromRuntimeFlagsToRuntimeType_Quarkus(t *testing.T) {
	flags := &flag.RuntimeTypeFlags{
		Runtime: "quarkus",
	}

	runtimeType := FromRuntimeFlagsToRuntimeType(flags)
	assert.Equal(t, api.QuarkusRuntimeType, runtimeType)
}

func Test_FromArgsToRuntime_NonBinaryBuild(t *testing.T) {
	flags := &flag.RuntimeTypeFlags{
		Runtime: "quarkus",
	}

	runtimeType, err := FromArgsToRuntimeType(flags, flag.GitRepositoryResource, "https://github.com/kiegroup/kogito-examples/blob/stable/process-scripts-quarkus/src/main/resources/org/acme/travels/scripts.bpmn")
	assert.Nil(t, err)
	assert.Equal(t, api.QuarkusRuntimeType, runtimeType)
}

func Test_FromArgsToRuntime_BinaryBuild_SpringBoot(t *testing.T) {
	flags := &flag.RuntimeTypeFlags{
		Runtime: "quarkus",
	}

	tmpDir := test.TempDirWithFile("target", "*-runner-*.jar")
	defer os.RemoveAll(tmpDir)

	runtimeType, err := FromArgsToRuntimeType(flags, flag.LocalBinaryDirectoryResource, tmpDir)
	assert.Nil(t, err)
	assert.Equal(t, api.SpringBootRuntimeType, runtimeType)
}

func Test_FromArgsToRuntime_BinaryBuild_Quarkus(t *testing.T) {
	flags := &flag.RuntimeTypeFlags{
		Runtime: "quarkus",
	}

	tmpDir := test.TempDirWithFile("target", "*-runner.jar")
	defer os.RemoveAll(tmpDir)

	runtimeType, err := FromArgsToRuntimeType(flags, flag.LocalBinaryDirectoryResource, tmpDir)
	assert.Nil(t, err)
	assert.Equal(t, api.QuarkusRuntimeType, runtimeType)
}

func Test_FromArgsToRuntime_BinaryBuild_QuarkusNative(t *testing.T) {
	flags := &flag.RuntimeTypeFlags{
		Runtime: "quarkus",
	}

	tmpDir := test.TempDirWithFile("target", "*-runner")
	defer os.RemoveAll(tmpDir)

	runtimeType, err := FromArgsToRuntimeType(flags, flag.LocalBinaryDirectoryResource, tmpDir)
	assert.Nil(t, err)
	assert.Equal(t, api.QuarkusRuntimeType, runtimeType)
}
