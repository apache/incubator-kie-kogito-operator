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

package config

import (
	"flag"
	"path/filepath"

	"github.com/kiegroup/kogito-cloud-operator/pkg/version"
)

// TestConfig contains the information about the tests environment
type TestConfig struct {
	// tests configuration
	smoke            bool
	performance      bool
	loadFactor       int
	localTests       bool
	ciName           string
	crDeploymentOnly bool
	containerEngine  string
	domainSuffix     string
	imageCacheMode   string
	httpRetryNumber  int
	olmNamespace     string

	// operator information
	operatorImageName  string
	operatorImageTag   string
	operatorNamespaced bool

	// files/binaries
	operatorYamlURI string
	cliPath         string

	// runtime
	servicesImageRegistry             string
	servicesImageNamespace            string
	servicesImageNameSuffix           string
	servicesImageVersion              string
	dataIndexImageTag                 string
	explainabilityImageTag            string
	jobsServiceImageTag               string
	mgmtConsoleImageTag               string
	taskConsoleImageTag               string
	trustyImageTag                    string
	trustyUIImageTag                  string
	runtimeApplicationImageRegistry   string
	runtimeApplicationImageNamespace  string
	runtimeApplicationImageNameSuffix string
	runtimeApplicationImageVersion    string

	// build
	customMavenRepoURL                 string
	customMavenRepoReplaceDefault      bool
	mavenMirrorURL                     string
	mavenIgnoreSelfSignedCertificate   bool
	buildImageRegistry                 string
	buildImageNamespace                string
	buildImageNameSuffix               string
	buildImageVersion                  string
	buildS2iImageTag                   string
	buildRuntimeImageTag               string
	disableMavenNativeBuildInContainer bool

	// examples repository
	examplesRepositoryURI string
	examplesRepositoryRef string

	// dev options
	showScenarios bool
	showSteps     bool
	dryRun        bool
	keepNamespace bool
	namespaceName string
	localCluster  bool
}

const (
	defaultOperatorImageName = "quay.io/kiegroup/kogito-cloud-operator"

	defaultOperatorYamlURI = "../kogito-operator.yaml"
	defaultCliPath         = "../build/_output/bin/kogito"

	defaultKogitoExamplesURI = "https://github.com/kiegroup/kogito-examples"

	defaultLoadFactor      = 1
	defaultHTTPRetryNumber = 3

	defaultContainerEngine = "podman"
)

var (
	defaultOperatorImageTag = version.Version

	env = TestConfig{}
)

