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

package kogitoservice

import (
	"fmt"
	"github.com/kiegroup/kogito-operator/api"
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	"github.com/kiegroup/kogito-operator/core/operator"
	"k8s.io/apimachinery/pkg/types"
)

// ConfigMapReferenceReconciler ...
type ConfigMapReferenceReconciler interface {
	Reconcile() error
}

type configMapReferenceReconciler struct {
	operator.Context
	instance          api.KogitoService
	serviceDefinition *ServiceDefinition
	configMapHandler  infrastructure.ConfigMapHandler
}

func newConfigMapReferenceReconciler(context operator.Context, instance api.KogitoService, serviceDefinition *ServiceDefinition) ConfigMapReferenceReconciler {
	context.Log = context.Log.WithValues("resource", "ConfigMapReference")
	return &configMapReferenceReconciler{
		Context:           context,
		instance:          instance,
		serviceDefinition: serviceDefinition,
		configMapHandler:  infrastructure.NewConfigMapHandler(context),
	}
}

// Reconcile reconcile Kogito infra object
func (i *configMapReferenceReconciler) Reconcile() error {
	configMapReference := i.instance.GetSpec().GetPropertiesConfigMap()
	if len(configMapReference.GetName()) > 0 {
		i.Log.Debug("Custom Configmap instance reference is provided")
		namespace := i.instance.GetNamespace()
		configMapInstance, resultErr := i.configMapHandler.FetchConfigMap(types.NamespacedName{Name: configMapReference.GetName(), Namespace: namespace})
		if resultErr != nil {
			return resultErr
		}
		if configMapInstance == nil {
			return fmt.Errorf("configmap(%s) not found in namespace %s", configMapReference.GetName(), namespace)
		}
	} else {
		return fmt.Errorf("configmap resource name not provided")
	}

	i.updateConfigMapReferenceInStatus(configMapReference)
	return nil
}

func (i *configMapReferenceReconciler) updateConfigMapReferenceInStatus(configMapReference api.ConfigMapReferenceInterface) {
	i.serviceDefinition.ConfigMapReferences = append(i.serviceDefinition.ConfigMapReferences, configMapReference)
}
