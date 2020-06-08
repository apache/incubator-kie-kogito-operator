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
	yaml "gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
)

const (
	kogitoInfinispanUser = "developer"
)

// Identity is the struct for the secret holding the credential for the Infinispan server
type Identity struct {
	Credentials []Credential `yaml:"credentials"`
}

// Credential holds the information to authenticate into an infinispan server
type Credential struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// getDeveloperCredential will return the credential to be used by internal services
func getDeveloperCredential(secret *v1.Secret) (*Credential, error) {
	secretFileData := secret.Data[IdentityFileName]
	identity := &Identity{}
	if len(secretFileData) == 0 {
		// support for DataGrid operator based on Infinispan 0.3.x
		return &Credential{
			Username: string(secret.Data[infrastructure.InfinispanSecretUsernameKey]),
			Password: string(secret.Data[infrastructure.InfinispanSecretPasswordKey]),
		}, nil
	}
	err := yaml.Unmarshal(secretFileData, identity)
	if err != nil {
		return nil, err
	}
	for _, c := range identity.Credentials {
		if c.Username == kogitoInfinispanUser {
			return &c, nil
		}
	}
	return &Credential{}, nil
}
