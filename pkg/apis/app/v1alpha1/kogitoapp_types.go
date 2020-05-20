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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KogitoAppCRDName is the name of the KogitoApp CRD in the cluster
const KogitoAppCRDName = "kogitoapps.app.kiegroup.org"

// KogitoAppSpec defines the desired state of KogitoApp
// +k8s:openapi-gen=true
type KogitoAppSpec struct {
	KogitoServiceSpec `json:",inline"`

	// The name of the runtime used, either Quarkus or Springboot
	// Default value: quarkus
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Runtime"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:label"
	// +kubebuilder:validation:Enum=quarkus;springboot
	Runtime RuntimeType `json:"runtime,omitempty"`

	// S2I Build configuration
	// Default value: nil
	Build *KogitoAppBuildObject `json:"build"`

	// Kubernetes Service configuration
	// Default value: nil
	Service KogitoAppServiceObject `json:"service,omitempty"`

	// Annotates the pods managed by the operator with the required metadata for Istio to setup its sidecars, enabling the mesh. Defaults to false.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Enable Istio"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	EnableIstio bool `json:"enableIstio,omitempty"`

	// Set this property to true to tell the operator to deploy an instance of Infinispan via the Infinispan Operator and
	// configure this service to connect to the deployed server.
	// For Quarkus runtime, it sets QUARKUS_INFINISPAN_CLIENT_* environment variables. For Spring Boot, these variables start with SPRING_INFINISPAN_CLIENT_*.
	// More info: https://github.com/kiegroup/kogito-cloud-operator#kogito-services.
	// Set to false or ignore it if your service does not need persistence or if you are going to configure the persistence infrastructure yourself
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Enable Persistence"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	EnablePersistence bool `json:"enablePersistence,omitempty"`

	// Set this property to true to tell the operator to deploy an instance of Kafka via the Strimzi Operator and configure this service with
	// the proper information to connect to the Kafka cluster.
	// The Kafka cluster service endpoint will be injected in the Kogito Service container via an environment variable named "KAFKA_BOOTSTRAP_SERVERS" e.g.: kafka-kogito:9092.
	// Set to false or ignore it if your service does not need messaging or if you are going to configure the messaging infrastructure yourself
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Enable Events"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	EnableEvents bool `json:"enableEvents,omitempty"`
}

// GetRuntime ...
func (k *KogitoAppSpec) GetRuntime() RuntimeType {
	return k.Runtime
}

// GetBuild ...
func (k *KogitoAppSpec) GetBuild() *KogitoAppBuildObject {
	if k == nil {
		return nil
	}
	return k.Build
}

// IsGitURIEmpty checks if the provided Git URI is empty or not
func (k *KogitoAppSpec) IsGitURIEmpty() bool {
	if k == nil {
		return true
	}
	if k.Build == nil || &k.Build.GitSource == nil {
		return true
	}
	return len(k.Build.GitSource.URI) == 0
}

// KogitoAppBuildObject Data to define how to build an application from source
// +k8s:openapi-gen=true
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Kogito Service Build"
type KogitoAppBuildObject struct {
	Incremental bool `json:"incremental,omitempty"`
	// Environment variables used during build time
	// +listType=atomic
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Build Env Variables"
	Envs []corev1.EnvVar `json:"envs,omitempty"`
	// Information about the git repository where the Kogito App source code resides.
	// If set, the operator will use source to image strategy build.
	// +optional
	GitSource GitSource `json:"gitSource,omitempty"`
	// WebHook secrets for build configs
	// +listType=map
	// +listMapKey=type
	Webhooks []WebhookSecret `json:"webhooks,omitempty"`
	// + optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Images Version"
	// Image version for the Kogito official images used during the build. E.g.: 0.6.0. Default to current Operator version.
	ImageVersion string `json:"imageVersion,omitempty"`
	// Custom image used by the source to image process to build the Kogito Service binaries. Takes precedence over ImageVersion attribute.
	// + optional
	// +kubebuilder:validation:Pattern=`(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]/([a-z0-9-]+)/([a-z0-9-]+):(([a-z0-9\.-]+))`
	ImageS2ITag string `json:"imageS2ITag,omitempty"`
	// Custom image used by the source to image process to build the final Kogito Service image. Takes precedence over ImageVersion attribute.
	// + optional
	// +kubebuilder:validation:Pattern=`(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]/([a-z0-9-]+)/([a-z0-9-]+):(([a-z0-9\.-]+))`
	ImageRuntimeTag string `json:"imageRuntimeTag,omitempty"`
	// Native indicates if the Kogito Service built should be compiled to run on native mode when Runtime is Quarkus. For more information, see https://www.graalvm.org/docs/reference-manual/aot-compilation/.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Native Build"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	Native bool `json:"native,omitempty"`
	// Resources for S2I builder pods.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:resourceRequirements"
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// Internal Maven Mirror to be used during source-to-image builds to considerably increase build speed
	MavenMirrorURL string `json:"mavenMirrorURL,omitempty"`
}

// AddEnvironmentVariable adds new environment variable to build environment variables
func (k *KogitoAppBuildObject) AddEnvironmentVariable(name, value string) {
	env := corev1.EnvVar{
		Name:  name,
		Value: value,
	}
	k.Envs = append(k.Envs, env)
	return
}

