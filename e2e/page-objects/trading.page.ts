/**
 * Page Object for Booker Trading page (/trade)
 */
import { Page, expect } from '@playwright/test';
import { TimeoutValue } from '@config/test.config';

export class TradingPage {
  constructor(private readonly page: Page) {}

  // ----- Locators -----

  private get heading() {
    return this.page.getByRole('heading', { name: 'Trade' });
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
}
