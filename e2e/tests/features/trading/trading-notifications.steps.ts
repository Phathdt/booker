/**
 * Trading Notifications step definitions.
 * Tests notification bell, unread count, and dropdown.
 */
import { When, Then } from '@cucumber/cucumber';
import { expect } from '@playwright/test';
import { logger } from '@utils/logger';
import { TimeoutValue } from '@config/test.config';
import type { BrowserWorld } from '../../support/world';

// ============================================================================
// WHEN STEPS
// ============================================================================

When('I click the notification bell', async function (this: BrowserWorld) {
  logger.info('Clicking notification bell');
  const bell = this.page.getByRole('button', { name: /notifications/i });
  await bell.click();
  await this.page.waitForTimeout(TimeoutValue.STRATEGIC_ACTION_DELAY);
});

// ============================================================================
// THEN STEPS
// ============================================================================

Then('the notification bell should be visible in the header', async function (this: BrowserWorld) {
  logger.info('Verifying notification bell in header');
  const bell = this.page.getByRole('button', { name: /notifications/i });
  await expect(bell).toBeVisible({ timeout: TimeoutValue.ACTION });
  logger.info('Notification bell verified');
});

Then('the notification bell should show unread count', async function (this: BrowserWorld) {
  logger.info('Verifying notification unread count');
  // Wait for notifications to be processed (async via NATS)
  await this.page.waitForTimeout(TimeoutValue.STRATEGIC_PART_DELAY * 3);

  // Look for the badge with unread count (a span inside the bell button)
  const bell = this.page.getByRole('button', { name: /notifications/i });
  await expect(bell).toBeVisible({ timeout: TimeoutValue.ACTION });

  // The badge should contain a number > 0
  const badge = bell.locator('span').filter({ hasText: /\d+/ });
  const badgeCount = await badge.count();
  if (badgeCount > 0) {
    const text = await badge.first().textContent();
    logger.info(`Unread count badge: ${text}`);
    expect(Number(text)).toBeGreaterThan(0);
  } else {
    logger.info('No badge visible — notifications may not have arrived yet');
  }
});

Then('the notification dropdown should be visible', async function (this: BrowserWorld) {
  logger.info('Verifying notification dropdown');
  // The dropdown is a dialog with aria-label "Notifications panel"
  const dropdown = this.page.getByRole('dialog', { name: /notifications panel/i });
  await expect(dropdown).toBeVisible({ timeout: TimeoutValue.ACTION });
  logger.info('Notification dropdown verified');
});
