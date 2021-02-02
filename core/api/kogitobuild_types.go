// Copyright 2021 Red Hat, Inc. and/or its affiliates
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

package api

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// KogitoBuildInterface ...
// +kubebuilder:object:generate=false
type KogitoBuildInterface interface {
	metav1.Object
	runtime.Object
	// GetSpec gets the Kogito Service specification structure.
	GetSpec() KogitoBuildSpecInterface
	// GetStatus gets the Kogito Service Status structure.
	GetStatus() KogitoBuildStatusInterface
}

// KogitoBuildSpecInterface ...
// +kubebuilder:object:generate=false
type KogitoBuildSpecInterface interface {
	GetType() KogitoBuildType
	SetType(buildType KogitoBuildType)
	IsDisableIncremental() bool
	SetDisableIncremental(disableIncremental bool)
	GetEnv() []corev1.EnvVar
	SetEnv(env []corev1.EnvVar)
	GetGitSource() GitSource
	SetGitSource(gitSource GitSource)
	GetRuntime() RuntimeType
	SetRuntime(runtime RuntimeType)
	GetWebHooks() []WebHookSecret
	SetWebHooks(webhooks []WebHookSecret)
	IsNative() bool
	SetNative(native bool)
	GetResources() corev1.ResourceRequirements
	SetResources(resources corev1.ResourceRequirements)
	GetMavenMirrorURL() string
	SetMavenMirrorURL(mavenMirrorURL string)
	GetBuildImage() string
	SetBuildImage(buildImage string)
	GetRuntimeImage() string
	SetRuntimeImage(runtime string)
	GetTargetKogitoRuntime() string
	SetTargetKogitoRuntime(targetRuntime string)
	GetArtifact() Artifact
	SetArtifact(artifact Artifact)
	IsEnableMavenDownloadOutput() bool
	SetEnableMavenDownloadOutput(enableMavenDownloadOutput bool)
}

// KogitoBuildStatusInterface ...
// +kubebuilder:object:generate=false
type KogitoBuildStatusInterface interface {
	GetLatestBuild() string
	SetLatestBuild(latestBuild string)
	GetConditions() []KogitoBuildConditions
	SetConditions(conditions []KogitoBuildConditions)
	GetBuilds() Builds
	SetBuilds(builds Builds)
}

// KogitoBuildHandler ...
// +kubebuilder:object:generate=false
type KogitoBuildHandler interface {
	FetchKogitoBuildInstance(key types.NamespacedName) (KogitoBuildInterface, error)
}

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
	Env []corev1.EnvVar `json:"env,omitempty"`

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
	WebHooks []WebHookSecret `json:"webHooks,omitempty"`

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
	// Example: "quay.io/kiegroup/kogito-jvm-builder:latest".
	// On OpenShift an ImageStream will be created in the current namespace pointing to the given image.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Build Image"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +optional
	BuildImage string `json:"buildImage,omitempty"`

	// Image used as the base image for the final Kogito service. This image only has the required packages to run the application.
	// For example: quarkus based services will have only JVM installed, native services only the packages required by the OS.
	// The operator will use the one provided by the Kogito Team based on the "Runtime" field.
	// Example: "quay.io/kiegroup/kogito-jvm-builder:latest".
	// On OpenShift an ImageStream will be created in the current namespace pointing to the given image.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Base Image"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +optional
	RuntimeImage string `json:"runtimeImage,omitempty"`

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

// AddResourceRequest adds new resource request. Works also on an uninitialized Requests field.
func (k *KogitoBuildSpec) AddResourceRequest(name, value string) {
	if k.Resources.Requests == nil {
		k.Resources.Requests = corev1.ResourceList{}
	}

	k.Resources.Requests[corev1.ResourceName(name)] = resource.MustParse(value)
}

// AddResourceLimit adds new resource limit. Works also on an uninitialized Limits field.
func (k *KogitoBuildSpec) AddResourceLimit(name, value string) {
	if k.Resources.Limits == nil {
		k.Resources.Limits = corev1.ResourceList{}
	}

	k.Resources.Limits[corev1.ResourceName(name)] = resource.MustParse(value)
}

// GetType ...
func (k *KogitoBuildSpec) GetType() KogitoBuildType {
	return k.Type
}

// SetType ...
func (k *KogitoBuildSpec) SetType(buildType KogitoBuildType) {
	k.Type = buildType
}

// IsDisableIncremental ...
func (k *KogitoBuildSpec) IsDisableIncremental() bool {
	return k.DisableIncremental
}

// SetDisableIncremental ...
func (k *KogitoBuildSpec) SetDisableIncremental(disableIncremental bool) {
	k.DisableIncremental = disableIncremental
}

// GetEnv ...
func (k *KogitoBuildSpec) GetEnv() []corev1.EnvVar {
	return k.Env
}

// SetEnv ...
func (k *KogitoBuildSpec) SetEnv(env []corev1.EnvVar) {
	k.Env = env
}

