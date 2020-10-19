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
	infinispan "github.com/infinispan/infinispan-operator/pkg/apis/infinispan/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// InstanceName is the default name for the Infinispan provisioned instance
	secretName   = "kogito-infinispan-credential"
	replicasSize = 1
)

var log = logger.GetLogger("kogitoinfinispan_resource")

func loadDeployedInfinispanInstance(cli *client.Client, instanceName string, namespace string) (*infinispan.Infinispan, error) {
	log.Debug("fetching deployed kogito infinispan instance")
	infinispanInstance := &infinispan.Infinispan{}
	if exits, err := kubernetes.ResourceC(cli).FetchWithKey(types.NamespacedName{Name: instanceName, Namespace: namespace}, infinispanInstance); err != nil {
		log.Error("Error occurs while fetching kogito infinispan instance")
		return nil, err
	} else if !exits {
		log.Debug("Kogito infinispan instance is not exists")
		return nil, nil
	} else {
		log.Debug("Kogito infinispan instance found")
		return infinispanInstance, nil
	}
}

func createNewInfinispanInstance(cli *client.Client, name string, namespace string, instance *v1alpha1.KogitoInfra, scheme *runtime.Scheme) (*infinispan.Infinispan, error) {
	log.Debug("Going to create kogito infinispan instance")
	infinispanRes := &infinispan.Infinispan{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: infinispan.InfinispanSpec{
			Replicas: replicasSize,
		},
	}
	if err := controllerutil.SetOwnerReference(instance, infinispanRes, scheme); err != nil {
		return nil, err
	}
	if err := kubernetes.ResourceC(cli).Create(infinispanRes); err != nil {
		log.Error("Error occurs while creating kogito infinispan instance")
		return nil, err
	}
	log.Debug("Kogito infinispan instance created successfully")
	return infinispanRes, nil
}

func loadCustomKogitoInfinispanSecret(cli *client.Client, namespace string) (*v1.Secret, error) {
	log.Debugf("Fetching %s ", secretName)
	secret := &v1.Secret{}
	if exits, err := kubernetes.ResourceC(cli).FetchWithKey(types.NamespacedName{Name: secretName, Namespace: namespace}, secret); err != nil {
		log.Errorf("Error occurs while fetching %s", secretName)
		return nil, err
	} else if !exits {
		log.Errorf("%s not found", secretName)
		return nil, nil
	} else {
		log.Debugf("%s successfully fetched", secretName)
		return secret, nil
	}
}

func createCustomKogitoInfinispanSecret(cli *client.Client, namespace string, infinispanInstance *infinispan.Infinispan, instance *v1alpha1.KogitoInfra, scheme *runtime.Scheme) (*v1.Secret, error) {
	log.Debugf("Creating new secret %s", secretName)

	credentials, err := infrastructure.GetInfinispanCredential(cli, infinispanInstance)
	if err != nil {
		return nil, err
	}
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Type: v1.SecretTypeOpaque,
	}
	if credentials != nil {
		secret.StringData = map[string]string{
			infrastructure.InfinispanSecretUsernameKey: credentials.Username,
			infrastructure.InfinispanSecretPasswordKey: credentials.Password,
		}
	}
	if err := controllerutil.SetOwnerReference(instance, secret, scheme); err != nil {
		return nil, err
	}
	if err := kubernetes.ResourceC(cli).Create(secret); err != nil {
		log.Errorf("Error occurs while creating %s", secret)
		return nil, err
	}
	log.Debug("%s successfully created", secret)
	return secret, nil
}
