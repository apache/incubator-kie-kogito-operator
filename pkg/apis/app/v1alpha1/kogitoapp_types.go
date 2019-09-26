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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KogitoAppCRDName is the name of the Kogito App CRD in the cluster
const KogitoAppCRDName = "kogitoapps.app.kiegroup.org"

// KogitoAppSpec defines the desired state of KogitoApp
// +k8s:openapi-gen=true
type KogitoAppSpec struct {
	// The name of the runtime used, either quarkus or springboot, defaults to quarkus
	// +kubebuilder:validation:Enum=quarkus,springboot
	Runtime RuntimeType `json:"runtime,omitempty"`
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	Replicas *int32 `json:"replicas,omitempty"`
	Env      []Env  `json:"env,omitempty"`
	// The resources for the deployed pods, like memory and cpu
	Resources Resources `json:"resources,omitempty"`
	// S2I Build configuration
	Build *KogitoAppBuildObject `json:"build"`
	// Service configuration
	Service KogitoAppServiceObject `json:"service,omitempty"`
}

// Resources Data to define Resources needed for each deployed pod
// +k8s:openapi-gen=true
type Resources struct {
	Limits   []ResourceMap `json:"limits,omitempty"`
	Requests []ResourceMap `json:"requests,omitempty"`
}

// ResourceKind is the Resource Type accept for resources
type ResourceKind string

const (
	// ResourceCPU is the CPU resource
	ResourceCPU ResourceKind = "cpu"
	// ResourceMemory is the Memory resource
	ResourceMemory ResourceKind = "memory"
)

// ResourceMap Data to define a list of possible Resources
// +k8s:openapi-gen=true
type ResourceMap struct {
	// Resource type like cpu and memory
	// +kubebuilder:validation:Enum=cpu,memory
	Resource ResourceKind `json:"resource"`
	// Value of this resource in Kubernetes format
	Value string `json:"value"`
}

// Env Data to define environment variables in key/value pair fashion
// +k8s:openapi-gen=true
type Env struct {
	// Name of an environment variable
	Name string `json:"name,omitempty"`
	// Value for that environment variable
	Value string `json:"value,omitempty"`
}

// KogitoAppBuildObject Data to define how to build an application from source
// +k8s:openapi-gen=true
type KogitoAppBuildObject struct {
	Incremental bool       `json:"incremental,omitempty"`
	Env         []Env      `json:"env,omitempty"`
	GitSource   *GitSource `json:"gitSource"`
	// WebHook secrets for build configs
	Webhooks []WebhookSecret `json:"webhooks,omitempty"`
	// ImageS2I is used by build configurations to build the image from source
	ImageS2I Image `json:"imageS2I,omitempty"`
	// ImageRuntime is used build configurations to build a final runtime image based on a s2i configuration
	ImageRuntime Image `json:"imageRuntime,omitempty"`
	// Native indicates if the Kogito Service built should be compiled to run on native mode when Runtime is quarkus. See: https://www.graalvm.org/docs/reference-manual/aot-compilation/
	Native bool `json:"native,omitempty"`
	// Resources for build pods. Default limits are 1GB RAM/0.5 cpu on jvm and 4GB RAM/1 cpu for native builds.
	Resources Resources `json:"resources,omitempty"`
}

// KogitoAppServiceObject Data to define the service of the kogito app
// +k8s:openapi-gen=true
type KogitoAppServiceObject struct {
	// Labels for the application service
	Labels map[string]string `json:"labels,omitempty"`
}

// GitSource Git coordinates to locate the source code to build
// +k8s:openapi-gen=true
type GitSource struct {
	// Git URI for the s2i source
	URI *string `json:"uri"`
	// Branch to use in the git repository
	Reference string `json:"reference,omitempty"`
	// Context/subdirectory where the code is located, relatively to repo root
	ContextDir string `json:"contextDir,omitempty"`
}

// WebhookType literal type to distinguish between different types of Webhooks
type WebhookType string

const (
	// GitHubWebhook GitHub webhook
	GitHubWebhook WebhookType = "GitHub"
	// GenericWebhook Generic webhook
	GenericWebhook WebhookType = "Generic"
)

// WebhookSecret Secret to use for a given webhook
// +k8s:openapi-gen=true
type WebhookSecret struct {
	// WebHook type, either GitHub or Generic
	// +kubebuilder:validation:Enum=GitHub,Generic
	Type WebhookType `json:"type,omitempty"`
	// Secret value for webhook
	Secret string `json:"secret,omitempty"`
}

