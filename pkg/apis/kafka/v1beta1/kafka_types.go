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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KafkaSpec defines the desired state of Kafka
type KafkaSpec struct {
	KafkaClusterSpec `json:"kafka,omitempty"`
}

// KafkaClusterSpec defines the desired state of Kafka Cluster
type KafkaClusterSpec struct {
	Replicas int32 `json:"replicas,omitempty"`
}

// KafkaStatus defines the observed state of Kafka
type KafkaStatus struct {
	Listeners []ListenerStatus `json:"listeners,omitempty"`
}

// ListenerStatus defines a single listener
type ListenerStatus struct {
	Type      string            `json:"type,omitempty"`
	Addresses []ListenerAddress `json:"addresses,omitempty"`
}

// ListenerAddress defines a single address of particular listener
type ListenerAddress struct {
	Host string `json:"host,omitempty"`
	Port int32  `json:"port,omitempty"`
}

// Kafka is the Schema for the kafkas API
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Kafka struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaSpec   `json:"spec,omitempty"`
	Status KafkaStatus `json:"status,omitempty"`
}

// KafkaList contains a list of Kafka
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type KafkaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kafka `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Kafka{}, &KafkaList{})
}
