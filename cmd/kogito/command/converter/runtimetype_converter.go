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
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
	"io/ioutil"
	"strings"
)

// FromRuntimeFlagsToRuntimeType converts given RuntimeTypeFlags into RuntimeType
func FromRuntimeFlagsToRuntimeType(flags *flag.RuntimeTypeFlags) v1beta1.RuntimeType {
	return v1beta1.RuntimeType(flags.Runtime)
}

// FromArgsToRuntimeType determines what the runtime is based on
// arguments
func FromArgsToRuntimeType(flags *flag.RuntimeTypeFlags, resourceType flag.ResourceType, resource string) (v1beta1.RuntimeType, error) {
	runtimeType := FromRuntimeFlagsToRuntimeType(flags)

	// if given local binary directory, can determine what
	// runtime type is needed based on presence of runner file
	if resourceType == flag.LocalBinaryDirectoryResource {
		files, err := ioutil.ReadDir(resource)
		if err != nil {
			return runtimeType, err
		}

		for _, file := range files {
			if strings.HasSuffix(file.Name(), "-runner") || strings.HasSuffix(file.Name(), "-runner.jar") {
				return v1beta1.QuarkusRuntimeType, nil
			}
		}
		return v1beta1.SpringBootRuntimeType, nil
	}

	return runtimeType, nil
}
