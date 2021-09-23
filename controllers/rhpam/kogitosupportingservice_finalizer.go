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
	"github.com/kiegroup/kogito-operator/controllers/common"
	kogitocli "github.com/kiegroup/kogito-operator/core/client"
	"github.com/kiegroup/kogito-operator/internal/rhpam"
	rhpam2 "github.com/kiegroup/kogito-operator/version/rhpam"
	"k8s.io/apimachinery/pkg/runtime"
)

// NewFinalizeKogitoSupportingServiceReconciler ...
func NewFinalizeKogitoSupportingServiceReconciler(client *kogitocli.Client, scheme *runtime.Scheme) *common.FinalizeKogitoSupportingServiceReconciler {
	return &common.FinalizeKogitoSupportingServiceReconciler{
		Client:                   client,
		Scheme:                   scheme,
		Version:                  rhpam2.Version,
		SupportingServiceHandler: rhpam.NewKogitoSupportingServiceHandler,
		InfraHandler:             rhpam.NewKogitoInfraHandler,
	}
}
