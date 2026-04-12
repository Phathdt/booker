/**
 * Shared step definitions used across multiple features.
 * Avoids "ambiguous step" errors by defining common steps once.
 */
import { Given, Then } from '@cucumber/cucumber';
import { expect } from '@playwright/test';
import { LoginPage } from '@page-objects/login.page';
import { TradingPage } from '@page-objects/trading.page';
import { WalletPage } from '@page-objects/wallet.page';
import { logger } from '@utils/logger';
import { TimeoutValue } from '@config/test.config';
import { getAppUrl, getTestCredentials, URLS } from '@config/urls.config';
import type { BrowserWorld } from './world';

// ============================================================================
// SHARED GIVEN STEPS
// ============================================================================

Given('I am logged in to the platform', async function (this: BrowserWorld) {
  logger.info('Logging in to platform');
  const loginPage = new LoginPage(this.page);
  await loginPage.navigate();
  await loginPage.waitForPageLoad();

  const { email, password } = getTestCredentials();
  await loginPage.login(email, password);
  await this.page.waitForURL('**/trade', { timeout: TimeoutValue.NAVIGATION });
  logger.info('Logged in successfully');
});

Given('I am on the trading page', async function (this: BrowserWorld) {
  logger.info('Verifying on trading page');
  const tradingPage = new TradingPage(this.page);
  await tradingPage.waitForPageLoad();
  logger.info('Trading page ready');
});

Given('I am on the wallet page', async function (this: BrowserWorld) {
  logger.info('Navigating to wallet page');
  await this.page.goto(getAppUrl(URLS.ROUTES.WALLET), {
    waitUntil: 'domcontentloaded',
    timeout: TimeoutValue.NAVIGATION,
  });
  const walletPage = new WalletPage(this.page);
  await walletPage.waitForPageLoad();
  logger.info('Wallet page ready');
});

Given('I navigate to the login page', async function (this: BrowserWorld) {
  logger.info('Navigating to login page');
  const loginPage = new LoginPage(this.page);
  await loginPage.navigate();
  await loginPage.waitForPageLoad();
  logger.info('Login page loaded');
});

// ============================================================================
// SHARED THEN STEPS
// ============================================================================

Then('the trading page should display correctly', async function (this: BrowserWorld) {
  logger.info('Verifying trading page');
  const tradingPage = new TradingPage(this.page);
  await tradingPage.expectOnTradingPage();
  logger.info('Trading page verified');
});

Then('I should be redirected to the trading page', async function (this: BrowserWorld) {
  logger.info('Verifying redirect to trading page');
  await this.page.waitForURL('**/trade', { timeout: TimeoutValue.NAVIGATION });
  logger.info('Redirected to trading page');
});

Then('I should see a success message', async function (this: BrowserWorld) {
  logger.info('Verifying success toast');
  const toast = this.page.locator('[data-sonner-toast]');
  await expect(toast).toBeVisible({ timeout: TimeoutValue.ACTION });
  logger.info('Success toast displayed');
});
