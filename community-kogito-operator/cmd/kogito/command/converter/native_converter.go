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
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/community-kogito-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/community-kogito-operator/cmd/kogito/command/flag"
	"io/ioutil"
	"strings"
)

// FromArgsToNative determines whether the build should be
// native based on arguments
func FromArgsToNative(nativeFlag bool, resourceType flag.ResourceType, resource string) (bool, error) {
	switch resourceType {
	case flag.LocalBinaryDirectoryResource:
		files, err := ioutil.ReadDir(resource)
		if err != nil {
			return nativeFlag, err
		}

		native := false
		for _, file := range files {
			if strings.HasSuffix(file.Name(), "-runner") {
				native = true
				break
			}
		}

		if nativeFlag && !native {
			return native, fmt.Errorf("specified native binary build but no native executable found in '%s'", resource)
		} else if !nativeFlag && native {
			context.GetDefaultLogger().Infof("Did not specify native build but found native executable. Setting binary build to native.")
			return native, nil
		}
		return native, nil
	default:
		return nativeFlag, nil
	}
}
