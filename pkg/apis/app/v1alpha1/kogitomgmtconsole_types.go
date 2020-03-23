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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KogitoMgmtConsoleSpec defines the desired state of KogitoMgmtConsole
type KogitoMgmtConsoleSpec struct {
	KogitoServiceSpec `json:",inline"`
}

// KogitoMgmtConsoleStatus defines the observed state of KogitoMgmtConsole
type KogitoMgmtConsoleStatus struct {
	KogitoServiceStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KogitoMgmtConsole deploys the Kogito Management Console service in the given namespace
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=kogitomgmtconsoles,scope=Namespaced
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Kogito Management Console"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Deployments,apps/v1"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Services,v1"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="ImageStreams,image.openshift.io/v1"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Routes,route.openshift.io/v1"
type KogitoMgmtConsole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KogitoMgmtConsoleSpec   `json:"spec,omitempty"`
	Status KogitoMgmtConsoleStatus `json:"status,omitempty"`
}

// GetSpec ...
func (k *KogitoMgmtConsole) GetSpec() KogitoServiceSpecInterface {
	return &k.Spec
}

// GetStatus ...
func (k *KogitoMgmtConsole) GetStatus() KogitoServiceStatusInterface {
	return &k.Status
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KogitoMgmtConsoleList contains a list of KogitoMgmtConsole
type KogitoMgmtConsoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KogitoMgmtConsole `json:"items"`
}

// GetItemsCount ...
func (l *KogitoMgmtConsoleList) GetItemsCount() int {
	return len(l.Items)
}

// GetItemAt ...
func (l *KogitoMgmtConsoleList) GetItemAt(index int) KogitoService {
	if len(l.Items) > index {
		return KogitoService(&l.Items[index])
	}
	return nil
}

func init() {
	SchemeBuilder.Register(&KogitoMgmtConsole{}, &KogitoMgmtConsoleList{})
}
