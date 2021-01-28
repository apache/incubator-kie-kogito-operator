package api

import (
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KogitoRuntimeTest ...
// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type KogitoRuntimeTest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KogitoRuntimeSpecTest   `json:"spec,omitempty"`
	Status KogitoRuntimeStatusTest `json:"status,omitempty"`
}

// KogitoRuntimeSpecTest ...
type KogitoRuntimeSpecTest struct {
	api.KogitoServiceSpec `json:",inline"`
	EnableIstio           bool            `json:"enableIstio,omitempty"`
	Runtime               api.RuntimeType `json:"runtime,omitempty"`
}

// IsEnableIstio ...
func (k *KogitoRuntimeSpecTest) IsEnableIstio() bool {
	return k.EnableIstio
}

// SetEnableIstio ...
func (k *KogitoRuntimeSpecTest) SetEnableIstio(enableIstio bool) {
	k.EnableIstio = enableIstio
}

// GetRuntime ...
func (k *KogitoRuntimeSpecTest) GetRuntime() api.RuntimeType {
	return k.Runtime
}

// KogitoRuntimeStatusTest ...
type KogitoRuntimeStatusTest struct {
	api.KogitoServiceStatus `json:",inline"`
}

// GetSpec ...
func (k *KogitoRuntimeTest) GetSpec() api.KogitoServiceSpecInterface {
	return &k.Spec
}

// GetStatus ...
func (k *KogitoRuntimeTest) GetStatus() api.KogitoServiceStatusInterface {
	return &k.Status
}

// GetRuntimeSpec ...
func (k *KogitoRuntimeTest) GetRuntimeSpec() api.KogitoRuntimeSpecInterface {
	return &k.Spec
}

// GetRuntimeStatus ...
func (k *KogitoRuntimeTest) GetRuntimeStatus() api.KogitoRuntimeStatusInterface {
	return &k.Status
}

// KogitoRuntimeTestList ...
// +kubebuilder:object:root=true
// KogitoRuntimeList contains a list of KogitoRuntime.
type KogitoRuntimeTestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KogitoRuntimeTest `json:"items"`
}

// GetItems ...
func (k *KogitoRuntimeTestList) GetItems() []api.KogitoRuntimeInterface {
	models := make([]api.KogitoRuntimeInterface, len(k.Items))
	for i, v := range k.Items {
		models[i] = &v
	}
	return models
}

func init() {
	SchemeBuilder.Register(&KogitoRuntimeTest{}, &KogitoRuntimeTestList{})
}
