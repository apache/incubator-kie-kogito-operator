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

package message

const (
	// KubeConfigNoContext format in: expects kube config filepath
	KubeConfigNoContext = "There's no current context available in the kube config file %s. Please make sure to connect to the cluster first via 'oc/kubectl login' "
	// KubeConfigErrorWriteFile format in: filename, error
	KubeConfigErrorWriteFile = "Error while trying to update kube config file %s: %s "
)
