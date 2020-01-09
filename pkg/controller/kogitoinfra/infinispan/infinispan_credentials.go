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
	"crypto/md5"
	"encoding/hex"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	yaml "gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"math/rand"
)

const (
	kogitoInfinispanUser = "developer"
	passwordSize         = 10
	passwordAllowedBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var (
	defaultInfinispanUsers = []string{kogitoInfinispanUser, "operator"}
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

// generateDefaultCredentials will generate random Kogito credentials to inject into Infinispan Operator
// Deprecated: not in use right now because of this issue: https://github.com/infinispan/infinispan-operator/issues/211
func generateDefaultCredentials() (yamlFile, fileMD5 string, err error) {
	credentials := make([]Credential, len(defaultInfinispanUsers))
	for i, user := range defaultInfinispanUsers {
		credentials[i].Password = generateRandomPassword()
		credentials[i].Username = user
	}
	identity := Identity{Credentials: credentials}

	data, err := yaml.Marshal(&identity)
	if err != nil {
		return "", "", err
	}
	return string(data), getMD5FromBytes(data), nil
}

// generateRandomPassword generates a random password to inject into infinispan identities
// Deprecated: not in use right now because of this issue: https://github.com/infinispan/infinispan-operator/issues/211
func generateRandomPassword() string {
	b := make([]byte, passwordSize)
	for i := range b {
		b[i] = passwordAllowedBytes[rand.Int63()%int64(len(passwordAllowedBytes))]
	}
	return string(b)
}

func getMD5FromBytes(data []byte) string {
	hash := md5.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

// hasKogitoUser will find if the kogitoUser is set in the identities file. If we can't decode the file for some reason, will return false
// something odd could be happened to the file, so let's others decide to regenerate the file. An error message will be displayed in logs
func hasKogitoUser(secretFileData []byte) bool {
	identity := &Identity{}
	err := yaml.Unmarshal(secretFileData, identity)
	if err != nil {
		log.Errorf("Error (%s) while trying to unmarshal secret file: %s", err, secretFileData)
		return false
	}
	for _, c := range identity.Credentials {
		if c.Username == kogitoInfinispanUser {
			return true
		}
	}
	return false
}

// getDeveloperCredential will return the credential to be used by internal services
func getDeveloperCredential(secret *v1.Secret) (*Credential, error) {
	secretFileData := secret.Data[identityFileName]
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
