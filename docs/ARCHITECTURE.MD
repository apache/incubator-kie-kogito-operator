# Kogito Operator architecture

The Kogito Operator architecture uses two [controllers](https://godoc.org/github.com/kubernetes-sigs/controller-runtime/pkg#hdr-Controller) to deploy the application from source and to deploy the [Kogito Data Index Service](https://github.com/kiegroup/kogito-runtimes/wiki/Data-Index-Service). One more controller to handle SSO is on the [Kogito roadmap](https://github.com/kiegroup/kogito-runtimes/wiki/Roadmap).

The following diagram illustrates the general Kogito Operator architecture:

![Kogito Operator General Architecture](img/general_archictecture.png?raw=true)

An important function of the controller is the [Reconcile Loop](https://github.com/operator-framework/operator-sdk/blob/master/doc/user-guide.md#reconcile-loop). Inside this "loop" the controller ensures that all resources (Kubernetes and OpenShift objects) are created and updated accordingly.

To optimize Kogito Operator performance, the reconcile loop is intentionally not a monolithic system that performs all necessary functions for the Kogito Operator. Instead, the reconcile loop delegates Kubernetes and OpenShift API calls to the [`client`](../pkg/client) package, and the Kubernetes and OpenShift resources that the controller requires are defined and created in the [`resource`](../pkg/controller/kogitoapp/resource) package. The `resource` package communicates with the `client` package to bind or create the objects in the cluster. The `controller` also makes calls to  the `client` to perform certain tasks during the `reconcile` loop.

The following diagram illustrates the relationship between the controller, client, and resource:

![Kogito Operator Packages Structure](img/packages_structure.png?raw=true)

The `Controller` orchestrates all operations through `Client` and `Resource` calls by using its domain model (`Type`). The `Controller` also delegates to `Resource` the binding and creation of resources.

## Client package

The `client` package handles all Kubernetes and OpenShift API calls, transforming those operations into meaningful functions that can be used across all controller operations.

For example, consider the following `CreateIfNotExists` function:

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

This function fetches a named resource, such as `ServiceAccount`, and if the resource does not exist, the function creates a new one using the API.

## Resource package

The `resource` package defines the structure and dependencies of every resource according to the controller requirements. The following diagram illustrates the relationship between the OpenShift resources for deploying a new Kogito Application through the `KogitoApp` controller:

![Kogito App Resources Composition](img/kogitoapp_resource_composition.png?raw=true)

The `resource` package ensures that each object is created accordingly.

For example, consider the following `NewRoute` function:

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

This function creates a new `Route` that depends on the `Service` object and its references, labels, and annotations.
