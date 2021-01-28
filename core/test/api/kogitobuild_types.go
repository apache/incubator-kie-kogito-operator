package api

import (
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KogitoBuildTest ..
// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type KogitoBuildTest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   api.KogitoBuildSpec   `json:"spec,omitempty"`
	Status api.KogitoBuildStatus `json:"status,omitempty"`
}

// GetSpec ..
func (k *KogitoBuildTest) GetSpec() api.KogitoBuildSpecInterface {
	return &k.Spec
}

// GetStatus ...
func (k *KogitoBuildTest) GetStatus() api.KogitoBuildStatusInterface {
	return &k.Status
}

func init() {
	SchemeBuilder.Register(&KogitoBuildTest{})
}