// GetGitSource ...
func (k *KogitoBuildSpec) GetGitSource() GitSource {
	return k.GitSource
}

// SetGitSource ...
func (k *KogitoBuildSpec) SetGitSource(gitSource GitSource) {
	k.GitSource = gitSource
}

// GetRuntime ...
func (k *KogitoBuildSpec) GetRuntime() RuntimeType {
	return k.Runtime
}

// SetRuntime ...
func (k *KogitoBuildSpec) SetRuntime(runtime RuntimeType) {
	k.Runtime = runtime
}

// GetWebHooks ...
func (k *KogitoBuildSpec) GetWebHooks() []WebHookSecret {
	return k.WebHooks
}

// SetWebHooks ...
func (k *KogitoBuildSpec) SetWebHooks(webhooks []WebHookSecret) {
	k.WebHooks = webhooks
}

// IsNative ...
func (k *KogitoBuildSpec) IsNative() bool {
	return k.Native
}

// SetNative ...
func (k *KogitoBuildSpec) SetNative(native bool) {
	k.Native = native
}

// GetResources ...
func (k *KogitoBuildSpec) GetResources() corev1.ResourceRequirements {
	return k.Resources
}

// SetResources ...
func (k *KogitoBuildSpec) SetResources(resources corev1.ResourceRequirements) {
	k.Resources = resources
}

// GetMavenMirrorURL ...
func (k *KogitoBuildSpec) GetMavenMirrorURL() string {
	return k.MavenMirrorURL
}

// SetMavenMirrorURL ...
func (k *KogitoBuildSpec) SetMavenMirrorURL(mavenMirrorURL string) {
	k.MavenMirrorURL = mavenMirrorURL
}

// GetBuildImage ...
func (k *KogitoBuildSpec) GetBuildImage() string {
	return k.BuildImage
}

// SetBuildImage ...
func (k *KogitoBuildSpec) SetBuildImage(buildImage string) {
	k.BuildImage = buildImage
}

// GetRuntimeImage ...
func (k *KogitoBuildSpec) GetRuntimeImage() string {
	return k.RuntimeImage
}

// SetRuntimeImage ...
func (k *KogitoBuildSpec) SetRuntimeImage(runtime string) {
	k.RuntimeImage = runtime
}

// GetTargetKogitoRuntime ...
func (k *KogitoBuildSpec) GetTargetKogitoRuntime() string {
	return k.TargetKogitoRuntime
}

// SetTargetKogitoRuntime ....
func (k *KogitoBuildSpec) SetTargetKogitoRuntime(targetRuntime string) {
	k.TargetKogitoRuntime = targetRuntime
}

// GetArtifact ...
func (k *KogitoBuildSpec) GetArtifact() Artifact {
	return k.Artifact
}

// SetArtifact ...
func (k *KogitoBuildSpec) SetArtifact(artifact Artifact) {
	k.Artifact = artifact
}

// IsEnableMavenDownloadOutput ...
func (k *KogitoBuildSpec) IsEnableMavenDownloadOutput() bool {
	return k.EnableMavenDownloadOutput
}

// SetEnableMavenDownloadOutput ...
func (k *KogitoBuildSpec) SetEnableMavenDownloadOutput(enableMavenDownloadOutput bool) {
	k.EnableMavenDownloadOutput = enableMavenDownloadOutput
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

// GetLatestBuild ...
func (k *KogitoBuildStatus) GetLatestBuild() string {
	return k.LatestBuild
}

// SetLatestBuild ...
func (k *KogitoBuildStatus) SetLatestBuild(latestBuild string) {
	k.LatestBuild = latestBuild
}

// GetConditions ...
func (k *KogitoBuildStatus) GetConditions() []KogitoBuildConditions {
	return k.Conditions
}

// SetConditions ...
func (k *KogitoBuildStatus) SetConditions(conditions []KogitoBuildConditions) {
	k.Conditions = conditions
}

// GetBuilds ...
func (k *KogitoBuildStatus) GetBuilds() Builds {
	return k.Builds
}

// SetBuilds ...
func (k *KogitoBuildStatus) SetBuilds(builds Builds) {
	k.Builds = builds
}

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

// Builds ...
// +k8s:openapi-gen=true
type Builds struct {
	// Builds are being created.
	// +listType=set
	New []string `json:"new,omitempty"`
	// Builds are about to start running.
	// +listType=set
	Pending []string `json:"pending,omitempty"`
	// Builds are running.
	// +listType=set
	Running []string `json:"running,omitempty"`
	// Builds have executed and succeeded.
	// +listType=set
	Complete []string `json:"complete,omitempty"`
	// Builds have executed and failed.
	// +listType=set
	Failed []string `json:"failed,omitempty"`
	// Builds have been prevented from executing by an error.
	// +listType=set
	Error []string `json:"error,omitempty"`
	// Builds have been stopped from executing.
	// +listType=set
	Cancelled []string `json:"cancelled,omitempty"`
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
