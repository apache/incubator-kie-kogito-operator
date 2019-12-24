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

package shared

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
	olmapiv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1"
	olmapiv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/operator-framework/operator-marketplace/pkg/apis/operators/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

const (
	defaultOperatorPackageName   = "kogito-operator"
	defaultOperatorChannelName   = "alpha"
	communityOperatorSource      = "community-operators"
	operatorMarketplaceNamespace = "openshift-marketplace"
)

// isOperatorAvailableInOperatorHub will check if the Kogito Operator is available in OperatorHub (on OpenShift)
func isOperatorAvailableInOperatorHub(kubeCli *client.Client) (bool, error) {
	log := context.GetDefaultLogger()
	log.Debug("Trying to find if Kogito Operator is available in the OperatorHub")
	operatorSource := &v1.OperatorSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      communityOperatorSource,
			Namespace: operatorMarketplaceNamespace,
		},
	}
	exists, err := kubernetes.ResourceC(kubeCli).Fetch(operatorSource)
	if err != nil {
		return false, err
	}

	log.Debugf("Finishing fetch the OperatorHub for Kogito Operator in namespace %s", operatorMarketplaceNamespace)
	log.Debugf("OperatorSource named as %s created at %s", operatorSource.Name, operatorSource.CreationTimestamp)
	log.Debugf("OperatorSource %s has the following packages: %s", operatorSource.Name, operatorSource.Status.Packages)
	if !exists {
		return false, nil
	}

	return strings.Contains(operatorSource.Status.Packages, defaultOperatorPackageName), nil
}

// installOperatorWithOperatorHub installs the Kogito Operator via OperatorHub custom resources, works for OCP 4.x
// checks if a subscription to the given Kogito Operator package already exists. Doesn't create if one is in place.
// see: https://docs.openshift.com/container-platform/4.2/operators/olm-adding-operators-to-cluster.html#olm-installing-operator-from-operatorhub-using-cli_olm-adding-operators-to-a-cluster
func installOperatorWithOperatorHub(namespace string, cli *client.Client) error {
	log := context.GetDefaultLogger()
	log.Debug("Trying to install Kogito Operator via Subscription to the OperatorHub")
	if err := createOperatorGroupIfNotExists(namespace, cli); err != nil {
		return err
	}

	subs := &olmapiv1alpha1.SubscriptionList{}
	if err := kubernetes.ResourceC(cli).ListWithNamespace(namespace, subs); err != nil {
		return err
	}

	for _, sub := range subs.Items {
		if sub.Spec.Package == defaultOperatorPackageName &&
			sub.Spec.CatalogSource == communityOperatorSource {
			log.Warnf("Found subscription %s with package %s and catalog source %s. Won't create a new one",
				sub.Name, sub.Spec.Package, sub.Spec.CatalogSource)
			return nil
		}
	}

	subscription := &olmapiv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultOperatorPackageName,
			Namespace: namespace,
		},
		Spec: &olmapiv1alpha1.SubscriptionSpec{
			Package:                defaultOperatorPackageName,
			CatalogSource:          communityOperatorSource,
			CatalogSourceNamespace: operatorMarketplaceNamespace,
			Channel:                defaultOperatorChannelName,
		},
	}
	log.Debug("About to create a new subscription for the Kogito Operator")
	if err := kubernetes.ResourceC(cli).Create(subscription); err != nil {
		return err
	}

	return nil
}

// createOperatorGroupIfNotExists creates a new OperatorGroup for `SingleNamespace` installations mode.
// Only creates it if we don't have one in the given namespace that targets the given namespace.
func createOperatorGroupIfNotExists(namespace string, cli *client.Client) error {
	groups := &olmapiv1.OperatorGroupList{}
	if err := kubernetes.ResourceC(cli).ListWithNamespace(namespace, groups); err != nil {
		return err
	}

	// inspect target namespace in the groups
	for _, group := range groups.Items {
		for _, ns := range group.Spec.TargetNamespaces {
			if ns == namespace {
				return nil
			}
		}
	}

	// we don't have a group, let's create a new one
	groupName := fmt.Sprintf("%s-%s", namespace, util.RandomSuffix())
	group := &olmapiv1.OperatorGroup{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: groupName},
		Spec:       olmapiv1.OperatorGroupSpec{TargetNamespaces: []string{namespace}},
	}
	if err := kubernetes.ResourceC(cli).Create(group); err != nil {
		return err
	}
	return nil
}
