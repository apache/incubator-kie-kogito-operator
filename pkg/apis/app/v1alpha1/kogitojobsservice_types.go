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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KogitoJobsServiceSpec defines the desired state of KogitoJobsService.
// +k8s:openapi-gen=true
type KogitoJobsServiceSpec struct {
	KogitoServiceSpec `json:",inline"`
}

// GetRuntime ...
func (j *KogitoJobsServiceSpec) GetRuntime() RuntimeType {
	return QuarkusRuntimeType
}

// KogitoJobsServiceStatus defines the observed state of KogitoJobsService.
// +k8s:openapi-gen=true
type KogitoJobsServiceStatus struct {
	KogitoServiceStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KogitoJobsService deploys the Kogito Jobs Service in the given namespace.
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=kogitojobsservices,scope=Namespaced
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".spec.replicas",description="Number of replicas set for this service"
// +kubebuilder:printcolumn:name="Image",type="string",JSONPath=".status.image",description="Base image for this service"
// +kubebuilder:printcolumn:name="Endpoint",type="string",JSONPath=".status.externalURI",description="External URI to access this service"
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Kogito Jobs Services"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Deployment,apps/v1,\"A Kubernetes Deployment\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Service,v1,\"A Kubernetes Service\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="ImageStream,image.openshift.io/v1,\"A Openshift ImageStream\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Route,route.openshift.io/v1,\"A Openshift Route\""
type KogitoJobsService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KogitoJobsServiceSpec   `json:"spec,omitempty"`
	Status KogitoJobsServiceStatus `json:"status,omitempty"`
}

// GetSpec ...
func (k *KogitoJobsService) GetSpec() KogitoServiceSpecInterface {
	return &k.Spec
}

// GetStatus ...
func (k *KogitoJobsService) GetStatus() KogitoServiceStatusInterface {
	return &k.Status
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KogitoJobsServiceList contains a list of KogitoJobsService.
type KogitoJobsServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KogitoJobsService `json:"items"`
}

// GetItemsCount ...
func (l *KogitoJobsServiceList) GetItemsCount() int {
	return len(l.Items)
}

// GetItemAt ...
func (l *KogitoJobsServiceList) GetItemAt(index int) KogitoService {
	if len(l.Items) > index {
		return KogitoService(&l.Items[index])
	}
	return nil
}

func init() {
	SchemeBuilder.Register(&KogitoJobsService{}, &KogitoJobsServiceList{})
}
