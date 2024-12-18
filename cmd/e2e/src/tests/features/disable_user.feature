Feature: Disable/Enable BHCE Users

  Background:
    Given Create a new user with "User" role with disabled status

  @e2e
  Scenario: User redirected to the disabled user page when disabled
    Given User visits the login page
    And User enters valid email
    And User enters valid password
    When User clicks on the login button
    Then login page displays "Your Account has been Disabled"
    Then login page displays "Please contact your system administrator for assistance"
