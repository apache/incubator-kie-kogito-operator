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

package grafana

import (
	"fmt"

	grafana "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// APIVersion refers to Grafana APIVersion
	APIVersion = "integreatly.org/v1alpha1"
	// Kind refers to grafana Kind
	Kind = "Grafana"
)

// InfraResource implementation of KogitoInfraResource
type InfraResource struct {
}

// GetWatchedObjects provide list of object that needs to be watched to maintain Grafana kogitoInfra resource
func GetWatchedObjects() []framework.WatchedObjects {
	return []framework.WatchedObjects{
		{
			GroupVersion: grafana.SchemeGroupVersion,
			AddToScheme:  grafana.SchemeBuilder.AddToScheme,
			Objects:      []runtime.Object{&grafana.Grafana{}},
		},
	}
}

// Reconcile reconcile Kogito infra object
func (k *InfraResource) Reconcile(client *client.Client, instance *v1alpha1.KogitoInfra, scheme *runtime.Scheme) (requeue bool, resultErr error) {

	var grafanaInstance *grafana.Grafana

	// Step 1: check whether user has provided custom Grafana instance reference
	isCustomReferenceProvided := len(instance.Spec.Resource.Name) > 0
	if isCustomReferenceProvided {
		log.Debugf("Custom grafana instance reference is provided")
		namespace := instance.Spec.Resource.Namespace
		if len(namespace) == 0 {
			namespace = instance.Namespace
			log.Debugf("Namespace is not provided for custom resource, taking instance namespace(%s) as default", namespace)
		}
		grafanaInstance, resultErr = loadDeployedGrafanaInstance(client, instance.Spec.Resource.Name, namespace)
		if resultErr != nil {
			return false, resultErr
		}
		if grafanaInstance == nil {
			return false, fmt.Errorf("custom grafana instance(%s) not found in namespace %s", instance.Spec.Resource.Name, namespace)
		}
	} else {
		// create/refer kogito-grafana instance
		log.Debugf("Custom grafana instance reference is not provided")

		// Verify grafana
		grafanaAvailable, err := infrastructure.IsGrafanaOperatorAvailable(client, instance.Namespace)
		if err != nil {
			resultErr = fmt.Errorf("Grafana operator is not available in the namespace %s, impossible to continue ", instance.Namespace)
			return false, resultErr
		}
		if !grafanaAvailable {
			err = fmt.Errorf("Grafana operator is not available in the namespace %s, impossible to continue ", instance.Namespace)
			return false, err
		}

		// check whether grafana instance exist
		grafanaInstance, resultErr = loadDeployedGrafanaInstance(client, instanceName, instance.Namespace)
		if resultErr != nil {
			return false, resultErr
		}

		if grafanaInstance == nil {
			// if not exist then create new Grafana instance. Grafana operator creates Grafana instance, secret & service resource
			_, resultErr = createNewGrafanaInstance(client, instanceName, instance.Namespace, instance, scheme)
			if resultErr != nil {
				return false, resultErr
			}
			return true, nil
		}
	}
	return false, nil
}
