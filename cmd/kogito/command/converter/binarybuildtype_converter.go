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
	"github.com/kiegroup/kogito-operator/apis"
	"github.com/kiegroup/kogito-operator/cmd/kogito/command/flag"
)

// FromArgsToBinaryBuildType determines what kind of binary
// build the user is creating based on their passed arguments.
func FromArgsToBinaryBuildType(resourceType flag.ResourceType, runtime api.RuntimeType, native bool, legacy bool) flag.BinaryBuildType {
	if resourceType == flag.LocalBinaryDirectoryResource {
		if runtime == api.SpringBootRuntimeType {
			return flag.BinarySpringBootJvmBuild
		}
		if native {
			return flag.BinaryQuarkusNativeBuild
		}
		if legacy {
			return flag.BinaryQuarkusLegacyJarJvmBuild
		}
		return flag.BinaryQuarkusFastJarJvmBuild
	}
	return flag.SourceToImageBuild
}
