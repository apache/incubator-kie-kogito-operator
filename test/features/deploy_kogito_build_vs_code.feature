Feature: Deploy Kogito Build - after assets processed by VS code
  

  Background:
    Given Namespace is created
    And Kogito Operator is deployed
    And Project kie-asset-library is cloned
    And Project kie-asset-library is built by maven with configuration:
      | profile | dmn                                    |
      | option  | -Dactive.structures=quarkus            |
      | option  | -Dactive.config.sets=scesim            |
      | native  | disabled                               |
    Then Project "<project>" is generated in temporary folder


  Scenario Outline: Build <runtime> binary build with native <native> using KogitoBuild
    Given Project "<project>" assets are re-marshalled by VS Code
    And Project "<project>" is built from temporary folder by maven
      
    When Build binary <runtime> service "<project>" from kie-asset-library target folder
      | config | native | <native> |

    Then Kogito Runtime "<project>" has 1 pods running within 5 minutes
    And Service "<project>" with decision name "call centre drd" is available within 2 minutes
    
    @rhpam
    @smoke
    @quarkus
    Examples:
      | runtime    | project                  | native   |
      | quarkus    | dmn-generated-quarkus    | disabled |