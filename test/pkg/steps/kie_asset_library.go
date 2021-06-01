// Copyright 2021 Red Hat, Inc. and/or its affiliates
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
	"os"

	"github.com/cucumber/godog"
	"github.com/kiegroup/kogito-operator/test/pkg/framework"
	"github.com/kiegroup/kogito-operator/test/pkg/steps/mappers"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// TODO
// Should be externalized to config.go ?
const (
	KieAssetLibraryGitRepositoryURI         = "https://github.com/jstastny-cz/kie-asset-library-poc"
	KieAssetLibraryGitRepositoryBranch      = "main"
	KieAssetReMarshallerGitRepositoryURI    = "https://github.com/jomarko/kie-assets-re-marshaller"
	KieAssetReMarshallerGitRepositoryBranch = "master"
)

// registerMavenSteps register all existing Maven steps
func registerKieAssetLibrarySteps(ctx *godog.ScenarioContext, data *Data) {
	ctx.Step("^Project kie-asset-re-marshaller is cloned$", data.projectKieAssetReMarshallerIsCloned)
	ctx.Step("^Project kie-asset-library is cloned$", data.projectKieAssetLibraryIsCloned)
	ctx.Step("^Project kie-asset-library is built by maven with configuration:$", data.projectKieAssetLibraryIsBuiltByMavenWithConfiguration)
	ctx.Step("^Project \"([^\"]*)\" is generated by kie-asset-library$", data.projectIsGeneratedByKieAssetLibrary)
	ctx.Step("^Project \"([^\"]*)\" in kie-asset-library is built by maven$", data.projectInKieAssetLibraryIsBuiltByMaven)
	ctx.Step("^Project \"([^\"]*)\" assets are re-marshalled by VS Code$", data.projectAssetsAreRemarshalledByVsCode)
	ctx.Step(`^Deploy (quarkus|springboot) project "([^"]*)" from kie-asset-library with configuration:$`, data.deployKieAssetLibraryProjectOnOpenshift)
}

func (data *Data) projectKieAssetLibraryIsBuiltByMavenWithConfiguration(table *godog.Table) error {

	mavenConfig := &mappers.MavenCommandConfig{}
	if table != nil && len(table.Rows) > 0 {
		err := mappers.MapMavenCommandConfigTable(table, mavenConfig)
		if err != nil {
			return err
		}
	}

	return data.localPathBuiltByMavenWithProfileAndOptions(data.Location[KieAssetLibrary], mavenConfig)
}

func (data *Data) projectKieAssetReMarshallerIsCloned() error {
	framework.GetLogger(data.Namespace).Info("Cloning kie-asset-re-marshaller project", "URI", KieAssetReMarshallerGitRepositoryURI, "branch", KieAssetReMarshallerGitRepositoryBranch, "clonedLocation", data.Location[KieAssetReMarshaller])

	cloneOptions := &git.CloneOptions{
		URL:          KieAssetReMarshallerGitRepositoryURI,
		SingleBranch: true,
	}

	var err error
	reference := KieAssetReMarshallerGitRepositoryBranch

	// Try cloning as branch reference
	cloneOptions.ReferenceName = plumbing.NewBranchReferenceName(reference)
	err = cloneRepository(data.Location[KieAssetReMarshaller], cloneOptions)
	// If branch clone was successful then return, otherwise try other cloning options
	if err == nil {
		return nil
	}

	// If branch cloning failed then try cloning as tag
	cloneOptions.ReferenceName = plumbing.NewTagReferenceName(reference)
	err = cloneRepository(data.Location[KieAssetReMarshaller], cloneOptions)

	return err
}

func (data *Data) projectKieAssetLibraryIsCloned() error {
	framework.GetLogger(data.Namespace).Info("Cloning kie-asset-library project", "URI", KieAssetLibraryGitRepositoryURI, "branch", KieAssetLibraryGitRepositoryBranch, "clonedLocation", data.Location[KieAssetLibrary])

	cloneOptions := &git.CloneOptions{
		URL:          KieAssetLibraryGitRepositoryURI,
		SingleBranch: true,
	}

	var err error
	reference := KieAssetLibraryGitRepositoryBranch

	// Try cloning as branch reference
	cloneOptions.ReferenceName = plumbing.NewBranchReferenceName(reference)
	err = cloneRepository(data.Location[KieAssetLibrary], cloneOptions)
	// If branch clone was successful then return, otherwise try other cloning options
	if err == nil {
		return nil
	}

	// If branch cloning failed then try cloning as tag
	cloneOptions.ReferenceName = plumbing.NewTagReferenceName(reference)
	err = cloneRepository(data.Location[KieAssetLibrary], cloneOptions)

	return err
}

func (data *Data) projectIsGeneratedByKieAssetLibrary(project string) error {
	if _, err := os.Stat(data.Location[KieAssetLibrary] + "/kie-assets-library-generate/target/" + project); !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (data *Data) projectInKieAssetLibraryIsBuiltByMaven(project string) error {
	projectLocation := data.Location[KieAssetLibrary] + "/kie-assets-library-generate/target/" + project
	output, errCode := framework.CreateMavenCommand(projectLocation).
		SkipTests().
		Execute("clean", "install")
	framework.GetLogger(data.Namespace).Debug(output)
	if errCode != nil {
		framework.GetLogger(data.Namespace).Warn(output)
		framework.GetLogger(data.Namespace).Warn(project + " 'mvn clean install' failed due to: " + errCode.Error())
	}
	return errCode
}

func (data *Data) projectAssetsAreRemarshalledByVsCode(project string) error {
	installOutput, installErr := framework.CreateCommand("npm", "install").
		InDirectory(data.Location[KieAssetReMarshaller]).
		Execute()
	framework.GetLogger(data.Namespace).Debug(installOutput)

	if installErr != nil {
		framework.GetLogger(data.Namespace).Warn(installOutput)
		framework.GetLogger(data.Namespace).Warn(project + " 'mvn clean install' failed due to: " + installErr.Error())
		return installErr
	}

	kieProjectParameter := "KIE_PROJECT=" + data.Location[KieAssetLibrary] + "/kie-assets-library-generate/target/" + project
	runOutput, runErr := framework.CreateCommand("npm", "run", "test:it", kieProjectParameter).
		InDirectory(data.Location[KieAssetReMarshaller]).
		Execute()
	framework.GetLogger(data.Namespace).Debug(runOutput)

	if runErr != nil {
		framework.GetLogger(data.Namespace).Warn(runOutput)
		framework.GetLogger(data.Namespace).Warn(project + " 'npm run test:it' failed due to: " + runErr.Error())
	}

	return runErr
}

func (data *Data) deployKieAssetLibraryProjectOnOpenshift(runtimeType, project string, table *godog.Table) error {
	binaryFolder := data.Location[KieAssetLibrary] + "/kie-assets-library-generate/target/" + project

	return data.deployTargetFolderOnOpenshift(runtimeType, project, binaryFolder, table)
}
