package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// OpenShiftObject ...
type OpenShiftObject interface {
	metav1.Object
	runtime.Object
}

// SubAppSpec defines the desired state of SubApp
// +k8s:openapi-gen=true
type SubAppSpec struct {
	SubSet
}

// SubSet ...
// +k8s:openapi-gen=true
type SubSet struct {
	Runtime   RuntimeType                 `json:"runtime,omitempty"`
	Name      string                      `json:"name,omitempty"`
	Replicas  *int32                      `json:"replicas,omitempty"`
	Env       []corev1.EnvVar             `json:"env,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources"`
	Build     *SubAppBuildObject          `json:"build,omitempty"` // S2I Build configuration
}

// SubAppBuildObject Data to define how to build an application from source
// +k8s:openapi-gen=true
type SubAppBuildObject struct {
	Incremental bool            `json:"incremental,omitempty"`
	Env         []corev1.EnvVar `json:"env,omitempty"`
	GitSource   GitSource       `json:"gitSource,omitempty"`
	Webhooks    []WebhookSecret `json:"webhooks,omitempty"`
}

// GitSource Git coordinates to locate the source code to build
// +k8s:openapi-gen=true
type GitSource struct {
	URI        string `json:"uri,omitempty"`
	Reference  string `json:"reference,omitempty"`
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
	Type   WebhookType `json:"type,omitempty"`
	Secret string      `json:"secret,omitempty"`
}

// SubAppStatus defines the observed state of SubApp
// +k8s:openapi-gen=true
type SubAppStatus struct {
	Conditions  []Condition `json:"conditions"`
	Route       string      `json:"route,omitempty"`
	Deployments Deployments `json:"deployments"`
}

// RuntimeType - type of condition
type RuntimeType string

const (
	// QuarkusRuntimeType - the subapp is deployed
	QuarkusRuntimeType RuntimeType = "quarkus"
	// SpringbootRuntimeType - the subapp is being provisioned
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
	// DeployedConditionType - the subapp is deployed
	DeployedConditionType ConditionType = "Deployed"
	// ProvisioningConditionType - the subapp is being provisioned
	ProvisioningConditionType ConditionType = "Provisioning"
	// FailedConditionType - the subapp is in a failed state
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

// Condition - The condition for the submarine-cloud-operator
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

// SubApp is the Schema for the subapps API
// +k8s:openapi-gen=true
type SubApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubAppSpec   `json:"spec,omitempty"`
	Status SubAppStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SubAppList contains a list of SubApp
type SubAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SubApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SubApp{}, &SubAppList{})
}