// BindFlags binds BDD tests env flags to given flag set
func BindFlags(set *flag.FlagSet) {
	prefix := "tests."
	developmentOptionsPrefix := prefix + "dev."

	// tests configuration
	set.BoolVar(&env.smoke, prefix+"smoke", false, "Launch only smoke tests")
	set.BoolVar(&env.performance, prefix+"performance", false, "Launch performance tests")
	set.IntVar(&env.loadFactor, prefix+"load-factor", defaultLoadFactor, "Set the tests load factor. Useful for the tests to take into account that the cluster can be overloaded, for example for the calculation of timeouts. Default value is 1.")
	set.BoolVar(&env.localTests, prefix+"local", false, "If tests are launch on local machine using either a local or remote cluster")
	set.StringVar(&env.ciName, prefix+"ci", "", "If tests are launch on ci machine, give the CI name")
	set.BoolVar(&env.crDeploymentOnly, prefix+"cr-deployment-only", false, "Use this option if you have no CLI to test against. It will use only direct CR deployments.")
	set.StringVar(&env.containerEngine, prefix+"container-engine", defaultContainerEngine, "Engine used to interact with images and local containers.")
	set.StringVar(&env.domainSuffix, prefix+"domain-suffix", "", "Set the domain suffix for exposed services. Ignored when running tests on Openshift.")
	set.StringVar(&env.imageCacheMode, prefix+"image-cache-mode", "if-available", "Use this option to specify whether you want to use image cache for runtime images. Available options are 'always', 'never' or 'if-available'(default).")
	set.IntVar(&env.httpRetryNumber, prefix+"http-retry-nb", defaultHTTPRetryNumber, "Set the retry number for all HTTP calls in case it fails (and response code != 500). Default value is 3.")
	set.StringVar(&env.olmNamespace, prefix+"olm-namespace", "openshift-operators", "Set the namespace which is used for cluster scope operators. Default is 'openshift-operators'.")

	// operator information
	set.StringVar(&env.operatorImageName, prefix+"operator-image-name", defaultOperatorImageName, "Operator image name")
	set.StringVar(&env.operatorImageTag, prefix+"operator-image-tag", defaultOperatorImageTag, "Operator image tag")
	set.BoolVar(&env.operatorNamespaced, prefix+"operator-namespaced", false, "Set to true to deploy Kogito operator into namespace used for scenario execution, false for cluster wide deployment. Default is false.")

	// files/binaries
	set.StringVar(&env.operatorYamlURI, prefix+"operator-yaml-uri", defaultOperatorYamlURI, "Url or Path to kogito-operator.yaml file")
	set.StringVar(&env.cliPath, prefix+"cli-path", defaultCliPath, "Path to built CLI to test")

	// runtime
	set.StringVar(&env.servicesImageRegistry, prefix+"services-image-registry", "", "Set the services (jobs-service, data-index, trusty, explainability) image registry")
	set.StringVar(&env.servicesImageNamespace, prefix+"services-image-namespace", "", "Set the services (jobs-service, data-index, trusty, explainability) image namespace")
	set.StringVar(&env.servicesImageNameSuffix, prefix+"services-image-name-suffix", "", "Set the services (jobs-service, data-index, trusty, explainability) image name suffix")
	set.StringVar(&env.servicesImageVersion, prefix+"services-image-version", "", "Set the services (jobs-service, data-index, trusty, explainability) image version")
	set.StringVar(&env.dataIndexImageTag, prefix+"data-index-image-tag", "", "Set the Kogito Data Index image tag ('services-image-version' is ignored)")
	set.StringVar(&env.explainabilityImageTag, prefix+"explainability-image-tag", "", "Set the Kogito Explainability image tag ('services-image-version' is ignored)")
	set.StringVar(&env.jobsServiceImageTag, prefix+"jobs-service-image-tag", "", "Set the Kogito Jobs Service image tag ('services-image-version' is ignored)")
	set.StringVar(&env.mgmtConsoleImageTag, prefix+"management-console-image-tag", "", "Set the Kogito Management Console image tag ('services-image-version' is ignored)")
	set.StringVar(&env.taskConsoleImageTag, prefix+"task-console-image-tag", "", "Set the Kogito Task Console image tag ('services-image-version' is ignored)")
	set.StringVar(&env.trustyImageTag, prefix+"trusty-image-tag", "", "Set the Kogito Trusty image tag ('services-image-version' is ignored)")
	set.StringVar(&env.trustyUIImageTag, prefix+"trusty-ui-image-tag", "", "Set the Kogito Trusty UI image tag ('services-image-version' is ignored)")
	set.StringVar(&env.runtimeApplicationImageRegistry, prefix+"runtime-application-image-registry", "", "Set the runtime application (built Kogito application image) image registry")
	set.StringVar(&env.runtimeApplicationImageNamespace, prefix+"runtime-application-image-namespace", "", "Set the runtime application (built Kogito application image) image namespace")
	set.StringVar(&env.runtimeApplicationImageNameSuffix, prefix+"runtime-application-image-name-suffix", "", "Set the runtime application (built Kogito application image) image name suffix")
	set.StringVar(&env.runtimeApplicationImageVersion, prefix+"runtime-application-image-version", "", "Set the runtime application (built Kogito application image) image version")

	// build
	set.StringVar(&env.customMavenRepoURL, prefix+"custom-maven-repo-url", "", "Set a custom Maven repository url for S2I builds, in case your artifacts are in a specific repository. See https://github.com/kiegroup/kogito-images/README.md for more information")
	set.BoolVar(&env.customMavenRepoReplaceDefault, prefix+"custom-maven-repo-replace-default", false, "If you specified the option 'tests.custom-maven-repo-url' and you want that one to replace the main JBoss repository (useful with snapshots).")
	set.StringVar(&env.mavenMirrorURL, prefix+"maven-mirror-url", "", "Maven mirror url to be used when building app in the tests")
	set.BoolVar(&env.mavenIgnoreSelfSignedCertificate, prefix+"maven-ignore-self-signed-certificate", false, "Set to true if maven build need to ignore self-signed certificate. This could happen when using internal maven mirror url.")
	set.StringVar(&env.buildImageRegistry, prefix+"build-image-registry", "", "Set the build image registry")
	set.StringVar(&env.buildImageNamespace, prefix+"build-image-namespace", "", "Set the build image namespace")
	set.StringVar(&env.buildImageNameSuffix, prefix+"build-image-name-suffix", "", "Set the build image name suffix")
	set.StringVar(&env.buildImageVersion, prefix+"build-image-version", "", "Set the build image version")
	set.StringVar(&env.buildS2iImageTag, prefix+"build-s2i-image-tag", "", "Set the S2I build image full tag")
	set.StringVar(&env.buildRuntimeImageTag, prefix+"build-runtime-image-tag", "", "Set the Runtime build image full tag")
	set.BoolVar(&env.disableMavenNativeBuildInContainer, prefix+"disable-maven-native-build-container", false, "By default, Maven native builds are done in container (via container engine). Possibility to disable it.")

	// examples repository
	set.StringVar(&env.examplesRepositoryURI, prefix+"examples-uri", defaultKogitoExamplesURI, "Set the URI for the kogito-examples repository")
	set.StringVar(&env.examplesRepositoryRef, prefix+"examples-ref", "", "Set the branch for the kogito-examples repository")

	// dev options
	set.BoolVar(&env.showScenarios, prefix+"show-scenarios", false, "Show all scenarios which will be executed.")
	set.BoolVar(&env.showSteps, prefix+"show-steps", false, "Show all scenarios and their steps which will be executed.")
	set.BoolVar(&env.dryRun, prefix+"dry-run", false, "Dry Run the tests.")
	set.BoolVar(&env.keepNamespace, prefix+"keep-namespace", false, "Do not delete namespace(s) after scenario run (WARNING: can be resources consuming ...)")
	set.StringVar(&env.namespaceName, developmentOptionsPrefix+"namespace-name", "", "Use the specified namespace for scenarios, don't generate random namespace.")
	set.BoolVar(&env.localCluster, developmentOptionsPrefix+"local-cluster", false, "If tests are launch using a local cluster")
}

