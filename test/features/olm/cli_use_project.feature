# Disabled until OLM dev deployment is in place => https://issues.redhat.com/browse/KOGITO-940
@disabled
@olm
@cli
Feature: CLI: Project

  Background:
    Given Namespace is created

  Scenario: CLI use project
    When CLI use namespace

    Then Kogito Operator should be installed with dependencies

#####

  Scenario: CLI use project and set Kogito Data Index
    When CLI use namespace with Kogito Data Index

    Then Kogito Data Index has 1 pods running within 5 minutes
    And GraphQL request on service "data-index" is successful within 2 minutes with path "graphql" and query:
      """
      {
        ProcessInstances{
          id
        }
      }
      """

#####

  # Disabled until https://issues.redhat.com/browse/KOGITO-910 has been implemented
  @disabled
  Scenario: CLI use project and set Kogito Jobs Service (use-project --install-data-index)
    When CLI use namespace with Kogito Jobs Service

    Then Kogito Jobs Service has 1 pods running within 5 minutes
    And HTTP POST request on service "jobs-service" is successful within 2 minutes with path "jobs" and body:
      """
      { 
        "id": "1",
        "priority": "1",
        "expirationTime": "2100-01-29T18:19:00Z",
        "callbackEndpoint": "http://localhost:8080/callback"
      }
      """

    Then HTTP GET request on service "jobs-service" with path "jobs/1" is successful within 1 minutes