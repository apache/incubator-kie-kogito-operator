// Copyright 2021 Red Hat, Inc. and/or its affiliates
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

package api

// GitSource Git coordinates to locate the source code to build.
// +k8s:openapi-gen=true
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Kogito Git Source"
type GitSource struct {
	// Git URI for the s2i source.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Git URI"
	URI string `json:"uri"`
	// Branch to use in the Git repository.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Git Reference"
	Reference string `json:"reference,omitempty"`
	// Context/subdirectory where the code is located, relative to the repo root.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Git Context"
	ContextDir string `json:"contextDir,omitempty"`
}
