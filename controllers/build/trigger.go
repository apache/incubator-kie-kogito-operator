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
	"context"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	v1 "github.com/openshift/api/build/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"strings"
	"sync"
	"time"
)

const (
	cancelUpdateTimeout = 30 * time.Second
	poolWaitTimeout     = 500 * time.Millisecond
	triggeredBy         = "KogitoBuild controller from Kogito Operator"
)

// StartNewBuild starts a new build for the given KogitoBuild and BuildConfig.
// This action will cancel any other running builds for the given BC
func StartNewBuild(buildConfig *v1.BuildConfig, client *client.Client) error {
	if err := cancelRunningBuilds(buildConfig, client); err != nil {
		return err
	}
	if _, err := openshift.BuildConfigC(client).TriggerBuild(buildConfig, triggeredBy); err != nil {
		log.Error(err, "Failed to start a new build", "For Build Config", buildConfig.Name)
		return err
	}
	return nil
}

// cancelRunningBuilds cancels any running builds for the given BuildConfig
func cancelRunningBuilds(buildConfig *v1.BuildConfig, client *client.Client) error {
	builds, err := client.BuildCli.Builds(buildConfig.Namespace).List(context.TODO(),
		metav1.ListOptions{LabelSelector: strings.Join([]string{openshift.BuildConfigLabelSelector, buildConfig.Name}, "=")},
	)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	for _, b := range builds.Items {
		if b.Status.Phase == v1.BuildPhaseNew ||
			b.Status.Phase == v1.BuildPhasePending ||
			b.Status.Phase == v1.BuildPhaseRunning {
			var cancelError error
			wg.Add(1)
			go func(build *v1.Build) {
				defer wg.Done()
				err := wait.Poll(poolWaitTimeout, cancelUpdateTimeout, func() (bool, error) {
					build.Status.Cancelled = true
					_, err := client.BuildCli.Builds(build.Namespace).Update(context.TODO(), build, metav1.UpdateOptions{})
					if err == nil {
						return true, nil
					} else if errors.IsConflict(err) {
						// try again, someone just updated our status
						build, err = client.BuildCli.Builds(build.Namespace).Get(context.TODO(), build.Name, metav1.GetOptions{})
						return false, err
					}
					return true, err
				})
				if err != nil {
					log.Error(err, "Failed to cancel", "Build", build.Name)
					cancelError = err
					return
				}
				// wait for the build to be cancelled
				err = wait.Poll(poolWaitTimeout, cancelUpdateTimeout, func() (bool, error) {
					updatedBuild, err := client.BuildCli.Builds(build.Namespace).Get(context.TODO(), build.Name, metav1.GetOptions{})
					if err != nil {
						return true, err
					}
					if updatedBuild.Status.Phase == v1.BuildPhaseCancelled {
						log.Info("Successfully cancelled", "Build", build.Name, "Namespace", build.Namespace)
						return true, nil
					}
					return false, nil
				})
				if err != nil {
					log.Error(err, "Failed to fetch build during cancelling check phase", "Build", build.Name)
					cancelError = err
					return
				}
			}(&b)
			wg.Wait()
			return cancelError
		}
	}
	return nil
}
