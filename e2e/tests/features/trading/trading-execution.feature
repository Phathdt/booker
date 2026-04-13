# ============================================================================
# Trading Execution - E2E Test
# ============================================================================
#
# How to Run:
#   yarn test --tags @trading-execution   # Run execution tests only
#   yarn test --tags @trading             # Run all trading tests
#
# Note: Uses 2 different users to avoid self-trade prevention.
#   trader1@booker.dev places a sell order
#   trader2@booker.dev places a buy order at the same price → match
# ============================================================================

@trading @trading-execution
Feature: Trading Order Execution
  As a Booker trader
  I want my buy and sell orders to match with other traders
  So that trades are completed on the platform

  @execution @priority_high
  Scenario: Buy and sell orders from different users match and execute
    # Trader 1 places a sell order
    Given I am logged in as "trader1@booker.dev"
    And I am on the trading page
    When I select a trading pair
    And I fill in the sell order form with price "42000" and quantity "0.01"
    And I submit the sell order
    Then I should see a success message

    # Trader 2 places a matching buy order
    When I logout and login as "trader2@booker.dev"
    And I am on the trading page
    And I select a trading pair
    And I fill in the buy order form with price "42000" and quantity "0.01"
    And I submit the buy order
    Then I should see a success message
    And the buy order should be filled
