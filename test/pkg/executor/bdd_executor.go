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

package executor

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"

	"github.com/kiegroup/kogito-operator/test/pkg/config"
	"github.com/kiegroup/kogito-operator/test/pkg/framework"
	"github.com/kiegroup/kogito-operator/test/pkg/gherkin"
	"github.com/kiegroup/kogito-operator/test/pkg/installers"
	"github.com/kiegroup/kogito-operator/test/pkg/steps"

	flag "github.com/spf13/pflag"
)

const (
	disabledTag    = "@disabled"
	cliTag         = "@cli"
	smokeTag       = "@smoke"
	performanceTag = "@performance"
)

var (
	godogOpts = godog.Options{
		Output:    colors.Colored(os.Stdout),
		Format:    "junit",
		Randomize: time.Now().UTC().UnixNano(),
		Tags:      disabledTag,
	}

	// PreRegisterStepsHook appends hooks to be executed before default steps are registered
	PreRegisterStepsHook func(ctx *godog.ScenarioContext, d *steps.Data)
	// AfterScenarioHook appends hooks to be executed before default AfterScenario phase
	AfterScenarioHook func(scenario *godog.Scenario, d *steps.Data) error
)

func init() {
	config.BindFlags(flag.CommandLine)
	godog.BindCommandLineFlags("godog.", &godogOpts)
}

// ExecuteBDDTests executes BDD tests
func ExecuteBDDTests(beforeTestsExecution func(godogOpts *godog.Options) error) {
	flag.Parse()
	godogOpts.Paths = flag.Args()

	configureTags()
	configureTestOutput()

	if beforeTestsExecution != nil {
		if err := beforeTestsExecution(&godogOpts); err != nil {
			panic(err)
		}
	}

	features, err := gherkin.ParseFeatures(godogOpts.Tags, godogOpts.Paths)
	if err != nil {
		panic(fmt.Errorf("Error parsing features: %v", err))
	}
	if config.IsShowScenarios() || config.IsShowSteps() {
		showScenarios(features, config.IsShowSteps())
	}

	if !config.IsDryRun() {
		if !config.IsCrDeploymentOnly() || gherkin.MatchingFeatureWithTags(cliTag, features) {
			// Check CLI binary is existing if needed
			if exits, err := framework.CheckCliBinaryExist(); err != nil {
				panic(fmt.Errorf("Error trying to get CLI binary %v", err))
			} else if !exits {
				panic("CLI Binary does not exist on specified path")
			}
		}

		status := godog.TestSuite{
			Name:                 "godogs",
			TestSuiteInitializer: initializeTestSuite,
			ScenarioInitializer:  initializeScenario,
			Options:              &godogOpts,
		}.Run()

		os.Exit(status)
	}
	os.Exit(0)
}

func configureTags() {
	if config.IsSmokeTests() {
		// Filter with smoke tag
		appendTag(smokeTag)
	} else if !strings.Contains(godogOpts.Tags, performanceTag) {
		if config.IsPerformanceTests() {
			// Turn on performance tests
			appendTag(performanceTag)
		} else {
			// Turn off performance tests
			appendTag("~" + performanceTag)
		}
	}

	if !strings.Contains(godogOpts.Tags, disabledTag) {
		// Ignore disabled tag
		appendTag("~" + disabledTag)
	}
}

func appendTag(tag string) {
	if len(godogOpts.Tags) > 0 {
		godogOpts.Tags += " && "
	}
	godogOpts.Tags += tag
}

func configureTestOutput() {
	logFolder := framework.GetLogFolder()
	if err := framework.CreateFolder(logFolder); err != nil {
		panic(fmt.Errorf("Error while creating log folder %s: %v", logFolder, err))
	}

	mainLogFile, err := os.Create(fmt.Sprintf("%s/%s", logFolder, "junit.xml"))
	if err != nil {
		panic(fmt.Errorf("Error creating junit file: %v", err))
	}

	godogOpts.Output = io.MultiWriter(godogOpts.Output, mainLogFile)
}

func initializeTestSuite(ctx *godog.TestSuiteContext) {
	// Verify Setup
	if err := framework.CheckSetup(); err != nil {
		panic(err)
	}

	// Initialization of cluster wide resources
	ctx.BeforeSuite(func() {
		if config.IsOperatorProfiling() {
			framework.GetMainLogger().Info("Running testing with profiling")
		}

		monitorOlmNamespace()

		if config.IsOperatorInstalledByOlm() {
			if err := installKogitoOperatorCatalogSource(); err != nil {
				panic(err)
			}
		}
	})

	// Final cleanup once test suite finishes
	ctx.AfterSuite(func() {
		if config.IsOperatorProfiling() {
			retrieveProfilingData()
		}

		if !config.IsKeepNamespace() {
			// Delete all operators created by test suite
			if success := installers.UninstallServicesFromCluster(); !success {
				framework.GetMainLogger().Warn("Some services weren't uninstalled propertly from cluster, see error logs above")
			}
		}

		if config.IsOperatorInstalledByOlm() {
			deleteKogitoOperatorCatalogSource()
		}

		stopOlmNamespaceMonitoring()
	})
}

