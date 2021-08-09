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

// ConfigMapReference ...
type ConfigMapReference struct {
	// This must match the Name of a ConfigMap.
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Type of ConfigMap
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
func (c *ConfigMapReference) GetName() string {
	return c.Name
}

// SetName ...
func (c *ConfigMapReference) SetName(name string) {
	c.Name = name
}

// GetMountType ...
func (c *ConfigMapReference) GetMountType() api.MountType {
	return c.MountType
}

// SetMountType ...
func (c *ConfigMapReference) SetMountType(mountType api.MountType) {
	c.MountType = mountType
}

// GetMountPath ...
func (c *ConfigMapReference) GetMountPath() string {
	return c.MountPath
}

// SetMountPath ...
func (c *ConfigMapReference) SetMountPath(mountPath string) {
	c.MountPath = mountPath
}

// IsOptional ...
func (c *ConfigMapReference) IsOptional() *bool {
	return c.Optional
}

// SetOptional ....
func (c *ConfigMapReference) SetOptional(optional *bool) {
	c.Optional = optional
}

// GetFileMode ...
func (c *ConfigMapReference) GetFileMode() *int32 {
	return c.FileMode
}

// SetFileMode ...
func (c *ConfigMapReference) SetFileMode(fileMode *int32) {
	c.FileMode = fileMode
}
