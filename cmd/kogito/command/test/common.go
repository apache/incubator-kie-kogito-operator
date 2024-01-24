// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package test

import (
	"bytes"
	"strings"

	"github.com/apache/incubator-kie-kogito-operator/core/test"
	"github.com/spf13/cobra"

	"github.com/apache/incubator-kie-kogito-operator/cmd/kogito/command/context"
	"github.com/apache/incubator-kie-kogito-operator/core/client"
	clitest "github.com/apache/incubator-kie-kogito-operator/core/client/test"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	testErr     *bytes.Buffer
	testOut     *bytes.Buffer
	rootCommand *cobra.Command
)

// CLITestContext holds the text context for the CLI Unit Tests.
// Use SetupCliTest or SetupCliTestWithKubeClient to get a reference for your test cases
type CLITestContext interface {
	ExecuteCli() (string, string, error)
	ExecuteCliCmd(cmd string) (string, string, error)
	GetClient() *client.Client
}

type cliTestContext struct {
	*context.CommandContext
	client *client.Client
}

// SetupCliTest creates the CLI default test environment. The mocked Kubernetes client does not support OpenShift.
func SetupCliTest(cli string, factory context.CommandFactory, kubeObjects ...runtime.Object) CLITestContext {
	return SetupCliTestWithKubeClient(cli, factory, test.NewFakeClientBuilder().AddK8sObjects(kubeObjects...).Build())
}

// SetupCliTestWithKubeClient Setup a CLI test environment with the given Kubernetes client
func SetupCliTestWithKubeClient(cmd string, factory context.CommandFactory, kubeCli *client.Client) CLITestContext {
	testErr = new(bytes.Buffer)
	testOut = new(bytes.Buffer)

	ctx := &context.CommandContext{Client: kubeCli}

	kogitoRootCmd := context.NewRootCommand(ctx, testOut)
	kogitoRootCmd.Command().SetArgs(strings.Split(cmd, " "))
	kogitoRootCmd.Command().SetOut(testOut)
	kogitoRootCmd.Command().SetErr(testErr)

	rootCommand = kogitoRootCmd.Command()

	factory.BuildCommands(ctx, rootCommand)

	return &cliTestContext{CommandContext: ctx, client: kubeCli}
}

// ExecuteCli executes the CLI setup before executing the test
func (c *cliTestContext) ExecuteCli() (string, string, error) {
	err := rootCommand.Execute()
	return testOut.String(), testErr.String(), err
}

// ExecuteCliCmd executes the given command in the actual context
func (c *cliTestContext) ExecuteCliCmd(cmd string) (string, string, error) {
	rootCommand.SetArgs(strings.Split(cmd, " "))
	err := rootCommand.Execute()
	return testOut.String(), testErr.String(), err
}

func (c *cliTestContext) GetClient() *client.Client {
	return c.client
}

// OverrideKubeConfig overrides the default KUBECONFIG location to a temporary one
func OverrideKubeConfig() (teardown func()) {
	_, teardown = clitest.OverrideDefaultKubeConfig()
	return
}

// OverrideKubeConfigAndCreateContextInNamespace overrides the default KUBECONFIG location to a temporary one and creates a mock context in the given namespace
func OverrideKubeConfigAndCreateContextInNamespace(namespace string) (teardown func()) {
	_, teardown = clitest.OverrideDefaultKubeConfigWithNamespace(namespace)
	return
}

// OverrideKubeConfigAndCreateDefaultContext initializes the default KUBECONFIG location to a temporary one and creates a mock context in the "default" namespace
func OverrideKubeConfigAndCreateDefaultContext() (teardown func()) {
	_, teardown = clitest.OverrideDefaultKubeConfigEmptyContext()
	return
}
