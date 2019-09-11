# Contributing to Kogito Operator

Many thanks for your interest in contributing to this project. We love contributions! :heart:

Any kind of contribution is welcome: code, design ideas, bug reporting, documentation (including this one!) and so on.

## Some tips before you start

If you understand that you can contribute with code, here's some advice that might help during your journey:

1. Make yourself comfortable with the Go language. If you're a new programmer, take a look at the [official docs](https://golang.org/doc/). For a more advanced/intermediate level, [The Go Programming Language](https://www.amazon.com/gp/product/0134190440/) book might be a good starting point.
2. Become familiar with [the Operator SDK](https://github.com/operator-framework/operator-sdk). There's a [useful guide](https://github.com/operator-framework/operator-sdk/blob/master/doc/user-guide.md) that you can follow using [Memcached Operator](https://github.com/operator-framework/operator-sdk-samples/tree/master/memcached-operator) as an example.
3. Make sure to have all [requirements](../README.md#requirements) set on your local machine.

## Contributors Guide

Please follow the sections below to be able to start contributing to the Kogito Operator project.

### Environment

Start with having all the [requirements](../README.md#requirements) set on your local machine and **make sure** you stick with the versions described. 

The operator SDK has regular updates. We try to update our code to use the most recent updated SDK as soon as possible. Since they are not on a major version yet, incompatibilities might happen.

If you don't have a preferred IDE, we recommend using VS Code with the following plugins:

- [`vscode-go`](https://github.com/Microsoft/vscode-go) for Go lang tools support

To use Go modules with VS Code, follow the ["Go modules support in VS Code"](https://github.com/Microsoft/vscode-go/wiki/Go-modules-support-in-Visual-Studio-Code) guide to configure the IDE properly.

If you care to debug, follow the ["Debugging Go code using VS Code"](https://github.com/Microsoft/vscode-go/wiki/Debugging-Go-code-using-VS-Code) guide to be able to configure the IDE properly.

### Testing

We love tests and we do our best to always have a unit test to ensure that a certain feature was implemented correctly. 

The Operator SDK has a great documentation explaining how to perform [unit testing](https://github.com/operator-framework/operator-sdk/blob/master/doc/user/unit-testing.md) to an operator. Take your time and have a good read of it to have an overall understanding.

In general, we follow the recommendations listed in that article to write unit tests. There's a small issue that might take you off guard when mocking up specific OpenShift APIs like `BuildConfig` and `DeploymentConfig` that are not listed there. Take a look at [`kogitoapp_controller_test.go`](../pkg/controller/kogitoapp/kogitoapp_controller_test.go) file as a reference of how to set up your test cases and create fake API calls.

[End-to-end (e2e)](https://github.com/operator-framework/operator-sdk/blob/master/doc/test-framework/writing-e2e-tests.md) tests is a topic that we are pursuing and it's on our [backlog](https://issues.jboss.org/browse/KOGITO-130) to have it solved soon. 

E2e is basically an automated integration test written to have it running on a local Kubernetes/OpenShift cluster.

### Sending a Pull Request

Before starting to work on a new feature that you found to be missing or on a fix for a bug that you might be facing, [please open an issue](https://github.com/kiegroup/kogito-cloud-operator/issues) to discuss your idea or bug report with the maintainers.

If you're willing to collaborate, you can always work on a [JIRA issue](https://issues.jboss.org/issues/?jql=project+%3D+KOGITO+AND+component+%3D+Operator) that has been reported. The chances are that we already have someone on it, but it's always worth leaving a comment at the JIRA asking if they need some help.

Once you have your feature/bug fixed, open a [Pull Request (PR)](https://help.github.com/en/articles/about-pull-requests) to us and make sure that:

1. You have a JIRA associated with the PR
2. Your PR has the name of the JIRA in the title, e.g. `[KOGITO-XXX] - Awesome feature that solves it all`
3. The PR solves only the problem described in the JIRA
4. You have written unit tests for the particular fix/feature
5. You ran `make test` before submitting the PR and everything is working accordingly
6. You tested the feature on an actual OpenShift cluster

After sending your PR, a maintainer will review your code and might ask for changes. That said, you might be asked to [squash your commit](https://stackoverflow.com/questions/5189560/squash-my-last-x-commits-together-using-git) before we merge.

In case of any more questions, please fell free to reach us via the [issues page](https://github.com/kiegroup/kogito-cloud-operator/issues).