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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KogitoInfraSpec defines the desired state of KogitoInfra
// +k8s:openapi-gen=true
type KogitoInfraSpec struct {
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// InstallInfinispan indicates if Infinispan should be installed or not using Infinispan Operator.
	// Please note that the Infinispan Operator must be installed manually on environments that doesn't have OLM installed.
	InstallInfinispan bool `json:"installInfinispan"`
}

// KogitoInfraStatus defines the observed state of KogitoInfra
// +k8s:openapi-gen=true
type KogitoInfraStatus struct {
	Condition  KogitoInfraCondition    `json:"condition,omitempty"`
	Infinispan InfinispanInstallStatus `json:"infinispan,omitempty"`
}

// KogitoInfraCondition ...
type KogitoInfraCondition struct {
	Type               KogitoInfraConditionType `json:"type"`
	Status             v1.ConditionStatus       `json:"status"`
	LastTransitionTime metav1.Time              `json:"lastTransitionTime,omitempty"`
	Message            string                   `json:"message,omitempty"`
}

// InfinispanInstallStatus defines the Infinispan installation status
type InfinispanInstallStatus struct {
	Service          string             `json:"service,omitempty"`
	CredentialSecret string             `json:"credentialSecret,omitempty"`
	Name             string             `json:"name,omitempty"`
	Condition        []InstallCondition `json:"condition,omitempty"`
}

// InstallCondition defines the installation condition for the infrastructure actor
type InstallCondition struct {
	Type               InstallConditionType `json:"type"`
	Status             v1.ConditionStatus   `json:"status"`
	LastTransitionTime metav1.Time          `json:"lastTransitionTime,omitempty"`
	Message            string               `json:"message,omitempty"`
}

// InstallConditionType defines the possibles conditions that a install might have
type InstallConditionType string

const (
	// FailedInstallConditionType indicates failed condition
	FailedInstallConditionType InstallConditionType = "Failed"
	// ProvisioningInstallConditionType indicates provisioning condition
	ProvisioningInstallConditionType InstallConditionType = "Provisioning"
	// SuccessInstallConditionType indicates success condition
	SuccessInstallConditionType InstallConditionType = "Success"
)

// KogitoInfraConditionType ...
type KogitoInfraConditionType string

const (
	// SuccessInfraConditionType ...
	SuccessInfraConditionType KogitoInfraConditionType = "Success"
	// FailureInfraConditionType ...
	FailureInfraConditionType KogitoInfraConditionType = "Failure"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KogitoInfra is the Schema for the kogitoinfras API
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=kogitoinfras,scope=Namespaced
type KogitoInfra struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KogitoInfraSpec   `json:"spec,omitempty"`
	Status KogitoInfraStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KogitoInfraList contains a list of KogitoInfra
type KogitoInfraList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KogitoInfra `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KogitoInfra{}, &KogitoInfraList{})
}
