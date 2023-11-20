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
	api "github.com/kiegroup/kogito-operator/apis"
	"github.com/kiegroup/kogito-operator/apis/app/v1beta1"
	"github.com/kiegroup/kogito-operator/core/kogitosupportingservice"
	"github.com/kiegroup/kogito-operator/test/pkg/config"
	bddtypes "github.com/kiegroup/kogito-operator/test/pkg/types"
)

// InstallKogitoTrustyUI install the Kogito Management Console component
func InstallKogitoTrustyUI(installerType InstallerType, trustyUI *bddtypes.KogitoServiceHolder) error {
	return InstallService(trustyUI, installerType, "trusty-ui")
}

// WaitForKogitoTrustyUIService wait for Kogito Management Console to be deployed
func WaitForKogitoTrustyUIService(namespace string, replicas int, timeoutInMin int) error {
	return WaitForService(namespace, getTrustyUIServiceName(), replicas, timeoutInMin)
}

func getTrustyUIServiceName() string {
	return kogitosupportingservice.DefaultTrustyUIName
}

// GetKogitoTrustyUIResourceStub Get basic KogitoTrustyUI stub with all needed fields initialized
func GetKogitoTrustyUIResourceStub(namespace string, replicas int) *v1beta1.KogitoSupportingService {
	return &v1beta1.KogitoSupportingService{
		ObjectMeta: NewObjectMetadata(namespace, getTrustyUIServiceName()),
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType:       api.TrustyUI,
			KogitoServiceSpec: NewKogitoServiceSpec(int32(replicas), config.GetServiceImageTag(config.TrustyImageType, config.EphemeralPersistenceType), kogitosupportingservice.DefaultTrustyUIImageName),
		},
	}
}
