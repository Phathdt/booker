# ============================================================================
# Trading Notifications - E2E Test
# ============================================================================
#
# How to Run:
#   yarn test --tags @notifications      # Run notification tests only
#   yarn test --tags @trading            # Run all trading tests
# ============================================================================

@trading @notifications
Feature: Trading Notifications
  As a Booker trader
  I want to receive notifications about my trades and orders
  So that I stay informed about my trading activity

  @notification @smoke @priority_high
  Scenario: Notification bell is visible in header
    Given I am logged in to the platform
    And I am on the trading page
    Then the notification bell should be visible in the header

  @notification @priority_high
  Scenario: Notification bell shows unread count after trade
    Given I am logged in to the platform
    And I am on the trading page
    When I select a trading pair
    And I fill in the sell order form with price "39000" and quantity "0.01"
    And I submit the sell order
    Then I should see a success message
    When I fill in the buy order form with price "39000" and quantity "0.01"
    And I submit the buy order
    Then I should see a success message
    And the notification bell should show unread count

  @notification @priority_medium
  Scenario: Click notification bell opens dropdown
    Given I am logged in to the platform
    And I am on the trading page
    When I click the notification bell
    Then the notification dropdown should be visible
