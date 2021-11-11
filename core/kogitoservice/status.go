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

	api "github.com/kiegroup/kogito-operator/apis"
	"github.com/kiegroup/kogito-operator/core/client/kubernetes"
	"github.com/kiegroup/kogito-operator/core/framework"
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	"github.com/kiegroup/kogito-operator/core/operator"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// StatusHandler ...
type StatusHandler interface {
	HandleStatusUpdate(instance api.KogitoService, err *error)
}

type statusHandler struct {
	operator.Context
	errorHandler framework.ReconciliationErrorHandler
}

// NewStatusHandler ...
func NewStatusHandler(context operator.Context) StatusHandler {
	return &statusHandler{
		Context:      context,
		errorHandler: framework.NewReconciliationErrorHandler(context),
	}
}

func (s *statusHandler) HandleStatusUpdate(instance api.KogitoService, err *error) {
	s.Log.Info("Updating status for Kogito Service", "err", err)
	if statusErr := s.ensureResourcesStatusChanges(instance, *err); statusErr != nil {
		s.Log.Error(statusErr, "Error while updating Status for Kogito Service")
		return
	}
	s.Log.Info("Finished Kogito Service reconciliation")
}

func (s *statusHandler) ensureResourcesStatusChanges(instance api.KogitoService, errCondition error) (err error) {
	if instance.GetStatus().GetConditions() == nil {
		instance.GetStatus().SetConditions(&[]metav1.Condition{})
	}
	if errCondition != nil {
		if err = s.setFailedConditions(instance, s.errorHandler.GetReasonForError(errCondition), errCondition); err != nil {
			return err
		}
	} else {
		if err = s.handleConditionTransition(instance); err != nil {
			return err
		}
		if err = s.updateImageStatus(instance); err != nil {
			return err
		}
		if err = s.updateRouteStatus(instance); err != nil {
			return err
		}
		if err = s.updateDeploymentStatus(instance); err != nil {
			return err
		}
	}
	if err := s.updateStatus(instance); err != nil {
		s.Log.Error(err, "Error while trying to update status")
		return err
	}
	return nil
}

func (s *statusHandler) setFailedConditions(instance api.KogitoService, reason framework.ConditionReason, errCondition error) error {
	s.setFailed(instance.GetStatus().GetConditions(), metav1.ConditionTrue, reason, errCondition.Error())
	if s.errorHandler.IsReconciliationError(errCondition) {
		s.setProvisioning(instance.GetStatus().GetConditions(), metav1.ConditionTrue, framework.ProvisioningInProgressReason)
	} else {
		s.setProvisioning(instance.GetStatus().GetConditions(), metav1.ConditionFalse, framework.FailedProvisioningReason)
	}

	availableReplicas, err := s.fetchReadyReplicas(instance)
	if err != nil {
		return err
	}
	if availableReplicas > 0 {
		s.setDeployed(instance.GetStatus().GetConditions(), metav1.ConditionTrue)
	} else {
		s.setDeployed(instance.GetStatus().GetConditions(), metav1.ConditionFalse)
	}
	return nil
}

func (s *statusHandler) handleConditionTransition(instance api.KogitoService) error {
	s.InvalidateFailedCondition(instance.GetStatus().GetConditions())
	availableReplicas, err := s.fetchReadyReplicas(instance)
	if err != nil {
		return err
	}
	expectedReplicas := *instance.GetSpec().GetReplicas()
	if expectedReplicas == availableReplicas {
		s.setDeployed(instance.GetStatus().GetConditions(), metav1.ConditionTrue)
		s.setProvisioning(instance.GetStatus().GetConditions(), metav1.ConditionFalse, framework.FinishedProvisioningReason)
	} else if availableReplicas > 0 && availableReplicas < expectedReplicas {
		s.setDeployed(instance.GetStatus().GetConditions(), metav1.ConditionTrue)
		s.setProvisioning(instance.GetStatus().GetConditions(), metav1.ConditionTrue, framework.ProvisioningInProgressReason)
	} else if availableReplicas == 0 {
		s.setDeployed(instance.GetStatus().GetConditions(), metav1.ConditionFalse)
		s.setProvisioning(instance.GetStatus().GetConditions(), metav1.ConditionTrue, framework.ProvisioningInProgressReason)
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

func (s *statusHandler) updateImageStatus(instance api.KogitoService) error {
	deploymentHandler := infrastructure.NewDeploymentHandler(s.Context)
	deployment, err := deploymentHandler.FetchDeployment(types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})
	if err != nil {
		return err
	} else if deployment == nil {
		return nil
	}
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		image := deployment.Spec.Template.Spec.Containers[0].Image
		instance.GetStatus().SetImage(image)
	}
	return nil
}

