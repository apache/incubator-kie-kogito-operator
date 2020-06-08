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
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_getDeveloperCresdentials(t *testing.T) {
	secreteMap := make(map[string][]byte)
	secreteMap[infrastructure.InfinispanSecretPasswordKey] = []byte("password")
	secreteMap[infrastructure.InfinispanSecretUsernameKey] = []byte(kogitoInfinispanUser)
	infinispanSecret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      getInfinispanGeneratedSecretName()[0],
			Namespace: t.Name(),
		},
		Data: secreteMap,
	}
	credential, err := getDeveloperCredential(infinispanSecret)
	assert.NoError(t, err)
	assert.Equal(t, kogitoInfinispanUser, credential.Username)
	assert.Equal(t, "password", credential.Password)

}
