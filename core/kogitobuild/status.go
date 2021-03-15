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

package kogitobuild

import (
	"github.com/kiegroup/kogito-cloud-operator/api"
	"github.com/kiegroup/kogito-cloud-operator/core/client"
	"github.com/kiegroup/kogito-cloud-operator/core/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/core/client/openshift"
	"github.com/kiegroup/kogito-cloud-operator/core/framework"
	"github.com/kiegroup/kogito-cloud-operator/core/operator"
	buildv1 "github.com/openshift/api/build/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"sort"
	"strings"
)

// StatusHandler ...
type StatusHandler interface {
	HandleStatusChange(instance api.KogitoBuildInterface, err error)
}

type statusHandler struct {
	*operator.Context
}

// NewStatusHandler ...
func NewStatusHandler(context *operator.Context) StatusHandler {
	return &statusHandler{
		Context: context,
	}
}

func (s *statusHandler) HandleStatusChange(instance api.KogitoBuildInterface, err error) {
	updateStatus := false
	if err != nil {
		updateStatus = s.setFailedCondition(instance.GetStatus(), api.OperatorFailureReason, err.Error())
	} else {
		if updateStatus, err = s.handleConditionTransition(instance, s.Client); err != nil {
			s.Log.Error(err, "Failed to update build status")
		}
	}
	if updateStatus {
		if err = s.updateStatus(instance); err != nil {
			s.Log.Error(err, "Failed to update KogitoBuild")
		}
	}
}

// newSuccessfulCondition ...
func (s *statusHandler) newSuccessfulCondition(status metav1.ConditionStatus) metav1.Condition {
	return metav1.Condition{
		Type:               string(api.KogitoBuildSuccessful),
		Status:             status,
		LastTransitionTime: metav1.Now(),
	}
}

// newRunningCondition ...
func (s *statusHandler) newRunningCondition(status metav1.ConditionStatus) metav1.Condition {
	return metav1.Condition{
		Type:               string(api.KogitoBuildRunning),
		Status:             status,
		LastTransitionTime: metav1.Now(),
	}
}

// NewFailedCondition ...
func (s *statusHandler) newFailedCondition(reason api.KogitoBuildConditionReason, message string) metav1.Condition {
	return metav1.Condition{
		Type:               string(api.KogitoBuildFailure),
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             string(reason),
		Message:            message,
	}
}

// SetProvisioning Sets the condition type to Provisioning and status True if not yet set.
func (s *statusHandler) setSuccessful(c api.ConditionMetaInterface, status metav1.ConditionStatus) bool {
	for _, condition := range c.GetConditions() {
		if condition.Type == string(api.KogitoBuildSuccessful) {
			if condition.Status == status {
				return false
			}
			condition.Status = status
			condition.LastTransitionTime = metav1.Now()
			return true
		}
	}
	successfulCondition := s.newSuccessfulCondition(status)
	c.AddCondition(successfulCondition)
	return true
}

// SetProvisioning Sets the condition type to Provisioning and status True if not yet set.
func (s *statusHandler) setRunning(c api.ConditionMetaInterface, status metav1.ConditionStatus) bool {
	for _, condition := range c.GetConditions() {
		if condition.Type == string(api.KogitoBuildRunning) {
			if condition.Status == status {
				return false
			}
			condition.Status = status
			condition.LastTransitionTime = metav1.Now()
			return true
		}
	}
	runningCondition := s.newRunningCondition(status)
	c.AddCondition(runningCondition)
	return true
}

// SetProvisioning Sets the condition type to Provisioning and status True if not yet set.
func (s *statusHandler) setFailed(c api.ConditionMetaInterface, reason api.KogitoBuildConditionReason, message string) bool {
	for _, condition := range c.GetConditions() {
		if condition.Type == string(api.KogitoBuildFailure) {
			if condition.Message != message {
				condition.Message = message
				condition.LastTransitionTime = metav1.Now()
				return true
			}
			return false
		}
	}
	failedCondition := s.newFailedCondition(reason, message)
	c.AddCondition(failedCondition)
	return true
}

func (s *statusHandler) removeFailedCondition(c api.ConditionMetaInterface) bool {
	for i, condition := range c.GetConditions() {
		if condition.Type == string(api.KogitoBuildFailure) {
			c.RemoveCondition(i)
			return true
		}
	}
	return false
}

