# Guide for Core/Runtimes team to Smoke test local changes on Openshift/k8s Cluster.

This guide aims to help core/runtime developer to test there local changes with respect to the operator.
The operator will be deployed over a openshift/kubernetes cluster. In this example we've taken minikube to deploy the kubernetes cluster.
# Table of contents

* [Install Minikube](#install-minikube)
* [Build Artifacts](#build-artifacts)
* [Build the Kogito Application Images](#build-the-kogito-application-images)
* [Build the Kogito Service Image](#build-the-kogito-service-image)
* [Install the Operator](#install-the-operator)
* [Run Kogito Service with Kogito Operator](#run-kogito-service-with-kogito-operator)

## Install Minikube

We decided to go with Minikube as the k8s cluster as it is very resource efficient and can be started easily on any system.
 ### Prerequisites
  * Install kubectl binaries on your system [see](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
  * Have a hypervisor installed (kvm  is recommended for linux)

Alternatively, you can install it using podman or docker driver which will launch the minikube cluster inside the containers.
Finally for installing the Minikube cluster please follow [this tutorial](https://kubernetes.io/docs/tasks/tools/install-minikube/)

## Build Artifacts
Before proceeding read the README files on the following repos to know the dependencies you might need to install to build and deploy the artifacts

For generating the jar artifacts you can use, `mvn clean install` to have them locally on your system.

So for every repository, [runtimes](https://github.com/kiegroup/kogito-runtimes) / [apps](https://github.com/kiegroup/kogito-apps) / [examples](https://github.com/kiegroup/kogito-examples). From the root of the repository you can just run:

`mvn clean install -DskipTests(For skipping tests)` 
This will deploy your artifacts in `~/.m2/repository/org/kie/kogito`. You can just build the artifacts you modified or added new feature in and update them in the Images

## Build the Kogito Application Images
We’re going to use our pre-build images which are available in the `quay.io/kiegroup/` namespace.

 ### Images list
  * kogito-data-index
  * kogito-jobs-service
  * kogito-management-console

Now let’s take an example of updating the kogito-jobs-service artifact in the image.
Create a directory e.g: `~/images/`
Inside the directory copy the artifacts you need to update,
`# cp ~/.m2/repository/org/kie/kogito/jobs-service/8.0.0-SNAPSHOT/jobs-service-8.0.0-SNAPSHOT-runner.jar .`
Inside the directory you can have a Dockerfile which can be used to update the artifact.

` $ vim jobs-service-Dockerfile` 
```
FROM quay.io/kiegroup/kogito-jobs-service
COPY jobs-service-8.0.0-SNAPSHOT-runner.jar ${KOGITO_HOME}/bin/kogito-jobs-service-runner.jar
```
Note: Normally all the latest artifacts would be in 8.0.0-SNAPSHOT
Note: You can check the location of the artifact in the image from the [kogito-images](https://github.com/kiegroup/kogito-images) repository.

Now just build your image. Inside the above created directory run:

`$ podman build -t <prefered-tag-name>  -f jobs-service-Dockerfile  .`

The above command will build your image with <prefered-tag-name> and can be seen in the output of `$ podman images`

Note: We added the -f flag to specify the Dockerfile in case you want to build multiple images, you can then have different files and specify them with the -f flag.
This image needs to be on a public container registry from which it can be pulled later.
```
Run: podman tag <prefered-tag-name> quay.io/<your-namespace>/<image-name>
Eg: podman tag jobs-service quay.io/myorg/jobs-service:latest
$ podman push quay.io/<your-namespace>/<image-name>
```

## Build the Kogito Service Image

 ### Base Images
  * kogito-quarkus-ubi8
  * kogito-quarkus-jvm-ubi8
  * kogito-springboot-ubi8

As an example for building the Image for your Kogito custom Service we’re going to build an example from [kogito-examples](https://github.com/kiegroup/kogito-examples) of `process-business-rules-quarkus`

```
cd kogito-examples/process-business-rules-quarkus
mvn clean package
```
Now create a Dockerfile inside the `kogito-examples/process-business-rules-quarkus`

With the following content:

```
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
First we would need to enable the OLM in our Minikube cluster for that just run:

`$ minikube addons enable olm`

Launch the OLM console locally, 
Just clone the [operator-lifecycle-manager](https://github.com/operator-framework/operator-lifecycle-manager) repo and from the root of the project run: `$ make run-console-local`.
 
 This will run the operatorhub console on http://localhost:9000 

Note: Needs to have [`jq`](https://stedolan.github.io/jq/manual/) installed and 9000 port available on the system.

Create a different namespace where kogito-operator and all the dependent operator will run

`$ kubectl create ns kogito`

Now on your browser and visit  https://localhost:9000. Select `Operators > OperatorHub` and search for kogito. 
Select the kogito operator by Redhat and install it with defaults only changing the namespace where it needs to be installed. You can select the `kogito` namespace which was created for this purpose.

You can see the pods by:

`$ kubectl get po -n kogito`

Wait till all pods are in running state (It is installing kogito-operator and all the dependent operator for kogito)


Alternatively you if you are having trouble in having the operators shown in the OperatorHub.
You can also install it from [here](https://operatorhub.io/operator/kogito-operator)

## Run Kogito Service with Kogito Operator

Now we’ll create a `KogitoRuntime` object with the Application image that we built earlier.

```
apiVersion: app.kiegroup.org/v1alpha1
kind: KogitoRuntime
metadata:
  name: process-business-rules-quarkus
spec:
  replicas: 1
  image:
    namespace: <your-quay-namespace>
```
Create a myapp.yml with the above content.

Just create the above object with

`$ kubectl create -f myapp.yml -n <operator-namespace>`

You can see your runtime object by 

`$ kubectl get KogitoRuntime -n <operator-namespace>`

Now expose the service on NodePort so it’s easily accessible.

`$ kubectl expose deployment process-business-rules-quarkus -n <operator-namespace>  --type=NodePort --name=process-business-rules-quarkus`

`$ minikube service process-business-rules-quarkus -n <operator-namspace>`

