# ============================================================================
# Trading Execution - E2E Test
# ============================================================================
#
# How to Run:
#   yarn test --tags @trading-execution   # Run execution tests only
#   yarn test --tags @trading             # Run all trading tests
# ============================================================================

@trading @trading-execution
Feature: Trading Order Execution
  As a Booker trader
  I want my buy and sell orders to match and execute
  So that trades are completed on the platform

  @execution @priority_high
  Scenario: Buy and sell orders match and execute
    Given I am logged in to the platform
    And I am on the trading page
    When I select a trading pair
    And I fill in the sell order form with price "40000" and quantity "0.01"
    And I submit the sell order
    Then I should see a success message
    When I fill in the buy order form with price "40000" and quantity "0.01"
    And I submit the buy order
    Then I should see a success message
    And the matching orders should be executed
