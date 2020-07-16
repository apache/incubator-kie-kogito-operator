Feature: Deploy Kogito Runtime

  Background:
    Given Namespace is created

  @smoke
  Scenario: Install Kogito Jobs Service without persistence
    Given Kogito Operator is deployed

    When Install Kogito Jobs Service with 1 replicas
    And Kogito Jobs Service has 1 pods running within 10 minutes
    And HTTP POST request on service "jobs-service" is successful within 2 minutes with path "jobs" and body:
      """json
      {
        "id": "1",
        "priority": "1",
        "expirationTime": "2100-01-29T18:19:00Z",
        "callbackEndpoint": "http://localhost:8080/callback"
      }
      """

    Then HTTP GET request on service "jobs-service" with path "jobs/1" is successful within 1 minutes

#####

  @security
  Scenario: Install Kogito Jobs Service with Keycloak security
    Given Kogito Operator is deployed with Keycloak operator
    And Keycloak instance with realm "kogito-realm" and client "kogito-jobs-service" is deployed
    And Keycloak user "my-user" with password "my-password" is deployed
    And Stores Keycloak endpoint into variable "auth-server-url"

    When Install Kogito Jobs Service with 1 replicas with configuration:
      | runtime-env  | quarkus.oidc.tenant-enabled                          | true                                             |
      | runtime-env  | quarkus.oidc.tls.verification                        | none                                             |
      | runtime-env  | quarkus.oidc.auth-server-url                         | {auth-server-url}/auth/realms/kogito-realm       |
      | runtime-env  | quarkus.oidc.client-id                               | kogito-jobs-service                              |
      | runtime-env  | quarkus.http.auth.permission.secure.paths            | /jobs*                                           |
      | runtime-env  | quarkus.http.auth.permission.secure.policy           | authenticated                                    |
    And Kogito Jobs Service has 1 pods running within 10 minutes

    Then HTTP GET request on service "jobs-service" with path "jobs" is forbidden within 1 minutes
    
    When Stores access token for user "my-user" and password "my-password" on realm "kogito-realm" and client "kogito-jobs-service" using server "{auth-server-url}" into variable "my-user-token"
    When HTTP POST request on service "jobs-service" using access token "{my-user-token}" is successful within 2 minutes with path "jobs" and body:
      """json
      {
        "id": "1",
        "priority": "1",
        "expirationTime": "2100-01-29T18:19:00Z",
        "callbackEndpoint": "http://localhost:8080/callback"
      }
      """
    Then HTTP GET request on service "jobs-service" using access token "{my-user-token}" with path "jobs/1" is successful within 1 minutes
    