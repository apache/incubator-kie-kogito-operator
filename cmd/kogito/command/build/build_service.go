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

package build

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/converter"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/message"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	buildutil "github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/util"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
	buildv1 "github.com/openshift/api/build/v1"
	"github.com/spf13/cobra"
	"io"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/url"
)

type buildFlags struct {
	flag.OperatorFlags
	flag.GitSourceFlags
	flag.RuntimeFlags
	flag.PodResourceFlags
	flag.ArtifactFlags
	flag.WebHookFlags
	name                      string
	project                   string
	incrementalBuild          bool
	env                       []string
	native                    bool
	mavenMirrorURL            string
	buildImage                string
	runtimeImage              string
	targetRuntime             string
	enableMavenDownloadOutput bool
}
type buildCommand struct {
	context.CommandContext
	command *cobra.Command
	flags   buildFlags
	Parent  *cobra.Command
}

// initDeployCommand is the constructor for the deploy command
func initBuildServiceCommand(ctx *context.CommandContext, parent *cobra.Command) context.KogitoCommand {
	cmd := &buildCommand{CommandContext: *ctx, Parent: parent}
	cmd.RegisterHook()
	cmd.InitHook()
	return cmd
}

func (i *buildCommand) Command() *cobra.Command {
	return i.command
}

