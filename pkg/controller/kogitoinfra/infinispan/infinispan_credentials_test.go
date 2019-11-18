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

package infinispan

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

func Test_generateDefaultCredentials(t *testing.T) {
	yamlFile, fileMD5, err := generateDefaultCredentials()
	assert.NoError(t, err)
	assert.NotEmpty(t, yamlFile)
	assert.NotEmpty(t, fileMD5)

	actualMD5 := getMD5FromBytes([]byte(yamlFile))
	assert.NoError(t, err)
	assert.Equal(t, actualMD5, fileMD5)
}

func Test_hasKogitoUser(t *testing.T) {
	defaultFile, _, _ := generateDefaultCredentials()
	developersCredential, _ := yaml.Marshal(&Identity{Credentials: []Credential{{Username: "johndoe", Password: generateRandomPassword()}}})
	type args struct {
		secretFileData []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"default file has kogito user", args{secretFileData: []byte(defaultFile)}, true},
		{"developer file hasn't kogito " +
			"user", args{secretFileData: developersCredential}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasKogitoUser(tt.args.secretFileData); got != tt.want {
				t.Errorf("hasKogitoUser() = %v, want %v", got, tt.want)
			}
		})
	}
}
