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
  Scenario: Install Kogito Trusty with persistence using external Infinispan
    Given Infinispan instance "external-infinispan" is deployed with configuration:
      | username | developer |
      | password | mypass |

    When Install Kogito Trusty with 1 replicas with configuration:
      | infinispan | username | developer                 |
      | infinispan | password | mypass                    |
      | infinispan | uri      | external-infinispan:11222 |

    Then Kogito Trusty has 1 pods running within 10 minutes

# External Kafka testing is covered in deploy_quarkus_service and deploy_springboot_service as it checks integration between Trusty and KogitoApp