func (i *buildCommand) RegisterHook() {
	i.command = &cobra.Command{
		Use:     "build-service NAME [SOURCE]",
		Short:   "Build kogito service",
		Aliases: []string{"build"},
		Long: `Build the provided SOURCE code of Kogito service and create a image for it. 
	If the [SOURCE] is provided, the build will take place on the cluster.
	If not, you can also provide a dmn/drl/bpmn/bpmn2 file or a directory containing one or more of those files, using the --from-file
	`,
		RunE:    i.Exec,
		PreRun:  i.CommonPreRun,
		PostRun: i.CommonPostRun,
		// Args validation
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) <= 1 {
				return fmt.Errorf("requires 1 arg, received %v", len(args))
			}
			if err := buildutil.CheckImageTag(i.flags.runtimeImage); err != nil {
				return err
			}
			if err := buildutil.CheckImageTag(i.flags.buildImage); err != nil {
				return err
			}
			if err := flag.CheckOperatorArgs(&i.flags.OperatorFlags); err != nil {
				return err
			}
			if err := flag.CheckGitSourceArgs(&i.flags.GitSourceFlags); err != nil {
				return err
			}
			if err := flag.CheckRuntimeArgs(&i.flags.RuntimeFlags); err != nil {
				return err
			}
			if err := flag.CheckResourceArgs(&i.flags.PodResourceFlags); err != nil {
				return err
			}
			if err := flag.CheckArtifactArgs(&i.flags.ArtifactFlags); err != nil {
				return err
			}
			if err := flag.CheckWebHookArgs(&i.flags.WebHookFlags); err != nil {
				return err
			}
			if err := util.ParseStringsForKeyPair(i.flags.env); err != nil {
				return fmt.Errorf("build environment variables are in the wrong format. Valid are key pairs like 'env=value', received %s", i.flags.env)
			}
			if i.flags.native {
				if converter.FromRuntimeFlagsToRuntimeType(&i.flags.RuntimeFlags) != v1alpha1.QuarkusRuntimeType {
					return fmt.Errorf("native builds are only supported with %s runtime", v1alpha1.QuarkusRuntimeType)
				}
			}
			if len(i.flags.mavenMirrorURL) > 0 {
				if _, err := url.ParseRequestURI(i.flags.mavenMirrorURL); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func (i *buildCommand) InitHook() {
	i.Parent.AddCommand(i.command)
	i.flags = buildFlags{
		OperatorFlags:    flag.OperatorFlags{},
		GitSourceFlags:   flag.GitSourceFlags{},
		RuntimeFlags:     flag.RuntimeFlags{},
		PodResourceFlags: flag.PodResourceFlags{},
		ArtifactFlags:    flag.ArtifactFlags{},
		WebHookFlags:     flag.WebHookFlags{},
	}
	flag.AddOperatorFlags(i.command, &i.flags.OperatorFlags)
	flag.AddGitSourceFlags(i.command, &i.flags.GitSourceFlags)
	flag.AddRuntimeFlags(i.command, &i.flags.RuntimeFlags)
	flag.AddResourceFlags(i.command, &i.flags.PodResourceFlags)
	flag.AddArtifactFlags(i.command, &i.flags.ArtifactFlags)
	flag.AddWebHookFlags(i.command, &i.flags.WebHookFlags)
	i.command.Flags().StringVarP(&i.flags.project, "project", "p", "", "The project name where the service will be deployed")
	i.command.Flags().BoolVar(&i.flags.incrementalBuild, "incremental-build", true, "Build should be incremental?")
	i.command.Flags().StringArrayVar(&i.flags.env, "env", nil, "Key/pair value environment variables that will be set during the build. For example 'MY_CUSTOM_ENV=my_custom_value'. Can be set more than once.")
	i.command.Flags().BoolVar(&i.flags.native, "native", false, "Use native builds? Be aware that native builds takes more time and consume much more resources from the cluster. Defaults to false")
	i.command.Flags().StringVar(&i.flags.mavenMirrorURL, "maven-mirror-url", "", "Internal Maven Mirror to be used during source-to-image builds to considerably increase build speed, e.g: https://my.internal.nexus/content/group/public")
	i.command.Flags().StringVar(&i.flags.buildImage, "build-image", "", "Custom image tag for the s2i build to build the application binaries, e.g: quay.io/mynamespace/myimage:latest")
	i.command.Flags().StringVar(&i.flags.runtimeImage, "runtime-image", "", "Custom image tag for the s2i build, e.g: quay.io/mynamespace/myimage:latest")
	i.command.Flags().StringVar(&i.flags.targetRuntime, "target-runtime", "", "Set this field targeting the desired KogitoRuntime when this KogitoBuild instance has a different name than the KogitoRuntime")
	i.command.Flags().BoolVarP(&i.flags.enableMavenDownloadOutput, "maven-output", "m", false, "If set to true will print the logs for downloading/uploading of maven dependencies. Defaults to false")
}

func (i *buildCommand) Exec(cmd *cobra.Command, args []string) (err error) {
	log := context.GetDefaultLogger()
	i.flags.name = args[0]
	if i.flags.project, err = shared.EnsureProject(i.Client, i.flags.project); err != nil {
		return err
	}

	resourceType, err := buildutil.GetResourceType(args[1])
	if err != nil {
		return nil
	}

	if resourceType == buildutil.GitRepositoryResource {
		i.flags.GitSourceFlags.Source = args[1]
	}

	kogitoBuild := v1alpha1.KogitoBuild{
		ObjectMeta: v1.ObjectMeta{
			Name:      i.flags.name,
			Namespace: i.flags.project,
		},
		Spec: v1alpha1.KogitoBuildSpec{
			Type:                      converter.FromResourceTypeToKogitoBuildType(resourceType),
			DisableIncremental:        !i.flags.incrementalBuild,
			Envs:                      converter.FromStringArrayToEnvs(i.flags.env),
			GitSource:                 converter.FromGitSourceFlagsToGitSource(&i.flags.GitSourceFlags),
			WebHooks:                  converter.FromWebHookFlagsToWebHookSecret(&i.flags.WebHookFlags),
			Runtime:                   converter.FromRuntimeFlagsToRuntimeType(&i.flags.RuntimeFlags),
			Native:                    i.flags.native,
			Resources:                 converter.FromPodResourceFlagsToResourceRequirement(&i.flags.PodResourceFlags),
			MavenMirrorURL:            i.flags.mavenMirrorURL,
			BuildImage:                converter.FromImageTagToImage(i.flags.buildImage),
			RuntimeImage:              converter.FromImageTagToImage(i.flags.runtimeImage),
			TargetKogitoRuntime:       i.flags.targetRuntime,
			Artifact:                  converter.FromArtifactFlagsToArtifact(&i.flags.ArtifactFlags),
			EnableMavenDownloadOutput: i.flags.enableMavenDownloadOutput,
		},
		Status: v1alpha1.KogitoBuildStatus{
			Conditions: []v1alpha1.KogitoBuildConditions{},
		},
	}

	log.Debugf("Trying to build Kogito Service '%s'", kogitoBuild.Name)

	// Create the Kogito application
	err = shared.
		ServicesInstallationBuilder(i.Client, i.flags.project).
		SilentlyInstallOperatorIfNotExists(shared.KogitoChannelType(i.flags.Channel)).
		BuildService(&kogitoBuild).
		GetError()
	if err != nil {
		return err
	}

	if err := createBuildIfRequires(i.flags.name, i.flags.project, args[1], resourceType); err != nil {
		return nil
	}

	return nil
}

func createBuildIfRequires(name, namespace, resource string, resourceType buildutil.ResourceType) error {
	switch resourceType {
	case buildutil.GitRepositoryResource:
		handleGitRepositoryBuild(name, namespace)
	case buildutil.GitFileResource:
		if err := handleGitFileResourceBuild(name, namespace, resource); err != nil {
			return err
		}
	case buildutil.LocalDirectoryResource:
		if err := handleLocalDirectoryResourceBuild(name, namespace, resource); err != nil {
			return err
		}
	case buildutil.LocalFileResource:
		if err := handleLocalFileResourceBuild(name, namespace, resource); err != nil {
			return err
		}
	case buildutil.BinaryResource:
		handleBinaryResourceBuild(name, namespace)
	}
	return nil
}

func handleGitRepositoryBuild(name, namespace string) {
	log := context.GetDefaultLogger()
	log.Infof(message.KogitoAppViewDeploymentStatus, name, namespace)
	log.Infof(message.KogitoAppViewBuildStatus, name, namespace)
}

func handleGitFileResourceBuild(name, namespace, resource string) error {
	fileReader, fileName, err := buildutil.LoadGitFileIntoMemory(resource)
	if err != nil {
		return err
	}
	if err = triggerBuild(name, namespace, fileReader, fileName); err != nil {
		return err
	}
	return nil
}

func handleLocalDirectoryResourceBuild(name, namespace, resource string) error {
	fileReader, fileName, err := buildutil.ZipAndLoadLocalDirectoryIntoMemory(resource)
	if err != nil {
		return err
	}
	if err = triggerBuild(name, namespace, fileReader, fileName); err != nil {
		return err
	}
	return nil
}

func handleLocalFileResourceBuild(name, namespace, resource string) error {
	fileReader, fileName, err := buildutil.LoadLocalFileIntoMemory(resource)
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
	log.Infof(message.KogitoAppUploadBinariesInstruction, name, namespace)
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

	log.Infof(message.KogitoAppSuccessfullyUploadedFile, name, namespace)
	return nil
}
