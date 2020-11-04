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

	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
)

// knativeInfraReconciler for Knative resources reconciliation
type knativeInfraReconciler struct {
	targetContext
}

// Reconcile ...
func (i *knativeInfraReconciler) Reconcile() (requeue bool, resultErr error) {

	if !infrastructure.IsKnativeEventingAvailable(i.client) {
		return false, errorForResourceAPINotFound(&i.instance.Spec.Resource)
	}

	if len(i.instance.Spec.Resource.Name) > 0 {
		var log = logger.GetLogger("kogitoinfra-knative-reconcile")
		ns := i.instance.Spec.Resource.Namespace
		if len(ns) == 0 {
			log.Debugf("Namespace not defined, fetching from the current ns: %s", i.instance.Namespace)
			ns = i.instance.Namespace
		}
		broker := eventingv1.Broker{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: i.instance.Spec.Resource.Name}}
		if exists, resultErr := kubernetes.ResourceC(i.client).Fetch(&broker); resultErr != nil {
			return false, resultErr
		} else if !exists {
			return false, errorForResourceNotFound(infrastructure.KnativeEventingBrokerKind, broker.Name, broker.Namespace)
		}
	} else {
		return false,
			fmt.Errorf("No Knative Eventing Broker resource defined in the KogitoInfra CR %s on namespace %s, impossible to continue ", i.instance.Name, i.instance.Namespace)
	}
	return false, nil
}
