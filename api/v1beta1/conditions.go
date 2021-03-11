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
	"github.com/kiegroup/kogito-cloud-operator/api"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Condition is the detailed condition for the resource
type Condition struct {
	Type               api.ConditionType                `json:"type"`
	Status             corev1.ConditionStatus           `json:"status"`
	LastTransitionTime metav1.Time                      `json:"lastTransitionTime,omitempty"`
	Reason             api.KogitoServiceConditionReason `json:"reason,omitempty"`
	Message            string                           `json:"message,omitempty"`
}

// GetType ...
func (c Condition) GetType() api.ConditionType {
	return c.Type
}

// GetStatus ...
func (c Condition) GetStatus() corev1.ConditionStatus {
	return c.Status
}

// GetLastTransitionTime ...
func (c Condition) GetLastTransitionTime() metav1.Time {
	return c.LastTransitionTime
}

// GetReason ...
func (c Condition) GetReason() api.KogitoServiceConditionReason {
	return c.Reason
}

// GetMessage ...
func (c Condition) GetMessage() string {
	return c.Message
}

// ConditionsMeta definition of a Condition structure
type ConditionsMeta struct {
	// +listType=atomic
	// History of conditions for the resource
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.x-descriptors="urn:alm:descriptor:io.kubernetes.conditions"
	Conditions []Condition `json:"conditions"`
}

// GetConditions returns the conditions history
func (c *ConditionsMeta) GetConditions() []api.ConditionInterface {
	conditions := make([]api.ConditionInterface, len(c.Conditions))
	for i, v := range c.Conditions {
		conditions[i] = api.ConditionInterface(v)
	}
	return conditions
}

// SetConditions sets the conditions history
func (c *ConditionsMeta) SetConditions(conditions []api.ConditionInterface) {
	var newConditions []Condition
	for _, condition := range conditions {
		if newCondition, ok := condition.(Condition); ok {
			newConditions = append(newConditions, newCondition)
		}

	}
	c.Conditions = newConditions
}

// NewDeployedCondition ...
func (c *ConditionsMeta) NewDeployedCondition() api.ConditionInterface {
	return Condition{
		Type:               api.DeployedConditionType,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
	}
}

// NewProvisioningCondition ...
func (c *ConditionsMeta) NewProvisioningCondition() api.ConditionInterface {
	return Condition{
		Type:               api.ProvisioningConditionType,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
	}
}

// NewFailedCondition ...
func (c *ConditionsMeta) NewFailedCondition(reason api.KogitoServiceConditionReason, err error) api.ConditionInterface {
	return Condition{
		Type:               api.FailedConditionType,
		Status:             corev1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            err.Error(),
	}
}
