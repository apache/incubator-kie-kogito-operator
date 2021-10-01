Many thanks for submitting your Pull Request :heart:! 

Please make sure your PR meets the following requirements:

- [ ] You have read the [contributors' guide](https://github.com/kiegroup/kogito-operator/blob/main/README.md#contributing-to-the-kogito-operator)
- [ ] Pull Request title is properly formatted: `[KOGITO-XYZ] Subject`
- [ ] Pull Request contains a link to the JIRA issue
- [ ] Pull Request contains a description of the issue
- [ ] Pull Request does not include fixes for issues other than the main ticket
- [ ] Your feature/bug fix has a unit test that verifies it
- [ ] You've ran `make before-pr` and everything is working accordingly
- [ ] You've tested the new feature/bug fix in an actual OpenShift cluster
- [ ] You've added a [RELEASE_NOTES.md](https://github.com/kiegroup/kogito-operator/blob/main/RELEASE_NOTES.md) entry regarding this change

<details>
<summary>
How to retest this PR or trigger a specific build:
</summary>

* <b>Run operator BDD testing</b>
  Please add comment: <b>/jenkins test</b>

* <b>Run RHPAM operator BDD testing</b>
  Please add comment: <b>/jenkins RHPAM test</b>
</details>