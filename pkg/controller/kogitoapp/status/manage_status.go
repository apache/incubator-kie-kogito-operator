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

package status

import (
	"context"
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/resource"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
	appsv1 "github.com/openshift/api/apps/v1"
	obuildv1 "github.com/openshift/api/build/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	cachev1 "sigs.k8s.io/controller-runtime/pkg/cache"
	"time"
)

var log = logger.GetLogger("kogitoapp.controller")

// UpdateResourcesResult contains the results of the update of the resources
type UpdateResourcesResult struct {
	Updated     bool
	ErrorReason v1alpha1.ReasonType
	Err         error
}

// UpdateStatusResult contains the results of the update of the status
type UpdateStatusResult struct {
	Updated      bool
	RequeueAfter time.Duration
	Err          error
}

func createResult(updated bool, requeueAfter time.Duration, err error, result *UpdateStatusResult) {
	if updated {
		result.Updated = true
	}

	if requeueAfter > 0 && result.RequeueAfter <= 0 {
		result.RequeueAfter = requeueAfter
	}

	if result.Err == nil && err != nil {
		result.Err = err
	}
}

// ManageStatus will guarantee the status changes
func ManageStatus(instance *v1alpha1.KogitoApp, resources *resource.KogitoAppResources, client *client.Client,
	resourcesUpdateResult *UpdateResourcesResult, cache cachev1.Cache, namespacedName types.NamespacedName) *UpdateStatusResult {
	result := &UpdateStatusResult{false, 0, nil}

	{
		requeueAfter, updated, err := statusUpdateForImageBuild(instance, resources, client)
		createResult(updated, requeueAfter, err, result)
	}

	{
		updated, err := statusUpdateForDeployment(instance, resources, client)
		createResult(updated, 0, err, result)
	}

	{
		requeueAfter, updated, err := statusUpdateForRoute(instance, resources, client)
		createResult(updated, requeueAfter, err, result)
	}

	{
		updated, err := statusUpdateForResources(instance, resourcesUpdateResult)
		createResult(updated, 0, err, result)
	}

	// Fetch the cached KogitoApp instance
	cachedInstance := &v1alpha1.KogitoApp{}
	{
		err := cache.Get(context.TODO(), namespacedName, cachedInstance)
		if err != nil {
			if errors.IsNotFound(err) {
				// Request object not found, could have been deleted after reconcile request.
				// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
				// Return and don't requeue

			} else {
				log.Infof("Failed to read the cached instance")
				// Error reading the object - requeue the request.
				instance.Status.SetFailed(v1alpha1.UnknownReason, err)
				createResult(true, 0, nil, result)
			}
		} else {
			updated := statusUpdateForKogitoApp(instance, cachedInstance)
			createResult(updated, 0, nil, result)
		}
	}

	if result.Updated {
		// UpdateObj reconciles the given object
		updated, err := updateObj(instance, client)
		createResult(updated, 0, err, result)
	} else if result.RequeueAfter <= 0 && result.Err == nil {
		if instance.Status.SetDeployed() {
			log.Infof("'%s' successfully deployed", instance.Name)
			if instance.ResourceVersion == cachedInstance.ResourceVersion {
				updated, err := updateObj(instance, client)
				createResult(updated, 0, err, result)
			}
		}
	}

	return result
}

func updateObj(obj meta.ResourceObject, client *client.Client) (bool, error) {
	log := log.With("kind", obj.GetObjectKind().GroupVersionKind().Kind, "name", obj.GetName(), "namespace", obj.GetNamespace())
	log.Info("Updating")
	err := kubernetes.ResourceC(client).Update(obj)
	if err != nil {
		log.Warn("Failed to update object. ", err)
		return false, err
	}
	// Object updated - return and requeue
	return true, nil
}

func statusUpdateForResources(instance *v1alpha1.KogitoApp, result *UpdateResourcesResult) (bool, error) {
	if result != nil {
		if result.Err != nil {
			instance.Status.SetFailed(result.ErrorReason, result.Err)
			return true, result.Err
		} else if result.Updated {
			instance.Status.SetProvisioning()
			return true, nil
		}
	}

	return false, nil
}

