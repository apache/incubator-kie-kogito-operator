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
	"github.com/kiegroup/kogito-operator/apis/app/v1beta1"
	"github.com/kiegroup/kogito-operator/controllers/common"
	kogitocli "github.com/kiegroup/kogito-operator/core/client"
	app2 "github.com/kiegroup/kogito-operator/internal/app"
	"github.com/kiegroup/kogito-operator/version/app"
	"k8s.io/apimachinery/pkg/runtime"
)

// NewFinalizeKogitoRuntimeReconciler ...
func NewFinalizeKogitoRuntimeReconciler(client *kogitocli.Client, scheme *runtime.Scheme) *common.FinalizeKogitoRuntimeReconciler {
	return &common.FinalizeKogitoRuntimeReconciler{
		Client:            client,
		Scheme:            scheme,
		Version:           app.Version,
		RuntimeHandler:    app2.NewKogitoRuntimeHandler,
		InfraHandler:      app2.NewKogitoInfraHandler,
		ReconcilingObject: &v1beta1.KogitoRuntime{},
	}
}
