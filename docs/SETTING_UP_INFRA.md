# Setting up 3rd Party Infrastructure
As of [KOGITO-4199](https://issues.redhat.com/browse/KOGITO-4199), 
the operator no longer sets up 3rd party infrastructure as 
to decouple the Kogito operator functionality from other 
operators as well as allowing the user to setup the 3rd 
party infrastructure more to their needs. As such, there is 
now the onus on the user to setup the 3rd party 
infrastructure, and this document is meant to help guide the 
user to do so.

# Infinispan
## Installing Operator
Currently, Kogito only supports running with Infinispan 11. 
As such, the [Infinispan operator 2.0](https://infinispan.org/docs/infinispan-operator/2.0.x/operator.html) 
must be installed.

### minikube
The instructions for installing Infinispan operator 2.0 on 
minikube can be found on [OperatorHub](https://operatorhub.io/operator/infinispan/2.0.x/infinispan-operator.v2.0.6).

*Note*: The provided YAML in step 2 will install the 
operator into the namespace `my-infinispan`, and it will 
only be usable from there. To change this namespace,
you should download the YAML and fill in your desired 
namespace like in the following YAML:
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: [desired namespace]
---
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: operatorgroup
  namespace: [desired namespace]
spec:
  targetNamespaces:
  - [desired namespace]
---
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: my-infinispan
  namespace: [desired namespace]
spec:
  channel: 2.0.x
  name: infinispan
  source: operatorhubio-catalog
  sourceNamespace: olm
```

Once the operator is installed, you should see a pod named 
`infinispan-operator-*` in your namespace. 

## Setting Up Infinispan Server for Kogito
Kogito requires an Infinispan server named `kogito-infinispan` in the same namespace of the Kogito application. 
An example YAML for this can be found [here](../examples/infinispan-server.yaml). Alternatively, you can 
refer to the [official documentation](https://infinispan.org/infinispan-operator/master/operator.html#creating_minimal_clusters-start) 
and change the name to `kogito-infinispan`.

Once the YAML is deployed, you should see a pod in your 
namespace called `kogito-infinispan-0` (the `0` can vary), 
and running 
`kubectl logs kogito-infinispan-0` should show `Infinispan 
Server 11.* started in...` at the end.
