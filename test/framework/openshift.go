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

package framework

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ocapps "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	"github.com/kiegroup/kogito-cloud-operator/test/config"
)

// WaitForBuildComplete waits for a build to be completed
func WaitForBuildComplete(namespace, buildName string, timeoutInMin int) error {
	return WaitForOnOpenshift(namespace, fmt.Sprintf("Build %s complete", buildName), timeoutInMin,
		func() (bool, error) {
			bc := buildv1.BuildConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      buildName,
					Namespace: namespace,
				},
			}
			builds, err := openshift.BuildConfigC(kubeClient).GetBuildsStatus(&bc, fmt.Sprintf("%s=%s", openshift.BuildConfigLabelSelector, buildName))

			if err != nil {
				return false, fmt.Errorf("Error while fetching buildconfig %s: %v", buildName, err)
			} else if builds == nil || len(builds.Complete) < 1 {
				return false, nil
			}

			return true, nil
		})
}

// WaitForBuildConfigCreated waits for a build config to be created
func WaitForBuildConfigCreated(namespace, buildConfigName string, timeoutInMin int) error {
	return WaitForOnOpenshift(namespace, fmt.Sprintf("BuildConfig %s created", buildConfigName), timeoutInMin,
		func() (bool, error) {
			if bc, err := getBuildConfig(namespace, buildConfigName); err != nil {
				return false, err
			} else if bc == nil {
				return false, nil
			}
			return true, nil
		})
}

// WaitForBuildConfigCreatedWithWebhooks waits for a build config to be created with webhooks
func WaitForBuildConfigCreatedWithWebhooks(namespace, buildConfigName string, expectedWebhooks []v1beta1.WebHookSecret, timeoutInMin int) error {
	return WaitForOnOpenshift(namespace, fmt.Sprintf("BuildConfig %s created with webhooks", buildConfigName), timeoutInMin,
		func() (bool, error) {
			if bc, err := getBuildConfig(namespace, buildConfigName); err != nil {
				return checkWebhooksInBuildConfig(namespace, bc.Spec.Triggers, expectedWebhooks), err
			} else if bc == nil {
				return false, nil
			}
			return true, nil
		})
}

func checkWebhooksInBuildConfig(namespace string, actual []buildv1.BuildTriggerPolicy, expected []v1beta1.WebHookSecret) bool {
	for _, expectedWebhook := range expected {
		for _, actualTrigger := range actual {
			var typedTrigger *buildv1.WebHookTrigger
			switch expectedWebhook.Type {
			case v1beta1.GitHubWebHook:
				typedTrigger = actualTrigger.GitHubWebHook
			case v1beta1.GenericWebHook:
				typedTrigger = actualTrigger.GenericWebHook
			}

			if typedTrigger == nil || typedTrigger.SecretReference.Name != expectedWebhook.Secret {
				return false
			}
		}
	}

	return true
}

func getBuildConfig(namespace, buildConfigName string) (*buildv1.BuildConfig, error) {
	bc := &buildv1.BuildConfig{}
	if exists, err := kubernetes.ResourceC(kubeClient).FetchWithKey(types.NamespacedName{Name: buildConfigName, Namespace: namespace}, bc); err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("Error while trying to look for BuildConfig %s: %v ", buildConfigName, err)
	} else if errors.IsNotFound(err) || !exists {
		return nil, nil
	}
	return bc, nil
}

// WaitForDeploymentConfigRunning waits for a deployment config to be running, with a specific number of pod
func WaitForDeploymentConfigRunning(namespace, dcName string, podNb int, timeoutInMin int) error {
	return WaitForOnOpenshift(namespace, fmt.Sprintf("DeploymentConfig %s running", dcName), timeoutInMin,
		func() (bool, error) {
			if dc, err := GetDeploymentConfig(namespace, dcName); err != nil {
				return false, err
			} else if dc == nil {
				return false, nil
			} else {
				GetLogger(namespace).Debug("Deployment config has", "available replicas", dc.Status.AvailableReplicas)
				return dc.Status.AvailableReplicas == int32(podNb), nil
			}
		}, CheckPodsByDeploymentConfigInError(namespace, dcName))
}

// GetDeploymentConfig retrieves a deployment config
func GetDeploymentConfig(namespace, dcName string) (*ocapps.DeploymentConfig, error) {
	dc := &ocapps.DeploymentConfig{}
	if exists, err := kubernetes.ResourceC(kubeClient).FetchWithKey(types.NamespacedName{Name: dcName, Namespace: namespace}, dc); err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("Error while trying to look for DeploymentConfig %s: %v ", dcName, err)
	} else if errors.IsNotFound(err) || !exists {
		return nil, nil
	}
	return dc, nil
}

// WaitForRoute waits for a route to be available
func WaitForRoute(namespace, routeName string, timeoutInMin int) error {
	return WaitForOnOpenshift(namespace, fmt.Sprintf("Route %s available", routeName), timeoutInMin,
		func() (bool, error) {
			route, err := GetRoute(namespace, routeName)
			if err != nil || route == nil {
				return false, err
			}

			return true, nil
		})
}

// GetRoute retrieves a route
func GetRoute(namespace, routeName string) (*routev1.Route, error) {
	route := &routev1.Route{}
	if exists, err :=
		kubernetes.ResourceC(kubeClient).FetchWithKey(types.NamespacedName{Name: routeName, Namespace: namespace}, route); err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	} else {
		return route, nil
	}
}

func createHTTPRoute(namespace, serviceName string) error {
	GetLogger(namespace).Info("Creating HTTP route", "serviceName", serviceName)

	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: serviceName,
			},
		},
	}
	if err := kubernetes.ResourceC(kubeClient).Create(route); err != nil {
		return err
	}
	return nil
}

// GetRouteURI retrieves a route URI
func GetRouteURI(namespace, serviceName string) (string, error) {
	if err := WaitForRoute(namespace, serviceName, 2); err != nil {
		return "", fmt.Errorf("Route %s does not exist in namespace %s: %v", serviceName, namespace, err)
	}

	route, err := GetRoute(namespace, serviceName)
	if err != nil || route == nil {
		return "", err
	}
	host := route.Spec.Host

	protocol := "http"
	port := "80"
	if route.Spec.TLS != nil {
		protocol = "https"
		port = "443"
	}

	uri := protocol + "://" + host + ":" + port
	return uri, nil
}

// WaitForOnOpenshift waits for a specification condition
func WaitForOnOpenshift(namespace, display string, timeoutInMin int, condition func() (bool, error), errorConditions ...func() (bool, error)) error {
	return WaitFor(namespace, display, GetOpenshiftDurationFromTimeInMin(timeoutInMin), condition, errorConditions...)
}

// GetOpenshiftDurationFromTimeInMin will calculate the time depending on the configured cluster load factor
func GetOpenshiftDurationFromTimeInMin(timeoutInMin int) time.Duration {
	return time.Duration(timeoutInMin*config.GetLoadFactor()) * time.Minute
}
