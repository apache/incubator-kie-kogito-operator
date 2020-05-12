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

package resource

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createTestKogitoApp(runtime v1alpha1.RuntimeType) *v1alpha1.KogitoApp {
	return &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: v1alpha1.KogitoAppSpec{
			Runtime: runtime,
			Build:   &v1alpha1.KogitoAppBuildObject{},
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{
				Envs: []corev1.EnvVar{
					{
						Name:      "TEST_BOOTSTRAP_SERVERS",
						Value:     "",
						ValueFrom: nil,
					},
				},
			},
		},
	}
}

func createTestKogitoInfra() *v1alpha1.KogitoInfra {
	return &v1alpha1.KogitoInfra{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallInfinispan: true,
		},
		Status: v1alpha1.KogitoInfraStatus{
			Infinispan: v1alpha1.InfinispanInstallStatus{
				InfraComponentInstallStatusType: v1alpha1.InfraComponentInstallStatusType{
					Service: "test",
				},
				CredentialSecret: "test",
			},
			Kafka: v1alpha1.InfraComponentInstallStatusType{
				Service: "test",
			},
		},
	}
}
