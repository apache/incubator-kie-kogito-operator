/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package app

import (
	"github.com/kiegroup/kogito-operator/apis/app/v1beta1"
	"github.com/kiegroup/kogito-operator/controllers/common"
	"github.com/kiegroup/kogito-operator/core/client"
	"github.com/kiegroup/kogito-operator/internal/app"
	app2 "github.com/kiegroup/kogito-operator/version/app"
	"k8s.io/apimachinery/pkg/runtime"
)

//+kubebuilder:rbac:groups=app.kiegroup.org,resources=kogitobuilds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.kiegroup.org,resources=kogitobuilds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.kiegroup.org,resources=kogitobuilds/finalizers,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments;replicasets,verbs=get;create;list;watch;delete;update
//+kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update
//+kubebuilder:rbac:groups=build.openshift.io,resources=builds;buildconfigs,verbs=get;create;list;watch;delete;update
//+kubebuilder:rbac:groups=image.openshift.io,resources=imagestreams;imagestreamtags,verbs=get;create;list;watch;delete;update

// NewKogitoBuildReconciler ...
func NewKogitoBuildReconciler(client *client.Client, scheme *runtime.Scheme) *common.KogitoBuildReconciler {
	return &common.KogitoBuildReconciler{
		Client:            client,
		Scheme:            scheme,
		Version:           app2.Version,
		BuildHandler:      app.NewKogitoBuildHandler,
		ReconcilingObject: &v1beta1.KogitoBuild{},
	}
}
