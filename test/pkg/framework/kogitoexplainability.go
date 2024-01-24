// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package framework

import (
	api "github.com/apache/incubator-kie-kogito-operator/apis"
	"github.com/apache/incubator-kie-kogito-operator/apis/app/v1beta1"
	"github.com/apache/incubator-kie-kogito-operator/core/kogitosupportingservice"
	"github.com/apache/incubator-kie-kogito-operator/test/pkg/config"
	bddtypes "github.com/apache/incubator-kie-kogito-operator/test/pkg/types"
)

// InstallKogitoExplainabilityService install the Kogito Explainability service
func InstallKogitoExplainabilityService(namespace string, installerType InstallerType, explainability *bddtypes.KogitoServiceHolder) error {
	// Persistence is already configured internally by the Explainability service, so we don't need to add any additional persistence step here.
	return InstallService(explainability, installerType, "explainability")
}

// WaitForKogitoExplainabilityService wait for Kogito Explainability to be deployed
func WaitForKogitoExplainabilityService(namespace string, replicas int, timeoutInMin int) error {
	return WaitForService(namespace, getExplainabilityServiceName(), replicas, timeoutInMin)
}

func getExplainabilityServiceName() string {
	return kogitosupportingservice.DefaultExplainabilityName
}

// GetKogitoExplainabilityResourceStub Get basic KogitoExplainability stub with all needed fields initialized
func GetKogitoExplainabilityResourceStub(namespace string, replicas int) *v1beta1.KogitoSupportingService {
	return &v1beta1.KogitoSupportingService{
		ObjectMeta: NewObjectMetadata(namespace, getExplainabilityServiceName()),
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType:       api.Explainability,
			KogitoServiceSpec: NewKogitoServiceSpec(int32(replicas), config.GetServiceImageTag(config.ExplainabilityImageType, config.EphemeralPersistenceType), kogitosupportingservice.DefaultExplainabilityImageName),
		},
	}
}
