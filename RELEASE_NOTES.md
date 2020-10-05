## Enhancements

- [KOGITO-3270](https://issues.redhat.com/browse/KOGITO-3270) - Operator should support ConfigMap injecting in KogitoServices
- [KOGITO-2905](https://issues.redhat.com/browse/KOGITO-2905) - Change in image definition in the CRDs
- [KOGITO-3039](https://issues.redhat.com/browse/KOGITO-3039) - Agnostic infrastructure layer configuration on Operator
- [KOGITO-2113](https://issues.redhat.com/browse/KOGITO-2113) - Simplify binary build deployment
- [KOGITO-3094](https://issues.redhat.com/browse/KOGITO-3094) - Enable persistence selection in Data Index controller


## Bug Fixes

- [KOGITO-3382](https://issues.redhat.com/browse/KOGITO-3382) - Transitive dependency of Keycloak operator module missing


## Known Issues
- OptaPlanner applications are not compatible with native builds duo to an incompatibility with Graalvm version 20.2.0 included in Kogito Images. Please see PLANNER-2084 [PLANNER-2084](https://issues.redhat.com/browse/PLANNER-2084) for more details.