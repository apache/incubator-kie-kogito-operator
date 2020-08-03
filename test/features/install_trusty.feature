@trusty
@infinispan
@kafka
Feature: Kogito Trusty

  Background:
    Given Namespace is created
    And Kogito Operator is deployed with Infinispan and Kafka operators

  @smoke
  Scenario: Install Kogito Trusty
    When Install Kogito Trusty with 1 replicas
    Then Kogito Trusty has 1 pods running within 10 minutes
    
#####

  @externalcomponent
  @infinispan
  @kafka
  Scenario: Install Kogito Trusty with persistence using external Infinispan
    Given Infinispan instance "external-infinispan" is deployed with configuration:
      | username | developer |
      | password | mypass |
    And Kafka instance "external-kafka" is deployed
    When Install Kogito Trusty with 1 replicas with configuration:
      | infinispan | username | developer                 |
      | infinispan | password | mypass                    |
      | infinispan | uri      | external-infinispan:11222 |
      | kafka | externalURI | external-kafka-kafka-bootstrap:9092 |
    Then Kogito Trusty has 1 pods running within 10 minutes
