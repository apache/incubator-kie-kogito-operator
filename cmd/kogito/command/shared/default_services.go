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

package shared

import (
	"github.com/kiegroup/kogito-cloud-operator/api"
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/core/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/core/kogitosupportingservice"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var defaultReplicas = int32(1)
var defaultServiceStatus = v1beta1.KogitoServiceStatus{ConditionsMeta: v1beta1.ConditionsMeta{Conditions: []v1beta1.Condition{}}}

// GetDefaultDataIndex gets the default Data Index instance
func GetDefaultDataIndex(namespace string) v1beta1.KogitoSupportingService {
	return v1beta1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{Name: kogitosupportingservice.DefaultDataIndexName, Namespace: namespace},
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType: api.DataIndex,
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{
				Replicas: &defaultReplicas,
				Infra: []string{
					infrastructure.InfinispanInstanceName,
					infrastructure.KafkaInstanceName,
				},
			},
		},
		Status: v1beta1.KogitoSupportingServiceStatus{KogitoServiceStatus: defaultServiceStatus},
	}
}

// GetDefaultJobsService gets the default Jobs Service instance
func GetDefaultJobsService(namespace string) v1beta1.KogitoSupportingService {
	return v1beta1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{Name: kogitosupportingservice.DefaultJobsServiceName, Namespace: namespace},
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType: api.JobsService,
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{
				Replicas: &defaultReplicas,
			},
		},
		Status: v1beta1.KogitoSupportingServiceStatus{KogitoServiceStatus: defaultServiceStatus},
	}
}

// GetDefaultMgmtConsole gets the default Management Console instance
func GetDefaultMgmtConsole(namespace string) v1beta1.KogitoSupportingService {
	return v1beta1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{Name: kogitosupportingservice.DefaultMgmtConsoleName, Namespace: namespace},
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType: api.MgmtConsole,
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{
				Replicas: &defaultReplicas,
			},
		},
		Status: v1beta1.KogitoSupportingServiceStatus{KogitoServiceStatus: defaultServiceStatus},
	}
}

// GetDefaultPersistenceInfra provides kogitoInfra instance for Infinispan
func GetDefaultPersistenceInfra(namespace string) *v1beta1.KogitoInfra {
	return &v1beta1.KogitoInfra{
		ObjectMeta: metav1.ObjectMeta{
			Name:      infrastructure.InfinispanInstanceName,
			Namespace: namespace,
		},
		Spec: v1beta1.KogitoInfraSpec{
			Resource: v1beta1.Resource{
				Kind:       infrastructure.InfinispanAPIVersion,
				APIVersion: infrastructure.InfinispanKind,
			},
		},
	}
}

// GetDefaultMessagingInfra provides kogitoInfra instance for Kafka
func GetDefaultMessagingInfra(namespace string) *v1beta1.KogitoInfra {
	return &v1beta1.KogitoInfra{
		ObjectMeta: metav1.ObjectMeta{
			Name:      infrastructure.KafkaInstanceName,
			Namespace: namespace,
		},
		Spec: v1beta1.KogitoInfraSpec{
			Resource: v1beta1.Resource{
				Kind:       infrastructure.KafkaKind,
				APIVersion: infrastructure.KafkaAPIVersion,
			},
		},
	}
}
