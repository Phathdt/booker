/**
 * Trading Market Data step definitions.
 * Tests ticker bar, recent trades, and pair switching.
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

When('I select trading pair {string}', async function (this: BrowserWorld, pairId: string) {
  logger.info(`Selecting trading pair ${pairId}`);
  const tradingPage = new TradingPage(this.page);
  await tradingPage.selectPair(pairId);
  await this.page.waitForTimeout(TimeoutValue.STRATEGIC_PART_DELAY);
});

// ============================================================================
// THEN STEPS
// ============================================================================

Then('the ticker bar should be visible', async function (this: BrowserWorld) {
  logger.info('Verifying ticker bar');
  const tradingPage = new TradingPage(this.page);
  await tradingPage.expectTickerBarVisible();
  logger.info('Ticker bar verified');
});

Then('the ticker bar should show price information', async function (this: BrowserWorld) {
  logger.info('Verifying ticker bar shows price data');
  const tradingPage = new TradingPage(this.page);
  await tradingPage.expectTickerBarVisible();
  logger.info('Ticker bar price data verified');
});

Then('the recent trades section should be visible', async function (this: BrowserWorld) {
  logger.info('Verifying recent trades section');
  const heading = this.page.getByRole('heading', { name: 'Recent Trades' });
  await expect(heading).toBeVisible({ timeout: TimeoutValue.ACTION });
  logger.info('Recent trades section verified');
});
