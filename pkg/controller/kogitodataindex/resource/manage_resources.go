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
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"reflect"
)

/* TODO : should be rethinked by KOGITO-601 */

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
		kafkaUpdate, err := ensureKafka(instance, resources.StatefulSet, client)
		if err != nil {
			return err
		}
		infinispanUpdate, err := ensureInfinispan(instance, resources.StatefulSet, client)
		if err != nil {
			return err
		}

		if err := ensureKafkaTopics(instance, resources.KafkaTopics, client); err != nil {
			return err
		}
		volumeUpdate, err := ensureVolumes(resources, client)
		if err != nil {
			return err
		}
		if replicaUpdate || imgUpdate || envUpdate || resUpdate || kafkaUpdate || infinispanUpdate || volumeUpdate {
			if err := kubernetes.ResourceC(client).Update(resources.StatefulSet); err != nil {
				return err
			}
		}
	}

	return nil
}

func ensureVolumes(resources *KogitoDataIndexResources, cli *client.Client) (bool, error) {
	if len(resources.StatefulSet.Spec.Template.Spec.Containers) == 0 {
		return false, nil
	}
	copyss := resources.StatefulSet.DeepCopy()
	copyss.Spec.Template.Spec.Volumes = nil
	copyss.Spec.Template.Spec.Containers[0].VolumeMounts = nil
	if err := mountProtoBufConfigMaps(copyss, cli); err != nil {
		return false, err
	}
	if !reflect.DeepEqual(copyss.Spec.Template.Spec.Volumes, resources.StatefulSet.Spec.Template.Spec.Volumes) {
		resources.StatefulSet = copyss
		return true, nil
	}
	return false, nil
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
		envs := framework.MapToEnvVar(instance.Spec.Env)
		managedEnvs := extractManagedEnvVars(&statefulset.Spec.Template.Spec.Containers[0])

		if !framework.EnvVarCheck(envs, statefulset.Spec.Template.Spec.Containers[0].Env) {
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
	if len(statefulset.Spec.Template.Spec.Containers) == 0 || &instance.Spec.InfinispanProperties == nil {
		return false, nil
	}

	secret := &corev1.Secret{}
	if &instance.Spec.InfinispanProperties.Credentials != nil {
		var err error
		secret, err = infrastructure.FetchInfinispanCredentials(&instance.Spec, instance.Namespace, client)
		if err != nil {
			return false, err
		}
		if secret == nil && len(instance.Spec.InfinispanProperties.Credentials.SecretName) > 0 {
			log.Warnf("Secret %s not found, skipping Infinispan credentials update", instance.Spec.InfinispanProperties.Credentials.SecretName)
			return false, nil
		}
	}

	infinispanEnvs := infrastructure.FromInfinispanToStringMap(instance.Spec.InfinispanProperties)
	currentInfinispan := infrastructure.GetInfinispanVars(statefulset.Spec.Template.Spec.Containers[0])

	if framework.EnvVarCheck(currentInfinispan, framework.MapToEnvVar(infinispanEnvs)) {
		return false, nil
	}

	log.Debugf("Encountered differences in the Infinispan properties: %s. Setting to %s.", currentInfinispan, infinispanEnvs)
	updateInfinispanVars(&statefulset.Spec.Template.Spec.Containers[0], infinispanEnvs)
	return true, nil
}

func ensureKafka(instance *v1alpha1.KogitoDataIndex, statefulset *appsv1.StatefulSet, client *client.Client) (bool, error) {
	if len(statefulset.Spec.Template.Spec.Containers) == 0 {
		return false, nil
	}

	if externalURI, err := getKafkaServerURI(instance.Spec.Kafka, instance.Namespace, client); err != nil {
		return false, err
	} else if len(externalURI) > 0 {
		updated := false
		for _, kafkaEnv := range managedKafkaKeys {
			currentURI := framework.GetEnvVarFromContainer(kafkaEnv, statefulset.Spec.Template.Spec.Containers[0])
			if externalURI != currentURI {
				log.Debugf("Found differences in the Kafka URI (%s). Updating to '%s'.", currentURI, externalURI)
				framework.SetEnvVar(kafkaEnv, externalURI, &statefulset.Spec.Template.Spec.Containers[0])
				updated = true
			}
		}
		return updated, nil
	}

	return false, nil
}

func ensureKafkaTopics(instance *v1alpha1.KogitoDataIndex, kafkaTopics []kafkabetav1.KafkaTopic, client *client.Client) error {
	if len(kafkaTopics) == 0 {
		return nil
	}

	kafkaName, kafkaReplicas, err := getKafkaServerReplicas(instance.Spec.Kafka, instance.Namespace, client)
	if err != nil {
		return err
	} else if len(kafkaName) <= 0 || kafkaReplicas <= 0 {
		return nil
	}

	for _, kafkaTopic := range kafkaTopics {
		if kafkaTopic.Labels[kafkaClusterLabel] != kafkaName || kafkaTopic.Spec.Replicas != kafkaReplicas {
			kafkaTopic.Labels[kafkaClusterLabel] = kafkaName
			kafkaTopic.Spec.Replicas = kafkaReplicas
			if err := kubernetes.ResourceC(client).Update(&kafkaTopic); err != nil {
				log.Error("Error while updating Kafka Topic with new files: ", err)
				return err
			}
		}
	}

	return nil
}
