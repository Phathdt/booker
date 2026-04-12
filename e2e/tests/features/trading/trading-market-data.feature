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
