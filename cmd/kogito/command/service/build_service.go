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
	"fmt"
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
	"go.uber.org/zap"
	"io"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BuildService is interface to perform Kogito Build
type BuildService interface {
	InstallBuildService(cli *client.Client, flags *flag.BuildFlags, resource string) (err error)
	DeleteBuildService(cli *client.Client, name, project string) (err error)
}

type buildService struct {
	resourceCheckService shared.IResourceCheckService
}

// NewBuildService create and return buildService value
func NewBuildService() BuildService {
	return buildService{
		resourceCheckService: shared.InitResourceCheckService(),
	}
}

// InstallBuildService install Kogito build service
func (i buildService) InstallBuildService(cli *client.Client, flags *flag.BuildFlags, resource string) (err error) {
	log := context.GetDefaultLogger()
	log.Debugf("Installing Kogito build : %s", flags.Name)

	if err = validatePreRequisite(cli, flags, log, i); err != nil {
		return err
	}

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
			Envs:                      converter.FromStringArrayToEnvs(flags.Env, flags.SecretEnv),
			GitSource:                 converter.FromGitSourceFlagsToGitSource(&flags.GitSourceFlags),
			Runtime:                   converter.FromRuntimeFlagsToRuntimeType(&flags.RuntimeTypeFlags),
			WebHooks:                  converter.FromWebHookFlagsToWebHookSecret(&flags.WebHookFlags),
			Native:                    flags.Native,
			Resources:                 converter.FromPodResourceFlagsToResourceRequirement(&flags.PodResourceFlags),
			MavenMirrorURL:            flags.MavenMirrorURL,
			BuildImage:                converter.FromImageTagToImage(flags.BuildImage, flags.ImageVersion),
			RuntimeImage:              converter.FromImageTagToImage(flags.RuntimeImage, flags.ImageVersion),
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
		return err
	}

	return nil
}

func validatePreRequisite(cli *client.Client, flags *flag.BuildFlags, log *zap.SugaredLogger, i buildService) error {

	if !cli.IsOpenshift() {
		log.Info("Kogito Build is only supported on Openshift.")
		return fmt.Errorf("kogito build only supported on Openshift. Provide image flag to deploy Kogito service on K8s")
	}

	if err := i.resourceCheckService.CheckKogitoBuildNotExists(cli, flags.Name, flags.Project); err != nil {
		return err
	}

	if flags.Native {
		if v1alpha1.RuntimeType(flags.RuntimeTypeFlags.Runtime) != v1alpha1.QuarkusRuntimeType {
			return fmt.Errorf("native builds are only supported with %s runtime", v1alpha1.QuarkusRuntimeType)
		}
	}
	return nil
}

// DeleteBuildService delete Kogito build service
func (i buildService) DeleteBuildService(cli *client.Client, name, project string) (err error) {
	log := context.GetDefaultLogger()

	if !cli.IsOpenshift() {
		log.Info("Delete Kogito Build is only supported on OpenShift.")
		return nil
	}
	if err := i.resourceCheckService.CheckKogitoBuildExists(cli, name, project); err != nil {
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

	log.Info(message.BuildTriggeringNewBuild)
	build, err := openshift.BuildConfigC(cli).TriggerBuildFromFile(namespace, fileReader, options)
	if err != nil {
		return err
	}

	log.Infof(message.KogitoBuildSuccessfullyUploadedFile, build.Name, name, namespace)
	return nil
}
