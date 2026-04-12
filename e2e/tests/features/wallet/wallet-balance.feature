# ============================================================================
# Wallet Balance Management - E2E Test
# ============================================================================
#
# How to Run:
#   yarn test --tags @wallet-balance      # Run balance tests only
#   yarn test --tags @wallet              # Run all wallet tests
#   yarn test --tags @smoke               # Run smoke tests
# ============================================================================

@wallet @wallet-balance
Feature: Wallet Balance Management
  As a Booker trader
  I want to manage my wallet balances
  So that I can deposit and withdraw funds

  @balance @smoke @priority_high
  Scenario: View wallet page with balance table
    Given I am logged in to the platform
    When I navigate to the wallet page
    Then the wallet page should display correctly
    And the balance table should be visible
    And the balance table should contain asset information

  @balance @priority_high
  Scenario: Navigate from trading to wallet
    Given I am logged in to the platform
    And I am on the trading page
    When I navigate to the wallet page
    Then the wallet page should display correctly

  @balance @priority_medium
  Scenario: Deposit funds into an asset
    Given I am logged in to the platform
    And I am on the wallet page
    When I click the deposit button for "USDT"
    And I fill in the deposit amount with "100"
    And I submit the deposit
    Then I should see a success message
    And the deposit dialog should close

  @balance @priority_medium
  Scenario: Withdraw funds from an asset
    Given I am logged in to the platform
    And I am on the wallet page
    When I click the withdraw button for "USDT"
    And I fill in the withdraw amount with "50"
    And I submit the withdraw
    Then I should see a success message
    And the withdraw dialog should close

  @balance @priority_low
  Scenario: Multiple assets visible in wallet
    Given I am logged in to the platform
    And I am on the wallet page
    Then the balance table should contain asset "USDT"
    And the balance table should contain asset "BTC"
