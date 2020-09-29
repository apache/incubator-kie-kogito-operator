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

package keycloak

import (
	"fmt"
	keycloakv1alpha1 "github.com/keycloak/keycloak-operator/pkg/apis/keycloak/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// APIVersion refers to kafka APIVersion
	APIVersion = "keycloak.org/v1alpha1"
	// Kind refers to kafka Kind
	Kind = "Keycloak"
)

// InfraResource implementation of KogitoInfraResource
type InfraResource struct {
}

// GetWatchedObjects provide list of object that needs to be watched to maintain Keycloak kogitoInfra resource
func GetWatchedObjects() []framework.WatchedObjects {
	return []framework.WatchedObjects{
		{
			GroupVersion: keycloakv1alpha1.SchemeGroupVersion,
			AddToScheme:  keycloakv1alpha1.SchemeBuilder.AddToScheme,
			Objects:      []runtime.Object{&keycloakv1alpha1.Keycloak{}},
		},
	}
}

// Reconcile reconcile Kogito infra object
func (k *InfraResource) Reconcile(client *client.Client, instance *v1alpha1.KogitoInfra, scheme *runtime.Scheme) (requeue bool, resultErr error) {
	var keycloakInstance *keycloakv1alpha1.Keycloak

	// Step 1: check whether user has provided custom keycloak instance reference
	isCustomReferenceProvided := len(instance.Spec.Resource.Name) > 0
	if isCustomReferenceProvided {
		log.Debugf("Custom Keycloak instance reference is provided")
		namespace := instance.Spec.Resource.Namespace
		if len(namespace) == 0 {
			namespace = instance.Namespace
			log.Debugf("Namespace is not provided for custom resource, taking instance namespace(%s) as default", namespace)
		}
		keycloakInstance, resultErr = loadDeployedKeycloakInstance(client, instance.Spec.Resource.Name, namespace)
		if resultErr != nil {
			return false, resultErr
		}
		if keycloakInstance == nil {
			return false, fmt.Errorf("custom Keycloak instance(%s) not found in namespace %s", instance.Spec.Resource.Name, namespace)
		}
	} else {
		log.Debugf("Custom Keycloak instance reference is not provided")

		// Step 1: Validation
		keycloakAvailable := infrastructure.IsKeycloakAvailable(client)
		if !keycloakAvailable {
			resultErr = fmt.Errorf("Keycloak is not available in the namespace %s, impossible to continue ", instance.Namespace)
			return
		}

		// check whether Keycloak instance exist
		keycloakInstance, resultErr := loadDeployedKeycloakInstance(client, InstanceName, instance.Namespace)
		if resultErr != nil {
			return false, resultErr
		}

		if keycloakInstance == nil {
			// if not exist then create new Keycloak instance. Keycloak operator creates Keycloak instance, secret & service resource
			_, resultErr = createNewKeycloakInstance(client, InstanceName, instance.Namespace, instance, scheme)
			if resultErr != nil {
				return false, resultErr
			}
			return true, nil
		}
	}
	return false, nil
}
