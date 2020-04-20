@quarkus
Feature: Deploy quarkus service

  Background:
    Given Namespace is created

  Scenario Outline: Deploy drools-quarkus-example service without persistence and native <native>
    Given Kogito Operator is deployed
    When Deploy quarkus example service "ruleunit-quarkus-example" with configuration:
      | config | native | <native> |
    Then Kogito application "ruleunit-quarkus-example" has 1 pods running within <minutes> minutes
    And HTTP POST request on service "ruleunit-quarkus-example" is successful within 2 minutes with path "find-approved" and body:
       """json
       {
         "maxAmount" : 5000,
         "loanApplications": [{ 
            "id" : "ABC10001",
            "amount" : 2000,
            "deposit" : 100,
            "applicant": {
              "age" : 45,
              "name" : "John"
            }
          }, {
            "id" : "ABC10002",
            "amount" : 5000,
            "deposit" : 100,
            "applicant" : {
              "age" : 25,
              "name" : "Paul"
            }
          }, {
            "id" : "ABC10015",
            "amount" : 1000,
            "deposit" : 100,
            "applicant" : {
              "age" : 12,
              "name" : "George"
            }
          }]
        }
      """

    @smoke  
    Examples:
      | native   | minutes |
      | disabled | 10      |

    @native
    Examples:
      | native  | minutes |
      | enabled | 20      |

#####

  @persistence
  Scenario Outline: Deploy process-quarkus-example service with persistence and native <native>
    Given Kogito Operator is deployed with Infinispan operator
    And Deploy quarkus example service "process-quarkus-example" with configuration:
      | config | native      | <native> |
      | config | persistence | enabled  |
    And Kogito application "process-quarkus-example" has 1 pods running within <minutes> minutes
    And HTTP GET request on service "process-quarkus-example" with path "orders" is successful within 3 minutes
    And HTTP POST request on service "process-quarkus-example" with path "orders" and body:
      """json
      {
        "approver" : "john", 
        "order" : {
          "orderNumber" : "12345", 
          "shipped" : false
        }
      }
      """
    And HTTP GET request on service "process-quarkus-example" with path "orders" should return an array of size 1 within 1 minutes
    
    When Scale Kogito application "process-quarkus-example" to 0 pods within 2 minutes
    And Scale Kogito application "process-quarkus-example" to 1 pods within 2 minutes
    
    Then HTTP GET request on service "process-quarkus-example" with path "orders" should return an array of size 1 within 2 minutes
    
    Examples:
      | native   | minutes |
      | disabled | 10      |

    @native
    Examples:
      | native  | minutes |
      | enabled | 20      |

#####

  # Disabled as long as https://issues.redhat.com/browse/KOGITO-1163 and https://issues.redhat.com/browse/KOGITO-1166 is not solved
  @disabled
  @jobsservice
  Scenario Outline: Deploy process-timer-quarkus service with Jobs service and native <native>
    Given Kogito Operator is deployed
    And Install Kogito Jobs Service with 1 replicas
    And Deploy quarkus example service "process-timer-quarkus" with configuration:
      | config | native | <native> |
    And Kogito application "process-timer-quarkus" has 1 pods running within <minutes> minutes

    When HTTP POST request on service "process-timer-quarkus" is successful within 2 minutes with path "timer" and body:
      """json
      { }
      """

    # Implement retrieving of job information from Jobs service once https://issues.redhat.com/browse/KOGITO-1163 is fixed
    Then Kogito application "process-timer-quarkus" log contains text "Before timer" within 1 minutes
    And Kogito application "process-timer-quarkus" log contains text "After timer" within 1 minutes

    Examples:
      | native   | minutes |
      | disabled | 10      |

    # Disabled as long as https://issues.redhat.com/browse/KOGITO-1179 is not solved
    @disabled
    @native
    Examples:
      | native  | minutes |
      | enabled | 20      |

#####

  @events
  @persistence
  Scenario Outline: Data Index retrieves Quarkus process' events with native <native>
    Given Kogito Operator is deployed with Infinispan and Kafka operators
    And Install Kogito Data Index with 1 replicas
    And Deploy quarkus example service "process-quarkus-example" with configuration:
      | config | native      | <native> |
      | config | persistence | enabled  |
      | config | events      | enabled  |
    And Kogito application "process-quarkus-example" has 1 pods running within <minutes> minutes
    And HTTP GET request on service "process-quarkus-example" with path "orders" is successful within 3 minutes

    When HTTP POST request on service "process-quarkus-example" with path "orders" and body:
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
      | native   | minutes |
      | disabled | 10      |

    @native
    Examples:
      | native  | minutes |
      | enabled | 20      |
