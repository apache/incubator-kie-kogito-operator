@springboot
Feature: Deploy spring boot service

  Background:
    Given Namespace is created

  Scenario Outline: Deploy jbpm-springboot-example service without persistence
    Given Kogito Operator is deployed
    
    When "<installer>" deploy spring boot example service "jbpm-springboot-example"

    Then Kogito application "jbpm-springboot-example" has 1 pods running within 10 minutes
    And HTTP GET request on service "jbpm-springboot-example" with path "orders" is successful within 2 minutes

    @cr
    Examples: CR
      | installer |
      | CR        |

    @smoke
    @cli
    @wip
    Examples: CLI
      | installer |
      | CLI       |

#####

  # Disabled because of https://issues.redhat.com/browse/KOGITO-948
  @disabled
  @persistence
  Scenario Outline: Deploy jbpm-springboot-example service with persistence
    Given Kogito Operator is deployed with Infinispan operator
    And "<installer>" deploy spring boot example service "jbpm-springboot-example" with persistence
    And Kogito application "jbpm-springboot-example" has 1 pods running within 10 minutes
    And HTTP GET request on service "jbpm-springboot-example" with path "orders" is successful within 3 minutes
    And HTTP POST request on service "jbpm-springboot-example" with path "orders" and body:
      """json
      {
        "approver" : "john", 
        "order" : {
          "orderNumber" : "12345", 
          "shipped" : false
        }
      }
      """
    And HTTP GET request on service "jbpm-springboot-example" with path "orders" should return an array of size 1 within 1 minutes
    
    When Scale Kogito application "jbpm-springboot-example" to 0 pods within 2 minutes
    And Scale Kogito application "jbpm-springboot-example" to 1 pods within 2 minutes
    
    Then HTTP GET request on service "jbpm-springboot-example" with path "orders" should return an array of size 1 within 2 minutes

    @cr
    Examples: CR
      | installer |
      | CR        |

    @cli
    Examples: CLI
      | installer |
      | CLI       |
#####

  # Disabled as long as https://issues.redhat.com/browse/KOGITO-1163 and https://issues.redhat.com/browse/KOGITO-1166 is not solved
  @disabled
  @cr
  @jobsservice
  Scenario: CR deploy timer-springboot-example service with Jobs service
    Given Kogito Operator is deployed
    And "CR" install Kogito Jobs Service with 1 replicas
    And "CR" deploy spring boot example service "timer-springboot-example"
    And Kogito application "timer-springboot-example" has 1 pods running within 10 minutes

    When HTTP POST request on service "timer-springboot-example" is successful within 2 minutes with path "timer" and body:
      """json
      { }
      """

    # Implement retrieving of job information from Jobs service once https://issues.redhat.com/browse/KOGITO-1163 is fixed
    Then Kogito application "timer-springboot-example" log contains text "Before timer" within 1 minutes
    And Kogito application "timer-springboot-example" log contains text "After timer" within 1 minutes