func (s *statusHandler) updateDeploymentStatus(instance api.KogitoService) error {
	deploymentHandler := infrastructure.NewDeploymentHandler(s.Context)
	deployment, err := deploymentHandler.FetchDeployment(types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})
	if err != nil {
		return err
	} else if deployment == nil {
		return nil
	}
	instance.GetStatus().SetDeploymentConditions(deployment.Status.Conditions)
	return nil
}

func (s *statusHandler) updateRouteStatus(instance api.KogitoService) error {
	if s.Client.IsOpenshift() {
		if instance.GetStatus().GetRouteConditions() == nil {
			instance.GetStatus().SetRouteConditions(&[]metav1.Condition{})
		}

		routeHandler := infrastructure.NewRouteHandler(s.Context)
		if instance.GetSpec().IsRouteDisabled() {
			// update route condition that route was disabled
			s.Log.Debug("Routes are disabled.")
			successCondition := s.newFailedCondition(metav1.ConditionFalse, framework.RouteProcessed, "Routes are disabled.")
			meta.SetStatusCondition(instance.GetStatus().GetRouteConditions(), successCondition)
			return nil
		}
		routeKey := types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()}
		// check if route is successfully created and update runtime conditions with error message
		if isAdmitted, err := routeHandler.ValidateRouteStatus(routeKey); err != nil {
			s.Log.Info("Error occured during route creation", "err", err)
			errCondition := framework.ErrorForRouteCreation(err)
			failedCondition := s.newFailedCondition(metav1.ConditionTrue, s.errorHandler.GetReasonForError(errCondition), errCondition.Error())
			meta.SetStatusCondition(instance.GetStatus().GetRouteConditions(), failedCondition)
			return nil

		} else if isAdmitted {
			successCondition := s.newFailedCondition(metav1.ConditionFalse, framework.RouteProcessed, "Route admitted.")
			meta.SetStatusCondition(instance.GetStatus().GetRouteConditions(), successCondition)

		}
		route, err := routeHandler.GetHostFromRoute(routeKey)
		if err != nil {
			return err
		}

		if len(route) > 0 {
			uri := fmt.Sprintf("http://%s", route)
			instance.GetStatus().SetExternalURI(uri)
		}
	}
	return nil
}

// NewDeployedCondition ...
func (s *statusHandler) newDeployedCondition(status metav1.ConditionStatus) metav1.Condition {
	reason := framework.SuccessfulDeployedReason
	if status == metav1.ConditionFalse {
		reason = framework.FailedDeployedReason
	}
	return metav1.Condition{
		Type:   string(api.DeployedConditionType),
		Status: status,
		Reason: string(reason),
	}
}

// NewProvisioningCondition ...
func (s *statusHandler) newProvisioningCondition(status metav1.ConditionStatus, reason framework.ConditionReason) metav1.Condition {
	return metav1.Condition{
		Type:   string(api.ProvisioningConditionType),
		Status: status,
		Reason: string(reason),
	}
}

// NewFailedCondition ...
func (s *statusHandler) newFailedCondition(status metav1.ConditionStatus, reason framework.ConditionReason, message string) metav1.Condition {
	return metav1.Condition{
		Type:    string(api.FailedConditionType),
		Status:  status,
		Reason:  string(reason),
		Message: message,
	}
}

// SetProvisioning Sets the condition type to Provisioning and status True if not yet set.
func (s *statusHandler) setProvisioning(conditions *[]metav1.Condition, status metav1.ConditionStatus, reason framework.ConditionReason) {
	provisionCondition := s.newProvisioningCondition(status, reason)
	meta.SetStatusCondition(conditions, provisionCondition)
}

// SetProvisioning Sets the condition type to Provisioning and status True if not yet set.
func (s *statusHandler) setDeployed(conditions *[]metav1.Condition, status metav1.ConditionStatus) {
	deployedCondition := s.newDeployedCondition(status)
	meta.SetStatusCondition(conditions, deployedCondition)
}

// SetProvisioning Sets the condition type to Provisioning and status True if not yet set.
func (s *statusHandler) setFailed(conditions *[]metav1.Condition, status metav1.ConditionStatus, reason framework.ConditionReason, message string) {
	failedCondition := s.newFailedCondition(status, reason, message)
	meta.SetStatusCondition(conditions, failedCondition)
}

func (s *statusHandler) InvalidateFailedCondition(conditions *[]metav1.Condition) {
	failedCondition := meta.FindStatusCondition(*conditions, string(api.FailedConditionType))
	if failedCondition != nil {
		s.setFailed(conditions, metav1.ConditionFalse, framework.ConditionReason(failedCondition.Reason), failedCondition.Message)
	}
}

func (s *statusHandler) fetchReadyReplicas(instance api.KogitoService) (int32, error) {
	deploymentHandler := infrastructure.NewDeploymentHandler(s.Context)
	readyReplicas, err := deploymentHandler.FetchReadyReplicas(types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})
	if err != nil {
		return 0, err
	}
	return readyReplicas, nil
}
