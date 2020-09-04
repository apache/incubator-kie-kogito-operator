// Copyright 2020 Red Hat, Inc. and/or its affiliates
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KogitoExplainabilityCRDName is the name of the Kogito Explainability CRD in the cluster.
const KogitoExplainabilityCRDName = "kogitoexplainabilities.app.kiegroup.org"

// KogitoExplainabilitySpec defines the desired state of KogitoExplainability.
// +k8s:openapi-gen=true
type KogitoExplainabilitySpec struct {
	KogitoServiceSpec `json:",inline"`
	KafkaMeta         `json:",inline"`
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// GetRuntime ...
func (d *KogitoExplainabilitySpec) GetRuntime() RuntimeType {
	return QuarkusRuntimeType
}

// KogitoExplainabilityStatus defines the observed state of KogitoExplainability.
// +k8s:openapi-gen=true
type KogitoExplainabilityStatus struct {
	KogitoServiceStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KogitoExplainability defines the Explainability Service infrastructure deployment.
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=kogitoexplainabilities,scope=Namespaced
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".spec.replicas",description="Number of replicas set for this service"
// +kubebuilder:printcolumn:name="Image",type="string",JSONPath=".status.image",description="Base image for this service"
// +kubebuilder:printcolumn:name="Endpoint",type="string",JSONPath=".status.externalURI",description="External URI to access this service"
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Kogito Explainability"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Deployment,apps/v1"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Route,route.openshift.io/v1"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Service,v1"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="KafkaTopic,kafka.strimzi.io/v1beta1"
type KogitoExplainability struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KogitoExplainabilitySpec   `json:"spec,omitempty"`
	Status KogitoExplainabilityStatus `json:"status,omitempty"`
}

// GetSpec ...
func (k *KogitoExplainability) GetSpec() KogitoServiceSpecInterface {
	return &k.Spec
}

// GetStatus ...
func (k *KogitoExplainability) GetStatus() KogitoServiceStatusInterface {
	return &k.Status
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KogitoExplainabilityList contains a list of KogitoExplainability.
type KogitoExplainabilityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// +listType=atomic
	Items []KogitoExplainability `json:"items"`
}

// GetItemsCount ...
func (l *KogitoExplainabilityList) GetItemsCount() int {
	return len(l.Items)
}

// GetItemAt ...
func (l *KogitoExplainabilityList) GetItemAt(index int) KogitoService {
	if len(l.Items) > index {
		return KogitoService(&l.Items[index])
	}
	return nil
}

func init() {
	SchemeBuilder.Register(&KogitoExplainability{}, &KogitoExplainabilityList{})
}
