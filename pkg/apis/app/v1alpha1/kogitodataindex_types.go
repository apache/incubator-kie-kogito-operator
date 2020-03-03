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

// KogitoDataIndexCRDName is the name of the Kogito Data Index CRD in the cluster
const KogitoDataIndexCRDName = "kogitodataindices.app.kiegroup.org"

// KogitoDataIndexSpec defines the desired state of KogitoDataIndex
// +k8s:openapi-gen=true
type KogitoDataIndexSpec struct {
	InfinispanMeta    `json:",inline"`
	KafkaMeta         `json:",inline"`
	KogitoServiceSpec `json:",inline"`
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// HttpPort will set the environment env KOGITO_DATA_INDEX_HTTP_PORT to define which port data-index service will listen internally.
	// +optional
	HTTPPort int32 `json:"httpPort,omitempty"`
}

// KogitoDataIndexStatus defines the observed state of KogitoDataIndex
// +k8s:openapi-gen=true
type KogitoDataIndexStatus struct {
	KogitoServiceStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KogitoDataIndex defines the Data Index Service infrastructure deployment
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

// GetSpec ...
func (k *KogitoDataIndex) GetSpec() KogitoServiceSpecInterface {
	return &k.Spec
}

// GetStatus ...
func (k *KogitoDataIndex) GetStatus() KogitoServiceStatusInterface {
	return &k.Status
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KogitoDataIndexList contains a list of KogitoDataIndex
type KogitoDataIndexList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// +listType=atomic
	Items []KogitoDataIndex `json:"items"`
}

// GetItemsCount ...
func (l *KogitoDataIndexList) GetItemsCount() int {
	return len(l.Items)
}

// GetItemAt ...
func (l *KogitoDataIndexList) GetItemAt(index int) KogitoService {
	if len(l.Items) > index {
		return KogitoService(&l.Items[index])
	}
	return nil
}

func init() {
	SchemeBuilder.Register(&KogitoDataIndex{}, &KogitoDataIndexList{})
}
