@asset
Feature: Deploy source files (dmn, drl, bpmn, bpmn2, ...) with CLI

  Background:
    Given Namespace is created
    And Kogito Operator is deployed

  @rhpam
  Scenario: Deploy .dmn source files with CLI using runtime <runtime>
    Given Clone Kogito examples into local directory
    
    When Deploy <runtime> file "Traffic Violation.dmn" from example service "dmn-<runtime>-example" with env:
      | MAVEN_ARGS_APPEND | <maven_args> |
    And Build "dmn-<runtime>-example-builder" is complete after 10 minutes
    And Build "dmn-<runtime>-example" is complete after 5 minutes

    Then HTTP POST request on service "dmn-<runtime>-example" is successful within 2 minutes with path "Traffic Violation" and body:
      """json
      {
          "Driver":{"Points":2},
          "Violation":{
              "Type":"speed",
              "Actual Speed":120,
              "Speed Limit":100
          }
      }
      """

    # Running just SpringBoot for smoke tests as quarkus is executed for directory deployment below
    @smoke
    @springboot
    Examples:
      | runtime    | maven_args           |
      | springboot | -Dstarters=decisions |

    @quarkus
    Examples:
      | runtime    | maven_args |
      | quarkus    |            |
  
  Scenario: Deploy .bpmn source files with CLI
    Given Clone Kogito examples into local directory
    
    When Deploy quarkus file "org/acme/travels/scripts.bpmn" from example service "process-scripts-quarkus"
    And Build "process-scripts-quarkus-builder" is complete after 10 minutes
    And Build "process-scripts-quarkus" is complete after 5 minutes

    Then HTTP POST request on service "process-scripts-quarkus" is successful within 2 minutes with path "scripts" and body:
      """json
      {
          "name" : "john"
      }
      """

  @smoke
  @rhpam
  Scenario: Deploy source files in folder with CLI
    Given Clone Kogito examples into local directory
    
    When Deploy quarkus folder from example service "dmn-quarkus-example"
    And Build "dmn-quarkus-example-builder" is complete after 10 minutes
    And Build "dmn-quarkus-example" is complete after 5 minutes

    Then HTTP POST request on service "dmn-quarkus-example" is successful within 2 minutes with path "Traffic Violation" and body:
      """json
      {
          "Driver":{"Points":2},
          "Violation":{
              "Type":"speed",
              "Actual Speed":120,
              "Speed Limit":100
          }
      }
      """
