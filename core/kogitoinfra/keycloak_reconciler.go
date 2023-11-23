// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package kogitoinfra

import (
	"github.com/kiegroup/kogito-operator/core/client/kubernetes"
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	keycloakv1alpha1 "github.com/kiegroup/kogito-operator/core/infrastructure/keycloak/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
)

// keycloakInfraReconciler implementation of KogitoInfraResource
type keycloakInfraReconciler struct {
	infraContext
}

func initkeycloakInfraReconciler(context infraContext) Reconciler {
	context.Log = context.Log.WithValues("resource", "keycloak")
	return &keycloakInfraReconciler{
		infraContext: context,
	}
}

// AppendKeycloakWatchedObjects ...
func AppendKeycloakWatchedObjects(b *builder.Builder) *builder.Builder {
	return b
}

// Reconcile reconcile Kogito infra object
func (k *keycloakInfraReconciler) Reconcile() (resultErr error) {
	var keycloakInstance *keycloakv1alpha1.Keycloak
	keycloakHandler := infrastructure.NewKeycloakHandler(k.Context)
	if !keycloakHandler.IsKeycloakAvailable() {
		return errorForResourceAPINotFound(k.instance.GetSpec().GetResource().GetAPIVersion())
	}

	if len(k.instance.GetSpec().GetResource().GetName()) > 0 {
		k.Log.Debug("Custom Keycloak instance reference is provided")
		namespace := k.instance.GetSpec().GetResource().GetNamespace()
		if len(namespace) == 0 {
			namespace = k.instance.GetNamespace()
			k.Log.Debug("Namespace is not provided for custom resource, taking instance", "Namespace", namespace)
		}
		if keycloakInstance, resultErr = k.loadDeployedKeycloakInstance(k.instance.GetSpec().GetResource().GetName(), namespace); resultErr != nil {
			return resultErr
		} else if keycloakInstance == nil {
			return errorForResourceNotFound("Keycloak", k.instance.GetSpec().GetResource().GetName(), namespace)
		}
	} else {
		return errorForResourceConfigError(k.instance, "No Keycloak resource name given")
	}
	return nil
}

func (k *keycloakInfraReconciler) loadDeployedKeycloakInstance(name string, namespace string) (*keycloakv1alpha1.Keycloak, error) {
	k.Log.Debug("fetching deployed kogito Keycloak instance")
	keycloakInstance := &keycloakv1alpha1.Keycloak{}
	if exits, err := kubernetes.ResourceC(k.Client).FetchWithKey(types.NamespacedName{Name: name, Namespace: namespace}, keycloakInstance); err != nil {
		k.Log.Error(err, "Error occurs while fetching kogito Keycloak instance")
		return nil, err
	} else if !exits {
		k.Log.Debug("Kogito Keycloak instance is not exists")
		return nil, nil
	} else {
		k.Log.Debug("Kogito Keycloak instance found")
		return keycloakInstance, nil
	}
}
