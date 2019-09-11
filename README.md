# Kogito Operator

[![Go Report Card](https://goreportcard.com/badge/github.com/kiegroup/kogito-cloud-operator)](https://goreportcard.com/report/github.com/kiegroup/kogito-cloud-operator) [![CircleCI](https://circleci.com/gh/kiegroup/kogito-cloud-operator.svg?style=svg)](https://circleci.com/gh/kiegroup/kogito-cloud-operator)

Kogito Operator was designed to deploy [Kogito Runtimes](https://github.com/kiegroup/kogito-runtimes) services from source and every piece of infrastructure that the services might need like SSO ([Keycloak](https://github.com/integr8ly/keycloak-operator)) and Persistence ([Infinispan](https://github.com/infinispan/infinispan-operator)).

## Requirements

- go v1.12+
- [operator-sdk](https://github.com/operator-framework/operator-sdk/releases) v0.9.0
- ocp 4.x (you can use [CRC](https://github.com/code-ready/crc) for local deployment)
- [kogito s2i imagestreams](https://raw.githubusercontent.com/kiegroup/kogito-cloud/master/s2i/kogito-imagestream.yaml) installed

## Build

```bash
$ make
```

## Installation

Kogito Operator is not available in the OperatorHub [yet](https://issues.jboss.org/browse/KOGITO-67), hence has to be installed manually on [OpenShift 4.x](#deploy-to-openshift-4x) or [OpenShift 3.11](#deploy-to-openshift-311-manually).

You can also [run the operator locally](#running-locally) if you have the [requirements](#requirements) configured in your local machine.

### Deploy to OpenShift 4.x

First make sure that the Kogito image stream is created in the cluster:

```bash
$ oc apply -f https://raw.githubusercontent.com/kiegroup/kogito-cloud/master/s2i/kogito-imagestream.yaml -n openshift
```

Then create an entry in the OperatorHub catalog with:

```bash
$ oc create -f deploy/catalog_resources/courier/kogitocloud-operatorsource.yaml
```

It will take a few minutes for the operator to become visible under the _OperatorHub_ section of the OpenShift console _Catalog_. The Operator can be easily found by filtering by _Kogito_ name.

Verify operator availability by running:

```bash
$ oc describe operatorsource.operators.coreos.com/kogitocloud-operator -n openshift-marketplace
```

### Deploy to OpenShift 3.11 manually

```bash
## kogito imagestreams should already be installed/available ... e.g.
$ oc apply -f https://raw.githubusercontent.com/kiegroup/kogito-cloud/master/s2i/kogito-imagestream.yaml -n openshift
$ oc new-project <project-name>
$ ./hack/3.11deploy.sh
```

### Trigger a KogitoApp deployment

Use the OLM console to subscribe to the `kogito` Operator Catalog Source within your namespace. Once subscribed, use the console to `Create KogitoApp` or create one manually as seen below.

```bash
$ oc create -f deploy/crs/app_v1alpha1_kogitoapp_cr.yaml
kogitoapp.app.kiegroup.org/example-quarkus created
```

Or you can always use the [CLI](#kogito-cli) to deploy your services.

To look at the Operator logs, first identify where the Operator is deployed:

```bash
$ oc get pods

NAME                                     READY   STATUS      RESTARTS   AGE
kogito-cloud-operator-6d7b6d4466-9ng8t   1/1     Running     0          26m
```

Use the pod name as the input of the following command:

```bash
$ oc logs -f kogito-cloud-operator-6d7b6d4466-9ng8t
```

### Clean up a KogitoApp deployment

```bash
$ oc delete kogitoapp example-quarkus
```

### Deploy Data Index Service

The Kogito Operator is able to deploy the [Data Index Service](https://github.com/kiegroup/kogito-runtimes/wiki/Data-Index-Service) as a [Custom Resource](deploy/crds/app_v1alpha1_kogitodataindex_cr.yaml) (`KogitoDataIndex`). Since Data Index Service depends on Kafka and Infinispan, it's necessary to manually deploy an Apache Kafka Cluster and an Infinispan Server (10.x) in the same namespace.

| :information_source: It's planned for future releases that the Kogito Operator will deploy an Infinispan and a Kafka cluster when deploying the Data Index Service. |
| --- |

#### Deploy Infinispan

To deploy an Infinispan Server, you can leverage from `oc new-app [docker image]` command as follows:

```bash
$ oc new-app jboss/infinispan-server:10.0.0.Beta3
```

Expect a similar output like this one:

```bash
--> Found Docker image caaa296 (5 months old) from Docker Hub for "jboss/infinispan-server:10.0.0.Beta3"

    Infinispan Server 
    ----------------- 
    Provides a scalable in-memory distributed database designed for fast access to large volumes of data.

    Tags: datagrid, java, jboss

    * An image stream tag will be created as "infinispan-server:10.0.0.Beta3" that will track this image
    * This image will be deployed in deployment config "infinispan-server"
    * Ports 11211/tcp, 11222/tcp, 57600/tcp, 7600/tcp, 8080/tcp, 8181/tcp, 8888/tcp, 9990/tcp will be load balanced by service "infinispan-server"
      * Other containers can access this service through the hostname "infinispan-server"

--> Creating resources ...
    imagestream.image.openshift.io "infinispan-server" created
    deploymentconfig.apps.openshift.io "infinispan-server" created
    service "infinispan-server" created
--> Success
    Application is not exposed. You can expose services to the outside world by executing one or more of the commands below:
     'oc expose svc/infinispan-server' 
    Run 'oc status' to view your app.
```

OpenShift will create everything you need for Infinispan Server to work in the namespace. Make sure that the pod is running:

```bash
$ oc get pods -l app=infinispan-server
```

Take a look at the logs by running:

```bash
# take the pod name from the command you ran before
$ oc logs -f <pod name>
```

The Infinispan server should be accessed within the namespace by port 11222:

```bash
$ oc get svc -l app=infinispan-server

NAME                TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                                                                      AGE
infinispan-server   ClusterIP   172.30.193.214   <none>        7600/TCP,8080/TCP,8181/TCP,8888/TCP,9990/TCP,11211/TCP,11222/TCP,57600/TCP   4m19s
```

#### Deploy Strimzi

Deploying [Strimzi](https://strimzi.io/) is much easier since it's an Operator and should be available in the [OperatorHub](https://operatorhub.io/operator/strimzi-kafka-operator). On OpenShift Web Console, go to the left menu, Catalog, OperatorHub and search for `Strimzi`.

Follow the on screen instructions to install the Strimzi Operator. At the end, you should see the Strimzi Operator on the Installed Operators tab:

![Strimzi Installed](docs/img/strimzi_installed.png?raw=true)

Next, you need to create a Kafka cluster and a Kafka Topic for the Data Index Service to connect. Click on the name of the Strimzi Operator, then on `Kafka` tab and `Create Kafka`. Accept the default options to create a 3 node Kafka cluster. If it's a development environment, consider setting the Zookeeper and Kafka replicas to 1 to save resources.

After a few minutes you should see the pods running and the services available:

```bash
$ oc get svc -l strimzi.io/cluster=my-cluster

NAME                          TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                      AGE
my-cluster-kafka-bootstrap    ClusterIP   172.30.228.90    <none>        9091/TCP,9092/TCP,9093/TCP   9d
my-cluster-kafka-brokers      ClusterIP   None             <none>        9091/TCP,9092/TCP,9093/TCP   9d
my-cluster-zookeeper-client   ClusterIP   172.30.241.146   <none>        2181/TCP                     9d
my-cluster-zookeeper-nodes    ClusterIP   None             <none>        2181/TCP,2888/TCP,3888/TCP   9d
```

The service you're interested in is `my-cluster-kafka-bootstrap:9092`. We will use it to deploy the Data Index Service later.

Having the cluster up and running, the next step is creating a `Kafka Topic` required by the Data Index Service.

In the OpenShift Web Console, go to the Installed Operators, Strimzi Operator, Kafka Topic tab. From there, create a new `Kafka Topic` and name it as `kogito-processinstances-events` like in the example below:

```yaml
apiVersion: kafka.strimzi.io/v1beta1
kind: KafkaTopic
metadata:
  name: kogito-processinstances-events
  labels:
    strimzi.io/cluster: my-cluster
  namespace: kogito
spec:
  partitions: 10
  replicas: 3
  config:
    retention.ms: 604800000
    segment.bytes: 1073741824
```

To check if everything was created successfully run the following command:

```bash
$ oc describe kafkatopic/kogito-processinstances-events

Name:         kogito-processinstances-events
Namespace:    kogito
Labels:       strimzi.io/cluster=my-cluster
Annotations:  <none>
API Version:  kafka.strimzi.io/v1beta1
Kind:         KafkaTopic
Metadata:
  Creation Timestamp:  2019-08-28T18:09:41Z
  Generation:          2
  Resource Version:    5673235
  Self Link:           /apis/kafka.strimzi.io/v1beta1/namespaces/kogito/kafkatopics/kogito-processinstances-events
  UID:                 0194989e-c9bf-11e9-8160-0615e4bfa428
Spec:
  Config:
    message.format.version:  2.3-IV1
    retention.ms:            604800000
    segment.bytes:           1073741824
  Partitions:                10
  Replicas:                  1
  Topic Name:                kogito-processinstances-events
Events:                      <none>
```

Now that you have the required infrastrucuture, it's safe to deploy the Kogito Data Index Service.

#### Deploy Data Index

Having [installed](#installation) the Kogito Operator, create a new `Kogito Data Index` resource using the services URIs from Infinispan and Kafka:

```bash
$ oc get svc -l app=infinispan-server

NAME                TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                                                                      AGE
infinispan-server   ClusterIP   172.30.193.214   <none>        7600/TCP,8080/TCP,8181/TCP,8888/TCP,9990/TCP,11211/TCP,11222/TCP,57600/TCP   4m19s
```

In this example, the Infinispan Server service is `infinispan-server:11222`.

Then grab the Kafka cluster URI:

```bash
$ oc get svc -l strimzi.io/cluster=my-cluster | grep bootstrap

my-cluster-kafka-bootstrap    ClusterIP   172.30.228.90    <none>        9091/TCP,9092/TCP,9093/TCP   9d
```

In this case the Kafka Cluster service is `my-cluster-kafka-bootstrap:9092`.

Use this information to create the Kogito Data Index resource. 

##### Deploy Data Index with Kogito CLI

If you have installed the [Kogito CLI](#kogito-cli), you can simply run:

```bash
$ kogito deploy-data-index -p my-project --infinispan-url infinispan-server:11222 --kafka-url my-cluster-kafka-bootstrap:9092
```

##### Deploy Data Index with Operator Catalog (OLM)

If you're running on OCP 4.x, you might use the OperatorHub user interface. In the left menu go to Installed Operators, Kogito Operator, Kogito Data Index tab. From there, click on "Create Kogito Data Index" and create a new resource like in the example below using the Infinispan and Kafka services:

```yaml
apiVersion: app.kiegroup.org/v1alpha1
kind: KogitoDataIndex
metadata:
  name: kogito-data-index
spec:
  # If not informed, these default values will be set for you
  name: "kogito-data-index"
  # environment variables to set in the runtime container. Example: JAVAOPTS: "-Dquarkus.log.level=DEBUG"
  env: {}
  # number of pods to be deployed
  replicas: 1
  # image to use for this deploy
  image: "quay.io/kiegroup/kogito-data-index:latest"
  # Limits and requests for the Data Index pod
  memoryLimit: ""
  memoryRequest: ""
  cpuLimit: ""
  cpuRequest: ""
  # details about the kafka connection
  kafka:
    # the service name and port for the kafka cluster. Example: my-kafka-cluster:9092
    serviceURI: my-cluster-kafka-bootstrap:9092
  # details about the connected infinispan
  infinispan:
    # the service name and port of the infinispan cluster. Example: my-infinispan:11222
    serviceURI: infinispan-server:11222
```

##### Deploy Data Index with oc client

You can use the CR file showed above as a reference and create the custom resource from the command line:

```bash
# clone this repo
$ git clone https://github.com/kiegroup/kogito-cloud-operator.git
$ cd kogito-cloud-operator
# make your changes
$ vi deploy/crds/app_v1alpha1_kogitodataindex_cr.yaml
# deploy to the cluster
$ oc create -f deploy/crds/app_v1alpha1_kogitodataindex_cr.yaml -n my-project
```

You should be able to access the GraphQL interface via the route created for you:

```bash
$ oc get routes -l app=kogito-data-index

NAME                HOST/PORT                                                                      PATH   SERVICES            PORT   TERMINATION   WILDCARD
kogito-data-index   kogito-data-index-kogito.apps.mycluster.example.com                                   kogito-data-index   8180                 None
```

## Kogito CLI

A CLI tool is available to make it easy to deploy new Kogito Services from source instead of relying on CRs yaml files.

### CLI Requirements

1. [`oc` client](https://docs.okd.io/latest/cli_reference/get_started_cli.html) installed
2. An authenticated OpenShift user with permissions to create resources in a given namespace
3. Kogito Operator [installed](#installation) in the cluster

### CLI Install

#### For Linux

1. [Download](https://github.com/kiegroup/kogito-cloud-operator/releases) the correct distribution for your machine

2. Unpack the binary: `tar -xvf release.tar.gz`

3. You should see an executable named `kogito`. Move the binary to a pre-existing directory in your `PATH`. For example: `# cp /path/to/kogito /usr/local/bin`

#### For Windows

Just download the [latest 64-bit Windows release](https://github.com/kiegroup/kogito-cloud-operator/releases). Extract the zip file through a file browser. Add the extracted directory to your PATH. You can now use `kogito` from the command line.

### Build CLI from source

| :warning: Building CLI from source requires that [go is installed](https://golang.org/doc/install) and available in your `PATH`. |
| --- |

Build and install the CLI by running:

```bash
$ git clone https://github.com/kiegroup/kogito-cloud-operator
$ cd kogito-cloud-operator
$ make install-cli
```

The `kogito` CLI should be available in your path:

```bash
$ kogito
Kogito CLI deploys your Kogito Services into an OpenShift cluster

Usage:
  kogito [command]

Available Commands:
  delete-project    Deletes a Kogito Project - i.e., the Kubernetes/OpenShift namespace
  delete-service    Deletes a Kogito Runtime Service deployed in the namespace/project
  deploy-data-index Deploys the Kogito Data Index Service in the given Project
  deploy-service    Deploys a new Kogito Runtime Service into the given Project
  help              Help about any command
  new-project       Creates a new Kogito Project for your Kogito Services
  use-project       Sets the Kogito Project where your Kogito Service will be deployed
  version           Prints the kogito CLI version

Flags:
      --config string   config file (default is $HOME/.kogito.json)
  -h, --help            help for kogito
  -v, --verbose         verbose output

Use "kogito [command] --help" for more information about a command.
```

### Deploy a Kogito Service from source with CLI

After [installing](#installation) the Kogito Operator, it's possible to deploy a new Kogito Service by using the CLI:

```bash
# creates a new namespace in your cluster
$ kogito new-project kogito-cli

# deploys a new Kogito Runtime Service from source
$ kogito deploy-service example-drools https://github.com/kiegroup/kogito-examples --context-dir drools-quarkus-example
```

If you are using OpenShift 3.11 as described in the [previous chapter](#deploy-to-openshift-311-manually), you shall use the existing namespace you created during the manual deployment, by using the following CLI commands:

```bash
# use the provisioned namespace in your OpenShift 3.11 cluster
$ kogito use-project <project-name>

# deploys a new kogito service from source
$ kogito deploy-service example-drools https://github.com/kiegroup/kogito-examples --context-dir drools-quarkus-example
```

This can be shorten to:

```bash
$ kogito deploy-service example-drools https://github.com/kiegroup/kogito-examples --context-dir drools-quarkus-example --project <project-name>
```

## Development

While fixing issues or adding new features to the Kogito Operator, please consider taking a look at [Contributions](docs/CONTRIBUTING.MD) and [Architecture](docs/ARCHITECTURE.MD) documentation.

### Deploy to OpenShift 4.x for development purposes

To install this operator on OpenShift 4 for end-to-end testing, make sure you have access to a quay.io account to create an application repository. Follow the [authentication](https://github.com/operator-framework/operator-courier/#authentication) instructions for Operator Courier to obtain an account token. This token is in the form of "basic XXXXXXXXX" and both words are required for the command.

Push the operator bundle to your quay application repository as follows:

```bash
$ operator-courier push deploy/catalog_resources/courier/0.4.0 namespace kogitocloud-operator 0.4.0 "basic XXXXXXXXX"
```

If pushing to another quay repository, replace _namespace_ with your username or other namespace. Notice that the push command does not overwrite an existing repository, and the bundle needs to be deleted before a new version can be built and uploaded. Once the bundle has been uploaded, create an [Operator Source](https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md#linking-the-quay-application-repository-to-your-openshift-40-cluster) to load your operator bundle in OpenShift.

Note that the OpenShift cluster needs access to the created application. Make sure that the application is **public** or you have configured the private repository credentials in the cluster. To make the application public, go to your quay.io account, and in the _Applications_ tab look for the `kogitocloud-operator` application. Under the settings section click on _make public_ button.

```bash
## kogito imagestreams should already be installed/available ... e.g.
$ oc apply -f https://raw.githubusercontent.com/kiegroup/kogito-cloud/master/s2i/kogito-imagestream.yaml -n openshift
$ oc create -f deploy/catalog_resources/courier/kogitocloud-operatorsource.yaml
```

Remember to replace _registryNamespace_ in the `kogitocloud-operatorsource.yaml` with your quay namespace. The name, display name and publisher of the operator are the only other attributes that may be modified.

It will take a few minutes for the operator to become visible under the _OperatorHub_ section of the OpenShift console _Catalog_. The Operator can be easily found by filtering the provider type to _Custom_.

It's possible to verify the operator status by running:

```bash
$ oc describe operatorsource.operators.coreos.com/kogitocloud-operator -n openshift-marketplace
```

### Running Locally

Change log level at runtime with the `DEBUG` environment variable. e.g. -

```bash
$ make mod
$ make clean
$ DEBUG="true" operator-sdk up local --namespace=<namespace>
```

Before submitting PR, please be sure to read the [contributors guide](docs/CONTRIBUTING.MD).

It's always worth noting that you should generate, vet, format, lint, and test your code. This all can be done with one command.

```bash
$ make test
```

## Contributing

Please take a look at the [Contributing to Kogito Operator](docs/CONTRIBUTING.MD) guide.
