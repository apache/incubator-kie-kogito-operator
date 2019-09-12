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

package builder

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
)

var log = logger.GetLogger("builder_kogitoapp")

// KogitoAppInventory has a reference for every resource needed to deploy the KogitoApp
type KogitoAppInventory struct {
	ResourceInventoryStatus
	BuildConfigS2I     *buildv1.BuildConfig
	BuildConfigService *buildv1.BuildConfig
	DeploymentConfig   *appsv1.DeploymentConfig
	Route              *routev1.Route
	Service            *corev1.Service
}

// ResourceInventoryStatusKind defines the kind of the resource status in the cluster
type ResourceInventoryStatusKind struct {
	IsNew bool
}

// ResourceInventoryStatus defines the resource status in the cluster
type ResourceInventoryStatus struct {
	BuildConfigS2IStatus     ResourceInventoryStatusKind
	BuildConfigServiceStatus ResourceInventoryStatusKind
	DeploymentConfigStatus   ResourceInventoryStatusKind
	RouteStatus              ResourceInventoryStatusKind
	ServiceStatus            ResourceInventoryStatusKind
}
