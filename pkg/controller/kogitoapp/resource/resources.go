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

package resource

import (
	monv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	dockerv10 "github.com/openshift/api/image/docker10"
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
)

var log = logger.GetLogger("builder_kogitoapp")

// KogitoAppResources has a reference for every resource needed to deploy the KogitoApp
type KogitoAppResources struct {
	KogitoAppResourcesStatus
	BuildConfigS2I     *buildv1.BuildConfig
	BuildConfigRuntime *buildv1.BuildConfig
	ImageStreamS2I     *imgv1.ImageStream
	ImageStreamRuntime *imgv1.ImageStream
	DeploymentConfig   *appsv1.DeploymentConfig
	Route              *routev1.Route
	Service            *corev1.Service
	ServiceMonitor     *monv1.ServiceMonitor
	RuntimeImage       *dockerv10.DockerImage
}

// KogitoAppResourceStatusKind defines the kind of the resource status in the cluster
type KogitoAppResourceStatusKind struct {
	IsNew bool
}

// KogitoAppResourcesStatus defines the resource status in the cluster
type KogitoAppResourcesStatus struct {
	BuildConfigS2IStatus     KogitoAppResourceStatusKind
	BuildConfigRuntimeStatus KogitoAppResourceStatusKind
	DeploymentConfigStatus   KogitoAppResourceStatusKind
	RouteStatus              KogitoAppResourceStatusKind
	ServiceStatus            KogitoAppResourceStatusKind
}
