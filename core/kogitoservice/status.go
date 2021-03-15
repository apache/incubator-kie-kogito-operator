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

package kogitoservice

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/api"
	"github.com/kiegroup/kogito-cloud-operator/core/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/core/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/core/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
)

// StatusHandler ...
type StatusHandler interface {
	HandleStatusUpdate(instance api.KogitoService, err *error)
}

type statusHandler struct {
	*operator.Context
}

// NewStatusHandler ...
func NewStatusHandler(context *operator.Context) StatusHandler {
	return &statusHandler{
		context,
	}
}

func (s *statusHandler) HandleStatusUpdate(instance api.KogitoService, err *error) {
	s.Log.Info("Updating status for Kogito Service")
	if statusErr := s.ensureResourcesStatusChanges(instance, *err); statusErr != nil {
		s.Log.Error(statusErr, "Error while updating Status for Kogito Service")
		return
	}
	s.Log.Info("Finished Kogito Service reconciliation")
}

func (s *statusHandler) ensureResourcesStatusChanges(instance api.KogitoService, errCondition error) (err error) {
	updateStatus := false
	changed := false
	if errCondition != nil {
		changed = s.setFailed(instance.GetStatus(), reasonForError(errCondition), errCondition)
		updateStatus = updateStatus || changed

		if isReconciliationError(errCondition) {
			changed = s.setProvisioning(instance.GetStatus(), metav1.ConditionTrue)
			updateStatus = updateStatus || changed
		} else {
			changed = s.setProvisioning(instance.GetStatus(), metav1.ConditionFalse)
			updateStatus = updateStatus || changed
		}

		availableReplicas, err := s.fetchReadyReplicas(instance)
		if err != nil {
			return err
		}
		if availableReplicas > 0 {
			changed = s.setDeployed(instance.GetStatus(), metav1.ConditionTrue)
			updateStatus = updateStatus || changed
		} else {
			changed = s.setDeployed(instance.GetStatus(), metav1.ConditionFalse)
			updateStatus = updateStatus || changed
		}
	} else {
		changed = s.removeFailedCondition(instance.GetStatus())
		updateStatus = updateStatus || changed

		availableReplicas, err := s.fetchReadyReplicas(instance)
		if err != nil {
			return err
		}
		expectedReplicas := *instance.GetSpec().GetReplicas()
		if expectedReplicas == availableReplicas {
			changed = s.setDeployed(instance.GetStatus(), metav1.ConditionTrue)
			updateStatus = updateStatus || changed
			changed = s.setProvisioning(instance.GetStatus(), metav1.ConditionFalse)
			updateStatus = updateStatus || changed
		} else if availableReplicas > 0 && availableReplicas < expectedReplicas {
			changed = s.setDeployed(instance.GetStatus(), metav1.ConditionTrue)
			updateStatus = updateStatus || changed
			changed = s.setProvisioning(instance.GetStatus(), metav1.ConditionTrue)
			updateStatus = updateStatus || changed
		} else if availableReplicas == 0 {
			changed = s.setDeployed(instance.GetStatus(), metav1.ConditionFalse)
			updateStatus = updateStatus || changed
			changed = s.setProvisioning(instance.GetStatus(), metav1.ConditionTrue)
			updateStatus = updateStatus || changed
		}

		if changed, err = s.updateImageStatus(instance); err != nil {
			return err
		}
		updateStatus = changed || updateStatus

		if changed, err = s.updateRouteStatus(instance); err != nil {
			return err
		}
		updateStatus = changed || updateStatus

		if changed, err = s.updateDeploymentStatus(instance); err != nil {
			return err
		}
		updateStatus = changed || updateStatus
	}

	if updateStatus {
		if err := s.updateStatus(instance); err != nil {
			s.Log.Error(err, "Error while trying to update status")
			return err
		}
	}
	return nil
}

func (s *statusHandler) updateStatus(instance api.KogitoService) error {
	err := kubernetes.ResourceC(s.Client).UpdateStatus(instance)
	if err != nil {
		return err
	}
	return nil
}

func (s *statusHandler) updateImageStatus(instance api.KogitoService) (bool, error) {
	deploymentHandler := infrastructure.NewDeploymentHandler(s.Context)
	deployment, err := deploymentHandler.MustFetchDeployment(types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})
	if err != nil {
		return false, err
	}
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		image := deployment.Spec.Template.Spec.Containers[0].Image
		if len(image) > 0 && image != instance.GetStatus().GetImage() {
			instance.GetStatus().SetImage(deployment.Spec.Template.Spec.Containers[0].Image)
			return true, nil
		}
	}
	return false, nil
}

