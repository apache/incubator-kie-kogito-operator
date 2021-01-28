package api

import (
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KogitoSupportingServiceTest ...
// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type KogitoSupportingServiceTest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KogitoSupportingServiceSpecTest   `json:"spec,omitempty"`
	Status KogitoSupportingServiceStatusTest `json:"status,omitempty"`
}

// KogitoSupportingServiceSpecTest ...
type KogitoSupportingServiceSpecTest struct {
	api.KogitoServiceSpec `json:",inline"`
	ServiceType           api.ServiceType `json:"serviceType"`
}

// GetServiceType ...
func (k *KogitoSupportingServiceSpecTest) GetServiceType() api.ServiceType {
	return k.ServiceType
}

// SetServiceType ...
func (k *KogitoSupportingServiceSpecTest) SetServiceType(serviceType api.ServiceType) {
	k.ServiceType = serviceType
}

// GetRuntime ...
func (k *KogitoSupportingServiceSpecTest) GetRuntime() api.RuntimeType {
	return api.QuarkusRuntimeType
}

// KogitoSupportingServiceStatusTest ...
type KogitoSupportingServiceStatusTest struct {
	api.KogitoServiceStatus `json:",inline"`
}

// GetSpec ...
func (k *KogitoSupportingServiceTest) GetSpec() api.KogitoServiceSpecInterface {
	return &k.Spec
}

// GetStatus ...
func (k *KogitoSupportingServiceTest) GetStatus() api.KogitoServiceStatusInterface {
	return &k.Status
}

// GetSupportingServiceSpec ...
func (k *KogitoSupportingServiceTest) GetSupportingServiceSpec() api.KogitoSupportingServiceSpecInterface {
	return &k.Spec
}

// GetSupportingServiceStatus ...
func (k *KogitoSupportingServiceTest) GetSupportingServiceStatus() api.KogitoSupportingServiceStatusInterface {
	return &k.Status
}

// KogitoSupportingServiceTestList ...
// +kubebuilder:object:root=true
// KogitoSupportingServiceList contains a list of KogitoSupportingService.
type KogitoSupportingServiceTestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KogitoSupportingServiceTest `json:"items"`
}

// GetItems ...
func (k *KogitoSupportingServiceTestList) GetItems() []api.KogitoSupportingServiceInterface {
	models := make([]api.KogitoSupportingServiceInterface, len(k.Items))
	for i, v := range k.Items {
		models[i] = &v
	}
	return models
}

func init() {
	SchemeBuilder.Register(&KogitoSupportingServiceTest{}, &KogitoSupportingServiceTestList{})
}
