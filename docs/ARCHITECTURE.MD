# Kogito Operator architecture

The Kogito Operator is built on top of the [Operator SDK framework](https://github.com/operator-framework/operator-sdk)
, written in Golang. It’s recommended to read the SDK documentation before going further to understand the concept of controllers, 
operators, custom resources and so on. 
Good starting points are the [Memcached Operator](https://github.com/operator-framework/operator-sdk-samples/tree/master/go/memcached-operator)
example and the [Operator SDK User Guide](https://github.com/operator-framework/operator-sdk/blob/master/doc/user-guide.md).

You should learn how to use the Kubernetes API since most of the Operator’s work relies on 
making successive calls to this API.

## The Operator Goals

The main goal of the Kogito Operator is to deploy a custom Kogito Service, 
the Kogito support services (such as the Data Index), and the infrastructure needed to offer capabilities 
such as persistence, messaging and security. 

At this point we have a few CRDs (Custom Resources Definitions) to achieve the mentioned goals:

- **KogitoApp**: describes a custom Kogito Service build and deployment attributes to enable the operator
 to create a specific build and deployment configuration. 
 An instance of this CRD will build and deploy a given Kogito Service in an OpenShift cluster;
- **KogitoMgmtConsole, KogitoDataIndex, KogitoJobsService**: they are mostly the same thing and share a 
common interface, apart from some specific details such as persistence characteristics. 
They enable the deployment of the Management Console, Data Index and Jobs Service respectively;
- **KogitoInfra**: it’s an internal object managed by the operator and deployed only on scenarios that a 
Kogito Service needs middleware infrastructure to be handled by third party operators. 
It delegates to these operators the deployment of required middleware services, such as Infinispan for persistence 
and Kafka for messaging.

## Operator Dependencies

We depend on three operators to offer developers fast setup on development scenarios, 
so they won’t need to bother deploying their own middleware infrastructure to test and try their 
Kogito Services on an OpenShift cluster.

For example, given this KogitoApp CR:

```yaml
apiVersion: app.kiegroup.org/v1alpha1
kind: KogitoApp
metadata:
  name: kogito-travel-agency
spec:
  enableEvents: true
  enablePersistence: true
  build:
    gitSource:
      contextDir: kogito-travel-agency
      uri: "https://github.com/kiegroup/kogito-examples/"
```

The operator will deploy Kafka through the Strimzi Operator to enable events for this service. 
To enable persistence, the operator will deploy an Infinispan server through the Infinispan Operator.

Those operators are installed automatically by the [OLM (Operator Lifecycle Management)](https://github.com/operator-framework/operator-lifecycle-manager)
once the Kogito Operator is installed in the namespace by reading its [manifest (CSV file)](https://github.com/operator-framework/operator-sdk/blob/master/doc/user/olm-catalog/generating-a-csv.md):

```yaml
required:
    - description: Represents a Infinispan cluster (...)
      displayName: Infinispan Cluster
      kind: Infinispan
      name: infinispans.infinispan.org
      version: v1
    - description: Represents a Kafka cluster
      displayName: Kafka
      kind: Kafka
      name: kafkas.kafka.strimzi.io
      version: v1beta1
    - description: Represents a topic inside a Kafka cluster
      displayName: Kafka Topic
      kind: KafkaTopic
      name: kafkatopics.kafka.strimzi.io
      version: v1beta1
    - description: Represents a Keycloak server to provide SSO (...)      
      displayName: Keycloak
      kind: Keycloak
      name: keycloaks.keycloak.org
      version: v1alpha1
```

If those kinds are not found by the Kogito Operator during runtime, the capability to handle middleware 
infrastructure will be disabled. This means that in environments without OLM or where users 
do not install the operator dependencies, the Kogito Operator will work normally, but the infrastructure 
must be provided by the end user.

Same thing for Kogito Support services such as the Data Index:

```yaml
apiVersion: app.kiegroup.org/v1alpha1
kind: KogitoDataIndex
metadata:
  name: data-index
spec:
  replicas: 1
  kafka:
    useKogitoInfra: true
  infinispan:
    useKogitoInfra: true
```

In cases like this, the Data Index will be deployed after the Kogito Operator deploys the infrastructure 
to support Kafka and Infinispan for the service.

To not use the provided infrastructure (such as on production environments), one can define the CR like this:

```yaml
apiVersion: app.kiegroup.org/v1alpha1
kind: KogitoDataIndex
metadata:
  name: data-index
spec:
  replicas: 1
  kafka:
    useKogitoInfra: true
  infinispan:
    useAuth: true
    credentials:
      secretName: infinispan-credentials
      usernameKey: username
      passwordKey: password
    uri: my-infinispan-server:11222
    saslMechanism: DIGEST-MD5
    useKogitoInfra: false
```

In this case, the operator **will not** try to deploy an Infinispan server, 
but to connect to the provided server `my-infinispan-server` with the given credentials. 

Another option would be setting `useKogitoInfra` to false and providing all the Infinispan connectivity 
properties through the ConfigMap `data-index-properties` created by the operator.

## Packages Structure

To start contributing to the Kogito Operator is important to understand the overall package structure 
and their responsibilities within the project. Take a look at the picture below.

![](docs/img/kogito_operator_packages.png)

The **cmd** package contains the entry point for the Kogito Operator image in the internal **manager** package. 
The **kogito** package is the CLI implementation. In the future this package could be part 
of a separate project, using the Kogito Operator types and infrastructure code as reference.

The **test** package contains the implementation of BDD tests. 
Initially we were relying on the [e2e tests using the Operator SDK](https://github.com/operator-framework/operator-sdk/blob/master/doc/test-framework/writing-e2e-tests.md),
but it came up that reading and understanding those tests was quite difficult, 
as we were performing a succession of many steps within one test. 
We needed a better approach on that side. Also, as the new framework should allow us to test 
full use cases, we needed a more powerful framework due to the complexity of our deployments.

That is why, we decided to implement those tests using Business Driven Development tests with 
the [Godog project](https://github.com/cucumber/godog), now owned by Cucumber teams. 
It allows us to separate the test scenario itself, which is much more readable as it is 
plain-english, and the implementation of the tests.

In those BDD tests, we are testing full built cloud images in an Openshift cluster, 
and not the code itself directly as it was the case for e2e operator framework.

The operator implementation lies within the **pkg** package, where:

- **apis**: defines the CRD (Custom Resource Definition) types for the resources managed 
by the operator. Initially generated by the Operator SDK, these types are the backbone for the 
operator implementation. It’s based on them that all the controller logic is handled;
- **client**: it’s a wrapper for the Kubernetes and Openshift clients. 
Underneath, it uses the same client provided by the Operator SDK. This is the same client used 
by the CLI and the Operator;
- **controller**: business logic for the operator. Initially generated by the Operator SDK and 
expanded to meet our needs. All the logic related to how to react to a change in any resource 
managed by the operator should be handled here;
- **framework**: has common code related to any Kubernetes operator and not tied to a specific 
use case. Later, code in this package could move to the [Operator Utils](https://github.com/RHsyseng/operator-utils)
project. When modifying code in this package, have this in mind;
- **infrastructure**: code related to the Kogito Operator common infrastructure, 
such as injecting external endpoints among the services managed by the operator. 
Code in this package should be shared between all controllers, otherwise keep it on the 
controller package itself;
- **logger**: log implementation shared among all other packages. 
It’s a wrapper for the [Uber Zap log library](https://github.com/uber-go/zap);
- **util**: common Go utilities used across the project. 

## Future

There are some key topics that we are actively working on to improve the Kogito Operator use case 
that should be the focus for the next versions.

### Kubernetes Support

One of the key missing features for the Kogito Operator is the lack of Kubernetes (vanilla) support. 
At the time of this writing, only OpenShift is supportable mainly because of:

1. [S2i (source to image)](https://docs.openshift.com/container-platform/4.3/builds/build-strategies.html#build-strategy-s2i_build-strategies) 
dependency to build a custom Kogito Service image and deploy it;
2. [Routes](https://docs.openshift.com/container-platform/4.3/networking/routes/route-configuration.html)
dependency to expose Kogito services endpoints;

For the s2i lack of support on Kubernetes, we came up with two possibilities:

1. Use [Tekton Operator](https://github.com/tektoncd/operator) to build the images instead of using OpenShift BuildConfig resources
2. Build the images locally via Dockerfile or [s2i client](https://github.com/openshift/source-to-image)

Regarding the Routes dependency, the alternative is to use the [Kubernetes Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/)
object instead. Although this will require a small change in the Kogito CRs to handle details 
of the ingress such as the host name.

Despite those two dependencies the Kogito Operator code is already safe to run outside of an 
OpenShift instance, since every API call to it has some kind of guard code to avoid runtime errors.

This feature will be handled by the JIRA [KOGITO-952](https://issues.redhat.com/browse/KOGITO-952).

### Building

At the time of writing we still have some challenges regarding the custom Kogito Service image 
builds that we need to overcome. Each one of them will be detailed in the next sections.

#### S2i builds running on OpenShift takes too much time

Taking the code from a Git repository and building it on an OpenShift cluster takes a lot of time 
mainly because all dependencies have to be downloaded from the Maven public repository.

To overcome this problem, the recommendation is to have a Maven cache available such as a Nexus 
server. The problem with this approach is that developers should set up another infrastructure 
component to speed up the build process, which could be cumbersome and resource hungry.

Another approach is something that we are working on ([KOGITO-602](https://issues.redhat.com/browse/KOGITO-602)),
which is to provide the Kogito dependencies libraries that Red Hat already distribute within the 
s2i images, like Quarkus. This won’t require a Nexus server ready and will speed up the build 
process by nearly 70%. The problem with this solution is that the build images will be 
considerably large and Spring Boot based services won’t gain much since we can’t redistribute 
their libraries with our images.

Since native builds are also supported by Kogito Services based on Quarkus, when building such 
services on OpenShift, it will require a sizable amount of memory and CPU resources. 
In local development scenarios this could be nearly impossible to achieve, since in our tests a 
small [example can take up to 10 GB of memory to build](https://github.com/kiegroup/kogito-cloud-operator/#native-x-jvm-builds). 

#### Local Builds

A more affordable solution in development scenarios is for the developer to provide their already
built artifacts. To support such scenarios, we created the Binary Build solution, which is a 
specially configured image and OpenShift’s build configuration that takes a local built artifact
and uploads it to the cluster. For more information, please refer see the [Operator Documentation](https://github.com/kiegroup/kogito-cloud-operator/#binary-builds). 

Since we are aiming for Kubernetes support, Binary Build is not a supported solution since 
`BuildConfig` is an OpenShift exclusive object. In such scenarios, not only the Kogito service 
binary needs to be provided by the developer/administrator, but the entire image.

One approach is to have a template Dockerfile being generated for the developer once they create 
a Kogito project via the Maven plugin. This Dockerfile will be based on the Kogito UBI image with 
a new layer to copy the generated project’s artifact and library from the target dir.

Another way of achieving local builds is to use the s2i client instead of a Dockerfile. 
Since building and maintaining Dockerfiles is a well-known approach, it’s preferable over the 
s2i client.

Since we have the “Kogito CLI” tool that facilitates the integration with the operator, 
one possible solution is to incorporate the local image build to it. This way, the full circle 
could be achieved by the tool:

1. Kogito CLI Generates the Kogito project’s skeleton
2. A developer works on their project, creating process, rules, decisions, and so on
3. Kogito CLI creates a new Kogito project on a Kubernetes cluster
4. Kogito CLI pushes the image to a registry and deploys the Kogito Service into the cluster

#### Pipeline Builds

For production usage, would be worth it to also support pipeline builds. In this case, 
the workflow would be fairly simple and much the same of what we would have for local builds. 
From the Kogito Operator perspective, it doesn't matter where the image is since its reachable. 

## Operator Dependencies

Like stated earlier in this document, the Kogito Operator depends on Infinispan, Strimzi and 
Keycloak to provide required infrastructure for Kogito Services.

Since there’s a quite complex architecture that needs to be provided to run a Kogito service with 
persistence, messaging and security, having the operator to handle it comes in handy during 
development, but could be a problem in production environments since load, resources, cross 
dependencies and so on could be quite different.

That said, it is not recommended having the Kogito Operator to handle the entire infrastructure 
architecture in production environments. At the time of writing, we don’t have a way to declare 
a “development” or “production” mode to the operator. Thus, if an administrator creates a Kogito 
Data Index setting `useKogitoInfra` to `true` in a production namespace, the operator would deploy
both Infinispan and Strimzi to enable those services to the Data Index.

In the near future, the operator should be aware of the environment to handle those properties accordingly.
