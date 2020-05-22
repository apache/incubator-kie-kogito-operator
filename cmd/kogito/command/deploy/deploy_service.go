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

package deploy

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	kogitoerror "github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/error"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/message"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	buildutil "github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/util"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
	buildv1 "github.com/openshift/api/build/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/cobra"
)

const defaultDeployRuntime = string(v1alpha1.QuarkusRuntimeType)

var (
	deployRuntimeValidEntries = []string{string(v1alpha1.QuarkusRuntimeType), string(v1alpha1.SpringbootRuntimeType)}
)

type deployFlags struct {
	CommonFlags
	name              string
	runtime           string
	serviceLabels     []string
	incrementalBuild  bool
	buildEnv          []string
	reference         string
	contextDir        string
	source            string
	imageS2I          string
	imageRuntime      string
	native            bool
	buildLimits       []string
	buildRequests     []string
	imageVersion      string
	mavenMirrorURL    string
	enableIstio       bool
	enablePersistence bool
	enableEvents      bool
}

type deployCommand struct {
	context.CommandContext
	command *cobra.Command
	flags   deployFlags
	Parent  *cobra.Command
}

// initDeployCommand is the constructor for the deploy command
func initDeployCommand(ctx *context.CommandContext, parent *cobra.Command) context.KogitoCommand {
	cmd := &deployCommand{CommandContext: *ctx, Parent: parent}
	cmd.RegisterHook()
	cmd.InitHook()
	return cmd
}

func (i *deployCommand) Command() *cobra.Command {
	return i.command
}

func (i *deployCommand) RegisterHook() {
	i.command = &cobra.Command{
		Use:     "deploy-service NAME [SOURCE]",
		Short:   "Deploys a new Kogito Runtime Service into the given Project",
		Aliases: []string{"deploy"},
		Long: `deploy-service will create a new Kogito Runtime Service in the Project context. 
	If the [SOURCE] is provided, the build will take place on the cluster.
	If not, you can also provide a dmn/drl/bpmn/bpmn2 file or a directory containing one or more of those files, using the --from-file
	Or you can also later upload directly the application binaries via "oc start-build [NAME-binary] --from-dir=target
			
	Project context is the namespace (Kubernetes) or project (OpenShift) where the Service will be deployed.
	To know what's your context, use "kogito project". To set a new Project in the context use "kogito use-project NAME".

	Please note that this command requires the Kogito Operator installed in the cluster.
	For more information about the Kogito Operator installation please refer to https://github.com/kiegroup/kogito-cloud-operator#kogito-operator-installation.
		`,
		RunE:    i.Exec,
		PreRun:  i.CommonPreRun,
		PostRun: i.CommonPostRun,
		// Args validation
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 2 {
				return fmt.Errorf("requires 1 arg, received %v", len(args))
			}
			if len(args) == 0 {
				return fmt.Errorf("the service requires a name ")
			}
			if err := util.ParseStringsForKeyPair(i.flags.buildEnv); err != nil {
				return fmt.Errorf("build environment variables are in the wrong format. Valid are key pairs like 'env=value', received %s", i.flags.buildEnv)
			}
			if err := util.ParseStringsForKeyPair(i.flags.serviceLabels); err != nil {
				return fmt.Errorf("service labels are in the wrong format. Valid are key pairs like 'service=myservice', received %s", i.flags.serviceLabels)
			}
			if err := util.ParseStringsForKeyPair(i.flags.buildLimits); err != nil {
				return fmt.Errorf("build-limits are in the wrong format. Valid are key pairs like 'cpu=1', received %s", i.flags.buildLimits)
			}
			if err := util.ParseStringsForKeyPair(i.flags.buildRequests); err != nil {
				return fmt.Errorf("build-requests are in the wrong format. Valid are key pairs like 'cpu=1', received %s", i.flags.buildRequests)
			}
			if !util.Contains(i.flags.runtime, deployRuntimeValidEntries) {
				return fmt.Errorf("runtime not valid. Valid runtimes are %s. Received %s", deployRuntimeValidEntries, i.flags.runtime)
			}
			if err := CheckImageTag(i.flags.imageRuntime); err != nil {
				return err
			}
			if err := CheckImageTag(i.flags.imageS2I); err != nil {
				return err
			}
			if err := CheckDeployArgs(&i.flags.CommonFlags); err != nil {
				return err
			}
			return nil
		},
	}
}

