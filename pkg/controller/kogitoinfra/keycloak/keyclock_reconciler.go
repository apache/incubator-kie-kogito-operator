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
		resourceName := instance.Spec.Resource.Name
		resourceNameSpace := instance.Spec.Resource.Namespace

		keycloakInstance, resultErr = loadDeployedKeycloakInstance(client, resourceName, resourceNameSpace)
		if resultErr != nil {
			return false, resultErr
		}
		if keycloakInstance == nil {
			return false, fmt.Errorf("custom Keycloak instance(%s) not found in namespace %s", resourceName, resourceNameSpace)
		}
	} else {
		log.Debugf("Custom Keycloak instance reference is not provided")
		// if resource name is not provided then Keycloak instance should be created with default name = kogito-keycloak
		resourceName := InstanceName

		// Keycloak resource should be created in same namespace as kogitoInfra
		resourceNameSpace := instance.Namespace

		// Step 1: Validation
		keycloakAvailable := infrastructure.IsKeycloakAvailable(client)
		if !keycloakAvailable {
			resultErr = fmt.Errorf("Keycloak is not available in the namespace %s, impossible to continue ", instance.Namespace)
			return
		}

		// check whether Keycloak instance exist
		keycloakInstance, resultErr := loadDeployedKeycloakInstance(client, resourceName, resourceNameSpace)
		if resultErr != nil {
			return false, resultErr
		}

		if keycloakInstance == nil {
			// if not exist then create new Keycloak instance. Keycloak operator creates Keycloak instance, secret & service resource
			_, resultErr = createNewKeycloakInstance(client, resourceName, resourceNameSpace, instance, scheme)
			if resultErr != nil {
				return false, resultErr
			}
			return true, nil
		}
	}
	return false, nil
}
