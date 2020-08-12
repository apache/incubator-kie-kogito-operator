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
	"context"
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"io"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	buildv1 "github.com/openshift/api/build/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

const (
	checkBcRetries         = 30
	checkBcRetriesInterval = 2 * time.Second
	// BuildConfigLabelSelector default build selector for buildconfigs
	BuildConfigLabelSelector = "buildconfig"
)

// BuildState describes the state of the build
type BuildState struct {
	ImageExists bool
	Builds      *v1alpha1.Builds
}

// BuildConfigInterface exposes OpenShift BuildConfig operations
type BuildConfigInterface interface {
	EnsureImageBuild(bc *buildv1.BuildConfig, labelSelector, imageName string) (BuildState, error)
	TriggerBuild(bc *buildv1.BuildConfig, triggeredBy string) (bool, error)
	TriggerBuildFromFile(namespace string, r io.Reader, options *buildv1.BinaryBuildRequestOptions) (*buildv1.Build, error)
	GetBuildsStatus(bc *buildv1.BuildConfig, labelSelector string) (*v1alpha1.Builds, error)
	GetBuildsStatusByLabel(namespace, labelSelector string) (*v1alpha1.Builds, error)
}

func newBuildConfig(c *client.Client) BuildConfigInterface {
	client.MustEnsureClient(c)
	return &buildConfig{
		client:                 c,
		checkBcRetries:         checkBcRetries,
		checkBcRetriesInterval: checkBcRetriesInterval,
	}
}

// internal use for unit tests, do not make it public
func newBuildConfigWithBCRetries(c *client.Client, retries int, retriesInterval time.Duration) BuildConfigInterface {
	client.MustEnsureClient(c)
	return &buildConfig{
		client:                 c,
		checkBcRetries:         retries,
		checkBcRetriesInterval: retriesInterval,
	}
}

type buildConfig struct {
	client                 *client.Client
	checkBcRetries         int
	checkBcRetriesInterval time.Duration
}

// EnsureImageBuild checks for the corresponding image for the build and retrieves the status of the builds.
// Returns a BuildState structure describing it's results. Label selector is used to query for the right bc
func (b *buildConfig) EnsureImageBuild(bc *buildv1.BuildConfig, labelSelector, imageName string) (BuildState, error) {
	state := BuildState{
		ImageExists: false,
	}
	imageNamed := types.NamespacedName{
		Name:      imageName,
		Namespace: bc.Namespace,
	}
	if img, err := ImageStreamC(b.client).FetchDockerImage(imageNamed); err != nil {
		return state, err
	} else if img == nil {
		log.Debugf("ImageStream not found for build %s", bc.Name)
	} else {
		state.ImageExists = true
	}

	if builds, err := b.GetBuildsStatus(bc, labelSelector); builds != nil {
		state.Builds = builds
	} else if err != nil {
		return state, err
	}

	return state, nil
}

// TriggerBuild triggers a new build
func (b *buildConfig) TriggerBuild(bc *buildv1.BuildConfig, triggeredBy string) (bool, error) {
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
	buildRequest := newBuildRequest(triggeredBy, bc)
	build, err := b.client.BuildCli.BuildConfigs(bc.Namespace).Instantiate(context.TODO(), bc.Name, &buildRequest, metav1.CreateOptions{})
	if err != nil {
		return false, err
	}

	log.Info("Build triggered: ", build.Name)
	return true, nil
}

