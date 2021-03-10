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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// KogitoInfraConditionType ...
type KogitoInfraConditionType string

const (
	// SuccessInfraConditionType ...
	SuccessInfraConditionType KogitoInfraConditionType = "Success"
	// FailureInfraConditionType ...
	FailureInfraConditionType KogitoInfraConditionType = "Failure"
)

// KogitoInfraConditionReason describes the reasons for reconciliation failure
type KogitoInfraConditionReason string

const (
	// ReconciliationFailure generic failure on reconciliation
	ReconciliationFailure KogitoInfraConditionReason = "ReconciliationFailure"
	// ResourceNotFound target resource not found
	ResourceNotFound KogitoInfraConditionReason = "ResourceNotFound"
	// ResourceAPINotFound API not available in the cluster
	ResourceAPINotFound KogitoInfraConditionReason = "ResourceAPINotFound"
	// UnsupportedAPIKind API defined in the KogitoInfra CR not supported
	UnsupportedAPIKind KogitoInfraConditionReason = "UnsupportedAPIKind"
	// ResourceNotReady related resource is not ready
	ResourceNotReady KogitoInfraConditionReason = "ResourceNotReady"
	// ResourceConfigError related resource is not configured properly
	ResourceConfigError KogitoInfraConditionReason = "ResourceConfigError"
	// ResourceMissingResourceConfig related resource is missing a config information to continue
	ResourceMissingResourceConfig KogitoInfraConditionReason = "ResourceMissingConfig"
)

// KogitoInfraInterface ...
type KogitoInfraInterface interface {
	metav1.Object
	runtime.Object
	// GetSpec gets the Kogito Service specification structure.
	GetSpec() KogitoInfraSpecInterface
	// GetStatus gets the Kogito Service Status structure.
	GetStatus() KogitoInfraStatusInterface
}

// KogitoInfraSpecInterface ...
type KogitoInfraSpecInterface interface {
	GetResource() ResourceInterface
	GetInfraProperties() map[string]string
}

// ResourceInterface ...
type ResourceInterface interface {
	GetAPIVersion() string
	SetAPIVersion(apiVersion string)
	GetKind() string
	SetKind(kind string)
	GetNamespace() string
	SetNamespace(namespace string)
	GetName() string
	SetName(name string)
}

// KogitoInfraStatusInterface ...
type KogitoInfraStatusInterface interface {
	GetCondition() KogitoInfraConditionInterface
	SetCondition(condition KogitoInfraConditionInterface)
	GetRuntimeProperties() RuntimePropertiesMap
	AddRuntimeProperties(runtimeType RuntimeType, runtimeProperties RuntimePropertiesInterface)
	GetVolumes() []KogitoInfraVolumeInterface
	SetVolumes(infraVolumes []KogitoInfraVolumeInterface)
}

// RuntimePropertiesMap defines the map that KogitoInfraStatus
// will use to link the runtime to their variables.
type RuntimePropertiesMap map[RuntimeType]RuntimePropertiesInterface

// KogitoInfraConditionInterface ...
type KogitoInfraConditionInterface interface {
	GetType() KogitoInfraConditionType
	SetType(infraConditionType KogitoInfraConditionType)
	GetStatus() v1.ConditionStatus
	SetStatus(status v1.ConditionStatus)
	GetLastTransitionTime() metav1.Time
	SetLastTransitionTime(lastTransitionTime metav1.Time)
	GetMessage() string
	SetMessage(message string)
	GetReason() KogitoInfraConditionReason
	SetReason(reason KogitoInfraConditionReason)
}

// ConfigVolumeSourceInterface ...
type ConfigVolumeSourceInterface interface {
	GetSecret() *v1.SecretVolumeSource
	SetSecret(secret *v1.SecretVolumeSource)
	GetConfigMap() *v1.ConfigMapVolumeSource
	SetConfigMap(configMap *v1.ConfigMapVolumeSource)
}

// ConfigVolumeInterface ...
type ConfigVolumeInterface interface {
	ConfigVolumeSourceInterface
	GetName() string
	SetName(name string)
	ToKubeVolume() v1.Volume
}

// KogitoInfraVolumeInterface ...
type KogitoInfraVolumeInterface interface {
	GetMount() v1.VolumeMount
	GetNamedVolume() ConfigVolumeInterface
}

// RuntimePropertiesInterface ...
type RuntimePropertiesInterface interface {
	GetAppProps() map[string]string
	GetEnv() []v1.EnvVar
}
