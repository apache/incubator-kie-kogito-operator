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

package steps

import (
	"fmt"

	"github.com/kiegroup/kogito-operator/api"

	"github.com/cucumber/godog"
	"github.com/kiegroup/kogito-operator/test/pkg/config"
	"github.com/kiegroup/kogito-operator/test/pkg/framework"
)

const sourceLocation = "src/main/resources"

func registerKogitoDeployFilesSteps(ctx *godog.ScenarioContext, data *Data) {
	// Deploy steps
	ctx.Step(`^Deploy (quarkus|springboot) file "([^"]*)" from example service "([^"]*)"$`, data.deployFileFromExampleService)
	ctx.Step(`^Deploy (quarkus|springboot) folder from example service "([^"]*)"$`, data.deployFolderFromExampleService)
}

// Deploy steps

func (data *Data) deployFileFromExampleService(runtimeType, file, serviceName string) error {
	sourceFilePath := fmt.Sprintf(`%s/%s/%s/%s`, data.Location[KogitoExamples], serviceName, sourceLocation, file)
	return deploySourceFilesFromPath(data.Namespace, runtimeType, serviceName, sourceFilePath)
}

func (data *Data) deployFolderFromExampleService(runtimeType, serviceName string) error {
	sourceFolderPath := fmt.Sprintf(`%s/%s/%s`, data.Location[KogitoExamples], serviceName, sourceLocation)
	return deploySourceFilesFromPath(data.Namespace, runtimeType, serviceName, sourceFolderPath)
}

func deploySourceFilesFromPath(namespace, runtimeType, serviceName, path string) error {
	framework.GetLogger(namespace).Info("Deploying example with source files", "runtimeType", runtimeType, "serviceName", serviceName, "path", path)

	buildHolder, err := getKogitoBuildConfiguredStub(namespace, runtimeType, serviceName, nil)
	if err != nil {
		return err
	}

	buildHolder.KogitoBuild.GetSpec().SetType(api.LocalSourceBuildType)
	buildHolder.KogitoBuild.GetSpec().GetGitSource().SetURI(path)

	err = framework.DeployKogitoBuild(namespace, framework.GetDefaultInstallerType(), buildHolder)
	if err != nil {
		return err
	}

	// If we don't use Kogito CLI then upload target folder using OC client
	if config.IsCrDeploymentOnly() {
		return framework.WaitForOnOpenshift(namespace, fmt.Sprintf("Build '%s-builder' to start", serviceName), defaultTimeoutToStartBuildInMin,
			func() (bool, error) {
				_, err := framework.CreateCommand("oc", "start-build", serviceName+"-builder", "--from-file="+path, "-n", namespace).WithLoggerContext(namespace).Execute()
				return err == nil, err
			})
	}

	return nil
}
