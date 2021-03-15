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

package api

import (
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

// ConditionMetaInterface defines the base information for kogito services conditions
type ConditionMetaInterface interface {
	GetConditions() []metav1.Condition
	AddCondition(condition metav1.Condition)
	RemoveCondition(index int)
}
