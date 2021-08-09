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

package v1beta1

import "github.com/kiegroup/kogito-operator/api"

// SecretReference ...
type SecretReference struct {
	// This must match the Name of a Secret.
	Name string `json:"name protobuf:"bytes,1,opt,name=name"`
	// Type of Secret
	// +optional
	MountType api.MountType `json:"mountType,omitempty" protobuf:"bytes,2,opt,name=mountPath"`
	// Path within the container at which the volume should be mounted.  Must
	// not contain ':'.
	// +optional
	MountPath string `json:"mountPath,omitempty" protobuf:"bytes,3,opt,name=mountPath"`
	// Optional: mode bits used to set permissions on created files by default.
	// Must be an octal value between 0000 and 0777 or a decimal value between 0 and 511.
	// YAML accepts both octal and decimal values, JSON requires decimal values
	// for mode bits. Defaults to 0644.
	// Directories within the path are not affected by this setting.
	// This might be in conflict with other options that affect the file
	// mode, like fsGroup, and the result can be other mode bits set.
	// +optional
	FileMode *int32 `json:"fileMode,omitempty" protobuf:"bytes,4,opt,name=fileMode"`
	// Specify whether the Secret or its keys must be defined
	// +optional
	Optional *bool `json:"optional,omitempty" protobuf:"varint,5,opt,name=optional"`
}

// GetName ...
func (s *SecretReference) GetName() string {
	return s.Name
}

// SetName ...
func (s *SecretReference) SetName(name string) {
	s.Name = name
}

// GetMountType ...
func (s *SecretReference) GetMountType() api.MountType {
	return s.MountType
}

// SetMountType ...
func (s *SecretReference) SetMountType(mountType api.MountType) {
	s.MountType = mountType
}

// GetMountPath ...
func (s *SecretReference) GetMountPath() string {
	return s.MountPath
}

// SetMountPath ...
func (s *SecretReference) SetMountPath(mountPath string) {
	s.MountPath = mountPath
}

// IsOptional ...
func (s *SecretReference) IsOptional() *bool {
	return s.Optional
}

// SetOptional ....
func (s *SecretReference) SetOptional(optional *bool) {
	s.Optional = optional
}

// GetFileMode ...
func (s *SecretReference) GetFileMode() *int32 {
	return s.FileMode
}

// SetFileMode ...
func (s *SecretReference) SetFileMode(fileMode *int32) {
	s.FileMode = fileMode
}
