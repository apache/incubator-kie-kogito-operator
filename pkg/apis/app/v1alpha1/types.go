/// Copyright 2019 Red Hat, Inc. and/or its affiliates
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

package v1alpha1

/*
Collection of common types among CRDs
*/

// Image is a definition of a Docker image
type Image struct {
	Domain    string `json:"domain,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
	Tag       string `json:"tag,omitempty"`
}

// InfinispanSaslMechanismType is the possible SASL Mechanism used during infinispan connection.
// For more information, see https://en.wikipedia.org/wiki/Simple_Authentication_and_Security_Layer#SASL_mechanisms.
type InfinispanSaslMechanismType string

const (
	// SASLPlain is the PLAIN type
	SASLPlain InfinispanSaslMechanismType = "PLAIN"
	// SASLDigestMD5 is the DIGEST-MD5 type
	SASLDigestMD5 InfinispanSaslMechanismType = "DIGEST-MD5"
)

// SecretCredentialsType is the data structure for specifying credentials within a Secret
type SecretCredentialsType struct {
	// +optional
	// SecretName is the name of the secret where the credentials are set
	SecretName string `json:"secretName,omitempty"`

	// +optional
	// UsernameKey is the key pointing to a value in a Secret holding the username value
	UsernameKey string `json:"usernameKey,omitempty"`

	// +optional
	// PasswordKey is the key pointing to a value in a Secret holding the password value
	PasswordKey string `json:"passwordKey,omitempty"`
}

// InfinispanConnectionProperties is the configuration needed for authenticating an Infinispan cluster
// For more information, see https://docs.jboss.org/infinispan/10.0/apidocs/org/infinispan/client/hotrod/configuration/package-summary.html#package.description
// +k8s:openapi-gen=true
type InfinispanConnectionProperties struct {
	// +optional
	Credentials SecretCredentialsType `json:"credentials,omitempty"`

	// UseAuth is set to true if the credentials are set. This also sets the property infinispan.client.hotrod.use_auth.
	// +optional
	UseAuth bool `json:"useAuth,omitempty"`

	// Name of the Infinispan authentication realm. This sets the property infinispan.client.hotrod.auth_realm.
	// +optional
	AuthRealm string `json:"authRealm,omitempty"`

	// +kubebuilder:validation:Enum=PLAIN;DIGEST-MD5
	// +optional
	// SaslMechanism defined for the authentication. This sets the property infinispan.client.hotrod.sasl_mechanism.
	SaslMechanism InfinispanSaslMechanismType `json:"saslMechanism,omitempty"`

	// +optional
	// URI to connect to the Infinispan cluster (can it be an internal service or external URI), for example, myinfinispan-cluster:11222
	URI string `json:"uri,omitempty"`

	// +optional
	// UseKogitoInfra flags if the instance will use a provided infrastructure by KogitoInfra CR.
	// Setting this to true will deploy a new KogitoInfra CR into the namespace that will install Infinispan via Infinispan Operator.
	// Infinispan Operator MUST be installed in the namespace for this to work. On OpenShift, OLM should install it for you.
	// If running on Kubernetes without OLM installed, please install Infinispan Operator first.
	// Set this to false and fill all other properties to provide your own infrastructure
	UseKogitoInfra bool `json:"useKogitoInfra,omitempty"`
}

// InfinispanAware defines a spec with InfinispanProperties awareness
type InfinispanAware interface {
	// GetInfinispanProperties ...
	GetInfinispanProperties() InfinispanConnectionProperties
	// SetInfinispanProperties ...
	SetInfinispanProperties(props InfinispanConnectionProperties)
	// AreInfinispanPropertiesBlank checks if the connection properties have been set
	AreInfinispanPropertiesBlank() bool
}

// InfinispanMeta defines a structure for specs that need InfinispanProperties integration
type InfinispanMeta struct {
	// +optional
	// Has the data used by the service to connect to the Infinispan cluster.
	InfinispanProperties InfinispanConnectionProperties `json:"infinispan,omitempty"`
}

// GetInfinispanProperties ...
func (i *InfinispanMeta) GetInfinispanProperties() InfinispanConnectionProperties {
	return i.InfinispanProperties
}

// SetInfinispanProperties ...
func (i *InfinispanMeta) SetInfinispanProperties(props InfinispanConnectionProperties) {
	i.InfinispanProperties = props
}

// AreInfinispanPropertiesBlank checks if the connection properties have been set
func (i *InfinispanMeta) AreInfinispanPropertiesBlank() bool {
	return &i.InfinispanProperties == nil ||
		&i.InfinispanProperties.Credentials == nil ||
		&i.InfinispanProperties.UseKogitoInfra == nil ||
		len(i.InfinispanProperties.URI) == 0
}

// KafkaConnectionProperties has the data needed to connect to a Kafka cluster
type KafkaConnectionProperties struct {
	// +optional
	// URI is the service URI to connect to the Kafka cluster, for example, my-cluster-kafka-bootstrap:9092
	ExternalURI string `json:"externalURI,omitempty"`

	// +optional
	// Instance is the Kafka instance to be used, for example, kogito-kafka
	Instance string `json:"instance,omitempty"`

	// +optional
	// UseKogitoInfra flags if the instance will use a provided infrastructure by KogitoInfra CR.
	// Setting this to true will configure a KogitoInfra CR to install Kafka via Strimzi Operator.
	// Strimzi Operator MUST be installed in the namespace for this to work. On OpenShift, OLM should install it for you.
	// If running on Kubernetes without OLM installed, please install Strimzi Operator first.
	// Set this to false and fill other properties to provide your own infrastructure
	UseKogitoInfra bool `json:"useKogitoInfra,omitempty"`
}

// KafkaAware defines a spec with KafkaProperties awareness
type KafkaAware interface {
	// GetKafkaProperties ...
	GetKafkaProperties() KafkaConnectionProperties
	// SetKafkaProperties ...
	SetKafkaProperties(props KafkaConnectionProperties)
	// AreKafkaPropertiesBlank checks if the connection properties have been set
	AreKafkaPropertiesBlank() bool
}

// KafkaMeta defines a structure for specs that need KafkaProperties integration
type KafkaMeta struct {
	// +optional
	// Has the data used by the service to connect to the Kafka cluster.
	KafkaProperties KafkaConnectionProperties `json:"kafka,omitempty"`
}

// GetKafkaProperties ...
func (k *KafkaMeta) GetKafkaProperties() KafkaConnectionProperties {
	return k.KafkaProperties
}

// SetKafkaProperties ...
func (k *KafkaMeta) SetKafkaProperties(props KafkaConnectionProperties) {
	k.KafkaProperties = props
}

// AreKafkaPropertiesBlank checks if the connection properties have been set
func (k *KafkaMeta) AreKafkaPropertiesBlank() bool {
	return len(k.KafkaProperties.ExternalURI) == 0 &&
		len(k.KafkaProperties.Instance) == 0 &&
		!k.KafkaProperties.UseKogitoInfra
}

// KeycloakConnectionProperties has the data needed to connect to a Keycloak cluster
type KeycloakConnectionProperties struct {
	// +optional
	// Keycloak
	Keycloak string `json:"keycloak,omitempty"`

	// +optional
	// KeycloakRealm
	KeycloakRealm string `json:"keycloakRealm,omitempty"`

	// +optional
	// AuthServerURL
	AuthServerURL string `json:"authServerUrl,omitempty"`

	// +optional
	// RealmName
	RealmName string `json:"realmName,omitempty"`

	// +optional
	// Labels
	Labels map[string]string `json:"labels,omitempty"`
}

// KeycloakMeta defines a structure for specs that need KeycloakProperties integration
type KeycloakMeta struct {
	// +optional
	// KeycloakProperties has the data used by the service to connect to the Keycloak cluster.
	KeycloakProperties KeycloakConnectionProperties `json:"keycloak,omitempty"`

	// +optional
	// EnableSecurity
	EnableSecurity bool `json:"enableSecurity,omitempty"`
}

// KeycloakAware defines a spec with KeycloakProperties awareness
type KeycloakAware interface {
	// GetKeycloakProperties ...
	GetKeycloakProperties() KeycloakConnectionProperties
	// SetKeycloakProperties ...
	SetKeycloakProperties(props KeycloakConnectionProperties)
}

// GetKeycloakProperties ...
func (k *KeycloakMeta) GetKeycloakProperties() KeycloakConnectionProperties {
	return k.KeycloakProperties
}

// SetKeycloakProperties ...
func (k *KeycloakMeta) SetKeycloakProperties(props KeycloakConnectionProperties) {
	k.KeycloakProperties = props
}
