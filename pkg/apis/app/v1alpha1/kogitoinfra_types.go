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

// Resource provide reference infra resource
type Resource struct {

	// APIVersion describes the API Version of referred Kubernetes resource for example, infinispan.org/v1
	APIVersion string `json:"apiVersion"`

	// Kind describes the kind of referred Kubernetes resource for example, Infinispan
	Kind string `json:"kind"`

	// Namespace where referred resource exists.
	Namespace string `json:"namespace,omitempty"`

	// Name of referred resource.
	Name string `json:"name,omitempty"`
}

// KogitoInfraSpec defines the desired state of KogitoInfra.
// +k8s:openapi-gen=true
type KogitoInfraSpec struct {
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// +optional
	// Resource for the service. Example: Infinispan/Kafka/Keycloak/Grafana.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Resource Resource `json:"resource,omitempty"`
}

// KogitoInfraStatus defines the observed state of KogitoInfra.
// +k8s:openapi-gen=true
type KogitoInfraStatus struct {
	Condition KogitoInfraCondition `json:"condition,omitempty"`

	// +optional
	// +mapType=atomic
	// Application properties extracted from the linked resource that will be added to the deployed Kogito service.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	AppProps map[string]string `json:"appProps,omitempty"`

	// +optional
	// +listType=atomic
	// Environment variables extracted from the linked resource that will be added to the deployed Kogito service.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Env []v1.EnvVar `json:"env,omitempty"`
}

/*
	TODO: Change `LastTransitionTime` to Time type when k8s implements a way of validating date-time on non array objects:
	https://github.com/coreos/prometheus-operator/issues/2399#issuecomment-466320464
*/

// KogitoInfraCondition ...
type KogitoInfraCondition struct {
	Type               KogitoInfraConditionType `json:"type"`
	Status             v1.ConditionStatus       `json:"status"`
	LastTransitionTime string                   `json:"lastTransitionTime,omitempty"`
	Message            string                   `json:"message,omitempty"`
}

// KogitoInfraConditionType ...
type KogitoInfraConditionType string

const (
	// SuccessInfraConditionType ...
	SuccessInfraConditionType KogitoInfraConditionType = "Success"
	// FailureInfraConditionType ...
	FailureInfraConditionType KogitoInfraConditionType = "Failure"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KogitoInfra is the resource to bind a Custom Resource (CR) not managed by Kogito Operator to a given deployed Kogito service.
// It holds the reference of a CR managed by another operator such as Strimzi. For example: one can create a Kafka CR via Strimzi
// and link this resource using KogitoInfra to a given Kogito service (custom or supporting, such as Data Index).
// Please refer to the Kogito Operator documentation (https://docs.jboss.org/kogito/release/latest/html_single/) for more information.
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=kogitoinfras,scope=Namespaced
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Kogito Infra"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Kafka,ksafka.strimzi.io/v1beta1,\"A Kafka instance\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Infinispan,infinispan.org/v1,\"A Infinispan instance\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Keycloak,keycloak.org/v1alpha1,\"A Keycloak Instance\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Grafana,integreatly.org/v1alpha,\"A Grafana Instance\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Secret,v1,\"A Kubernetes Secret\""
type KogitoInfra struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KogitoInfraSpec   `json:"spec,omitempty"`
	Status KogitoInfraStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KogitoInfraList contains a list of KogitoInfra.
type KogitoInfraList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KogitoInfra `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KogitoInfra{}, &KogitoInfraList{})
}
