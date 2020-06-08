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

package install

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/test"

	"github.com/stretchr/testify/assert"

	operatorMarketplace "github.com/operator-framework/operator-marketplace/pkg/apis/operators/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_InstallOperator(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("install operator -p %s", ns)
	test.SetupCliTest(cli, context.CommandFactory{BuildCommands: BuildCommands}, &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})

	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Operator successfully deployed")
}

func Test_InstallOperatorNoNamespace(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("install operator -p %s", ns)
	test.SetupCliTest(cli, context.CommandFactory{BuildCommands: BuildCommands}, &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	_, _, err := test.ExecuteCli()
	assert.NoError(t, err)

	cli = "install operator --install-data-index"
	test.SetupCliTest(cli, context.CommandFactory{BuildCommands: BuildCommands}, &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Operator successfully deployed")
	assert.Contains(t, lines, "Kogito Data Index Service successfully installed in the Project")
}

func Test_InstallOperatorNoNamespaceWithForceFlag(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("install operator -p %s --force --image my-cool-image:latest", ns)
	test.SetupCliTest(cli, context.CommandFactory{BuildCommands: BuildCommands}, &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "Forcing installation of operator with custom image my-cool-image:latest")
	assert.Contains(t, lines, "Kogito Operator successfully deployed")
}

func Test_InstallOperatorNoNamespaceWithForceFlagWitNoCustomImage(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("install operator -p %s --force", ns)
	test.SetupCliTest(cli, context.CommandFactory{BuildCommands: BuildCommands}, &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	lines, _, err := test.ExecuteCli()
	assert.Error(t, err)
	assert.Contains(t, lines, "Error: force install flag is enabled but the custom operator image is missing")
}

func Test_InstallOperatorWithSupportServices(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("install operator -p %s --install-data-index --install-jobs-service --install-mgmt-console", ns)
	test.SetupCliTest(cli, context.CommandFactory{BuildCommands: BuildCommands}, &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "Data Index")
	assert.Contains(t, lines, "Jobs Service")
	assert.Contains(t, lines, "Management Console")
}

func Test_InstallOperatorWithDefaultChannel(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("install operator -p %s", ns)
	test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&operatorMarketplace.OperatorSource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      communityOperatorSource,
				Namespace: operatorMarketplaceNamespace,
			},
			Status: operatorMarketplace.OperatorSourceStatus{
				Packages: defaultOperatorPackageName,
			},
		})
	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Operator successfully subscribed")
}

func Test_InstallOperatorWithValidChannel(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("install operator -p %s --channel %s", ns, shared.DevPreviewChannel)
	test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&operatorMarketplace.OperatorSource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      communityOperatorSource,
				Namespace: operatorMarketplaceNamespace,
			},
			Status: operatorMarketplace.OperatorSourceStatus{
				Packages: defaultOperatorPackageName,
			},
		})
	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Operator successfully subscribed")
}

func Test_InstallOperatorWithInvalidChannel(t *testing.T) {
	ns := t.Name()
	invalidChannel := "testChannel"
	cli := fmt.Sprintf("install operator -p %s --channel %s", ns, invalidChannel)
	test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&operatorMarketplace.OperatorSource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      communityOperatorSource,
				Namespace: operatorMarketplaceNamespace,
			},
			Status: operatorMarketplace.OperatorSourceStatus{
				Packages: defaultOperatorPackageName,
			},
		})
	lines, _, err := test.ExecuteCli()
	assert.Error(t, err)
	assert.Contains(t, lines, "Invalid Kogito channel type testChannel, only alpha/dev-preview channels are allowed")
}

const (
	defaultOperatorPackageName   = "kogito-operator"
	communityOperatorSource      = "community-operators"
	operatorMarketplaceNamespace = "openshift-marketplace"
)
