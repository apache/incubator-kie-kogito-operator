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

// ServiceType define resource type of supporting service
type ServiceType string

const (
	//SupportingServiceAnnotationKey is the key, needs to be defined when creating kogitoSupportingService
	SupportingServiceAnnotationKey = "org.kie.kogito/supportServiceType"
)

const (
	// DataIndex supporting service resource type
	DataIndex ServiceType = "DataIndex"
	// JobsService supporting service resource type
	JobsService ServiceType = "JobsService"
	// MgmtConsole supporting service resource type
	MgmtConsole ServiceType = "MgmtConsole"
	// Explainablity supporting service resource type
	Explainablity ServiceType = "Explainablity"
	// TrustyAI supporting service resource type
	TrustyAI ServiceType = "TrustyAI"
	// TrustyUI supporting service resource type
	TrustyUI ServiceType = "TrustyUI"
)

// KogitoSupportingServiceSpec defines the desired state of KogitoSupportingService.
// +k8s:openapi-gen=true
type KogitoSupportingServiceSpec struct {
	KogitoServiceSpec `json:",inline"`

	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Resource Type"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:text"
	// +kubebuilder:validation:Enum=DataIndex;JobsService;MgmtConsole;Explainablity;TrustyAI;TrustyUI
	// Defines the type for the supporting service, eg: DataIndex, JobsService
	ServiceType ServiceType `json:"serviceType"`
}

// GetRuntime ...
func (k *KogitoSupportingServiceSpec) GetRuntime() RuntimeType {
	return QuarkusRuntimeType
}

// KogitoSupportingServiceStatus defines the observed state of KogitoSupportingService.
// +k8s:openapi-gen=true
type KogitoSupportingServiceStatus struct {
	KogitoServiceStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KogitoSupportingService deploys the Supporting service in the given namespace.
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=kogitosupportingservices,scope=Namespaced
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".spec.replicas",description="Number of replicas set for this service"
// +kubebuilder:printcolumn:name="Image",type="string",JSONPath=".status.image",description="Base image for this service"
// +kubebuilder:printcolumn:name="Endpoint",type="string",JSONPath=".status.externalURI",description="External URI to access this service"
// +kubebuilder:printcolumn:name="Service Type",type="string",JSONPath=".spec.serviceType",description="Supporting Service Type"
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Kogito Supporting Service"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Deployment,apps/v1,\"A Kubernetes Deployment\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Service,v1,\"A Kubernetes Service\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="ImageStream,image.openshift.io/v1,\"A Openshift ImageStream\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Route,route.openshift.io/v1,\"A Openshift Route\""
type KogitoSupportingService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KogitoSupportingServiceSpec   `json:"spec,omitempty"`
	Status KogitoSupportingServiceStatus `json:"status,omitempty"`
}

// GetSpec ...
func (k *KogitoSupportingService) GetSpec() KogitoServiceSpecInterface {
	return &k.Spec
}

// GetStatus ...
func (k *KogitoSupportingService) GetStatus() KogitoServiceStatusInterface {
	return &k.Status
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KogitoSupportingServiceList contains a list of KogitoSupportingService.
type KogitoSupportingServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KogitoSupportingService `json:"items"`
}

// GetItemsCount ...
func (l *KogitoSupportingServiceList) GetItemsCount() int {
	return len(l.Items)
}

// GetItemAt ...
func (l *KogitoSupportingServiceList) GetItemAt(index int) KogitoService {
	if len(l.Items) > index {
		return KogitoService(&l.Items[index])
	}
	return nil
}

// SetAnnotations overwrite the main function
func (k *KogitoSupportingService) SetAnnotations(annotations map[string]string) {
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[SupportingServiceAnnotationKey] = string(k.Spec.ServiceType)
	k.Annotations = annotations
}

// GetAnnotations overwrite the main function
func (k *KogitoSupportingService) GetAnnotations() map[string]string {
	if k.Annotations == nil {
		k.Annotations = map[string]string{}
	}
	k.Annotations[SupportingServiceAnnotationKey] = string(k.Spec.ServiceType)
	return k.Annotations
}

func init() {
	SchemeBuilder.Register(&KogitoSupportingService{}, &KogitoSupportingServiceList{})
}
