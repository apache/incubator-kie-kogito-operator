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

package kogitoservice

import (
	"github.com/kiegroup/kogito-operator/api"
	"github.com/kiegroup/kogito-operator/core/client/kubernetes"
	"github.com/kiegroup/kogito-operator/core/framework"
	"github.com/kiegroup/kogito-operator/core/operator"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	trustStoreMountPath          = operator.KogitoHomeDir + "/certs/custom-truststore"
	trustStoreSecretKey          = "keyStorePassword"
	trustStoreEnvVarPassword     = "CUSTOM_TRUSTSTORE_PASSWORD"
	trustStoreEnvVarCertFileName = "CUSTOM_TRUSTSTORE"
	trustStoreVolumeName         = "trustStore"
)

// TODO: refactor in a handler with private functions and move to infra

// MountTrustStore ...
func MountTrustStore(context *operator.Context, deployment *appsv1.Deployment, service api.KogitoService) error {
	if len(service.GetSpec().GetTrustStore().GetConfigMapName()) == 0 {
		return nil
	}

	// key
	if len(service.GetSpec().GetTrustStore().GetPasswordSecretName()) > 0 {
		secret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: service.GetNamespace(),
				Name:      service.GetSpec().GetTrustStore().GetPasswordSecretName(),
			},
		}
		if exists, err := kubernetes.ResourceC(context.Client).Fetch(secret); err != nil {
			return err
		} else if !exists {
			return errorForTrustStoreMount("Failed to find Secret named " + secret.Name + " in the namespace " + secret.Namespace)
		}
		if _, ok := secret.StringData[trustStoreSecretKey]; !ok {
			return errorForTrustStoreMount("Failed to find a key named " + trustStoreSecretKey + " in the Secret " + secret.Name + " for the container's Truststore")
		}
		framework.EnvOverride(deployment.Spec.Template.Spec.Containers[0].Env,
			framework.CreateSecretEnvVar(trustStoreEnvVarPassword, secret.Name, trustStoreSecretKey))
	}

	// truststore
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      service.GetSpec().GetTrustStore().GetConfigMapName(),
			Namespace: service.GetNamespace(),
		},
	}
	if exists, err := kubernetes.ResourceC(context.Client).Fetch(cm); err != nil {
		return err
	} else if !exists {
		return errorForTrustStoreMount("Failed to find ConfigMap named " + cm.Name + " in the namespace " + cm.Namespace)
	}
	if len(cm.Data) == 0 {
		return errorForTrustStoreMount("Failed to mount Truststore. ConfigMap " + cm.Name + " has no data")
	}
	if len(cm.Data) > 1 {
		return errorForTrustStoreMount("Failed to mount Truststore. ConfigMap " + cm.Name + " has more than one key. Truststore ConfigMap must have only one file")
	}

	trustStoreVolume := v1.Volume{
		Name: trustStoreVolumeName,
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: cm.Name,
				},
				Items: []v1.KeyToPath{
					{
						// TODO: get the key
						Key:  "",
						Path: "",
					},
				},
				DefaultMode: &framework.ModeForCertificates,
			},
		},
	}
	trustStoreMount := v1.VolumeMount{
		Name:      trustStoreVolumeName,
		MountPath: trustStoreMountPath,
		ReadOnly:  true,
	}

	// TODO: add to framework.AddVolumeToDeployment(volume, mount)
	deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, trustStoreVolume)
	deployment.Spec.Template.Spec.Containers[0].VolumeMounts = append(deployment.Spec.Template.Spec.Containers[0].VolumeMounts, trustStoreMount)

	framework.EnvOverride(deployment.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{
		Name: trustStoreEnvVarCertFileName,
		// TODO: ConfigMapKey
		Value: "",
	})

	return nil
}
