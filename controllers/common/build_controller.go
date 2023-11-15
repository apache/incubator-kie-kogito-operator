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

package common

import (
	"context"
	"fmt"
	"github.com/kiegroup/kogito-operator/core/framework"
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	"github.com/kiegroup/kogito-operator/core/kogitobuild"
	"github.com/kiegroup/kogito-operator/core/logger"
	"github.com/kiegroup/kogito-operator/core/manager"
	"github.com/kiegroup/kogito-operator/core/operator"
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	kogitocli "github.com/kiegroup/kogito-operator/core/client"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	imageStreamCreationReconcileTimeout = 10 * time.Second
)

// KogitoBuildReconciler reconciles a KogitoBuild object
type KogitoBuildReconciler struct {
	*kogitocli.Client
	Scheme            *runtime.Scheme
	Version           string
	BuildHandler      func(context operator.Context) manager.KogitoBuildHandler
	ReconcilingObject client.Object
	Labels            map[string]string
}

// Reconcile reads that state of the cluster for a KogitoBuild object and makes changes based on the state read
// and what is in the KogitoBuild.Spec
func (r *KogitoBuildReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, resultErr error) {
	log := logger.FromContext(ctx)
	log.Info("Reconciling for KogitoBuild")

	// create buildContext
	buildContext := operator.Context{
		Client:  r.Client,
		Log:     log,
		Scheme:  r.Scheme,
		Version: r.Version,
		Labels:  r.Labels,
	}

	// fetch the requested instance
	buildHandler := r.BuildHandler(buildContext)
	instance, resultErr := buildHandler.FetchKogitoBuildInstance(req.NamespacedName)
	if resultErr != nil {
		return
	} else if instance == nil {
		log.Warn("Kogito Build not found")
		return
	}

	buildStatusHandler := kogitobuild.NewStatusHandler(buildContext, buildHandler)
	defer buildStatusHandler.HandleStatusChange(instance, resultErr)

	if len(instance.GetSpec().GetRuntime()) == 0 {
		instance.GetSpec().SetRuntime(api.QuarkusRuntimeType)
	}
	envs := instance.GetSpec().GetEnv()
	instance.GetSpec().SetEnv(framework.EnvOverride(envs, corev1.EnvVar{Name: infrastructure.RuntimeTypeKey, Value: string(instance.GetSpec().GetRuntime())}))
	if len(instance.GetSpec().GetTargetKogitoRuntime()) == 0 {
		instance.GetSpec().SetTargetKogitoRuntime(instance.GetName())
	}

	// create the Kogito Image Streams to build the service if needed
	buildImageHandler := kogitobuild.NewImageSteamHandler(buildContext)
	created, resultErr := buildImageHandler.CreateRequiredKogitoImageStreams(instance)
	if resultErr != nil {
		return result, fmt.Errorf("Error while creating Kogito ImageStreams: %s ", resultErr)
	}
	if created {
		result = reconcile.Result{RequeueAfter: imageStreamCreationReconcileTimeout, Requeue: true}
		return
	}

	// get the build manager to start the reconciliation logic
	deltaProcessor, resultErr := kogitobuild.NewDeltaProcessor(buildContext, instance, buildHandler)
	if resultErr != nil {
		return
	}
	resultErr = deltaProcessor.ProcessDelta()
	return
}

// SetupWithManager registers the controller with manager
func (r *KogitoBuildReconciler) SetupWithManager(mgr ctrl.Manager) error {
	b := ctrl.NewControllerManagedBy(mgr).For(r.ReconcilingObject)
	if r.IsOpenshift() {
		b.Owns(&buildv1.BuildConfig{}).Owns(&imagev1.ImageStream{})
	}
	return b.Complete(r)
}
