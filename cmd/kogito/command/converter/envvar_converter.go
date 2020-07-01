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

package converter

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
	v1 "k8s.io/api/core/v1"
)

// FromStringArrayToEnvs converts a string array in the format of key=value pairs to the required type for the Kubernetes EnvVar type
func FromStringArrayToEnvs(keyValuePair []string) []v1.EnvVar {
	if keyValuePair == nil {
		return nil
	}
	return framework.MapToEnvVar(util.FromStringsKeyPairToMap(keyValuePair))
}
