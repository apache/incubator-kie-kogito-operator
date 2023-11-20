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

package util

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func Test_ValidateKeyPair_CorrectInput(t *testing.T) {
	keyValuePair := []string{
		"VAR1=key1",
		"VAR2=key2",
	}

	err := CheckKeyPair(keyValuePair)
	assert.NoError(t, err)
}

func Test_ValidateKeyPair_InCorrectInput(t *testing.T) {
	keyValuePair := []string{
		"VAR1=key1",
		"VAR2",
	}

	err := CheckKeyPair(keyValuePair)
	assert.Error(t, err)
}

func Test_ValidateSecretEnvVar_CorrectInput(t *testing.T) {
	keyValuePair := []string{
		"VAR1=secretName1#secretKey1",
		"VAR2=secretName2#secretKey2",
	}

	err := CheckSecretKeyPair(keyValuePair)
	assert.NoError(t, err)
}

func Test_ValidateSecretEnvVar_InCorrectInput(t *testing.T) {
	keyValuePair := []string{
		"VAR1=secretName1#secretKey1",
		"VAR2=secretName2@secretKey2",
	}

	err := CheckSecretKeyPair(keyValuePair)
	assert.Error(t, err)
}

func Test_ValidateImageTag_CorrectInput(t *testing.T) {
	image := "quay.io/kiegroup/data_index:1.0"
	err := CheckImageTag(image)
	assert.NoError(t, err)
}

func TestCheckFileExists(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "application.properties")
	assert.NoError(t, err)
	tempDir, err := ioutil.TempDir("", "checkdir")
	assert.NoError(t, err)
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			"File exists", args{path: tempFile.Name()}, true, false,
		},
		{
			"It's a directory", args{path: tempDir}, false, false,
		},
		{
			"File does not exists", args{"/an/imaginary/path"}, false, false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckFileExists(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckFileExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CheckFileExists() got = %v, want %v", got, tt.want)
			}
		})
	}
}
