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
	"strings"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
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

	// KogitoQuarkusUbi8Image Quarkus runtime builder image
	KogitoQuarkusUbi8Image = "kogito-quarkus-ubi8"
	// KogitoQuarkusJVMUbi8Image Quarkus jvm runtime builder image
	KogitoQuarkusJVMUbi8Image = "kogito-quarkus-jvm-ubi8"
	// KogitoQuarkusUbi8s2iImage Quarkus s2i builder image
	KogitoQuarkusUbi8s2iImage = "kogito-quarkus-ubi8-s2i"
	// KogitoSpringBootUbi8Image SpringBoot runtime builder image
	KogitoSpringBootUbi8Image = "kogito-springboot-ubi8"
	// KogitoSpringBootUbi8s2iImage SpringBoot s2i builder image
	KogitoSpringBootUbi8s2iImage = "kogito-springboot-ubi8-s2i"
)

var (
	// KogitoImages maps the default Kogito Images on a matrix of RuntimeType and its purpose
	KogitoImages = map[v1alpha1.RuntimeType]map[bool]string{
		v1alpha1.QuarkusRuntimeType: {
			true:  KogitoQuarkusUbi8s2iImage,
			false: KogitoQuarkusJVMUbi8Image,
		},
		v1alpha1.SpringBootRuntimeType: {
			true:  KogitoSpringBootUbi8s2iImage,
			false: KogitoSpringBootUbi8Image,
		},
	}
)

// GetKogitoImageVersion gets the Kogito Runtime latest micro version based on the Operator current version
// E.g. Operator version is 0.9.0, the latest image version is 0.9.x-latest
func GetKogitoImageVersion() string {
	return getKogitoImageVersion(version.Version)
}

// unit test friendly unexported function
// in this case we are considering only micro updates, that's 0.9.0 -> 0.9, thus for 1.0.0 => 1.0
// in the future this should be managed with carefully if we desire a behavior like 1.0.0 => 1, that's minor upgrades
func getKogitoImageVersion(v string) string {
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
		return strings.Join(versionPrefix[:lastIndex], versionSeparator)
	}
	return LatestTag
}
