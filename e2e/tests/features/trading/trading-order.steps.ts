/**
 * Trading Order step definitions.
 * Shared steps: "I am logged in to the platform", "I am on the trading page",
 *               "the trading page should display correctly", "I should see a success message"
 *               — defined in shared-steps.ts
 */
import { When, Then } from '@cucumber/cucumber';
import { expect } from '@playwright/test';
import { TradingPage } from '@page-objects/trading.page';
import { logger } from '@utils/logger';
import { TimeoutValue } from '@config/test.config';
import type { BrowserWorld } from '../../support/world';

// ============================================================================
// WHEN STEPS
// ============================================================================

When('I navigate to the trading page', async function (this: BrowserWorld) {
  logger.info('Navigating to trading page');
  await this.page.waitForURL('**/trade', { timeout: TimeoutValue.NAVIGATION });
  const tradingPage = new TradingPage(this.page);
  await tradingPage.waitForPageLoad();
  logger.info('Trading page loaded');
});

When('I select a trading pair', async function (this: BrowserWorld) {
  logger.info('Selecting trading pair BTC_USDT');
  const tradingPage = new TradingPage(this.page);
  await tradingPage.selectPair('BTC_USDT');
  await this.page.waitForTimeout(TimeoutValue.STRATEGIC_PART_DELAY);
});

When('I fill in the buy order form with price {string} and quantity {string}', async function (this: BrowserWorld, price: string, quantity: string) {
  logger.info(`Filling buy order: price=${price}, quantity=${quantity}`);
  this.data.lastBuyPrice = price;
  this.data.lastBuyQuantity = quantity;
  const tradingPage = new TradingPage(this.page);
  await tradingPage.fillBuyOrder(price, quantity);
});

When('I fill in the sell order form with price {string} and quantity {string}', async function (this: BrowserWorld, price: string, quantity: string) {
  logger.info(`Filling sell order: price=${price}, quantity=${quantity}`);
  this.data.lastSellPrice = price;
  this.data.lastSellQuantity = quantity;
  const tradingPage = new TradingPage(this.page);
  await tradingPage.fillSellOrder(price, quantity);
});

When('I submit the buy order', async function (this: BrowserWorld) {
  logger.info('Submitting buy order');
  const tradingPage = new TradingPage(this.page);
  await tradingPage.submitBuyOrder();
  await this.page.waitForTimeout(TimeoutValue.STRATEGIC_PART_DELAY);
});

When('I submit the sell order', async function (this: BrowserWorld) {
  logger.info('Submitting sell order');
  const tradingPage = new TradingPage(this.page);
  await tradingPage.submitSellOrder();
  await this.page.waitForTimeout(TimeoutValue.STRATEGIC_PART_DELAY);
});

When('I try to submit the buy order without filling any fields', async function (this: BrowserWorld) {
  logger.info('Attempting to submit buy order without filling fields');
  const tradingPage = new TradingPage(this.page);
  await tradingPage.submitBuyOrder();
  await this.page.waitForTimeout(TimeoutValue.STRATEGIC_ACTION_DELAY);
});

When('I cancel the first open order', async function (this: BrowserWorld) {
  logger.info('Cancelling the first open order');
  const tradingPage = new TradingPage(this.page);
  this.data.activeOrderCountBeforeCancel = await tradingPage.getActiveOrderCount();
  await tradingPage.cancelFirstOpenOrder();
  await this.page.waitForTimeout(TimeoutValue.STRATEGIC_PART_DELAY);
});

// ============================================================================
// THEN STEPS
// ============================================================================

Then('the order book should be visible', async function (this: BrowserWorld) {
  logger.info('Verifying order book');
  const heading = this.page.getByRole('heading', { name: 'Order Book' });
  await expect(heading).toBeVisible({ timeout: TimeoutValue.ACTION });
});

Then('the open orders section should be visible', async function (this: BrowserWorld) {
  logger.info('Verifying open orders section');
  const heading = this.page.getByRole('heading', { name: 'Open Orders' });
  await expect(heading).toBeVisible({ timeout: TimeoutValue.ACTION });
});

Then('the order should appear in the open orders', async function (this: BrowserWorld) {
  logger.info('Verifying order in open orders table');
  const table = this.page.locator('table');
  await expect(table).toBeVisible({ timeout: TimeoutValue.ACTION });
});

Then('the total should be calculated correctly', async function (this: BrowserWorld) {
  const price = parseFloat(this.data.lastBuyPrice);
  const quantity = parseFloat(this.data.lastBuyQuantity);
  const expectedTotal = (price * quantity).toFixed(8);
  logger.info(`Verifying total: ${expectedTotal}`);
  const tradingPage = new TradingPage(this.page);
  await tradingPage.expectTotalDisplayed(expectedTotal);
});

Then('the buy order should not be submitted', async function (this: BrowserWorld) {
  logger.info('Verifying buy order not submitted');
  const tradingPage = new TradingPage(this.page);
  await tradingPage.expectOnTradingPage();
});

Then('the cancelled order should be removed from the open orders', async function (this: BrowserWorld) {
  logger.info('Verifying cancelled order removed from open orders');
  const tradingPage = new TradingPage(this.page);
  const activeCountBefore = this.data.activeOrderCountBeforeCancel as number;
  // After cancelling, the cancelled order becomes inactive (no cancel button),
  // so active order count should decrease by 1
  const activeCountAfter = await tradingPage.getActiveOrderCount();
  expect(activeCountAfter).toBe(activeCountBefore - 1);
});

Then('I should see an error message', async function (this: BrowserWorld) {
  logger.info('Verifying error toast');
  const toast = this.page.locator('[data-sonner-toast][data-type="error"]');
  await expect(toast).toBeVisible({ timeout: TimeoutValue.ACTION });
  logger.info('Error toast displayed');
});

Then('the matching orders should be executed', async function (this: BrowserWorld) {
  logger.info('Verifying matching orders executed');
  const tradingPage = new TradingPage(this.page);
  await tradingPage.expectOrderExecuted();
  logger.info('Order execution verified');
});
