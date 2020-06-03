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
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultInfinispanSaslMechanism = v1alpha1.SASLPlain
)

// CreateInfinispanProperties creates Infinispan properties by reading information from the KogitoInfra
func CreateInfinispanProperties(cli *client.Client, kogitoInfra *v1alpha1.KogitoInfra, kogitoApp *v1alpha1.KogitoApp) (envs []corev1.EnvVar, appProps map[string]string, err error) {
	if kogitoApp != nil &&
		(kogitoInfra != nil && &kogitoInfra.Status != nil && &kogitoInfra.Status.Infinispan != nil) {
		uri, err := infrastructure.GetInfinispanServiceURI(cli, kogitoInfra)
		if err != nil {
			return nil, nil, err
		}
		secret, err := infrastructure.GetInfinispanCredentialsSecret(cli, kogitoInfra)
		if err != nil {
			return nil, nil, err
		}

		appProps = map[string]string{}

		// inject credentials
		vars := services.PropertiesInfinispanQuarkus
		if kogitoApp.Spec.Runtime == v1alpha1.SpringbootRuntimeType {
			vars = services.PropertiesInfinispanSpring
		}

		appProps[vars[services.AppPropInfinispanUseAuth]] = "true"
		appProps[vars[services.AppPropInfinispanServerList]] = uri
		appProps[vars[services.AppPropInfinispanSaslMechanism]] = string(defaultInfinispanSaslMechanism)

		envs = append(envs, corev1.EnvVar{Name: vars[services.EnvVarInfinispanUser], ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: secret.Name},
				Key:                  infrastructure.InfinispanSecretUsernameKey,
			},
		}})
		envs = append(envs, corev1.EnvVar{Name: vars[services.EnvVarInfinispanPassword], ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: secret.Name},
				Key:                  infrastructure.InfinispanSecretPasswordKey,
			},
		}})
	}
	return envs, appProps, nil
}