func initializeScenario(ctx *godog.ScenarioContext) {
	// Register Steps
	data := &steps.Data{}

	if PreRegisterStepsHook != nil {
		PreRegisterStepsHook(ctx, data)
	}

	data.RegisterAllSteps(ctx)

	// Scenario handlers
	ctx.BeforeScenario(func(scenario *godog.Scenario) {
		if err := data.BeforeScenario(scenario); err != nil {
			framework.GetLogger(data.Namespace).Error(err, "Error in configuring data for before scenario")
		}
	})
	ctx.AfterScenario(func(scenario *godog.Scenario, err error) {

		if AfterScenarioHook != nil {
			if err := AfterScenarioHook(scenario, data); err != nil {
				framework.GetLogger(data.Namespace).Error(err, "Error in executing AfterScenarioHook")
			}
		}

		if err := data.AfterScenario(scenario, err); err != nil {
			framework.GetLogger(data.Namespace).Error(err, "Error in configuring data for After scenario")
		}

		// Namespace should be deleted after all other operations have been done
		if !config.IsKeepNamespace() {
			if success := installers.UninstallServicesFromNamespace(data.Namespace); !success {
				framework.GetMainLogger().Warn("Some services weren't uninstalled propertly from namespace, see error logs above", "namespace", data.Namespace)
			}

			deleteNamespaceIfExists(data.Namespace)
		}
	})

	// Step handlers
	ctx.BeforeStep(func(s *godog.Step) {
		framework.GetLogger(data.Namespace).Info("Step", "stepText", s.Text)
	})
	ctx.AfterStep(func(s *godog.Step, err error) {
		if err != nil {
			framework.GetLogger(data.Namespace).Error(err, "Error in step", "step", s.Text)
		}
	})
}

func deleteNamespaceIfExists(namespace string) {
	err := framework.OperateOnNamespaceIfExists(namespace, func(namespace string) error {
		framework.GetLogger(namespace).Info("Delete created namespace", "namespace", namespace)
		if e := framework.DeleteNamespace(namespace); e != nil {
			return fmt.Errorf("Error while deleting the namespace: %v", e)
		}
		return nil
	})
	if err != nil {
		framework.GetLogger(namespace).Error(err, "Error while doing operator on namespace")
	}
}

func showScenarios(features []*gherkin.Feature, showSteps bool) {
	mainLogger := framework.GetMainLogger()
	mainLogger.Info("------------------ SHOW SCENARIOS ------------------")
	for _, ft := range features {
		// Placeholders in names are now replaced directly into names for each scenario
		if len(ft.Scenarios) > 0 {
			mainLogger.Info(fmt.Sprintf("Feature: %s", ft.Document.GetFeature().GetName()))
			for _, scenario := range ft.Scenarios {
				mainLogger.Info(fmt.Sprintf("    Scenario: %s", scenario.GetName()))
				if showSteps {
					for _, step := range scenario.Steps {
						mainLogger.Info(fmt.Sprintf("        Step: %s", step.GetText()))
					}
				}
			}
		}
	}
	mainLogger.Info("------------------ END SHOW SCENARIOS ------------------")
}

func monitorOlmNamespace() {
	monitorNamespace(config.GetOlmNamespace())
}

func monitorNamespace(namespace string) {
	go func() {
		err := framework.StartPodLogCollector(namespace)
		if err != nil {
			framework.GetLogger(namespace).Error(err, "Error starting log collector", "namespace", namespace)
		}
	}()
}

func stopOlmNamespaceMonitoring() {
	stopNamespaceMonitoring(config.GetOlmNamespace())
}

func stopNamespaceMonitoring(namespace string) {
	if err := framework.StopPodLogCollector(namespace); err != nil {
		framework.GetMainLogger().Error(err, "Error stopping log collector", "namespace", namespace)
	}
	if err := framework.BumpEvents(namespace); err != nil {
		framework.GetMainLogger().Error(err, "Error bumping events", "namespace", namespace)
	}
}

// Install cluster wide Kogito operator from OLM
func installKogitoOperatorCatalogSource() error {
	// Create CatalogSource
	if _, err := framework.CreateKogitoOperatorCatalogSource(); err != nil {
		return fmt.Errorf("Error installing custer wide Kogito operator using OLM: %v", err)
	}

	// Wait for the CatalogSource
	if err := framework.WaitForKogitoOperatorCatalogSourceReady(); err != nil {
		return fmt.Errorf("Error while waiting for Kogito operator CatalogSource initialization: %v", err)
	}

	return nil
}

func deleteKogitoOperatorCatalogSource() {
	if err := framework.DeleteKogitoOperatorCatalogSource(); err != nil {
		framework.GetMainLogger().Error(err, "Error deleting Kogito operator CatalogSource")
	}
}

func retrieveProfilingData() {
	framework.GetMainLogger().Info("Retrieve Profiling Data")

	if err := framework.RemoveKogitoOperatorDeployment(installers.KogitoNamespace); err != nil {
		framework.GetMainLogger().Error(err, "Unable to delete Kogito Operator Deployment")
		return
	}

	// Apply dataaccess
	if _, err := framework.CreateCommand("oc", "apply", "-f", config.GetOperatorProfilingDataAccessYamlURI()).Execute(); err != nil {
		framework.GetMainLogger().Error(err, "Error while installing Kogito operator from YAML file")
		return
	}

	// Wait for dataaccess pod
	if err := framework.WaitForPodsWithLabel(installers.KogitoNamespace, "name", "profiling-data-access", 1, 2); err != nil {
		framework.GetMainLogger().Error(err, "Error while waiting for profiling data access pod")
		return
	}

	// Copy coverage data
	dataFileInContainer := fmt.Sprintf("%s:/data/cover.out", "kogito-operator-profiling-data-access")
	if _, err := framework.CreateCommand("oc", "cp", dataFileInContainer, config.GetOperatorProfilingOutputFileURI(), "-n", installers.KogitoNamespace).Execute(); err != nil {
		framework.GetMainLogger().Error(err, "Error while installing Kogito operator from YAML file")
		return
	}
}
