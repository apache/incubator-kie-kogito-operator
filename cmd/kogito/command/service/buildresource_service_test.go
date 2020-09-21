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

package service

import "testing"

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
