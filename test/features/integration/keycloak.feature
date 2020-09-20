@infinispan
@kafka
@keycloak
@security
Feature: Kogito integration with Keycloak

  Background:
    Given Namespace is created
    And Kogito Operator is deployed with Infinispan, Kafka and Keycloak operators

  @dataindex
  Scenario: Install Kogito Data Index with Keycloak security
    Given Keycloak instance with realm "kogito-realm" and client "kogito-dataindex-service" is deployed
    And Keycloak user "my-user" with password "my-password" is deployed

    When Install Kogito Data Index with 1 replicas with configuration:
      | runtime-env  | quarkus.oidc.tenant-enabled                    | true                                           |
      | runtime-env  | quarkus.oidc.tls.verification                  | none                                           |
      | runtime-env  | quarkus.oidc.auth-server-url                   | https://keycloak:8443/auth/realms/kogito-realm |
      | runtime-env  | quarkus.oidc.client-id                         | kogito-dataindex-service                       |
      | runtime-env  | quarkus.http.auth.permission.unsecure.paths    | /health/*                                      |
      | runtime-env  | quarkus.http.auth.permission.unsecure.policy   | permit                                         |
      | runtime-env  | quarkus.http.auth.permission.secure.paths      | /*                                             |
      | runtime-env  | quarkus.http.auth.permission.secure.policy     | authenticated                                  |
    And Kogito Data Index has 1 pods running within 10 minutes
    And Stores access token for user "my-user" and password "my-password" on realm "kogito-realm" and client "kogito-dataindex-service" into variable "my-user-token"

    Then GraphQL request on service "data-index" is successful using access token "{my-user-token}" within 2 minutes with path "graphql" and query:
    """
    {
      ProcessInstances{
        id
      }
    }
    """

#####

  @jobsservice
  Scenario: Install Kogito Jobs Service with Keycloak security
    Given Keycloak instance with realm "kogito-realm" and client "kogito-jobs-service" is deployed
    And Keycloak user "my-user" with password "my-password" is deployed

    When Install Kogito Jobs Service with 1 replicas with configuration:
      | runtime-env  | quarkus.oidc.tenant-enabled                    | true                                           |
      | runtime-env  | quarkus.oidc.tls.verification                  | none                                           |
      | runtime-env  | quarkus.oidc.auth-server-url                   | https://keycloak:8443/auth/realms/kogito-realm |
      | runtime-env  | quarkus.oidc.client-id                         | kogito-jobs-service                            |
      | runtime-env  | quarkus.http.auth.permission.unsecure.paths    | /health/*                                      |
      | runtime-env  | quarkus.http.auth.permission.unsecure.policy   | permit                                         |
      | runtime-env  | quarkus.http.auth.permission.secure.paths      | /*                                             |
      | runtime-env  | quarkus.http.auth.permission.secure.policy     | authenticated                                  |
    And Kogito Jobs Service has 1 pods running within 10 minutes

    Then HTTP GET request on service "jobs-service" with path "jobs" is forbidden within 1 minutes
    
    When Stores access token for user "my-user" and password "my-password" on realm "kogito-realm" and client "kogito-jobs-service" into variable "my-user-token"
    And HTTP POST request on service "jobs-service" using access token "{my-user-token}" is successful within 2 minutes with path "jobs" and body:
      """json
      {
        "id": "1",
        "priority": "1",
        "expirationTime": "2100-01-29T18:19:00Z",
        "callbackEndpoint": "http://localhost:8080/callback"
      }
      """

    Then HTTP GET request on service "jobs-service" using access token "{my-user-token}" with path "jobs/1" is successful within 1 minutes
    