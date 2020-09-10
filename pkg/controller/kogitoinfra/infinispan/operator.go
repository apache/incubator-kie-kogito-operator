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

package infinispan

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	infinispanOperatorGeneratedSecret         = "%s-generated-secret"
	infinispanOperatorAppRealmGeneratedSecret = "%s-app-generated-secret"
)

// getOperatorGeneratedSecret will fetch for the generated secret created by the Infinispan Operator
func getOperatorGeneratedSecret(cli *client.Client, namespace string) (*corev1.Secret, error) {
	secretNames := getInfinispanGeneratedSecretName()
	for _, secretName := range secretNames {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      secretName,
			},
			Data: map[string][]byte{},
		}

		if exists, err := kubernetes.ResourceC(cli).Fetch(secret); err != nil {
			return nil, err
		} else if exists {
			return secret, nil
		}
	}
	// return the supposed generated default one, in the next reconcile phase will get the managed by the Operator
	return createEmptySecret(secretNames[0], namespace), nil
}

// getInfinispanGeneratedSecretName gets the formatted name for the generated Infinispan Operator secret
func getInfinispanGeneratedSecretName() []string {
	return []string{
		fmt.Sprintf(infinispanOperatorGeneratedSecret, InstanceName),
		fmt.Sprintf(infinispanOperatorAppRealmGeneratedSecret, InstanceName),
	}
}

func createEmptySecret(secretName, secretNamespace string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: secretNamespace,
			Name:      secretName,
		},
		Data: map[string][]byte{},
	}
}