// TriggerBuildFromFile will be called by kogito-cli when a build from file is performed.
// When called a new build will be triggered with the request kogito resource or a tgz file.
func (b *buildConfig) TriggerBuildFromFile(namespace string, bodyPost io.Reader, options *buildv1.BinaryBuildRequestOptions) (*buildv1.Build, error) {

	result := &buildv1.Build{}
	buildName := fmt.Sprintf("%s%s", options.Name, "-builder")

	// before upload the file, make sure that the build exist
	err := b.waitForBuildConfig(b.checkBcRetries, b.checkBcRetriesInterval, func() error {
		if _, err := b.client.BuildCli.BuildConfigs(namespace).Get(context.TODO(), buildName, metav1.GetOptions{}); errors.IsNotFound(err) {
			log.Debugf("BuildConfig %s not found in the %s namespace", buildName, namespace)
			return err
		} else if err != nil {
			log.Debugf("Error while retrieving BuildConfig %s in namespace %s", buildName, namespace)
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	errPost := b.client.BuildCli.RESTClient().Post().
		Namespace(namespace).
		Resource("buildconfigs").
		Name(buildName).
		SubResource("instantiatebinary").
		Body(bodyPost).
		VersionedParams(options, runtime.NewParameterCodec(meta.GetRegisteredSchema())).
		Do(context.TODO()).
		Into(result)
	return result, errPost
}

// GetBuildsStatusByLabel checks the status of the builds for all builds with the given label
func (b *buildConfig) GetBuildsStatusByLabel(namespace, labelSelector string) (*v1alpha1.Builds, error) {
	list, err := b.client.BuildCli.Builds(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}
	status := &v1alpha1.Builds{}

	for _, item := range list.Items {
		log.Debugf("Checking status of build '%s'", item.Name)
		switch item.Status.Phase {
		case buildv1.BuildPhaseNew:
			status.New = append(status.New, item.Name)
		case buildv1.BuildPhasePending:
			status.Pending = append(status.Pending, item.Name)
		case buildv1.BuildPhaseRunning:
			status.Running = append(status.Running, item.Name)
		case buildv1.BuildPhaseComplete:
			status.Complete = append(status.Complete, item.Name)
		case buildv1.BuildPhaseFailed:
			status.Failed = append(status.Failed, item.Name)
		case buildv1.BuildPhaseError:
			status.Error = append(status.Error, item.Name)
		case buildv1.BuildPhaseCancelled:
			status.Cancelled = append(status.Cancelled, item.Name)
		default:
			status.New = append(status.New, item.Name)
		}
		log.Debugf("Build %s status is %s", item.Name, item.Status.Phase)
	}

	return status, nil
}

// GetBuildsStatus checks the status of the builds for the BuildConfig
func (b *buildConfig) GetBuildsStatus(bc *buildv1.BuildConfig, labelSelector string) (*v1alpha1.Builds, error) {
	if exists, err := b.checkBuildConfigExists(bc); !exists {
		return nil, err
	}

	list, err := b.client.BuildCli.Builds(bc.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}

	status := &v1alpha1.Builds{}

	for _, item := range list.Items {
		// it's the build from our buildConfig
		if strings.HasPrefix(item.Name, bc.Name) {
			log.Debugf("Checking status of build '%s'", item.Name)
			switch item.Status.Phase {
			case buildv1.BuildPhaseNew:
				status.New = append(status.New, item.Name)
			case buildv1.BuildPhasePending:
				status.Pending = append(status.Pending, item.Name)
			case buildv1.BuildPhaseRunning:
				status.Running = append(status.Running, item.Name)
			case buildv1.BuildPhaseComplete:
				status.Complete = append(status.Complete, item.Name)
			case buildv1.BuildPhaseFailed:
				status.Failed = append(status.Failed, item.Name)
			case buildv1.BuildPhaseError:
				status.Error = append(status.Error, item.Name)
			case buildv1.BuildPhaseCancelled:
				status.Cancelled = append(status.Cancelled, item.Name)
			default:
				status.New = append(status.New, item.Name)
			}
			log.Debugf("Build %s status is %s", item.Name, item.Status.Phase)
		}
	}

	return status, nil
}

func (b *buildConfig) checkBuildConfigExists(bc *buildv1.BuildConfig) (bool, error) {
	if _, err := b.client.BuildCli.BuildConfigs(bc.Namespace).Get(context.TODO(), bc.Name, metav1.GetOptions{}); err != nil && errors.IsNotFound(err) {
		log.Warnf("BuildConfig not found in namespace %s", bc.Namespace)
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

// newBuildRequest creates a new BuildRequest for the build
func newBuildRequest(triggeredBy string, bc *buildv1.BuildConfig) buildv1.BuildRequest {
	buildRequest := buildv1.BuildRequest{ObjectMeta: metav1.ObjectMeta{Name: bc.Name}}
	buildRequest.TriggeredBy = []buildv1.BuildTriggerCause{{Message: fmt.Sprintf("Triggered by %s operator", triggeredBy)}}
	meta.SetGroupVersionKind(&buildRequest.TypeMeta, meta.KindBuildRequest)
	return buildRequest
}

func (b *buildConfig) waitForBuildConfig(retries int, retryInterval time.Duration, f func() error) (err error) {
	for i := 0; ; i++ {
		err = f()
		if err == nil {
			return
		}

		if i >= (retries - 1) {
			break
		}
		time.Sleep(retryInterval)
		log.Debug("retrying after error:", err)
	}
	return fmt.Errorf("after %d attempts, last error: %v", retries, err)
}
