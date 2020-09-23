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
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/test"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_FromArgsToNative_NotBinaryBuild(t *testing.T) {
	native, err := FromArgsToNative(true, flag.LocalDirectoryResource, "/tmp")
	assert.True(t, native)
	assert.Nil(t, err)
}

func Test_FromArgsToNative_BinaryBuild_NotNative(t *testing.T) {
	tmpDir := test.TempDirWithFile("target", "*-runner.jar")
	defer os.RemoveAll(tmpDir)

	// test correct use case of no native flag, returns false
	native, err := FromArgsToNative(false, flag.LocalBinaryDirectoryResource, tmpDir)
	assert.False(t, native)
	assert.Nil(t, err)

	// test incorrect use case of native flag, returns error
	_, err = FromArgsToNative(true, flag.LocalBinaryDirectoryResource, tmpDir)
	assert.Error(t, err)
}

func Test_FromArgsToNative_BinaryBuild_Native(t *testing.T) {
	tmpDir := test.TempDirWithFile("target", "*-runner")
	defer os.RemoveAll(tmpDir)

	// test correct use case of native flag, returns true
	native, err := FromArgsToNative(true, flag.LocalBinaryDirectoryResource, tmpDir)
	assert.True(t, native)
	assert.Nil(t, err)

	// test use case of no native flag but native directory, returns true
	native, err = FromArgsToNative(false, flag.LocalBinaryDirectoryResource, tmpDir)
	assert.True(t, native)
	assert.Nil(t, err)
}
