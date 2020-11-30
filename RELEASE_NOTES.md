## Enhancements  
- [KOGITO-3610](https://issues.redhat.com/browse/KOGITO-3610) Updating Kogito Service interface to describe CloudEvents
- [KOGITO-3754](https://issues.redhat.com/browse/KOGITO-3754) Making KogitoInfra create a truststore secret to hold PKCS certs for infinispan/services connections
- [KOGITO-3874](https://issues.redhat.com/browse/KOGITO-3874) Allow configuration of readiness and liveness probes
- [KOGITO-3215](https://issues.redhat.com/browse/KOGITO-3215) - Support cluster-scoped deployment for operator
- [KOGITO-3937](https://issues.redhat.com/browse/KOGITO-3937) Move probe configuration in to `probes` subproperty
- [KOGITO-3898](https://issues.redhat.com/browse/KOGITO-3898) Provide CLI flags for subset of probe parameters

## Bug Fixes
- [KOGITO-3866](https://issues.redhat.com/browse/KOGITO-3866) Operator print error logs when KogitoRuntime delete

## Known Issues
The protobuf ConfigMap does not update in Spring Boot due to [this issue](https://issues.redhat.com/browse/KOGITO-3406).
