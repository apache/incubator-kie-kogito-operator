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

package util

import (
	"reflect"
	"testing"
)

func TestFromStringsKeyPairToMap(t *testing.T) {
	type args struct {
		array []string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{"happy path", args{[]string{"env1=value", "env2=value"}}, map[string]string{"env1": "value", "env2": "value"}},
		{"happy path2", args{[]string{"env1=value"}}, map[string]string{"env1": "value"}},
		{"only key", args{[]string{"env1="}}, map[string]string{"env1": ""}},
		{"only key without sep", args{[]string{"env1"}}, map[string]string{"env1": ""}},
		{"no key no value", args{[]string{""}}, map[string]string{}},
		{"no key with value", args{[]string{"=value"}}, map[string]string{}},
		{"various seps", args{[]string{"env1=value=value"}}, map[string]string{"env1": "value=value"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromStringsKeyPairToMap(tt.args.array); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromStringsKeyPairToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
