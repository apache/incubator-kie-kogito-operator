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

package converter

import (
	"io/ioutil"
	"strings"

	"github.com/apache/incubator-kie-kogito-operator/apis"
	"github.com/apache/incubator-kie-kogito-operator/cmd/kogito/command/flag"
)

var (
	quarkusLegacyJarSuffix       = "-runner.jar"
	quarkusRuntimeTargetSuffixes = []string{quarkusLegacyJarSuffix, "quarkus-app", "-runner"}
)

// FromRuntimeFlagsToRuntimeType converts given RuntimeTypeFlags into RuntimeType
func FromRuntimeFlagsToRuntimeType(flags *flag.RuntimeTypeFlags) api.RuntimeType {
	return api.RuntimeType(flags.Runtime)
}

// FromArgsToRuntimeType determines what the runtime is based on
// arguments
func FromArgsToRuntimeType(flags *flag.RuntimeTypeFlags, resourceType flag.ResourceType, resource string) (api.RuntimeType, error) {
	runtimeType := FromRuntimeFlagsToRuntimeType(flags)

	// if given local binary directory, can determine what
	// runtime type is needed based on presence of runner file
	if resourceType == flag.LocalBinaryDirectoryResource {
		files, err := ioutil.ReadDir(resource)
		if err != nil {
			return runtimeType, err
		}

		for _, file := range files {
			for _, fileSuffix := range quarkusRuntimeTargetSuffixes {
				if strings.HasSuffix(file.Name(), fileSuffix) {
					return api.QuarkusRuntimeType, nil
				}
			}
		}
		return api.SpringBootRuntimeType, nil
	}

	return runtimeType, nil
}

// ToQuarkusLegacyJarType determines what the type of quarkus jar is based on resources
func ToQuarkusLegacyJarType(resourceType flag.ResourceType, resource string) (bool, error) {
	switch resourceType {
	case flag.LocalBinaryDirectoryResource:
		files, err := ioutil.ReadDir(resource)
		if err != nil {
			return false, err
		}

		for _, file := range files {
			if strings.HasSuffix(file.Name(), quarkusLegacyJarSuffix) {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, nil
	}
}
