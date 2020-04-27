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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
	buildv1 "github.com/openshift/api/build/v1"
)

const (
	buildConfigLabelSelector = "buildconfig"
)

// Trigger defines how to interact with build triggers in a OpenShift cluster
type Trigger interface {
	// SelectOneBuildConfigWithLabel restricts the trigger only on the first build with the given labels
	SelectOneBuildConfigWithLabel(labelSelector map[string]string, buildConfigs ...*buildv1.BuildConfig) Trigger
	// OnBuildConfig defines the build configuration to trigger the build.
	// Another approach is using `SelectOneBuildConfigWithLabel` to select the build to trigger
	OnBuildConfig(buildConfig *buildv1.BuildConfig) Trigger
	// StartNewBuildIfNotRunning starts a new build for the filtered build config
	StartNewBuildIfNotRunning() (result TriggerResult, err error)
	// HasBuildConfiguration returns true if a build configuration was selected for this instance
	HasBuildConfiguration() bool
}

// TriggerResult structure to hold build trigger information
type TriggerResult struct {
	Started   bool
	BuildName string
}

type trigger struct {
	triggeredBy string
	client      *client.Client
	buildConfig *buildv1.BuildConfig
}

func (t *trigger) HasBuildConfiguration() bool {
	return t.buildConfig != nil
}

func (t *trigger) OnBuildConfig(buildConfig *buildv1.BuildConfig) Trigger {
	t.buildConfig = buildConfig
	return t
}

func (t *trigger) SelectOneBuildConfigWithLabel(labelSelector map[string]string, buildConfigs ...*buildv1.BuildConfig) Trigger {
	for _, bc := range buildConfigs {
		if util.MapContainsMap(bc.Labels, labelSelector) {
			t.buildConfig = bc
			return t
		}
	}
	return t
}

func (t *trigger) StartNewBuildIfNotRunning() (result TriggerResult, err error) {
	result = TriggerResult{
		Started:   false,
		BuildName: "",
	}
	if t.buildConfig == nil {
		return
	}
	result.BuildName = t.buildConfig.Name
	builds, err := openshift.BuildConfigC(t.client).GetBuildsStatus(t.buildConfig, fmt.Sprintf("%s=%s", buildConfigLabelSelector, t.buildConfig.Name))
	if err != nil {
		return
	}
	if !isBuildRunning(builds) {
		if _, err = openshift.BuildConfigC(t.client).TriggerBuild(t.buildConfig, t.triggeredBy); err != nil {
			return
		}
		result.Started = true
	}
	return
}

func isBuildRunning(builds *v1alpha1.Builds) bool {
	return builds != nil && (len(builds.Running) > 0 || len(builds.Pending) > 0 || len(builds.New) > 0)
}

// NewTrigger creates a new build Trigger
func NewTrigger(client *client.Client, triggeredBy string) Trigger {
	return &trigger{
		triggeredBy: triggeredBy,
		client:      client,
		buildConfig: nil,
	}
}
