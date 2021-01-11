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

package v1beta1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Resource provide reference infra resource
type Resource struct {

	// APIVersion describes the API Version of referred Kubernetes resource for example, infinispan.org/v1
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="APIVersion"
	APIVersion string `json:"apiVersion"`

	// Kind describes the kind of referred Kubernetes resource for example, Infinispan
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Kind"
	Kind string `json:"kind"`

	// Namespace where referred resource exists.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Namespace"
	Namespace string `json:"namespace,omitempty"`

	// Name of referred resource.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Name"
	Name string `json:"name,omitempty"`
}

// KogitoInfraSpec defines the desired state of KogitoInfra.
// +k8s:openapi-gen=true
type KogitoInfraSpec struct {
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// Resource for the service. Example: Infinispan/Kafka/Keycloak.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Resource Resource `json:"resource,omitempty"`

	// +optional
	// +mapType=atomic
	// Optional properties which would be needed to setup correct runtime/service configuration, based on the resource type.
	// For example, MongoDB will require `username` and `database` as properties for a correct setup, else it will fail
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	InfraProperties map[string]string `json:"infraProperties,omitempty"`
}

// RuntimeProperties defines the variables that will be
// extracted from the linked resource and added to the
// deployed Kogito service.
type RuntimeProperties struct {
	AppProps map[string]string `json:"appProps,omitempty"`
	Env      []v1.EnvVar       `json:"env,omitempty"`
}

// RuntimePropertiesMap defines the map that KogitoInfraStatus
// will use to link the runtime to their variables.
type RuntimePropertiesMap map[RuntimeType]RuntimeProperties

// KogitoInfraStatus defines the observed state of KogitoInfra.
// +k8s:openapi-gen=true
type KogitoInfraStatus struct {
	Condition KogitoInfraCondition `json:"condition,omitempty"`

	// +optional
	// Runtime variables extracted from the linked resource that will be added to the deployed Kogito service.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	RuntimeProperties RuntimePropertiesMap `json:"runtimeProperties,omitempty"`

	// +optional
	// +listType=atomic
	// List of volumes that should be added to the services bound to this infra instance
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Volume []KogitoInfraVolume `json:"volumes,omitempty"`
}

// KogitoInfraConditionReason describes the reasons for reconciliation failure
type KogitoInfraConditionReason string

const (
	// ReconciliationFailure generic failure on reconciliation
	ReconciliationFailure KogitoInfraConditionReason = "ReconciliationFailure"
	// ResourceNotFound target resource not found
	ResourceNotFound KogitoInfraConditionReason = "ResourceNotFound"
	// ResourceAPINotFound API not available in the cluster
	ResourceAPINotFound KogitoInfraConditionReason = "ResourceAPINotFound"
	// UnsupportedAPIKind API defined in the KogitoInfra CR not supported
	UnsupportedAPIKind KogitoInfraConditionReason = "UnsupportedAPIKind"
	// ResourceNotReady related resource is not ready
	ResourceNotReady KogitoInfraConditionReason = "ResourceNotReady"
	// ResourceConfigError related resource is not configured properly
	ResourceConfigError KogitoInfraConditionReason = "ResourceConfigError"
	// ResourceMissingResourceConfig related resource is missing a config information to continue
	ResourceMissingResourceConfig KogitoInfraConditionReason = "ResourceMissingConfig"
)

// KogitoInfraCondition ...
type KogitoInfraCondition struct {
	// Type ...
	Type KogitoInfraConditionType `json:"type"`
	// Status ...
	Status v1.ConditionStatus `json:"status"`
	// LastTransitionTime ...
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Message ...
	Message string `json:"message,omitempty"`
	// Reason ...
	Reason KogitoInfraConditionReason `json:"reason,omitempty"`
}

// KogitoInfraConditionType ...
type KogitoInfraConditionType string

const (
	// SuccessInfraConditionType ...
	SuccessInfraConditionType KogitoInfraConditionType = "Success"
	// FailureInfraConditionType ...
	FailureInfraConditionType KogitoInfraConditionType = "Failure"
)

// KogitoInfraVolume describes the data structure for volumes that should be mounted in the given service provided by this infra instance
type KogitoInfraVolume struct {
	// Mount is the Kubernetes VolumeMount referenced by this instance
	Mount v1.VolumeMount `json:"mount"`
	// NamedVolume describes the pod Volume reference
	NamedVolume ConfigVolume `json:"volume"`
}

/*
BEGIN VOLUME
This was created to not add excessive attributes to our CRD files. As the feature grows, we can keep adding sources.
*/

// ConfigVolume is the Kubernetes Core `Volume` type that holds only configuration volume sources.
type ConfigVolume struct {
	// Volume's name.
	// Must be a DNS_LABEL and unique within the pod.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// ConfigVolumeSource represents the location and type of the mounted volume.
	ConfigVolumeSource `json:",inline" protobuf:"bytes,2,opt,name=volumeSource"`
}

// ConfigVolumeSource is the Kubernetes Core `VolumeSource` type for ConfigMap and Secret only
type ConfigVolumeSource struct {
	// Secret represents a secret that should populate this volume.
	// More info: https://kubernetes.io/docs/concepts/storage/volumes#secret
	// +optional
	Secret *v1.SecretVolumeSource `json:"secret,omitempty" protobuf:"bytes,6,opt,name=secret"`
	// ConfigMap represents a configMap that should populate this volume
	// +optional
	ConfigMap *v1.ConfigMapVolumeSource `json:"configMap,omitempty" protobuf:"bytes,19,opt,name=configMap"`
}

// ToKubeVolume converts the current ConfigVolume instance to Kubernetes Core Volume type.
func (v ConfigVolume) ToKubeVolume() v1.Volume {
	volume := v1.Volume{Name: v.Name}
	volume.Secret = v.Secret
	volume.ConfigMap = v.ConfigMap
	return volume
}

/* END VOLUME */

// +kubebuilder:object:root=true

// KogitoInfra is the resource to bind a Custom Resource (CR) not managed by Kogito Operator to a given deployed Kogito service.
// It holds the reference of a CR managed by another operator such as Strimzi. For example: one can create a Kafka CR via Strimzi
// and link this resource using KogitoInfra to a given Kogito service (custom or supporting, such as Data Index).
// Please refer to the Kogito Operator documentation (https://docs.jboss.org/kogito/release/latest/html_single/) for more information.
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=kogitoinfras,scope=Namespaced
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Resource Name",type="string",JSONPath=".spec.resource.name",description="Third Party Infrastructure Resource"
// +kubebuilder:printcolumn:name="Kind",type="string",JSONPath=".spec.resource.kind",description="Kubernetes CR Kind"
// +kubebuilder:printcolumn:name="API Version",type="string",JSONPath=".spec.resource.apiVersion",description="Kubernetes CR API Version"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.condition.status",description="General Status of this resource bind"
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=".status.condition.reason",description="Status reason"
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Kogito Infra"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Kafka,kafka.strimzi.io/v1beta1,\"A Kafka instance\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Infinispan,infinispan.org/v1,\"A Infinispan instance\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Keycloak,keycloak.org/v1alpha1,\"A Keycloak Instance\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Secret,v1,\"A Kubernetes Secret\""
type KogitoInfra struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KogitoInfraSpec   `json:"spec,omitempty"`
	Status KogitoInfraStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KogitoInfraList contains a list of KogitoInfra.
type KogitoInfraList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KogitoInfra `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KogitoInfra{}, &KogitoInfraList{})
}
