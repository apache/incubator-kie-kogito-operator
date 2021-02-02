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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// KogitoSupportingServiceInterface ...
// +kubebuilder:object:generate=false
type KogitoSupportingServiceInterface interface {
	KogitoService
	// GetSpec gets the Kogito Service specification structure.
	GetSupportingServiceSpec() KogitoSupportingServiceSpecInterface
	// GetStatus gets the Kogito Service Status structure.
	GetSupportingServiceStatus() KogitoSupportingServiceStatusInterface
}

// KogitoSupportingServiceSpecInterface ...
// +kubebuilder:object:generate=false
type KogitoSupportingServiceSpecInterface interface {
	KogitoServiceSpecInterface
	GetServiceType() ServiceType
	SetServiceType(serviceType ServiceType)
}

// KogitoSupportingServiceStatusInterface ...
// +kubebuilder:object:generate=false
type KogitoSupportingServiceStatusInterface interface {
	KogitoServiceStatusInterface
}

// KogitoSupportingServiceListInterface ...
// +kubebuilder:object:generate=false
type KogitoSupportingServiceListInterface interface {
	runtime.Object
	// GetItems gets all items
	GetItems() []KogitoSupportingServiceInterface
}

// KogitoSupportingServiceHandler ...
// +kubebuilder:object:generate=false
type KogitoSupportingServiceHandler interface {
	FetchKogitoSupportingService(key types.NamespacedName) (KogitoSupportingServiceInterface, error)
	FetchKogitoSupportingServiceList(namespace string) (KogitoSupportingServiceListInterface, error)
}

// ServiceType define resource type of supporting service
type ServiceType string

const (
	// DataIndex supporting service resource type
	DataIndex ServiceType = "DataIndex"
	// Explainability supporting service resource type
	Explainability ServiceType = "Explainability"
	// JobsService supporting service resource type
	JobsService ServiceType = "JobsService"
	// MgmtConsole supporting service resource type
	MgmtConsole ServiceType = "MgmtConsole"
	// TaskConsole supporting service resource type
	TaskConsole ServiceType = "TaskConsole"
	// TrustyAI supporting service resource type
	TrustyAI ServiceType = "TrustyAI"
	// TrustyUI supporting service resource type
	TrustyUI ServiceType = "TrustyUI"
)
