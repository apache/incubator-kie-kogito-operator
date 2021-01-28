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

package kogitoinfra

import (
	keycloakv1alpha1 "github.com/keycloak/keycloak-operator/pkg/apis/keycloak/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/core/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// keycloakMetricsExtension default extension enabled in Keycloak default installations
	keycloakMetricsExtension = "https://github.com/aerogear/keycloak-metrics-spi/releases/download/1.0.4/keycloak-metrics-spi-1.0.4.jar"
)

func initkeycloakInfraReconciler(context targetContext) *keycloakInfraReconciler {
	context.log = context.log.WithValues("resource", "keycloak")
	return &keycloakInfraReconciler{
		targetContext: context,
	}
}

// keycloakInfraReconciler implementation of KogitoInfraResource
type keycloakInfraReconciler struct {
	targetContext
}

// AppendKeycloakWatchedObjects ...
func AppendKeycloakWatchedObjects(b *builder.Builder) *builder.Builder {
	return b
}

// Reconcile reconcile Kogito infra object
func (k *keycloakInfraReconciler) Reconcile() (requeue bool, resultErr error) {
	var keycloakInstance *keycloakv1alpha1.Keycloak
	keycloakHandler := infrastructure.NewKeycloakHandler(k.client, k.log)
	if !keycloakHandler.IsKeycloakAvailable() {
		return false, errorForResourceAPINotFound(k.instance.GetSpec().GetResource().APIVersion)
	}

	if len(k.instance.GetSpec().GetResource().Name) > 0 {
		k.log.Debug("Custom Keycloak instance reference is provided")
		namespace := k.instance.GetSpec().GetResource().Namespace
		if len(namespace) == 0 {
			namespace = k.instance.GetNamespace()
			k.log.Debug("Namespace is not provided for custom resource, taking instance", "Namespace", namespace)
		}
		if keycloakInstance, resultErr = k.loadDeployedKeycloakInstance(k.instance.GetSpec().GetResource().Name, namespace); resultErr != nil {
			return false, resultErr
		} else if keycloakInstance == nil {
			return false, errorForResourceNotFound("Keycloak", k.instance.GetSpec().GetResource().Name, namespace)
		}
	} else {
		k.log.Debug("Custom Keycloak instance reference is not provided")
		// check whether Keycloak instance exist
		keycloakInstance, resultErr := k.loadDeployedKeycloakInstance(infrastructure.KeycloakInstanceName, k.instance.GetNamespace())
		if resultErr != nil {
			return false, resultErr
		}

		if keycloakInstance == nil {
			// if not exist then create new Keycloak instance. Keycloak operator creates Keycloak instance, secret & service resource
			_, resultErr = k.createNewKeycloakInstance(infrastructure.KeycloakInstanceName, k.instance.GetNamespace())
			if resultErr != nil {
				return false, resultErr
			}
			return true, nil
		}
	}
	return false, nil
}

func (k *keycloakInfraReconciler) loadDeployedKeycloakInstance(name string, namespace string) (*keycloakv1alpha1.Keycloak, error) {
	k.log.Debug("fetching deployed kogito Keycloak instance")
	keycloakInstance := &keycloakv1alpha1.Keycloak{}
	if exits, err := kubernetes.ResourceC(k.client).FetchWithKey(types.NamespacedName{Name: name, Namespace: namespace}, keycloakInstance); err != nil {
		k.log.Error(err, "Error occurs while fetching kogito Keycloak instance")
		return nil, err
	} else if !exits {
		k.log.Debug("Kogito Keycloak instance is not exists")
		return nil, nil
	} else {
		k.log.Debug("Kogito Keycloak instance found")
		return keycloakInstance, nil
	}
}

func (k *keycloakInfraReconciler) createNewKeycloakInstance(name string, namespace string) (*keycloakv1alpha1.Keycloak, error) {
	k.log.Debug("Going to create kogito Keycloak instance")
	k.log.Debug("Creating default resources for Keycloak installation for Kogito Infra", "Namespace", namespace)
	keycloakInstance := &keycloakv1alpha1.Keycloak{
		ObjectMeta: v1.ObjectMeta{Namespace: namespace, Name: name},
		Spec: keycloakv1alpha1.KeycloakSpec{
			Extensions:     []string{keycloakMetricsExtension},
			Instances:      1,
			ExternalAccess: keycloakv1alpha1.KeycloakExternalAccess{Enabled: true},
		},
	}
	if err := controllerutil.SetOwnerReference(k.instance, keycloakInstance, k.scheme); err != nil {
		return nil, err
	}
	if err := kubernetes.ResourceC(k.client).Create(keycloakInstance); err != nil {
		k.log.Error(err, "Error occurs while creating kogito Keycloak instance")
		return nil, err
	}
	k.log.Debug("Successfully created Kogito Keycloak instance")
	return keycloakInstance, nil
}
