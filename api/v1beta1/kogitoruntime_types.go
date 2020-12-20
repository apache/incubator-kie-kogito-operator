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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KogitoRuntimeSpec defines the desired state of KogitoRuntime.
type KogitoRuntimeSpec struct {
	KogitoServiceSpec `json:",inline"`

	// Annotates the pods managed by the operator with the required metadata for Istio to setup its sidecars, enabling the mesh. Defaults to false.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Enable Istio"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	EnableIstio bool `json:"enableIstio,omitempty"`

	// The name of the runtime used, either Quarkus or SpringBoot.
	// Default value: quarkus
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="runtime"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:label"
	// +kubebuilder:validation:Enum=quarkus;springboot
	Runtime RuntimeType `json:"runtime,omitempty"`
}

// KogitoRuntimeStatus defines the observed state of KogitoRuntime.
type KogitoRuntimeStatus struct {
	KogitoServiceStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// KogitoRuntime is a custom Kogito service.
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=kogitoruntimes,scope=Namespaced
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".spec.replicas",description="Number of replicas set for this service"
// +kubebuilder:printcolumn:name="Image",type="string",JSONPath=".status.image",description="Image of this service"
// +kubebuilder:printcolumn:name="Endpoint",type="string",JSONPath=".status.externalURI",description="External URI to access this service"
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Kogito service"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Deployment,apps/v1,\"A Kubernetes Deployment\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Route,route.openshift.io/v1,\"A Openshift Route\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="ConfigMap,v1,\"A Kubernetes ConfigMap\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Service,v1,\"A Kubernetes Service\""
type KogitoRuntime struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KogitoRuntimeSpec   `json:"spec,omitempty"`
	Status KogitoRuntimeStatus `json:"status,omitempty"`
}

// GetRuntime ...
func (k *KogitoRuntimeSpec) GetRuntime() RuntimeType {
	if len(k.Runtime) == 0 {
		k.Runtime = QuarkusRuntimeType
	}
	return k.Runtime
}

// GetSpec ...
func (k *KogitoRuntime) GetSpec() KogitoServiceSpecInterface {
	return &k.Spec
}

// GetStatus ...
func (k *KogitoRuntime) GetStatus() KogitoServiceStatusInterface {
	return &k.Status
}

// +kubebuilder:object:root=true

// KogitoRuntimeList contains a list of KogitoRuntime.
type KogitoRuntimeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KogitoRuntime `json:"items"`
}

// GetItemsCount ...
func (l *KogitoRuntimeList) GetItemsCount() int {
	return len(l.Items)
}

// GetItemAt ...
func (l *KogitoRuntimeList) GetItemAt(index int) KogitoService {
	if len(l.Items) > index {
		return KogitoService(&l.Items[index])
	}
	return nil
}

func init() {
	SchemeBuilder.Register(&KogitoRuntime{}, &KogitoRuntimeList{})
}
