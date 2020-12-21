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

package controllers

import (
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/controllers/build"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	buildv1 "github.com/openshift/api/build/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"sort"
	"strings"
)

const (
	// maxConditionsBuffer describes the max count of Conditions in the status field
	maxConditionsBuffer = 5
)

var (
	buildConditionStatus = map[buildv1.BuildPhase]v1beta1.KogitoBuildConditionType{
		buildv1.BuildPhaseError:     v1beta1.KogitoBuildFailure,
		buildv1.BuildPhaseFailed:    v1beta1.KogitoBuildFailure,
		buildv1.BuildPhaseCancelled: v1beta1.KogitoBuildFailure,
		buildv1.BuildPhaseNew:       v1beta1.KogitoBuildRunning,
		buildv1.BuildPhasePending:   v1beta1.KogitoBuildRunning,
		buildv1.BuildPhaseRunning:   v1beta1.KogitoBuildRunning,
		buildv1.BuildPhaseComplete:  v1beta1.KogitoBuildSuccessful,
	}
)

func addConditionError(instance *v1beta1.KogitoBuild, err error) {
	if err != nil {
		instance.Status.Conditions = append(instance.Status.Conditions, v1beta1.KogitoBuildConditions{
			Type:               v1beta1.KogitoBuildFailure,
			Status:             v1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             v1beta1.OperatorFailureReason,
			Message:            err.Error(),
		})
	}
}

func addCondition(instance *v1beta1.KogitoBuild, condition v1beta1.KogitoBuildConditionType, reason v1beta1.KogitoBuildConditionReason, message string) bool {
	if len(instance.Status.Conditions) == 0 ||
		instance.Status.Conditions[len(instance.Status.Conditions)-1].Type != condition {
		instance.Status.Conditions = append(instance.Status.Conditions, v1beta1.KogitoBuildConditions{
			Type:               condition,
			Status:             v1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             reason,
			Message:            message,
		})
		return true
	}
	return false
}

func updateBuildsStatus(instance *v1beta1.KogitoBuild, client *client.Client) (changed bool, err error) {
	buildsStatus, err := openshift.BuildConfigC(client).GetBuildsStatusByLabel(
		instance.Namespace,
		strings.Join([]string{
			strings.Join([]string{framework.LabelAppKey, build.GetApplicationName(instance)}, "="),
			strings.Join([]string{build.LabelKeyBuildType, string(instance.Spec.Type)}, "="),
		}, ","))
	if err != nil {
		return false, err
	}
	if buildsStatus != nil && !reflect.DeepEqual(buildsStatus, instance.Status.Builds) {
		instance.Status.Builds = *buildsStatus
		return true, nil
	}
	return false, nil
}

func handleConditionTransition(instance *v1beta1.KogitoBuild, client *client.Client) (changed bool, err error) {
	if changed, err = updateBuildsStatus(instance, client); err != nil {
		return false, err
	}
	builds := &buildv1.BuildList{}
	err = kubernetes.ResourceC(client).ListWithNamespaceAndLabel(
		instance.Namespace, builds,
		map[string]string{
			framework.LabelAppKey:   build.GetApplicationName(instance),
			build.LabelKeyBuildType: string(instance.Spec.Type)})
	if err != nil {
		return changed, err
	}
	if len(builds.Items) > 0 {
		sort.SliceStable(builds.Items, func(i, j int) bool {
			return builds.Items[i].CreationTimestamp.After(builds.Items[j].CreationTimestamp.Time)
		})
		instance.Status.LatestBuild = builds.Items[0].Name
		condition := buildConditionStatus[builds.Items[0].Status.Phase]
		if condition == v1beta1.KogitoBuildFailure {
			return addCondition(instance, condition, v1beta1.BuildFailureReason, builds.Items[0].Status.Message), nil
		}
		return addCondition(instance, condition, "", builds.Items[0].Status.Message), nil
	}
	return addCondition(
		instance,
		v1beta1.KogitoBuildRunning,
		"", "") || changed, nil
}

func trimConditions(instance *v1beta1.KogitoBuild) {
	if len(instance.Status.Conditions) > maxConditionsBuffer {
		low := len(instance.Status.Conditions) - maxConditionsBuffer
		high := len(instance.Status.Conditions)
		instance.Status.Conditions = instance.Status.Conditions[low:high]
	}
}

func sortConditionsByTransitionTime(instance *v1beta1.KogitoBuild) {
	sort.SliceStable(instance.Status.Conditions, func(i, j int) bool {
		return instance.Status.Conditions[i].LastTransitionTime.Before(&instance.Status.Conditions[j].LastTransitionTime)
	})
}

func (r *KogitoBuildReconciler) handleStatusChange(instance *v1beta1.KogitoBuild, err error) {
	needUpdate := false
	sortConditionsByTransitionTime(instance)
	if err != nil {
		needUpdate = true
		addConditionError(instance, err)
	} else {
		if needUpdate, err = handleConditionTransition(instance, r.Client); err != nil {
			r.Log.Error(err, "Failed to update build status for", "Instance", instance.Name)
		}
	}
	trimConditions(instance)
	if needUpdate {
		if err = r.updateStatus(instance); err != nil {
			r.Log.Error(err, "Failed to update KogitoBuild for", "Instance", instance.Name)
		}
	}
}

func (r *KogitoBuildReconciler) updateStatus(instance *v1beta1.KogitoBuild) error {
	if err := kubernetes.ResourceC(r.Client).UpdateStatus(instance); err != nil {
		return err
	}
	return nil
}
