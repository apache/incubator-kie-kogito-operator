// Copyright 2020 Red Hat, Inc. and/or its affiliates
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

package rhpam

import (
	v1 "github.com/kiegroup/kogito-operator/apis/rhpam/v1"
	"github.com/kiegroup/kogito-operator/controllers/common"
	"github.com/kiegroup/kogito-operator/internal/rhpam"
	rhpam2 "github.com/kiegroup/kogito-operator/version/rhpam"

	kogitocli "github.com/kiegroup/kogito-operator/core/client"
	"k8s.io/apimachinery/pkg/runtime"
)

//+kubebuilder:rbac:groups=rhpam.kiegroup.org,resources=kogitoinfras,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rhpam.kiegroup.org,resources=kogitoinfras/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rhpam.kiegroup.org,resources=kogitoinfras/finalizers,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments;replicasets,verbs=get;create;list;watch;create;delete;update
//+kubebuilder:rbac:groups=infinispan.org,resources=infinispans,verbs=get;create;list;delete;watch
//+kubebuilder:rbac:groups=kafka.strimzi.io,resources=kafkas;kafkatopics,verbs=get;create;list;delete;watch
//+kubebuilder:rbac:groups=keycloak.org,resources=keycloaks,verbs=get;create;list;delete;watch
//+kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update
//+kubebuilder:rbac:groups=eventing.knative.dev,resources=brokers,verbs=get;list;watch
//+kubebuilder:rbac:groups=eventing.knative.dev,resources=triggers,verbs=get;list;watch;create;delete;update
//+kubebuilder:rbac:groups=sources.knative.dev,resources=sinkbindings,verbs=get;list;watch;create;delete;update
//+kubebuilder:rbac:groups=integreatly.org,resources=grafanadashboards,verbs=get;create;list;watch;create;delete;update
//+kubebuilder:rbac:groups=mongodbcommunity.mongodb.com,resources=mongodb,verbs=get;create;list;watch;delete

// NewKogitoInfraReconciler ...
func NewKogitoInfraReconciler(client *kogitocli.Client, scheme *runtime.Scheme) *common.KogitoInfraReconciler {
	return &common.KogitoInfraReconciler{
		Client:            client,
		Scheme:            scheme,
		Version:           rhpam2.Version,
		InfraHandler:      rhpam.NewKogitoInfraHandler,
		ReconcilingObject: &v1.KogitoInfra{},
	}
}
