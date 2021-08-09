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
	"github.com/kiegroup/kogito-operator/api"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KogitoInfraSpec defines the desired state of KogitoInfra.
// +k8s:openapi-gen=true
type KogitoInfraSpec struct {
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// Resource for the service. Example: Infinispan/Kafka/Keycloak.
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Resource InfraResource `json:"resource,omitempty"`

	// +optional
	// +mapType=atomic
	// Optional properties which would be needed to setup correct runtime/service configuration, based on the resource type.
	// For example, MongoDB will require `username` and `database` as properties for a correct setup, else it will fail
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	InfraProperties map[string]string `json:"infraProperties,omitempty"`

	// +optional
	// +listType=atomic
	// Environment variables to be added to the runtime container. Keys must be a C_IDENTIFIER.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Envs []corev1.EnvVar `json:"envs,omitempty"`

	// +optional
	// +listType=atomic
	// List of configmap that should be added to the services bound to this infra instance
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	ConfigMapReferences []ConfigMapReference `json:"configMapReferences,omitempty"`

	// +optional
	// +listType=atomic
	// List of secret that should be mount to the services bound to this infra instance
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	SecretReferences []SecretReference `json:"secretReferences,omitempty"`
}

// GetResource ...
func (k *KogitoInfraSpec) GetResource() api.ResourceInterface {
	return &k.Resource
}

// GetInfraProperties ...
func (k *KogitoInfraSpec) GetInfraProperties() map[string]string {
	return k.InfraProperties
}

// GetEnvs ...
func (k *KogitoInfraSpec) GetEnvs() []corev1.EnvVar {
	return k.Envs
}

// GetConfigMapReferences ...
func (k *KogitoInfraSpec) GetConfigMapReferences() []api.ConfigMapReferenceInterface {
	newConfigMapReferences := make([]api.ConfigMapReferenceInterface, len(k.ConfigMapReferences))
	for i, v := range k.ConfigMapReferences {
		item := v
		newConfigMapReferences[i] = &item
	}
	return newConfigMapReferences
}

// GetSecretReferences ...
func (k *KogitoInfraSpec) GetSecretReferences() []api.SecretReferenceInterface {
	newSecretReferences := make([]api.SecretReferenceInterface, len(k.SecretReferences))
	for i, v := range k.SecretReferences {
		item := v
		newSecretReferences[i] = &item
	}
	return newSecretReferences
}

// KogitoInfraStatus defines the observed state of KogitoInfra.
// +k8s:openapi-gen=true
type KogitoInfraStatus struct {
	// +listType=atomic
	// History of conditions for the resource
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.x-descriptors="urn:alm:descriptor:io.kubernetes.conditions"
	Conditions *[]metav1.Condition `json:"conditions"`

	// +optional
	// +listType=atomic
	// Environment variables to be added to the runtime container. Keys must be a C_IDENTIFIER.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Envs []corev1.EnvVar `json:"envs,omitempty"`

	// +optional
	// +listType=atomic
	// List of configmap that should be added to the services bound to this infra instance
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	ConfigMapReferences []ConfigMapReference `json:"configMapReferences,omitempty"`

	// +optional
	// +listType=atomic
	// List of secret that should be munted to the services bound to this infra instance
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	SecretReferences []SecretReference `json:"secretReferences,omitempty"`
}

// GetConditions ...
func (k *KogitoInfraStatus) GetConditions() *[]metav1.Condition {
	return k.Conditions
}

// SetConditions ...
func (k *KogitoInfraStatus) SetConditions(conditions *[]metav1.Condition) {
	k.Conditions = conditions
}

// GetEnvs ...
func (k *KogitoInfraStatus) GetEnvs() []corev1.EnvVar {
	return k.Envs
}

// SetEnvs ...
func (k *KogitoInfraStatus) SetEnvs(envs []corev1.EnvVar) {
	k.Envs = envs
}

// GetConfigMapReferences ...
func (k *KogitoInfraStatus) GetConfigMapReferences() []api.ConfigMapReferenceInterface {
	newConfigMapReferences := make([]api.ConfigMapReferenceInterface, len(k.ConfigMapReferences))
	for i, v := range k.ConfigMapReferences {
		item := v
		newConfigMapReferences[i] = &item
	}
	return newConfigMapReferences
}

