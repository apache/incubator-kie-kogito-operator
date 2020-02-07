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

package framework

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"reflect"
	"testing"
)

func TestFromStringToImage(t *testing.T) {
	type args struct {
		imageTag string
	}
	tests := []struct {
		name string
		args args
		want v1alpha1.Image
	}{
		{"empty", args{""}, v1alpha1.Image{}},
		{"with registry name", args{"quay.io/openshift/myimage:1.0"}, v1alpha1.Image{Name: "myimage", Tag: "1.0", Namespace: "openshift", Domain: "quay.io"}},
		{"full name", args{"openshift/myimage:1.0"}, v1alpha1.Image{Name: "myimage", Tag: "1.0", Namespace: "openshift"}},
		{"namespace empty", args{"myimage:1.0"}, v1alpha1.Image{Name: "myimage", Tag: "1.0", Namespace: "", Domain: ""}},
		{"tag empty", args{"myimage"}, v1alpha1.Image{Name: "myimage", Tag: "latest", Namespace: "", Domain: ""}},
		{"tag empty with a trick", args{"myimage:"}, v1alpha1.Image{Name: "myimage", Tag: "latest", Namespace: "", Domain: ""}},
		{"just tag", args{":1.0"}, v1alpha1.Image{Name: "", Tag: "1.0", Namespace: "", Domain: ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertImageTagToImage(tt.args.imageTag); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertImageTagToImage() = %v, want %v", got, tt.want)
			}
		})
	}
}
