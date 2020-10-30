## Enhancements
- [KOGITO-2562](https://issues.redhat.com/browse/KOGITO-2562) - Knative Eventing resources created for Kogito Applications that uses [Knative Addon](https://github.com/kiegroup/kogito-examples/tree/stable/process-knative-quickstart-quarkus).
- [KOGITO-816](https://issues.redhat.com/browse/KOGITO-816) - Operator provisioning for Task Console.
- [KOGITO-1006](https://issues.redhat.com/browse/KOGITO-1006) - Add support to install Task Console to CLI.  
- [KOGITO-3094](https://issues.redhat.com/browse/KOGITO-3094) - Provide data-index-infinispan/mongodb as independent images
     - To switch to Mongodb persistence provider just provide the `quay.io/kiegroup/kogito-data-index-mongodb:1.0.0-snapshot` on the Kogito CR  
- [KOGITO-3624](https://issues.redhat.com/browse/KOGITO-3624) - Grafana dashboard resources created for Kogito Applications that uses [Monitoring Addon](https://blog.kie.org/2020/07/trustyai-meets-kogito-decision-monitoring.html)
- [KOGITO-3273](https://issues.redhat.com/browse/KOGITO-3273) - Research to remove supporting Kogito Services CRD
- [KOGITO-3568](https://issues.redhat.com/browse/KOGITO-3273) - Remove spec.httpPort attribute from Operator and HTTP_PORT env var from Kogito images
- [KOGITO-3376](https://issues.redhat.com/browse/KOGITO-3376) - Update protobuf ConfigMap using fetched files
- [KOGITO-3708](https://issues.redhat.com/browse/KOGITO-3708) - Decreasing the amount of reconciliation loops for `KogitoInfra` resources when
third party infrastructure resources are not yet available
- [KOGITO-3556](https://issues.redhat.com/browse/KOGITO-3556) - Review CRDs Status resource update
 
## Bug Fixes

## Known Issues
The protobuf ConfigMap does not update in Spring Boot due to [this issue](https://issues.redhat.com/browse/KOGITO-3406).
