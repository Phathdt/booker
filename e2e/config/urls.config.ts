/**
 * Centralized URL configuration for Booker E2E tests
 */

// Base URL from environment - we only interact via browser, not API directly
export const URLS = {
  APP_URL: process.env.APP_URL || 'http://booker.localhost',

  // Application routes (matches web/src/core/router/index.tsx)
  ROUTES: {
    LOGIN: '/login',
    TRADE: '/trade',
    WALLET: '/wallet',
  },
};

export function getAppUrl(path = ''): string {
  return `${URLS.APP_URL}${path}`;
}

/**
 * Get test user credentials from environment
 */
export function getTestCredentials(): { email: string; password: string } {
  const email = process.env.TEST_USER_EMAIL || 'user1@test.com';
  const password = process.env.TEST_USER_PASSWORD || 'password123';
  return { email, password };
}
