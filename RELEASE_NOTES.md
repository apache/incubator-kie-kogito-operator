## Enhancements

- [KOGITO-2562](https://issues.redhat.com/browse/KOGITO-2562) - Knative Eventing resources created for Kogito Applications that uses [Knative Addon](https://github.com/kiegroup/kogito-examples/tree/stable/process-knative-quickstart-quarkus).  
- [KOGITO-3624](https://issues.redhat.com/browse/KOGITO-3624) - Grafana dashboard resources created for Kogito Applications that uses [Monitoring Addon](https://blog.kie.org/2020/07/trustyai-meets-kogito-decision-monitoring.html)

## Bug Fixes

## Known Issues
- OptaPlanner applications are not compatible with native builds duo to an incompatibility with Graalvm version 20.2.0 included in Kogito Images. Please see PLANNER-2084 [PLANNER-2084](https://issues.redhat.com/browse/PLANNER-2084) for more details.