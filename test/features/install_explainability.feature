@explainability
@kafka
Feature: Kogito Data Index

  Background:
    Given Namespace is created
    And Kogito Operator is deployed with Kafka operator

  @smoke
  Scenario: Install Kogito Explainability
    When Install Kogito Explainability with 1 replicas
    Then Kogito Explainability has 1 pods running within 10 minutes
    And HTTP GET request on service "explainability" is successful within 2 minutes with path "health/live"
