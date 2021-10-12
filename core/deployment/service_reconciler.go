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

package deployment

import (
	"github.com/kiegroup/kogito-operator/core/framework"
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	"github.com/kiegroup/kogito-operator/core/operator"
	v12 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceReconciler ...
type ServiceReconciler interface {
	Reconcile() error
}

type serviceReconciler struct {
	operator.Context
	deployment     *v12.Deployment
	serviceHandler infrastructure.ServiceHandler
	deltaProcessor framework.DeltaProcessor
}

func newServiceReconciler(context operator.Context, deployment *v12.Deployment) ServiceReconciler {
	return &serviceReconciler{
		Context:        context,
		deployment:     deployment,
		serviceHandler: infrastructure.NewServiceHandler(context),
		deltaProcessor: framework.NewDeltaProcessor(context),
	}
}

func (i *serviceReconciler) Reconcile() error {

	// Create Required resource
	requestedResources, err := i.createRequiredResources()
	if err != nil {
		return err
	}

	// Get Deployed resource
	deployedResources, err := i.getDeployedResources()
	if err != nil {
		return err
	}

	// Process Delta
	if err = i.processDelta(requestedResources, deployedResources); err != nil {
		return err
	}

	return nil
}

func (i *serviceReconciler) createRequiredResources() (map[reflect.Type][]client.Object, error) {
	resources := make(map[reflect.Type][]client.Object)
	service := i.serviceHandler.CreateService(types.NamespacedName{Name: i.deployment.GetName(), Namespace: i.deployment.GetNamespace()})
	if err := framework.SetOwner(i.deployment, i.Scheme, service); err != nil {
		return nil, err
	}
	resources[reflect.TypeOf(v1.Service{})] = []client.Object{service}
	return resources, nil
}

func (i *serviceReconciler) getDeployedResources() (map[reflect.Type][]client.Object, error) {
	resources := make(map[reflect.Type][]client.Object)
	service, err := i.serviceHandler.FetchService(types.NamespacedName{Name: i.deployment.GetName(), Namespace: i.deployment.GetNamespace()})
	if err != nil {
		return nil, err
	}
	if service != nil {
		resources[reflect.TypeOf(v1.Service{})] = []client.Object{service}
	}
	return resources, nil
}

func (i *serviceReconciler) processDelta(requestedResources map[reflect.Type][]client.Object, deployedResources map[reflect.Type][]client.Object) (err error) {
	comparator := i.serviceHandler.GetComparator()
	_, err = i.deltaProcessor.ProcessDelta(comparator, requestedResources, deployedResources)
	return
}
