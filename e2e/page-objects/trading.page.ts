/**
 * Page Object for Booker Trading page (/trade)
 */
import { Page, expect } from '@playwright/test';
import { TimeoutValue } from '@config/test.config';

export class TradingPage {
  constructor(private readonly page: Page) {}

  // ----- Locators -----

  private get heading() {
    return this.page.getByRole('heading', { name: 'Trade', exact: true });
  }

  private get orderBookHeading() {
    return this.page.getByRole('heading', { name: 'Order Book' });
  }

  private get openOrdersHeading() {
    return this.page.getByRole('heading', { name: 'Open Orders' });
  }

  private get buyTab() {
    return this.page.getByRole('tab', { name: 'Buy' });
  }

  private get sellTab() {
    return this.page.getByRole('tab', { name: 'Sell' });
  }

  private buyPriceInput() {
    return this.page.locator('#buy-price');
  }

  private buyQuantityInput() {
    return this.page.locator('#buy-quantity');
  }

  private sellPriceInput() {
    return this.page.locator('#sell-price');
  }

  private sellQuantityInput() {
    return this.page.locator('#sell-quantity');
  }

  private get pairSelector() {
    return this.page.getByRole('combobox');
  }

  // ----- Actions -----

  async waitForPageLoad(): Promise<void> {
    await expect(this.heading).toBeVisible({ timeout: TimeoutValue.NAVIGATION });
  }

  async selectPair(pairId: string): Promise<void> {
    const currentValue = await this.pairSelector.inputValue().catch(() => '');
    if (currentValue === pairId) return; // Already selected

    await this.pairSelector.click();
    // Labels use "BTC / USDT" format, ids use "BTC_USDT"
    const label = pairId.replace('_', ' / ');
    await this.page.getByRole('option', { name: label }).click();
  }

  async switchToBuyTab(): Promise<void> {
    await this.buyTab.click();
  }

  async switchToSellTab(): Promise<void> {
    await this.sellTab.click();
  }

  async fillBuyOrder(price: string, quantity: string): Promise<void> {
    await this.switchToBuyTab();
    await this.buyPriceInput().fill(price);
    await this.buyQuantityInput().fill(quantity);
  }

  async submitBuyOrder(): Promise<void> {
    const button = this.page.getByRole('button', { name: /buy/i });
    await button.click();
  }

  async fillSellOrder(price: string, quantity: string): Promise<void> {
    await this.switchToSellTab();
    await this.sellPriceInput().fill(price);
    await this.sellQuantityInput().fill(quantity);
  }

  async submitSellOrder(): Promise<void> {
    const button = this.page.getByRole('button', { name: /sell/i });
    await button.click();
  }

  async placeBuyOrder(price: string, quantity: string): Promise<void> {
    await this.fillBuyOrder(price, quantity);
    await this.submitBuyOrder();
  }

  async placeSellOrder(price: string, quantity: string): Promise<void> {
    await this.fillSellOrder(price, quantity);
    await this.submitSellOrder();
  }

  async cancelFirstOpenOrder(): Promise<void> {
    const openOrdersSection = this.page.locator('section, div').filter({ has: this.openOrdersHeading });
    const cancelButton = openOrdersSection.getByRole('button', { name: /cancel/i }).first();
    await cancelButton.click();
  }

  async getOpenOrderCount(): Promise<number> {
    const openOrdersSection = this.page.locator('section, div').filter({ has: this.openOrdersHeading });
    const rows = openOrdersSection.locator('table tbody tr');
    return rows.count();
  }

  async getActiveOrderCount(): Promise<number> {
    // Count only orders with status "new" or "partial" (orders that can be cancelled)
    const openOrdersSection = this.page.locator('section, div').filter({ has: this.openOrdersHeading });
    const rows = openOrdersSection.locator('table tbody tr');

    let count = 0;
    const rowCount = await rows.count();
    for (let i = 0; i < rowCount; i++) {
      const row = rows.nth(i);
      const statusBadge = row.locator('[class*="badge"]');
      const statusText = await statusBadge.textContent();
      if (statusText === 'new' || statusText === 'partial') {
        count++;
      }
    }
    return count;
  }

  async getFilledOrderCount(): Promise<number> {
    // Count only orders with status "filled"
    const openOrdersSection = this.page.locator('section, div').filter({ has: this.openOrdersHeading });
    const rows = openOrdersSection.locator('table tbody tr');

    let count = 0;
    const rowCount = await rows.count();
    for (let i = 0; i < rowCount; i++) {
      const row = rows.nth(i);
      const statusBadge = row.locator('[class*="badge"]');
      const statusText = await statusBadge.textContent();
      if (statusText === 'filled') {
        count++;
      }
    }
    return count;
  }

  // ----- Assertions -----

  async expectOnTradingPage(): Promise<void> {
    await expect(this.heading).toBeVisible({ timeout: TimeoutValue.ACTION });
    await expect(this.orderBookHeading).toBeVisible();
    await expect(this.openOrdersHeading).toBeVisible();
  }

  async expectTotalDisplayed(expectedTotal: string): Promise<void> {
    const totalText = this.page.locator('.font-mono').filter({ hasText: 'USDT' });
    await expect(totalText).toContainText(expectedTotal);
  }

  async expectNoOpenOrders(): Promise<void> {
    const openOrdersSection = this.page.locator('section, div').filter({ has: this.openOrdersHeading });
    const emptyState = openOrdersSection.getByText(/no orders found/i);
    const rows = openOrdersSection.locator('table tbody tr');
    // Either empty state text is shown or the table has no rows
    const hasEmptyState = await emptyState.isVisible();
    if (!hasEmptyState) {
      const rowCount = await rows.count();
      expect(rowCount).toBe(0);
    }
  }

  async expectTickerBarVisible(): Promise<void> {
    // Ticker bar contains "Last Price" and "24h Change" labels
    const lastPrice = this.page.getByText('Last Price');
    const change = this.page.getByText('24h Change');
    // Either the ticker data or "No ticker data" placeholder should be visible
    const noData = this.page.getByText('No ticker data');
    const hasData = await lastPrice.isVisible({ timeout: TimeoutValue.ACTION }).catch(() => false);
    if (!hasData) {
      await expect(noData).toBeVisible({ timeout: TimeoutValue.ACTION });
    } else {
      await expect(change).toBeVisible();
    }
  }

  async expectOrderExecuted(): Promise<void> {
    // After matching, both orders should be filled (status: "filled").
    // They remain in the table but are no longer active (can't be cancelled).
    await this.page.waitForTimeout(TimeoutValue.STRATEGIC_PART_DELAY);
    // Verify we are still on the trading page (no error redirect)
    await this.expectOnTradingPage();
    // Verify no active (new/partial) orders remain — all should be filled
    const activeOrderCount = await this.getActiveOrderCount();
    expect(activeOrderCount).toBe(0);
  }
}
