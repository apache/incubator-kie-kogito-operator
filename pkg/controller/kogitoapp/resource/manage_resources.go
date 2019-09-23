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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/shared"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	routev1 "github.com/openshift/api/route/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"reflect"
	"strconv"
)

// UpdateResourcesResult contains the results of the update of the resources
type UpdateResourcesResult struct {
	Updated     bool
	ErrorReason v1alpha1.ReasonType
	Err         error
}

func createResult(updated bool, err error, errorReason v1alpha1.ReasonType, result *UpdateResourcesResult) {
	if updated {
		result.Updated = true
	}

	if result.Err == nil && err != nil {
		result.Err = err
		result.ErrorReason = errorReason
	}
}

// ManageResources will trigger updates on resources if needed
func ManageResources(instance *v1alpha1.KogitoApp, resources *KogitoAppResources, client *client.Client) *UpdateResourcesResult {
	result := &UpdateResourcesResult{false, v1alpha1.ReasonType(""), nil}

	{
		updated, err := ensureBuildConfigS2I(instance, resources.BuildConfigS2I, client)
		createResult(updated, err, v1alpha1.BuildS2IFailedReason, result)
	}

	{
		updated, err := ensureBuildConfigRuntime(instance, resources.BuildConfigRuntime, client)
		createResult(updated, err, v1alpha1.BuildRuntimeFailedReason, result)
	}

	{
		updated, err := ensureDeploymentConfig(instance, resources.DeploymentConfig, client)
		createResult(updated, err, v1alpha1.DeploymentFailedReason, result)
	}

	{
		updated, err := ensureService(instance, resources.Service, client)
		createResult(updated, err, v1alpha1.ServiceFailedReason, result)
	}

	{
		updated, err := ensureRoute(instance, resources.Route, client)
		createResult(updated, err, v1alpha1.RouteFailedReason, result)
	}

	return result
}

func ensureDeploymentConfig(instance *v1alpha1.KogitoApp, depConfig *appsv1.DeploymentConfig, client *client.Client) (bool, error) {
	if depConfig != nil {
		log := log.With("kind", depConfig.GetObjectKind().GroupVersionKind().Kind, "name", depConfig.Name, "namespace", depConfig.Namespace)

		if _, err := kubernetes.ResourceC(client).Fetch(depConfig); err != nil {
			return false, err
		}

		update := false

		if instance.Spec.Replicas != nil && *instance.Spec.Replicas != depConfig.Spec.Replicas {
			log.Debug("Changes detected in 'Replicas'.", " OLD - ", depConfig.Spec.Replicas, " NEW - ", *instance.Spec.Replicas)
			depConfig.Spec.Replicas = *instance.Spec.Replicas
			update = true
		}

		resources := shared.FromResourcesToResourcesRequirements(instance.Spec.Resources)
		if !reflect.DeepEqual(depConfig.Spec.Template.Spec.Containers[0].Resources, resources) {
			log.Debug("Changes detected in 'Resources'.", " OLD - ", depConfig.Spec.Template.Spec.Containers[0].Resources, " NEW - ", resources)
			depConfig.Spec.Template.Spec.Containers[0].Resources = resources
			update = true
		}

		if !reflect.DeepEqual(util.EnvToMap(instance.Spec.Env), util.EnvVarToMap(depConfig.Spec.Template.Spec.Containers[0].Env)) {
			log.Debug("Changes detected in 'Env'.", " OLD - ", depConfig.Spec.Template.Spec.Containers[0].Env, " NEW - ", instance.Spec.Env)
			depConfig.Spec.Template.Spec.Containers[0].Env = shared.FromEnvToEnvVar(instance.Spec.Env)
			update = true
		}

		if update {
			log.Info("Updating DeploymentConfig")

			if err := kubernetes.ResourceC(client).Update(depConfig); err != nil {
				return update, err
			}

			return update, nil
		}
	}

	return false, nil
}

func ensureService(instance *v1alpha1.KogitoApp, service *corev1.Service, client *client.Client) (bool, error) {
	if service != nil {
		log := log.With("kind", service.GetObjectKind().GroupVersionKind().Kind, "name", service.Name, "namespace", service.Namespace)

		if _, err := kubernetes.ResourceC(client).Fetch(service); err != nil {
			return false, err
		}

		if update := ensureServiceLabels(instance, service.Labels, log); update {
			log.Info("Updating Service")

			if err := kubernetes.ResourceC(client).Update(service); err != nil {
				return update, err
			}

			return update, nil
		}
	}

	return false, nil
}

