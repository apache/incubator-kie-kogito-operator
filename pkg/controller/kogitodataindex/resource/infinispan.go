// Copyright 2019 Red Hat, Inc. and/or its affiliates
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
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/infinispan"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// updateInfinispanVars will update the infinispan environment variables based on a KEY=VALUE map
func updateInfinispanVars(container *corev1.Container, newVars map[string]string) {
	for k, v := range newVars {
		util.SetEnvVar(k, v, container)
	}
}

// getInfinispanVars will get the infinispan env vars from the container env
func getInfinispanVars(container corev1.Container) []corev1.EnvVar {
	envs := []corev1.EnvVar{}
	if &container == nil {
		return envs
	}

	for _, env := range container.Env {
		for _, infinispanKey := range managedInfinispanKeys {
			if env.Name == infinispanKey {
				envs = append(envs, env)
			}
		}
	}

	return envs
}

// fromInfinispanToStringMap will convert the InfinispanConnectionProperties into a Map of environment variables to be set in the container
func fromInfinispanToStringMap(infinispan v1alpha1.InfinispanConnectionProperties) map[string]string {
	propsmap := map[string]string{}

	if &infinispan == nil {
		return propsmap
	}

	if &infinispan.Credentials != nil && len(infinispan.Credentials.SecretName) > 0 {
		propsmap[infinispanEnvKeyUsername] = ""
		propsmap[infinispanEnvKeyPassword] = ""
		propsmap[infinispanEnvKeyUseAuth] = "true"
		propsmap[infinispanEnvKeyCredSecret] = infinispan.Credentials.SecretName
		if len(infinispan.SaslMechanism) == 0 {
			infinispan.SaslMechanism = v1alpha1.SASLPlain
		}
	} else {
		propsmap[infinispanEnvKeyUseAuth] = "false"
	}
	if len(infinispan.AuthRealm) > 0 {
		propsmap[infinispanEnvKeyAuthRealm] = infinispan.AuthRealm
	}
	if len(infinispan.SaslMechanism) > 0 {
		propsmap[infinispanEnvKeySasl] = string(infinispan.SaslMechanism)
	}
	if len(infinispan.ServiceURI) > 0 {
		propsmap[interfaceEnvKeyServiceURI] = infinispan.ServiceURI
	}

	log.Debugf("Infinispan properties created as %s", propsmap)
	return propsmap
}

// setInfinispanCredentialsSecret will bind Infinispan Credentials in the container using the given secret as a reference
func setInfinispanCredentialsSecret(infinispanProps v1alpha1.InfinispanConnectionProperties, secret *corev1.Secret, container *corev1.Container) {
	if &infinispanProps.Credentials != nil &&
		len(infinispanProps.Credentials.SecretName) > 0 &&
		secret != nil &&
		container != nil {
		util.SetEnvVarFromSecret(infinispanEnvKeyUsername, infinispan.SecretUsernameKey, secret, container)
		util.SetEnvVarFromSecret(infinispanEnvKeyPassword, infinispan.SecretPasswordKey, secret, container)
	}
}

// fetchInfinispanCredentials will fetch the secret defined in KogitoDataIndex.Spec.Infinispan.Credentials.Secret
func fetchInfinispanCredentials(instance *v1alpha1.KogitoDataIndex, client *client.Client) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	if len(instance.Spec.Infinispan.Credentials.SecretName) > 0 {
		secret.ObjectMeta = metav1.ObjectMeta{
			Name:      instance.Spec.Infinispan.Credentials.SecretName,
			Namespace: instance.Namespace,
		}
		exist, err := kubernetes.ResourceC(client).Fetch(secret)
		if err != nil {
			return nil, err
		}
		if !exist {
			return nil, nil
		}
	}

	return secret, nil
}
