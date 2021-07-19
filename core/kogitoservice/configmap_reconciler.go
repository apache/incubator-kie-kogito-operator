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
	"github.com/kiegroup/kogito-operator/core/manager"
	"github.com/kiegroup/kogito-operator/core/operator"
	"time"
)

// ConfigMapReconciler ...
type ConfigMapReconciler interface {
	Reconcile() (time.Duration, error)
}

type configMapReconciler struct {
	operator.Context
	serviceDeployer *serviceDeployer
	infraHandler    manager.KogitoInfraHandler
}

// NewConfigMapReconciler ...
func NewConfigMapReconciler(context operator.Context, serviceDeployer *serviceDeployer, infraHandler manager.KogitoInfraHandler) ConfigMapReconciler {
	return &configMapReconciler{
		Context:         context,
		serviceDeployer: serviceDeployer,
		infraHandler:    infraHandler,
	}
}

func (c *configMapReconciler) Reconcile() (reconcileInterval time.Duration, err error) {
	if c.serviceDeployer.definition.OnConfigMapReconcile != nil {
		reconcileInterval, err = c.serviceDeployer.definition.OnConfigMapReconcile()
		if err != nil {
			return
		} else if reconcileInterval > 0 {
			return reconcileInterval, nil
		}
	}

	userProvidedConfigConfigMapReconciler := NewUserProvidedConfigConfigMapReconciler(c.Context, c.serviceDeployer.instance)
	reconcileInterval, err = userProvidedConfigConfigMapReconciler.Reconcile()
	if err != nil {
		return
	} else if reconcileInterval > 0 {
		return reconcileInterval, nil
	}

	customConfigMapReconciler := NewAppConfigMapReconciler(c.Context, c.serviceDeployer.instance, c.infraHandler)
	reconcileInterval, err = customConfigMapReconciler.Reconcile()
	if err != nil {
		return
	} else if reconcileInterval > 0 {
		return reconcileInterval, nil
	}

	return
}
