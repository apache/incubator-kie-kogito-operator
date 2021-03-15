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

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConditionsMeta definition of a Condition structure
type ConditionsMeta struct {
	// +listType=atomic
	// History of conditions for the resource
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.x-descriptors="urn:alm:descriptor:io.kubernetes.conditions"
	Conditions []metav1.Condition `json:"conditions"`
}

// GetConditions returns the conditions history
func (c *ConditionsMeta) GetConditions() []metav1.Condition {
	return c.Conditions
}

// AddCondition ...
func (c *ConditionsMeta) AddCondition(condition metav1.Condition) {
	c.Conditions = append(c.Conditions, condition)
}

// RemoveCondition ...
func (c *ConditionsMeta) RemoveCondition(index int) {
	c.Conditions = append(c.Conditions[:index], c.Conditions[index+1:]...)
}
