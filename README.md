# Kogito Operator

[![Go Report Card](https://goreportcard.com/badge/github.com/kiegroup/kogito-cloud-operator)](https://goreportcard.com/report/github.com/kiegroup/kogito-cloud-operator) [![CircleCI](https://circleci.com/gh/kiegroup/kogito-cloud-operator.svg?style=svg)](https://circleci.com/gh/kiegroup/kogito-cloud-operator)

Kogito Operator was designed to deploy [Kogito Runtimes](https://github.com/kiegroup/kogito-runtimes) services from source and every piece of infrastructure that the services might need like SSO ([Keycloak](https://github.com/integr8ly/keycloak-operator)) and Persistence ([Infinispan](https://github.com/infinispan/infinispan-operator)).

## Requirements

- go v1.12+
- [operator-sdk](https://github.com/operator-framework/operator-sdk/releases) v0.9.0
- ocp 4.x (you can use [CRC](https://github.com/code-ready/crc) for local deployment)
- [kogito s2i imagestreams](https://raw.githubusercontent.com/kiegroup/kogito-cloud/master/s2i/kogito-imagestream.yaml) installed

## Architecture

The actual architecture has only one [controller](https://godoc.org/github.com/kubernetes-sigs/controller-runtime/pkg#hdr-Controller) that's responsible for deploying the application from source. It's on the [roadmap](https://github.com/kiegroup/kogito-runtimes/wiki/Roadmap) to have two more controllers to handle SSO and Persistence. The following image illustrates the general idea:

![Kogito Operator General Architecture](docs/img/general_archictecture.png?raw=true)

One of the most important responsibilities of the controller is the [Reconcile Loop](https://github.com/operator-framework/operator-sdk/blob/master/doc/user-guide.md#reconcile-loop). Inside this "loop" the controller will ensure that it has every resource (Kubernetes and OpenShift objects) created and updated accordingly.

We aim to avoid having a huge monolith inside the reconcile loop that does it all. With that in mind, we separated the responsibility of making the Kubernetes and OpenShift API calls to a package that we call [`client`](pkg/client). Kubernetes/OpenShift resources that the controller need is defined and created inside the [`builder`](pkg/controller/kogitoapp/builder) package. `Builder` communicates with the `client` package to bind or create the objects in the cluster. The `Controller` also make calls to `client` to perform certain taks during the `reconcile` loop.

Take a look at the following diagram to have a general idea of what we're talking about:

![Kogito Operator Packages Structure](docs/img/packages_structure.png?raw=true)

`Controller` will orchestrate all operations through `client` and `builder` calls by using its domain model (`Type`). `Controller` also will delegate to `Builder` the resources bind and creation.

### Client

In this package we handle all Kubernetes/OpenShift API calls, transforming those operations into meaningful functions that can be used across all controller operations. Take for example the `CreateIfNotExists` function:

```go
func CreateIfNotExists(resource meta.ResourceObject) (bool, error) {
	if exists, err := Fetch(resource); err == nil && !exists {
		err = cli.Create(context.TODO(), resource)
		if err != nil {
			return false, err
		}
		return true, nil
	} else if err != nil {
		return false, err
	}
	return false, nil
}
```

It will fetch a particular named resource (e.g. a `ServiceAccount`), and if does not exist, the function will create a new one using the API.

We try to do our best to have a code easy to read even for those who are not familiar with the Go language.

### Builder

The `builder` package defines the structure and dependencies of every resource according to the controller requirements. The following diagram illustrates the relationship between the OpenShift resources for deploying a new Kogito Application through the `KogitoApp` controller:

![Kogito App Resources Composition](docs/img/kogitoapp_resource_composition.png?raw=true)

The `builder` package ensures that each object is created accordingly. Take "for example" the `NewRoleBinding` function:

```go
func NewRoleBinding(kogitoApp *v1alpha1.KogitoApp, serviceAccount *corev1.ServiceAccount) (roleBinding rbacv1.RoleBinding) {
	roleBinding = rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			...
		},
		RoleRef: rbacv1.RoleRef{
			...
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      KindServiceAccount.Name,
				Namespace: serviceAccount.Namespace,
				Name:      serviceAccount.Name,
			},
		},
    }
    ...
	return roleBinding
}
```

This function will create a new `RoleBinding` that depends on the `ServiceAccount`, with the default role references, labels and annotations.

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

If pushing to another quay repository, replace _kiegroup_ with your username or other namespace. Notice that the push command does not overwrite an existing repository, and the bundle needs to be deleted before a new version can be built and uploaded. Once the bundle has been uploaded, create an [Operator Source](https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md#linking-the-quay-application-repository-to-your-openshift-40-cluster) to load your operator bundle in OpenShift.

Note that the OpenShift cluster needs access to the created application. Make sure that the application is **public** or you have configured the private repository credentials in the cluster. To make the application public, go to your quay.io account, and in the _Applications_ tab look for the `kogitocloud-operator` application. Under the settings section click on _make public_ button.

```bash
## kogito imagestreams should already be installed/available ... e.g.
oc apply -f https://raw.githubusercontent.com/kiegroup/kogito-cloud/master/s2i/kogito-imagestream.yaml -n openshift
oc create -f deploy/catalog_resources/courier/kiecloud-operatorsource.yaml
```

Remember to replace _registryNamespace_ with your quay namespace. The name, display name and publisher of the operator are the only other attributes that may be modified.

It will take a few minutes for the operator to become visible under the _OperatorHub_ section of the OpenShift console _Catalog_. The Operator can be easily found by filtering the provider type to _Custom_.


It's possible to verify the operator status by running:

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

## Running Locally 

Change log level at runtime with the `DEBUG` environment variable. e.g. -

```bash
make mod
make clean
DEBUG="true" operator-sdk up local --namespace=<namespace>
```

Before submitting PR, please be sure to read the [contributors guide](CONTRIBUTING.MD##contributors-guide). 

It's always worth noting that you should generate, vet, format, lint, and test your code. This all can be done with one command.

```bash
make test
```

## Contributing

Please take a look at the [Contributing to Kogito Operator](CONTRIBUTING.MD) guide.