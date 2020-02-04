@cr
@dataindex
Feature: Install Kogito Data Index Service

  Background:
    Given Namespace is created
    And Kogito Operator is deployed with dependencies

  Scenario: Install Kogito Data Index Service
    When Install Kogito Data Index with 1 replicas

    Then Kogito Data Index has 1 pods running within 5 minutes
    And GraphQL request on service "kogito-data-index" is successful within 2 minutes with path "graphql" and query:
    """
    {
      ProcessInstances{
        id
      }
    }
    """

#####

  @events
  @persistence
  Scenario Outline: Process instance events are stored in Data Index
    Given Install Kogito Data Index with 1 replicas
    And Deploy quarkus example service "jbpm-quarkus-example" with persistence enabled and native <native> and events "enabled"
    And Kogito application "jbpm-quarkus-example" has 1 pods running within <minutes> minutes
    And HTTP GET request on service "jbpm-quarkus-example" with path "orders" is successful within 3 minutes

    When HTTP POST request on service "jbpm-quarkus-example" with path "orders" and body:
      """json
      {
        "approver" : "john", 
        "order" : {
          "orderNumber" : "12345", 
          "shipped" : false
        }
      }
      """

    Then GraphQL request on Data Index service returns ProcessInstances processName "orders" within 2 minutes

    Examples:
      | native | minutes |
      # Commented as long as https://issues.redhat.com/browse/KOGITO-842 is not solved
      #| "enabled" | 20 |
      | "disabled" | 10 |