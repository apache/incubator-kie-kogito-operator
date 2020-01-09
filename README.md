# Kogito Operator

[![Go Report Card](https://goreportcard.com/badge/github.com/kiegroup/kogito-cloud-operator)](https://goreportcard.com/report/github.com/kiegroup/kogito-cloud-operator) [![CircleCI](https://circleci.com/gh/kiegroup/kogito-cloud-operator.svg?style=svg)](https://circleci.com/gh/kiegroup/kogito-cloud-operator)

The Kogito Operator deploys [Kogito Runtimes](https://github.com/kiegroup/kogito-runtimes) services from source and every piece of infrastructure that the services might need, such as SSO ([Keycloak](https://github.com/integr8ly/keycloak-operator)) and Persistence ([Infinispan](https://github.com/infinispan/infinispan-operator)).

Table of Contents
=================

   * [Kogito Operator](#kogito-operator)
   * [Table of Contents](#table-of-contents)
      * [Kogito Operator requirements](#kogito-operator-requirements)
      * [Kogito Operator installation](#kogito-operator-installation)
         * [Deploying to OpenShift 4.x](#deploying-to-openshift-4x)
            * [Automatically in OperatorHub](#automatically-in-operatorhub)
            * [Manually in OperatorHub](#manually-in-operatorhub)
            * [Locally on your system](#locally-on-your-system)
         * [Deploying to OpenShift 3.11](#deploying-to-openshift-311)
      * [Kogito Runtimes service deployment](#kogito-runtimes-service-deployment)
         * [Deploying a new service](#deploying-a-new-service)
         * [Cleaning up a Kogito service deployment](#cleaning-up-a-kogito-service-deployment)
         * [Native X JVM builds](#native-x-jvm-builds)
         * [Troubleshooting Kogito Runtimes service deployment](#troubleshooting-kogito-runtimes-service-deployment)
            * [No builds are running](#no-builds-are-running)
      * [Kogito Data Index Service deployment](#kogito-data-index-service-deployment)
         * [Deploying Infinispan](#deploying-infinispan)
         * [Deploying Strimzi](#deploying-strimzi)
            * [Kafka](#kafka)
            * [Kafka Topics](#kafka-topics)
         * [Kogito Data Index Service installation](#kogito-data-index-service-installation)
            * [Retrieving Kafka internal URLs](#retrieving-kafka-internal-urls)
            * [Installing the Kogito Data Index Service with the Kogito CLI](#installing-the-kogito-data-index-service-with-the-kogito-cli)
            * [Installing the Kogito Data Index Service with the Operator Catalog (OLM)](#installing-the-kogito-data-index-service-with-the-operator-catalog-olm)
            * [Installing the Kogito Data Index Service with the oc client](#installing-the-kogito-data-index-service-with-the-oc-client)
         * [Kogito Data Index Integration with persistent Kogito Services](#kogito-data-index-integration-with-persistent-kogito-services)
      * [Kogito Jobs Service deployment](#kogito-jobs-service-deployment)
         * [Kogito Jobs Service installation](#kogito-jobs-service-installation)
            * [Installing the Kogito Jobs Service with the Kogito CLI](#installing-the-kogito-jobs-service-with-the-kogito-cli)
            * [Installing the Kogito Jobs Service with the Operator Catalog (OLM)](#installing-the-kogito-jobs-service-with-the-operator-catalog-olm)
            * [Installing the Kogito Data Index Service with the oc client](#installing-the-kogito-data-index-service-with-the-oc-client-1)
         * [Enabling Persistence with Infinispan](#enabling-persistence-with-infinispan)
      * [Kogito CLI](#kogito-cli)
         * [Kogito CLI requirements](#kogito-cli-requirements)
         * [Kogito CLI installation](#kogito-cli-installation)
            * [For Linux](#for-linux)
            * [For Windows](#for-windows)
            * [Building the Kogito CLI from source](#building-the-kogito-cli-from-source)
         * [Deploying a Kogito service from source with the Kogito CLI](#deploying-a-kogito-service-from-source-with-the-kogito-cli)
      * [Prometheus integration with the Kogito Operator](#prometheus-integration-with-the-kogito-operator)
         * [Prometheus annotations](#prometheus-annotations)
         * [Prometheus Operator](#prometheus-operator)
      * [Infinispan integration](#infinispan-integration)
         * [Infinispan for Kogito Services](#infinispan-for-kogito-services)
         * [Infinispan for Data Index Service](#infinispan-for-data-index-service)
      * [Kafka integration](#kafka-integration)
         * [Kafka for Kogito Services](#kafka-for-kogito-services)
         * [Kafka For Data Index](#kafka-for-data-index)
      * [Kogito Operator development](#kogito-operator-development)
         * [Building the Kogito Operator](#building-the-kogito-operator)
         * [Deploying to OpenShift 4.x for development purposes](#deploying-to-openshift-4x-for-development-purposes)
         * [Running End-to-End (E2E) tests](#running-end-to-end-e2e-tests)
            * [With the Kogito Operator SDK](#with-the-kogito-operator-sdk)
            * [With the Kogito CLI](#with-the-kogito-cli)
         * [Running Smoke tests](#running-smoke-tests)
            * [Running smoke tests with current branch](#running-smoke-tests-with-current-branch)
         * [Running the Kogito Operator locally](#running-the-kogito-operator-locally)
      * [Contributing to the Kogito Operator](#contributing-to-the-kogito-operator)

Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc)

## Kogito Operator requirements

- Go v1.12 is installed.
- The [operator-sdk](https://github.com/operator-framework/operator-sdk/releases) v0.11.0 is installed.
- OpenShift 3.11 or 4.x is installed. (You can use [CRC](https://github.com/code-ready/crc) for local deployment.)

## Kogito Operator installation

### Deploying to OpenShift 4.x

The Kogito operator is a namespaced operator, so you must install it into the namespace where you want your Kogito application to run.

(Optional) You can import the Kogito image stream using the `oc client` manually with the following command:

```bash
$ oc apply -f https://raw.githubusercontent.com/kiegroup/kogito-cloud/master/s2i/kogito-imagestream.yaml -n openshift
```

This step is optional because the Kogito Operator creates the required imagestreams when it installs a new application.

#### Automatically in OperatorHub

The Kogito Operator is available in the OperatorHub as a community operator. To find the Operator, search by the _Kogito_ name.

You can also verify the Operator availability in the catalog by running the following command:

```bash
$ oc describe operatorsource.operators.coreos.com/kogito-operator -n openshift-marketplace
```
Follow the OpenShift Web Console instructions in the **Catalog** -> **OperatorHub** section in the left menu to install it in any namespace in the cluster.

![Kogito Operator in the Catalog](docs/img/catalog-kogito-operator.png?raw=true)

#### Manually in OperatorHub

If you cannot find the Kogito Operator in OperatorHub, you can install it manually by creating an entry in the OperatorHub Catalog:

```bash
$ oc create -f deploy/olm-catalog/kogito-operator/kogito-operator-operatorsource.yaml
```

After several minutes, the Operator appears under the **Catalog** -> **OperatorHub** section in the OpenShift Web Console. To find the Operator, search by the _Kogito_ name. You can then install the Operator as described in the [Automatically in OperatorHub](#automatically-in-operatorhub) section.

#### Locally on your system

You can also [run the Kogito Operator locally](#running-the-kogito-operator-locally) if you have the [requirements](#kogito-operator-requirements) configured on your local system.

### Deploying to OpenShift 3.11

The OperatorHub catalog is not available by default for OpenShift 3.11, so you must manually install the Kogito Operator on OpenShift 3.11.

```bash
## Kogito imagestreams should already be installed and available, for example:
$ oc apply -f https://raw.githubusercontent.com/kiegroup/kogito-cloud/master/s2i/kogito-imagestream.yaml -n openshift
$ oc new-project <project-name>
$ ./hack/3.11deploy.sh
```

## Kogito Runtimes service deployment

### Deploying a new service

Use the OLM console to subscribe to the `kogito` Operator Catalog Source within your namespace. After you subscribe, use the console to `Create KogitoApp` or create one manually as shown in the following example:

```bash
$ oc create -f deploy/crds/app.kiegroup.org_v1alpha1_kogitoapp_cr.yaml
kogitoapp.app.kiegroup.org/example-quarkus created
```

Alternatively, you can use the [Kogito CLI](#kogito-cli) to deploy your services:

```bash
$ kogito deploy-service example-quarkus https://github.com/kiegroup/kogito-examples/ --context-dir=drools-quarkus-example
```

### Cleaning up a Kogito service deployment

```bash
$ kogito delete-service example-quarkus
```

### Native X JVM builds

By default, the Kogito services are built with traditional `java` compilers to save time and resources. This means that the final generated artifact is a JAR file with the chosen runtime (default to Quarkus) with its dependencies in the image user's home directory: `/home/kogito/bin/lib`.

Kogito services implemented with [Quarkus](https://quarkus.io/guides/kogito-guide) can be built to native binary. This means very low footprint on runtime (see [performance examples](https://www.graalvm.org/docs/examples/java-performance-examples/)), but a lot of resources during build time. For more information about AOT compilation, see [GraalVM Native Image](https://www.graalvm.org/docs/reference-manual/aot-compilation/).

In Kogito Operator tests, native builds take approximately 10 minutes and the build pod can consume up to 10GB of RAM and 1.5 CPU cores. Ensure that you have these resources available when running native builds.

To deploy a service using native builds, run the `deploy-service` command with the `--native` flag:

```bash
$ kogito deploy-service example-quarkus https://github.com/kiegroup/kogito-examples/ --context-dir=drools-quarkus-example --native
```

### Troubleshooting Kogito Runtimes service deployment

#### No builds are running

If you do not see any builds running nor any resources created in the namespace, review the Kogito Operator log.

To view the Operator logs, first identify where the operator is deployed:

```bash
$ oc get pods

NAME                                     READY   STATUS      RESTARTS   AGE
kogito-operator-6d7b6d4466-9ng8t   1/1     Running     0          26m
```

Use the pod name as the input in the following command:

```bash
$ oc logs -f kogito-operator-6d7b6d4466-9ng8t
```

## Kogito Data Index Service deployment

The Kogito Operator can deploy the [Data Index Service](https://github.com/kiegroup/kogito-runtimes/wiki/Data-Index-Service) as a [Custom Resource](deploy/crds/app.kiegroup.org_v1alpha1_kogitodataindex_cr.yaml) (`KogitoDataIndex`).

The Data Index Service depends on Kafka, starting with version 0.6.0, the Kogito Operator deploys an Apache Kafka Cluster (based on [Strimzi](https://strimzi.io/) operator) in the same namespace.

The Data Index Service also depends on Infinispan, but starting with version 0.6.0 of the Kogito Operator, Infinispan Server is automatically deployed for you.

### Deploying Infinispan

If you plan to use the Data Index Service to connect to an Infinispan Server instance deployed within the same namespace, the Kogito Operator can handle this deployment for you.

When you install the Kogito Operator from OperatorHub, the Infinispan Operator is installed in the same namespace. If you do not have access to OperatorHub or OLM in your cluster, you can [manually deploy the Infinispan Operator](https://infinispan.org/infinispan-operator/master/documentation/asciidoc/titles/operator.html#deploying_operator_manually).

After you deploy the Infinispan Operator, see [Deploying Strimzi](#deploying-strimzi) for next steps.

### Deploying Strimzi

#### Kafka

[Strimzi](https://strimzi.io/) Operator should be available in the [OperatorHub](https://operatorhub.io/operator/strimzi-kafka-operator). In the OpenShift Web Console, go to **Catalog** -> **Installed Operators** in the left menu and search for `Strimzi`.

You should see the Strimzi Operator in the **Installed Operators** tab:

![Strimzi Installed](docs/img/strimzi_installed.png?raw=true)

Next, create a Kafka cluster for the Data Index Service to connect. Click the name of the Strimzi Operator, then click the **Kafka** tab and click **Create Kafka**. Accept the default options to create a three-node Kafka cluster. If this is a development environment, consider setting the Zookeeper and Kafka replicas to 1 to conserve resources.

After a few minutes, you should see the pods running and the services available:

```bash
$ oc get svc -l strimzi.io/cluster=my-cluster

NAME                          TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                      AGE
my-cluster-kafka-bootstrap    ClusterIP   172.30.228.90    <none>        9091/TCP,9092/TCP,9093/TCP   9d
my-cluster-kafka-brokers      ClusterIP   None             <none>        9091/TCP,9092/TCP,9093/TCP   9d
my-cluster-zookeeper-client   ClusterIP   172.30.241.146   <none>        2181/TCP                     9d
my-cluster-zookeeper-nodes    ClusterIP   None             <none>        2181/TCP,2888/TCP,3888/TCP   9d
```

The service that you will use to deploy the Data Index Service is `my-cluster-kafka-bootstrap:9092`.

#### Kafka Topics

If Strimzi Operator is installed in the namespace, the Kogito Operator will create the following `Kafka Topic`s that are required by the Data Index Service: 

- `kogito-processinstances-events`
- `kogito-usertaskinstances-events`
- `kogito-processdomain-events`
- `kogito-usertaskdomain-events`

Now that you have the required infrastructure, you can deploy the Kogito Data Index Service.

### Kogito Data Index Service installation

#### Retrieving Kafka internal URLs

After you configure the Kafka Operator, you need the Kafka service URI or Kafka instance name to install the Data Index Service.

Run the following command to retrieve the Kafka internal URI:

```bash
$ oc get svc -l strimzi.io/cluster=my-cluster | grep bootstrap

my-cluster-kafka-bootstrap    ClusterIP   172.30.228.90    <none>        9091/TCP,9092/TCP,9093/TCP   9d
```

In this case, the Kafka Cluster service is `my-cluster-kafka-bootstrap:9092`.

Or run the following command to retrieve the Kafka instance name:
```bash
$ oc get kafka

NAME         AGE
my-cluster   17s
``` 

In this case the Kafka instance name is `my-cluster`.

Use this information to create the Kogito Data Index resource.

#### Installing the Kogito Data Index Service with the Kogito CLI

If you have installed the [Kogito CLI](#kogito-cli), run the following command to create the Kogito Data Index resource. Replace the URLs with the URLs you retrieved for your environment:

```bash
$ kogito install data-index -p my-project --kafka-url my-cluster-kafka-bootstrap:9092
```

Or run the following command to create the Kogito Data Index resource with the Kafka instance name:

```bash
$ kogito install data-index -p my-project --kafka-instance my-cluster
```

Infinispan is deployed for you using the Infinispan Operator. Ensure that the Infinispan deployment is running in your project. If the deployment fails, the following error message appears:

```
Infinispan Operator is not available in the Project: my-project. Please make sure to install it before deploying Data Index without infinispan-url provided
```

To resolve the error, review the deployment procedure to this point to ensure that all steps have been successful.

#### Installing the Kogito Data Index Service with the Operator Catalog (OLM)

If you are running on OpenShift 4.x, you can use the OperatorHub user interface to create the Kogito Data Index resource. In the OpenShift Web Console, go to **Installed Operators** -> **Kogito Operator** -> **Kogito Data Index**. Click **Create Kogito Data Index** and create a new resource that uses the Infinispan and Kafka services, as shown in the following example:

```yaml
apiVersion: app.kiegroup.org/v1alpha1
kind: KogitoDataIndex
metadata:
  name: kogito-data-index
spec:
  # Number of pods to be deployed
  replicas: 1
  # Image to use for this deployment
  image: "quay.io/kiegroup/kogito-data-index:latest"
  kafka:
    # Service name and port for the Kafka cluster, for example, my-kafka-cluster:9092
    externalURI: my-cluster-kafka-bootstrap:9092
```

#### Installing the Kogito Data Index Service with the oc client

To create the Kogito Data Index resource using the oc client, you can use the CR file from the previous example as a reference and create the custom resource from the command line as shown in the following example:

```bash
# Clone this repository
$ git clone https://github.com/kiegroup/kogito-cloud-operator.git
$ cd kogito-cloud-operator
# Make your changes
$ vi deploy/crds/app.kiegroup.org_v1alpha1_kogitodataindex_cr.yaml
# Deploy to the cluster
$ oc create -f deploy/crds/app.kiegroup.org_v1alpha1_kogitodataindex_cr.yaml -n my-project
```

You can access the GraphQL interface through the route that was created for you:

```bash
$ oc get routes -l app=kogito-data-index

NAME                HOST/PORT                                                  PATH   SERVICES            PORT   TERMINATION   WILDCARD
kogito-data-index   kogito-data-index-kogito.apps.mycluster.example.com               kogito-data-index   8080   None
```

### Kogito Data Index Integration with persistent Kogito Services

If your Kogito Service has [persistence](https://github.com/kiegroup/kogito-runtimes/wiki/Persistence) enabled, Data Index will mount a volume based on a `configMap` created for you during the deployment of the service. 

This `configMap` has the `-protobuf-files` suffix and inside it you'll see the protobuf files that your service generated 
during build time. For example:

```yaml
kind: ConfigMap
apiVersion: v1
metadata:
  name: example-quarkus-protobuf-files
  labels:
    kogito-protobuf: true
data:
  visaApplications.proto: |-
    syntax = "proto2"; 
    package org.acme.travels.visaApplications; 
    import "kogito-index.proto";
    import "kogito-types.proto";
    option kogito_model = "VisaApplications";
    option kogito_id = "visaApplications";
    # data suppressed for brevit
 ```    
During the deployment of a new set of protobuf files (e.g. a new persistent Kogito Service is deployed), Data Index 
will spin up a new `pod` referring to this new volume attached.

Updated protobuf files will be automatically refreshed by Kubernetes volumes after some time, this means that if you add a new
property in your domain data, this data will be reflect automatically in the Data Index without restarts.

_Please note that removed Kogito Services will remove the protobuf files associated to it as well. This means that you won't
be able to see the data through the Data Index anymore, although the data still persisted in Infinispan._

## Kogito Jobs Service deployment

Like Data Index, [Jobs Service](https://github.com/kiegroup/kogito-runtimes/wiki/Jobs-Service) can be deployed via Operator or CLI. 
If persistence is required, the operator will also deploy an Infinispan server using Infinispan Operator.

### Kogito Jobs Service installation

There's a couple of ways to install the Jobs Service into your namespace using Kogito Operator.

#### Installing the Kogito Jobs Service with the Kogito CLI

If you have installed the [Kogito CLI](https://github.com/kiegroup/kogito-cloud-operator#kogito-cli), 
run the following command to create the Kogito Jobs Service resource:

``` 
$ kogito install jobs-service -p my-project
```

There's some options to customize the Jobs Service deployment with CLI. 
Run `kogito install jobs-service --help` to understand and set them according to your requirements.  

#### Installing the Kogito Jobs Service with the Operator Catalog (OLM)

If you are running on OpenShift 4.x, you can use the OperatorHub user interface to create the Kogito Jobs Service resource. 
In the OpenShift Web Console, go to **Installed Operators** -> **Kogito Operator** -> **Kogito Jobs Service**. 
Click **Create Kogito Jobs Service** and create a new resource as shown in the following example:

```yaml
apiVersion: app.kiegroup.org/v1alpha1
kind: KogitoJobsService
metadata:
  name: jobs-service
spec:
  replicas: 1
```

#### Installing the Kogito Data Index Service with the oc client

To create the Kogito Jobs Service resource using the oc client, you can use the CR file from the previous example as a reference 
and create the custom resource from the command line as shown in the following example:

```bash
 # Clone this repository
 $ git clone https://github.com/kiegroup/kogito-cloud-operator.git
 $ cd kogito-cloud-operator
 # Make your changes
 $ vi deploy/crds/app.kiegroup.org_v1alpha1_kogitojobsservice_cr.yaml
 # Deploy to the cluster
 $ oc create -f deploy/crds/app.kiegroup.org_v1alpha1_kogitojobsservice_cr.yaml -n my-project
```

### Enabling Persistence with Infinispan

Jobs Service supports persistence with Infinispan by setting the property `spec.infinispan.useKogitoInfra` to `true` in the
CR or the flag `--enable-persistence` in the CLI.

When doing this Kogito Operator deploys a new Infinispan server using Infinispan Operator for you within the same namespace
that you're deploying Jobs Service. It also sets all information regarding server authentication.

For this to work, bear in mind that Infinispan Operator must be installed in the namespace. If the Kogito Operator was installed
with OLM, it means that the Infinispan Operator would also be installed. If it was installed manually, you will also have to manually
install the Infinispan Operator.

It's also possible to fine tune the Infinispan integration by setting the properties `spec.infinispan.credentials`, 
`spec.infinispan.uri` and `spec.infinispan.useKogitoInfra` to `false` in the CR. This way the Infinispan server won't be deployed
and the Jobs Service will try to connect to the given URI. Just make sure that your cluster have access to this URI.

This process behaves similarly to the one defined by the [Data Index Service](#deploying-infinispan). 

## Kogito CLI

The Kogito CLI tool enables you to deploy new Kogito services from source instead of relying on CRs and YAML files.

### Kogito CLI requirements

- The [`oc` client](https://docs.okd.io/latest/cli_reference/get_started_cli.html) is installed.
- You are an authenticated OpenShift user with permissions to create resources in a given namespace.

### Kogito CLI installation

#### For Linux

1. Download the correct [Kogito distribution](https://github.com/kiegroup/kogito-cloud-operator/releases) for your machine.

2. Unpack the binary: `tar -xvf release.tar.gz`

   You should see an executable named `kogito`.

3. Move the binary to a pre-existing directory in your `PATH`, for example, `# cp /path/to/kogito /usr/local/bin`.

#### For Windows

1. Download the latest 64-bit Windows release of the [Kogito distribution](https://github.com/kiegroup/kogito-cloud-operator/releases).

2. Extract the zip file through a file browser.

3. Add the extracted directory to your `PATH`. You can now use `kogito` from the command line.

#### Building the Kogito CLI from source

| :warning: To build the Kogito CLI from source, ensure that [Go is installed](https://golang.org/doc/install) and available in your `PATH`. |
| --- |

Run the following command to build and install the Kogito CLI:

```bash
$ git clone https://github.com/kiegroup/kogito-cloud-operator
$ cd kogito-cloud-operator
$ make install-cli
```

The `kogito` CLI appears in your path:

```bash
$ kogito
Kogito CLI deploys your Kogito services into an OpenShift cluster

Usage:
  kogito [command]

Available Commands:
  delete-project Deletes a Kogito Project - i.e., the Kubernetes/OpenShift namespace
  delete-service Deletes a Kogito Runtime Service deployed in the namespace/project
  deploy-service Deploys a new Kogito Runtime Service into the given Project
  help           Help about any command
  install        Install all sort of infrastructure components to your Kogito project
  new-project    Creates a new Kogito Project for your Kogito services
  use-project    Sets the Kogito Project where your Kogito service will be deployed

Flags:
      --config string   config file (default is $HOME/.kogito.json)
  -h, --help            help for kogito
  -v, --verbose         verbose output
      --version         display version

Use "kogito [command] --help" for more information about a command.
```

### Deploying a Kogito service from source with the Kogito CLI

After you complete the [Kogito Operator installation](#kogito-operator-installation), you can deploy a new Kogito service by using the Kogito CLI:

```bash
# creates a new namespace in your cluster
$ kogito new-project kogito-cli

# deploys a new Kogito Runtime Service from source
$ kogito deploy-service example-drools https://github.com/kiegroup/kogito-examples --context-dir drools-quarkus-example
```

If you are using OpenShift 3.11 as described in [Deploying to OpenShift 3.11](#deploying-to-openshift-311), use the existing namespace that you created during the manual deployment, as shown in the following example:

```bash
# Use the provisioned namespace in your OpenShift 3.11 cluster
$ kogito use-project <project-name>

# Deploys new Kogito service from source
$ kogito deploy-service example-drools https://github.com/kiegroup/kogito-examples --context-dir drools-quarkus-example
```

You can shorten the previous command as shown in the following example:

```bash
$ kogito deploy-service example-drools https://github.com/kiegroup/kogito-examples --context-dir drools-quarkus-example --project <project-name>
```
## Prometheus integration with the Kogito Operator

### Prometheus annotations

By default, if your Kogito Runtimes service contains the `monitoring-prometheus-addon` dependency, metrics for the Kogito service are enabled. For more information about Prometheus metrics in Kogito services, see [Enabling metrics](https://github.com/kiegroup/kogito-runtimes/wiki/Configuration#enabling-metrics).

The Kogito Operator adds Prometheus annotations to the pod and service of the deployed application, as shown in the following example:

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

The [Prometheus Operator](https://github.com/helm/charts/tree/master/stable/prometheus-operator#prometheusioscrape) does not directly support the Prometheus annotations that the Kogito Operator adds to your Kogito services. If you are deploying the Kogito Operator on OpenShift 4.x, then you are likely using the Prometheus Operator.

Therefore, in a scenario where Prometheus is deployed and managed by the Prometheus Operator, and if metrics for the Kogito service are enabled, a new [`ServiceMonitor`](https://github.com/coreos/prometheus-operator/blob/master/example/prometheus-operator-crd/servicemonitor.crd.yaml) resource is deployed by the Kogito Operator to expose the metrics for Prometheus to scrape:

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

You must manually configure your [`Prometheus`](https://github.com/coreos/prometheus-operator/blob/master/example/prometheus-operator-crd/prometheus.crd.yaml) resource that is managed by the Prometheus Operator to select the `ServiceMonitor` resource:

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

After you configure your Prometheus resource with the `ServiceMonitor` resource, you can see the endpoint being scraped by Prometheus in the **Targets** page of the Prometheus web console:

![](docs/img/prometheus_target.png)

The metrics exposed by the Kogito service appear in the **Graph** view:

![](docs/img/prometheus_graph.png)

For more information about the Prometheus Operator, see the [Prometheus Operator](https://github.com/coreos/prometheus-operator/blob/master/Documentation/user-guides/getting-started.md) documentation.

## Infinispan integration

To help you start and run an Infinispan Server instance in your project, the Kogito Operator has a resource called `KogitoInfra` to handle Infinispan deployment for you.

The `KogitoInfra` resource use the [Infinispan Operator](https://github.com/infinispan/infinispan-operator) to deploy new Infinispan server instances if needed.

You can freely edit and manage the Infinispan instance. Kogito Operator do not manage or handle the Infinispan instances. 
For example, if you have plans to scale the Infinispan cluster, you can edit the `replicas` field in the [Infinispan CR](https://github.com/infinispan/infinispan-operator/blob/master/pkg/apis/infinispan/v1/infinispan_types.go) to meet your requirements.

By default, the `KogitoInfra` resource creates a secret that holds the user name and password for Infinispan authentication. 
To view the credentials, run the following command:

```bash
$ oc get secret/kogito-infinispan-credential -o yaml

apiVersion: v1
data:
  password: VzNCcW9DeXdpMVdXdlZJZQ==
  username: ZGV2ZWxvcGVy
kind: Secret
(...)
```

The key values are masked by a Base64 algorithm. To view the password from the previous example output in your terminal, 
run the following command:

```bash
$ echo VzNCcW9DeXdpMVdXdlZJZQ== | base64 -d

W3BqoCywi1WWvVIe
```

For more information about Infinispan Operator, please see [their official documentation](https://infinispan.org/infinispan-operator/master/documentation/asciidoc/titles/operator.html).

**Note:** *Sometimes the [OperatorHub will install DataGrid operator instead of Infinispan](https://github.com/operator-framework/operator-lifecycle-manager/issues/1158) when installing Kogito Operator. 
If this happens, please uninstall DataGrid and install Infinispan manually since they are not compatible*

### Infinispan for Kogito Services

If your Kogito Service depends on the [persistence add-on](https://github.com/kiegroup/kogito-runtimes/wiki/Persistence), 
Kogito Operator installs Infinispan and inject the connection properties as environment variables into the service. 
Depending on the runtime, this variables will differ. See the table below:


|Quarkus Runtime                          |Springboot Runtime              | Description                                       |Example                |
|-----------------------------------------|--------------------------------|---------------------------------------------------|-----------------------| 
|QUARKUS_INFINISPAN_CLIENT_SERVER_LIST    |INFINISPAN_REMOTE_SERVER_LIST   |Service URI from deployed Infinispan               |kogito-infinispan:11222|
|QUARKUS_INFINISPAN_CLIENT_AUTH_USERNAME  |INFINISPAN_REMOTE_AUTH_USER_NAME|Default username generated by Infinispan Operator  |developer              |
|QUARKUS_INFINISPAN_CLIENT_AUTH_PASSWORD  |INFINISPAN_REMOTE_AUTH_PASSWORD |Random password generated by Infinispan Operator   |Z1Nz34JpuVdzMQKi       |
|QUARKUS_INFINISPAN_CLIENT_SASL_MECHANISM |INFINISPAN_REMOTE_SASL_MECHANISM|Default to `PLAIN`                                 |`PLAIN`                |

Just make sure that your Kogito Service can read these properties in runtime. 
Those variables names are the same as the ones used by Infinispan clients from Quarkus and Springboot.

[On Quarkus versions below 1.1.0 (Kogito 0.6.0)](https://github.com/quarkusio/quarkus/issues/5708), make sure that your `aplication.properties` file has the properties listed like the example below:

```properties
quarkus.infinispan-client.server-list=
quarkus.infinispan-client.auth-username=
quarkus.infinispan-client.auth-password=
quarkus.infinispan-client.sasl-mechanism=
```

These properties are replaced by the environment variables in runtime.

You can control the installation method for the Infinispan by using the flag `infinispan-install` in the Kogito CLI or 
editing the `spec.infra.installInfinispan` in `KogitoApp` custom resource:

- **`Auto`** - The operator tries to discover if the service needs persistence by scanning the runtime image for the `org.kie/persistence/required` label attribute
- **`Always`** - Infinispan is installed in the namespace without checking if the service needs persistence or not
- **`Never`** - Infinispan is not installed, even if the service requires persistence. Use this option only if you intend to deploy your own persistence mechanism and you know how to configure your service to access it
  

### Infinispan for Data Index Service

For the Data Index Service, if you do not provide a service URL to connect to Infinispan, a new server is deployed via `KogitoInfra`.

A random password for the `developer` user is created and injected into the Data Index automatically.
You do not need to do anything for both services to work together.

## Kafka integration

Like Infinispan, Kogito Operator can deploy a Kafka cluster for your Kogito services via `KogitoInfra` custom resource.

To deploy a Kafka cluster with Zookeeper to support sending and receiving messages within a process, Kogito Operator relies
on the [Strimzi Operator](https://strimzi.io/docs/latest/).

You can freely edit the Kafka instance deployed by the operator to fulfill any requirement that you need. 
The Kafka instance is not managed by Kogito, instead it's managed by Strimzi. 
That's why Kogito Operator is dependant on Strimzi Operator and it's installed once you install the Kogito Operator using
OLM.

_**Note:** Sometimes the OperatorHub will install AMQ Streams operator instead of Strimzi when installing Kogito Operator. 
If this happens, please uninstall AMQ Streams and install Strimzi manually since they are not compatible_

### Kafka for Kogito Services

To enable Kafka installation during deployment of your service, use the following Kogito CLI command:

```bash
$ kogito deploy kogito-kafka-quickstart-quarkus https://github.com/mswiderski/kogito-quickstarts --install-kafka Always \
--build-env MAVEN_ARGS_APPEND="-pl kogito-kafka-quickstart-quarkus -am" ARTIFACT_DIR="kogito-kafka-quickstart-quarkus/target"  
```

Or using the custom resource (CR) yaml file:

```bash
apiVersion: app.kiegroup.org/v1alpha1
kind: KogitoApp
metadata:
  name: kogito-kafka-quickstart-quarkus
spec:
  infra:
    installKafka: Always
  build:
    env:
    - name: MAVEN_ARGS_APPEND
      value: -pl kogito-kafka-quickstart-quarkus -am
    - name: ARTIFACT_DIR
      value: kogito-kafka-quickstart-quarkus/target
    gitSource:
      uri: https://github.com/mswiderski/kogito-quickstarts
```

The flag `--install-kafka Always` in the CLI and the attribute `installKafka: Always` in the CR tells to the operator
to deploy a Kafka cluster in the namespace if no Kafka cluster owned by Kogito Operator is found.

A variable named `KAFKA_BOOTSTRAP_SERVERS` is injected into the service container. For Quarkus runtimes, this works
out of the box when using [Kafka Client](https://quarkus.io/guides/kafka-streams#topic-configuration) version 1.x or greater. 
For Springboot you might need to rely on property substitution in the `application.properties` like:

```properties
spring.kafka.bootstrap.servers=${KAFKA_BOOTSTRAP_SERVERS}
```

Also, if the container has any environment variable with the suffix `_BOOTSTRAP_SERVERS`, they are injected by the 
value of `KAFKA_BOOTSTRAP_SERVERS` variable as well. For example, by running:
 
```bash
$ kogito deploy kogito-kafka-quickstart-quarkus https://github.com/mswiderski/kogito-quickstarts --install-kafka Always \
--build-env MAVEN_ARGS_APPEND="-pl kogito-kafka-quickstart-quarkus -am" ARTIFACT_DIR="kogito-kafka-quickstart-quarkus/target \
-e MP_MESSAGING_INCOMING_TRAVELLERS_BOOTSTRAP_SERVERS -e MP_MESSAGING_OUTGOING_PROCESSEDTRAVELLERS_BOOTSTRAP_SERVERS"  
```
The variables `MP_MESSAGING_INCOMING_TRAVELLERS_BOOTSTRAP_SERVERS` and `MP_MESSAGING_OUTGOING_PROCESSEDTRAVELLERS_BOOTSTRAP_SERVERS`
will have the deployed Kafka service URL inject into them.

Please note that for services with [Quarkus version below 1.1.0](https://github.com/quarkusio/quarkus/issues/5708) (Kogito Runtimes 0.6.0), it's required to add these Kafka properties in the 
`application.properties`. Otherwise they won't be replaced in runtime by environment variables injected by the operator.

### Kafka For Data Index

Data Index allows you to inform the service name defined by the Kafka cluster deployed in the same namespace.
You can let `KogitoInfra` to deploy Kafka for you via Strimzi Operator. 

Just deploy the `KogitoInfra` CR with `installKafka` attribute set to `true`:

```yaml
apiVersion: app.kiegroup.org/v1alpha1
kind: KogitoInfra
metadata:
  name: kogito-infra
spec:
  # let's install both since Data Index needs persistence and messaging
  installInfinispan: true
  installKafka: true
```

Then deploy Data Index service CR like:

```yaml
apiVersion: app.kiegroup.org/v1alpha1
kind: KogitoDataIndex
metadata:
  name: kogito-data-index
spec:
  replicas: 1
  image: "quay.io/kiegroup/kogito-data-index:0.6.0"
  kafka:
    # default kafka instance deployed by KogitoInfra
    instance: kogito-kafka
  infinispan:
    # reuse the Infinispan deployed by KogitoInfra
    useKogitoInfra: true
```

Or use Kogito CLI:

```bash
kogito install data-index --kafka-instance kogito-kafka
```

That's it! The information required to connect to Kafka will be automatically set for you by the operator to the Data Index service.

## Kogito Operator development

Before you begin fixing issues or adding new features to the Kogito Operator, see [Contributing to the Kogito Operator](docs/CONTRIBUTING.MD) and [Kogito Operator architecture](docs/ARCHITECTURE.MD).

### Building the Kogito Operator

To build the Kogito Operator, use the following command:

```bash
$ make
```

The output of this command is a ready-to-use Kogito Operator image that you can deploy in any namespace.

### Deploying to OpenShift 4.x for development purposes

To install the Kogito Operator on OpenShift 4.x for end-to-end (E2E) testing, ensure that you have access to a `quay.io` account to create an application repository. Follow the Operator Courier [authentication](https://github.com/operator-framework/operator-courier/#authentication) instructions to obtain an account token. This token is in the format `basic XXXXXXXXX` and both words are required for the command.

Push the Operator bundle to your quay application repository as shown in the following example:

```bash
$ operator-courier push deploy/olm-catalog/kogito-operator/ namespace kogito-operator 0.6.0 "basic XXXXXXXXX"
```

If you push to another quay repository, replace `namespace` with your user name or the other namespace. The push command does not overwrite an existing repository, so you must delete the bundle before you can build and upload a new version. After you upload the bundle, create an [Operator Source](https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md#linking-the-quay-application-repository-to-your-openshift-40-cluster) to load your operator bundle in OpenShift.

The OpenShift cluster needs access to the created application. Ensure that the application is **public** or that you have configured the private repository credentials in the cluster. To make the application public, go to your `quay.io` account, and in the **Applications** tab look for the `kogito-operator` application. Under the settings section, click **make public**.

```bash
## Kogito imagestreams should already be installed and available, for example:
$ oc apply -f https://raw.githubusercontent.com/kiegroup/kogito-cloud/master/s2i/kogito-imagestream.yaml -n openshift
$ oc create -f deploy/olm-catalog/kogito-operator/kogito-operator-operatorsource.yaml
```

Replace `registryNamespace` in the `kogito-operator-operatorsource.yaml` file with your quay namespace. The name, display name, and publisher of the Operator are the only other attributes that you can modify.

After several minutes, the Operator appears under **Catalog** -> **OperatorHub** in the OpenShift Web Console. To find the Operator, filter the provider type by _Custom_.

To verify the operator status, run the following command:

```bash
$ oc describe operatorsource.operators.coreos.com/kogito-operator -n openshift-marketplace
```

### Running End-to-End (E2E) tests

#### With the Kogito Operator SDK

If you have an OpenShift cluster and admin privileges, you can run E2E tests with the following command:

```bash
$ make run-e2e namespace=<namespace> tag=<tag> maven_mirror=<maven_mirror_url> image=<image_tag> tests=<full|jvm|native>
```

where:

- `namespace` (required) is a given temporary namespace where the test will run. You do not need to create the namespace because it will be created and deleted after the test runs.
- `tag` (optional, default is current release) is the image tag for the Kogito image builds, for example, `0.6.0-rc1`. This is helpful in situations where [Kogito S2I images](https://github.com/kiegroup/kogito-cloud/tree/master/s2i) have not been released yet and are under a temporary tag.
- `maven_mirror` (optional, default is empty) is the Maven mirror URL. This is helpful when you need to speed up the build time by referring to a closer Maven repository.
- `image` (optional, default is empty) indicates whether the E2E test should be executed against a specified Kogito Operator image. If the value is empty, then the local Operator source code is used for the test execution.
- `tests` (optional, default is `full`) indicates what types of tests should be executed. Possible values are `full`, `jvm`, and `native`. If you specify `full` or specify no parameter, then both JVM and native tests are executed.

If any errors are detected during this test, a detailed log appears in your command terminal.

To save the test output in a local file for future reference, run the following command:

```bash
make run-e2e namespace=kogito-e2e  2>&1 | tee log.out
```

#### With the Kogito CLI

You can run a smoke test using the Kogito CLI during development to make sure that at least the basic use case is covered.

On OpenShift 4.x, before you run this test, install the Kogito Operator in the namespace where the test will run. On OpenShift 3.11, the Kogito CLI installs the Kogito Operator for you.

To run an E2E test using the Kogito CLI, run the following command:

 ```bash
 $ make run-e2e-cli namespace=<namespace> tag=<tag> native=<true|false> maven_mirror=<maven_mirror_url> skip_build=<true|false>
 ```

where:

- `namespace` (required) is a given temporary namespace where the test will run.
- `tag` (optional, default is current release) is the image tag for the Kogito image builds, for example, `0.6.0-rc1`. This is helpful in situations where [Kogito S2I images](https://github.com/kiegroup/kogito-cloud/tree/master/s2i) have not been released yet and are under a temporary tag.
- `native` (optional, default is `false`) indicates whether the E2E test should use `native` or `jvm` builds. For more information, see [Native X JVM builds](#native-x-jvm-builds).
- `maven_mirror` (optional, default is empty) is the Maven mirror URL. This is helpful when you need to speed up the build time by referring to a closer Maven repository.
- `skip_build` (optional, default is `true`) is set to `true` to skip building the CLI before running the test.

### Running Smoke tests

**REQUIREMENT:** You need to be authenticated to the cluster before running the tests.

If you have an OpenShift cluster and admin privileges, you can run Smoke tests with the following command:

```bash
$ make run-smoke [key=value]*
```

You can set those optional keys:

- `deploy_uri` is the URI where you can find operator deployment yaml files and crds. It can be a URL or a Filesystem absolute path. Default is https://raw.githubusercontent.com/kiegroup/kogito-cloud-operator/master/deploy/.
- `maven_mirror` is the Maven mirror URL. This is helpful when you need to speed up the build time by referring to a closer Maven repository.
- `operator_image` is the Operator image full name. Example: *operator_image=quay.io/kiegroup/kogito-cloud-operator*.
- `operator_tag` is the Operator image tag. Default is the current version.
- `feature` is a specific feature you want to run. Example: *feature=deploy_quarkus_with_persistence*
- `local` to be set to true if running tests in local. Default is false.

Logs will be shown on the Terminal.

To save the test output in a local file for future reference, run the following command:

```bash
make run-smoke 2>&1 | tee log.out
```

**NOTE:** There is some more environment variables that can be set. Please see [smoke tests configuration](./test/smoke/README.md).

#### Running smoke tests with current branch 

```
$ make
$ docker tag quay.io/kiegroup/kogito-cloud-operator:0.7.0-rc1 quay.io/{USERNAME}/kogito-cloud-operator:0.7.0-rc1 
$ docker push quay.io/{USERNAME}/kogito-cloud-operator:0.7.0-rc1
$ make run-smoke --ope_name quay.io/{USERNAME}/kogito-cloud-operator
```

**NOTE:** Replace {USERNAME} with the username/group you want to push to. Docker needs to be logged in to quay.io and be able to push to your username/group.

### Running the Kogito Operator locally

To run the Kogito Operator locally, change the log level at runtime with the `DEBUG` environment variable, as shown in the following example:

```bash
$ make mod
$ make clean
$ DEBUG=true operator-sdk up local --namespace=<namespace>
```

Before submitting a [pull request](https://help.github.com/en/articles/about-pull-requests) to the Kogito Operator repository, review the instructions for [Contributing to the Kogito Operator](docs/CONTRIBUTING.MD).

You can use the following command to vet, format, lint, and test your code:

```bash
$ make test
```

## Contributing to the Kogito Operator

For information about submitting bug fixes or proposed new features for the Kogito Operator, see [Contributing to the Kogito Operator](docs/CONTRIBUTING.MD).
