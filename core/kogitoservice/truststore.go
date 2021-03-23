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
	trustStoreVolumeName         = "custom-truststore"
)

// TrustStoreHandler takes care of mounting the custom TrustStore in a given Deployment based on api.KogitoService spec
type TrustStoreHandler interface {
	MountTrustStore(deployment *appsv1.Deployment, service api.KogitoService) error
}

// NewTrustStoreHandler creates a new TrustStoreHandler with the given context
func NewTrustStoreHandler(context *operator.Context) TrustStoreHandler {
	return &trustStoreHandler{
		context: context,
	}
}

type trustStoreHandler struct {
	context *operator.Context
}

// MountTrustStore mounts the given custom TrustStore based on api.KogitoService
func (t *trustStoreHandler) MountTrustStore(deployment *appsv1.Deployment, service api.KogitoService) error {
	if len(service.GetSpec().GetTrustStore().GetConfigMapName()) == 0 {
		return nil
	}

	if err := t.mapTrustStorePassword(deployment, service); err != nil {
		return err
	}

	if err := t.mountTrustStoreFile(deployment, service); err != nil {
		return err
	}

	return nil
}

func (t *trustStoreHandler) mountTrustStoreFile(deployment *appsv1.Deployment, service api.KogitoService) error {
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      service.GetSpec().GetTrustStore().GetConfigMapName(),
			Namespace: service.GetNamespace(),
		},
	}
	if exists, err := kubernetes.ResourceC(t.context.Client).Fetch(cm); err != nil {
		return err
	} else if !exists {
		return errorForTrustStoreMount("Failed to find ConfigMap named " + cm.Name + " in the namespace " + cm.Namespace)
	}
	if len(cm.BinaryData) != 1 {
		return errorForTrustStoreMount("Failed to mount Truststore. ConfigMap " + cm.Name + " must have only one file.")
	}
	trustStoreFileName := framework.GetFirstConfigMapBinaryKey(cm)
	trustStoreVolume := v1.Volume{
		Name: trustStoreVolumeName,
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{Name: cm.Name},
				Items:                []v1.KeyToPath{{Key: trustStoreFileName, Path: trustStoreFileName}},
				DefaultMode:          &framework.ModeForCertificates,
			},
		},
	}
	trustStoreMount := v1.VolumeMount{
		Name:      trustStoreVolumeName,
		MountPath: trustStoreMountPath + "/" + trustStoreFileName,
		SubPath:   trustStoreFileName,
	}

	framework.AddVolumeToDeployment(deployment, trustStoreMount, trustStoreVolume)
	deployment.Spec.Template.Spec.Containers[0].Env = framework.EnvOverride(deployment.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{
		Name:  trustStoreEnvVarCertFileName,
		Value: trustStoreFileName,
	})

	return nil
}

func (t *trustStoreHandler) mapTrustStorePassword(deployment *appsv1.Deployment, service api.KogitoService) error {
	if len(service.GetSpec().GetTrustStore().GetPasswordSecretName()) > 0 {
		secret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: service.GetNamespace(),
				Name:      service.GetSpec().GetTrustStore().GetPasswordSecretName(),
			},
		}
		if exists, err := kubernetes.ResourceC(t.context.Client).Fetch(secret); err != nil {
			return err
		} else if !exists {
			return errorForTrustStoreMount("Failed to find Secret named " + secret.Name + " in the namespace " + secret.Namespace)
		}
		if _, ok := secret.Data[trustStoreSecretKey]; !ok {
			return errorForTrustStoreMount("Failed to find a key named " + trustStoreSecretKey + " in the Secret " + secret.Name + " for the container's Truststore")
		}
		deployment.Spec.Template.Spec.Containers[0].Env = framework.EnvOverride(deployment.Spec.Template.Spec.Containers[0].Env,
			framework.CreateSecretEnvVar(trustStoreEnvVarPassword, secret.Name, trustStoreSecretKey))
	}
	return nil
}
