/**
 * Page Object for Booker Wallet page (/wallet)
 */
import { Page, expect } from '@playwright/test';
import { TimeoutValue } from '@config/test.config';

export class WalletPage {
  constructor(private readonly page: Page) {}

  // ----- Locators -----

  private get heading() {
    return this.page.getByRole('heading', { name: 'Wallet' });
  }

  private get balancesHeading() {
    return this.page.getByRole('heading', { name: 'Balances' });
  }

  private get balanceTable() {
    return this.page.locator('table');
  }

  // ----- Actions -----

  async waitForPageLoad(): Promise<void> {
    await expect(this.heading).toBeVisible({ timeout: TimeoutValue.NAVIGATION });
  }

  async clickDepositButton(asset: string): Promise<void> {
    const row = this.page.locator('tr').filter({ hasText: asset });
    await row.getByRole('button', { name: /deposit/i }).click();
  }

  async clickWithdrawButton(asset: string): Promise<void> {
    const row = this.page.locator('tr').filter({ hasText: asset });
    await row.getByRole('button', { name: /withdraw/i }).click();
  }

  async fillDepositAmount(amount: string): Promise<void> {
    const dialog = this.page.locator('[role="dialog"]');
    await dialog.locator('input[type="number"], input[placeholder*="0"]').fill(amount);
  }

  async submitDeposit(): Promise<void> {
    const dialog = this.page.locator('[role="dialog"]');
    await dialog.getByRole('button', { name: /deposit/i }).click();
  }

  async fillWithdrawAmount(amount: string): Promise<void> {
    const dialog = this.page.locator('[role="dialog"]');
    await dialog.locator('input[type="number"], input[placeholder*="0"]').fill(amount);
  }

  async submitWithdraw(): Promise<void> {
    const dialog = this.page.locator('[role="dialog"]');
    await dialog.getByRole('button', { name: /withdraw/i }).click();
  }

  // ----- Assertions -----

  async expectOnWalletPage(): Promise<void> {
    await expect(this.heading).toBeVisible({ timeout: TimeoutValue.ACTION });
    await expect(this.balancesHeading).toBeVisible();
  }

  async expectBalanceTableVisible(): Promise<void> {
    await expect(this.balanceTable).toBeVisible({ timeout: TimeoutValue.ACTION });
  }

  async expectAssetInTable(asset: string): Promise<void> {
    const row = this.page.locator('tr').filter({ hasText: asset });
    await expect(row).toBeVisible({ timeout: TimeoutValue.ACTION });
  }
}
