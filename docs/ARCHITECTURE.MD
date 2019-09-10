# Architecture

The actual architecture has two [controllers](https://godoc.org/github.com/kubernetes-sigs/controller-runtime/pkg#hdr-Controller) that's responsible for deploying the application from source and to deploy the [Data Index Service](https://github.com/kiegroup/kogito-runtimes/wiki/Data-Index-Service). It's on the [roadmap](https://github.com/kiegroup/kogito-runtimes/wiki/Roadmap) to have one more controller to handle SSO. The following image illustrates the general idea:

![Kogito Operator General Architecture](img/general_archictecture.png?raw=true)

One of the most important responsibilities of the controller is the [Reconcile Loop](https://github.com/operator-framework/operator-sdk/blob/master/doc/user-guide.md#reconcile-loop). Inside this "loop" the controller will ensure that it has every resource (Kubernetes and OpenShift objects) created and updated accordingly.

We aim to avoid having a huge monolith inside the reconcile loop that does it all. With that in mind, we separated the responsibility of making the Kubernetes and OpenShift API calls to a package that we call [`client`](../pkg/client). Kubernetes/OpenShift resources, which the controller needs, are defined and created inside the [`resource`](../pkg/controller/kogitoapp/resource) package. `resource` communicates with the `client` package to bind or create the objects in the cluster. The `controller` also make calls to `client` to perform certain tasks during the `reconcile` loop.

Take a look at the following diagram to have a general idea of what we're talking about:

![Kogito Operator Packages Structure](img/packages_structure.png?raw=true)

`Controller` will orchestrate all operations through `Client` and `Resource` calls by using its domain model (`Type`). `Controller` also will delegate to `Resource` the resources bind and creation.

## Client

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

## Resource

The `resource` package defines the structure and dependencies of every resource according to the controller requirements. The following diagram illustrates the relationship between the OpenShift resources for deploying a new Kogito Application through the `KogitoApp` controller:

![Kogito App Resources Composition](img/kogitoapp_resource_composition.png?raw=true)

The `resource` package ensures that each object is created accordingly. Take "for example" the `NewRoute` function:

```go
func NewRoute(kogitoApp *v1alpha1.KogitoApp, service *corev1.Service) (route *routev1.Route, err error) {
	...
	route = &routev1.Route{
		ObjectMeta: service.ObjectMeta,
		Spec: routev1.RouteSpec{
			Port: &routev1.RoutePort{
				...
			},
			To: routev1.RouteTargetReference{
				Kind: meta.KindService.Name,
				Name: service.Name,
			},
		},
	}
	...
	return route, nil
}
```

This function will create a new `Route` that depends on the `Service`, with the service's references, labels and annotations.
