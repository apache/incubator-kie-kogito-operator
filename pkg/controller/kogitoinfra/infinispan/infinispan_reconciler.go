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

package infinispan

import (
	"fmt"
	infinispanv1 "github.com/infinispan/infinispan-operator/pkg/apis/infinispan/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// InfraResource implementation of KogitoInfraResource
type InfraResource struct {
}

// GetWatchedObjects provide list of object that needs to be watched to maintain Infinispan kogitoInfra resource
func GetWatchedObjects() []framework.WatchedObjects {
	return []framework.WatchedObjects{
		{
			GroupVersion: infinispanv1.SchemeGroupVersion,
			AddToScheme:  infinispanv1.AddToScheme,
			Objects:      []runtime.Object{&infinispanv1.Infinispan{}},
		},
		{
			Objects: []runtime.Object{&corev1.Secret{}},
		},
	}
}

// Reconcile reconcile Kogito infra object
func (i *InfraResource) Reconcile(client *client.Client, instance *v1alpha1.KogitoInfra, scheme *runtime.Scheme) (requeue bool, resultErr error) {

	var infinispanInstance *infinispanv1.Infinispan

	// Step 1: check whether user has provided custom infinispan instance reference
	isCustomReferenceProvided := len(instance.Spec.Resource.Name) > 0
	if isCustomReferenceProvided {
		log.Debugf("Custom infinispan instance reference is provided")

		infinispanInstance, resultErr = loadDeployedInfinispanInstance(client, instance.Spec.Resource.Name, instance.Spec.Resource.Namespace)
		if resultErr != nil {
			return false, resultErr
		}
		if infinispanInstance == nil {
			return false, fmt.Errorf("custom Infinispan instance(%s) not found in namespace %s", instance.Spec.Resource.Name, instance.Spec.Resource.Namespace)
		}
	} else {
		// create/refer kogito-infinispan instance
		log.Debugf("Custom infinispan instance reference is not provided")

		// Verify Infinispan
		infinispanAvailable, err := infrastructure.IsInfinispanOperatorAvailable(client, instance.Namespace)
		if err != nil {
			return false, err
		}
		if !infinispanAvailable {
			err = fmt.Errorf("Infinispan operator is not available in the namespace %s, impossible to continue ", instance.Namespace)
			return false, err
		}

		// Step 1: check whether infinispan instance exist
		infinispanInstance, resultErr = loadDeployedInfinispanInstance(client, InstanceName, instance.Namespace)
		if resultErr != nil {
			return false, resultErr
		}

		if infinispanInstance == nil {
			// if not exist then create new Infinispan instance. Infinispan operator creates Infinispan instance, secret & service resource
			_, resultErr = createNewInfinispanInstance(client, InstanceName, instance.Namespace, instance, scheme)
			if resultErr != nil {
				return false, resultErr
			}
			return true, nil
		}
	}
	if resultErr := updateAppPropsInStatus(client, infinispanInstance, instance); resultErr != nil {
		return false, nil
	}
	if resultErr := updateEnvVarsInStatus(client, infinispanInstance, instance, scheme); resultErr != nil {
		return false, nil
	}
	return false, resultErr
}
