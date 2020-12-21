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

package flag

import (
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/util"
	"github.com/spf13/cobra"
	"net/url"
)

// ResourceType represents mediums through which user can trigger build
type ResourceType string

const (
	// LocalFileResource build using file on local system
	LocalFileResource ResourceType = "LocalFileResource"
	// LocalDirectoryResource s2i build using directory on local system
	LocalDirectoryResource ResourceType = "LocalDirectoryResource"
	// LocalBinaryDirectoryResource binary build using directory on local system
	LocalBinaryDirectoryResource ResourceType = "LocalBinaryDirectoryResource"
	// GitFileResource build using file in Git Repo
	GitFileResource ResourceType = "GitFileResource"
	// GitRepositoryResource build using Git Repo
	GitRepositoryResource ResourceType = "GitRepositoryResource"
	// BinaryResource build using user generated binaries
	BinaryResource ResourceType = "BinaryResource"
)

// BuildFlags is common properties used to configure Build
type BuildFlags struct {
	GitSourceFlags
	RuntimeTypeFlags
	PodResourceFlags
	ArtifactFlags
	WebHookFlags
	EnvVarFlags
	Name                      string
	Project                   string
	IncrementalBuild          bool
	Native                    bool
	MavenMirrorURL            string
	BuildImage                string
	RuntimeImage              string
	TargetRuntime             string
	EnableMavenDownloadOutput bool
}

// AddBuildFlags adds the BuildFlags to the given command
func AddBuildFlags(command *cobra.Command, flags *BuildFlags) {
	AddGitSourceFlags(command, &flags.GitSourceFlags)
	AddPodResourceFlags(command, &flags.PodResourceFlags, "build")
	AddArtifactFlags(command, &flags.ArtifactFlags)
	AddWebHookFlags(command, &flags.WebHookFlags)
	AddEnvVarFlags(command, &flags.EnvVarFlags, "build-env", "")
	command.Flags().BoolVar(&flags.IncrementalBuild, "incremental-build", true, "Build should be incremental?")
	command.Flags().BoolVar(&flags.Native, "native", false, "Use native builds? Be aware that native builds takes more time and consume much more resources from the cluster. Defaults to false. Currently only works with s2i (requires [SOURCE] argument).")
	command.Flags().StringVar(&flags.MavenMirrorURL, "maven-mirror-url", "", "Internal Maven Mirror to be used during source-to-image builds to considerably increase build speed, e.g: https://my.internal.nexus/content/group/public")
	command.Flags().StringVar(&flags.BuildImage, "image-s2i", "", "Custom image tag for the s2i build to build the application binaries, e.g: quay.io/mynamespace/myimage:latest")
	command.Flags().StringVar(&flags.RuntimeImage, "image-runtime", "", "Custom image tag for the s2i build, e.g: quay.io/mynamespace/myimage:latest")
	command.Flags().StringVar(&flags.TargetRuntime, "target-runtime", "", "Set this field targeting the desired KogitoService when this KogitoBuild instance has a different name than the KogitoService")
	command.Flags().BoolVarP(&flags.EnableMavenDownloadOutput, "maven-output", "m", false, "If set to true will print the logs for downloading/uploading of maven dependencies. Defaults to false")
}

// CheckBuildArgs validates the BuildFlags flags
func CheckBuildArgs(flags *BuildFlags) error {
	if err := util.CheckImageTag(flags.RuntimeImage); err != nil {
		return err
	}
	if err := util.CheckImageTag(flags.BuildImage); err != nil {
		return err
	}
	if err := CheckGitSourceArgs(&flags.GitSourceFlags); err != nil {
		return err
	}
	if err := CheckPodResourceArgs(&flags.PodResourceFlags); err != nil {
		return err
	}
	if err := CheckArtifactArgs(&flags.ArtifactFlags); err != nil {
		return err
	}
	if err := CheckWebHookArgs(&flags.WebHookFlags); err != nil {
		return err
	}
	if err := CheckEnvVarArgs(&flags.EnvVarFlags); err != nil {
		return err
	}
	if len(flags.MavenMirrorURL) > 0 {
		if _, err := url.ParseRequestURI(flags.MavenMirrorURL); err != nil {
			return err
		}
	}
	return nil
}