func statusUpdateForDeployment(instance *v1alpha1.KogitoApp, resources *resource.KogitoAppResources, client *client.Client) (bool, error) {
	if resources.DeploymentConfig != nil {
		dc := &appsv1.DeploymentConfig{}

		if exists, err :=
			kubernetes.ResourceC(client).FetchWithKey(
				types.NamespacedName{Name: resources.DeploymentConfig.Name, Namespace: resources.DeploymentConfig.Namespace}, dc); err != nil {
			return false, err
		} else if exists {
			var ready, starting, stopped []string

			if dc.Spec.Replicas == 0 {
				stopped = append(stopped, dc.Name)
			} else if dc.Status.Replicas == 0 {
				stopped = append(stopped, dc.Name)
			} else if dc.Status.ReadyReplicas < dc.Status.Replicas {
				starting = append(starting, dc.Name)
			} else {
				ready = append(ready, dc.Name)
			}

			log.Debugf("Found DC with status stopped [%s], starting [%s], and ready [%s]", stopped, starting, ready)

			if len(instance.Status.Deployments.Stopped) != len(stopped) ||
				len(instance.Status.Deployments.Starting) != len(starting) ||
				len(instance.Status.Deployments.Ready) != len(ready) {
				log.Infof("Updating deployment status")

				instance.Status.Deployments = v1alpha1.Deployments{
					Stopped:  stopped,
					Starting: starting,
					Ready:    ready,
				}

				return true, nil
			}
		}
	} else {
		log.Debugf("DeploymentConfig is nil, impossible to get the deployment status", resources.DeploymentConfig)
	}

	return false, nil
}

func statusUpdateForRoute(instance *v1alpha1.KogitoApp, resources *resource.KogitoAppResources, client *client.Client) (requeue time.Duration, updated bool, err error) {
	// Setting route to the status
	if resources.Route != nil {
		log.Debugf("Trying to get the host for the route %s", resources.Route.Name)
		if exists, route, err := openshift.RouteC(client).GetHostFromRoute(types.NamespacedName{Name: resources.Route.Name, Namespace: resources.Route.Namespace}); err != nil {
			return 0, false, err
		} else if exists {
			fmtRoute := fmt.Sprintf("http://%s", route)
			if fmtRoute != instance.Status.Route {
				log.Infof("Updating route status")
				instance.Status.Route = fmtRoute
				return 0, true, nil
			}

			return 0, false, nil
		}

		log.Infof("Failed to get the host of the route %s", resources.Route.Name)
		return time.Duration(500) * time.Millisecond, false, nil
	}

	log.Debugf("Route is nil, impossible to get host to set in the status", resources.Route)
	return 0, false, nil
}

func statusUpdateForImageBuild(instance *v1alpha1.KogitoApp, resources *resource.KogitoAppResources, client *client.Client) (
	requeue time.Duration, updated bool, err error) {
	// ensure builds
	log.Infof("Checking if build for '%s' is finished", instance.Name)
	var imageExists, building, runtimeFailed, s2iFailed bool
	if imageExists, building, updated, runtimeFailed, s2iFailed, err = ensureApplicationImageExists(instance, resources, client); err != nil {
		return 0, false, err
	}

	if building {
		// let's wait for the build to finish
		if instance.Status.SetProvisioning() {
			updated = true
		}

		requeue = time.Duration(50) * time.Second
	}

	if runtimeFailed {
		instance.Status.SetFailed(v1alpha1.BuildRuntimeFailedReason, fmt.Errorf("runtime image build failed"))
		updated = true
	}

	if s2iFailed {
		instance.Status.SetFailed(v1alpha1.BuildS2IFailedReason, fmt.Errorf("s2i image build failed"))
		updated = true
	}

	if !imageExists {
		log.Infof("Build for '%s' still running", instance.Name)
		requeue = time.Duration(50) * time.Second
	}

	return requeue, updated, nil
}

