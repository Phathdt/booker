/**
 * Auth Register step definitions.
 * Shared steps: "I navigate to the login page", "I should be redirected to the trading page"
 *               — defined in shared-steps.ts
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

When('I switch to the register tab', async function (this: BrowserWorld) {
  logger.info('Switching to register tab');
  const loginPage = new LoginPage(this.page);
  await loginPage.switchToRegisterTab();
});

When('I enter valid registration details', async function (this: BrowserWorld) {
  const email = `e2e-${Date.now()}@test.com`;
  const password = 'TestPassword123!';
  logger.info(`Filling register form: ${email}`);
  this.data.registerEmail = email;
  const loginPage = new LoginPage(this.page);
  await loginPage.fillRegisterForm(email, password);
  await this.attach(`Register email: ${email}`, 'text/plain');
});

When('I enter an email and a short password', async function (this: BrowserWorld) {
  const email = `e2e-short-${Date.now()}@test.com`;
  logger.info('Filling register form with short password');
  const loginPage = new LoginPage(this.page);
  await loginPage.fillRegisterForm(email, '123');
});

When('I enter an existing user email with valid password', async function (this: BrowserWorld) {
  const { email } = getTestCredentials();
  logger.info(`Filling register form with existing email: ${email}`);
  const loginPage = new LoginPage(this.page);
  await loginPage.fillRegisterForm(email, 'TestPassword123!');
});

When('I submit the registration form', async function (this: BrowserWorld) {
  logger.info('Submitting registration form');
  const loginPage = new LoginPage(this.page);
  await loginPage.submitRegister();
  await this.page.waitForTimeout(TimeoutValue.STRATEGIC_PART_DELAY);
});

// ============================================================================
// THEN STEPS
// ============================================================================

Then('I should see a registration error message', async function (this: BrowserWorld) {
  logger.info('Verifying registration error');
  const toast = this.page.locator('[data-sonner-toast]');
  await expect(toast).toBeVisible({ timeout: TimeoutValue.ACTION });
});
