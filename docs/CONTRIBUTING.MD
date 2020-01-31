# Contributing to the Kogito Operator

Thank you for your interest in contributing to this project!

Any kind of contribution is welcome: code, design ideas, bug reporting, or documentation (including this page).

## Prerequisites

For code contributions, review the following prerequisites:

- Become familiar with the Go language. For an introduction to Go, see the official [Go Documentation](https://golang.org/doc/). For an intermediate or advanced Go resource, see [The Go Programming Language](https://www.amazon.com/gp/product/0134190440/) book.
- Become familiar with the [Operator SDK](https://github.com/operator-framework/operator-sdk). For more information, see the [Operator SDK User Guide](https://github.com/operator-framework/operator-sdk/blob/master/doc/user-guide.md) and use the [Memcached Operator](https://github.com/operator-framework/operator-sdk-samples/tree/master/memcached-operator) as an example.
- Ensure that you have all [Kogito Operator requirements](../README.md#kogito-operator-requirements) set on your local machine. **You must use the listed versions.**

## Kogito Operator environment

The Operator SDK is updated regularly and the Kogito Operator code typically uses the most recent SDK updates as soon as possible. The Operate SDK has not reached a major version yet, so incompatibilities might occur.

If you do not have a preferred IDE, use Visual Studio Code with the [`vscode-go`](https://github.com/Microsoft/vscode-go) plugin for Go language tools support.

To use Go modules with VS Code, see [Go modules support in VS Code](https://github.com/Microsoft/vscode-go/wiki/Go-modules-support-in-Visual-Studio-Code).

To debug Go in your VS code, see [Debugging Go code using VS Code](https://github.com/Microsoft/vscode-go/wiki/Debugging-Go-code-using-VS-Code).

### Kogito Operator unit tests

For information about Operator SDK testing, see [Unit testing with the Operator SDK](https://github.com/operator-framework/operator-sdk/blob/master/doc/user/unit-testing.md).

In general, the unit tests that are provided with the Kogito Operator are based on that Operator SDK testing resource. You might encounter minor issues as you create specific OpenShift APIs such as `BuildConfig` and `DeploymentConfig` that are not listed there. For an example test case with sample API calls, see the [`kogitoapp_controller_test.go`](../pkg/controller/kogitoapp/kogitoapp_controller_test.go) test file.

### Kogito Operator collaboration and pull requests

Before you start to work on a new proposed feature or on a fix for a bug, [open an issue](https://github.com/kiegroup/kogito-cloud-operator/issues) to discuss your idea or bug report with the maintainers. You can also work on a [JIRA issue](https://issues.jboss.org/issues/?jql=project+%3D+KOGITO+AND+component+%3D+Operator) that has been reported. A developer might already be assigned to address the issue, but you can leave a comment in the JIRA asking if they need some help.

After you update the source with your new proposed feature or bug fix, open a [pull request (PR)](https://help.github.com/en/articles/about-pull-requests) that meets the following requirements:

- You have a JIRA associated with the PR.
- Your PR has the name of the JIRA in the title, for example, `[KOGITO-XXX] - Awesome feature that solves it all`.
- The PR solves only the problem described in the JIRA.
- You have written unit tests for the particular fix or feature.
- You ran `make test` before submitting the PR and everything is working accordingly.
- You tested the feature on an actual OpenShift cluster.

After you send your PR, a maintainer will review your code and might ask you to make changes and to [squash your commits](https://stackoverflow.com/questions/5189560/squash-my-last-x-commits-together-using-git) before we can merge.

If you have any questions, contact a Kogito Operator maintainer in the [issues page](https://github.com/kiegroup/kogito-cloud-operator/issues).
