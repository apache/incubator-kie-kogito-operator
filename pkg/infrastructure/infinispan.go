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

package infrastructure

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// InfinispanOperatorName is the Infinispan Operator default name
	InfinispanOperatorName = "infinispan-operator"
	infinispanServerGroup  = "infinispan.org"
	defaultInfinispanPort  = 11222
	// InfinispanSecretUsernameKey is the secret username key set in the linked secret
	InfinispanSecretUsernameKey = "username"
	// InfinispanSecretPasswordKey is the secret password key set in the linked secret
	InfinispanSecretPasswordKey = "password"
)

// IsInfinispanAvailable checks whether Infinispan CRD is available or not
func IsInfinispanAvailable(cli *client.Client) bool {
	return cli.HasServerGroup(infinispanServerGroup)
}

// IsInfinispanOperatorAvailable verify if Infinispan Operator is running in the given namespace and the CRD is available
func IsInfinispanOperatorAvailable(cli *client.Client, namespace string) (bool, error) {
	log.Debugf("Checking if Infinispan Operator is available in the namespace %s", namespace)
	// first check for CRD
	if IsInfinispanAvailable(cli) {
		log.Debugf("Infinispan CRDs available. Checking if Infinispan Operator is deployed in the namespace %s", namespace)
		// then check if there's an Infinispan Operator deployed
		deployment := &v1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: InfinispanOperatorName}}
		exists := false
		var err error
		if exists, err = kubernetes.ResourceC(cli).Fetch(deployment); err != nil {
			return false, nil
		}
		if exists {
			log.Debugf("Infinispan Operator is available in the namespace %s", namespace)
			return true, nil
		}
	} else {
		log.Debug("Couldn't find Infinispan CRDs")
	}
	log.Debugf("Looks like Infinispan Operator is not available in the namespace %s", namespace)
	return false, nil
}

// GetInfinispanServiceURI fetches for the Infinispan service linked with the given KogitoInfra and returns a formatted URI
func GetInfinispanServiceURI(cli *client.Client, infra *v1alpha1.KogitoInfra) (uri string, err error) {
	if &infra == nil || &infra.Status == nil || &infra.Status.Infinispan == nil {
		return "", nil
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: infra.Status.Infinispan.Service, Namespace: infra.Namespace},
	}
	exists := false
	if exists, err = kubernetes.ResourceC(cli).Fetch(service); err != nil {
		return "", err
	}

	if exists {
		for _, port := range service.Spec.Ports {
			if port.TargetPort.IntVal == defaultInfinispanPort {
				return fmt.Sprintf("%s:%d", service.Name, port.TargetPort.IntVal), nil
			}
		}
		return "", fmt.Errorf("Infinispan default port (%d) not found in service %s ", defaultInfinispanPort, service.Name)
	}

	return "", nil
}

// GetInfinispanCredentialsSecret will fetch for the secret created to hold Infinispan credentials
func GetInfinispanCredentialsSecret(cli *client.Client, infra *v1alpha1.KogitoInfra) (secret *corev1.Secret, err error) {
	err = nil
	if &infra == nil || &infra.Status == nil || &infra.Status.Infinispan == nil {
		return
	}
	secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: infra.Status.Infinispan.CredentialSecret, Namespace: infra.Namespace},
	}
	_, err = kubernetes.ResourceC(cli).Fetch(secret)
	return
}