func (i *deployCommand) InitHook() {
	i.flags = deployFlags{CommonFlags: CommonFlags{}}
	i.Parent.AddCommand(i.command)
	AddDeployFlags(i.command, &i.flags.CommonFlags)
	i.command.Flags().StringVarP(&i.flags.runtime, "runtime", "r", defaultDeployRuntime, "The runtime which should be used to build the Service. Valid values are 'quarkus' or 'springboot'. Default to '"+defaultDeployRuntime+"'.")
	i.command.Flags().StringVarP(&i.flags.reference, "branch", "b", "", "Git branch to use in the git repository")
	i.command.Flags().StringVarP(&i.flags.contextDir, "context-dir", "c", "", "Context/subdirectory where the code is located, relatively to repository root")
	i.command.Flags().StringSliceVar(&i.flags.serviceLabels, "svc-labels", nil, "Labels that should be applied to the internal endpoint of the Kogito Service. Used by the service discovery engine. Example: 'label=value'. Can be set more than once.")
	i.command.Flags().BoolVar(&i.flags.incrementalBuild, "incremental-build", true, "Build should be incremental?")
	i.command.Flags().BoolVar(&i.flags.native, "native", false, "Use native builds? Be aware that native builds takes more time and consume much more resources from the cluster. Defaults to false")
	i.command.Flags().StringArrayVar(&i.flags.buildEnv, "build-env", nil, "Key/pair value environment variables that will be set during the build. For example 'MY_CUSTOM_ENV=my_custom_value'. Can be set more than once.")
	i.command.Flags().StringSliceVar(&i.flags.buildLimits, "build-limits", nil, "Resource limits for the s2i build pod. Valid values are 'cpu' and 'memory'. For example 'cpu=1'. Can be set more than once.")
	i.command.Flags().StringSliceVar(&i.flags.buildRequests, "build-requests", nil, "Resource requests for the s2i build pod. Valid values are 'cpu' and 'memory'. For example 'cpu=1'. Can be set more than once.")
	i.command.Flags().StringVar(&i.flags.imageS2I, "image-s2i", "", "Custom image tag for the s2i build to build the application binaries, e.g: quay.io/mynamespace/myimage:latest")
	i.command.Flags().StringVar(&i.flags.imageRuntime, "image-runtime", "", "Custom image tag for the s2i build, e.g: quay.io/mynamespace/myimage:latest")
	i.command.Flags().BoolVar(&i.flags.enablePersistence, "enable-persistence", false, "If set to true will install Infinispan in the same namespace and inject the environment variables to configure the service connection to the Infinispan server.")
	i.command.Flags().BoolVar(&i.flags.enableEvents, "enable-events", false, "If set to true will install a Kafka cluster via the Strimzi Operator. The environment variable 'KAFKA_BOOTSTRAP_SERVERS' will be available for the service during runtime.")
	i.command.Flags().StringVar(&i.flags.imageVersion, "image-version", "", "Image version for standard Kogito build images. Ignored if a custom image is set for image-s2i or image-runtime.")
	i.command.Flags().StringVar(&i.flags.mavenMirrorURL, "maven-mirror-url", "", "Internal Maven Mirror to be used during source-to-image builds to considerably increase build speed, e.g: https://my.internal.nexus/content/group/public")
	i.command.Flags().BoolVar(&i.flags.enableIstio, "enable-istio", false, "Enable Istio integration by annotating the Kogito service pods with the right value for Istio controller to inject sidecars on it. Defaults to false")
}

