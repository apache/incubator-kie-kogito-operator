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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KogitoBuildType describes the build types supported by the KogitoBuild CR
type KogitoBuildType string

const (
	// BinaryBuildType builds takes an uploaded binary file already compiled and creates a Kogito service image from it.
	BinaryBuildType KogitoBuildType = "Binary"
	// RemoteSourceBuildType builds pulls the source code from a Git repository, builds the binary and then the final Kogito service image.
	RemoteSourceBuildType KogitoBuildType = "RemoteSource"
	// LocalSourceBuildType builds takes an uploaded resource files such as DRL (rules), DMN (decision) or BPMN (process), builds the binary and the final Kogito service image.
	LocalSourceBuildType KogitoBuildType = "LocalSource"
)

// KogitoBuildSpec defines the desired state of KogitoBuild.
type KogitoBuildSpec struct {

	// Sets the type of build that this instance will handle:
	// Binary - takes an uploaded binary file already compiled and creates a Kogito service image from it.
	// RemoteSource - pulls the source code from a Git repository, builds the binary and then the final Kogito service image.
	// LocalSource - takes an uploaded resource file such as DRL (rules), DMN (decision) or BPMN (process), builds the binary and the final Kogito service image.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="DisableIncremental Builds"
	// +kubebuilder:validation:Enum=Binary;RemoteSource;LocalSource
	Type KogitoBuildType `json:"type"`

	// DisableIncremental indicates that source to image builds should NOT be incremental. Defaults to false.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="DisableIncremental Builds"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	// +optional
	DisableIncremental bool `json:"disableIncremental,omitempty"`

	// Environment variables used during build time.
	// +listType=atomic
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Build Env Variables"
	// +optional
	Envs []corev1.EnvVar `json:"envs,omitempty"`

	// Information about the git repository where the Kogito Service source code resides.
	// Ignored for binary builds.
	// +optional
	GitSource GitSource `json:"gitSource,omitempty"`

	// Which runtime Kogito service base image to use when building the Kogito service.
	// If "BuildImage" is set, this value is ignored by the operator.
	// Default value: quarkus.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Runtime"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:label"
	// +optional
	// +kubebuilder:validation:Enum=quarkus;springboot
	Runtime RuntimeType `json:"runtime,omitempty"`

	// WebHooks secrets for source to image builds based on Git repositories (Remote Sources).
	// +listType=atomic
	// +optional
	WebHooks []WebhookSecret `json:"webHooks,omitempty"`

	// Native indicates if the Kogito Service built should be compiled to run on native mode when Runtime is Quarkus (Source to Image build only).
	// For more information, see https://www.graalvm.org/docs/reference-manual/aot-compilation/.
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Native Build"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	Native bool `json:"native,omitempty"`

	// Resources Requirements for builder pods.
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:resourceRequirements"
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Maven Mirror URL to be used during source-to-image builds (Local and Remote) to considerably increase build speed.
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Maven Mirror URL"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:label"
	MavenMirrorURL string `json:"mavenMirrorURL,omitempty"`

	// Image used to build the Kogito Service from source (Local and Remote).
	// The operator will use the one provided by the Kogito Team based on the "Runtime" field.
	// Example: Domain: quay.io, Namespace: kiegroup, Name: kogito-jvm-builder, Tag: latest.
	// On OpenShift an ImageStream will be created in the current namespace pointing to the given image.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Build Image"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +optional
	BuildImage Image `json:"buildImage,omitempty"`

	// Image used as the base image for the final Kogito service. This image only has the required packages to run the application.
	// For example: quarkus based services will have only JVM installed, native services only the packages required by the OS.
	// The operator will use the one provided by the Kogito Team based on the "Runtime" field.
	// Example: Domain: quay.io, Namespace: kiegroup, Name: kogito-jvm-builder, Tag: latest.
	// On OpenShift an ImageStream will be created in the current namespace pointing to the given image.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Base Image"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +optional
	RuntimeImage Image `json:"runtimeImage,omitempty"`

	// Set this field targeting the desired KogitoRuntime when this KogitoBuild instance has a different name than the KogitoRuntime.
	// By default this KogitoBuild instance will generate a final image named after its own name (.metadata.name).
	// On OpenShift, an ImageStream will be created causing a redeployment on any KogitoRuntime with the same name.
	// On Kubernetes, the final image will be pushed to the KogitoRuntime deployment.
	// If you have multiple KogitoBuild instances (let's say BinaryBuildType and Remote Source), you might need that both target the same KogitoRuntime.
	// Both KogitoBuilds will update the same ImageStream or generate a final image to the same KogitoRuntime deployment.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Base Image"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +optional
	TargetKogitoRuntime string `json:"targetKogitoRuntime,omitempty"`

	// Artifact contains override information for building the Maven artifact (used for Local Source builds).
	// You might want to override this information when building from decisions, rules or process files.
	// In this scenario the Kogito Images will generate a new Java project for you underneath.
	// This information will be used to generate this project.
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Final Artifact"
	Artifact Artifact `json:"artifact,omitempty"`

	// If set to true will print the logs for downloading/uploading of maven dependencies. Defaults to false.
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Enable Maven Download Output"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	EnableMavenDownloadOutput bool `json:"enableMavenDownloadOutput,omitempty"`
}

