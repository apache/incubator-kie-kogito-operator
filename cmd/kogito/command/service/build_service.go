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
	"github.com/kiegroup/kogito-operator/apis/app/v1beta1"
	"github.com/kiegroup/kogito-operator/core/kogitobuild"
	"github.com/kiegroup/kogito-operator/core/manager"
	"github.com/kiegroup/kogito-operator/core/operator"
	"io"

	"github.com/kiegroup/kogito-operator/apis"
	"github.com/kiegroup/kogito-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-operator/cmd/kogito/command/converter"
	"github.com/kiegroup/kogito-operator/cmd/kogito/command/flag"
	"github.com/kiegroup/kogito-operator/cmd/kogito/command/message"
	"github.com/kiegroup/kogito-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/kogito-operator/core/client/kubernetes"
	"github.com/kiegroup/kogito-operator/meta"
	buildv1 "github.com/openshift/api/build/v1"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BuildService is interface to perform Kogito Build
type BuildService interface {
	InstallBuildService(flags *flag.BuildFlags, resource string) (err error)
	DeleteBuildService(name, project string) (err error)
}

type buildService struct {
	operator.Context
	resourceCheckService shared.ResourceCheckService
	buildHandler         manager.KogitoBuildHandler
}

// NewBuildService create and return buildService value
func NewBuildService(context operator.Context, buildHandler manager.KogitoBuildHandler) BuildService {
	return buildService{
		Context:              context,
		resourceCheckService: shared.NewResourceCheckService(),
		buildHandler:         buildHandler,
	}
}

// InstallBuildService install Kogito build service
func (i buildService) InstallBuildService(flags *flag.BuildFlags, resource string) (err error) {
	log := context.GetDefaultLogger()
	log.Debugf("Installing Kogito build : %s", flags.Name)

	if err = i.validatePreRequisite(flags, log); err != nil {
		return err
	}

	resourceType, err := GetResourceType(resource)
	if err != nil {
		return err
	}

	if resourceType == flag.GitRepositoryResource {
		flags.GitSourceFlags.Source = resource
	}

	native, err := converter.FromArgsToNative(flags.Native, resourceType, resource)
	if err != nil {
		return err
	}

	runtime, err := converter.FromArgsToRuntimeType(&flags.RuntimeTypeFlags, resourceType, resource)
	if err != nil {
		return err
	}

	legacy, err := converter.ToQuarkusLegacyJarType(resourceType, resource)
	if err != nil {
		return err
	}

	kogitoBuild := v1beta1.KogitoBuild{
		ObjectMeta: v1.ObjectMeta{
			Name:      flags.Name,
			Namespace: flags.Project,
		},
		Spec: v1beta1.KogitoBuildSpec{
			Type:                      converter.FromResourceTypeToKogitoBuildType(resourceType),
			DisableIncremental:        !flags.IncrementalBuild,
			Env:                       converter.FromStringArrayToEnvs(flags.Env, flags.SecretEnv),
			GitSource:                 converter.FromGitSourceFlagsToGitSource(&flags.GitSourceFlags),
			Runtime:                   runtime,
			WebHooks:                  converter.FromWebHookFlagsToWebHookSecret(&flags.WebHookFlags),
			Native:                    native,
			Resources:                 converter.FromPodResourceFlagsToResourceRequirement(&flags.PodResourceFlags),
			MavenMirrorURL:            flags.MavenMirrorURL,
			BuildImage:                flags.BuildImage,
			RuntimeImage:              flags.RuntimeImage,
			TargetKogitoRuntime:       flags.TargetRuntime,
			Artifact:                  converter.FromArtifactFlagsToArtifact(&flags.ArtifactFlags),
			EnableMavenDownloadOutput: flags.EnableMavenDownloadOutput,
		},
	}

	log.Debugf("Trying to build Kogito Service '%s'", kogitoBuild.Name)

	// Create the Kogito application
	err = shared.
		ServicesInstallationBuilder(i.Client, flags.Project).
		CheckOperatorCRDs().
		InstallBuildService(&kogitoBuild).
		GetError()
	if err != nil {
		return err
	}

	binaryBuildType := converter.FromArgsToBinaryBuildType(resourceType, runtime, native, legacy)
	return i.createBuildIfRequires(flags.Name, flags.Project, resource, resourceType, binaryBuildType)
}

