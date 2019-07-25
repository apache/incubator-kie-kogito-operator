# Kogito Operator

[![Go Report Card](https://goreportcard.com/badge/github.com/kiegroup/kogito-cloud-operator)](https://goreportcard.com/report/github.com/kiegroup/kogito-cloud-operator)

## Requirements

- go v1.11+
- dep v0.5.x
- operator-sdk v0.7.0
- ocp 4.x
- kogito s2i imagestreams installed

## Build

```bash
make
```

## Upload to a container registry

e.g.

```bash
docker push quay.io/kiegroup/kogito-cloud-operator:<version>
```

## Deploy to OpenShift 4 using OLM

To install this operator on OpenShift 4 for end-to-end testing, make sure you have access to a quay.io account to create an application repository. Follow the [authentication](https://github.com/operator-framework/operator-courier/#authentication) instructions for Operator Courier to obtain an account token. This token is in the form of "basic XXXXXXXXX" and both words are required for the command.

Push the operator bundle to your quay application repository as follows:

```bash
operator-courier push deploy/catalog_resources/courier/0.1.0 kiegroup kogitocloud-operator 0.1.0 "basic XXXXXXXXX"
```

If pushing to another quay repository, replace _kiegroup_ with your username or other namespace. Also note that the push command does not overwrite an existing repository, and it needs to be deleted before a new version can be built and uploaded. Once the bundle has been uploaded, create an [Operator Source](https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md#linking-the-quay-application-repository-to-your-openshift-40-cluster) to load your operator bundle in OpenShift.

Note that the OpenShift cluster needs access to the created application. Make sure that the application is **public** or you have configured the private repository credentials in the cluster. To make the application public, go to your quay.io account, and in the _Applications_ tab look for the `kogitocloud-operator` application. Under the settings section click on _make public_ button.

```bash
## kogito imagestreams should already be installed/available ... e.g.
oc apply -f https://raw.githubusercontent.com/kiegroup/kogito-cloud/master/s2i/kogito-imagestream.yaml -n openshift
oc create -f deploy/catalog_resources/courier/kiecloud-operatorsource.yaml
```

Remember to replace _registryNamespace_ with your quay namespace. The name, display name and publisher of the operator are the only other attributes that may be modified.

It will take a few minutes for the operator to become visible under the _OperatorHub_ section of the OpenShift console _Catalog_. It can be easily found by filtering the provider type to _Custom_.


Also it's possible to verify the operator status by running:

```bash
oc describe operatorsource.operators.coreos.com/kiecloud-operators -n openshift-marketplace
```

## Deploy to OpenShift 3.11 manually

```bash
## kogito imagestreams should already be installed/available ... e.g.
oc apply -f https://raw.githubusercontent.com/kiegroup/kogito-cloud/master/s2i/kogito-imagestream.yaml -n openshift
oc new-project <project-name>
./hack/3.11deploy.sh
```

### Trigger a KogitoApp deployment

Use the OLM console to subscribe to the `kogito` Operator Catalog Source within your namespace. Once subscribed, use the console to `Create KogitoApp` or create one manually as seen below.

```bash
$ oc create -f deploy/crs/app_v1alpha1_kogitoapp_cr.yaml
kogitoapp.app.kiegroup.org/example-quarkus created
```

### Clean up a KogitoApp deployment

```bash
oc delete kogitoapp example-quarkus
```

## Development

Change log level at runtime w/ the `DEBUG` environment variable. e.g. -

```bash
make dep
make clean
DEBUG="true" operator-sdk up local --namespace=<namespace>
```

Before submitting PR, please be sure to generate, vet, format, and test your code. This all can be done with one command.

```bash
make test
```
