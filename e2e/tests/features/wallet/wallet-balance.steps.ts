/**
 * Wallet Balance step definitions.
 * Shared steps: "I am logged in to the platform", "I am on the trading page",
 *               "I am on the wallet page", "I should see a success message"
 *               — defined in shared-steps.ts
 */
import { When, Then } from '@cucumber/cucumber';
import { expect } from '@playwright/test';
import { WalletPage } from '@page-objects/wallet.page';
import { logger } from '@utils/logger';
import { TimeoutValue } from '@config/test.config';
import { getAppUrl, URLS } from '@config/urls.config';
import type { BrowserWorld } from '../../support/world';

// ============================================================================
// WHEN STEPS
// ============================================================================

When('I navigate to the wallet page', async function (this: BrowserWorld) {
  logger.info('Navigating to wallet page');
  await this.page.goto(getAppUrl(URLS.ROUTES.WALLET), {
    waitUntil: 'domcontentloaded',
    timeout: TimeoutValue.NAVIGATION,
  });
  const walletPage = new WalletPage(this.page);
  await walletPage.waitForPageLoad();
  logger.info('Wallet page loaded');
});

When('I click the deposit button for {string}', async function (this: BrowserWorld, asset: string) {
  logger.info(`Clicking deposit for ${asset}`);
  const walletPage = new WalletPage(this.page);
  await walletPage.clickDepositButton(asset);
  await this.page.waitForTimeout(TimeoutValue.STRATEGIC_ACTION_DELAY);
});

When('I fill in the deposit amount with {string}', async function (this: BrowserWorld, amount: string) {
  logger.info(`Filling deposit amount: ${amount}`);
  this.data.lastDepositAmount = amount;
  const walletPage = new WalletPage(this.page);
  await walletPage.fillDepositAmount(amount);
});

When('I submit the deposit', async function (this: BrowserWorld) {
  logger.info('Submitting deposit');
  const walletPage = new WalletPage(this.page);
  await walletPage.submitDeposit();
  await this.page.waitForTimeout(TimeoutValue.STRATEGIC_PART_DELAY);
});

When('I click the withdraw button for {string}', async function (this: BrowserWorld, asset: string) {
  logger.info(`Clicking withdraw for ${asset}`);
  const walletPage = new WalletPage(this.page);
  await walletPage.clickWithdrawButton(asset);
  await this.page.waitForTimeout(TimeoutValue.STRATEGIC_ACTION_DELAY);
});

When('I fill in the withdraw amount with {string}', async function (this: BrowserWorld, amount: string) {
  logger.info(`Filling withdraw amount: ${amount}`);
  this.data.lastWithdrawAmount = amount;
  const walletPage = new WalletPage(this.page);
  await walletPage.fillWithdrawAmount(amount);
});

When('I submit the withdraw', async function (this: BrowserWorld) {
  logger.info('Submitting withdraw');
  const walletPage = new WalletPage(this.page);
  await walletPage.submitWithdraw();
  await this.page.waitForTimeout(TimeoutValue.STRATEGIC_PART_DELAY);
});

// ============================================================================
// THEN STEPS
// ============================================================================

Then('the wallet page should display correctly', async function (this: BrowserWorld) {
  logger.info('Verifying wallet page');
  const walletPage = new WalletPage(this.page);
  await walletPage.expectOnWalletPage();
});

Then('the balance table should be visible', async function (this: BrowserWorld) {
  logger.info('Verifying balance table');
  const walletPage = new WalletPage(this.page);
  await walletPage.expectBalanceTableVisible();
});

Then('the balance table should contain asset information', async function (this: BrowserWorld) {
  logger.info('Verifying balance table has data');
  const rows = this.page.locator('tr');
  const rowCount = await rows.count();
  logger.info(`Balance table: ${rowCount} rows`);
  expect(rowCount).toBeGreaterThan(0);
});

Then('the deposit dialog should close', async function (this: BrowserWorld) {
  logger.info('Verifying deposit dialog closed');
  const dialog = this.page.locator('[role="dialog"]');
  await expect(dialog).not.toBeVisible({ timeout: TimeoutValue.ACTION });
});

Then('the withdraw dialog should close', async function (this: BrowserWorld) {
  logger.info('Verifying withdraw dialog closed');
  const dialog = this.page.locator('[role="dialog"]');
  await expect(dialog).not.toBeVisible({ timeout: TimeoutValue.ACTION });
});

Then('the balance table should contain asset {string}', async function (this: BrowserWorld, asset: string) {
  logger.info(`Verifying ${asset} in balance table`);
  const walletPage = new WalletPage(this.page);
  await walletPage.expectAssetInTable(asset);
});
