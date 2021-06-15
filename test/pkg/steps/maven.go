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

package steps

import (
	"github.com/cucumber/godog"
	"github.com/kiegroup/kogito-operator/test/pkg/framework"
	"github.com/kiegroup/kogito-operator/test/pkg/steps/mappers"
)

/*
	DataTable for Maven:
	| profile | profile        |
	| profile | profile2       |
	| option  | -Doption=true  |
	| option  | -Doption2=true |
	| native  | enabled        |
*/

const (
	// DefaultMavenBuiltExampleRegex is the regex for building maven example
	DefaultMavenBuiltExampleRegex = "Local example service \"([^\"]*)\" is built by Maven"
	nativeProfile                 = "native"
)

// registerMavenSteps register all existing Maven steps
func registerMavenSteps(ctx *godog.ScenarioContext, data *Data) {
	ctx.Step("^"+DefaultMavenBuiltExampleRegex+"$", data.localServiceBuiltByMaven)
	ctx.Step("^"+DefaultMavenBuiltExampleRegex+" with configuration:$", data.localServiceBuiltByMavenWithConfiguration)
}

// Build local service
func (data *Data) localServiceBuiltByMaven(serviceName string) error {
	return data.localServiceBuiltByMavenWithConfiguration(serviceName, nil)
}

// Build local service with configuration
func (data *Data) localServiceBuiltByMavenWithConfiguration(serviceName string, table *godog.Table) error {
	mavenConfig := &mappers.MavenCommandConfig{}
	if table != nil && len(table.Rows) > 0 {
		err := mappers.MapMavenCommandConfigTable(table, mavenConfig)
		if err != nil {
			return err
		}
	}
	return data.localServiceBuiltByMavenWithProfileAndOptions(serviceName, mavenConfig)
}

// Build local service with profile and additional options
func (data *Data) localServiceBuiltByMavenWithProfileAndOptions(serviceName string, mavenConfig *mappers.MavenCommandConfig) error {
	serviceRepositoryPath := data.Location[KogitoExamples] + "/" + serviceName
	return data.localPathBuiltByMavenWithProfileAndOptions(serviceRepositoryPath, mavenConfig)
}

func (data *Data) localPathBuiltByMavenWithProfileAndOptions(serviceRepositoryPath string, mavenConfig *mappers.MavenCommandConfig) error {

	mvnCmd := framework.CreateMavenCommand(serviceRepositoryPath).
		SkipTests().
		UpdateArtifacts().
		Options(mavenConfig.Options...).
		Profiles(mavenConfig.Profiles...).
		WithLoggerContext(data.Namespace)

	if mavenConfig.Native {
		mvnCmd = mvnCmd.Profiles(nativeProfile)
	}

	output, err := mvnCmd.Execute("clean", "package")
	framework.GetLogger(data.Namespace).Debug(output)

	if err != nil {
		framework.GetLogger(data.Namespace).Warn("'mvn clean package' failed for: " + serviceRepositoryPath)
	}

	return err
}
