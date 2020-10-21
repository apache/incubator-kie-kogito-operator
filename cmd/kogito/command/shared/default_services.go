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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var defaultReplicas = int32(1)
var defaultServiceStatus = v1alpha1.KogitoServiceStatus{ConditionsMeta: v1alpha1.ConditionsMeta{Conditions: []v1alpha1.Condition{}}}
var defaultServiceSpec = v1alpha1.KogitoServiceSpec{
	Replicas: &defaultReplicas,
	Infra: []string{
		infrastructure.InfinispanInstanceName,
		infrastructure.KafkaInstanceName,
	},
}

// GetDefaultDataIndex gets the default Data Index instance
func GetDefaultDataIndex(namespace string) v1alpha1.KogitoSupportingService {
	return v1alpha1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{Name: infrastructure.DefaultDataIndexName, Namespace: namespace},
		Spec: v1alpha1.KogitoSupportingServiceSpec{
			ServiceType:       v1alpha1.DataIndex,
			KogitoServiceSpec: defaultServiceSpec,
		},
		Status: v1alpha1.KogitoSupportingServiceStatus{KogitoServiceStatus: defaultServiceStatus},
	}
}

// GetDefaultJobsService gets the default Jobs Service instance
func GetDefaultJobsService(namespace string) v1alpha1.KogitoSupportingService {
	return v1alpha1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{Name: infrastructure.DefaultJobsServiceName, Namespace: namespace},
		Spec: v1alpha1.KogitoSupportingServiceSpec{
			ServiceType:       v1alpha1.JobsService,
			KogitoServiceSpec: defaultServiceSpec,
		},
		Status: v1alpha1.KogitoSupportingServiceStatus{KogitoServiceStatus: defaultServiceStatus},
	}
}

// GetDefaultMgmtConsole gets the default Management Console instance
func GetDefaultMgmtConsole(namespace string) v1alpha1.KogitoSupportingService {
	return v1alpha1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{Name: infrastructure.DefaultMgmtConsoleName, Namespace: namespace},
		Spec: v1alpha1.KogitoSupportingServiceSpec{
			ServiceType:       v1alpha1.MgmtConsole,
			KogitoServiceSpec: defaultServiceSpec,
		},
		Status: v1alpha1.KogitoSupportingServiceStatus{KogitoServiceStatus: defaultServiceStatus},
	}
}

// GetDefaultPersistenceInfra provides kogitoInfra instance for Infinispan
func GetDefaultPersistenceInfra(namespace string) *v1alpha1.KogitoInfra {
	return &v1alpha1.KogitoInfra{
		ObjectMeta: metav1.ObjectMeta{
			Name:      infrastructure.InfinispanInstanceName,
			Namespace: namespace,
		},
		Spec: v1alpha1.KogitoInfraSpec{
			Resource: v1alpha1.Resource{
				Kind:       infrastructure.InfinispanAPIVersion,
				APIVersion: infrastructure.InfinispanKind,
			},
		},
	}
}

// GetDefaultMessagingInfra provides kogitoInfra instance for Kafka
func GetDefaultMessagingInfra(namespace string) *v1alpha1.KogitoInfra {
	return &v1alpha1.KogitoInfra{
		ObjectMeta: metav1.ObjectMeta{
			Name:      infrastructure.KafkaInstanceName,
			Namespace: namespace,
		},
		Spec: v1alpha1.KogitoInfraSpec{
			Resource: v1alpha1.Resource{
				Kind:       infrastructure.KafkaKind,
				APIVersion: infrastructure.KafkaAPIVersion,
			},
		},
	}
}
