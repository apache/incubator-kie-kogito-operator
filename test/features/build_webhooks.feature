Feature: Configure Webhooks Triggers in Kogito Build

  Background:
    Given Namespace is created
    And Kogito Operator is deployed

  Scenario Outline: Configure <type> webhook trigger in remote S2I using KogitoBuild
    When Build quarkus example service "process-quarkus-example" with configuration:
      | webhook | type   | <type>    |
      | webhook | secret | <secret>  |
    Then BuildConfig "process-quarkus-example" is created with webhooks within 2 minutes:
      | webhook | type   | <type>    |
      | webhook | secret | <secret>  |

    Examples:
      | type    | secret         | 
      | GitHub  | github_secret  | 
      | Generic | generic_secret | 