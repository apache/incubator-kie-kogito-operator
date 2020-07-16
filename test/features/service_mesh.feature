Feature: Service Deployment: Service Mesh

  Background:
    Given Namespace is created 
    And Elasticsearch Operator is deployed
    And Jaeger Operator is deployed
    And Service Mesh Operator is deployed

  Scenario Outline: Deploy Example with Istio enabled
    Given Kogito Operator is deployed
    And Service Mesh instance is deployed
    And Clone Kogito examples into local directory
    And Local example service "<example-service>" is built by Maven using profile "<profile>" and deployed to runtime registry

    When Deploy <runtime> example service "<example-service>" from runtime registry with configuration:
      | istio | enabled | true |

    Then Kogito Runtime "<example-service>" has 1 pods running within 10 minutes
    And Kogito Runtime "<example-service>" is configured with Istio

    @springboot
    Examples:
      | runtime    | example-service            | profile |
      | springboot | process-springboot-example | default |