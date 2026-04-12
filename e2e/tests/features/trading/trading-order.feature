# ============================================================================
# Trading Order - E2E Test
# ============================================================================
#
# How to Run:
#   yarn test --tags @trading-order       # Run order tests only
#   yarn test --tags @trading             # Run all trading tests
#   yarn test --tags @smoke               # Run smoke tests
# ============================================================================

@trading @trading-order
Feature: Trading Order Management
  As a Booker trader
  I want to place buy and sell orders
  So that I can execute trades on the platform

  @order @smoke @priority_high
  Scenario: View trading page after login
    Given I am logged in to the platform
    When I navigate to the trading page
    Then the trading page should display correctly
    And the order book should be visible
    And the open orders section should be visible

  @order @priority_high
  Scenario: Place a buy order with valid price and quantity
    Given I am logged in to the platform
    And I am on the trading page
    When I select a trading pair
    And I fill in the buy order form with price "50000" and quantity "0.1"
    And I submit the buy order
    Then I should see a success message
    And the order should appear in the open orders

  @order @priority_high
  Scenario: Place a sell order with valid price and quantity
    Given I am logged in to the platform
    And I am on the trading page
    When I select a trading pair
    And I fill in the sell order form with price "51000" and quantity "0.05"
    And I submit the sell order
    Then I should see a success message
    And the order should appear in the open orders

  @order @priority_medium
  Scenario: Order form shows calculated total
    Given I am logged in to the platform
    And I am on the trading page
    When I select a trading pair
    And I fill in the buy order form with price "45000" and quantity "0.2"
    Then the total should be calculated correctly

  @order @priority_medium
  Scenario: Cannot submit order with empty fields
    Given I am logged in to the platform
    And I am on the trading page
    When I select a trading pair
    And I try to submit the buy order without filling any fields
    Then the buy order should not be submitted
