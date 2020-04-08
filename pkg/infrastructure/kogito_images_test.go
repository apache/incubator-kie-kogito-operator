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

package infrastructure

import "testing"

func Test_getRuntimeImageVersion(t *testing.T) {
	type args struct {
		v string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Usual case", args{v: "0.9.0"}, "0.9"},
		{"Micro update case", args{v: "0.9.1"}, "0.9"},
		{"Micro micro update case", args{v: "0.9.1.1"}, "0.9"},
		{"Unusual version", args{v: "0.9"}, "0.9"},
		{"Weird version", args{v: "0"}, "0"},
		{"No version", args{v: ""}, LatestTag},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRuntimeImageVersion(tt.args.v); got != tt.want {
				t.Errorf("getRuntimeImageVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