// KogitoAppStatus defines the observed state of KogitoApp
// +k8s:openapi-gen=true
type KogitoAppStatus struct {
	Conditions  []Condition `json:"conditions"`
	Route       string      `json:"route,omitempty"`
	Deployments Deployments `json:"deployments"`
	Builds      Builds      `json:"builds"`
}

// RuntimeType - type of condition
type RuntimeType string

const (
	// QuarkusRuntimeType - the kogitoapp is deployed
	QuarkusRuntimeType RuntimeType = "quarkus"
	// SpringbootRuntimeType - the kogitoapp is being provisioned
	SpringbootRuntimeType RuntimeType = "springboot"
)

// Image - image details
type Image struct {
	ImageStreamName      string `json:"imageStreamName,omitempty"`
	ImageStreamTag       string `json:"imageStreamTag,omitempty"`
	ImageStreamNamespace string `json:"imageStreamNamespace,omitempty"`
}

// ConditionType - type of condition
type ConditionType string

const (
	// DeployedConditionType - the kogitoapp is deployed
	DeployedConditionType ConditionType = "Deployed"
	// ProvisioningConditionType - the kogitoapp is being provisioned
	ProvisioningConditionType ConditionType = "Provisioning"
	// FailedConditionType - the kogitoapp is in a failed state
	FailedConditionType ConditionType = "Failed"
)

// ReasonType - type of reason
type ReasonType string

const (
	// ParseCRRequestFailedReason - Unable to resolve the CR request
	ParseCRRequestFailedReason ReasonType = "ParseCRRequestFailed"
	// RetrieveDeployedResourceFailedReason - Unable to retrieve the deployed resources
	RetrieveDeployedResourceFailedReason ReasonType = "RetrieveDeployedResourceFailed"
	// CreateResourceFailedReason - Unable to create the requested resources
	CreateResourceFailedReason ReasonType = "CreateResourceFailed"
	// RemoveResourceFailedReason - Unable to remove the requested resources
	RemoveResourceFailedReason ReasonType = "RemoveResourceFailed"
	// UpdateResourceFailedReason - Unable to update the requested resources
	UpdateResourceFailedReason ReasonType = "UpdateResourceFailed"
	// TriggerBuildFailedReason - Unable to trigger the builds
	TriggerBuildFailedReason ReasonType = "TriggerBuildFailed"
	// BuildS2IFailedReason - Unable to build with the s2i image
	BuildS2IFailedReason ReasonType = "BuildS2IFailedReason"
	// BuildRuntimeFailedReason - Unable to build the runtime image
	BuildRuntimeFailedReason ReasonType = "BuildRuntimeFailedReason"
	// UnknownReason - Unable to determine the error
	UnknownReason ReasonType = "Unknown"
)

// Condition - The condition for the kogito-cloud-operator
// +k8s:openapi-gen=true
type Condition struct {
	Type               ConditionType          `json:"type"`
	Status             corev1.ConditionStatus `json:"status"`
	LastTransitionTime metav1.Time            `json:"lastTransitionTime,omitempty"`
	Reason             ReasonType             `json:"reason,omitempty"`
	Message            string                 `json:"message,omitempty"`
}

// Deployments ...
// +k8s:openapi-gen=true
type Deployments struct {
	// Deployments are ready to serve requests
	Ready []string `json:"ready,omitempty"`
	// Deployments are starting, may or may not succeed
	Starting []string `json:"starting,omitempty"`
	// Deployments are not starting, unclear what next step will be
	Stopped []string `json:"stopped,omitempty"`
	// Deployments failed
	Failed []string `json:"failed,omitempty"`
}

// Builds ...
// +k8s:openapi-gen=true
type Builds struct {
	// Builds are being newly created
	New []string `json:"new,omitempty"`
	// Builds are about to start running
	Pending []string `json:"pending,omitempty"`
	// Builds are running
	Running []string `json:"running,omitempty"`
	// Builds have been successful
	Complete []string `json:"complete,omitempty"`
	// Builds have executed and failed
	Failed []string `json:"failed,omitempty"`
	// Builds have been prevented from executing by error
	Error []string `json:"error,omitempty"`
	// Builds have been stopped from executing
	Cancelled []string `json:"cancelled,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KogitoApp is the Schema for the kogitoapps API
// +k8s:openapi-gen=true
type KogitoApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KogitoAppSpec   `json:"spec,omitempty"`
	Status KogitoAppStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KogitoAppList contains a list of KogitoApp
type KogitoAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KogitoApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KogitoApp{}, &KogitoAppList{})
}
