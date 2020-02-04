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

// KogitoDataIndexCRDName is the name of the Kogito Data Index CRD in the cluster
const KogitoDataIndexCRDName = "kogitodataindices.app.kiegroup.org"

// KogitoDataIndexSpec defines the desired state of KogitoDataIndex
// +k8s:openapi-gen=true
type KogitoDataIndexSpec struct {
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// HttpPort will set the environment env KOGITO_DATA_INDEX_HTTP_PORT to define which port data-index service will listen internally.
	// +optional
	HTTPPort int32 `json:"httpPort,omitempty"`

	// Replicas is the number of pod replicas that the Data Index Service will create
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:validation:Minimum=0
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Replicas"
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	//Env is a collection of additional environment variables to add to the Data Index container
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Environment Variables"
	Env map[string]string `json:"env,omitempty"`

	// +optional
	// Image to use for this service
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Image"
	Image string `json:"image,omitempty"`

	// +optional
	// MemoryLimit is the limit of Memory for the container
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Memory Limit"
	MemoryLimit string `json:"memoryLimit,omitempty"`

	// +optional
	// MemoryRequest is the request of Memory for the container
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Memory Request"
	MemoryRequest string `json:"memoryRequest,omitempty"`

	// +optional
	// CPULimit is the limit of CPU for the container
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="CPU Limit"
	CPULimit string `json:"cpuLimit,omitempty"`

	// +optional
	// CPURequest is the request of CPU for the container
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="CPU Request"
	CPURequest string `json:"cpuRequest,omitempty"`

	InfinispanMeta `json:",inline"`

	KafkaMeta `json:",inline"`

	KeycloakMeta `json:",inline"`
}

// KogitoDataIndexStatus defines the observed state of KogitoDataIndex
// +k8s:openapi-gen=true
type KogitoDataIndexStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - Define observed state of cluster
	// IMPORTANT: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// Status of the Data Index Service Deployment created and managed by it
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=false
	DeploymentStatus appsv1.DeploymentStatus `json:"deploymentStatus,omitempty"`

	// Status of the Data Index Service created and managed by it
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=false
	ServiceStatus corev1.ServiceStatus `json:"serviceStatus,omitempty"`

	// OK when all resources are created successfully
	// +listType=atomic
	Conditions []DataIndexCondition `json:"conditions,omitempty"`

	// All dependencies OK means that everything was found within the namespace
	// +listType=set
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	DependenciesStatus []DataIndexDependenciesStatus `json:"dependenciesStatus,omitempty"`

	// Route is where the service is exposed
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.x-descriptors="urn:alm:descriptor:org.w3:link"
	Route string `json:"route,omitempty"`
}

// DataIndexDependenciesStatus indicates all possible statuses that the dependencies can have
type DataIndexDependenciesStatus string

const (
	//DataIndexDependenciesStatusOK - All dependencies have been met
	DataIndexDependenciesStatusOK DataIndexDependenciesStatus = "OK"
	//DataIndexDependenciesStatusMissingKafka - Kafka is missing
	DataIndexDependenciesStatusMissingKafka DataIndexDependenciesStatus = "Missing Kafka"
	//DataIndexDependenciesStatusMissingKeycloak - Keycloak is missing
	DataIndexDependenciesStatusMissingKeycloak DataIndexDependenciesStatus = "Missing Keycloak"
	//DataIndexDependenciesStatusMissingInfinispan - Infinispan is missing
	DataIndexDependenciesStatusMissingInfinispan DataIndexDependenciesStatus = "Missing Infinispan"
)

// DataIndexCondition indicates the possible conditions for the Data Index Service
type DataIndexCondition struct {
	Condition          DataIndexConditionType `json:"condition"`
	Message            string                 `json:"message,omitempty"`
	LastTransitionTime metav1.Time            `json:"lastTransitionTime,omitempty"`
}

// DataIndexConditionType indicates the possible status that the resource can have
type DataIndexConditionType string

const (
	// ConditionOK - Everything was created successfully
	ConditionOK DataIndexConditionType = "OK"
	// ConditionProvisioning - The service is still being deployed
	ConditionProvisioning DataIndexConditionType = "Provisioning"
	// ConditionFailed - The service and its dependencies failed to deploy
	ConditionFailed DataIndexConditionType = "Failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KogitoDataIndex is the Schema for the kogitodataindices API
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=kogitodataindices,scope=Namespaced
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Kogito Data Index"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Deployments,apps/v1"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Routes,route.openshift.io/v1"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="ConfigMaps,v1"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Services,v1"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="KafkaTopics,kafka.strimzi.io/v1beta1"
type KogitoDataIndex struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KogitoDataIndexSpec   `json:"spec,omitempty"`
	Status KogitoDataIndexStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KogitoDataIndexList contains a list of KogitoDataIndex
type KogitoDataIndexList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// +listType=atomic
	Items []KogitoDataIndex `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KogitoDataIndex{}, &KogitoDataIndexList{})
}