func (i buildService) validatePreRequisite(flags *flag.BuildFlags, log *zap.SugaredLogger) error {
	if !i.Client.IsOpenshift() {
		log.Info("Kogito Build is only supported on Openshift.")
		return fmt.Errorf("kogito build only supported on Openshift. Provide image flag to deploy Kogito service on K8s")
	}

	if flags.Native {
		if api.RuntimeType(flags.RuntimeTypeFlags.Runtime) != api.QuarkusRuntimeType {
			return fmt.Errorf("native builds are only supported with %s runtime", api.QuarkusRuntimeType)
		}
	}
	return nil
}

// DeleteBuildService delete Kogito build service
func (i buildService) DeleteBuildService(name, project string) (err error) {
	log := context.GetDefaultLogger()

	if !i.Client.IsOpenshift() {
		log.Info("Delete Kogito Build is only supported on OpenShift.")
		return nil
	}
	if err := i.resourceCheckService.CheckKogitoBuildExists(i.Client, name, project); err != nil {
		return err
	}
	log.Debugf("About to delete build %s in namespace %s", name, project)
	if err := kubernetes.ResourceC(i.Client).Delete(&v1beta1.KogitoBuild{
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

func (i buildService) createBuildIfRequires(name, namespace, resource string, resourceType flag.ResourceType, binaryBuildType flag.BinaryBuildType) error {
	switch resourceType {
	case flag.GitRepositoryResource:
		i.handleGitRepositoryBuild(name, namespace)
	case flag.GitFileResource:
		if err := i.handleGitFileResourceBuild(name, namespace, resource); err != nil {
			return err
		}
	case flag.LocalDirectoryResource, flag.LocalBinaryDirectoryResource:
		if err := i.handleLocalDirectoryResourceBuild(name, namespace, resource, binaryBuildType); err != nil {
			return err
		}
	case flag.LocalFileResource:
		if err := i.handleLocalFileResourceBuild(name, namespace, resource); err != nil {
			return err
		}
	case flag.BinaryResource:
		i.handleBinaryResourceBuild(name, namespace)
	}
	return nil
}

func (i buildService) handleGitRepositoryBuild(name, namespace string) {
	log := context.GetDefaultLogger()
	log.Infof(message.KogitoBuildViewDeploymentStatus, name, namespace)
	log.Infof(message.KogitoViewBuildStatus, name, namespace)
}

func (i buildService) handleGitFileResourceBuild(name, namespace, resource string) error {
	fileReader, fileName, err := LoadGitFileIntoMemory(resource)
	if err != nil {
		return err
	}
	return i.triggerBuild(name, namespace, fileReader, fileName, false)
}

func (i buildService) handleLocalDirectoryResourceBuild(name, namespace, resource string, binaryBuildType flag.BinaryBuildType) error {
	fileReader, fileName, err := ZipAndLoadLocalDirectoryIntoMemory(resource, binaryBuildType)
	if err != nil {
		return err
	}

	binaryBuild := true
	if binaryBuildType == flag.SourceToImageBuild {
		binaryBuild = false
	}
	return i.triggerBuild(name, namespace, fileReader, fileName, binaryBuild)
}

func (i buildService) handleLocalFileResourceBuild(name, namespace, resource string) error {
	fileReader, fileName, err := LoadLocalFileIntoMemory(resource)
	if err != nil {
		return err
	}
	return i.triggerBuild(name, namespace, fileReader, fileName, false)
}

func (i buildService) handleBinaryResourceBuild(name, namespace string) {
	log := context.GetDefaultLogger()
	log.Infof(message.KogitoBuildUploadBinariesInstruction, name, namespace)
}

func (i buildService) triggerBuild(name string, namespace string, fileReader io.Reader, fileName string, binaryBuild bool) error {
	log := context.GetDefaultLogger()
	options := &buildv1.BinaryBuildRequestOptions{}
	options.Name = name
	if len(fileName) > 0 {
		options.AsFile = fileName
	}

	log.Info(message.BuildTriggeringNewBuild)

	build, err := kogitobuild.NewBuildHandler(i.Context, i.buildHandler).TriggerBuildFromFile(namespace, fileReader, options, binaryBuild, meta.GetRegisteredSchema())
	if err != nil {
		return err
	}

	if binaryBuild {
		log.Infof(message.KogitoBuildSuccessfullyUploadedBinaries, build.Name, name, namespace)
	} else {
		log.Infof(message.KogitoBuildSuccessfullyUploadedFile, build.Name, name, namespace)
	}
	return nil
}
