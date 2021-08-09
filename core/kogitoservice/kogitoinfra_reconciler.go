// Copyright 2021 Red Hat, Inc. and/or its affiliates
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
	"github.com/kiegroup/kogito-operator/api"
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	"github.com/kiegroup/kogito-operator/core/manager"
	"github.com/kiegroup/kogito-operator/core/operator"
	"k8s.io/apimachinery/pkg/types"
)

// KogitoInfraReconciler ...
type KogitoInfraReconciler interface {
	Reconcile() error
}

type kogitoInfraReconciler struct {
	operator.Context
	instance          api.KogitoService
	serviceDefinition *ServiceDefinition
	infraHandler      manager.KogitoInfraHandler
	infraManager      manager.KogitoInfraManager
}

func newKogitoInfraReconciler(context operator.Context, instance api.KogitoService, serviceDefinition *ServiceDefinition, infraHandler manager.KogitoInfraHandler) KogitoInfraReconciler {
	return &kogitoInfraReconciler{
		Context:           context,
		instance:          instance,
		serviceDefinition: serviceDefinition,
		infraHandler:      infraHandler,
		infraManager:      manager.NewKogitoInfraManager(context, infraHandler),
	}
}

func (k *kogitoInfraReconciler) Reconcile() error {
	infraNames := k.instance.GetSpec().GetInfra()
	for _, infraName := range infraNames {

		infra, err := k.infraHandler.FetchKogitoInfraInstance(types.NamespacedName{Name: infraName, Namespace: k.instance.GetNamespace()})
		if err != nil {
			return err
		}
		if infra == nil {
			k.Log.Info("Infra not found", "Infra", infraName)
			return nil
		}

		if err := k.checkInfraDependencies(infra); err != nil {
			return err
		}
		// we need to take ownership of the provided KogitoInfra instances
		if err := k.takeKogitoInfraOwnership(infra); err != nil {
			return err
		}

		k.serviceDefinition.ConfigMapReferences = append(k.serviceDefinition.ConfigMapReferences, infra.GetStatus().GetConfigMapReferences()...)

		k.serviceDefinition.SecretReferences = append(k.serviceDefinition.SecretReferences, infra.GetStatus().GetSecretReferences()...)

		k.serviceDefinition.Envs = append(k.serviceDefinition.Envs, infra.GetStatus().GetEnvs()...)
	}
	return nil
}

// checkInfraDependencies verifies if every KogitoInfra resource have an ok status.
func (k *kogitoInfraReconciler) checkInfraDependencies(infra api.KogitoInfraInterface) error {
	if isReady, err := k.infraManager.IsKogitoInfraReady(types.NamespacedName{Name: infra.GetName(), Namespace: infra.GetNamespace()}); err != nil {
		return err
	} else if !isReady {
		conditionReason, err := k.infraManager.GetKogitoInfraFailureConditionReason(types.NamespacedName{Name: infra.GetName(), Namespace: infra.GetNamespace()})
		if err != nil {
			return err
		}
		return infrastructure.ErrorForInfraNotReady(k.instance.GetName(), infra.GetName(), conditionReason)
	}
	return nil
}

func (k *kogitoInfraReconciler) takeKogitoInfraOwnership(infra api.KogitoInfraInterface) error {
	if err := k.infraManager.TakeKogitoInfraOwnership(types.NamespacedName{Name: infra.GetName(), Namespace: infra.GetNamespace()}, k.instance); err != nil {
		return err
	}
	return nil
}
