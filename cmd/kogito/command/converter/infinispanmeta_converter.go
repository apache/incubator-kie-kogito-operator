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

package converter

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultInfinispanSecretName  = "kogito-operator-infinispan-credentials"
	defaultInfinispanUsernameKey = "username"
	defaultInfinispanPasswordKey = "password"
)

// FromInfinispanFlagsToInfinispanMeta converts given InfinispanFlags into InfinispanMeta
func FromInfinispanFlagsToInfinispanMeta(cli *client.Client, namespace string, infinispanFlags *flag.InfinispanFlags, enablePersistence bool) (v1alpha1.InfinispanMeta, error) {
	infinispanMeta := v1alpha1.InfinispanMeta{}
	// configure Infinispan connection properties only when enablePersistence flag is true
	if enablePersistence {
		infinispanProperties, err := fromInfinispanFlagsToInfinispanProperties(cli, namespace, infinispanFlags)
		if err != nil {
			return infinispanMeta, err
		}
		infinispanMeta.InfinispanProperties = infinispanProperties
	}
	return infinispanMeta, nil
}

// fromInfinispanFlagsToInfinispanProperties converts given InfinispanFlags into InfinispanConnectionProperties
func fromInfinispanFlagsToInfinispanProperties(cli *client.Client, namespace string, infinispanFlags *flag.InfinispanFlags) (v1alpha1.InfinispanConnectionProperties, error) {
	log := context.GetDefaultLogger()
	infinispanProperties := v1alpha1.InfinispanConnectionProperties{}
	// If User provided Infinispan URI then configure connection properties to user define values else
	// set UseKogitoInfra to true so Infinispan will be automatically deployed via Infinispan Operator
	if len(infinispanFlags.URI) > 0 {
		if err := initializeUserDefineInfinispanProperties(cli, namespace, infinispanFlags, &infinispanProperties); err != nil {
			return infinispanProperties, err
		}
	} else {
		log.Info("infinispan-url not informed, Infinispan will be automatically deployed via Infinispan Operator")
		infinispanProperties.UseKogitoInfra = true
	}
	return infinispanProperties, nil
}

// initializeUserDefineInfinispanProperties set Infinispan connection properties to user define values
func initializeUserDefineInfinispanProperties(cli *client.Client, namespace string, infinispanFlags *flag.InfinispanFlags, infinispanProperties *v1alpha1.InfinispanConnectionProperties) error {
	log := context.GetDefaultLogger()
	log.Infof("infinispan-url informed. Infinispan will NOT be provisioned for you. Make sure that %s url is accessible from the cluster", infinispanFlags.URI)
	// If user and password are sent, create a secret to hold them and attach them to the CRD
	if err := createInfinispanSecret(cli, namespace, infinispanFlags.InfinispanUser, infinispanFlags.InfinispanPassword); err != nil {
		return err
	}
	infinispanProperties.Credentials = v1alpha1.SecretCredentialsType{
		SecretName:  defaultInfinispanSecretName,
		UsernameKey: defaultInfinispanUsernameKey,
		PasswordKey: defaultInfinispanPasswordKey,
	}
	infinispanProperties.UseAuth = true
	infinispanProperties.AuthRealm = infinispanFlags.AuthRealm

	infinispanSasl := infinispanFlags.InfinispanSasl
	if len(infinispanSasl) == 0 {
		infinispanSasl = string(v1alpha1.SASLPlain)
	}
	infinispanProperties.SaslMechanism = v1alpha1.InfinispanSaslMechanismType(infinispanSasl)

	infinispanProperties.URI = infinispanFlags.URI
	infinispanProperties.UseKogitoInfra = false
	return nil
}

// createInfinispanSecret create new secret resource to persist user provided username and password. If Secret with name
//`kogito-operator-infinispan-credentials` already exists then delete old secret resource and then create new resource
func createInfinispanSecret(cli *client.Client, namespace, username, password string) error {
	infinispanSecret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: defaultInfinispanSecretName, Namespace: namespace},
	}

	// Check if Secret resource with name `kogito-operator-infinispan-credentials` is already exists
	if exist, err := kubernetes.ResourceC(cli).Fetch(&infinispanSecret); err != nil {
		return fmt.Errorf("Error while trying to fetch for the Infinispan Credentials Secret: %s ", err)
	} else if exist {
		// If Secret already exists then delete old resource before creating new secret resource
		if err := kubernetes.ResourceC(cli).Delete(&infinispanSecret); err != nil {
			return fmt.Errorf("Error while deleting Infinispan Credentials Secret: %s ", err)
		}
	}

	infinispanSecret.StringData = map[string]string{
		defaultInfinispanUsernameKey: username,
		defaultInfinispanPasswordKey: password,
	}

	if err := kubernetes.ResourceC(cli).Create(&infinispanSecret); err != nil {
		return fmt.Errorf("Error while trying to create an Infinispan Secret credentials: %s ", err)
	}
	return nil
}