func ensureRoute(instance *v1alpha1.KogitoApp, route *routev1.Route, client *client.Client) (bool, error) {
	if route != nil {
		log := log.With("kind", route.GetObjectKind().GroupVersionKind().Kind, "name", route.Name, "namespace", route.Namespace)

		if _, err := kubernetes.ResourceC(client).Fetch(route); err != nil {
			return false, err
		}

		if update := ensureServiceLabels(instance, route.Labels, log); update {
			log.Info("Updating Route")

			if err := kubernetes.ResourceC(client).Update(route); err != nil {
				return update, err
			}

			return update, nil
		}
	}

	return false, nil
}

func ensureServiceLabels(instance *v1alpha1.KogitoApp, serviceLabels map[string]string, log *zap.SugaredLogger) bool {
	update := false

	labels := map[string]string{}
	addDefaultLabels(&labels, instance)
	addServiceLabelsToMap(labels, instance)

	if !reflect.DeepEqual(labels, serviceLabels) {
		log.Debug("Changes detected in 'Labels'.", " OLD - ", serviceLabels, " NEW - ", labels)

		for oKey := range serviceLabels {
			delete(serviceLabels, oKey)
		}
		for nKey, nValue := range labels {
			serviceLabels[nKey] = nValue
		}

		update = true
	}

	return update
}

func ensureBuildConfigS2I(instance *v1alpha1.KogitoApp, buildConfigS2I *buildv1.BuildConfig, client *client.Client) (bool, error) {
	if buildConfigS2I != nil {
		log := log.With("kind", buildConfigS2I.GetObjectKind().GroupVersionKind().Kind, "name", buildConfigS2I.Name, "namespace", buildConfigS2I.Namespace)

		if _, err := kubernetes.ResourceC(client).Fetch(buildConfigS2I); err != nil {
			return false, err
		}

		updateS2I := false

		s2iImage := resolveS2IImage(instance)
		s2iImageName, s2iImageNamespace := parseImage(&s2iImage)
		if buildConfigS2I.Spec.Strategy.SourceStrategy.From.Name != s2iImageName {
			log.Debug("Changes detected in 'ImageName'.", " OLD - ", buildConfigS2I.Spec.Strategy.SourceStrategy.From.Name, " NEW - ", s2iImageName)

			buildConfigS2I.Spec.Strategy.SourceStrategy.From.Name = s2iImageName
			updateS2I = true
		}
		if buildConfigS2I.Spec.Strategy.SourceStrategy.From.Namespace != s2iImageNamespace {
			log.Debug("Changes detected in 'ImageNamespace'.", " OLD - ", buildConfigS2I.Spec.Strategy.SourceStrategy.From.Namespace, " NEW - ", s2iImageNamespace)

			buildConfigS2I.Spec.Strategy.SourceStrategy.From.Namespace = s2iImageNamespace
			updateS2I = true
		}

		resources := shared.FromResourcesToResourcesRequirements(instance.Spec.Build.Resources)
		if !reflect.DeepEqual(buildConfigS2I.Spec.Resources, resources) {
			log.Debug("Changes detected in 'Resources'.", " OLD - ", buildConfigS2I.Spec.Resources, " NEW - ", resources)
			buildConfigS2I.Spec.Resources = resources
			updateS2I = true
		}

		envs := util.EnvToMap(instance.Spec.Build.Env)
		if instance.Spec.Runtime == v1alpha1.QuarkusRuntimeType {
			envs[nativeBuildEnvVarKey] = strconv.FormatBool(instance.Spec.Build.Native)
		}
		limitCPU, limitMemory := getBCS2ILimitsAsIntString(buildConfigS2I)
		envs[buildS2IlimitCPUEnvVarKey] = limitCPU
		envs[buildS2IlimitMemoryEnvVarKey] = limitMemory
		if !reflect.DeepEqual(envs, util.EnvVarToMap(buildConfigS2I.Spec.Strategy.SourceStrategy.Env)) {
			log.Debug("Changes detected in 'Env'.", " OLD - ", buildConfigS2I.Spec.Strategy.SourceStrategy.Env, " NEW - ", envs)

			buildConfigS2I.Spec.Strategy.SourceStrategy.Env = util.MapToEnvVar(envs)
			updateS2I = true
		}

		if instance.Spec.Build.Incremental != *buildConfigS2I.Spec.Strategy.SourceStrategy.Incremental {
			log.Debug("Changes detected in 'Incremental'.", " OLD - ", *buildConfigS2I.Spec.Strategy.SourceStrategy.Incremental, " NEW - ", instance.Spec.Build.Incremental)

			buildConfigS2I.Spec.Strategy.SourceStrategy.Incremental = &instance.Spec.Build.Incremental
			updateS2I = true
		}

		if instance.Spec.Build.GitSource.ContextDir != buildConfigS2I.Spec.Source.ContextDir {
			log.Debug("Changes detected in 'ContextDir'.", " OLD - ", buildConfigS2I.Spec.Source.ContextDir, " NEW - ", instance.Spec.Build.GitSource.ContextDir)

			buildConfigS2I.Spec.Source.ContextDir = instance.Spec.Build.GitSource.ContextDir
			updateS2I = true
		}

		if *instance.Spec.Build.GitSource.URI != buildConfigS2I.Spec.Source.Git.URI {
			log.Debug("Changes detected in 'GitSourceURI'.", " OLD - ", buildConfigS2I.Spec.Source.Git.URI, " NEW - ", *instance.Spec.Build.GitSource.URI)

			buildConfigS2I.Spec.Source.Git.URI = *instance.Spec.Build.GitSource.URI
			updateS2I = true
		}
		if instance.Spec.Build.GitSource.Reference != buildConfigS2I.Spec.Source.Git.Ref {
			log.Debug("Changes detected in 'GitSourceRef'.", " OLD - ", buildConfigS2I.Spec.Source.Git.Ref, " NEW - ", instance.Spec.Build.GitSource.Reference)

			buildConfigS2I.Spec.Source.Git.Ref = instance.Spec.Build.GitSource.Reference
			updateS2I = true
		}

		if updateS2I {
			log.Info("Updating BuildConfig for S2I")

			if err := kubernetes.ResourceC(client).Update(buildConfigS2I); err != nil {
				return updateS2I, err
			}

			if _, err := openshift.BuildConfigC(client).TriggerBuild(buildConfigS2I, instance.Name); err != nil {
				return updateS2I, err
			}

			return updateS2I, nil
		}

	}

	return false, nil
}

