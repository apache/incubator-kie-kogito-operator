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

package util

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
)

// CheckImageTag checks the given image tag
func CheckImageTag(image string) error {
	if len(image) > 0 && !framework.DockerTagRegxCompiled.MatchString(image) {
		return fmt.Errorf("invalid name for image tag. Valid format is domain/namespace/image-name:tag. Received %s", image)
	}
	return nil
}
