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

package app

import (
	"github.com/kiegroup/kogito-operator/controllers/common"
	kogitocli "github.com/kiegroup/kogito-operator/core/client"
	app2 "github.com/kiegroup/kogito-operator/internal/app"
	"k8s.io/apimachinery/pkg/runtime"
)

//+kubebuilder:rbac:groups=apps,resources=deployments;replicasets,verbs=get;create;list;watch;delete;update
//+kubebuilder:rbac:groups=core,resources=configmaps;events;pods;secrets;serviceaccounts;services,verbs=create;delete;get;list;patch;update;watch

// NewKogitoRuntimeDeploymentReconciler ...
func NewKogitoRuntimeDeploymentReconciler(client *kogitocli.Client, scheme *runtime.Scheme) *common.RuntimeDeploymentReconciler {
	return &common.RuntimeDeploymentReconciler{
		Client:                client,
		Scheme:                scheme,
		RuntimeHandler:        app2.NewKogitoRuntimeHandler,
		SupportServiceHandler: app2.NewKogitoSupportingServiceHandler,
	}
}