func ensureApplicationImageExists(instance *v1alpha1.KogitoApp, resources *resource.KogitoAppResources, client *client.Client) (
	exists bool, running bool, updated bool, runtimeFailed bool, s2iFailed bool, err error) {

	runtimeState, err :=
		openshift.BuildConfigC(client).EnsureImageBuild(
			resources.BuildConfigRuntime,
			getBCLabelsAsUniqueSelectors(resources.BuildConfigRuntime))
	if err != nil {
		return false, false, false, false, false, err
	}

	// verify service build and image
	if runtimeState.ImageExists {
		log.Debugf("Final application image exists there's no need to trigger any build")
	} else {
		log.Warnf("Image not found for %s. The image could being built. Check with 'oc get is/%s -n %s'",
			resources.BuildConfigRuntime.Name,
			resources.BuildConfigRuntime.Name,
			instance.Namespace)
	}

	var runtimeUpdated, runtimeRunning, runtimeError bool
	if runtimeState.Builds == nil {
		runtimeRunning = true
	} else {
		runtimeUpdated, runtimeRunning, runtimeError = checkBuildsStatus(runtimeState.Builds, &instance.Status.Builds)
	}

	if runtimeRunning {
		log.Infof("Image for '%s' is being pushed to the registry", instance.Name)
	}

	// verify s2i build and image
	s2iState, err :=
		openshift.BuildConfigC(client).EnsureImageBuild(
			resources.BuildConfigS2I,
			getBCLabelsAsUniqueSelectors(resources.BuildConfigS2I))
	if err != nil {
		return false, runtimeRunning, runtimeUpdated, runtimeError, false, err
	}

	var s2iUpdated, s2iRunning, s2iError bool
	if s2iState.Builds == nil {
		s2iRunning = true
	} else {
		s2iUpdated, s2iRunning, s2iError = checkBuildsStatus(s2iState.Builds, &instance.Status.Builds)
	}

	if s2iRunning {
		// build is running, nothing to do
		log.Infof("Application '%s' build is still running. Won't trigger a new build.", instance.Name)
	} else if !s2iState.ImageExists && !s2iRunning {
		log.Warnf("There's no image nor build for '%s' running", resources.BuildConfigS2I.Name)
	}

	if runtimeState.ImageExists && !runtimeRunning && s2iState.ImageExists && !s2iRunning {
		log.Debugf("There are images for both builds, nothing to do")
	}

	if runtimeUpdated || s2iUpdated {
		instance.Status.Builds.New = append(runtimeState.Builds.New, s2iState.Builds.New...)
		instance.Status.Builds.Pending = append(runtimeState.Builds.Pending, s2iState.Builds.Pending...)
		instance.Status.Builds.Running = append(runtimeState.Builds.Running, s2iState.Builds.Running...)
		instance.Status.Builds.Error = append(runtimeState.Builds.Error, s2iState.Builds.Error...)
		instance.Status.Builds.Failed = append(runtimeState.Builds.Failed, s2iState.Builds.Failed...)
		instance.Status.Builds.Cancelled = append(runtimeState.Builds.Cancelled, s2iState.Builds.Cancelled...)
		instance.Status.Builds.Complete = append(runtimeState.Builds.Complete, s2iState.Builds.Complete...)
	}

	return runtimeState.ImageExists && s2iState.ImageExists, runtimeRunning || s2iRunning, runtimeUpdated || s2iUpdated, runtimeError, s2iError, nil
}

func checkBuildsStatus(state *v1alpha1.Builds, lastState *v1alpha1.Builds) (updated bool, running bool, newFailed bool) {
	if len(state.New) > 0 || len(state.Pending) > 0 || len(state.Running) > 0 {
		running = true
	}

	if !util.ContainsAll(lastState.New, state.New) ||
		!util.ContainsAll(lastState.Pending, state.Pending) ||
		!util.ContainsAll(lastState.Running, state.Running) ||
		!util.ContainsAll(lastState.Complete, state.Complete) {
		updated = true
	}

	if !util.ContainsAll(lastState.Failed, state.Failed) ||
		!util.ContainsAll(lastState.Error, state.Error) ||
		!util.ContainsAll(lastState.Cancelled, state.Cancelled) {
		updated = true
		newFailed = true
	}

	return updated, running, newFailed
}

func getBCLabelsAsUniqueSelectors(bc *obuildv1.BuildConfig) string {
	return fmt.Sprintf("%s=%s,%s=%s",
		resource.LabelKeyAppName,
		bc.Labels[resource.LabelKeyAppName],
		resource.LabelKeyBuildType,
		bc.Labels[resource.LabelKeyBuildType],
	)
}

func statusUpdateForKogitoApp(instance *v1alpha1.KogitoApp, cachedInstance *v1alpha1.KogitoApp) bool {
	// Update CR if needed
	if !reflect.DeepEqual(instance.Spec, cachedInstance.Spec) {
		if instance.Status.SetProvisioning() &&
			instance.ResourceVersion == cachedInstance.ResourceVersion {
			log.Infof("Instance spec updated")
			return true
		}
		return false
	}
	if !reflect.DeepEqual(instance.Status, cachedInstance.Status) {
		if instance.ResourceVersion == cachedInstance.ResourceVersion {
			log.Infof("Instance status updated")
			return true
		}
		return false
	}

	return false
}
