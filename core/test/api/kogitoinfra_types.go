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
