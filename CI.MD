# CI
Kogito-cloud-operator uses [CircleCI](https://circleci.com/) for automated image push to Quay registry.

#### Workflow
Each commit to master branch will trigger the CI to build and push an image to Quay with the version-RC tag. 
Please refer the version information [here](https://github.com/kiegroup/kogito-cloud-operator/blob/master/version/version.go#L19), also make ensure the version is sync with the version specified in [Makefile](https://github.com/sbuvaneshkumar/kogito-cloud-operator/blob/master/Makefile#L7).

If the git tags are specificed, make ensure to update the release version env in your CircleCI account under project settings.

#### Project settings
Following environment variables needs to be set under CircleCI project settings in order for the CI to function.
```
CIRCLE_TAG    - Release tag
REGISTRY_USER - Account username for registry.redhat.io
REGISTRY_PASS - Account password for registry.redhat.com
QUAY_USER     - Account username for Quay
QUAY_PASS     - Account password for Quay
```
#### Credentials
Only the application credential (from Quay) should be used to write in this specific repository.
Also, please make ensure the credential has write access to this repo.
