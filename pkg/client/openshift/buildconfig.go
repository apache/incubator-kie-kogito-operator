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

package openshift

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	buildv1 "github.com/openshift/api/build/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// BuildState describes the state of the build
type BuildState struct {
	ImageExists  bool
	BuildRunning bool
}

// BuildConfigInterface exposes OpenShift BuildConfig operations
type BuildConfigInterface interface {
	EnsureImageBuild(bc *buildv1.BuildConfig, labelSelector string) (BuildState, error)
	TriggerBuild(bc *buildv1.BuildConfig, triggedBy string) (bool, error)
	BuildIsRunning(bc *buildv1.BuildConfig, labelSelector string) (bool, error)
}

func newBuildConfig(c *client.Client) BuildConfigInterface {
	client.MustEnsureClient(c)
	return &buildConfig{
		client: c,
	}
}

type buildConfig struct {
	client *client.Client
}

// EnsureImageBuild checks for the corresponding image for the build. If there's no image, verifies if the build still running.
// Returns a BuildState structure describing it's results. Label selector is used to query for the right bc
func (b *buildConfig) EnsureImageBuild(bc *buildv1.BuildConfig, labelSelector string) (BuildState, error) {
	state := BuildState{
		ImageExists:  false,
		BuildRunning: false,
	}
	bcNamed := types.NamespacedName{
		Name:      bc.Name,
		Namespace: bc.Namespace,
	}
	if img, err := ImageStreamC(b.client).FetchDockerImage(bcNamed); err != nil {
		return state, err
	} else if img == nil {
		log.Debugf("Image not found for build %s", bc.Name)
		state.ImageExists = false
		if running, err := b.BuildIsRunning(bc, labelSelector); running {
			log.Debugf("Build %s is still running", bc.Name)
			state.BuildRunning = true
			return state, nil
		} else if err != nil {
			return state, err
		}
		// TODO: ensure that we don't have errors in the builds and inform this to the user
		log.Debugf("There's no image and no build running or pending for %s.", bc.Name)
		return state, nil
	}
	state.ImageExists = true
	return state, nil
}

// TriggerBuild triggers a new build
func (b *buildConfig) TriggerBuild(bc *buildv1.BuildConfig, triggedBy string) (bool, error) {
	if exists, err := b.checkBuildConfigExists(bc); !exists {
		log.Warnf("Impossible to trigger a new build for %s. Not exists.", bc.Name)
		return false, err
	}
	// catch panic when FakeClient Build is unable to handle dc properly
	defer func() {
		if err := recover(); err != nil {
			log.Info("Skip build triggering duo to a bug on FakeBuild: github.com/openshift/client-go/build/clientset/versioned/typed/build/v1/fake/fake_buildconfig.go:134")
		}
	}()
	buildRequest := newBuildRequest(triggedBy, bc)
	build, err := b.client.BuildCli.BuildConfigs(bc.Namespace).Instantiate(bc.Name, &buildRequest)
	if err != nil {
		return false, err
	}

	log.Info("Build triggered: ", build.Name)
	return true, nil
}

// BuildIsRunning checks if there's a build on New, Pending or Running state for the buildConfiguration
func (b *buildConfig) BuildIsRunning(bc *buildv1.BuildConfig, labelSelector string) (bool, error) {
	if exists, err := b.checkBuildConfigExists(bc); !exists {
		return false, err
	}
	list, err := b.client.BuildCli.Builds(bc.Namespace).List(metav1.ListOptions{
		LabelSelector:        labelSelector,
		IncludeUninitialized: false,
	})
	if err != nil {
		return false, err
	}
	for _, item := range list.Items {
		// it's the build from our buildConfig
		if strings.HasPrefix(item.Name, bc.Name) {
			log.Debugf("Checking status of build '%s'", item.Name)
			if item.Status.Phase == buildv1.BuildPhaseNew ||
				item.Status.Phase == buildv1.BuildPhasePending ||
				item.Status.Phase == buildv1.BuildPhaseRunning {
				log.Debugf("Build %s is still running", item.Name)
				return true, nil
			}
			log.Debugf("Build %s status is %s", item.Name, item.Status.Phase)
		}
	}
	return false, nil
}

func (b *buildConfig) checkBuildConfigExists(bc *buildv1.BuildConfig) (bool, error) {
	if _, err := b.client.BuildCli.BuildConfigs(bc.Namespace).Get(bc.Name, metav1.GetOptions{}); err != nil && errors.IsNotFound(err) {
		log.Warnf("BuildConfig not found in namespace")
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

// newBuildRequest creates a new BuildRequest for the build
func newBuildRequest(triggedby string, bc *buildv1.BuildConfig) buildv1.BuildRequest {
	buildRequest := buildv1.BuildRequest{ObjectMeta: metav1.ObjectMeta{Name: bc.Name}}
	buildRequest.TriggeredBy = []buildv1.BuildTriggerCause{{Message: fmt.Sprintf("Triggered by %s operator", triggedby)}}
	meta.SetGroupVersionKind(&buildRequest.TypeMeta, meta.KindBuildRequest)
	return buildRequest
}
