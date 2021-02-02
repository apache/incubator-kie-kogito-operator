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
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KogitoInfraTest ...
// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type KogitoInfraTest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   api.KogitoInfraSpec   `json:"spec,omitempty"`
	Status api.KogitoInfraStatus `json:"status,omitempty"`
}

// GetSpec ...
func (k *KogitoInfraTest) GetSpec() api.KogitoInfraSpecInterface {
	return &k.Spec
}

// GetStatus ...
func (k *KogitoInfraTest) GetStatus() api.KogitoInfraStatusInterface {
	return &k.Status
}

func init() {
	SchemeBuilder.Register(&KogitoInfraTest{})
}
