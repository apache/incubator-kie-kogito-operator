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
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/read"
	infinispan "github.com/infinispan/infinispan-operator/pkg/apis/infinispan/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	"k8s.io/apimachinery/pkg/api/errors"
	"reflect"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// InstanceName is the default name for the Infinispan provisioned instance
	InstanceName     = "kogito-infinispan"
	secretName       = "kogito-infinispan-credential"
	identityFileName = "identities.yaml"
	annotationKeyMD5 = "org.kie.kogito/infraInfinispanCredentialsFileHash"
	replicasSize     = 1
)

var log = logger.GetLogger("kogitoinfra_resource")

// GetDeployedResources will fetch for every resource already deployed using kogitoInfra instance as a reference
func GetDeployedResources(kogitoInfra *v1alpha1.KogitoInfra, cli *client.Client) (resources map[reflect.Type][]resource.KubernetesResource, err error) {
	if infrastructure.IsInfinispanAvailable(cli) {
		reader := read.New(cli.ControlCli).WithNamespace(kogitoInfra.Namespace).WithOwnerObject(kogitoInfra)
		// unfortunately the SecretList is buggy, so we have to fetch it manually: https://github.com/kubernetes-sigs/controller-runtime/issues/362
		resources, err = reader.ListAll(&infinispan.InfinispanList{})
		if err != nil {
			log.Warn("Failed to list deployed objects. ", err)
			return nil, err
		}
		// see comment above why we're fetching Secret manually https://github.com/kubernetes-sigs/controller-runtime/issues/362
		secretType := reflect.TypeOf(v1.Secret{})
		secret, err := reader.WithOwnerObject(kogitoInfra).WithNamespace(kogitoInfra.Namespace).Load(secretType, secretName)
		if err != nil && !errors.IsNotFound(err) {
			log.Warn("Failed to get deployed secret", err)
			return nil, err
		} else if err == nil {
			resources[secretType] = []resource.KubernetesResource{secret}
		}
	}
	return resources, nil
}

// CreateRequiredResources will create all resources needed for the deployment of Infinispan based on its operator.
// Return an empty array if `.Spec.InstallInfinispan` is set to `false`
func CreateRequiredResources(kogitoInfra *v1alpha1.KogitoInfra, cli *client.Client) (resources map[reflect.Type][]resource.KubernetesResource, err error) {
	resources = make(map[reflect.Type][]resource.KubernetesResource, 2)
	log.Debugf("Creating default resources for Infinispan installation for Kogito Infra on %s namespace", kogitoInfra.Namespace)
	if kogitoInfra.Spec.InstallInfinispan {
		// ignoring custom secrets for now: https://github.com/infinispan/infinispan-operator/issues/211
		/*
			secret := &v1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: kogitoInfra.Namespace, Name: secretName}}
			secretExists := false
			if secretExists, err = kubernetes.ResourceC(cli).Fetch(secret); err != nil {
				return resources, err
			} else if !secretExists ||
				len(secret.Data[identityFileName]) == 0 ||
				!hasKogitoUser(secret.Data[identityFileName]) {
				// we only generated a new one if there's no one created for us since each time we will have a new one because of the random password
				secret, err = newInfinispanCredentials(kogitoInfra)
				if err != nil {
					return resources, err
				}
			} else {
				// the md5 should correspond to the credential file set
				// if the user just changed the secret arbitrarily to add more users or to change the existent password,
				// we correct the md5 and the operator will change the secret correctly
				secret.Annotations[annotationKeyMD5] = getMD5FromBytes(secret.Data[identityFileName])
			}
		*/
		secret, err := newInfinispanLinkedSecret(kogitoInfra, cli)
		if err != nil {
			return nil, err
		}
		resources[reflect.TypeOf(v1.Secret{})] = []resource.KubernetesResource{secret}
		resources[reflect.TypeOf(infinispan.Infinispan{})] = []resource.KubernetesResource{newInfinispanResource(kogitoInfra)}
		log.Debugf("Requested objects created as %s", resources)
	}

	return resources, nil
}

func newInfinispanResource(kogitoInfra *v1alpha1.KogitoInfra) *infinispan.Infinispan {
	infinispanRes := &infinispan.Infinispan{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: kogitoInfra.Namespace,
			Name:      InstanceName,
		},
		Spec: infinispan.InfinispanSpec{
			Replicas: replicasSize,
			// ignoring generating secrets for now: https://github.com/infinispan/infinispan-operator/issues/211
			//Security: infinispan.InfinispanSecurity{
			//	EndpointSecretName: secret.Name,
			//},
		},
	}

	return infinispanRes
}

// newInfinispanCredentials will generate custom credentials for the Infinispan Operator
// Deprecated: not in use right now because of this issue: https://github.com/infinispan/infinispan-operator/issues/211
func newInfinispanCredentials(kogitoInfra *v1alpha1.KogitoInfra) (*v1.Secret, error) {
	credentialsFileData, MD5, err := generateDefaultCredentials()
	if err != nil {
		return nil, err
	}
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: kogitoInfra.Namespace,
			Annotations: map[string]string{
				annotationKeyMD5: MD5,
			},
		},
		Type: v1.SecretTypeOpaque,
		StringData: map[string]string{
			identityFileName: credentialsFileData,
		},
	}
	return secret, nil
}

// newInfinispanLinkedSecret will create a new secret based on the generated identity secret by the Infinispan Operator
// this secret will be used later by any client services in the namespace to connect to the Infinispan instance
func newInfinispanLinkedSecret(kogitoInfra *v1alpha1.KogitoInfra, cli *client.Client) (*v1.Secret, error) {
	infinispanSecret, err := getOperatorGeneratedSecret(kogitoInfra, cli)
	if err != nil {
		return nil, err
	}
	credentials, err := getDeveloperCredential(infinispanSecret)
	if err != nil {
		return nil, err
	}
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: kogitoInfra.Namespace,
		},
		Type: v1.SecretTypeOpaque,
		StringData: map[string]string{
			infrastructure.InfinispanSecretUsernameKey: credentials.Username,
			infrastructure.InfinispanSecretPasswordKey: credentials.Password,
		},
	}
	return secret, nil
}
