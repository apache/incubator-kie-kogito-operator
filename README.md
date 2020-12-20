# Kogito Operator

[![Go Report Card](https://goreportcard.com/badge/github.com/kiegroup/kogito-cloud-operator)](https://goreportcard.com/report/github.com/kiegroup/kogito-cloud-operator) [![CircleCI](https://circleci.com/gh/kiegroup/kogito-cloud-operator.svg?style=svg)](https://circleci.com/gh/kiegroup/kogito-cloud-operator)

The Kogito Operator deploys [Kogito Runtimes](https://github.com/kiegroup/kogito-runtimes) services from source and all infrastructure requirements for the services, such as persistence with [Infinispan](https://infinispan.org/) and messaging with [Apache Kafka](https://kafka.apache.org/). Kogito provides a command-line interface (CLI) that enables you to interact with the Kogito Operator for deployment tasks.

For information about the Kogito Operator architecture and instructions for using the operator and CLI to deploy Kogito services and infrastructures, see the official [Kogito Documentation](https://docs.jboss.org/kogito/release/latest/html_single/#chap-kogito-deploying-on-openshift) page.

# Table of Contents

* [Contributing to the Kogito Operator](#contributing-to-the-kogito-operator)
  * [Trying the Kogito Operator](#trying-the-kogito-operator)
  * [Prerequisites](#prerequisites)
  * [Kogito Operator environment](#kogito-operator-environment)
    * [Kogito Operator unit tests](#kogito-operator-unit-tests)
    * [Kogito Operator collaboration and pull requests](#kogito-operator-collaboration-and-pull-requests)
  * [Kogito Operator development](#kogito-operator-development)
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
  
Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc)

# Contributing to the Kogito Operator

Thank you for your interest in contributing to this project!

Any kind of contribution is welcome: code, design ideas, bug reporting, or documentation (including this page).

## Trying the Kogito Operator

You can quickly install the operator in your local cluster by executing the following command:

```shell script
NAMESPACE=mynamespace
VERSION=<current_operator_version>
kubectl apply -n "${NAMESPACE}" -f "https://github.com/kiegroup/kogito-cloud-operator/releases/download/${VERSION}kogito-operator.yaml"
```

Replace the values above to match your context. The version information you can grab from the [releases page](https://github.com/kiegroup/kogito-cloud-operator/releases).

Alternatively, if you cloned this repo just do:

```shell script
$ ./hack/install.sh
```

The script will download the latest version and install the resources for you in the current namespace.
You can set the `VERSION` and `NAMESPACE` variables before running the script to control which version to install in the given namespace.

## Prerequisites

For code contributions, review the following prerequisites:

- Become familiar with the Go language. For an introduction to Go, see the official [Go Documentation](https://golang.org/doc/). For an intermediate or advanced Go resource, see [The Go Programming Language](https://www.amazon.com/gp/product/0134190440/) book.
- Become familiar with the [Operator SDK](https://github.com/operator-framework/operator-sdk). For more information, see the [Operator SDK Documentation](https://sdk.operatorframework.io/docs/) and use the [Memcached Operator](https://github.com/operator-framework/operator-sdk-samples/tree/master/go/memcached-operator) as an example.
- Ensure that you have all [Kogito Operator requirements](#requirements) set on your local machine. **You must use the listed versions.**

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

In general, the unit tests that are provided with the Kogito Operator are based on that Operator SDK testing resource. You might encounter minor issues as you create specific OpenShift APIs such as `BuildConfig` and `DeploymentConfig` that are not listed there. For an example test case with sample API calls, see the [`kogitobuild_controller_test.go`](../master/pkg/controller/kogitobuild/kogitobuild_controller_test.go) test file.

### Kogito Operator collaboration and pull requests

Before you start to work on a new proposed feature or on a fix for a bug, [open an issue](https://github.com/kiegroup/kogito-cloud-operator/issues) to discuss your idea or bug report with the maintainers. You can also work on a [JIRA issue](https://issues.jboss.org/issues/?jql=project+%3D+KOGITO+AND+component+%3D+Operator) that has been reported. A developer might already be assigned to address the issue, but you can leave a comment in the JIRA asking if they need some help.

After you update the source with your new proposed feature or bug fix, open a [pull request (PR)](https://help.github.com/en/articles/about-pull-requests) that meets the following requirements:

- You have a JIRA associated with the PR. 
- Your PR has the name of the JIRA in the title, for example, `[KOGITO-XXX] - Awesome feature that solves it all`.
- The PR solves only the problem described in the JIRA.
- You have written unit tests for the particular fix or feature.
- You ran `make vet` and `make test` before submitting the PR and everything is working accordingly.
- You tested the feature on an actual OpenShift cluster.

After you send your PR, a maintainer will review your code and might ask you to make changes and to [squash your commits](https://stackoverflow.com/questions/5189560/squash-my-last-x-commits-together-using-git) before we can merge.

If you have any questions, contact a Kogito Operator maintainer in the [issues page](https://github.com/kiegroup/kogito-cloud-operator/issues).

## Kogito Operator development

Before you begin fixing issues or adding new features to the Kogito Operator, review the previous instructions for contributing to the Kogito Operator repository.

### Requirements

- [Docker](https://www.docker.com/)
- [Operator Courier](https://github.com/operator-framework/operator-courier) is used to build, validate and push Operator Artifacts
- [Operator SDK](https://github.com/operator-framework/operator-sdk) v1.2.0
- [Go](https://golang.org/) v1.14 is installed.
- [Golint dependency](https://pkg.go.dev/golang.org/x/lint/golint): go get -u golang.org/x/lint/golint
- [Golangci-lint](https://golangci-lint.run/usage/install/)

### Building the Kogito Operator

To build the Kogito Operator, use the following command:

```bash
$ make
```

The output of this command is a ready-to-use Kogito Operator image that you can deploy in any namespace.

### Deploying to OpenShift 4.x for development purposes

To install the Kogito Operator on OpenShift 4.x for end-to-end (E2E) testing, ensure that you have access to a `quay.io`
account to create an application repository.

Follow the steps below:

1. Run `make prepare-olm version=2.0.0-snapshot`. Bear in mind that if there're different versions
in the `deploy/olm-catalog/kogito-operator/kogito-operator.package.yaml` file, every CSV must
be included in the output folder. At this time, the script did not copy previous CSV versions to the
output folder, so it must be copied manually.

2. Grab [Quay credentials](https://github.com/operator-framework/operator-courier/#authentication) with:

```
$ export QUAY_USERNAME=youruser
$ export QUAY_PASSWORD=yourpass

$ AUTH_TOKEN=$(curl -sH "Content-Type: application/json" -XPOST https://quay.io/cnr/api/v1/users/login -d '
{
    "user": {
        "username": "'"${QUAY_USERNAME}"'",
        "password": "'"${QUAY_PASSWORD}"'"
    }
}' | jq -r '.token')
```

3. Set courier variables:

```
$ export OPERATOR_DIR=build/_output/operatorhub/
$ export QUAY_NAMESPACE=kiegroup # should be different in your environment
$ export PACKAGE_NAME=kogito-operator
$ export PACKAGE_VERSION=2.0.0-snapshot
$ export TOKEN=$AUTH_TOKEN
```

If you push to another quay repository, replace `QUAY_NAMESPACE` with your user name or the other namespace.
The push command does not overwrite an existing repository, so you must delete the bundle before you can
build and upload a new version. After you upload the bundle, create an
[Operator Source](https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md#linking-the-quay-application-repository-to-your-openshift-40-cluster)
to load your operator bundle in OpenShift.

4. Run `operator-courier` to publish the operator application to Quay:

```
operator-courier push "$OPERATOR_DIR" "$QUAY_NAMESPACE" "$PACKAGE_NAME" "$PACKAGE_VERSION" "$TOKEN"
```

5. Check if the application was pushed successfully in Quay.io. The OpenShift cluster needs access to the created application.
Ensure that the application is **public** or that you have configured the private repository credentials in the cluster.
To make the application public, go to your `quay.io` account, and in the **Applications** tab look for the `kogito-operator`
application. Under the settings section, click **make public**.

6. Publish the operator source to your OpenShift cluster:

```
$ oc create -f deploy/olm-catalog/kogito-operator/kogito-operator-operatorsource.yaml
```

Replace `registryNamespace` in the `kogito-operator-operatorsource.yaml` file with your quay namespace.
The name, display name, and publisher of the Operator are the only other attributes that you can modify.

After several minutes, the Operator appears under **Catalog** -> **OperatorHub** in the OpenShift Web Console.
To find the Operator, filter the provider type by _Custom_.

To verify the operator status, run the following command:

```bash
$ oc describe operatorsource.operators.coreos.com/kogito-operator -n openshift-marketplace
```

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
- `local` to be set to true if running tests in local using either a local or remote cluster.
  *Default is false.*
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
- `operator_image` is the Operator image full name.  
  *Default: operator_image=quay.io/kiegroup/kogito-cloud-operator*.
- `operator_tag` is the Operator image tag.  
  *Default is the current version*.
<!--- files/binaries -->
- `deploy_uri` set operator *deploy* folder.  
  *Default is ./deploy*.
- `cli_path` set the built CLI path.  
  *Default is ./build/_output/bin/kogito*.
<!--- runtime -->
- `services_image_version` sets the services (jobs-service, data-index, ...) image version.
- `services_image_namespace` sets the services (jobs-service, data-index, ...) image namespace.
- `services_image_registry` sets the services (jobs-service, data-index, ...) image registry.
- `data_index_image_tag` sets the Kogito Data Index image tag ('services_image_version' is ignored)
- `explainability_image_tag` sets the Kogito Explainability image tag ('services_image_version' is ignored)
- `jobs_service_image_tag` sets the Kogito Jobs Service image tag ('services_image_version' is ignored)
- `management_console_image_tag` sets the Kogito Management Console image tag ('services_image_version' is ignored)
- `task_console_image_tag` sets the Kogito Task Console image tag ('services_image_version' is ignored)
- `trusty_image_tag` sets the Kogito Trusty image tag ('services_image_version' is ignored)
- `trusty_ui_image_tag` sets the Kogito Trusty UI image tag ('services_image_version' is ignored)
<!--- build -->
- `custom_maven_repo` sets a custom Maven repository url for S2I builds, in case your artifacts are in a specific repository. See https://github.com/kiegroup/kogito-images/README.md for more information.
- `maven_mirror` is the Maven mirror URL.  
  This is helpful when you need to speed up the build time by referring to a closer Maven repository.
- `build_image_registry` sets the build image registry.
- `build_image_namespace` sets the build image namespace.
- `build_image_name_suffix` sets the build image name suffix to append to usual image names.
- `build_image_version` sets the build image version
- `build_s2i_image_tag` sets the build S2I image full tag.
- `build_runtime_image_tag` sets the build Runtime image full tag.
- `disable_maven_native_build_container` disables the default Maven native build done in container.
<!--- examples repository -->
- `examples_uri` sets the URI for the kogito-examples repository.  
  *Default is https://github.com/kiegroup/kogito-examples*.
- `examples_ref` sets the branch for the kogito-examples repository.
<!--- build runtime applications -->
- `runtime_application_image_registry` sets the registry for built runtime applications.
- `runtime_application_image_namespace` sets the namespace for built runtime applications.
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

Logs will be shown on the Terminal.

To save the test output in a local file for future reference, run the following command:

```bash
make run-tests 2>&1 | tee log.out
```

#### Running BDD tests with current branch

```
$ make
$ docker tag quay.io/kiegroup/kogito-cloud-operator:2.0.0-snapshot quay.io/{USERNAME}/kogito-cloud-operator:2.0.0-snapshot
$ docker push quay.io/{USERNAME}/kogito-cloud-operator:2.0.0-snapshot
$ make run-tests operator_image=quay.io/{USERNAME}/kogito-cloud-operator
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
| @olm               | OLM integration tests                                                              |
| @travelagency      | Travel agency tests                                                                |
|                    |                                                                                    |
| @disabled          | Disabled tests, usually with comment describing reasons                            |
| @cli               | Tests to be executed only using Kogito CLI                                         |
|                    |                                                                                    |
| @springboot        | SpringBoot tests                                                                   |
| @quarkus           | Quarkus tests                                                                      |
|                    |                                                                                    |
| @dataindex         | Tests including DataIndex                                                          |
| @trusty            | Tests including Trusty                                                             |
| @jobsservice       | Tests including Jobs service                                                       |
| @managementconsole | Tests including Management console                                                 |
| @trustyui          | Tests including Trusty UI                                                          |
| @infra             | Tests checking KogitoInfra functionality                                           |
|                    |                                                                                    |
| @binary            | Tests using Kogito applications built locally and uploaded to OCP as a binary file |
| @native            | Tests using native build                                                           |
| @persistence       | Tests verifying persistence capabilities                                           |
| @events            | Tests verifying eventing capabilities                                              |
| @discovery         | Tests checking service discovery functionality                                     |
| @usertasks         | Tests interacting with user tasks to check authentication/authorization            |
| @security          | Tests verifying security capabilities                                              |
| @failover          | Tests verifying failover capabilities                                              |
|                    |                                                                                    |
| @resources         | Tests checking resource requests and limits                                        |
|                    |                                                                                    |
| @infinispan        | Tests using the infinispan operator                                                |
| @kafka             | Tests using the kafka operator                                                     |
| @keycloak          | Tests using the keycloak operator                                                  |

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

If you made changes in the core/runtimes part of the Kogito and want to test your changes against the operator. Please follow this [guide](docs/GUIDE_TO_KOGITO_DEVS.md) to test your changes.