// tests configuration

// IsSmokeTests return whether smoke tests should be executed
func IsSmokeTests() bool {
	return env.smoke
}

// IsPerformanceTests return whether performance tests should be executed
func IsPerformanceTests() bool {
	return env.performance
}

// GetLoadFactor return the load factor of the cluster
func GetLoadFactor() int {
	return env.loadFactor
}

// IsLocalTests return whether tests are executed in local
func IsLocalTests() bool {
	return env.localTests
}

// GetCiName return the CI name that executes the tests, if any
func GetCiName() string {
	return env.ciName
}

// IsCrDeploymentOnly returns whether the deployment should be done only with CR
func IsCrDeploymentOnly() bool {
	return env.crDeploymentOnly
}

// GetContainerEngine returns engine used to interact with images and local containers
func GetContainerEngine() string {
	return env.containerEngine
}

// GetDomainSuffix returns the domain suffix for exposed services
func GetDomainSuffix() string {
	return env.domainSuffix
}

// GetImageCacheMode returns image cache mode
func GetImageCacheMode() ImageCacheMode {
	return ImageCacheMode(env.imageCacheMode)
}

// GetHTTPRetryNumber return the number of retries to be applied for http calls
func GetHTTPRetryNumber() int {
	return env.httpRetryNumber
}

// GetOlmNamespace returns namespace which is used for cluster scope operators
func GetOlmNamespace() string {
	return env.olmNamespace
}

// operator information

// GetOperatorImageName return the image name for the operator
func GetOperatorImageName() string {
	return env.operatorImageName
}

// GetOperatorImageTag return the image tag for the operator
func GetOperatorImageTag() string {
	return env.operatorImageTag
}

// IsOperatorNamespaced return true if the Kogito operator should be deployed in scenario namespace, false for cluster wide deployment
func IsOperatorNamespaced() bool {
	return env.operatorNamespaced
}

// files/binaries

// GetOperatorYamlURI return the uri for kogito-operator.yaml file
func GetOperatorYamlURI() string {
	return env.operatorYamlURI
}

// GetOperatorCliPath return the path to the kogito CLI binary
func GetOperatorCliPath() (string, error) {
	return filepath.Abs(env.cliPath)
}

// runtime

// GetServicesImageRegistry return the registry for the services images
func GetServicesImageRegistry() string {
	return env.servicesImageRegistry
}

// GetServicesImageNamespace return the namespace for the services images
func GetServicesImageNamespace() string {
	return env.servicesImageNamespace
}

// GetServicesImageNameSuffix return the name suffix for the services images
func GetServicesImageNameSuffix() string {
	return env.servicesImageNameSuffix
}

// GetServicesImageVersion return the version for the services images
func GetServicesImageVersion() string {
	return env.servicesImageVersion
}

// GetDataIndexImageTag return the Kogito Data Index image tag
func GetDataIndexImageTag() string {
	return env.dataIndexImageTag
}

