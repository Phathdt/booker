/**
 * Page Object for Booker Login/Register page (/login)
 */
import { Page, expect } from '@playwright/test';
import { TimeoutValue } from '@config/test.config';
import { getAppUrl, URLS } from '@config/urls.config';

export class LoginPage {
  constructor(private readonly page: Page) {}

  // ----- Locators -----

  private get loginTab() {
    return this.page.getByRole('tab', { name: 'Login' });
  }

  private get registerTab() {
    return this.page.getByRole('tab', { name: 'Register' });
  }

  private get loginEmailInput() {
    return this.page.locator('#login-email');
  }

  private get loginPasswordInput() {
    return this.page.locator('#login-password');
  }

  private get registerEmailInput() {
    return this.page.locator('#register-email');
  }

  private get registerPasswordInput() {
    return this.page.locator('#register-password');
  }

  private get signInButton() {
    return this.page.getByRole('button', { name: /sign in/i });
  }

  private get createAccountButton() {
    return this.page.getByRole('button', { name: /create account/i });
  }

  private get heading() {
    return this.page.getByRole('heading', { name: 'Booker' });
  }

  // ----- Actions -----

  async navigate(): Promise<void> {
    await this.page.goto(getAppUrl(URLS.ROUTES.LOGIN), {
      waitUntil: 'domcontentloaded',
      timeout: TimeoutValue.NAVIGATION,
    });
  }

  async waitForPageLoad(): Promise<void> {
    await expect(this.heading).toBeVisible({ timeout: TimeoutValue.NAVIGATION });
  }

  async switchToLoginTab(): Promise<void> {
    await this.loginTab.click();
  }

  async switchToRegisterTab(): Promise<void> {
    await this.registerTab.click();
  }

  async fillLoginForm(email: string, password: string): Promise<void> {
    await this.loginEmailInput.fill(email);
    await this.loginPasswordInput.fill(password);
  }

  async submitLogin(): Promise<void> {
    await this.signInButton.click();
  }

  async login(email: string, password: string): Promise<void> {
    await this.switchToLoginTab();
    await this.fillLoginForm(email, password);
    await this.submitLogin();
  }

  async fillRegisterForm(email: string, password: string): Promise<void> {
    await this.registerEmailInput.fill(email);
    await this.registerPasswordInput.fill(password);
  }

  async submitRegister(): Promise<void> {
    await this.createAccountButton.click();
  }

  async register(email: string, password: string): Promise<void> {
    await this.switchToRegisterTab();
    await this.fillRegisterForm(email, password);
    await this.submitRegister();
  }

  // ----- Assertions -----

  async expectSignInButtonDisabled(): Promise<void> {
    await expect(this.signInButton).toBeDisabled();
  }

  async expectSignInButtonEnabled(): Promise<void> {
    await expect(this.signInButton).toBeEnabled();
  }

  async expectOnLoginPage(): Promise<void> {
    await expect(this.heading).toBeVisible({ timeout: TimeoutValue.ACTION });
  }

  async expectToastError(message: string): Promise<void> {
    const toast = this.page.locator('[data-sonner-toast][data-type="error"]');
    await expect(toast).toBeVisible({ timeout: TimeoutValue.ACTION });
    await expect(toast).toContainText(message);
  }
}
