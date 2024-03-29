@resources
@quarkus
@rhpam
Feature: Deploy the service by configuring the resource requests and limits

  Background:
    Given Namespace is created
    And Kogito Operator is deployed

  Scenario Outline: Setting runtime resource requests cpu <runtime-cpu-request>, mem <runtime-memory-request> and limits cpu <runtime-cpu-limit>, mem <runtime-memory-limit>
    Given Clone Kogito examples into local directory
    And Local example service "kogito-quarkus-examples/ruleunit-quarkus-example" is built by Maven and deployed to runtime registry

    When Deploy quarkus example service "ruleunit-quarkus-example" from runtime registry with configuration:
      | runtime-request | cpu    | <runtime-cpu-request>    |
      | runtime-request | memory | <runtime-memory-request> |
      | runtime-limit   | cpu    | <runtime-cpu-limit>      |
      | runtime-limit   | memory | <runtime-memory-limit>   |

    Then Deployment "ruleunit-quarkus-example" has 1 pods with runtime resources within 2 minutes:
      | runtime-request | cpu    | <runtime-cpu-request>    |
      | runtime-request | memory | <runtime-memory-request> |
      | runtime-limit   | cpu    | <runtime-cpu-limit>      |
      | runtime-limit   | memory | <runtime-memory-limit>   |

    Examples:
      | runtime-cpu-request | runtime-memory-request | runtime-cpu-limit | runtime-memory-limit |
      | 500m                | 1Gi                    | 1000m             | 2Gi                  |


  Scenario Outline: Setting build resource requests cpu <build-cpu-request>, mem <build-memory-request> and limits cpu <build-cpu-limit>, mem <build-memory-limit>
    When Build quarkus example service "kogito-quarkus-examples/ruleunit-quarkus-example" with configuration:
      | build-request | cpu    | <build-cpu-request>    |
      | build-request | memory | <build-memory-request> |
      | build-limit   | cpu    | <build-cpu-limit>      |
      | build-limit   | memory | <build-memory-limit>   |
    Then BuildConfig "ruleunit-quarkus-example-builder" is created with build resources within 2 minutes:
      | build-request | cpu    | <build-cpu-request>    |
      | build-request | memory | <build-memory-request> |
      | build-limit   | cpu    | <build-cpu-limit>      |
      | build-limit   | memory | <build-memory-limit>   |

    Examples:
      | build-cpu-request | build-memory-request | build-cpu-limit | build-memory-limit |
      | 500m              | 1Gi                  | 1000m           | 2Gi                |