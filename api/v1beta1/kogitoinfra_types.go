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
	"github.com/kiegroup/kogito-cloud-operator/api"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KogitoInfraSpec defines the desired state of KogitoInfra.
// +k8s:openapi-gen=true
type KogitoInfraSpec struct {
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// Resource for the service. Example: Infinispan/Kafka/Keycloak.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Resource Resource `json:"resource"`

	// +optional
	// +mapType=atomic
	// Optional properties which would be needed to setup correct runtime/service configuration, based on the resource type.
	// For example, MongoDB will require `username` and `database` as properties for a correct setup, else it will fail
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	InfraProperties map[string]string `json:"infraProperties,omitempty"`
}

// GetResource ...
func (k *KogitoInfraSpec) GetResource() api.ResourceInterface {
	return &k.Resource
}

// GetInfraProperties ...
func (k *KogitoInfraSpec) GetInfraProperties() map[string]string {
	return k.InfraProperties
}

// RuntimeProperties defines the variables that will be
// extracted from the linked resource and added to the
// deployed Kogito service.
type RuntimeProperties struct {
	AppProps map[string]string `json:"appProps,omitempty"`
	Env      []v1.EnvVar       `json:"env,omitempty"`
}

// GetAppProps ...
func (r RuntimeProperties) GetAppProps() map[string]string {
	return r.AppProps
}

// GetEnv ...
func (r RuntimeProperties) GetEnv() []v1.EnvVar {
	return r.Env
}

// KogitoInfraStatus defines the observed state of KogitoInfra.
// +k8s:openapi-gen=true
type KogitoInfraStatus struct {
	Condition KogitoInfraCondition `json:"condition,omitempty"`

	// +optional
	// Runtime variables extracted from the linked resource that will be added to the deployed Kogito service.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	RuntimeProperties map[api.RuntimeType]RuntimeProperties `json:"runtimeProperties,omitempty"`

	// +optional
	// +listType=atomic
	// List of volumes that should be added to the services bound to this infra instance
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Volumes []KogitoInfraVolume `json:"volumes,omitempty"`
}

// GetCondition ...
func (k *KogitoInfraStatus) GetCondition() api.KogitoInfraConditionInterface {
	return &k.Condition
}

// SetCondition ...
func (k *KogitoInfraStatus) SetCondition(condition api.KogitoInfraConditionInterface) {
	if newCondition, ok := condition.(*KogitoInfraCondition); ok {
		k.Condition = *newCondition
	}
}

// GetRuntimeProperties ...
func (k *KogitoInfraStatus) GetRuntimeProperties() api.RuntimePropertiesMap {
	runtimeProperties := make(api.RuntimePropertiesMap, len(k.RuntimeProperties))
	for key, value := range k.RuntimeProperties {
		runtimeProperties[key] = api.RuntimePropertiesInterface(value)
	}
	return runtimeProperties
}

// AddRuntimeProperties ...
func (k *KogitoInfraStatus) AddRuntimeProperties(runtimeType api.RuntimeType, runtimeProperties api.RuntimePropertiesInterface) {
	rp := k.RuntimeProperties
	if rp == nil {
		rp = make(map[api.RuntimeType]RuntimeProperties)
	}
	rp[runtimeType] = runtimeProperties.(RuntimeProperties)
	k.RuntimeProperties = rp
}

// GetVolumes ...
func (k *KogitoInfraStatus) GetVolumes() []api.KogitoInfraVolumeInterface {
	volumes := make([]api.KogitoInfraVolumeInterface, len(k.Volumes))
	for i, v := range k.Volumes {
		volumes[i] = api.KogitoInfraVolumeInterface(v)
	}
	return volumes
}

// SetVolumes ...
func (k *KogitoInfraStatus) SetVolumes(infraVolumes []api.KogitoInfraVolumeInterface) {
	var volumes []KogitoInfraVolume
	for _, volume := range infraVolumes {
		if newVolume, ok := volume.(KogitoInfraVolume); ok {
			volumes = append(volumes, newVolume)
		}
	}
	k.Volumes = volumes
}

// KogitoInfraCondition ...
type KogitoInfraCondition struct {
	// Type ...
	Type api.KogitoInfraConditionType `json:"type"`
	// Status ...
	Status v1.ConditionStatus `json:"status"`
	// LastTransitionTime ...
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Message ...
	Message string `json:"message,omitempty"`
	// Reason ...
	Reason api.KogitoInfraConditionReason `json:"reason,omitempty"`
}

// GetType ...
func (k *KogitoInfraCondition) GetType() api.KogitoInfraConditionType {
	return k.Type
}

// SetType ...
func (k *KogitoInfraCondition) SetType(infraConditionType api.KogitoInfraConditionType) {
	k.Type = infraConditionType
}

