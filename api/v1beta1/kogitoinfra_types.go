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

package v1beta1

import (
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true

// KogitoInfra is the resource to bind a Custom Resource (CR) not managed by Kogito Operator to a given deployed Kogito service.
// It holds the reference of a CR managed by another operator such as Strimzi. For example: one can create a Kafka CR via Strimzi
// and link this resource using KogitoInfra to a given Kogito service (custom or supporting, such as Data Index).
// Please refer to the Kogito Operator documentation (https://docs.jboss.org/kogito/release/latest/html_single/) for more information.
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=kogitoinfras,scope=Namespaced
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Resource Name",type="string",JSONPath=".spec.resource.name",description="Third Party Infrastructure Resource"
// +kubebuilder:printcolumn:name="Kind",type="string",JSONPath=".spec.resource.kind",description="Kubernetes CR Kind"
// +kubebuilder:printcolumn:name="API Version",type="string",JSONPath=".spec.resource.apiVersion",description="Kubernetes CR API Version"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.condition.status",description="General Status of this resource bind"
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=".status.condition.reason",description="Status reason"
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Kogito Infra"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Kafka,kafka.strimzi.io/v1beta1,\"A Kafka instance\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Infinispan,infinispan.org/v1,\"A Infinispan instance\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Keycloak,keycloak.org/v1alpha1,\"A Keycloak Instance\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Secret,v1,\"A Kubernetes Secret\""
type KogitoInfra struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   api.KogitoInfraSpec   `json:"spec,omitempty"`
	Status api.KogitoInfraStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KogitoInfraList contains a list of KogitoInfra.
type KogitoInfraList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KogitoInfra `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KogitoInfra{}, &KogitoInfraList{})
}

// GetSpec provide spec of Kogito infra
func (k *KogitoInfra) GetSpec() api.KogitoInfraSpecInterface {
	return &k.Spec
}

// GetStatus provide status of Kogito infra
func (k *KogitoInfra) GetStatus() api.KogitoInfraStatusInterface {
	return &k.Status
}