// AddResourceRequest adds new resource request. Works also on an uninitialized Requests field.
func (k *KogitoAppBuildObject) AddResourceRequest(name, value string) {
	if k.Resources.Requests == nil {
		k.Resources.Requests = corev1.ResourceList{}
	}

	k.Resources.Requests[corev1.ResourceName(name)] = resource.MustParse(value)
}

// AddResourceLimit adds new resource limit. Works also on an uninitialized Limits field.
func (k *KogitoAppBuildObject) AddResourceLimit(name, value string) {
	if k.Resources.Limits == nil {
		k.Resources.Limits = corev1.ResourceList{}
	}

	k.Resources.Limits[corev1.ResourceName(name)] = resource.MustParse(value)
}

// KogitoAppServiceObject Data to define the service of the Kogito application
// +k8s:openapi-gen=true
type KogitoAppServiceObject struct {
	// Labels for the application service
	Labels map[string]string `json:"labels,omitempty"`
}

// GitSource Git coordinates to locate the source code to build
// +k8s:openapi-gen=true
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Kogito Git Source"
type GitSource struct {
	// Git URI for the s2i source
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Git URI"
	URI string `json:"uri"`
	// Branch to use in the Git repository
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Git Reference"
	Reference string `json:"reference,omitempty"`
	// Context/subdirectory where the code is located, relative to the repo root
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Git Context"
	ContextDir string `json:"contextDir,omitempty"`
}

// WebhookType literal type to distinguish between different types of webhooks
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
	// +kubebuilder:validation:Enum=GitHub;Generic
	Type WebhookType `json:"type,omitempty"`
	// Secret value for webhook
	Secret string `json:"secret,omitempty"`
}

// KogitoAppStatus defines the observed state of KogitoApp
// +k8s:openapi-gen=true
type KogitoAppStatus struct {
	ConditionsMeta `json:",inline"`
	// External URL for the service
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.displayName="Route"
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.x-descriptors="urn:alm:descriptor:org.w3:link"
	Route string `json:"route,omitempty"`
	// History of service deployments status
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.displayName="Deployments"
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:podStatuses"
	Deployments Deployments `json:"deployments"`
	// History of service builds status
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.displayName="Builds"
	Builds Builds `json:"builds"`
}

// RuntimeType - type of condition
type RuntimeType string

const (
	// QuarkusRuntimeType - The KogitoApp is deployed
	QuarkusRuntimeType RuntimeType = "quarkus"
	// SpringbootRuntimeType - The KogitoApp is being provisioned
	SpringbootRuntimeType RuntimeType = "springboot"
)

// Deployments ...
// +k8s:openapi-gen=true
type Deployments struct {
	// Deployments are ready to serve requests
	// +listType=set
	Ready []string `json:"ready,omitempty"`
	// Deployments are starting
	// +listType=set
	Starting []string `json:"starting,omitempty"`
	// Deployments are not starting and the next step is unclear
	// +listType=set
	Stopped []string `json:"stopped,omitempty"`
	// Deployments failed
	// +listType=set
	Failed []string `json:"failed,omitempty"`
}

// Builds ...
// +k8s:openapi-gen=true
type Builds struct {
	// Builds are being created
	// +listType=set
	New []string `json:"new,omitempty"`
	// Builds are about to start running
	// +listType=set
	Pending []string `json:"pending,omitempty"`
	// Builds are running
	// +listType=set
	Running []string `json:"running,omitempty"`
	// Builds have executed and succeeded
	// +listType=set
	Complete []string `json:"complete,omitempty"`
	// Builds have executed and failed
	// +listType=set
	Failed []string `json:"failed,omitempty"`
	// Builds have been prevented from executing by an error
	// +listType=set
	Error []string `json:"error,omitempty"`
	// Builds have been stopped from executing
	// +listType=set
	Cancelled []string `json:"cancelled,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KogitoApp is a project prescription running a Kogito Runtime Service.
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=kogitoapps,scope=Namespaced
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".spec.replicas",description="Number of replicas set for this service"
// +kubebuilder:printcolumn:name="Runtime",type="string",JSONPath=".spec.runtime",description="Runtime used to build the service"
// +kubebuilder:printcolumn:name="Enable Persistence",type="boolean",JSONPath=".spec.enablePersistence",description="Indicates if persistence is enabled"
// +kubebuilder:printcolumn:name="Enable Events",type="boolean",JSONPath=".spec.enableEvents",description="Indicates if events is enabled"
// +kubebuilder:printcolumn:name="Image Version",type="string",JSONPath=".spec.build.imageVersion",description="Build image version"
// +kubebuilder:printcolumn:name="Endpoint",type="string",JSONPath=".status.route",description="External URI to access this service"
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Kogito Service"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="DeploymentConfigs,apps.openshift.io/v1"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="ImageStreams,image.openshift.io/v1"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="BuildConfigs,build.openshift.io/v1"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Routes,route.openshift.io/v1"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="ConfigMaps,v1"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Services,v1"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="ServiceMonitors,monitoring.coreos.com/v1"
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
	// +listType=atomic
	Items []KogitoApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KogitoApp{}, &KogitoAppList{})
}