func (s *statusHandler) handleConditionTransition(instance api.KogitoBuildInterface, client *client.Client) (updateStatus bool, err error) {
	changed := false
	changed, err = updateBuildsStatus(instance, client)
	if err != nil {
		return false, err
	}
	updateStatus = updateStatus || changed

	builds := &buildv1.BuildList{}
	err = kubernetes.ResourceC(client).ListWithNamespaceAndLabel(
		instance.GetNamespace(), builds,
		map[string]string{
			framework.LabelAppKey: GetApplicationName(instance),
			LabelKeyBuildType:     string(instance.GetSpec().GetType())})
	if err != nil {
		return false, err
	}
	if len(builds.Items) > 0 {
		sort.SliceStable(builds.Items, func(i, j int) bool {
			return builds.Items[i].CreationTimestamp.After(builds.Items[j].CreationTimestamp.Time)
		})
		latestBuild := builds.Items[0]
		if latestBuild.Name != instance.GetStatus().GetLatestBuild() {
			instance.GetStatus().SetLatestBuild(latestBuild.Name)
			changed = true
		}
		updateStatus = updateStatus || changed

		changed = s.addCondition(latestBuild, instance.GetStatus())
		updateStatus = updateStatus || changed
		return updateStatus, nil
	}
	changed = s.setRunningCondition(instance.GetStatus())
	return updateStatus || changed, nil
}

func updateBuildsStatus(instance api.KogitoBuildInterface, client *client.Client) (changed bool, err error) {
	buildsStatus, err := openshift.BuildConfigC(client).GetBuildsStatusByLabel(
		instance.GetNamespace(),
		strings.Join([]string{
			strings.Join([]string{framework.LabelAppKey, GetApplicationName(instance)}, "="),
			strings.Join([]string{LabelKeyBuildType, string(instance.GetSpec().GetType())}, "="),
		}, ","))
	if err != nil {
		return false, err
	}
	if buildsStatus != nil && !reflect.DeepEqual(buildsStatus, instance.GetStatus().GetBuilds()) {
		instance.GetStatus().SetBuilds(buildsStatus)
		return true, nil
	}
	return false, nil
}

func (s *statusHandler) addCondition(build buildv1.Build, c api.ConditionMetaInterface) bool {
	switch build.Status.Phase {
	case buildv1.BuildPhaseFailed, buildv1.BuildPhaseCancelled:
		return s.setFailedCondition(c, api.BuildFailureReason, build.Status.Message)
	case buildv1.BuildPhaseNew, buildv1.BuildPhasePending, buildv1.BuildPhaseRunning:
		return s.setRunningCondition(c)
	case buildv1.BuildPhaseComplete:
		return s.setSuccessfulCondition(c)
	}
	return false
}

func (s *statusHandler) setFailedCondition(c api.ConditionMetaInterface, reason api.KogitoBuildConditionReason, message string) (updateStatus bool) {
	changed := false
	changed = s.setFailed(c, reason, message)
	updateStatus = updateStatus || changed

	changed = s.setRunning(c, metav1.ConditionFalse)
	updateStatus = updateStatus || changed

	changed = s.setSuccessful(c, metav1.ConditionFalse)
	updateStatus = updateStatus || changed
	return
}

func (s *statusHandler) setRunningCondition(c api.ConditionMetaInterface) (updateStatus bool) {
	changed := false
	changed = s.removeFailedCondition(c)
	updateStatus = updateStatus || changed

	changed = s.setRunning(c, metav1.ConditionTrue)
	updateStatus = updateStatus || changed

	changed = s.setSuccessful(c, metav1.ConditionFalse)
	updateStatus = updateStatus || changed
	return
}

func (s *statusHandler) setSuccessfulCondition(c api.ConditionMetaInterface) (updateStatus bool) {
	changed := false
	changed = s.removeFailedCondition(c)
	updateStatus = updateStatus || changed

	changed = s.setRunning(c, metav1.ConditionFalse)
	updateStatus = updateStatus || changed

	changed = s.setSuccessful(c, metav1.ConditionTrue)
	updateStatus = updateStatus || changed
	return
}

func (s *statusHandler) updateStatus(instance api.KogitoBuildInterface) error {
	if err := kubernetes.ResourceC(s.Client).UpdateStatus(instance); err != nil {
		return err
	}
	return nil
}
