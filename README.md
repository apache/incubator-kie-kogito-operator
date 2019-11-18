# Kogito Operator

[![Go Report Card](https://goreportcard.com/badge/github.com/kiegroup/kogito-cloud-operator)](https://goreportcard.com/report/github.com/kiegroup/kogito-cloud-operator) [![CircleCI](https://circleci.com/gh/kiegroup/kogito-cloud-operator.svg?style=svg)](https://circleci.com/gh/kiegroup/kogito-cloud-operator)

Kogito Operator was designed to deploy [Kogito Runtimes](https://github.com/kiegroup/kogito-runtimes) services from source and every piece of infrastructure that the services might need, such as SSO ([Keycloak](https://github.com/integr8ly/keycloak-operator)) and Persistence ([Infinispan](https://github.com/infinispan/infinispan-operator)).

Table of Contents
=================

   * [Kogito Operator](#kogito-operator)
   * [Table of Contents](#table-of-contents)
      * [Requirements](#requirements)
      * [Installation](#installation)
         * [Deploy to OpenShift 4.x](#deploy-to-openshift-4x)
            * [Optional Step](#optional-step)
            * [Via OperatorHub automatically](#via-operatorhub-automatically)
            * [Via OperatorHub manually](#via-operatorhub-manually)
            * [Via local Operator](#via-local-operator)
            * [Check Operator's availability](#check-operators-availability)
         * [Deploy to OpenShift 3.11](#deploy-to-openshift-311)
      * [Trigger a Kogito Runtime Service deployment](#trigger-a-kogito-runtime-service-deployment)
         * [Deploy a new service](#deploy-a-new-service)
         * [Clean up a Kogito Service deployment](#clean-up-a-kogito-service-deployment)
         * [Native X JVM Builds](#native-x-jvm-builds)
         * [Troubleshooting](#troubleshooting)
            * [I'm not seeing any builds running!](#im-not-seeing-any-builds-running)
            * [Native build fails with "The build pod was killed due to an out of memory condition."](#native-build-fails-with-the-build-pod-was-killed-due-to-an-out-of-memory-condition)
      * [Deploy Data Index Service](#deploy-data-index-service)
         * [Deploy Infinispan](#deploy-infinispan)
         * [Deploy Strimzi](#deploy-strimzi)
            * [Create Required Topics](#create-required-topics)
         * [Install Data Index Service](#install-data-index-service)
            * [Retrieve Kafka internal url](#retrieve-kafka-internal-url)
            * [Install Data Index Service with Kogito CLI](#install-data-index-service-with-kogito-cli)
            * [Install Data Index Service with Operator Catalog (OLM)](#install-data-index-service-with-operator-catalog-olm)
            * [Install Data Index Service with oc client](#install-data-index-service-with-oc-client)
      * [Kogito CLI](#kogito-cli)
         * [CLI Requirements](#cli-requirements)
         * [CLI Install](#cli-install)
            * [For Linux](#for-linux)
            * [For Windows](#for-windows)
            * [Build CLI from source](#build-cli-from-source)
         * [Deploy a Kogito Service from source with CLI](#deploy-a-kogito-service-from-source-with-cli)
      * [Development](#development)
         * [Build Operator](#build-operator)
         * [Deploy to OpenShift 4.x for development purposes](#deploy-to-openshift-4x-for-development-purposes)
         * [Running End to End tests](#running-end-to-end-tests)
            * [Operator SDK](#operator-sdk)
            * [Kogito CLI](#kogito-cli-1)
         * [Running Locally](#running-locally)
      * [Prometheus Integration](#prometheus-integration)
         * [Prometheus Annotations](#prometheus-annotations)
         * [Prometheus Operator](#prometheus-operator)
      * [Infinispan Integration](#infinispan-integration)
      * [Contributing](#contributing)

Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc)

## Requirements

- go v1.12
- [operator-sdk](https://github.com/operator-framework/operator-sdk/releases) v0.11.0
- ocp 3.11/4.x (you can use [CRC](https://github.com/code-ready/crc) for local deployment)
- [kogito s2i imagestreams](https://raw.githubusercontent.com/kiegroup/kogito-cloud/master/s2i/kogito-imagestream.yaml) installed

## Installation

### Deploy to OpenShift 4.x

Kogito operator is a namespaced operator, which means that you will need to install it into the namespace you want your Kogito application to run in.

#### Optional Step

You can import the Kogito image stream using the `oc client` manually with the following command:

```bash
$ oc apply -f https://raw.githubusercontent.com/kiegroup/kogito-cloud/master/s2i/kogito-imagestream.yaml -n openshift
```

But, this step is not mandatory anymore, when kogito is going to install a new app, it will create the required imagestreams
before.

#### Via OperatorHub automatically

The installation on OpenShift 4.x is pretty straightforward since Kogito Operator is available in the OperatorHub as a community operator. The Operator can be easily found by filtering by _Kogito_ name.

Just follow the OpenShift Web Console instructions in the _Catalog_, _OperatorHub_ section in the left menu to install it in any namespace in the cluster. 

![Kogito Operator in the Catalog](docs/img/catalog-kogito-operator.png?raw=true)

#### Via OperatorHub manually

If it does not appear, you can still install it manually, by creating an entry in the OperatorHub Catalog:

```bash
$ oc create -f deploy/olm-catalog/kogito-cloud-operator/kogitocloud-operatorsource.yaml
```

It will take a few minutes for the operator to become visible under the _OperatorHub_ section of the OpenShift console _Catalog_. The Operator can be easily found by filtering by _Kogito_ name.

Then, you can install as decribed in [the automatic process](#via-operatorhub-automatically).

#### Via local Operator

Or you can also [run the operator locally](#running-locally) if you have the [requirements](#requirements) configured in your local machine.

#### Check Operator's availability

You can verify the operator's availability in catalog by running:

```bash
$ oc describe operatorsource.operators.coreos.com/kogitocloud-operator -n openshift-marketplace
```

### Deploy to OpenShift 3.11

Installation on OpenShift 3.11 has to be done manually since the OperatorHub catalog is not available by default:

```bash
## kogito imagestreams should already be installed/available ... e.g.
$ oc apply -f https://raw.githubusercontent.com/kiegroup/kogito-cloud/master/s2i/kogito-imagestream.yaml -n openshift
$ oc new-project <project-name>
$ ./hack/3.11deploy.sh
```

## Trigger a Kogito Runtime Service deployment

### Deploy a new service

Use the OLM console to subscribe to the `kogito` Operator Catalog Source within your namespace. Once subscribed, use the console to `Create KogitoApp` or create one manually as seen below.

```bash
$ oc create -f deploy/crds/app.kiegroup.org_v1alpha1_kogitoapp_cr.yaml
kogitoapp.app.kiegroup.org/example-quarkus created
```

Alternatively, you can use the [CLI](#kogito-cli) to deploy your services: 

```bash
$ kogito deploy-service example-quarkus https://github.com/kiegroup/kogito-examples/ --context-dir=drools-quarkus-example
```

### Clean up a Kogito Service deployment

```bash
$ kogito delete-service example-quarkus
```

### Native X JVM Builds

By default, the Kogito Services will be built with traditional `java` compilers to speed up the time and save resources. 

This means that the final generated artifact will be a jar file with the chosen runtime (default to Quarkus) with its dependencies in the image user's home dir `/home/kogito/bin/lib`.

Kogito Services when implemented with [Quarkus](https://quarkus.io/guides/kogito-guide) can be built to native binary. This means low ([really low](https://www.graalvm.org/docs/examples/java-performance-examples/)) footprint on runtime, but will demand a lot of resources during build time. Read more about AOT compilation [here](https://www.graalvm.org/docs/reference-manual/aot-compilation/).

In our tests, native builds takes approximately 10 minutes and the build pod can consume up to 10GB of RAM and 1.5 CPU cores. Make sure that you have this resources available when running native builds.

To deploy a service using native builds, run the `deploy-service` command with `--native` flag:

```bash
$ kogito deploy-service example-quarkus https://github.com/kiegroup/kogito-examples/ --context-dir=drools-quarkus-example --native
```

### Troubleshooting 

#### I'm not seeing any builds running!

If you don't see any builds running nor any resources created in the namespace, try to take a look at the Kogito Operator log.

To look at the operator logs, first identify where the operator is deployed:

```bash
$ oc get pods

NAME                                     READY   STATUS      RESTARTS   AGE
kogito-cloud-operator-6d7b6d4466-9ng8t   1/1     Running     0          26m
```

Use the pod name as the input of the following command:

```bash
$ oc logs -f kogito-cloud-operator-6d7b6d4466-9ng8t
```

#### Native build fails with "The build pod was killed due to an out of memory condition."

By default, the operator will set the limit in the s2i build pod to 10GB of memory and 1 CPU. It might happen that your Kogito service would require even more resources from the cluster. 

To increase this limit you can edit the `kogitoApp` custom resource in the `build` section:

```yaml
(...)
spec:
  build:
    gitSource:
      contextDir: onboarding-example/onboarding
      uri: 'https://github.com/kiegroup/kogito-examples'
    native: true
    resources:
      limits:
        - resource: cpu
          value: '2'
        - resource: memory
          value: 12Gi
(...)
``` 

Using the CLI, specify how much of memory and CPU your service will use:

```bash
kogito deploy onboarding-service https://github.com/kiegroup/kogito-examples --context-dir onboarding-example/onboarding --native --build-limits memory=12Gi cpu=2  
```

For more information, see [Native X JVM Builds](#native-x-jvm-builds).

## Deploy Data Index Service

The Kogito Operator is able to deploy the [Data Index Service](https://github.com/kiegroup/kogito-runtimes/wiki/Data-Index-Service) as a [Custom Resource](deploy/crds/app.kiegroup.org_v1alpha1_kogitodataindex_cr.yaml) (`KogitoDataIndex`). 

Since Data Index Service depends on Kafka, it's necessary to manually deploy an Apache Kafka Cluster (we use [Strimzi](https://strimzi.io/)) in the same namespace.
 
Data Index also depends on Infinispan, but since 0.6.0 version, Infinispan Server is *automagically* deployed for you.

| :information_source: It's planned for future releases that the Kogito Operator will deploy a Kafka cluster when deploying the Data Index Service. |
| --- |

### Deploy Infinispan

If you plan to have the Data Index to connect to the Infinispan server deployed within the same namespace, Kogito Operator can handle this deployment for you.

When installing the Kogito Operator from OperatorHub, the Infinispan Operator will be available in the same namespace. Otherwise, if you don't have OperatorHub or OLM available in your cluster, you'll have to [subscribe to Infinispan Operator](https://infinispan.org/infinispan-operator/master/documentation/asciidoc/titles/operator.html#deploying_operator_manually) manually.

Having the Infinispan Operator deployed, just jump to [Install Data Index Service section](#install-data-index-service).

### Deploy Strimzi

Deploying [Strimzi](https://strimzi.io/) is easy since it's an Operator and should be available in the [OperatorHub](https://operatorhub.io/operator/strimzi-kafka-operator). On OpenShift Web Console, go to the left menu, Catalog, OperatorHub and search for `Strimzi`.

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

#### Create Required Topics

Having the cluster up and running, the next step is to create the `Kafka Topic`s required by the Data Index Service: 

- `kogito-processinstances-events`
- `kogito-usertaskinstances-events` 
- `kogito-processdomain-events`
- `kogito-usertaskdomain-events`

To create the required topics, in the OpenShift Web Console go to the Installed Operators, Strimzi Operator, Kafka Topic tab. 

From there, create a new `Kafka Topic` and name it as `kogito-processinstances-events` like in the example below:

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
And then do the same for the others topics listed above.

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

Now that you have the required infrastructure, it's safe to deploy the Kogito Data Index Service.

### Install Data Index Service

#### Retrieve Kafka internal url

Having configured the Kafka Operator, you'll need the Kafka service URI to install the Data Index Service. 

Run the following command to take the internal URI:

```bash
$ oc get svc -l strimzi.io/cluster=my-cluster | grep bootstrap

my-cluster-kafka-bootstrap    ClusterIP   172.30.228.90    <none>        9091/TCP,9092/TCP,9093/TCP   9d
```

In this case the Kafka Cluster service is `my-cluster-kafka-bootstrap:9092`.

Use this information to create the Kogito Data Index resource. 

#### Install Data Index Service with Kogito CLI

If you have installed the [Kogito CLI](#kogito-cli), you can simply run:

```bash
$ kogito install data-index -p my-project --kafka-url my-cluster-kafka-bootstrap:9092
```

Infinispan will be deployed for you using the Infinispan Operator. Just make sure that it's running in your project. Case not, an error message will be displayed:

```
Infinispan Operator is not available in the Project: my-project. Please make sure to install it before deploying Data Index without infinispan-url provided 
```

#### Install Data Index Service with Operator Catalog (OLM)

If you're running on OCP 4.x, you might use the OperatorHub user interface. In the left menu go to Installed Operators, Kogito Operator, Kogito Data Index tab. From there, click on "Create Kogito Data Index" and create a new resource like in the example below using the Infinispan and Kafka services:

```yaml
apiVersion: app.kiegroup.org/v1alpha1
kind: KogitoDataIndex
metadata:
  name: kogito-data-index
spec:
  # number of pods to be deployed
  replicas: 1
  # image to use for this deploy
  image: "quay.io/kiegroup/kogito-data-index:latest"
  kafka:
    # the service name and port for the kafka cluster. Example: my-kafka-cluster:9092
    serviceURI: my-cluster-kafka-bootstrap:9092
```

#### Install Data Index Service with oc client

You can use the CR file showed above as a reference and create the custom resource from the command line:

```bash
# clone this repo
$ git clone https://github.com/kiegroup/kogito-cloud-operator.git
$ cd kogito-cloud-operator
# make your changes
$ vi deploy/crds/app.kiegroup.org_v1alpha1_kogitodataindex_cr.yaml
# deploy to the cluster
$ oc create -f deploy/crds/app.kiegroup.org_v1alpha1_kogitodataindex_cr.yaml -n my-project
```

You should be able to access the GraphQL interface via the route created for you:

```bash
$ oc get routes -l app=kogito-data-index

NAME                HOST/PORT                                                                      PATH   SERVICES            PORT   TERMINATION   WILDCARD
kogito-data-index   kogito-data-index-kogito.apps.mycluster.example.com                                   kogito-data-index   8080                 None
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

#### Build CLI from source

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
  delete-project Deletes a Kogito Project - i.e., the Kubernetes/OpenShift namespace
  delete-service Deletes a Kogito Runtime Service deployed in the namespace/project
  deploy-service Deploys a new Kogito Runtime Service into the given Project
  help           Help about any command
  install        Install all sort of infrastructure components to your Kogito project
  new-project    Creates a new Kogito Project for your Kogito Services
  use-project    Sets the Kogito Project where your Kogito Service will be deployed

Flags:
      --config string   config file (default is $HOME/.kogito.json)
  -h, --help            help for kogito
  -v, --verbose         verbose output
      --version         display version

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

If you are using OpenShift 3.11 as described in the [previous chapter](#deploy-to-openshift-311), you shall use the existing namespace you created during the manual deployment, by using the following CLI commands:

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

### Build Operator

We have a script ready for you. The output of this command is a ready to use Kogito Operator image to be deployed in any namespace.

```bash
$ make
```

### Deploy to OpenShift 4.x for development purposes

To install this operator on OpenShift 4 for end-to-end testing, make sure you have access to a quay.io account to create an application repository. Follow the [authentication](https://github.com/operator-framework/operator-courier/#authentication) instructions for Operator Courier to obtain an account token. This token is in the form of "basic XXXXXXXXX" and both words are required for the command.

Push the operator bundle to your quay application repository as follows:

```bash
$ operator-courier push deploy/olm-catalog/kogito-cloud-operator/ namespace kogitocloud-operator 0.6.0 "basic XXXXXXXXX"
```

If pushing to another quay repository, replace _namespace_ with your username or other namespace. Notice that the push command does not overwrite an existing repository, and the bundle needs to be deleted before a new version can be built and uploaded. Once the bundle has been uploaded, create an [Operator Source](https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md#linking-the-quay-application-repository-to-your-openshift-40-cluster) to load your operator bundle in OpenShift.

Note that the OpenShift cluster needs access to the created application. Make sure that the application is **public** or you have configured the private repository credentials in the cluster. To make the application public, go to your quay.io account, and in the _Applications_ tab look for the `kogitocloud-operator` application. Under the settings section click on _make public_ button.

```bash
## kogito imagestreams should already be installed/available ... e.g.
$ oc apply -f https://raw.githubusercontent.com/kiegroup/kogito-cloud/master/s2i/kogito-imagestream.yaml -n openshift
$ oc create -f deploy/olm-catalog/kogito-cloud-operator/kogitocloud-operatorsource.yaml
```

Remember to replace _registryNamespace_ in the `kogitocloud-operatorsource.yaml` with your quay namespace. The name, display name and publisher of the operator are the only other attributes that may be modified.

It will take a few minutes for the operator to become visible under the _OperatorHub_ section of the OpenShift console _Catalog_. The Operator can be easily found by filtering the provider type to _Custom_.

It's possible to verify the operator status by running:

```bash
$ oc describe operatorsource.operators.coreos.com/kogitocloud-operator -n openshift-marketplace
```

### Running End to End tests

#### Operator SDK

If you have an OpenShift cluster and admin privileges, you can run e2e tests with the following command:

```bash
$ make run-e2e namespace=<namespace> tag=<tag> maven_mirror=<maven_mirror_url> image=<image_tag> tests=<full|jvm|native>
```

Where:

1. `namespace` (required) is a given temporary namespace where the test will run. You don't need to create the namespace, since it will be created and deleted after running the tests
2. `tag` (optional, default is current release) is the image tag for the Kogito image builds, for example: `0.6.0-rc1`. Useful on situations where [Kogito Cloud images](https://github.com/kiegroup/kogito-cloud/tree/master/s2i) haven't released yet and are under a temporary tag
3. `maven_mirror` (optional, default is empty) the Maven mirror URL. Useful when you need to speed up the build time by referring to a closer maven repository
4. `image` (optional, default is empty) indicates if the e2e test should be executed against specified Kogito operator image. If value is empty then local operator source code is used for test execution.
4. `tests` (optional, default is `full`) indicates what types of tests should be executed. Possible values are `full`, `jvm` and `native`. If full is specified or parameter is not provided then both JVM and native tests are executed.

In case of errors while running this test, a huge log dump will appear in your terminal. To save the test output in a local file to be analysed later, use the command below:

```bash
make run-e2e namespace=kogito-e2e  2>&1 | tee log.out
```

#### Kogito CLI

You can run a smoke test using the Kogito CLI during development to make to sure that at least the basic use case is covered. 

Before running this test, if on OpenShift 4.x install the Kogito Operator first in the namespace that the test will run. On OpenShift 3.11, the CLI will install it for you.

The command is pretty much similar to the one described for the Operator SDK:

 ```bash
 $ make run-e2e-cli namespace=<namespace> tag=<tag> native=<true|false> maven_mirror=<maven_mirror_url> skip_build=<true|false>
 ```

Where:

1. `namespace` (required) is a given namespace where the test will run.
2. `tag` (optional, default is current release) is the image tag for the Kogito image builds, for example: `0.6.0-rc1`. Useful on situations where [Kogito Cloud images](https://github.com/kiegroup/kogito-cloud/tree/master/s2i) haven't released yet and are under a temporary tag
3. `native` (optional, default is `false`) indicates if the e2e test should use native or jvm builds. See [Native X JVM Builds](#native-x-jvm-builds)
4. `maven_mirror` (optional, default is empty) the Maven mirror URL. Useful when you need to speed up the build time by referring to a closer maven repository
5. `skip_build` (optional, default is `true`) set to `true` to skip building the CLI before running the test

### Running Locally

Change log level at runtime with the `DEBUG` environment variable. e.g. -

```bash
$ make mod
$ make clean
$ DEBUG=true operator-sdk up local --namespace=<namespace>
```

Before submitting PR, please be sure to read the [contributors guide](docs/CONTRIBUTING.MD).

It's always worth noting that you should generate, vet, format, lint, and test your code. This all can be done with one command.

```bash
$ make test
```

## Prometheus Integration

### Prometheus Annotations

By default, if your Kogito Runtime Service has the [`monitoring-prometheus-addon`](https://github.com/kiegroup/kogito-runtimes/wiki/Configuration#enabling-metrics) dependency, metrics for the Kogito Service will be enabled. The Kogito Operator will add annotations to the pod and service of the deployed application, for example:

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    org.kie.kogito/managed-by: Kogito Operator
    org.kie.kogito/operator-crd: KogitoApp
    prometheus.io/path: /metrics
    prometheus.io/port: "8080"
    prometheus.io/scheme: http
    prometheus.io/scrape: "true"
  labels:
    app: onboarding-service
    onboarding: process
  name: onboarding-service
  namespace: kogito
  ownerReferences:
  - apiVersion: app.kiegroup.org/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: KogitoApp
    name: onboarding-service
spec:
  clusterIP: 172.30.173.165
  ports:
  - name: http
    port: 8080
    protocol: TCP
    targetPort: 8080
  selector:
    app: onboarding-service
    onboarding: process
  sessionAffinity: None
  type: ClusterIP
status:
  loadBalancer: {}
``` 

### Prometheus Operator

Those annotations [won't work for the Prometheus Operator](https://github.com/helm/charts/tree/master/stable/prometheus-operator#prometheusioscrape). If you're deploying on OpenShift 4.x, chances are that you're using the Prometheus Operator. 

In a scenario where the Prometheus is deployed and managed by the Prometheus Operator, and if metrics for the Kogito Service are enabled, a new [`ServiceMonitor`](https://github.com/coreos/prometheus-operator/blob/master/example/prometheus-operator-crd/servicemonitor.crd.yaml) resource will be deployed by the Kogito Operator to expose the metrics to Prometheus to scrape:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    app: onboarding-service
  name: onboarding-service
  namespace: kogito
spec:
  endpoints:
  - path: /metrics
    targetPort: 8080
    scheme: http
  namespaceSelector:
    matchNames:
    - kogito
  selector:
    matchLabels:
      app: onboarding-service
```

A [`Prometheus`](https://github.com/coreos/prometheus-operator/blob/master/example/prometheus-operator-crd/prometheus.crd.yaml) resource which is managed by the Prometheus Operator should be manually configured to select the ServiceMonitor:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
spec:
  serviceAccountName: prometheus
  serviceMonitorSelector:
    matchLabels:
      app: onboarding-service
```

Then you can see the endpoint being scraped by Prometheus in the _Target_ web console:

![](docs/img/prometheus_target.png)

The metrics exposed by the Kogito Service can be seen on the _Graph_, for example:

![](docs/img/prometheus_graph.png)

For more information about the Prometheus Operator, check the [Getting Started guide](https://github.com/coreos/prometheus-operator/blob/master/Documentation/user-guides/getting-started.md). 

## Infinispan Integration

To make it easy to have an Infinispan server up and running in your project, Kogito Operator has a resource called `KogitoInfra` to handle Infinispan deployment for you. 

For the Data Index service, if not provided a service URL to connect to the Infinispan server, one will be deployed using the [Infinispan Operator](https://github.com/infinispan/infinispan-operator).

A random password for the `developer` user will be created and injected into the Data Index automatically. You don't have to do anything to have both services to work together.

If you have plans to scale the Infinispan cluster, you can easily edit the [Infinispan CR](https://github.com/infinispan/infinispan-operator/blob/master/pkg/apis/infinispan/v1/infinispan_types.go) to meet your requirements. 

By default, `KogitoInfra` will create a secret that holds the username and password to authenticate to this server. To see it, just run:

```bash
$ oc get secret/kogito-infinispan-credential -o yaml

apiVersion: v1
data:
  password: VzNCcW9DeXdpMVdXdlZJZQ==
  username: ZGV2ZWxvcGVy
kind: Secret
(...)
```

The key values are masked by Base64 algorithm, to view the password from the above output in your terminal use:

```bash
$ echo VzNCcW9DeXdpMVdXdlZJZQ== | base64 -d

W3BqoCywi1WWvVIe
```

## Contributing

Please take a look at the [Contributing to Kogito Operator](docs/CONTRIBUTING.MD) guide.
