package api

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// KogitoRuntimeInterface ...
// +kubebuilder:object:generate=false
type KogitoRuntimeInterface interface {
	KogitoService
	// GetSpec gets the Kogito Service specification structure.
	GetRuntimeSpec() KogitoRuntimeSpecInterface
	// GetStatus gets the Kogito Service Status structure.
	GetRuntimeStatus() KogitoRuntimeStatusInterface
}

// KogitoRuntimeListInterface ...
// +kubebuilder:object:generate=false
type KogitoRuntimeListInterface interface {
	runtime.Object
	// GetItems gets all items
	GetItems() []KogitoRuntimeInterface
}

// KogitoRuntimeSpecInterface ...
// +kubebuilder:object:generate=false
type KogitoRuntimeSpecInterface interface {
	KogitoServiceSpecInterface
	IsEnableIstio() bool
	SetEnableIstio(enableIstio bool)
}

// KogitoRuntimeStatusInterface ...
// +kubebuilder:object:generate=false
type KogitoRuntimeStatusInterface interface {
	KogitoServiceStatusInterface
}

// KogitoRuntimeHandler ...
// +kubebuilder:object:generate=false
type KogitoRuntimeHandler interface {
	FetchKogitoRuntimeInstance(key types.NamespacedName) (KogitoRuntimeInterface, error)
	FetchAllKogitoRuntimeInstances(namespace string) (KogitoRuntimeListInterface, error)
}
