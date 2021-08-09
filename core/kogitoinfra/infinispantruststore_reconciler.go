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

package kogitoinfra

import (
	"github.com/RHsyseng/operator-utils/pkg/resource"
	v1 "github.com/infinispan/infinispan-operator/pkg/apis/infinispan/v1"
	"github.com/kiegroup/kogito-operator/api"
	"github.com/kiegroup/kogito-operator/api/v1beta1"
	"github.com/kiegroup/kogito-operator/core/framework"
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	"github.com/kiegroup/kogito-operator/core/operator"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"software.sslmate.com/src/go-pkcs12"
)

const (
	truststoreSecretName = "kogito-infinispan-truststore"
	certMountPath        = operator.KogitoHomeDir + "/certs/infinispan"
	truststoreSecretKey  = "truststore.p12"
	truststoreMountPath  = certMountPath + "/" + truststoreSecretKey
)

type infinispanTrustStoreReconciler struct {
	infraContext       infraContext
	infinispanInstance *v1.Infinispan
	secretHandler      infrastructure.SecretHandler
}

func newInfinispanTrustStoreReconciler(context infraContext, infinispanInstance *v1.Infinispan) Reconciler {
	return &infinispanTrustStoreReconciler{
		infraContext:       context,
		infinispanInstance: infinispanInstance,
		secretHandler:      infrastructure.NewSecretHandler(context.Context),
	}
}

func (i *infinispanTrustStoreReconciler) Reconcile() (err error) {

	if !isInfinispanEncryptionEnabled(i.infinispanInstance) {
		return nil
	}
	// Create Required resource
	requestedResources, err := i.createRequiredResources()
	if err != nil {
		return
	}

	// Get Deployed resource
	deployedResources, err := i.getDeployedResources()
	if err != nil {
		return
	}

	// Process Delta
	if err = i.processDelta(requestedResources, deployedResources); err != nil {
		return err
	}

	secretReference := &v1beta1.SecretReference{
		Name:      truststoreSecretName,
		MountType: api.Volume,
		MountPath: truststoreMountPath,
		FileMode:  &framework.ModeForCertificates,
	}
	i.updateSecretReferenceInStatus(secretReference)
	return nil
}

func (i *infinispanTrustStoreReconciler) createRequiredResources() (map[reflect.Type][]resource.KubernetesResource, error) {
	resources := make(map[reflect.Type][]resource.KubernetesResource)

	ispnSecret, err := i.secretHandler.MustFetchSecret(types.NamespacedName{Name: i.infinispanInstance.Spec.Security.EndpointEncryption.CertSecretName, Namespace: i.infinispanInstance.GetNamespace()})
	if err != nil {
		return nil, err
	}
	trustStore, err := framework.CreatePKCS12TrustStoreFromSecret(ispnSecret, pkcs12.DefaultPassword, infinispanTLSSecretKey)
	if err != nil {
		return nil, err
	}

	kogitoInfraEncryptionSecret := &v12.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      truststoreSecretName,
			Namespace: i.infraContext.instance.GetNamespace(),
		},
		Type: v12.SecretTypeOpaque,
		Data: map[string][]byte{
			truststoreSecretKey: trustStore,
		},
	}
	if err := framework.SetOwner(i.infraContext.instance, i.infraContext.Scheme, kogitoInfraEncryptionSecret); err != nil {
		return resources, err
	}
	resources[reflect.TypeOf(v12.Secret{})] = []resource.KubernetesResource{kogitoInfraEncryptionSecret}
	return resources, nil
}

func (i *infinispanTrustStoreReconciler) getDeployedResources() (map[reflect.Type][]resource.KubernetesResource, error) {
	resources := make(map[reflect.Type][]resource.KubernetesResource)
	// fetch owned image stream
	deployedSecret, err := i.secretHandler.FetchSecret(types.NamespacedName{Name: truststoreSecretName, Namespace: i.infraContext.instance.GetNamespace()})
	if err != nil {
		return nil, err
	}
	if deployedSecret != nil {
		resources[reflect.TypeOf(v12.Secret{})] = []resource.KubernetesResource{deployedSecret}
	}
	return resources, nil
}

func (i *infinispanTrustStoreReconciler) processDelta(requestedResources map[reflect.Type][]resource.KubernetesResource, deployedResources map[reflect.Type][]resource.KubernetesResource) (err error) {
	comparator := i.secretHandler.GetComparator()
	deltaProcessor := infrastructure.NewDeltaProcessor(i.infraContext.Context)
	_, err = deltaProcessor.ProcessDelta(comparator, requestedResources, deployedResources)
	return err
}

func isInfinispanEncryptionEnabled(infinispanInstance *v1.Infinispan) bool {
	return !(*infinispanInstance.Spec.Security.EndpointAuthentication || infinispanInstance.Spec.Security.EndpointEncryption == nil || len(infinispanInstance.Spec.Security.EndpointEncryption.CertSecretName) == 0)
}

func (i *infinispanTrustStoreReconciler) updateSecretReferenceInStatus(secretReference *v1beta1.SecretReference) {
	instance := i.infraContext.instance
	secretReferences := append(instance.GetStatus().GetSecretReferences(), secretReference)
	instance.GetStatus().SetSecretReferences(secretReferences)
}
