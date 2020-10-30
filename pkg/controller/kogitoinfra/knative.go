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
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
)

// kafkaInfraResource for Knative resources reconciliation
type knativeInfraResource struct{}

// Reconcile ...
func (i *knativeInfraResource) Reconcile(client *client.Client, instance *v1alpha1.KogitoInfra, scheme *runtime.Scheme) (requeue bool, resultErr error) {

	if !infrastructure.IsKnativeEventingAvailable(client) {
		return false, errorForResourceAPINotFound(&instance.Spec.Resource)
	}

	if len(instance.Spec.Resource.Name) > 0 {
		var log = logger.GetLogger("kogitoinfra-knative-reconcile")
		ns := instance.Spec.Resource.Namespace
		if len(ns) == 0 {
			log.Debugf("Namespace not defined, fetching from the current ns: %s", instance.Namespace)
			ns = instance.Namespace
		}
		broker := eventingv1.Broker{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: instance.Spec.Resource.Name}}
		if exists, resultErr := kubernetes.ResourceC(client).Fetch(&broker); resultErr != nil {
			return false, resultErr
		} else if !exists {
			return false, errorForResourceNotFound(infrastructure.KnativeEventingBrokerKind, broker.Name, broker.Namespace)
		}
	} else {
		return false,
			fmt.Errorf("No Knative Eventing Broker resource defined in the KogitoInfra CR %s on namespace %s, impossible to continue ", instance.Name, instance.Namespace)
	}
	return false, nil
}
