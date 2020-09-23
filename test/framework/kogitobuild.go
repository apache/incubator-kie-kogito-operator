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

package framework

import (
	"fmt"

	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitobuild/build"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/test/config"
	"github.com/kiegroup/kogito-cloud-operator/test/framework/mappers"
	bddtypes "github.com/kiegroup/kogito-cloud-operator/test/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// DeployKogitoBuild deploy a KogitoBuild
func DeployKogitoBuild(namespace string, installerType InstallerType, buildHolder *bddtypes.KogitoBuildHolder) error {
	GetLogger(namespace).Infof("%s deploy %s example %s with name %s and native %v", installerType, buildHolder.KogitoBuild.Spec.Runtime, buildHolder.KogitoBuild.Spec.GitSource.ContextDir, buildHolder.KogitoBuild.Name, buildHolder.KogitoBuild.Spec.Native)

	switch installerType {
	case CLIInstallerType:
		return cliDeployKogitoBuild(buildHolder)
	case CRInstallerType:
		return crDeployKogitoBuild(buildHolder)
	default:
		panic(fmt.Errorf("Unknown installer type %s", installerType))
	}
}

func crDeployKogitoBuild(buildHolder *bddtypes.KogitoBuildHolder) error {
	if _, err := kubernetes.ResourceC(kubeClient).CreateIfNotExists(buildHolder.KogitoBuild); err != nil {
		return fmt.Errorf("Error creating example build %s: %v", buildHolder.KogitoBuild.Name, err)
	}
	if _, err := kubernetes.ResourceC(kubeClient).CreateIfNotExists(buildHolder.KogitoService); err != nil {
		return fmt.Errorf("Error creating example service %s: %v", buildHolder.KogitoService.GetName(), err)
	}
	return nil
}

func cliDeployKogitoBuild(buildHolder *bddtypes.KogitoBuildHolder) error {
	cmd := []string{"deploy", buildHolder.KogitoBuild.GetName()}

	// If GIT URI is defined then it needs to be appended as second parameter
	if gitURI := buildHolder.KogitoBuild.Spec.GitSource.URI; len(gitURI) > 0 {
		cmd = append(cmd, gitURI)
	} else if len(buildHolder.BuiltBinaryFolder) > 0 {
		cmd = append(cmd, buildHolder.BuiltBinaryFolder)
	}

	cmd = append(cmd, mappers.GetBuildCLIFlags(buildHolder.KogitoBuild)...)
	cmd = append(cmd, mappers.GetServiceCLIFlags(buildHolder.KogitoServiceHolder)...)

	_, err := ExecuteCliCommandInNamespace(buildHolder.KogitoBuild.GetNamespace(), cmd...)
	return err
}

// GetKogitoBuildStub Get basic KogitoBuild stub with all needed fields initialized
func GetKogitoBuildStub(namespace, runtimeType, name string) *v1alpha1.KogitoBuild {
	kogitoBuild := &v1alpha1.KogitoBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Status: v1alpha1.KogitoBuildStatus{
			Conditions: []v1alpha1.KogitoBuildConditions{},
		},
		Spec: v1alpha1.KogitoBuildSpec{
			Runtime:        v1alpha1.RuntimeType(runtimeType),
			MavenMirrorURL: config.GetMavenMirrorURL(),
		},
	}

	if len(config.GetCustomMavenRepoURL()) > 0 {
		kogitoBuild.Spec.Envs = framework.EnvOverride(kogitoBuild.Spec.Envs, corev1.EnvVar{Name: "MAVEN_REPO_URL", Value: config.GetCustomMavenRepoURL()})
	}

	return kogitoBuild
}

// GetKogitoBuild returns the KogitoBuild type
func GetKogitoBuild(namespace, buildName string) (*v1alpha1.KogitoBuild, error) {
	build := &v1alpha1.KogitoBuild{}
	if exists, err := kubernetes.ResourceC(kubeClient).FetchWithKey(types.NamespacedName{Name: buildName, Namespace: namespace}, build); err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("Error while trying to look for KogitoBuild %s: %v ", buildName, err)
	} else if errors.IsNotFound(err) || !exists {
		return nil, nil
	}
	return build, nil
}

// SetupKogitoBuildImageStreams sets the correct images for the KogitoBuild
func SetupKogitoBuildImageStreams(kogitoBuild *v1alpha1.KogitoBuild) {
	kogitoBuild.Spec.BuildImage = getKogitoBuildS2IImage(kogitoBuild)
	kogitoBuild.Spec.RuntimeImage = getKogitoBuildRuntimeImage(kogitoBuild)
}

func getKogitoBuildS2IImage(kogitoBuild *v1alpha1.KogitoBuild) string {
	if len(config.GetBuildS2IImageStreamTag()) > 0 {
		return config.GetBuildS2IImageStreamTag()
	}

	return getKogitoBuildImage(build.KogitoImages[kogitoBuild.Spec.Runtime][true])
}

func getKogitoBuildRuntimeImage(kogitoBuild *v1alpha1.KogitoBuild) string {
	if len(config.GetBuildRuntimeImageStreamTag()) > 0 {
		return config.GetBuildRuntimeImageStreamTag()
	}
	imageName := build.KogitoImages[kogitoBuild.Spec.Runtime][false]
	if kogitoBuild.Spec.Native {
		imageName = build.KogitoQuarkusUbi8Image
	}
	return getKogitoBuildImage(imageName)
}

// getKogitoBuildImage returns a build image with defaults set
func getKogitoBuildImage(imageName string) string {
	image := v1alpha1.Image{
		Domain:    config.GetBuildImageRegistry(),
		Namespace: config.GetBuildImageNamespace(),
		Tag:       config.GetBuildImageVersion(),
	}

	// Update image name with suffix if provided
	if len(config.GetBuildImageNameSuffix()) > 0 {
		image.Name = fmt.Sprintf("%s-%s", imageName, config.GetBuildImageNameSuffix())
	}
	return framework.ConvertImageToImageTag(image)
}
