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

// GetDefaultDataIndex gets the default Data Index instance
func GetDefaultDataIndex(namespace string) v1alpha1.KogitoDataIndex {
	return v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{Name: infrastructure.DefaultDataIndexName, Namespace: namespace},
		Spec: v1alpha1.KogitoDataIndexSpec{
			KogitoServiceSpec: defaultServiceSpec,
		},
		Status: v1alpha1.KogitoDataIndexStatus{KogitoServiceStatus: defaultServiceStatus},
	}
}

// GetDefaultJobsService gets the default Jobs Service instance
func GetDefaultJobsService(namespace string, enablePersistence bool, enableEvents bool) v1alpha1.KogitoJobsService {
	return v1alpha1.KogitoJobsService{
		ObjectMeta: metav1.ObjectMeta{Name: infrastructure.DefaultJobsServiceName, Namespace: namespace},
		Spec: v1alpha1.KogitoJobsServiceSpec{
			KogitoServiceSpec: defaultServiceSpec,
		},
		Status: v1alpha1.KogitoJobsServiceStatus{KogitoServiceStatus: defaultServiceStatus},
	}
}

// GetDefaultMgmtConsole gets the default Management Console instance
func GetDefaultMgmtConsole(namespace string) v1alpha1.KogitoMgmtConsole {
	return v1alpha1.KogitoMgmtConsole{
		ObjectMeta: metav1.ObjectMeta{Name: infrastructure.DefaultMgmtConsoleName, Namespace: namespace},
		Spec:       v1alpha1.KogitoMgmtConsoleSpec{KogitoServiceSpec: defaultServiceSpec},
		Status:     v1alpha1.KogitoMgmtConsoleStatus{KogitoServiceStatus: defaultServiceStatus},
	}
}