func (i *deployCommand) Exec(cmd *cobra.Command, args []string) (err error) {
	log := context.GetDefaultLogger()
	i.flags.name = args[0]
	hasSource := false
	options := &buildv1.BinaryBuildRequestOptions{}
	var userProvidedFile io.Reader

	if len(args) > 1 {
		userProvidedFile, options, err = matchKogitoResources(args[1])
		if err != nil {
			switch t := err.(type) {
			case *os.PathError:
				return t

			case *kogitoerror.KogitoAssetFileBuildError:
				return t

			case *url.Error:
				return t

			default:
				log.Warnf(message.KogitoAppAssetNotSupported, args[1])
			}

		}
		// if the parameter does not match the expected input to perform a build from file, try a build from source instead
		if userProvidedFile == nil {
			log.Info("Trying to perform a build from source...")
			i.flags.source = args[1]
			hasSource = true
		}
	}

	if len(i.flags.mavenMirrorURL) > 0 {
		if _, err := url.ParseRequestURI(i.flags.mavenMirrorURL); err != nil {
			return err
		}
	}

	if i.flags.Project, err = shared.EnsureProject(i.Client, i.flags.Project); err != nil {
		return err
	}

	if installed, err := shared.SilentlyInstallOperatorIfNotExists(i.flags.Project, "", i.Client); err != nil {
		return err
	} else if !installed {
		return nil
	}

	if err := shared.CheckKogitoAppNotExists(i.Client, i.flags.name, i.flags.Project); err != nil {
		return err
	}

	log.Debugf("About to deploy a new kogito service: %s, runtime %s source %s on namespace %s",
		i.flags.name,
		i.flags.runtime,
		i.flags.source,
		i.flags.Project,
	)

	// Build the application
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: v1.ObjectMeta{
			Name:      i.flags.name,
			Namespace: i.flags.Project,
		},
		Spec: v1alpha1.KogitoAppSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{
				Replicas: &i.flags.Replicas,
				Envs:     shared.FromStringArrayToEnvs(i.flags.Env),
				Resources: corev1.ResourceRequirements{
					Limits:   shared.FromStringArrayToResources(i.flags.Limits),
					Requests: shared.FromStringArrayToResources(i.flags.Requests),
				},
			},

			Runtime: v1alpha1.RuntimeType(i.flags.runtime),
			Build: &v1alpha1.KogitoAppBuildObject{
				Envs: shared.FromStringArrayToEnvs(i.flags.buildEnv),
				Resources: corev1.ResourceRequirements{
					Limits:   shared.FromStringArrayToResources(i.flags.buildLimits),
					Requests: shared.FromStringArrayToResources(i.flags.buildRequests),
				},
				Incremental:     i.flags.incrementalBuild,
				ImageS2ITag:     i.flags.imageS2I,
				ImageRuntimeTag: i.flags.imageRuntime,
				ImageVersion:    i.flags.imageVersion,
				Native:          i.flags.native,
				MavenMirrorURL:  i.flags.mavenMirrorURL,
			},
			Service: v1alpha1.KogitoAppServiceObject{
				Labels: util.FromStringsKeyPairToMap(i.flags.serviceLabels),
			},
			EnablePersistence: i.flags.enablePersistence,
			EnableEvents:      i.flags.enableEvents,
			EnableIstio:       i.flags.enableIstio,
		},
		Status: v1alpha1.KogitoAppStatus{
			ConditionsMeta: v1alpha1.ConditionsMeta{Conditions: make([]v1alpha1.Condition, 0)},
		},
	}

	if hasSource {
		kogitoApp.Spec.Build.GitSource = v1alpha1.GitSource{
			URI:        i.flags.source,
			ContextDir: i.flags.contextDir,
			Reference:  i.flags.reference,
		}
	}

	log.Debugf("Trying to deploy Kogito Service '%s'", kogitoApp.Name)
	// Create the Kogito application
	if err := kubernetes.ResourceC(i.Client).Create(kogitoApp); err != nil {
		return fmt.Errorf("Error while creating a new KogitoApp in the context: %v ", err)
	}

	log.Infof(message.KogitoAppSuccessfullyCreated, kogitoApp.Name, kogitoApp.Namespace)
	if hasSource {
		log.Infof(message.KogitoAppViewDeploymentStatus, i.flags.name, i.flags.Project)
		log.Infof(message.KogitoAppViewBuildStatus, i.flags.name, i.flags.Project)

	} else if userProvidedFile != nil {

		options.Name = kogitoApp.Name
		cli, err := client.NewClientBuilder().WithBuildClient().Build()
		if err != nil {
			return err
		}

		_, err = openshift.BuildConfigC(cli).TriggerBuildFromFile(i.flags.Project, userProvidedFile, options)
		if err != nil {
			return err
		}

		log.Infof(message.KogitoAppSuccessfullyUploadedFile, i.flags.name, i.flags.Project)
	} else {
		log.Infof(message.KogitoAppUploadBinariesInstruction, i.flags.name, i.flags.Project)
	}

	endpoint, err := infrastructure.GetManagementConsoleEndpoint(i.Client, i.flags.Project)
	if err != nil {
		return err
	}
	if endpoint.IsEmpty() {
		log.Info(message.KogitoAppNoMgmtConsole)
	} else {
		log.Infof(message.KogitoAppMgmtConsoleEndpoint, endpoint.HTTPRouteURI)
	}

	return nil
}

