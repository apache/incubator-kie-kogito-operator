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
	framework2 "github.com/apache/incubator-kie-kogito-operator/core/framework"
	"github.com/apache/incubator-kie-kogito-operator/core/kogitosupportingservice"
	"github.com/apache/incubator-kie-kogito-operator/test/pkg/config"
	bddtypes "github.com/apache/incubator-kie-kogito-operator/test/pkg/types"
)

// InstallKogitoDataIndexService install the Kogito Data Index service
func InstallKogitoDataIndexService(namespace string, installerType InstallerType, dataIndex *bddtypes.KogitoServiceHolder) error {
	// Persistence is already configured internally by the Data Index service, so we don't need to add any additional persistence step here.
	return InstallService(dataIndex, installerType, "data-index")
}

// WaitForKogitoDataIndexService wait for Kogito Data Index to be deployed
func WaitForKogitoDataIndexService(namespace string, replicas int, timeoutInMin int) error {
	if err := WaitForDeploymentRunning(namespace, getDataIndexServiceName(), replicas, timeoutInMin); err != nil {
		return err
	}

	// Data Index can be restarted after the deployment of KogitoRuntime, so 2 pods can run in parallel for a while.
	// We need to wait for only one (wait until the old one is deleted)
	return WaitForPodsWithLabel(namespace, framework2.LabelAppKey, getDataIndexServiceName(), replicas, timeoutInMin)
}

func getDataIndexServiceName() string {
	return kogitosupportingservice.DefaultDataIndexName
}

// GetKogitoDataIndexResourceStub Get basic KogitoDataIndex stub with all needed fields initialized
func GetKogitoDataIndexResourceStub(namespace string, replicas int) *v1beta1.KogitoSupportingService {
	return &v1beta1.KogitoSupportingService{
		ObjectMeta: NewObjectMetadata(namespace, getDataIndexServiceName()),
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType: api.DataIndex,
			// This should be changed to `ephemeral` once inmemory data-index is available
			KogitoServiceSpec: NewKogitoServiceSpec(int32(replicas), config.GetServiceImageTag(config.DataIndexImageType, config.InfinispanPersistenceType), kogitosupportingservice.DefaultDataIndexImageName),
		},
	}
}
