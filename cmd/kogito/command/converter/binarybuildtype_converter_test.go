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
	"reflect"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/api"

	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
)

func TestFromArgsToBinaryBuildType(t *testing.T) {
	type args struct {
		resourceType flag.ResourceType
		runtime      api.RuntimeType
		native       bool
		legacy       bool
	}
	tests := []struct {
		name string
		args args
		want flag.BinaryBuildType
	}{
		{"Spring Boot JVM binary build", args{flag.LocalBinaryDirectoryResource, api.SpringBootRuntimeType, false, false}, flag.BinarySpringBootJvmBuild},
		{"Quarkus native binary build", args{flag.LocalBinaryDirectoryResource, api.QuarkusRuntimeType, true, false}, flag.BinaryQuarkusNativeBuild},
		{"Quarkus Fast JVM binary build", args{flag.LocalBinaryDirectoryResource, api.QuarkusRuntimeType, false, false}, flag.BinaryQuarkusFastJarJvmBuild},
		{"Quarkus Legacy JVM binary build", args{flag.LocalBinaryDirectoryResource, api.QuarkusRuntimeType, false, true}, flag.BinaryQuarkusLegacyJarJvmBuild},
		{"s2i build", args{flag.GitFileResource, api.QuarkusRuntimeType, true, false}, flag.SourceToImageBuild},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromArgsToBinaryBuildType(tt.args.resourceType, tt.args.runtime, tt.args.native, tt.args.legacy); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromArgsToBinaryBuildType() = %v, want %v", got, tt.want)
			}
		})
	}
}
