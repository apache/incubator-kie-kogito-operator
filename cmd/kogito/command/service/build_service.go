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

package service

import (
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/converter"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/message"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	buildv1 "github.com/openshift/api/build/v1"
	"io"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InstallBuildService install Kogito build service
func InstallBuildService(cli *client.Client, flags *flag.BuildFlags, resource string) (err error) {
	log := context.GetDefaultLogger()
	log.Debugf("Installing Kogito build : %s", flags.Name)
	resourceType, err := GetResourceType(resource)
	if err != nil {
		return nil
	}

	if resourceType == flag.GitRepositoryResource {
		flags.GitSourceFlags.Source = resource
	}

	kogitoBuild := v1alpha1.KogitoBuild{
		ObjectMeta: v1.ObjectMeta{
			Name:      flags.Name,
			Namespace: flags.Project,
		},
		Spec: v1alpha1.KogitoBuildSpec{
			Type:                      converter.FromResourceTypeToKogitoBuildType(resourceType),
			DisableIncremental:        !flags.IncrementalBuild,
			Envs:                      converter.FromStringArrayToEnvs(flags.Env),
			GitSource:                 converter.FromGitSourceFlagsToGitSource(&flags.GitSourceFlags),
			WebHooks:                  converter.FromWebHookFlagsToWebHookSecret(&flags.WebHookFlags),
			Runtime:                   converter.FromRuntimeFlagsToRuntimeType(&flags.RuntimeTypeFlags),
			Native:                    flags.Native,
			Resources:                 converter.FromPodResourceFlagsToResourceRequirement(&flags.PodResourceFlags),
			MavenMirrorURL:            flags.MavenMirrorURL,
			BuildImage:                converter.FromImageTagToImage(flags.BuildImage),
			RuntimeImage:              converter.FromImageTagToImage(flags.RuntimeImage),
			TargetKogitoRuntime:       flags.TargetRuntime,
			Artifact:                  converter.FromArtifactFlagsToArtifact(&flags.ArtifactFlags),
			EnableMavenDownloadOutput: flags.EnableMavenDownloadOutput,
		},
		Status: v1alpha1.KogitoBuildStatus{
			Conditions: []v1alpha1.KogitoBuildConditions{},
		},
	}

	log.Debugf("Trying to build Kogito Service '%s'", kogitoBuild.Name)

	// Create the Kogito application
	err = shared.
		ServicesInstallationBuilder(cli, flags.Project).
		SilentlyInstallOperatorIfNotExists(shared.KogitoChannelType(flags.Channel)).
		InstallBuildService(&kogitoBuild).
		GetError()
	if err != nil {
		return err
	}

	if err := createBuildIfRequires(flags.Name, flags.Project, resource, resourceType); err != nil {
		return nil
	}

	return nil
}

func createBuildIfRequires(name, namespace, resource string, resourceType flag.ResourceType) error {
	switch resourceType {
	case flag.GitRepositoryResource:
		handleGitRepositoryBuild(name, namespace)
	case flag.GitFileResource:
		if err := handleGitFileResourceBuild(name, namespace, resource); err != nil {
			return err
		}
	case flag.LocalDirectoryResource:
		if err := handleLocalDirectoryResourceBuild(name, namespace, resource); err != nil {
			return err
		}
	case flag.LocalFileResource:
		if err := handleLocalFileResourceBuild(name, namespace, resource); err != nil {
			return err
		}
	case flag.BinaryResource:
		handleBinaryResourceBuild(name, namespace)
	}
	return nil
}

func handleGitRepositoryBuild(name, namespace string) {
	log := context.GetDefaultLogger()
	log.Infof(message.KogitoBuildViewDeploymentStatus, name, namespace)
	log.Infof(message.KogitoViewBuildStatus, name, namespace)
}

func handleGitFileResourceBuild(name, namespace, resource string) error {
	fileReader, fileName, err := LoadGitFileIntoMemory(resource)
	if err != nil {
		return err
	}
	if err = triggerBuild(name, namespace, fileReader, fileName); err != nil {
		return err
	}
	return nil
}

func handleLocalDirectoryResourceBuild(name, namespace, resource string) error {
	fileReader, fileName, err := ZipAndLoadLocalDirectoryIntoMemory(resource)
	if err != nil {
		return err
	}
	if err = triggerBuild(name, namespace, fileReader, fileName); err != nil {
		return err
	}
	return nil
}

func handleLocalFileResourceBuild(name, namespace, resource string) error {
	fileReader, fileName, err := LoadLocalFileIntoMemory(resource)
	if err != nil {
		return err
	}
	if err = triggerBuild(name, namespace, fileReader, fileName); err != nil {
		return err
	}
	return nil
}

func handleBinaryResourceBuild(name, namespace string) {
	log := context.GetDefaultLogger()
	log.Infof(message.KogitoBuildUploadBinariesInstruction, name, namespace)
}

func triggerBuild(name string, namespace string, fileReader io.Reader, fileName string) error {
	log := context.GetDefaultLogger()
	options := &buildv1.BinaryBuildRequestOptions{}
	options.Name = name
	if len(fileName) > 0 {
		options.AsFile = fileName
	}

	cli, err := client.NewClientBuilder().WithBuildClient().Build()
	if err != nil {
		return err
	}

	_, err = openshift.BuildConfigC(cli).TriggerBuildFromFile(namespace, fileReader, options)
	if err != nil {
		return err
	}

	log.Infof(message.KogitoBuildSuccessfullyUploadedFile, name, namespace)
	return nil
}

// DeleteBuildService delete Kogito build service
func DeleteBuildService(cli *client.Client, name, project string) (err error) {
	log := context.GetDefaultLogger()
	if err := shared.CheckKogitoBuildExists(cli, name, project); err != nil {
		return err
	}
	log.Debugf("About to delete build %s in namespace %s", name, project)
	if err := kubernetes.ResourceC(cli).Delete(&v1alpha1.KogitoBuild{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: project,
		},
	}); err != nil {
		return err
	}
	log.Infof("Successfully deleted Kogito Build %s in the Project %s", name, project)
	return nil
}
