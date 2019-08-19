package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KogitoAppSpec defines the desired state of KogitoApp
// +k8s:openapi-gen=true
type KogitoAppSpec struct {
	// The name of the runtime used, either quarkus or springboot, defaults to quarkus
	Runtime  RuntimeType `json:"runtime,omitempty"`
	Name     string      `json:"name,omitempty"`
	Replicas *int32      `json:"replicas,omitempty"`
	Env      []Env       `json:"env,omitempty"`
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
	Resource *ResourceKind `json:"resource"`
	// Value of this resource in Kubernetes format
	Value *string `json:"value"`
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
	BuilderImage         bool   `json:"builderImage,omitempty"`
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
	// DeploymentFailedReason - Unable to deploy the application
	DeploymentFailedReason ReasonType = "DeploymentFailed"
	// ConfigurationErrorReason - An invalid configuration caused an error
	ConfigurationErrorReason ReasonType = "ConfigurationError"
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
