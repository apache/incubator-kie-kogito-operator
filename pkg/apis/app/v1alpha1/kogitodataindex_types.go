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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KogitoDataIndexSpec defines the desired state of KogitoDataIndex
// +k8s:openapi-gen=true
type KogitoDataIndexSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// Replicas is the number of pod replicas that the Data Index Service will spin
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
	// MemoryLimit is the limit which will be available for the container
	MemoryLimit string `json:"memoryLimit,omitempty"`

	// +optional
	// MemoryRequest is the request which will be requested upon container creation
	MemoryRequest string `json:"memoryRequest,omitempty"`

	// +optional
	// CPULimit is the limit which will be available for the container
	CPULimit string `json:"cpuLimit,omitempty"`

	// +optional
	// CPURequest is the request which will be requested upon container creation
	CPURequest string `json:"cpuRequest,omitempty"`

	// +optional
	// Infinispan has the data used by the Kogito Data Index to connect to the Infinispan cluster
	Infinispan InfinispanConnectionProperties `json:"infinispan,omitempty"`

	// +optional
	// Kafka has the data used by the Kogito Data Index to connecto to a Kafka cluster
	Kafka KafkaConnectionProperties `json:"kafka,omitempty"`
}

// KafkaConnectionProperties has the data needed to connect to a Kafka cluster
type KafkaConnectionProperties struct {
	// +optional
	// ServiceURI is the service URI to connect to the Infinispan cluster, e.g. myinifisan-cluster:11222
	ServiceURI string `json:"serviceURI,omitempty"`
}

// InfinispanConnectionProperties is the configuration needed for authenticating on Infinispan cluster.
// If not set, Data Index will connect to an existing Infinispan deployed by KogitoInfra resource and will set these values for you
// More information can be found at https://docs.jboss.org/infinispan/10.0/apidocs/org/infinispan/client/hotrod/configuration/package-summary.html#package.description
// +k8s:openapi-gen=true
type InfinispanConnectionProperties struct {
	// +optional
	Credentials SecretCredentialsType `json:"credentials,omitempty"`

	// UseAuth will be set to true if the credentials are set. Will set the property infinispan.client.hotrod.use_auth
	// +optional
	UseAuth bool `json:"useAuth,omitempty"`

	// Name of the Infinispan authentication realm. Will set the property infinispan.client.hotrod.auth_realm
	// +optional
	AuthRealm string `json:"authRealm,omitempty"`

	// +kubebuilder:validation:Enum=PLAIN;DIGEST-MD5
	// +optional
	// SaslMechanism defined for the authentication. Will set the property infinispan.client.hotrod.sasl_mechanism
	SaslMechanism InfinispanSaslMechanismType `json:"saslMechanism,omitempty"`

	// +optional
	// ServiceURI is the service URI to connect to the Infinispan cluster, e.g. myinfinispan-cluster:11222
	ServiceURI string `json:"serviceURI,omitempty"`

	// +optional
	// UseKogitoInfra flags if this Data Index instance will use a provided infrastructure by KogitoInfra CR. Defaults to true.
	// Set this to false and fill all other properties to provide your own infrastructure
	UseKogitoInfra bool `json:"useKogitoInfra,omitempty"`
}

// InfinispanSaslMechanismType is the possible SASL Mechanism used during infinispan connection. See: https://en.wikipedia.org/wiki/Simple_Authentication_and_Security_Layer#SASL_mechanisms
type InfinispanSaslMechanismType string

const (
	// SASLPlain is the PLAIN type
	SASLPlain InfinispanSaslMechanismType = "PLAIN"
	// SASLDigestMD5 is the DIGEST-MD5 type
	SASLDigestMD5 InfinispanSaslMechanismType = "DIGEST-MD5"
)

// SecretCredentialsType is the data structure for specifyng credentials within a Secret
type SecretCredentialsType struct {
	// +optional
	// SecretName is the name of the secret where the credentials are set
	SecretName string `json:"secretName,omitempty"`

	// +optional
	// UsernameKey is the username info for the user which will be authenticated in the cluster. Will set the property infinispan.client.hotrod.auth_username
	UsernameKey string `json:"usernameKey,omitempty"`

	// +optional
	// PasswordKey is the password info for the Infinispan user. Will set the propery infinispan.client.hotrod.auth_password
	PasswordKey string `json:"passwordKey,omitempty"`
}

// KogitoDataIndexStatus defines the observed state of KogitoDataIndex
// +k8s:openapi-gen=true
type KogitoDataIndexStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
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

// DataIndexDependenciesStatus is all possible status that the dependencies can have
type DataIndexDependenciesStatus string

const (
	//DataIndexDependenciesStatusOK means that all dependencies have been met
	DataIndexDependenciesStatusOK DataIndexDependenciesStatus = "OK"
	//DataIndexDependenciesStatusMissingKafka means that Kafka is missing
	DataIndexDependenciesStatusMissingKafka DataIndexDependenciesStatus = "Missing Kafka"
	//DataIndexDependenciesStatusMissingInfinispan means that Infinispan is missing
	DataIndexDependenciesStatusMissingInfinispan DataIndexDependenciesStatus = "Missing Infinispan"
)

// DataIndexCondition possible conditions for the Data Index Service
type DataIndexCondition struct {
	Condition          DataIndexConditionType `json:"condition"`
	Message            string                 `json:"message,omitempty"`
	LastTransitionTime metav1.Time            `json:"lastTransitionTime,omitempty"`
}

// DataIndexConditionType possible status that the resource can be
type DataIndexConditionType string

const (
	// ConditionOK means that everything was created successfully
	ConditionOK DataIndexConditionType = "OK"
	// ConditionProvisioning means that we're still deploying the service
	ConditionProvisioning DataIndexConditionType = "Provisioning"
	// ConditionFailed means that we fail to deploy the service and its dependencies
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
