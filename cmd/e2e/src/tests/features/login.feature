Feature: BHE Login Page Tests

  Background:
    Given Create a new user with "Administrator" role 
    And User navigates to the login page

  @e2e  
  Scenario: User Authenticate with valid credentials
    And User enters valid username
    And User enters valid password
    When User clicks on the login button
    Then User is redirect to "explore" page

  @e2e 
  Scenario: User Authenticates with invalid credentials
    And User enters valid username
    And User enters invalid password
    When User clicks on the login button
    Then Page Displays Error Message
    
  @e2e 
  Scenario: Unauthenticated user accessing pages that require authentication
    When User visits "my-profile" page
    Then User is redirect back to "login" page
  @e2e 
  Scenario: User Authenticate with valid credentials
    And User enters valid username
    And User enters valid password
    When User clicks on the login button
    When User reloads the "explore" page
    Then User should be logged