// SetConfigMapReferences ...
func (k *KogitoInfraStatus) SetConfigMapReferences(configMapReferences []api.ConfigMapReferenceInterface) {
	var newProduces []ConfigMapReference
	for _, produce := range configMapReferences {
		if newProduce, ok := produce.(*ConfigMapReference); ok {
			newProduces = append(newProduces, *newProduce)
		}
	}
	k.ConfigMapReferences = newProduces
}

// GetSecretReferences ...
func (k *KogitoInfraStatus) GetSecretReferences() []api.SecretReferenceInterface {
	newSecretReferences := make([]api.SecretReferenceInterface, len(k.SecretReferences))
	for i, v := range k.SecretReferences {
		item := v
		newSecretReferences[i] = &item
	}
	return newSecretReferences
}

// SetSecretReferences ...
func (k *KogitoInfraStatus) SetSecretReferences(secretReferences []api.SecretReferenceInterface) {
	var newSecretReferences []SecretReference
	for _, produce := range secretReferences {
		if newProduce, ok := produce.(*SecretReference); ok {
			newSecretReferences = append(newSecretReferences, *newProduce)
		}
	}
	k.SecretReferences = newSecretReferences
}

// InfraResource provide reference infra resource
type InfraResource struct {

	// APIVersion describes the API Version of referred Kubernetes resource for example, infinispan.org/v1
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="APIVersion"
	APIVersion string `json:"apiVersion"`

	// Kind describes the kind of referred Kubernetes resource for example, Infinispan
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Kind"
	Kind string `json:"kind"`

	// +optional
	// Namespace where referred resource exists.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Namespace"
	Namespace string `json:"namespace,omitempty"`

	// Name of referred resource.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Name"
	Name string `json:"name"`
}

// GetAPIVersion ...
func (r *InfraResource) GetAPIVersion() string {
	return r.APIVersion
}

// SetAPIVersion ...
func (r *InfraResource) SetAPIVersion(apiVersion string) {
	r.APIVersion = apiVersion
}

// GetKind ...
func (r *InfraResource) GetKind() string {
	return r.Kind
}

// SetKind ...
func (r *InfraResource) SetKind(kind string) {
	r.Kind = kind
}

// GetNamespace ...
func (r *InfraResource) GetNamespace() string {
	return r.Namespace
}

// SetNamespace ...
func (r *InfraResource) SetNamespace(namespace string) {
	r.Namespace = namespace
}

// GetName ...
func (r *InfraResource) GetName() string {
	return r.Name
}

// SetName ...
func (r *InfraResource) SetName(name string) {
	r.Name = name
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +genclient
// +groupName=app.kiegroup.org
// +groupGoName=Kogito
// +kubebuilder:resource:path=kogitoinfras,scope=Namespaced
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Resource Name",type="string",JSONPath=".spec.resource.name",description="Third Party Infrastructure Resource"
// +kubebuilder:printcolumn:name="Kind",type="string",JSONPath=".spec.resource.kind",description="Kubernetes CR Kind"
// +kubebuilder:printcolumn:name="API Version",type="string",JSONPath=".spec.resource.apiVersion",description="Kubernetes CR API Version"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.condition.status",description="General Status of this resource bind"
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=".status.condition.reason",description="Status reason"
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Kogito Infra"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Kafka,kafka.strimzi.io/v1beta2,\"A Kafka instance\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Infinispan,infinispan.org/v1,\"A Infinispan instance\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Keycloak,keycloak.org/v1alpha1,\"A Keycloak Instance\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Secret,v1,\"A Kubernetes Secret\""

// KogitoInfra is the resource to bind a Custom Resource (CR) not managed by Kogito Operator to a given deployed Kogito service.
// It holds the reference of a CR managed by another operator such as Strimzi. For example: one can create a Kafka CR via Strimzi
// and link this resource using KogitoInfra to a given Kogito service (custom or supporting, such as Data Index).
// Please refer to the Kogito Operator documentation (https://docs.jboss.org/kogito/release/latest/html_single/) for more information.
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

// GetSpec provide spec of Kogito infra
func (k *KogitoInfra) GetSpec() api.KogitoInfraSpecInterface {
	return &k.Spec
}

// GetStatus provide status of Kogito infra
func (k *KogitoInfra) GetStatus() api.KogitoInfraStatusInterface {
	return &k.Status
}
