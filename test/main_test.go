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
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/test/config"
	"github.com/kiegroup/kogito-cloud-operator/test/framework"
	"github.com/kiegroup/kogito-cloud-operator/test/steps"
)

const (
	disabledTag    = "@disabled"
	cliTag         = "@cli"
	smokeTag       = "@smoke"
	performanceTag = "@performance"

	kogitoClusterWideNamespace = "kogito-operator-system"
)

var (
	godogOpts = godog.Options{
		Output:    colors.Colored(os.Stdout),
		Format:    "junit",
		Randomize: time.Now().UTC().UnixNano(),
		Tags:      disabledTag,
	}

	// Array of created cluster objects, used to cleanup environment when the tests finish
	clusterObjects []meta.ResourceObject
)

func init() {
	godog.BindFlags("godog.", flag.CommandLine, &godogOpts)
	config.BindFlags(flag.CommandLine)
}

func TestMain(m *testing.M) {
	flag.Parse()
	godogOpts.Paths = flag.Args()

	configureTags()
	configureTestOutput()

	features, err := parseFeatures(godogOpts.Tags, godogOpts.Paths)
	if err != nil {
		panic(fmt.Errorf("Error parsing features: %v", err))
	}
	if config.IsShowScenarios() || config.IsShowSteps() {
		showScenarios(features, config.IsShowSteps())
	}

	if !config.IsDryRun() {
		if !config.IsCrDeploymentOnly() || matchingFeature(cliTag, features) {
			// Check CLI binary is existing if needed
			if exits, err := framework.CheckCliBinaryExist(); err != nil {
				panic(fmt.Errorf("Error trying to get CLI binary %v", err))
			} else if !exits {
				panic("CLI Binary does not exist on specified path")
			}
		}

		status := godog.TestSuite{
			Name:                 "godogs",
			TestSuiteInitializer: InitializeTestSuite,
			ScenarioInitializer:  InitializeScenario,
			Options:              &godogOpts,
		}.Run()

		if st := m.Run(); st > status {
			status = st
		}
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
	if config.IsSmokeTests() {
		framework.SetLogSubFolder("smoke")
	} else {
		framework.SetLogSubFolder("full")
	}
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

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	// Create kube client
	if err := framework.InitKubeClient(); err != nil {
		panic(err)
	}

	// Verify Setup
	if err := framework.CheckSetup(); err != nil {
		panic(err)
	}

	// Initialization of cluster wide resources
	ctx.BeforeSuite(func() {
		monitorOlmNamespace()

		if !config.IsOperatorNamespaced() {
			if err := installClusterWideKogitoOperator(); err != nil {
				panic(err)
			}
			monitorKogitoOperatorNamespace()
		}
	})

	// Final cleanup once test suite finishes
	ctx.AfterSuite(func() {
		if !config.IsKeepNamespace() {
			// Delete all operators created by test suite
			deleteClusterWideTestOperators()
		}

		if !config.IsOperatorNamespaced() {
			if err := uninstallClusterWideKogitoOperator(); err != nil {
				panic(err)
			}
			stopKogitoOperatorNamespaceMonitoring()
		}

		stopOlmNamespaceMonitoring()
	})
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	// Register Steps
	data := &steps.Data{}
	data.RegisterAllSteps(ctx)

	// Scenario handlers
	ctx.BeforeScenario(func(scenario *godog.Scenario) {
		if err := data.BeforeScenario(scenario); err != nil {
			framework.GetLogger(data.Namespace).Error(err, "Error in configuring data for before scenario")
		}
	})
	ctx.AfterScenario(func(scenario *godog.Scenario, err error) {
		if err := data.AfterScenario(scenario, err); err != nil {
			framework.GetLogger(data.Namespace).Error(err, "Error in configuring data for After scenario")
		}

		// Namespace should be deleted after all other operations have been done
		if !config.IsKeepNamespace() {
			// Delete all objects created for the scenario
			for _, o := range data.GetCreatedOperatorObjects() {
				if err := framework.DeleteObject(o); err != nil {
					framework.GetMainLogger().Error(err, "Error removing created objects", "namespace", data.Namespace)
				}
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

func matchingFeature(filterTags string, features []*feature) bool {
	for _, ft := range features {
		if matchesScenarios(filterTags, ft.scenarios) {
			return true
		}
	}
	return false
}

func matchesScenarios(filterTags string, scenarios []*godog.Scenario) bool {
	for _, scenario := range scenarios {
		if matchesTags(filterTags, scenario.Tags) {
			return true
		}
	}
	return false
}

func showScenarios(features []*feature, showSteps bool) {
	mainLogger := framework.GetMainLogger()
	mainLogger.Info("------------------ SHOW SCENARIOS ------------------")
	for _, ft := range features {
		// Placeholders in names are now replaced directly into names for each scenario
		if len(ft.scenarios) > 0 {
			mainLogger.Info(fmt.Sprintf("Feature: %s", ft.document.GetFeature().GetName()))
			for _, scenario := range ft.scenarios {
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

func deleteClusterWideTestOperators() {
	subscriptions, err := framework.GetClusterWideTestSubscriptions()
	if err != nil {
		framework.GetMainLogger().Error(err, "Error retrieving cluster wide test subscriptions")
		return
	}

	for _, subscription := range subscriptions.Items {
		err := framework.DeleteSubscription(&subscription)
		if err != nil {
			framework.GetMainLogger().Error(err, "Error deleting cluster wide test subscription", "subscriptionName", subscription.Name)
		}
	}
}

func monitorOlmNamespace() {
	monitorNamespace(config.GetOlmNamespace())
}

func monitorKogitoOperatorNamespace() {
	monitorNamespace(kogitoClusterWideNamespace)
}

func monitorNamespace(namespace string) {
	go func() {
		err := framework.StartPodLogCollector(namespace)
		if err != nil {
			framework.GetLogger(namespace).Error(err, "Error starting log collector", "namespace", namespace, err)
		}
	}()
}

func stopOlmNamespaceMonitoring() {
	stopNamespaceMonitoring(config.GetOlmNamespace())
}

func stopKogitoOperatorNamespaceMonitoring() {
	stopNamespaceMonitoring(kogitoClusterWideNamespace)
}

func stopNamespaceMonitoring(namespace string) {
	if err := framework.StopPodLogCollector(namespace); err != nil {
		framework.GetMainLogger().Error(err, "Error stopping log collector", "namespace", namespace)
	}
	if err := framework.BumpEvents(namespace); err != nil {
		framework.GetMainLogger().Error(err, "Error bumping events", "namespace", namespace)
	}
}

// Install cluster wide Kogito operator
func installClusterWideKogitoOperator() error {
	// Check that Kogito operator is not deployed yet
	if running, err := framework.KogitoOperatorExists(kogitoClusterWideNamespace); err != nil {
		return fmt.Errorf("Error while checking whether Kogito operator in namespace %s is running: %v", kogitoClusterWideNamespace, err)
	} else if running {
		return fmt.Errorf("Kogito operator is already running in namespace %s. Please uninstall the operator first", kogitoClusterWideNamespace)
	}

	// Create the Kogito operator namespace if it doesn't exist
	if exists, err := framework.IsNamespace(kogitoClusterWideNamespace); err != nil {
		return err
	} else if !exists {
		// Create the namespace
		if err := framework.CreateNamespace(kogitoClusterWideNamespace); err != nil {
			return err
		}
	}

	// Deploy the operator
	created, err := framework.DeployClusterWideKogitoOperatorFromYaml(kogitoClusterWideNamespace)
	clusterObjects = created
	if err != nil {
		return err
	}

	// Wait until operator runs
	if err := framework.WaitForKogitoOperatorRunning(kogitoClusterWideNamespace); err != nil {
		return fmt.Errorf("Error while checking operator running: %v", err)
	}

	return nil
}

// Uninstall cluster wide Kogito operator
func uninstallClusterWideKogitoOperator() error {
	if !config.IsKeepNamespace() {
		// Delete all objects created for the Kogito operator
		for _, o := range clusterObjects {
			if err := framework.DeleteObject(o); err != nil {
				framework.GetMainLogger().Error(err, "Error removing created objects", "namespace", kogitoClusterWideNamespace)
			}
		}

		return framework.DeleteNamespace(kogitoClusterWideNamespace)
	}
	return nil
}
