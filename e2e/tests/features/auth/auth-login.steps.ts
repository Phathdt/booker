/**
 * Auth Login step definitions.
 * Shared steps: "I navigate to the login page", "I should be redirected to the trading page",
 *               "the trading page should display correctly" — defined in shared-steps.ts
 */
import { When, Then } from '@cucumber/cucumber';
import { expect } from '@playwright/test';
import { LoginPage } from '@page-objects/login.page';
import { logger } from '@utils/logger';
import { TimeoutValue } from '@config/test.config';
import { getTestCredentials } from '@config/urls.config';
import type { BrowserWorld } from '../../support/world';

// ============================================================================
// WHEN STEPS
// ============================================================================

When('I enter valid login credentials', async function (this: BrowserWorld) {
  const { email, password } = getTestCredentials();
  logger.info(`Filling login form with email: ${email}`);
  const loginPage = new LoginPage(this.page);
  await loginPage.fillLoginForm(email, password);
});

When('I enter invalid login credentials', async function (this: BrowserWorld) {
  logger.info('Filling login form with invalid credentials');
  const loginPage = new LoginPage(this.page);
  await loginPage.fillLoginForm('invalid@test.com', 'wrongpassword');
});

When('I submit the login form', async function (this: BrowserWorld) {
  logger.info('Submitting login form');
  const loginPage = new LoginPage(this.page);
  await loginPage.submitLogin();
  await this.page.waitForTimeout(TimeoutValue.STRATEGIC_PART_DELAY);
});

When('I leave the login fields empty', async function (this: BrowserWorld) {
  logger.info('Leaving login fields empty');
  // Fields are empty by default after page load
});

// ============================================================================
// THEN STEPS
// ============================================================================

Then('I should see a login error message', async function (this: BrowserWorld) {
  logger.info('Verifying login error toast');
  const toast = this.page.locator('[data-sonner-toast]');
  await expect(toast).toBeVisible({ timeout: TimeoutValue.ACTION });
  logger.info('Login error toast displayed');
});

Then('I should remain on the login page', async function (this: BrowserWorld) {
  logger.info('Verifying still on login page');
  const loginPage = new LoginPage(this.page);
  await loginPage.expectOnLoginPage();
});

Then('the sign in button should not submit the form', async function (this: BrowserWorld) {
  logger.info('Verifying empty form cannot be submitted');
  const loginPage = new LoginPage(this.page);
  await loginPage.expectOnLoginPage();
});
