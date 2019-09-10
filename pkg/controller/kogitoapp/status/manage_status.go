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
	appsv1 "github.com/openshift/api/apps/v1"
	obuildv1 "github.com/openshift/api/build/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	cachev1 "sigs.k8s.io/controller-runtime/pkg/cache"
)

var log = logger.GetLogger("kogitoapp.controller")

const maxBuffer = 30

// setProvisioning - Sets the condition type to Provisioning and status True if not yet set.
func setProvisioning(cr *v1alpha1.KogitoApp) bool {
	log := log.With("kind", cr.Kind, "name", cr.Name, "namespace", cr.Namespace)
	size := len(cr.Status.Conditions)
	if size > 0 && cr.Status.Conditions[size-1].Type == v1alpha1.ProvisioningConditionType {
		log.Debug("Status: unchanged status [provisioning].")
		return false
	}
	log.Debug("Status: set provisioning")
	condition := v1alpha1.Condition{
		Type:               v1alpha1.ProvisioningConditionType,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
	}
	cr.Status.Conditions = addCondition(cr.Status.Conditions, condition)
	return true
}

// setDeployed - Updates the condition with the DeployedCondition and True status
func setDeployed(cr *v1alpha1.KogitoApp) bool {
	log := log.With("kind", cr.Kind, "name", cr.Name, "namespace", cr.Namespace)
	size := len(cr.Status.Conditions)
	if size > 0 && cr.Status.Conditions[size-1].Type == v1alpha1.DeployedConditionType {
		log.Debug("Status: unchanged status [deployed].")
		return false
	}
	log.Debugf("Status: changed status [deployed].")
	condition := v1alpha1.Condition{
		Type:               v1alpha1.DeployedConditionType,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
	}
	cr.Status.Conditions = addCondition(cr.Status.Conditions, condition)
	return true
}

// setFailed - Sets the failed condition with the error reason and message
func setFailed(cr *v1alpha1.KogitoApp, reason v1alpha1.ReasonType, err error) {
	log := log.With("kind", cr.Kind, "name", cr.Name, "namespace", cr.Namespace)
	log.Debug("Status: set failed")
	condition := v1alpha1.Condition{
		Type:               v1alpha1.FailedConditionType,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            err.Error(),
	}
	cr.Status.Conditions = addCondition(cr.Status.Conditions, condition)
}

func addCondition(conditions []v1alpha1.Condition, condition v1alpha1.Condition) []v1alpha1.Condition {
	size := len(conditions) + 1
	first := 0
	if size > maxBuffer {
		first = size - maxBuffer
	}
	return append(conditions, condition)[first:size]
}

// UpdateStatusResult contains the results of the update of the status
type UpdateStatusResult struct {
	Updated      bool
	RequeueAfter bool
	Err          error
}

func createResult(updated bool, forceUpdate bool, requeueAfter bool, err error, result *UpdateStatusResult) {
	if updated {
		result.Updated = true
	} else if forceUpdate {
		result.Updated = false
	}

	if requeueAfter {
		result.RequeueAfter = true
	}

	if result.Err == nil && err != nil {
		result.Err = err
	}
}

