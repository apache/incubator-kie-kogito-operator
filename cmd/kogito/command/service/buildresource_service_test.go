/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package service

import (
	"fmt"
	"github.com/kiegroup/kogito-operator/cmd/kogito/command/flag"
	"github.com/kiegroup/kogito-operator/cmd/kogito/command/test"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func Test_getRawGitHubFileURL(t *testing.T) {
	type args struct {
		resource string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"file in branch", args{"https://github.com/kiegroup/kogito-examples/blob/stable/licensesheader.txt"}, "https://raw.githubusercontent.com/kiegroup/kogito-examples/stable/licensesheader.txt"},
		{"file in commit", args{"https://github.com/kiegroup/kogito-examples/blob/8bde586ed5e536abec46b16b08f2d0b108391107/licensesheader.txt"}, "https://raw.githubusercontent.com/kiegroup/kogito-examples/8bde586ed5e536abec46b16b08f2d0b108391107/licensesheader.txt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRawGitHubFileURL(tt.args.resource); got != tt.want {
				t.Errorf("getRawGitHubFileURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_GetResourceType_BinaryResource(t *testing.T) {
	resourceType, err := GetResourceType("")
	assert.Nil(t, err)
	assert.Equal(t, flag.BinaryResource, resourceType)
}

func Test_GetResourceType_GitRepositoryResource(t *testing.T) {
	resourceType, err := GetResourceType("https://github.com/kiegroup/kogito-examples")
	assert.Nil(t, err)
	assert.Equal(t, flag.GitRepositoryResource, resourceType)
}

func Test_GetResourceType_GitFileResource(t *testing.T) {
	resourceType, err := GetResourceType("https://github.com/kiegroup/kogito-examples/blob/stable/process-scripts-quarkus/src/main/resources/org/acme/travels/scripts.bpmn")
	assert.Nil(t, err)
	assert.Equal(t, flag.GitFileResource, resourceType)
}

func Test_GetResourceType_LocalFileResource_Invalid(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "*.unsupported")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpFile.Name())
	_, err = GetResourceType(tmpFile.Name())
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Errorf("invalid resource %s", tmpFile.Name()), err)
	}
}

func Test_GetResourceType_LocalFileResource_Valid(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "*.bpmn")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpFile.Name())
	resourceType, err := GetResourceType(tmpFile.Name())
	assert.Nil(t, err)
	assert.Equal(t, flag.LocalFileResource, resourceType)
}

func Test_GetResourceType_LocalDirectoryResource(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpDir)
	resourceType, err := GetResourceType(tmpDir)
	assert.Nil(t, err)
	assert.Equal(t, flag.LocalDirectoryResource, resourceType)
}

func Test_GetResourceType_LocalBinaryDirectoryResource(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}
	targetDir := tmpDir + "/target/"
	test.Mkdir(targetDir)
	defer os.RemoveAll(tmpDir)
	resourceType, err := GetResourceType(targetDir)
	assert.Nil(t, err)
	assert.Equal(t, flag.LocalBinaryDirectoryResource, resourceType)
}
