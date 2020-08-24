# Disabled until OLM dev deployment is in place => https://issues.redhat.com/browse/KOGITO-940
@disabled
@olm
@cli
Feature: CLI: Project

  Scenario: CLI create project    
    When CLI create namespace

    Then Kogito operator should be installed with dependencies

#####

  Scenario: CLI create project with Kogito Data Index   
    When CLI create namespace with Kogito Data Index

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
  Scenario: CLI create new project with Kogito Jobs Service   
    When CLI create namespace with Kogito Jobs Service

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

#####

  Scenario: CLI create project with Kogito Trusty   
    When CLI create namespace with Kogito Trusty

    Then Kogito Trusty has 1 pods running within 5 minutes
    And HTTP GET request on service "trusty" is successful within 2 minutes with path "executions"