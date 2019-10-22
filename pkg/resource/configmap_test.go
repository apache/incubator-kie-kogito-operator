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

package resource

import (
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"testing"
)

func TestAddFilesToConfigMap(t *testing.T) {
	files := map[string]string{"file1.proto": "this is the file content"}
	emptyCm := &v1.ConfigMap{}
	notEmptyCm := &v1.ConfigMap{Data: files}
	type args struct {
		files map[string]string
		cm    *v1.ConfigMap
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "config map has no values", args: args{
			files: files,
			cm:    emptyCm,
		}},
		{name: "config map has values", args: args{
			files: files,
			cm:    notEmptyCm,
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AddFilesToConfigMap(tt.args.files, tt.args.cm)
		})
	}

	assert.Equal(t, len(emptyCm.Data), 1)
	assert.Equal(t, len(notEmptyCm.Data), 1)

	assert.Contains(t, emptyCm.Data, "file1.proto")
	assert.Contains(t, notEmptyCm.Data, "file1.proto")
}
