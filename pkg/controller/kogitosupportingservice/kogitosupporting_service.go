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

package kogitosupportingservice

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitosupportingservice/explainability"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitosupportingservice/jobsservice"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitosupportingservice/mgmtconsole"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitosupportingservice/trustyai"
	"k8s.io/apimachinery/pkg/types"
)

func fetchKogitoSupportingService(client *client.Client, name string, namespace string) (*v1alpha1.KogitoSupportingService, error) {
	log.Debugf("going to fetch deployed kogito supporting service instance %s in namespace %s", name, namespace)
	instance := &v1alpha1.KogitoSupportingService{}
	if exists, resultErr := kubernetes.ResourceC(client).FetchWithKey(types.NamespacedName{Name: name, Namespace: namespace}, instance); resultErr != nil {
		log.Errorf("Error occurs while fetching deployed kogito supporting service instance %s", name)
		return nil, resultErr
	} else if !exists {
		return nil, fmt.Errorf("kogito supporting service resource with name %s not found in namespace %s", name, namespace)
	} else {
		log.Debugf("Successfully fetch deployed kogito supporting reference %s", name)
		return instance, nil
	}
}

func getKogitoSupportingResource(instance *v1alpha1.KogitoSupportingService) SupportingServiceResource {
	log.Debugf("going to fetch related kogito infra resource for given infra instance : %s", instance.Name)
	switch instance.Spec.ServiceType {
	case v1alpha1.DataIndex:
		log.Debugf("Kogito Supporting Service reference is for Data Index")
		return &DataIndexSupportingServiceResource{}
	case v1alpha1.JobsService:
		log.Debugf("Kogito Supporting Service reference is for Jobs Service")
		return &jobsservice.SupportingServiceResource{}
	case v1alpha1.MgmtConsole:
		log.Debugf("Kogito Supporting Service reference is for Management Console")
		return &mgmtconsole.SupportingServiceResource{}
	case v1alpha1.Explainablity:
		log.Debugf("Kogito Supporting Service reference is for Explainability Service")
		return &explainability.SupportingServiceResource{}
	case v1alpha1.TrustyAI:
		log.Debugf("Kogito Supporting Service reference is for TrustyAI")
		return &trustyai.SupportingServiceResource{}
	case v1alpha1.TrustyUI:
		log.Debugf("Kogito Supporting Service reference is for TrustyUI")
	}
	return nil

}

func ensureSingletonService(client *client.Client, namespace string, resourceType v1alpha1.ServiceType) error {
	supportingServiceList := &v1alpha1.KogitoSupportingServiceList{}
	if err := kubernetes.ResourceC(client).ListWithNamespace(namespace, supportingServiceList); err != nil {
		return err
	}

	var kogitoSupportingService []v1alpha1.KogitoSupportingService
	for _, service := range supportingServiceList.Items {
		if service.Spec.ServiceType == resourceType {
			kogitoSupportingService = append(kogitoSupportingService, service)
		}
	}

	if len(kogitoSupportingService) > 1 {
		return fmt.Errorf("kogito Supporting Service(%s) already exists, please delete the duplicate before proceeding", resourceType)
	}
	return nil
}

func contains(s []v1alpha1.ServiceType, e v1alpha1.ServiceType) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
