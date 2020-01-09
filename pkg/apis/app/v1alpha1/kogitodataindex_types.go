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
	InfinispanMeta `json:",inline"`
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// Replicas is the number of pod replicas that the Data Index Service will create
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:validation:Minimum=0
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	//Env is a collection of additional environment variables to add to the Data Index container
	Env map[string]string `json:"env,omitempty"`

	// +optional
	// Image to use for this service
	Image string `json:"image,omitempty"`

	// +optional
	// MemoryLimit is the limit of Memory for the container
	MemoryLimit string `json:"memoryLimit,omitempty"`

	// +optional
	// MemoryRequest is the request of Memory for the container
	MemoryRequest string `json:"memoryRequest,omitempty"`

	// +optional
	// CPULimit is the limit of CPU for the container
	CPULimit string `json:"cpuLimit,omitempty"`

	// +optional
	// CPURequest is the request of CPU for the container
	CPURequest string `json:"cpuRequest,omitempty"`

	// +optional
	// Kafka has the data used by the Kogito Data Index to connect to a Kafka cluster
	Kafka KafkaConnectionProperties `json:"kafka,omitempty"`
}

// KafkaConnectionProperties has the data needed to connect to a Kafka cluster
type KafkaConnectionProperties struct {
	// +optional
	// URI is the service URI to connect to the Kafka cluster, for example, my-cluster-kafka-bootstrap:9092
	ExternalURI string `json:"externalURI,omitempty"`

	// +optional
	// Instance is the Kafka instance to be used, for example, kogito-kafka
	Instance string `json:"instance,omitempty"`
}

// KogitoDataIndexStatus defines the observed state of KogitoDataIndex
// +k8s:openapi-gen=true
type KogitoDataIndexStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - Define observed state of cluster
	// IMPORTANT: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// Status of the Data Index Service Deployment created and managed by it
	DeploymentStatus appsv1.StatefulSetStatus `json:"deploymentStatus,omitempty"`

	// Status of the Data Index Service created and managed by it
	ServiceStatus corev1.ServiceStatus `json:"serviceStatus,omitempty"`

	// OK when all resources are created successfully
	// +listType=atomic
	Conditions []DataIndexCondition `json:"conditions,omitempty"`

	// All dependencies OK means that everything was found within the namespace
	// +listType=set
	DependenciesStatus []DataIndexDependenciesStatus `json:"dependenciesStatus,omitempty"`

	// Route is where the service is exposed
	Route string `json:"route,omitempty"`
}

// DataIndexDependenciesStatus indicates all possible statuses that the dependencies can have
type DataIndexDependenciesStatus string

const (
	//DataIndexDependenciesStatusOK - All dependencies have been met
	DataIndexDependenciesStatusOK DataIndexDependenciesStatus = "OK"
	//DataIndexDependenciesStatusMissingKafka - Kafka is missing
	DataIndexDependenciesStatusMissingKafka DataIndexDependenciesStatus = "Missing Kafka"
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
