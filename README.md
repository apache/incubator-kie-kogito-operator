# Kogito Operator

[![Go Report Card](https://goreportcard.com/badge/github.com/kiegroup/kogito-operator)](https://goreportcard.com/report/github.com/kiegroup/kogito-operator)

The Kogito Operator deploys [Kogito Runtimes](https://github.com/kiegroup/kogito-runtimes) services from source and all infrastructure requirements for the services, such as persistence with [Infinispan](https://infinispan.org/) and messaging with [Apache Kafka](https://kafka.apache.org/). Kogito provides a command-line interface (CLI) that enables you to interact with the Kogito Operator for deployment tasks.

For information about the Kogito Operator architecture and instructions for using the operator and CLI to deploy Kogito services and infrastructures, see the official [Kogito Documentation](https://docs.jboss.org/kogito/release/latest/html_single/#chap-kogito-deploying-on-openshift) page.

Table of Contents
=================

* [Trying the Kogito Operator](#trying-the-kogito-operator)
* [Using Kogito Custom Resources as a dependency](#using-kogito-custom-resources-as-a-dependency)
* [Contributing to the Kogito Operator](#contributing-to-the-kogito-operator)
  * [Kogito Operator environment](#kogito-operator-environment)
    * [Kogito Operator unit tests](#kogito-operator-unit-tests)
    * [Kogito Operator collaboration and pull requests](#kogito-operator-collaboration-and-pull-requests)
  * [Kogito Operator development](#kogito-operator-development)
    * [Prerequisites](#prerequisites)
    * [Requirements](#requirements)
    * [Building the Kogito Operator](#building-the-kogito-operator)
    * [Deploying to OpenShift 4.x for development purposes](#deploying-to-openshift-4x-for-development-purposes)
    * [Running BDD Tests](#running-bdd-tests)
      * [Running BDD tests with current branch](#running-bdd-tests-with-current-branch)
      * [Running BDD tests with custom Kogito Build images' version](#running-bdd-tests-with-custom-kogito-build-images-version)
      * [Running smoke tests](#running-smoke-tests)
      * [Running performance tests](#running-performance-tests)
      * [List of test tags](#list-of-test-tags)
    * [Running the Kogito Operator locally](#running-the-kogito-operator-locally)
    * [Remote Debug Kogito Operator using Intellij IDEA](#remote-debug-kogito-operator-using-intellij-idea)
  * [Guide for Kogito Developers](#guide-for-kogito-developers)
  * [Releasing Kogito Operator](#releasing-kogito-operator)
  * [Kogito Operator tested integrations](#kogito-operator-tested-integrations)

Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc)

# Trying the Kogito Operator

You can quickly install the operator in your local cluster by executing the following command:

```shell script
VERSION=<current_operator_version>
kubectl apply -f "https://github.com/kiegroup/kogito-operator/releases/download/${VERSION}/kogito-operator.yaml"
```

Replace the version above to match your context (example v1.8.0). The version information you can grab from the [releases page](https://github.com/kiegroup/kogito-operator/releases).

Alternatively, if you cloned this repo just do:

```shell script
$ ./hack/install.sh
```

The script will download the latest version and install the resources for you in the `kogito-operator-system` namespace.
You can set the `VERSION` variable before running the script to control which version to install.

For further information on how to install in other environments and configure
the Kogito Operator, [please see our official documentation](https://docs.jboss.org/kogito/release/latest/html_single/#chap-kogito-deploying-on-openshift).

# Using Kogito Custom Resources as a dependency

It's possible to use the Kubernetes Custom Resources managed by Kogito Operator
in your Golang project. For this to  work, you just have to import our `client` module in your `go.mod` file:

```shell script
# client depends on the `api` module
go install github.com/kiegroup/kogito-operator/client@{VERSION}
```

Replace `{VERSION}` with the [latest release](https://github.com/kiegroup/kogito-operator/releases).

The `client` module will provide your project with the Kubernetes client 
for Kogito custom resources such as `KogitoRuntime` and `KogitoBuild`.

Avoid importing the whole project unless extremely necessary. The Kogito Operator
has many dependencies since it can connect to different managed resources
from third-party vendors such as Strimzi and Infinispan.

# Contributing to the Kogito Operator

Thank you for your interest in contributing to this project!

Any kind of contribution is welcome: code, design ideas, bug reporting, or documentation (including this page).

## Kogito Operator environment

The Operator SDK is updated regularly, and the Kogito Operator code typically uses the most recent SDK updates as soon as possible.

If you do not have a preferred IDE, use Visual Studio Code with the [`vscode-go`](https://github.com/Microsoft/vscode-go) plugin for Go language tools support.

To use Go modules with VS Code, see [Go modules support in VS Code](https://github.com/golang/vscode-go#support-for-go-modules).

To debug Go in your VS code, see [Debugging Go code using VS Code](https://github.com/golang/vscode-go/blob/master/docs/debugging.md).

We check our code with `golangci-lint`, so it is recommended to add this to your IDE.
For adding the `golangci-lint` with goland, see [Go Linter](https://plugins.jetbrains.com/plugin/12496-go-linter).

For adding the `golangci-lint` with VScode, install the [Go Plugin](https://marketplace.visualstudio.com/items?itemName=ms-vscode.Go) and enable the linter from the plugins setting.

### Kogito Operator unit tests

For information about Operator SDK testing, see [Unit testing with the Operator SDK](https://sdk.operatorframework.io/docs/golang/legacy/unit-testing/).

In general, the unit tests that are provided with the Kogito Operator are based on that Operator SDK testing resource. You might encounter minor issues as you create specific OpenShift APIs such as `BuildConfig` and `DeploymentConfig` that are not listed there. For an example test case with sample API calls, see the [`kogitobuild_controller_test.go`](../main/controllers/kogitobuild_controller_test.go) test file.

### Kogito Operator collaboration and pull requests

Before you start to work on a new proposed feature or on a fix for a bug, [open an issue](https://github.com/kiegroup/kogito-operator/issues) to discuss your idea or bug report with the maintainers. You can also work on a [JIRA issue](https://issues.jboss.org/issues/?jql=project+%3D+KOGITO+AND+component+%3D+Operator) that has been reported. A developer might already be assigned to address the issue, but you can leave a comment in the JIRA asking if they need some help.

After you update the source with your new proposed feature or bug fix, open a [pull request (PR)](https://help.github.com/en/articles/about-pull-requests) that meets the following requirements:

- You have a JIRA associated with the PR. 
- Your PR has the name of the JIRA in the title, for example, `[KOGITO-XXX] - Awesome feature that solves it all`.
- The PR solves only the problem described in the JIRA.
- You have written unit tests for the particular fix or feature.
- You ran `make before-pr` before submitting the PR and everything is working accordingly.
- You tested the feature on an actual OpenShift cluster.
- You've added a [RELEASE_NOTES.md](RELEASE_NOTES.md) entry regarding this change.

After you send your PR, a maintainer will review your code and might ask you to make changes and to [squash your commits](https://stackoverflow.com/questions/5189560/squash-my-last-x-commits-together-using-git) before we can merge.

If you have any questions, contact a Kogito Operator maintainer in the [issues page](https://github.com/kiegroup/kogito-operator/issues).

## Kogito Operator development

Before you begin fixing issues or adding new features to the Kogito Operator, review the previous instructions for contributing to the Kogito Operator repository.

### Prerequisites

For code contributions, review the following prerequisites:

- Become familiar with the Go language. For an introduction to Go, see the official [Go Documentation](https://golang.org/doc/). For an intermediate or advanced Go resource, see [The Go Programming Language](https://www.amazon.com/gp/product/0134190440/) book.
- Become familiar with the [Operator SDK](https://github.com/operator-framework/operator-sdk). For more information, see the [Operator SDK Documentation](https://sdk.operatorframework.io/docs/) and use the [Memcached Operator](https://github.com/operator-framework/operator-sdk-samples/tree/master/go/memcached-operator) as an example.
- Ensure that you have all [Kogito Operator requirements](#requirements) set on your local machine. **You must use the listed versions.**

### Requirements

- [Docker](https://www.docker.com/)
- [Operator Courier](https://github.com/operator-framework/operator-courier) is used to build, validate and push Operator Artifacts
- [Operator SDK](https://github.com/operator-framework/operator-sdk) v1.21.0
- [Go](https://golang.org/) v1.17 is installed.
- [Golint dependency](https://pkg.go.dev/golang.org/x/lint/golint): go install golang.org/x/lint/golint@latest
- [Golangci-lint](https://golangci-lint.run/usage/install/)
- [Python 3.x](https://www.python.org/downloads/) v3.x is installed
- [Cekit](https://cekit.io/) v4.0.0+ is installed

### Building the Kogito Operator

Check if your $HOME/.config/containers/registries.conf is like this:
```bash
# Note that changing the order here may break tests.
unqualified-search-registries = ['docker.io']

[[registry]]
prefix="docker.io/library"
location="quay.io/libpod"

[[registry]]
location="localhost:5000"
insecure=true
```

To build the Kogito Operator, use the following command:

```bash
$ make
```

The output of this command is a ready-to-use Kogito Operator image that you can deploy in any namespace.

### Deploying to OpenShift 4.x for development purposes

Prerequisites:

1. Ensure that you [built the Kogito operator](#building-the-kogito-operator) first and pushed it to a registry reachable from the OpenShift cluster.

2. [Operator registry tool](https://github.com/operator-framework/operator-registry) is installed in your system and available as `opm` command


Follow the steps below:

1. Make sure that you defined environment variable `BUNDLE_IMG`. This variable will be used as an image tag used to store bundle image (containing all the operator metadata).

2. Make sure that you defined environment variable `CATALOG_IMG`. This variable will be used as an image tag used to store custom catalog image (containing reference to the custom bundle image).

3. Make sure that you defined environment variable `IMAGE` pointing to the custom Kogito operator image tag which you built.

4. Make sure that you defined environment variable `BUILDER` defining container runtime engine used to build and push images. Default value is `podman`.

5. Update the bundle metadata to point to the custom Kogito operator image by running this command:
```bash
$ make update-bundle
```

6. Build the bundle image by running this command:
```bash
$ make bundle-build
```

7. Push the built bundle image by running this command:
```bash
$ make bundle-push
```

8. Build the catalog image by running this command:
```bash
$ make catalog-build
```

9. Push the built catalog image by running this command:
```bash
$ make catalog-push
```

10. Define pushed catalog image as a catalog source for your OpenShift cluster by running this command:
```bash
$ cat << EOF | oc apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: custom-kogito-catalog
  namespace: openshift-marketplace
spec:
  sourceType: grpc
  image: ${CATALOG_IMG}
EOF
```

After several minutes, the Operator appears under **Catalog** -> **OperatorHub** in the OpenShift Web Console.
To find the Operator, filter the provider type by _custom-kogito-catalog_.

### Running BDD Tests

**REQUIREMENTS:**
* You need to be authenticated to the cluster before running the tests.
* Native tests need a node with at least 4 GiB of memory available (build resource request).

If you have an OpenShift cluster and admin privileges, you can run BDD tests with the following command:

```bash
$ make run-tests [key=value]*
```

You can set those optional keys:

<!--- tests configuration -->
- `feature` is a specific feature you want to run.  
  If you define a relative path, this has to be based on the "test" folder as the run is happening there.
  *Default are all enabled features from 'test/features' folder*  
  Example: feature=features/operator/deploy_quarkus_service.feature
- `tags` to run only specific scenarios. It is using tags filtering.  
  *Scenarios with '@disabled' tag are always ignored.*  
  Expression can be:
    - "@wip": run all scenarios with wip tag
    - "~@wip": exclude all scenarios with wip tag
    - "@wip && ~@new": run wip scenarios, but exclude new
    - "@wip,@undone": run wip or undone scenarios

  Complete list of supported tags and descriptions can be found in [List of test tags](#list-of-test-tags)
- `concurrent` is the number of concurrent tests to be ran.  
  *Default is 1.*
- `timeout` sets the timeout in minutes for the overall run.  
  *Default is 240 minutes.*
- `debug` to be set to true to activate debug mode.  
  *Default is false.*
- `load_factor` sets the tests load factor. Useful for the tests to take into account that the cluster can be overloaded, for example for the calculation of timeouts.  
  *Default is 1.*
- `ci` to be set if running tests with CI. Give CI name.
- `cr_deployment_only` to be set if you don't have a CLI built. Default will deploy applications via the CLI.
- `load_default_config` sets to true if you want to directly use the default test config (from test/.default_config)
- `container_engine` engine used to interact with images and local containers.
  *Default is docker.*
- `domain_suffix` domain suffix used for exposed services. Ignored when running tests on Openshift.
- `image_cache_mode` Use this option to specify whether you want to use image cache for runtime images. Available options are 'always', 'never' or 'if-available'(default).
- `http_retry_nb` sets the retry number for all HTTP calls in case it fails (and response code != 500).
  *Default is 3.*
- `olm_namespace` Set the namespace which is used for cluster scope operators. Default is 'openshift-operators'.
<!--- operator information -->
- `operator_image_tag` is the Operator image full name.
  *Default: quay.io/kiegroup/kogito-operator:${currentVersion}*.
- `operator_installation_source` Defines what source is used to install Kogito operator. Available options are `olm` and `yaml`.
  *Default is yaml*.
- `operator_catalog_image` Specifies catalog image containing Kogito operator bundle. Needs to be specified when `operator_installation_source` is set to `olm`.
- `use_product_operator` Set true if you want to run tests using product operator.
<!--- files/binaries -->
- `operator_yaml_uri` Url or Path to kogito-operator.yaml file.
*Default is ../kogito-operator.yaml*.
- `cli_path` set the built CLI path.  
  *Default is ./build/_output/bin/kogito*.
- `rhpam_operator_yaml_uri` Url or Path to rhpam-operator.yaml file.
*Default is ../rhpam-operator.yaml*.
<!--- runtime -->
- `services_{image_type}_{persistence_type}_image_tag` sets the services (jobs-service, data-index, ...) image tag.  
  image_type => data-index, explainibility, jobs-service, mgmt-console, task-console, trusty, trusty-ui  
  persistence_type => ephemeral, infinispan, mongodb, postgresql, redis  
  This will override those parameters for the given image: `services_image_version`, `services_image_namespace`, `services_image_registry`.
- `services_image_registry` sets the global services image registry.
- `services_image_name_suffix` sets the global services image name suffix to append to usual image names.
- `services_image_version` sets the global services image version.
<!--- build -->
- `custom_maven_repo_url` sets a custom Maven repository url for S2I builds, in case your artifacts are in a specific repository. See https://github.com/kiegroup/kogito-images/README.md for more information.
- `maven_mirror_url` is the Maven mirror URL.  
  This is helpful when you need to speed up the build time by referring to a closer Maven repository.
- `quarkus_platform_maven_mirror_url` is the Maven mirror url to be used when building app from source files with Quarkus, using the quarkus maven plugin.
- `build_builder_image_tag` sets the Builder image full tag.
- `build_runtime_jvm_image_tag` sets the Runtime JVM image full tag.
- `build_runtime_native_image_tag` sets the Runtime Native image full tag.
- `disable_maven_native_build_container` disables the default Maven native build done in container.
<!--- examples repository -->
- `examples_uri` sets the URI for the kogito-examples repository.  
  *Default is https://github.com/kiegroup/kogito-examples*.
- `examples_ref` sets the branch for the kogito-examples repository.
<!--- build runtime applications -->
- `runtime_application_image_registry` sets the registry for built runtime applications.
- `runtime_application_image_namespace` sets the namespace for built runtime applications.
- `runtime_application_image_name_prefix` sets the image name prefix to prepend to usual image names for built runtime applications.
- `runtime_application_image_name_suffix` sets the image name suffix to append to usual image names for built runtime applications.
- `runtime_application_image_version` sets the version for built runtime applications.
<!--- development options -->
- `show_scenarios` sets to true to display scenarios which will be executed.  
  *Default is false.*
- `show_steps` sets to true to display scenarios and their steps which will be executed.  
  *Default is false.*
- `dry_run` sets to true to execute a dry run of the tests, disable crds updates and display the scenarios which will be executed.  
  *Default is false.*
- `keep_namespace` sets to true to not delete namespace(s) after scenario run (WARNING: can be resources consuming ...).  
  *Default is false.*
- `namespace_name` to specify name of the namespace which will be used for scenario execution (intended for development purposes).
- `local_cluster` to be set to true if running tests using a local cluster.
  *Default is false.*
- `local_execution` to be set to true if running tests in local using either a local or remote cluster.
  *Default is false.*

Logs will be shown on the Terminal.

To save the test output in a local file for future reference, run the following command:

```bash
make run-tests 2>&1 | tee log.out
```

#### Running BDD tests with current branch

```
$ make
$ docker tag quay.io/kiegroup/kogito-operator:1.36.0 quay.io/{USERNAME}/kogito-operator:1.36.0
$ docker push quay.io/{USERNAME}/kogito-operator:1.36.0
$ make run-tests operator_image=quay.io/{USERNAME}/kogito-operator
```

**NOTE:** Replace {USERNAME} with the username/group you want to push to. Docker needs to be logged in to quay.io and be able to push to your username/group.

#### Running BDD tests with custom Kogito Build images' version

```bash
$ make run-tests build_image_version=<kogito_version>
```

#### Running smoke tests

The BDD tests do provide some smoke tests for a quick feedback on basic functionality:

```bash
$ make run-smoke-tests [key=value]*
```

It will run only tests tagged with `@smoke`.
All options from BDD tests do also apply here.

#### Running performance tests

The BDD tests also provide performance tests. These tests are ignored unless you
specifically provide the `@performance` tag or run:

```bash
$ make run-performance-tests [key=value]*
```

It will run only tests tagged with `@performance`.
All options from BDD tests do also apply here.

**NOTE:** Performance tests should be run without concurrency.

#### List of test tags

| Tag name           | Tag meaning                                                                        |
| ------------------ | ---------------------------------------------------------------------------------- |
| @smoke             | Smoke tests verifying basic functionality                                          |
| @performance       | Performance tests                                                                  |
| @travelagency      | Travel agency tests                                                                |
|                    |                                                                                    |
| @disabled          | Disabled tests, usually with comment describing reasons                            |
| @cli               | Tests to be executed only using Kogito CLI                                         |
|                    |                                                                                    |
| @rhpam             | Tests to be executed for RHPAM Kogito operator                                     |
|                    |                                                                                    |
| @springboot        | SpringBoot tests                                                                   |
| @quarkus           | Quarkus tests                                                                      |
|                    |                                                                                    |
| @dataindex         | Tests including DataIndex                                                          |
| @trusty            | Tests including Trusty                                                             |
| @jobsservice       | Tests including Jobs service                                                       |
| @managementconsole | Tests including Management console                                                 |
| @trustyui          | Tests including Trusty UI                                                          |
|                    |                                                                                    |
| @binary            | Tests using Kogito applications built locally and uploaded to OCP as a binary file |
| @asset             | Tests using Kogito applications built from assets uploaded to OCP                  |
| @native            | Tests using native build                                                           |
| @ignorelts         | Tests using native build that cannot be tested with Quarkus LTS version            |
| @persistence       | Tests verifying persistence capabilities                                           |
| @events            | Tests verifying eventing capabilities                                              |
| @discovery         | Tests checking service discovery functionality                                     |
| @usertasks         | Tests interacting with user tasks to check authentication/authorization            |
| @security          | Tests verifying security capabilities                                              |
| @failover          | Tests verifying failover capabilities                                              |
| @metrics           | Tests verifying metrics capabilities                                               |
|                    |                                                                                    |
| @resources         | Tests checking resource requests and limits                                        |
|                    |                                                                                    |
| @infinispan        | Tests using the infinispan operator                                                |
| @kafka             | Tests using the kafka operator                                                     |
| @keycloak          | Tests using the keycloak operator                                                  |
| @knative           | Tests using the Knative functionality                                              |
| @postgresql        | Tests using the PostgreSQL functionality                                           |
| @grafana           | Tests using the Grafana functionality                                              |
| @prometheus        | Tests using the Prometheus functionality                                           |

### Running the Kogito Operator locally

To run the Kogito Operator locally, change the log level at runtime with the `DEBUG` environment variable, as shown in the following example:

```bash
$ DEBUG=true make run
```

You can use the following command to vet, format, lint, and test your code:

```bash
$ make vet && make test
```

### Remote Debug Kogito Operator using Intellij IDEA

The operator will be deployed over an Kubernetes cluster. In this example we've taken minikube to deploy a Kubernetes cluster locally.

**Install Minikube:**

For installing the Minikube cluster please follow [this tutorial](https://kubernetes.io/docs/tasks/tools/install-minikube/)
```bash
$ minikube start
```
**Apply Operator manifests:**

```bash
$ export NAMESPACE=default
$ ./hack/install-manifests.sh
```
**Install Delve:**

Delve is a debugger for the Go programming language. For installing Delve please follow [this tutorial](https://github.com/go-delve/delve/tree/master/Documentation/installation)

**Start Operator in remote debug mode:**

```bash
$ cd cmd/manager
$ export WATCH_NAMESPACE=default
$ dlv debug --headless --listen=:2345 --api-version=2
```

verify logs on bash console for below line

```
API server listening at: [::]:2345
```

**Create the Go Remote run/debug configuration:**

1. Click `Edit | Run Configurations`. Alternatively, click the list of run/debug configurations on the toolbar and select `Edit Configurations`.  
![alt text](./docs/images/add_configuration.png?raw=true)
2. In the `Run/Debug Configurations` dialog, click the `Add` button (`+`) and select `Go Remote`.
![alt text](./docs/images/add_go_remote_config.png?raw=true)
3. In the Host field, type the host IP address (in our case `localhost`).
4. In the Port field, type the debugger port that you configured in above `dlv` command (in our case it's `2345`).
5. Click `OK`.          
![alt text](./docs/images/remote_debug_configurations.png?raw=true)
6. Put the breakpoint in your code.
7. From the list of `run/debug configurations` on the toolbar, select the created Go Remote configuration and click the `Debug <configuration_name>` button 

Running Kogito operator in remote debug on VSCode and GoLand is very similar to above procedure. Please follow these article to the setup remote debugger on [VSCode](https://dev.to/austincunningham/debug-kubernetes-operator-sdk-locally-using-vscode-130k) and [GoLand](https://dev.to/austincunningham/debug-kubernetes-operator-sdk-locally-in-goland-kl6)

## Guide for Kogito Developers
If you are a new developer looking for information on how to start working on the operator, please take a look at [this guide](docs/GUIDE_FOR_KOGITO_DEVS.md#introduction).

If you made changes in the core/runtimes part of the Kogito and want to test your changes against the operator, please follow [this guide](docs/GUIDE_FOR_RUNTIMES_DEVS.md#guide-for-coreruntimes-team-to-smoke-test-local-changes-on-openshiftk8s-cluster) to test your changes.

## Releasing Kogito Operator

When releasing the new version of kogito-operator in [community-operators](https://github.com/operator-framework/community-operators). One can use `make olm-manifests` it would create the manifests in the `build/_output/olm/<version>` directory in format which the community-operators repo expects.
One can then just copy the directory into the [kogito-operator](https://github.com/operator-framework/community-operators/tree/master/community-operators/kogito-operator) and raise the PR.

Note: One needs to create two separate PRs with this folder added in [upstream-kogito-operator](https://github.com/operator-framework/community-operators/tree/master/upstream-community-operators/kogito-operator) and [community-kogito-operator](https://github.com/operator-framework/community-operators/tree/master/community-operators/kogito-operator) respectively.

Before raising the PR, make sure the `replaces` field in the CSV is correct

## Kogito Operator tested integrations

Kogito operator integrates with various other technologies and operators. The tested certification matrix is listed below:

| Technology         | Tested version                                                |
| ------------------ | ------------------------------------------------------------- |
| Infinispan         | Infinispan operator 2.1.5 (deployed by OLM `2.1.x` channel)   |
| Kafka              | Strimzi 0.25.0 (deployed by OLM `stable` channel)             |
| Keycloak           | Keycloak operator 15.0.2 (deployed by OLM `alpha` channel)    |
| Prometheus         | Prometheus operator 0.47.0 (deployed by OLM `beta` channel)   |
| Grafana            | Grafana operator 3.10.3 (deployed by OLM `alpha` channel)     |
| Knative Eventing   | Knative Eventing 0.26.0                                       |
| MongoDB            | MongoDB Community Kubernetes Operator 0.2.2                   |
| PostgreSQL         | PostgreSQL 12.7 (deployed directly using image)               |
