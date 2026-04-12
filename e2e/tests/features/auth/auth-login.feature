# ============================================================================
# Auth Login - E2E Test
# ============================================================================
#
# How to Run:
#   yarn test --tags @auth-login       # Run login tests only
#   yarn test --tags @auth             # Run all auth tests
#   yarn test --tags @smoke            # Run smoke tests
# ============================================================================

@auth @auth-login
Feature: User Authentication - Login
  As a Booker user
  I want to log in with my credentials
  So that I can access the trading platform

  @login @smoke @priority_high
  Scenario: Successful login with valid credentials
    Given I navigate to the login page
    When I enter valid login credentials
    And I submit the login form
    Then I should be redirected to the trading page
    And the trading page should display correctly

  @login @priority_high
  Scenario: Failed login with invalid credentials
    Given I navigate to the login page
    When I enter invalid login credentials
    And I submit the login form
    Then I should see a login error message
    And I should remain on the login page

  @login @priority_medium
  Scenario: Cannot login with empty fields
    Given I navigate to the login page
    When I leave the login fields empty
    Then the sign in button should not submit the form
