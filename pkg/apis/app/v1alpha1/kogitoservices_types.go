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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// KogitoService defines the interface for any Kogito Service that the operator can handle, e.g. Data Index, Jobs Service, Runtimes, etc.
type KogitoService interface {
	metav1.Object
	runtime.Object
	// GetSpec gets the Kogito Service specification structure
	GetSpec() KogitoServiceSpecInterface
	// GetStatus gets the Kogito Service Status structure
	GetStatus() KogitoServiceStatusInterface
}

// KogitoServiceList defines a base interface for Kogito Service list
type KogitoServiceList interface {
	runtime.Object
	// GetItemsCount gets the number of items in the list
	GetItemsCount() int
	// GetItemAt gets the item at the given index
	GetItemAt(index int) KogitoService
}

// KogitoServiceStatusInterface defines the basic interface for the Kogito Service status
type KogitoServiceStatusInterface interface {
	ConditionMetaInterface
	GetDeploymentConditions() []appsv1.DeploymentCondition
	SetDeploymentConditions(deploymentConditions []appsv1.DeploymentCondition)
	GetImage() string
	SetImage(image string)
	GetExternalURI() string
	SetExternalURI(uri string)
}

// KogitoServiceStatus is the basic structure for any Kogito Service status
type KogitoServiceStatus struct {
	ConditionsMeta `json:",inline"`
	// General conditions for the Kogito Service deployment
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Deployment Conditions"
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.x-descriptors="urn:alm:descriptor:io.kubernetes.conditions"
	DeploymentConditions []appsv1.DeploymentCondition `json:"deploymentConditions,omitempty"`
	// Image is the resolved image for this service
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	Image string `json:"image,omitempty"`
	// URI is where the service is exposed
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.x-descriptors="urn:alm:descriptor:org.w3:link"
	ExternalURI string `json:"externalURI,omitempty"`
}

// GetDeploymentConditions gets the deployment conditions for the service
func (k *KogitoServiceStatus) GetDeploymentConditions() []appsv1.DeploymentCondition {
	return k.DeploymentConditions
}

// SetDeploymentConditions sets the deployment conditions for the service
func (k *KogitoServiceStatus) SetDeploymentConditions(deploymentConditions []appsv1.DeploymentCondition) {
	k.DeploymentConditions = deploymentConditions
}

// GetImage ...
func (k *KogitoServiceStatus) GetImage() string { return k.Image }

// SetImage ...
func (k *KogitoServiceStatus) SetImage(image string) { k.Image = image }

// GetExternalURI ...
func (k *KogitoServiceStatus) GetExternalURI() string { return k.ExternalURI }

// SetExternalURI ...
func (k *KogitoServiceStatus) SetExternalURI(uri string) { k.ExternalURI = uri }

// KogitoServiceSpecInterface defines the interface for the Kogito Service specification, it's the basic structure for any Kogito Service
type KogitoServiceSpecInterface interface {
	GetReplicas() int32
	SetReplicas(replicas int32)
	GetEnvs() []corev1.EnvVar
	SetEnvs(envs []corev1.EnvVar)
	GetImage() *Image
	SetImage(image Image)
	GetResources() corev1.ResourceRequirements
	SetResources(resources corev1.ResourceRequirements)
}

// KogitoServiceSpec is the basic structure for the Kogito Service specification
type KogitoServiceSpec struct {
	// Number of replicas that the service will have deployed in the cluster
	// Default value: 1
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Replicas"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:podCount"
	// +kubebuilder:validation:Minimum=0
	Replicas int32 `json:"replicas"`

	// +optional
	// +listType=atomic
	// Environment variables to be added to the runtime container. Keys must be a C_IDENTIFIER.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Envs []corev1.EnvVar `json:"envs,omitempty"`

	// +optional
	// Image definition for the service. Example: Domain: quay.io, Namespace: kiegroup, Name: kogito-jobs-service, Tag: latest.
	// On OpenShift an ImageStream will be created in the current namespace pointing to the given image.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Image Image `json:"image,omitempty"`

	// Defined Resources for the Jobs Service
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:resourceRequirements"
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// GetReplicas ...
func (k *KogitoServiceSpec) GetReplicas() int32 { return k.Replicas }

// SetReplicas ...
func (k *KogitoServiceSpec) SetReplicas(replicas int32) { k.Replicas = replicas }

// GetEnvs ...
func (k *KogitoServiceSpec) GetEnvs() []corev1.EnvVar { return k.Envs }

// SetEnvs ...
func (k *KogitoServiceSpec) SetEnvs(envs []corev1.EnvVar) { k.Envs = envs }

// GetImage ...
func (k *KogitoServiceSpec) GetImage() *Image { return &k.Image }

// SetImage ...
func (k *KogitoServiceSpec) SetImage(image Image) { k.Image = image }

// GetResources ...
func (k *KogitoServiceSpec) GetResources() corev1.ResourceRequirements { return k.Resources }

// SetResources ...
func (k *KogitoServiceSpec) SetResources(resources corev1.ResourceRequirements) {
	k.Resources = resources
}