// KogitoBuildStatus defines the observed state of KogitoBuild.
// +k8s:openapi-gen=true
type KogitoBuildStatus struct {
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.displayName="Latest Build"
	LatestBuild string `json:"latestBuild,omitempty"`
	// +listType=atomic
	// History of conditions for the resource, shows the status of the younger builder controlled by this instance
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.x-descriptors="urn:alm:descriptor:io.kubernetes.conditions"
	Conditions []KogitoBuildConditions `json:"conditions"`
	// History of builds
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.displayName="Builds"
	Builds Builds `json:"builds"`
}

// KogitoBuildConditionType ...
type KogitoBuildConditionType string

const (
	// KogitoBuildSuccessful condition for a successful build.
	KogitoBuildSuccessful KogitoBuildConditionType = "Successful"
	// KogitoBuildFailure condition for a failure build.
	KogitoBuildFailure KogitoBuildConditionType = "Failed"
	// KogitoBuildRunning condition for a running build.
	KogitoBuildRunning KogitoBuildConditionType = "Running"
)

// KogitoBuildConditionReason ...
type KogitoBuildConditionReason string

const (
	// OperatorFailureReason when operator fails to reconcile.
	OperatorFailureReason KogitoBuildConditionReason = "OperatorFailure"
	// BuildFailureReason when build fails.
	BuildFailureReason KogitoBuildConditionReason = "BuildFailure"
)

// KogitoBuildConditions describes the conditions for this build instance according to Kubernetes status interface.
type KogitoBuildConditions struct {
	// Type of this condition
	Type KogitoBuildConditionType `json:"type"`
	// Status ...
	Status corev1.ConditionStatus `json:"status"`
	// LastTransitionTime ...
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Reason of this condition
	Reason KogitoBuildConditionReason `json:"reason,omitempty"`
	// Message ...
	Message string `json:"message,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KogitoBuild handles how to build a custom Kogito service in a Kubernetes/OpenShift cluster.
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=kogitobuilds,scope=Namespaced
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type",description="Type of this build instance"
// +kubebuilder:printcolumn:name="Runtime",type="string",JSONPath=".spec.runtime",description="Runtime used to build the service"
// +kubebuilder:printcolumn:name="Native",type="boolean",JSONPath=".spec.native",description="Indicates it's a native build"
// +kubebuilder:printcolumn:name="Maven URL",type="string",JSONPath=".spec.mavenMirrorURL",description="URL for the proxy Maven repository"
// +kubebuilder:printcolumn:name="Kogito Runtime",type="string",JSONPath=".spec.targetKogitoRuntime",description="Target KogitoRuntime for this build"
// +kubebuilder:printcolumn:name="Git Repository",type="string",JSONPath=".spec.gitSource.uri",description="Git repository URL (RemoteSource builds only)"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="ImageStreams,image.openshift.io/v1"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="BuildConfigs,build.openshift.io/v1"
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Kogito Build"
type KogitoBuild struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KogitoBuildSpec   `json:"spec,omitempty"`
	Status KogitoBuildStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KogitoBuildList contains a list of KogitoBuild.
type KogitoBuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// +listType=atomic
	Items []KogitoBuild `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KogitoBuild{}, &KogitoBuildList{})
}
