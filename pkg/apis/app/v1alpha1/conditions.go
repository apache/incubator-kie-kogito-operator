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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConditionType is the type of condition
type ConditionType string

const (
	// DeployedConditionType - The KogitoService is deployed
	DeployedConditionType ConditionType = "Deployed"
	// ProvisioningConditionType - The KogitoService is being provisioned
	ProvisioningConditionType ConditionType = "Provisioning"
	// FailedConditionType - The KogitoService is in a failed state
	FailedConditionType ConditionType = "Failed"
)

// KogitoServiceConditionReason is the type of reason
type KogitoServiceConditionReason string

const (
	// CreateResourceFailedReason - Unable to create the requested resources
	CreateResourceFailedReason KogitoServiceConditionReason = "CreateResourceFailed"
	// KogitoInfraNotReadyReason - Unable to deploy Kogito Infra
	KogitoInfraNotReadyReason KogitoServiceConditionReason = "KogitoInfraNotReadyReason"
	// ServiceReconciliationFailure - Unable to determine the error
	ServiceReconciliationFailure KogitoServiceConditionReason = "ReconciliationFailure"
	// MessagingIntegrationFailureReason ...
	MessagingIntegrationFailureReason KogitoServiceConditionReason = "MessagingProvisionFailure"
	// MonitoringIntegrationFailureReason ...
	MonitoringIntegrationFailureReason KogitoServiceConditionReason = "MonitoringIntegrationFailure"
	// InternalServiceNotReachable ...
	InternalServiceNotReachable KogitoServiceConditionReason = "InternalServiceNotReachable"
)

// Condition is the detailed condition for the resource
// +k8s:openapi-gen=true
type Condition struct {
	Type               ConditionType                `json:"type"`
	Status             corev1.ConditionStatus       `json:"status"`
	LastTransitionTime metav1.Time                  `json:"lastTransitionTime,omitempty"`
	Reason             KogitoServiceConditionReason `json:"reason,omitempty"`
	Message            string                       `json:"message,omitempty"`
}

const maxBufferCondition = 5

// ConditionMetaInterface defines the base information for kogito services conditions
type ConditionMetaInterface interface {
	SetDeployed() bool
	SetProvisioning() bool
	SetFailed(reason KogitoServiceConditionReason, err error)
	GetConditions() []Condition
	SetConditions(conditions []Condition)
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
func (c *ConditionsMeta) GetConditions() []Condition {
	return c.Conditions
}

// SetConditions sets the conditions history
func (c *ConditionsMeta) SetConditions(conditions []Condition) {
	c.Conditions = conditions
}

// SetDeployed Updates the condition with the DeployedCondition and True status
func (c *ConditionsMeta) SetDeployed() bool {
	size := len(c.Conditions)
	if size > 0 && c.Conditions[size-1].Type == DeployedConditionType {
		return false
	}
	condition := Condition{
		Type:               DeployedConditionType,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
	}
	c.Conditions = c.addCondition(condition)
	return true
}

// SetProvisioning Sets the condition type to Provisioning and status True if not yet set.
func (c *ConditionsMeta) SetProvisioning() bool {
	size := len(c.Conditions)
	if size > 0 && c.Conditions[size-1].Type == ProvisioningConditionType {
		return false
	}
	condition := Condition{
		Type:               ProvisioningConditionType,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
	}
	c.Conditions = c.addCondition(condition)
	return true
}

// SetFailed Sets the failed condition with the error reason and message
func (c *ConditionsMeta) SetFailed(reason KogitoServiceConditionReason, err error) {
	condition := Condition{
		Type:               FailedConditionType,
		Status:             corev1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            err.Error(),
	}
	c.Conditions = c.addCondition(condition)
}

// addCondition adds a condition to the condition array ensuring the max buffer
func (c *ConditionsMeta) addCondition(condition Condition) []Condition {
	size := len(c.Conditions) + 1
	first := 0
	if size > maxBufferCondition {
		first = size - maxBufferCondition
	}
	return append(c.Conditions, condition)[first:size]
}