// GetStatus ...
func (k *KogitoInfraCondition) GetStatus() v1.ConditionStatus {
	return k.Status
}

// SetStatus ...
func (k *KogitoInfraCondition) SetStatus(status v1.ConditionStatus) {
	k.Status = status
}

// GetLastTransitionTime ...
func (k *KogitoInfraCondition) GetLastTransitionTime() metav1.Time {
	return k.LastTransitionTime
}

// SetLastTransitionTime ...
func (k *KogitoInfraCondition) SetLastTransitionTime(lastTransitionTime metav1.Time) {
	k.LastTransitionTime = lastTransitionTime
}

// GetMessage ...
func (k *KogitoInfraCondition) GetMessage() string {
	return k.Message
}

// SetMessage ...
func (k *KogitoInfraCondition) SetMessage(message string) {
	k.Message = message
}

// GetReason ...
func (k *KogitoInfraCondition) GetReason() api.KogitoInfraConditionReason {
	return k.Reason
}

// SetReason ...
func (k *KogitoInfraCondition) SetReason(reason api.KogitoInfraConditionReason) {
	k.Reason = reason
}

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
	Name string `json:"name"`
}

// GetAPIVersion ...
func (r *Resource) GetAPIVersion() string {
	return r.APIVersion
}

// SetAPIVersion ...
func (r *Resource) SetAPIVersion(apiVersion string) {
	r.APIVersion = apiVersion
}

// GetKind ...
func (r *Resource) GetKind() string {
	return r.Kind
}

// SetKind ...
func (r *Resource) SetKind(kind string) {
	r.Kind = kind
}

// GetNamespace ...
func (r *Resource) GetNamespace() string {
	return r.Namespace
}

// SetNamespace ...
func (r *Resource) SetNamespace(namespace string) {
	r.Namespace = namespace
}

// GetName ...
func (r *Resource) GetName() string {
	return r.Name
}

// SetName ...
func (r *Resource) SetName(name string) {
	r.Name = name
}

/*
BEGIN VOLUME
This was created to not add excessive attributes to our CRD files. As the feature grows, we can keep adding sources.
*/

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

// GetSecret ...
func (c *ConfigVolumeSource) GetSecret() *v1.SecretVolumeSource {
	return c.Secret
}

// SetSecret ...
func (c *ConfigVolumeSource) SetSecret(secret *v1.SecretVolumeSource) {
	c.Secret = secret
}

// GetConfigMap ...
func (c *ConfigVolumeSource) GetConfigMap() *v1.ConfigMapVolumeSource {
	return c.ConfigMap
}

// SetConfigMap ...
func (c *ConfigVolumeSource) SetConfigMap(configMap *v1.ConfigMapVolumeSource) {
	c.ConfigMap = configMap
}

// ConfigVolume is the Kubernetes Core `Volume` type that holds only configuration volume sources.
type ConfigVolume struct {
	// ConfigVolumeSource represents the location and type of the mounted volume.
	ConfigVolumeSource `json:",inline" protobuf:"bytes,2,opt,name=volumeSource"`
	// Volume's name.
	// Must be a DNS_LABEL and unique within the pod.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
}

// GetName ...
func (c *ConfigVolume) GetName() string {
	return c.Name
}

// SetName ...
func (c *ConfigVolume) SetName(name string) {
	c.Name = name
}

// ToKubeVolume converts the current ConfigVolume instance to Kubernetes Core Volume type.
func (c *ConfigVolume) ToKubeVolume() v1.Volume {
	volume := v1.Volume{Name: c.Name}
	volume.Secret = c.Secret
	volume.ConfigMap = c.ConfigMap
	return volume
}

/* END VOLUME */

// KogitoInfraVolume describes the data structure for volumes that should be mounted in the given service provided by this infra instance
type KogitoInfraVolume struct {
	// Mount is the Kubernetes VolumeMount referenced by this instance
	Mount v1.VolumeMount `json:"mount"`
	// NamedVolume describes the pod Volume reference
	NamedVolume ConfigVolume `json:"volume"`
}

// GetMount ...
func (k KogitoInfraVolume) GetMount() v1.VolumeMount {
	return k.Mount
}

// GetNamedVolume ...
func (k KogitoInfraVolume) GetNamedVolume() api.ConfigVolumeInterface {
	return &k.NamedVolume
}

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

// GetSpec provide spec of Kogito infra
func (k *KogitoInfra) GetSpec() api.KogitoInfraSpecInterface {
	return &k.Spec
}

// GetStatus provide status of Kogito infra
func (k *KogitoInfra) GetStatus() api.KogitoInfraStatusInterface {
	return &k.Status
}
