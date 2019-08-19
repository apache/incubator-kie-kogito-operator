package main

import (
	"bytes"
	"strings"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"k8s.io/apimachinery/pkg/runtime"
)

func runKogito() error {
	return Main()
}

func setupFakeKubeCli(initObjs ...runtime.Object) {
	kubeCli = &client.Client{ControlCli: fake.NewFakeClient(initObjs...)}
}

func kogitoCliTestSetup(arg string) (*bytes.Buffer, *bytes.Buffer) {
	testErr := new(bytes.Buffer)
	testOut := new(bytes.Buffer)

	rootCmd.SetArgs(strings.Split(arg, " "))
	rootCmd.SetOut(testOut)
	rootCmd.SetErr(testErr)
	setDefaultLog("kogito_cli_test", testOut)

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
