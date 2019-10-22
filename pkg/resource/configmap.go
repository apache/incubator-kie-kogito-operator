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

package resource

import (
	v1 "k8s.io/api/core/v1"
)

// AddFilesToConfigMap add files in the map format, where the key is the file name and the value it's contents, to a configMap
func AddFilesToConfigMap(files map[string]string, cm *v1.ConfigMap) {
	if cm.Data == nil {
		cm.Data = map[string]string{}
	}

	for key, value := range files {
		cm.Data[key] = value
	}
}
