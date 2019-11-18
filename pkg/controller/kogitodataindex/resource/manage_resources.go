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

package resource

import (
	"github.com/google/go-cmp/cmp"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// ManageResources will trigger updates on resources if needed
func ManageResources(instance *v1alpha1.KogitoDataIndex, resources *KogitoDataIndexResources, client *client.Client) error {
	if resources.StatefulSet != nil {
		if _, err := kubernetes.ResourceC(client).Fetch(resources.StatefulSet); err != nil {
			return err
		}

		replicaUpdate := ensureReplicas(instance, resources.StatefulSet)
		imgUpdate := ensureImage(instance, resources.StatefulSet)
		envUpdate := ensureEnvs(instance, resources.StatefulSet)
		resUpdate := ensureResources(instance, resources.StatefulSet)
		kafkaUpdate := ensureKafka(instance, resources.StatefulSet)
		infinispanUpdate, err := ensureInfinispan(instance, resources.StatefulSet, client)
		if err != nil {
			return err
		}

		if err := ensureProtoBufConfigMap(instance, resources.ProtoBufConfigMap, client); err != nil {
			return err
		}

		if replicaUpdate || imgUpdate || envUpdate || resUpdate || kafkaUpdate || infinispanUpdate {
			if err := kubernetes.ResourceC(client).Update(resources.StatefulSet); err != nil {
				return err
			}
		}
	}

	return nil
}

func ensureProtoBufConfigMap(instance *v1alpha1.KogitoDataIndex, cm *corev1.ConfigMap, client *client.Client) error {
	if cm == nil {
		return nil
	}

	files, err := getAllProtoFilesFromKogitoApps(client, instance.Namespace)
	if err != nil {
		log.Errorf("Error while fetching for protobuf files: %s ", err)
		return err
	}

	hashFiles := util.GenerateMD5Hash(files)
	hashCm := util.GenerateMD5Hash(cm.Data)

	if hashFiles != hashCm {
		log.Debugf("ProfoBuf Config Map will change from \n %s \n to \n %s", cm.Data, files)
		cm.Data = files
		if err := kubernetes.ResourceC(client).Update(cm); err != nil {
			log.Errorf("Error while updating configMap with new files: %s ", err)
			return err
		}
	}
	return nil
}

func ensureReplicas(instance *v1alpha1.KogitoDataIndex, statefulset *appsv1.StatefulSet) bool {
	size := instance.Spec.Replicas

	if *statefulset.Spec.Replicas != size {
		log.Debugf("Replicas changed to %s", size)
		statefulset.Spec.Replicas = &size
		return true
	}

	return false
}

func ensureImage(instance *v1alpha1.KogitoDataIndex, statefulset *appsv1.StatefulSet) bool {
	if len(statefulset.Spec.Template.Spec.Containers) > 0 {
		if statefulset.Spec.Template.Spec.Containers[0].Image != instance.Spec.Image {
			log.Debugf("Found difference in the deployment image (%s) was (%s)", instance.Spec.Image, statefulset.Spec.Template.Spec.Containers[0].Image)
			statefulset.Spec.Template.Spec.Containers[0].Image = instance.Spec.Image
			return true
		}
	}

	return false
}

func ensureEnvs(instance *v1alpha1.KogitoDataIndex, statefulset *appsv1.StatefulSet) bool {
	if len(statefulset.Spec.Template.Spec.Containers) > 0 {
		if instance.Spec.Env == nil {
			instance.Spec.Env = map[string]string{}
		}
		hasDiff := false
		removeManagedEnvVars(instance)
		envs := util.FromMapToEnvVar(instance.Spec.Env)
		managedEnvs := extractManagedEnvVars(&statefulset.Spec.Template.Spec.Containers[0])

		if !util.EnvVarCheck(envs, statefulset.Spec.Template.Spec.Containers[0].Env) {
			log.Debugf("Found difference in env vars (%s). Setting to %s", statefulset.Spec.Template.Spec.Containers[0].Env, envs)
			statefulset.Spec.Template.Spec.Containers[0].Env = envs
			hasDiff = true
		}
		// putting back managed envs
		statefulset.Spec.Template.Spec.Containers[0].Env = append(statefulset.Spec.Template.Spec.Containers[0].Env, managedEnvs...)
		return hasDiff
	}

	return false
}

func ensureResources(instance *v1alpha1.KogitoDataIndex, statefulset *appsv1.StatefulSet) bool {
	if len(statefulset.Spec.Template.Spec.Containers) > 0 {
		resourcesInstance := extractResources(instance)
		resourcesDeployment := statefulset.Spec.Template.Spec.Containers[0].Resources
		diff := cmp.Diff(resourcesDeployment, resourcesInstance)
		if diff != "" {
			log.Debugf("Found differences: '%s' in the resources (%s). Setting to %s", diff, resourcesDeployment, resourcesInstance)
			statefulset.Spec.Template.Spec.Containers[0].Resources = resourcesInstance
			return true
		}
	}

	return false
}

func ensureInfinispan(instance *v1alpha1.KogitoDataIndex, statefulset *appsv1.StatefulSet, client *client.Client) (bool, error) {
	if len(statefulset.Spec.Template.Spec.Containers) == 0 || &instance.Spec.Infinispan == nil {
		return false, nil
	}

	secret := &corev1.Secret{}
	if &instance.Spec.Infinispan.Credentials != nil {
		var err error
		secret, err = fetchInfinispanCredentials(instance, client)
		if err != nil {
			return false, err
		}
		if secret == nil && len(instance.Spec.Infinispan.Credentials.SecretName) > 0 {
			log.Warnf("Secret %s not found, skipping Infinispan credentials update", instance.Spec.Infinispan.Credentials.SecretName)
			return false, nil
		}
	}

	infinispanEnvs := fromInfinispanToStringMap(instance.Spec.Infinispan, secret)
	currentInfinispan := getInfinispanVars(statefulset.Spec.Template.Spec.Containers[0])

	if util.EnvVarCheck(currentInfinispan, util.FromMapToEnvVar(infinispanEnvs)) {
		return false, nil
	}

	log.Debugf("Encountered differences in the Infinispan properties: %s. Setting to %s.", currentInfinispan, infinispanEnvs)
	updateInfinispanVars(&statefulset.Spec.Template.Spec.Containers[0], infinispanEnvs)
	return true, nil
}

func ensureKafka(instance *v1alpha1.KogitoDataIndex, statefulset *appsv1.StatefulSet) bool {
	if &instance.Spec.Kafka == nil || len(statefulset.Spec.Template.Spec.Containers) == 0 {
		return false
	}

	updated := false
	for _, kafkaEnv := range managedKafkaKeys {
		currentURI := util.GetEnvVarFromContainer(kafkaEnv, statefulset.Spec.Template.Spec.Containers[0])
		if instance.Spec.Kafka.ServiceURI != currentURI {
			log.Debugf("Found differences in the Kafka ServiceURI (%s). Updating to '%s'.", currentURI, instance.Spec.Kafka.ServiceURI)
			util.SetEnvVar(kafkaEnv, instance.Spec.Kafka.ServiceURI, &statefulset.Spec.Template.Spec.Containers[0])
			updated = true
		}
	}

	return updated
}
