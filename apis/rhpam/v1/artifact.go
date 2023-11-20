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

package v1

// Artifact contains override information for building the Maven artifact.
// + optional
// +operator-sdk:csv:customresourcedefinitions:displayName="Final Artifact"
type Artifact struct {

	//Indicates the unique identifier of the organization or group that created the project.
	// + optional
	GroupID string `json:"groupId,omitempty"`

	//Indicates the unique base name of the primary artifact being generated.
	// + optional
	ArtifactID string `json:"artifactId,omitempty"`

	//Indicates the version of the artifact generated by the project.
	// + optional
	Version string `json:"version,omitempty"`
}

// GetGroupID ...
func (a *Artifact) GetGroupID() string {
	return a.GroupID
}

// SetGroupID ...
func (a *Artifact) SetGroupID(groupID string) {
	a.GroupID = groupID
}

// GetArtifactID ...
func (a *Artifact) GetArtifactID() string {
	return a.ArtifactID
}

// SetArtifactID ...
func (a *Artifact) SetArtifactID(artifactID string) {
	a.ArtifactID = artifactID
}

// GetVersion ...
func (a *Artifact) GetVersion() string {
	return a.Version
}

// SetVersion ...
func (a *Artifact) SetVersion(version string) {
	a.Version = version
}
