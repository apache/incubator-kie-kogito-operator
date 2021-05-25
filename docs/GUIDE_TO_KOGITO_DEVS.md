# Guide for Core/Runtimes team to Smoke test local changes on Openshift/k8s Cluster.

This guide aims to help core/runtimes developer to test their local changes with respect to the operator.
The operator will be deployed over an Openshift or Kubernetes cluster. In this example we've taken minikube to deploy a Kubernetes cluster.

# Table of contents

* [Guide for Core/Runtimes team to Smoke test local changes on Openshift/k8s Cluster.](#guide-for-coreruntimes-team-to-smoke-test-local-changes-on-openshiftk8s-cluster)
* [Table of contents](#table-of-contents)
  * [Install Minikube](#install-minikube)
    * [Prerequisites](#prerequisites)
  * [Build Artifacts](#build-artifacts)
  * [Build the Kogito Application Images](#build-the-kogito-application-images)
  * [Build the Kogito Service Image](#build-the-kogito-service-image)
    * [Base Images](#base-images)
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
mvn clean package -DskipTests #(For skipping tests) -DskipTestsIT #(For verification)
```

This will deploy your jar's in the target directory. For example, for jobs-service the jar would present at `/path/to/kogito-apps/jobs-service/jobs-service-common/target/jobs-service-common-{VERSION}-runner.jar`

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
$ cp /path/to/kogito-apps/jobs-service/jobs-service-common/target/jobs-service-common-{VERSION}-runner.jar .
```

Also, inside this directory, you can have a `Dockerfile` which can be used to update the artifact.

```shell-script
$ vim jobs-service-Dockerfile
``` 

```Dockerfile
FROM quay.io/kiegroup/kogito-jobs-service-ephemeral
COPY jobs-service-common-{VERSION}-runner.jar ${KOGITO_HOME}/bin/jobs-service-common-runner.jar
```
**Note**: Usually, latest artifacts are with version  `8.0.0-SNAPSHOT`
**Note**: You can check the location of the artifact in the image from the [kogito-images](https://github.com/kiegroup/kogito-images) repository.

Now you are going to build the image. Inside the above created directory, run:

```shell-script
$ podman build -t <prefered-tag-name>  -f jobs-service-Dockerfile  .
```

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
FROM quay.io/kiegroup/kogito-runtime-jvm:latest
ENV RUNTIME_TYPE quarkus

COPY target/quarkus-app/lib/ $KOGITO_HOME/bin/lib/
COPY target/quarkus-app/*.jar $KOGITO_HOME/bin
COPY target/quarkus-app/app/ $KOGITO_HOME/bin/app/
COPY target/quarkus-app/quarkus/ $KOGITO_HOME/bin/quarkus/
```

```shell-script
$ podman build --tag quay.io/<yournamespace>/process-business-rules-quarkus:latest .
```
It’ll build your image with the kogito service running on it.

Now let’s push the above built image to a public container registry.

```shell-script
$ podman push quay.io/<yournamespace>/process-business-rules-quarkus:latest
```

## Install the Operator

First we would need to enable the OLM in our Minikube cluster. For that, just run:

```shell-script
$ minikube addons enable olm
```
After enabling the olm please wait till all pods in `olm` namepspace are in running state.

```shell-script
$ watch kubectl get pods -n olm
```
**Note**: In some cases the `operator-catalog` pod state changes to `CrashBackLoopOff`, due to the failure in `readiness probe`. If that happens, please delete the olm namespace `$ kubectl delete namespace olm` and try installing the olm with this command. `$ curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.15.1/install.sh | bash -s 0.15.1` 

To launch the OLM console locally, run the following command:

 ```shell-script
$ curl -L https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/master/scripts/run_console_local.sh | bash
```

This will run the operatorhub console on http://localhost:9000 

**Note**: You will need to have [`jq`](https://stedolan.github.io/jq/manual/) and [`go`](https://golang.org/dl/) installed and 9000 port available on the system.

Create a different namespace where `kogito-operator` and all the dependent operator(s) will run

```shell-script
$ kubectl create ns kogito
```

Now open your browser and visit the OLM console on http://localhost:9000. Select `Operators > OperatorHub` and search for `kogito`.  
Select the Kogito Operator by Red Hat and install it with defaults options, choose the namespace `kogito` which was created for this purpose

You can see the pods by:

```shell-script
$ watch kubectl get pod -n kogito
```

Wait until all pods are in running state (It is installing kogito-operator and all the dependent operator(s) for kogito)

Alternatively, you can install the `Kogito Operator` from [here](https://operatorhub.io/operator/kogito-operator)

## Run Kogito Service with Kogito Operator

Now we’ll create a `KogitoRuntime` object with the image we built and pushed earlier.

```yaml
apiVersion: app.kiegroup.org/v1beta1
kind: KogitoRuntime
metadata:
  name: process-business-rules-quarkus
spec:
  replicas: 1
  image: quay.io/<your-quay-namespace>/process-business-rules-quarkus:latest
```
Create a `myapp.yml` file with the above content.

Just create the above object with

```shell-script
$ kubectl create -f myapp.yml -n kogito
```

You can see your runtime object by 

```shell-script
$ kubectl get KogitoRuntime -n kogito
```

Now expose the service on NodePort so it’s easily accessible.

```shell-script
$ kubectl expose deployment process-business-rules-quarkus -n kogito  --type=NodePort --name=process-business-rules-quarkus-np
$ minikube service process-business-rules-quarkus-np -n kogito
```
The above commands will expose the service on `nodePort`  and open the exposed service in your default browser.
