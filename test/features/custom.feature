Feature: Kogito In Service Mesh

  Background:
    Given Namespace is created

  Scenario: Install/Remove Kogito Infra Keycloak
    Given Kogito Operator is deployed with Keycloak operator
    When Install Kogito Infra "Keycloak"
    Then Kogito Infra "Keycloak" should be running within 10 minutes