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

package main

import (
	"bytes"
	"strings"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func runKogito() error {
	return Main()
}

func setupFakeKubeCli(initObjs ...runtime.Object) {
	s := meta.GetRegisteredSchema()
	kubeCli = &client.Client{ControlCli: fake.NewFakeClientWithScheme(s, initObjs...)}
}

func kogitoCliTestSetup(arg string) (*bytes.Buffer, *bytes.Buffer) {
	testErr := new(bytes.Buffer)
	testOut := new(bytes.Buffer)

	rootCmd.SetArgs(strings.Split(arg, " "))
	rootCmd.SetOut(testOut)
	rootCmd.SetErr(testErr)
	commandOutput = testOut

	return testOut, testErr
}

func executeCli(cli string, kubeObjects ...runtime.Object) (string, string, error) {
	setupFakeKubeCli(kubeObjects...)
	registerCommands()

	o, e := kogitoCliTestSetup(cli)

	err := rootCmd.Execute()

	defer func() {
		rootCmd = nil
	}()

	return o.String(), e.String(), err
}