// GetExplainabilityImageTag return the Kogito Explainability image tag
func GetExplainabilityImageTag() string {
	return env.explainabilityImageTag
}

// GetJobsServiceImageTag return the Kogito Jobs Service image tag
func GetJobsServiceImageTag() string {
	return env.jobsServiceImageTag
}

// GetManagementConsoleImageTag return the Kogito Management Console image tag
func GetManagementConsoleImageTag() string {
	return env.mgmtConsoleImageTag
}

// GetTaskConsoleImageTag return the Kogito Management Console image tag
func GetTaskConsoleImageTag() string {
	return env.taskConsoleImageTag
}

// GetTrustyImageTag return the Kogito Trusty image tag
func GetTrustyImageTag() string {
	return env.trustyImageTag
}

// GetTrustyUIImageTag return the Kogito Management Console image tag
func GetTrustyUIImageTag() string {
	return env.trustyUIImageTag
}

// GetRuntimeApplicationImageRegistry return the registry for the runtime application images
func GetRuntimeApplicationImageRegistry() string {
	return env.runtimeApplicationImageRegistry
}

// GetRuntimeApplicationImageNamespace return the namespace for runtime application images
func GetRuntimeApplicationImageNamespace() string {
	return env.runtimeApplicationImageNamespace
}

// GetRuntimeApplicationImageNameSuffix return the name suffix for runtime application images
func GetRuntimeApplicationImageNameSuffix() string {
	return env.runtimeApplicationImageNameSuffix
}

// GetRuntimeApplicationImageVersion return the version for runtime application images
func GetRuntimeApplicationImageVersion() string {
	return env.runtimeApplicationImageVersion
}

// build

// GetCustomMavenRepoURL return the custom maven repository url used by S2I builds
func GetCustomMavenRepoURL() string {
	return env.customMavenRepoURL
}

// IsCustomMavenRepoReplaceDefault return whether custom maven repo should replace the default JBoss repository
func IsCustomMavenRepoReplaceDefault() bool {
	return env.customMavenRepoReplaceDefault
}

// GetMavenMirrorURL return the maven mirror url used for building applications
func GetMavenMirrorURL() string {
	return env.mavenMirrorURL
}

// IsMavenIgnoreSelfSignedCertificate return whether self-signed certficate should be ignored
func IsMavenIgnoreSelfSignedCertificate() bool {
	return env.mavenIgnoreSelfSignedCertificate
}

// GetBuildImageRegistry return the registry for the build images
func GetBuildImageRegistry() string {
	return env.buildImageRegistry
}

// GetBuildImageNamespace return the namespace for the build images
func GetBuildImageNamespace() string {
	return env.buildImageNamespace
}

// GetBuildImageNameSuffix return the namespace for the build images
func GetBuildImageNameSuffix() string {
	return env.buildImageNameSuffix
}

// GetBuildImageVersion return the version for the build images
func GetBuildImageVersion() string {
	return env.buildImageVersion
}

// GetBuildS2IImageStreamTag return the tag for the s2i build image
func GetBuildS2IImageStreamTag() string {
	return env.buildS2iImageTag
}

// GetBuildRuntimeImageStreamTag return the tag for the runtime build image
func GetBuildRuntimeImageStreamTag() string {
	return env.buildRuntimeImageTag
}

// IsDisableMavenNativeBuildInContainer return whether Maven native build in container should be disabled
func IsDisableMavenNativeBuildInContainer() bool {
	return env.disableMavenNativeBuildInContainer
}

// examples repository

// GetExamplesRepositoryURI return the uri for the examples repository
func GetExamplesRepositoryURI() string {
	return env.examplesRepositoryURI
}

// GetExamplesRepositoryRef return the branch for the examples repository
func GetExamplesRepositoryRef() string {
	return env.examplesRepositoryRef
}

// dev options

// IsShowScenarios return whether we should display scenarios
func IsShowScenarios() bool {
	return env.showScenarios
}

// IsShowSteps return whether we should display scenarios's steps
func IsShowSteps() bool {
	return env.showSteps
}

// IsDryRun return whether we should do a dry run
func IsDryRun() bool {
	return env.dryRun
}

// IsKeepNamespace return whether we should keep namespace after scenario run
func IsKeepNamespace() bool {
	return env.keepNamespace
}

// GetNamespaceName return namespace name if it was defined
func GetNamespaceName() string {
	return env.namespaceName
}

// IsLocalCluster return whether tests are executed using a local cluster
func IsLocalCluster() bool {
	return env.localCluster
}
