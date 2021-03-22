// Copyright 2021 Red Hat, Inc. and/or its affiliates
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

package v1beta1

// TLSKeyStore (see api.TLSKeyStoreInterface)
type TLSKeyStore struct {
	// ConfigMapName is the name of the ConfigMap that has the KeyStore file
	ConfigMapName string `json:"configMapName"`
	// PasswordSecretName is the Secret name for the password for the given KeyStore.
	// It's expected that the secret has a key named `keyStorePassword` if there's more than one key.
	PasswordSecretName string `json:"passwordSecretName,omitempty"`
}

// GetConfigMapName ...
func (t *TLSKeyStore) GetConfigMapName() string {
	return t.ConfigMapName
}

// SetConfigMapName ...
func (t *TLSKeyStore) SetConfigMapName(name string) {
	t.ConfigMapName = name
}

// GetPasswordSecretName ...
func (t *TLSKeyStore) GetPasswordSecretName() string {
	return t.PasswordSecretName
}

// SetPasswordSecretName ...
func (t *TLSKeyStore) SetPasswordSecretName(name string) {
	t.PasswordSecretName = name
}
