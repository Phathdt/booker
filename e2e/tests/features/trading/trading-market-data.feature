# ============================================================================
# Trading Market Data - E2E Test
# ============================================================================
#
# How to Run:
#   yarn test --tags @trading-market     # Run market data tests only
#   yarn test --tags @trading            # Run all trading tests
# ============================================================================

@trading @trading-market
Feature: Trading Market Data Display
  As a Booker trader
  I want to see live market data on the trading page
  So that I can make informed trading decisions

  @market @smoke @priority_high
  Scenario: Trading page displays ticker bar and market sections
    Given I am logged in to the platform
    And I am on the trading page
    Then the ticker bar should be visible
    And the order book should be visible
    And the recent trades section should be visible
    And the open orders section should be visible

  @market @priority_high
  Scenario: Switching pairs updates all market components
    Given I am logged in to the platform
    And I am on the trading page
    When I select trading pair "ETH_USDT"
    Then the ticker bar should be visible
    And the order book should be visible
    And the recent trades section should be visible

  @market @priority_high
  Scenario: Order book shows bids and asks after placing orders
    Given I am logged in as "trader11@booker.dev"
    And I am on the trading page
    When I select a trading pair
    And I fill in the sell order form with price "48000" and quantity "0.01"
    And I submit the sell order
    Then I should see a success message
    When I logout and login as "trader12@booker.dev"
    And I am on the trading page
    And I select a trading pair
    And I fill in the buy order form with price "47000" and quantity "0.02"
    And I submit the buy order
    Then I should see a success message
    And the order book should show bid and ask levels
