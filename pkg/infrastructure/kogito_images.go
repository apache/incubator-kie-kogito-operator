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

import (
	"fmt"
	"strings"

	"github.com/kiegroup/kogito-cloud-operator/version"
)

const (
	versionSeparator = "."
	// LatestTag the default name for latest image tag
	LatestTag = "latest"
	// DefaultImageRegistry the default services image repository
	DefaultImageRegistry = "quay.io"
	// DefaultImageNamespace the default services image namespace
	DefaultImageNamespace = "kiegroup"
)

// GetRuntimeImageVersion gets the Kogito Runtime latest micro version based on the Operator current version
// E.g. Operator version is 0.9.0, the latest image version is 0.9.x-latest
func GetRuntimeImageVersion() string {
	return getRuntimeImageVersion(version.Version)
}

// unit test friendly unexported function
// in this case we are considering only micro updates, that's 0.9.0 -> 0.9, thus for 1.0.0 => 1.0
// in the future this should be managed with carefully if we desire a behavior like 1.0.0 => 1, that's minor upgrades
func getRuntimeImageVersion(v string) string {
	if len(v) == 0 {
		return LatestTag
	}

	versionPrefix := strings.Split(v, versionSeparator)
	length := len(versionPrefix)
	if length > 0 {
		lastIndex := 2   // micro updates
		if length <= 2 { // guard against unusual cases
			lastIndex = length
		}
		return fmt.Sprintf("%s", strings.Join(versionPrefix[:lastIndex], versionSeparator))
	}
	return LatestTag
}
