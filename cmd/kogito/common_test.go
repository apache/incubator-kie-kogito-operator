package main

import (
	"bytes"
	"strings"

	"github.com/kiegroup/kogito-cloud-operator/pkg/log"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kiegroup/kogito-cloud-operator/pkg/inventory"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func runKogito() error {
	return Main()
}

func setupFakeKubeCli(initObjs ...runtime.Object) {
	kubeCli = &inventory.Client{Cli: fake.NewFakeClient(initObjs...)}
}

func kogitoCliTestSetup(arg string) (*bytes.Buffer, *bytes.Buffer) {
	testErr := new(bytes.Buffer)
	testOut := new(bytes.Buffer)

	rootCmd.SetArgs(strings.Split(arg, " "))
	rootCmd.SetOut(testOut)
	rootCmd.SetErr(testErr)
	// every output command shares the same logger
	log.SetOutput(testOut)

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