// ManageStatus will guarantee the status changes
func ManageStatus(instance *v1alpha1.KogitoApp, resources *resource.KogitoAppResources, client *client.Client,
	resourcesUpdateResult *resource.UpdateResourcesResult, cache cachev1.Cache, namespacedName types.NamespacedName) *UpdateStatusResult {
	result := &UpdateStatusResult{false, false, nil}

	{
		requeueAfter, updated, err := statusUpdateForImageBuild(instance, resources, client)
		createResult(updated, false, requeueAfter, err, result)
	}

	{
		updated, err := statusUpdateForDeployment(instance, resources, client)
		createResult(updated, false, false, err, result)
	}

	{
		requeueAfter, updated, err := statusUpdateForRoute(instance, resources, client)
		createResult(updated, false, requeueAfter, err, result)
	}

	{
		updated, err := statusUpdateForResources(instance, resourcesUpdateResult)
		createResult(updated, false, false, err, result)
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
				setFailed(instance, v1alpha1.UnknownReason, err)
				createResult(true, false, false, nil, result)
			}
		} else {
			updated := statusUpdateForKogitoApp(instance, cachedInstance)
			createResult(updated, false, false, nil, result)
		}
	}

	if result.Updated {
		// UpdateObj reconciles the given object
		updated, err := updateObj(instance, client)
		createResult(updated, true, false, err, result)
	} else if !result.RequeueAfter && result.Err == nil {
		if setDeployed(instance) {
			log.Infof("'%s' successfully deployed", instance.Name)
			if instance.ResourceVersion == cachedInstance.ResourceVersion {
				updated, err := updateObj(instance, client)
				createResult(updated, true, false, err, result)
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

func statusUpdateForResources(instance *v1alpha1.KogitoApp, result *resource.UpdateResourcesResult) (bool, error) {
	if result != nil {
		if result.Err != nil {
			setFailed(instance, result.ErrorReason, result.Err)
			return true, result.Err
		} else if result.Updated {
			setProvisioning(instance)
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

func statusUpdateForRoute(instance *v1alpha1.KogitoApp, resources *resource.KogitoAppResources, client *client.Client) (requeue bool, updated bool, err error) {
	// Setting route to the status
	if resources.Route != nil {
		log.Debugf("Trying to get the host for the route %s", resources.Route.Name)
		if exists, route, err := openshift.RouteC(client).GetHostFromRoute(types.NamespacedName{Name: resources.Route.Name, Namespace: resources.Route.Namespace}); err != nil {
			return false, false, err
		} else if exists {
			fmtRoute := fmt.Sprintf("http://%s", route)
			if fmtRoute != instance.Status.Route {
				log.Infof("Updating route status")
				instance.Status.Route = fmtRoute
				return false, true, nil
			}

			return false, false, nil
		}

		log.Infof("Failed to get the host of the route %s", resources.Route.Name)
		return true, false, nil
	}

	log.Debugf("Route is nil, impossible to get host to set in the status", resources.Route)
	return false, false, nil
}

func statusUpdateForImageBuild(instance *v1alpha1.KogitoApp, resources *resource.KogitoAppResources, client *client.Client) (requeue bool, updated bool, err error) {
	// ensure builds
	log.Infof("Checking if build for '%s' is finished", instance.Name)
	if imageExists, building, err := ensureApplicationImageExists(instance, resources, client); err != nil {
		return false, false, err
	} else if !imageExists || building {
		// let's wait for the build to finish
		if setProvisioning(instance) {
			return false, true, nil
		}
		log.Infof("Build for '%s' still running", instance.Name)
		return true, false, nil
	}

	return false, false, nil
}

func ensureApplicationImageExists(instance *v1alpha1.KogitoApp, resources *resource.KogitoAppResources, client *client.Client) (exists bool, running bool, err error) {
	runtimeState, err :=
		openshift.BuildConfigC(client).EnsureImageBuild(
			resources.BuildConfigRuntime,
			getBCLabelsAsUniqueSelectors(resources.BuildConfigRuntime))
	if err != nil {
		return false, false, err
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

	if runtimeState.BuildRunning {
		log.Infof("Image for '%s' is being pushed to the registry", instance.Name)
	}

	// verify s2i build and image
	s2iState, err :=
		openshift.BuildConfigC(client).EnsureImageBuild(
			resources.BuildConfigS2I,
			getBCLabelsAsUniqueSelectors(resources.BuildConfigS2I))
	if err != nil {
		return false, runtimeState.BuildRunning, err
	} else if s2iState.BuildRunning {
		// build is running, nothing to do
		log.Infof("Application '%s' build is still running. Won't trigger a new build.", instance.Name)
	} else if !s2iState.ImageExists && !s2iState.BuildRunning {
		log.Warnf("There's no image nor build for '%s' running", resources.BuildConfigS2I.Name)
	}

	if runtimeState.ImageExists && !runtimeState.BuildRunning && s2iState.ImageExists && !s2iState.BuildRunning {
		log.Debugf("There are images for both builds, nothing to do")
		return true, false, nil
	}

	return runtimeState.ImageExists && s2iState.ImageExists, runtimeState.BuildRunning || s2iState.BuildRunning, nil
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
		if setProvisioning(instance) && instance.ResourceVersion == cachedInstance.ResourceVersion {
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
