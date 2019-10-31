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

package test

import (
	"bytes"
	"github.com/spf13/cobra"
	"strings"

	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	testErr     *bytes.Buffer
	testOut     *bytes.Buffer
	rootCommand *cobra.Command
)

// SetupFakeKubeCli will create a fake kube client for your tests
func SetupFakeKubeCli(initObjs ...runtime.Object) *client.Client {
	s := meta.GetRegisteredSchema()
	return &client.Client{ControlCli: fake.NewFakeClientWithScheme(s, initObjs...)}
}

// SetupCliTest creates the infrastructure for the CLI test
func SetupCliTest(cli string, factory context.CommandFactory, kubeObjects ...runtime.Object) (ctx *context.CommandContext) {
	kubeCli := SetupFakeKubeCli(kubeObjects...)
	testErr = new(bytes.Buffer)
	testOut = new(bytes.Buffer)

	ctx = &context.CommandContext{Client: kubeCli}

	kogitoRootCmd := context.NewRootCommand(ctx, testOut)
	kogitoRootCmd.Command().SetArgs(strings.Split(cli, " "))
	kogitoRootCmd.Command().SetOut(testOut)
	kogitoRootCmd.Command().SetErr(testErr)

	rootCommand = kogitoRootCmd.Command()

	factory.BuildCommands(ctx, rootCommand)

	return ctx
}

//ExecuteCli will execute the cli setup before
func ExecuteCli() (string, string, error) {
	if rootCommand == nil {
		panic("RootCommand reference not found. Try calling SetupCliTest first ")
	}
	err := rootCommand.Execute()

	defer func() {
		rootCommand = nil
		testErr = nil
		testOut = nil
	}()

	return testOut.String(), testErr.String(), err
}
