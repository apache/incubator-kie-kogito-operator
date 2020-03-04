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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KogitoAppCRDName is the name of the KogitoApp CRD in the cluster
const KogitoAppCRDName = "kogitoapps.app.kiegroup.org"

// KogitoAppSpec defines the desired state of KogitoApp
// +k8s:openapi-gen=true
type KogitoAppSpec struct {
	// The name of the runtime used, either Quarkus or Springboot
	// Default value: quarkus
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Runtime"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:label"
	// +kubebuilder:validation:Enum=quarkus;springboot
	Runtime RuntimeType `json:"runtime,omitempty"`

	// Number of replicas that the service will have deployed in the cluster
	// Default value: 1
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Replicas"
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	Replicas *int32 `json:"replicas,omitempty"`

	// Environment variables for the runtime service
	// Default value: nil
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Environment Variables"
	// +listType=map
	// +listMapKey=name
	Env []Env `json:"env,omitempty"`

	// The resources for the deployed pods, like memory and cpu
	// Default value: nil
	Resources Resources `json:"resources,omitempty"`

	// S2I Build configuration
	// Default value: nil
	Build *KogitoAppBuildObject `json:"build"`

	// Kubernetes Service configuration
	// Default value: nil
	Service KogitoAppServiceObject `json:"service,omitempty"`

	// Infrastructure definition
	Infra KogitoAppInfra `json:"infra,omitempty"`

	// Annotates the pods managed by the operator with the required metadata for Istio to setup its sidecars, enabling the mesh. Defaults to false.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Enable Istio"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	EnableIstio bool `json:"enableIstio,omitempty"`
}

// Resources Data to define Resources needed for each deployed pod
// +k8s:openapi-gen=true
type Resources struct {
	// +listType=map
	// +listMapKey=resource
	Limits []ResourceMap `json:"limits,omitempty"`
	// +listType=map
	// +listMapKey=resource
	Requests []ResourceMap `json:"requests,omitempty"`
}

// ResourceKind is the Resource Type accepted for resources
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
	// Resource type like CPU and memory
	// +kubebuilder:validation:Enum=cpu;memory
	Resource ResourceKind `json:"resource"`
	// Value of this resource in Kubernetes format
	Value string `json:"value"`
}

// Env Data to define environment variables in key-value pair fashion
// +k8s:openapi-gen=true
type Env struct {
	// Name of an environment variable
	Name string `json:"name,omitempty"`
	// Value for that environment variable
	Value string `json:"value,omitempty"`
}

// KogitoAppBuildObject Data to define how to build an application from source
// +k8s:openapi-gen=true
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Kogito Service Build"
type KogitoAppBuildObject struct {
	Incremental bool `json:"incremental,omitempty"`
	// Environment variables used during build time
	// +listType=map
	// +listMapKey=name
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Build Env Variables"
	Env       []Env      `json:"env,omitempty"`
	GitSource *GitSource `json:"gitSource"`
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
	// Resources for build pods. Default limits are 1GB RAM/0.5 CPU on JVM and 4GB RAM/1 CPU for native builds.
	Resources Resources `json:"resources,omitempty"`
	// Internal Maven Mirror to be used during source-to-image builds to considerably increase build speed
	MavenMirrorURL string `json:"mavenMirrorURL,omitempty"`
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
	URI *string `json:"uri"`
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

// KogitoAppInfraInstallInfinispanType defines the Infinispan installation mode
type KogitoAppInfraInstallInfinispanType string

const (
	// KogitoAppInfraInstallInfinispanAlways - Always installs Infinispan
	KogitoAppInfraInstallInfinispanAlways KogitoAppInfraInstallInfinispanType = "Always"
	// KogitoAppInfraInstallInfinispanNever - Never installs Infinispan
	KogitoAppInfraInstallInfinispanNever KogitoAppInfraInstallInfinispanType = "Never"
	// KogitoAppInfraInstallInfinispanAuto - The Operator will try to discover if the service needs persistence by scanning the runtime image metadata
	KogitoAppInfraInstallInfinispanAuto KogitoAppInfraInstallInfinispanType = "Auto"
)

// KogitoAppInfraInstallKafkaType defines the Kafka	 installation mode
type KogitoAppInfraInstallKafkaType string

const (
	// KogitoAppInfraInstallKafkaAlways - Always installs Kafka
	KogitoAppInfraInstallKafkaAlways KogitoAppInfraInstallKafkaType = "Always"
	// KogitoAppInfraInstallKafkaNever - Never installs Kafka
	KogitoAppInfraInstallKafkaNever KogitoAppInfraInstallKafkaType = "Never"
)

// KogitoAppInfra defines details regarding the Kogito Infrastructure to support the deployed Kogito Service
type KogitoAppInfra struct {
	// By default Kogito Operator installs an Infinispan instance in the namespace if the service needs persistence ('Auto').
	// Set to 'Never' to disable this behavior, e.g. if the service will use another persistence mechanism.
	// Set to 'Always' to always install Infinispan, even if the service won't need persistence.
	// For Quarkus runtime, it sets QUARKUS_INFINISPAN_CLIENT_* environment variables. For Spring Boot, these variables start with SPRING_INFINISPAN_CLIENT_*.
	// More info: https://github.com/kiegroup/kogito-cloud-operator#kogito-services.
	// Default to 'Auto', which means it installs Infinispan if the service requires persistence.
	// +kubebuilder:validation:Enum=Always;Never;Auto
	InstallInfinispan KogitoAppInfraInstallInfinispanType `json:"installInfinispan,omitempty"`
	// Set to 'Always' to have Kafka installed automatically via Strimzi Operator when deploying the Kogito Service. 'Never' otherwise.
	// The Kafka cluster service endpoint will be inject in the Kogito Service container via an environment variable named "KAFKA_BOOTSTRAP_SERVERS" e.g.: kafka-kogito:9092
	// Default to 'Never'
	// +kubebuilder:validation:Enum=Always;Never
	InstallKafka KogitoAppInfraInstallKafkaType `json:"installKafka,omitempty"`
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
