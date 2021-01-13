## Enhancements  
- [KOGITO-3610](https://issues.redhat.com/browse/KOGITO-3610) Updating Kogito Service interface to describe CloudEvents
- [KOGITO-3754](https://issues.redhat.com/browse/KOGITO-3754) Making KogitoInfra create a truststore secret to hold PKCS certs for infinispan/services connections
- [KOGITO-3874](https://issues.redhat.com/browse/KOGITO-3874) Allow configuration of readiness and liveness probes
- [KOGITO-3215](https://issues.redhat.com/browse/KOGITO-3215) Support cluster-scoped deployment for operator
- [KOGITO-3937](https://issues.redhat.com/browse/KOGITO-3937) Move probe configuration in to `probes` subproperty
- [KOGITO-3898](https://issues.redhat.com/browse/KOGITO-3898) Provide CLI flags for subset of probe parameters
- [KOGITO-3968](https://issues.redhat.com/browse/KOGITO-3968) Add support to build and deploy sw.json and sw.yaml files from source
- [KOGITO-2759](https://issues.redhat.com/browse/KOGITO-2759) Kogito operator support for MongoDB persistence

## Bug Fixes
- [KOGITO-3866](https://issues.redhat.com/browse/KOGITO-3866) Operator print error logs when KogitoRuntime delete
- [KOGITO-3864](https://issues.redhat.com/browse/KOGITO-3864) KogitoInfra gets deleted when refering KogitoRuntime/KogitoSupportingService delete
- [KOGITO-3947](https://issues.redhat.com/browse/KOGITO-3947) Infinite reconciliation loop when Jobs service is deployed
- [KOGITO-4046](https://issues.redhat.com/browse/KOGITO-4046) Deploying a new Kogito service from files in CLI not reading all files in the directory

## Known Issues
The protobuf ConfigMap does not update in Spring Boot due to [this issue](https://issues.redhat.com/browse/KOGITO-3406).

Default values for Quarkus probes and metrics URL were changed for Quarkus 1.11 [KOGITO-4173](https://issues.redhat.com/browse/KOGITO-4173)