// matchKogitoResources will verify if the given parameter matches one of the allowed extensions
// Accepted values are, supported file extensions as it is, directory or url
// See supportedExtensions env
// returns the file content as io.Reader if the provided file or url is valid.
func matchKogitoResources(resource string) (io.Reader, *buildv1.BinaryBuildRequestOptions, error) {
	log := context.GetDefaultLogger()
	options := &buildv1.BinaryBuildRequestOptions{}
	var fileName string
	// from url, do only basic check, for example, if the url has the suffix that's match the allowed ones
	switch {
	case strings.HasPrefix(resource, "http"):
		parsedURL, err := url.ParseRequestURI(resource)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			return nil, nil, &url.Error{
				URL: resource,
				Err: err,
			}
		}

		ff := strings.Split(parsedURL.Path, "/")
		fileName = strings.Join(strings.Fields(ff[len(ff)-1]), "")

		if buildutil.IsSuffixSupported(fileName) {
			response, err := http.Get(resource)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to download %s, error message: %s", resource, err.Error())
			}
			options.AsFile = fileName
			log.Infof(message.KogitoAppFoundAsset, fileName)
			return response.Body, options, nil
		}
		return nil, nil, nil

	// handle single file, files on dir.
	default:
		fileInfo, err := os.Stat(resource)
		if err != nil {
			return nil, nil, err
		}

		if fileInfo.Mode().IsRegular() {
			if buildutil.IsSuffixSupported(resource) {
				log.Infof(message.KogitoAppFoundFile, resource)
				ff := strings.Split(resource, "/")
				fileName = strings.Join(strings.Fields(ff[len(ff)-1]), "")
				fileReader, err := os.Open(resource)
				if err != nil {
					return nil, nil, err
				}
				options.AsFile = fileName
				return fileReader, options, nil
			}

		} else if fileInfo.Mode().IsDir() {
			log.Info(message.KogitoAppProvidedFileIsDir)

			ioTgzR, err := buildutil.ProduceTGZfile(resource)
			if err != nil {
				return nil, nil, err
			}
			return ioTgzR, options, nil
		}
		return nil, nil, &kogitoerror.KogitoAssetFileBuildError{Asset: resource}
	}
}
