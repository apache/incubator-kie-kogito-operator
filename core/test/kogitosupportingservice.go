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

package test

import (
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/client"
	"github.com/kiegroup/kogito-cloud-operator/core/client/kubernetes"
	api2 "github.com/kiegroup/kogito-cloud-operator/core/test/api"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type fakeKogitoSupportingServiceHandler struct {
	client *client.Client
}

// CreateFakeKogitoSupportingServiceHandler ...
func CreateFakeKogitoSupportingServiceHandler(client *client.Client) api.KogitoSupportingServiceHandler {
	return &fakeKogitoSupportingServiceHandler{
		client: client,
	}
}

// FetchKogitoSupportingService provide kogito supporting service instance
func (k fakeKogitoSupportingServiceHandler) FetchKogitoSupportingService(key types.NamespacedName) (api.KogitoSupportingServiceInterface, error) {
	instance := &api2.KogitoSupportingServiceTest{}
	if exists, resultErr := kubernetes.ResourceC(k.client).FetchWithKey(key, instance); resultErr != nil {
		return nil, resultErr
	} else if !exists {
		return nil, nil
	} else {
		return instance, nil
	}
}

func (k fakeKogitoSupportingServiceHandler) FetchKogitoSupportingServiceList(namespace string) (api.KogitoSupportingServiceListInterface, error) {
	supportingServiceList := &api2.KogitoSupportingServiceTestList{}
	if err := kubernetes.ResourceC(k.client).ListWithNamespace(namespace, supportingServiceList); err != nil {
		return nil, err
	}
	return supportingServiceList, nil
}

// CreateFakeDataIndex ...
func CreateFakeDataIndex(namespace string) *api2.KogitoSupportingServiceTest {
	return createFakeKogitoSupportingServiceInstance("data-index", namespace, api.DataIndex)
}

// CreateFakeJobsService ...
func CreateFakeJobsService(namespace string) *api2.KogitoSupportingServiceTest {
	return createFakeKogitoSupportingServiceInstance("jobs-service", namespace, api.JobsService)
}

// CreateFakeMgmtConsole ...
func CreateFakeMgmtConsole(namespace string) *api2.KogitoSupportingServiceTest {
	return createFakeKogitoSupportingServiceInstance("mgmt-console", namespace, api.MgmtConsole)
}

// CreateFakeExplainabilityService ...
func CreateFakeExplainabilityService(namespace string) *api2.KogitoSupportingServiceTest {
	return createFakeKogitoSupportingServiceInstance("explainability-service", namespace, api.Explainability)
}

// CreateFakeTaskConsole ...
func CreateFakeTaskConsole(namespace string) *api2.KogitoSupportingServiceTest {
	return createFakeKogitoSupportingServiceInstance("task-console", namespace, api.TaskConsole)
}

// CreateFakeTrustyAIService ...
func CreateFakeTrustyAIService(namespace string) *api2.KogitoSupportingServiceTest {
	return createFakeKogitoSupportingServiceInstance("trusty-ai", namespace, api.TrustyAI)
}

// CreateFakeTrustyUIService ...
func CreateFakeTrustyUIService(namespace string) *api2.KogitoSupportingServiceTest {
	return createFakeKogitoSupportingServiceInstance("trusty-ui", namespace, api.TrustyUI)
}

func createFakeKogitoSupportingServiceInstance(name, namespace string, serviceType api.ServiceType) *api2.KogitoSupportingServiceTest {
	replicas := int32(1)
	return &api2.KogitoSupportingServiceTest{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: api2.KogitoSupportingServiceSpecTest{
			ServiceType: serviceType,
			KogitoServiceSpec: api.KogitoServiceSpec{
				Replicas: &replicas,
			},
		},
	}
}
