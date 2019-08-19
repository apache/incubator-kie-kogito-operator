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
