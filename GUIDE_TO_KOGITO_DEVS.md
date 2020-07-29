# Guide for Core/Runtimes team to Smoke test local changes on Openshift/k8s Cluster.

This guide aims to help core/runtimes developer to test their local changes with respect to the operator.
The operator will be deployed over an Openshift or Kubernetes cluster. In this example we've taken minikube to deploy a Kubernetes cluster.

# Table of contents

* [Install Minikube](#install-minikube)
* [Build Artifacts](#build-artifacts)
* [Build the Kogito Application Images](#build-the-kogito-application-images)
* [Build the Kogito Service Image](#build-the-kogito-service-image)
* [Install the Operator](#install-the-operator)
* [Run Kogito Service with Kogito Operator](#run-kogito-service-with-kogito-operator)

## Install Minikube

We decided to go with Minikube as the Kubernetes cluster as it is very resource efficient and can be started easily on any system.
 
 ### Prerequisites
 
  * Install `kubectl` binaries on your system [see](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
  * Have a hypervisor installed (kvm  is recommended for linux)

For installing the Minikube cluster please follow [this tutorial](https://kubernetes.io/docs/tasks/tools/install-minikube/)

## Build Artifacts

Before proceeding, please read the README files on the following repositories to know the dependencies you might need to install to build and deploy the artifacts.
 * [kogito-runtimes](https://github.com/kiegroup/kogito-runtimes)
 * [kogito-apps](https://github.com/kiegroup/kogito-apps)
 * [kogito-examples](https://github.com/kiegroup/kogito-examples)
 
To generate the jar artifacts, you can just run the following command on each repository:

```shell-script
mvn clean install -DskipTests #(For skipping tests)
```

This will deploy your artifacts on your local Maven repository (default is `~/.m2/repository/org/kie/kogito`).

So, you can just build the artifacts you modified or where you added new feature in and update them so they can be built with the Images

## Build the Kogito Application Images

We’re going to use our pre-built images which are available in the `quay.io/kiegroup/` namespace.

  * kogito-data-index
  * kogito-jobs-service
  * kogito-management-console

Now let’s take an example of updating the `kogito-jobs-service` artifact in the image.
First, create a directory e.g: `~/images/`.
Then, inside this directory, copy the artifacts you need to update:

```shell-script
$ cp ~/.m2/repository/org/kie/kogito/jobs-service/8.0.0-SNAPSHOT/jobs-service-8.0.0-SNAPSHOT-runner.jar .
```

Also, inside this directory, you can have a `Dockerfile` which can be used to update the artifact.

` $ vim jobs-service-Dockerfile` 

```
FROM quay.io/kiegroup/kogito-jobs-service
COPY jobs-service-8.0.0-SNAPSHOT-runner.jar ${KOGITO_HOME}/bin/kogito-jobs-service-runner.jar
```
**Note**: Usually, latest artifacts are with version  `8.0.0-SNAPSHOT`
**Note**: You can check the location of the artifact in the image from the [kogito-images](https://github.com/kiegroup/kogito-images) repository.

Now you are going to build the image. Inside the above created directory, run:

`$ podman build -t <prefered-tag-name>  -f jobs-service-Dockerfile  .`

The above command will build your image with `quay.io/<your-namespace>/<image-name>` and can be seen in the output of `$ podman images`

**Note**: We added the `-f` flag to specify the `Dockerfile` in case you want to build multiple images, you can then have different files and specify them with the `-f` flag.
This image needs to be on a public container registry from which it can be pulled later.
```shell-script
$ podman push quay.io/<your-namespace>/<image-name>
```

## Build the Kogito Service Image

### Base Images
  * kogito-quarkus-ubi8
  * kogito-quarkus-jvm-ubi8
  * kogito-springboot-ubi8

As an example for building the Image for your Kogito custom Service we’re going to build the example  `process-business-rules-quarkus` from [kogito-examples](https://github.com/kiegroup/kogito-examples) repository.

```shell-script
$ git clone https://github.com/kiegroup/kogito-examples
cd kogito-examples/process-business-rules-quarkus
mvn clean package
```
Now create a Dockerfile inside the `kogito-examples/process-business-rules-quarkus`. With the following content:

```Dockerfile
FROM quay.io/kiegroup/kogito-quarkus-jvm-ubi8:latest

COPY target/*-runner.jar $KOGITO_HOME/bin    
COPY target/lib $KOGITO_HOME/bin/lib
```

```shell-script
$ podman build --tag quay.io/<yournamespace>/process-business-rules-quarkus:latest .
```
It’ll build your image with the kogito service running on it.

Now let’s push the above built image to a public container registry.

`$ podman push quay.io/<yournamespace>/process-business-rules-quarkus:latest`

## Install the Operator

First we would need to enable the OLM in our Minikube cluster. For that, just run:

`$ minikube addons enable olm`

To launch the OLM console locally, clone the [operator-lifecycle-manager](https://github.com/operator-framework/operator-lifecycle-manager) repository and from the root of the project run: `$ make run-console-local`.
 
This will run the operatorhub console on http://localhost:9000 

**Note**: You will need to have [`jq`](https://stedolan.github.io/jq/manual/) installed and 9000 port available on the system.

Create a different namespace where `kogito-operator` and all the dependent operator(s) will run

`$ kubectl create ns kogito`

Now open your browser and visit the OLM console on https://localhost:9000. Select `Operators > OperatorHub` and search for `kogito`.  
Select the Kogito Operator by Red Hat and install it with defaults options, choose the namespace `kogito` which was created for this purpose

You can see the pods by:

`$ watch kubectl get pod -n kogito`

Wait until all pods are in running state (It is installing kogito-operator and all the dependent operator(s) for kogito)

Alternatively, you can install the `Kogito Operator` from [here](https://operatorhub.io/operator/kogito-operator)

## Run Kogito Service with Kogito Operator

Now we’ll create a `KogitoRuntime` object with the image we built and pushed earlier.

```yaml
apiVersion: app.kiegroup.org/v1alpha1
kind: KogitoRuntime
metadata:
  name: process-business-rules-quarkus
spec:
  replicas: 1
  image:
    registry: quay.io
    name: process-business-rules-quarkus
    tag: latest
    namespace: <your-quay-namespace>
```
Create a `myapp.yml` file with the above content.

Just create the above object with

`$ kubectl create -f myapp.yml -n kogito`

You can see your runtime object by 

`$ kubectl get KogitoRuntime -n kogito`

Now expose the service on NodePort so it’s easily accessible.

`$ kubectl expose deployment process-business-rules-quarkus -n <operator-namespace>  --type=NodePort --name=process-business-rules-quarkus`

`$ minikube service process-business-rules-quarkus -n <operator-namspace>`

The above command will open the exposed service in your default browser.