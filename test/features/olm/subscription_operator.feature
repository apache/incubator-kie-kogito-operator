@olm
@release
Feature: Deploy Kogito Resources from subscription

  Background:
    Given Namespace is created

  Scenario Outline: Deploy Kogito Operator from subscription
    When Kogito Operator is deployed from subscription using channel "<channel>"
    Then Kogito Operator should be installed from subscription with dependencies

  Examples:
      | channel      |
      | alpha        |
      | dev-preview  |

