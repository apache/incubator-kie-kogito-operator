package api

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// KogitoInfraInterface ...
// +kubebuilder:object:generate=false
type KogitoInfraInterface interface {
	metav1.Object
	runtime.Object
	// GetSpec gets the Kogito Service specification structure.
	GetSpec() KogitoInfraSpecInterface
	// GetStatus gets the Kogito Service Status structure.
	GetStatus() KogitoInfraStatusInterface
}

// KogitoInfraSpecInterface ...
// +kubebuilder:object:generate=false
type KogitoInfraSpecInterface interface {
	GetResource() *Resource
	SetResource(resource Resource)
	GetInfraProperties() map[string]string
	SetInfraProperties(infraProperties map[string]string)
}

// KogitoInfraStatusInterface ...
// +kubebuilder:object:generate=false
type KogitoInfraStatusInterface interface {
	GetCondition() *KogitoInfraCondition
	SetCondition(condition KogitoInfraCondition)
	GetRuntimeProperties() RuntimePropertiesMap
	SetRuntimeProperties(runtimeProperties RuntimePropertiesMap)
	GetVolumes() []KogitoInfraVolume
	SetVolumes(infraVolumes []KogitoInfraVolume)
}

// KogitoInfraHandler ...
// +kubebuilder:object:generate=false
type KogitoInfraHandler interface {
	FetchKogitoInfraInstance(key types.NamespacedName) (KogitoInfraInterface, error)
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

// GetResource ...
func (k *KogitoInfraSpec) GetResource() *Resource {
	return &k.Resource
}

// SetResource ...
func (k *KogitoInfraSpec) SetResource(resource Resource) {
	k.Resource = resource
}

// GetInfraProperties ...
func (k *KogitoInfraSpec) GetInfraProperties() map[string]string {
	return k.InfraProperties
}

// SetInfraProperties ...
func (k *KogitoInfraSpec) SetInfraProperties(infraProperties map[string]string) {
	k.InfraProperties = infraProperties
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
	Volumes []KogitoInfraVolume `json:"volumes,omitempty"`
}

// GetCondition ...
func (k *KogitoInfraStatus) GetCondition() *KogitoInfraCondition {
	return &k.Condition
}

// SetCondition ...
func (k *KogitoInfraStatus) SetCondition(condition KogitoInfraCondition) {
	k.Condition = condition
}

// GetRuntimeProperties ...
func (k *KogitoInfraStatus) GetRuntimeProperties() RuntimePropertiesMap {
	return k.RuntimeProperties
}

// SetRuntimeProperties ...
func (k *KogitoInfraStatus) SetRuntimeProperties(runtimeProperties RuntimePropertiesMap) {
	k.RuntimeProperties = runtimeProperties
}

// GetVolumes ...
func (k *KogitoInfraStatus) GetVolumes() []KogitoInfraVolume {
	return k.Volumes
}

// SetVolumes ...
func (k *KogitoInfraStatus) SetVolumes(infraVolumes []KogitoInfraVolume) {
	k.Volumes = infraVolumes
}

// KogitoInfraConditionType ...
type KogitoInfraConditionType string

const (
	// SuccessInfraConditionType ...
	SuccessInfraConditionType KogitoInfraConditionType = "Success"
	// FailureInfraConditionType ...
	FailureInfraConditionType KogitoInfraConditionType = "Failure"
)

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

// ConfigVolume is the Kubernetes Core `Volume` type that holds only configuration volume sources.
type ConfigVolume struct {
	// Volume's name.
	// Must be a DNS_LABEL and unique within the pod.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// ConfigVolumeSource represents the location and type of the mounted volume.
	ConfigVolumeSource `json:",inline" protobuf:"bytes,2,opt,name=volumeSource"`
}

// ToKubeVolume converts the current ConfigVolume instance to Kubernetes Core Volume type.
func (v ConfigVolume) ToKubeVolume() v1.Volume {
	volume := v1.Volume{Name: v.Name}
	volume.Secret = v.Secret
	volume.ConfigMap = v.ConfigMap
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
