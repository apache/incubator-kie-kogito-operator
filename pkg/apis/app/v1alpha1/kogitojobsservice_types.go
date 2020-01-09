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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KogitoJobsServiceSpec defines the desired state of KogitoJobsService
// +k8s:openapi-gen=true
type KogitoJobsServiceSpec struct {
	InfinispanMeta `json:",inline"`
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// Number of replicas that the service will have deployed in the cluster
	// Default value: 1
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Replicas"
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1
	Replicas int32 `json:"replicas"`

	// +optional
	// +listType=atomic
	// Environment variables to be added to the runtime container. Keys must be a C_IDENTIFIER.
	Envs []corev1.EnvVar `json:"envs,omitempty"`

	// +optional
	// Image definition for the service. Example: Domain: quay.io, Namespace: kiegroup, Name: kogito-jobs-service, Tag: latest
	// Defaults to quay.io/kiegroup/kogito-jobs-service:latest
	// On OpenShift an ImageStream will be created in the current namespace pointing to the given image.
	Image Image `json:"image,omitempty"`

	// Defined Resources for the Jobs Service
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Resources"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.description="Resources"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:resourceRequirements"
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// +optional
	// Retry backOff time in milliseconds between the job execution attempts, in case of execution failure.
	// Default to service default, see: https://github.com/kiegroup/kogito-runtimes/wiki/Jobs-Service#configuration-properties
	BackOffRetryMillis int64 `json:"backOffRetryMillis,omitempty"`

	// +optional
	// Maximum interval in milliseconds when retrying to execute jobs, in case of failures.
	// Default to service default, see: https://github.com/kiegroup/kogito-runtimes/wiki/Jobs-Service#configuration-properties
	MaxIntervalLimitToRetryMillis int64 `json:"maxIntervalLimitToRetryMillis,omitempty"`
}

// KogitoJobsServiceStatus defines the observed state of KogitoJobsService
// +k8s:openapi-gen=true
type KogitoJobsServiceStatus struct {
	ConditionsMeta `json:",inline"`
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// DeploymentStatus is the detailed status for the Jobs Service deployment
	DeploymentStatus appsv1.DeploymentStatus `json:"deploymentStatus,omitempty"`
	// Image is the resolved image for this service
	Image string `json:"image,omitempty"`
	// URI is where the service is exposed
	ExternalURI string `json:"externalURI,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KogitoJobsService is the Schema for the kogitojobsservices API
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=kogitojobsservices,scope=Namespaced
type KogitoJobsService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KogitoJobsServiceSpec   `json:"spec,omitempty"`
	Status KogitoJobsServiceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KogitoJobsServiceList contains a list of KogitoJobsService
type KogitoJobsServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KogitoJobsService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KogitoJobsService{}, &KogitoJobsServiceList{})
}