func (s *statusHandler) updateDeploymentStatus(instance api.KogitoService) (update bool, err error) {
	deploymentHandler := infrastructure.NewDeploymentHandler(s.Context)
	deployment, err := deploymentHandler.MustFetchDeployment(types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})
	if err != nil {
		return false, err
	}

	if !reflect.DeepEqual(instance.GetStatus().GetDeploymentConditions(), deployment.Status.Conditions) {
		instance.GetStatus().SetDeploymentConditions(deployment.Status.Conditions)
		return true, nil
	}
	return false, nil
}

func (s *statusHandler) updateRouteStatus(instance api.KogitoService) (bool, error) {
	if s.Client.IsOpenshift() {
		routeHandler := infrastructure.NewRouteHandler(s.Context)
		route, err := routeHandler.GetHostFromRoute(types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})
		if err != nil {
			return false, err
		}

		if len(route) > 0 {
			uri := fmt.Sprintf("http://%s", route)
			if uri != instance.GetStatus().GetExternalURI() {
				instance.GetStatus().SetExternalURI(uri)
				return true, nil
			}
		}
	}
	return false, nil
}

// NewDeployedCondition ...
func (s *statusHandler) newDeployedCondition(status metav1.ConditionStatus) metav1.Condition {
	return metav1.Condition{
		Type:               string(api.DeployedConditionType),
		Status:             status,
		LastTransitionTime: metav1.Now(),
	}
}

// NewProvisioningCondition ...
func (s *statusHandler) newProvisioningCondition(status metav1.ConditionStatus) metav1.Condition {
	return metav1.Condition{
		Type:               string(api.ProvisioningConditionType),
		Status:             status,
		LastTransitionTime: metav1.Now(),
	}
}

// NewFailedCondition ...
func (s *statusHandler) newFailedCondition(reason api.KogitoServiceConditionReason, err error) metav1.Condition {
	return metav1.Condition{
		Type:               string(api.FailedConditionType),
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             string(reason),
		Message:            err.Error(),
	}
}

// SetProvisioning Sets the condition type to Provisioning and status True if not yet set.
func (s *statusHandler) setProvisioning(c api.ConditionMetaInterface, status metav1.ConditionStatus) bool {
	for _, condition := range c.GetConditions() {
		if condition.Type == string(api.ProvisioningConditionType) {
			if condition.Status == status {
				return false
			}
			condition.Status = status
			condition.LastTransitionTime = metav1.Now()
			return true
		}
	}
	provisionCondition := s.newProvisioningCondition(status)
	c.AddCondition(provisionCondition)
	return true
}

// SetProvisioning Sets the condition type to Provisioning and status True if not yet set.
func (s *statusHandler) setDeployed(c api.ConditionMetaInterface, status metav1.ConditionStatus) bool {
	for _, condition := range c.GetConditions() {
		if condition.Type == string(api.DeployedConditionType) {
			if condition.Status == status {
				return false
			}
			condition.Status = status
			condition.LastTransitionTime = metav1.Now()
			return true
		}
	}
	deployedCondition := s.newDeployedCondition(status)
	c.AddCondition(deployedCondition)
	return true
}

// SetProvisioning Sets the condition type to Provisioning and status True if not yet set.
func (s *statusHandler) setFailed(c api.ConditionMetaInterface, reason api.KogitoServiceConditionReason, err error) bool {
	for _, condition := range c.GetConditions() {
		if condition.Type == string(api.FailedConditionType) {
			if s.isFailedConditionChange(condition, reason, err) {
				condition.Reason = string(reason)
				condition.Message = err.Error()
				condition.LastTransitionTime = metav1.Now()
				return true
			}
			return false
		}
	}
	deployedCondition := s.newFailedCondition(reason, err)
	c.AddCondition(deployedCondition)
	return true
}

func (s *statusHandler) removeFailedCondition(c api.ConditionMetaInterface) bool {
	for i, condition := range c.GetConditions() {
		if condition.Type == string(api.FailedConditionType) {
			c.RemoveCondition(i)
			return true
		}
	}
	return false
}

func (s *statusHandler) isFailedConditionChange(oldCondition metav1.Condition, reason api.KogitoServiceConditionReason, err error) bool {
	return !(oldCondition.Reason == string(reason) && oldCondition.Message == err.Error())
}

func (s *statusHandler) fetchReadyReplicas(instance api.KogitoService) (int32, error) {
	deploymentHandler := infrastructure.NewDeploymentHandler(s.Context)
	readyReplicas, err := deploymentHandler.FetchReadyReplicas(types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})
	if err != nil {
		return 0, err
	}
	return readyReplicas, nil
}
