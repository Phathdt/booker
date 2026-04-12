# ============================================================================
# Auth Register - E2E Test
# ============================================================================
#
# How to Run:
#   yarn test --tags @auth-register    # Run register tests only
#   yarn test --tags @auth             # Run all auth tests
# ============================================================================

@auth @auth-register
Feature: User Authentication - Register
  As a new user
  I want to create an account on Booker
  So that I can start trading tokens

  @register @smoke @priority_high
  Scenario: Successful registration with valid credentials
    Given I navigate to the login page
    When I switch to the register tab
    And I enter valid registration details
    And I submit the registration form
    Then I should be redirected to the trading page

  @register @priority_medium
  Scenario: Cannot register with short password
    Given I navigate to the login page
    When I switch to the register tab
    And I enter an email and a short password
    And I submit the registration form
    Then I should remain on the login page

  @register @priority_medium
  Scenario: Cannot register with existing email
    Given I navigate to the login page
    When I switch to the register tab
    And I enter an existing user email with valid password
    And I submit the registration form
    Then I should see a registration error message