func ensureBuildConfigRuntime(instance *v1alpha1.KogitoApp, buildConfigRuntime *buildv1.BuildConfig, client *client.Client) (bool, error) {
	if buildConfigRuntime != nil {
		log := log.With("kind", buildConfigRuntime.GetObjectKind().GroupVersionKind().Kind, "name", buildConfigRuntime.Name, "namespace", buildConfigRuntime.Namespace)

		if _, err := kubernetes.ResourceC(client).Fetch(buildConfigRuntime); err != nil {
			return false, err
		}

		updateRuntime := false

		runtimeImage, buildType := resolveRuntimeImage(instance)
		if buildConfigRuntime.Labels[LabelKeyBuildType] != string(buildType) {
			log.Debug("Changes detected in 'BuildType'.", " OLD - ", buildConfigRuntime.Labels[LabelKeyBuildType], " NEW - ", buildType)

			buildConfigRuntime.Labels[LabelKeyBuildType] = string(buildType)
			updateRuntime = true
		}

		runtimeImageName, runtimeImageNamespace := parseImage(&runtimeImage)
		if buildConfigRuntime.Spec.Strategy.SourceStrategy.From.Name != runtimeImageName {
			log.Debug("Changes detected in 'ImageName'.", " OLD - ", buildConfigRuntime.Spec.Strategy.SourceStrategy.From.Name, " NEW - ", runtimeImageName)

			buildConfigRuntime.Spec.Strategy.SourceStrategy.From.Name = runtimeImageName
			updateRuntime = true
		}
		if buildConfigRuntime.Spec.Strategy.SourceStrategy.From.Namespace != runtimeImageNamespace {
			log.Debug("Changes detected in 'ImageNamespace'.", " OLD - ", buildConfigRuntime.Spec.Strategy.SourceStrategy.From.Namespace, " NEW - ", runtimeImageNamespace)

			buildConfigRuntime.Spec.Strategy.SourceStrategy.From.Namespace = runtimeImageNamespace
			updateRuntime = true
		}

		if updateRuntime {
			log.Info("Updating BuildConfig for Runtime")

			if err := kubernetes.ResourceC(client).Update(buildConfigRuntime); err != nil {
				return updateRuntime, err
			}

			if _, err := openshift.BuildConfigC(client).TriggerBuild(buildConfigRuntime, instance.Name); err != nil {
				return updateRuntime, err
			}

			return updateRuntime, nil
		}
	}

	return false, nil
}
