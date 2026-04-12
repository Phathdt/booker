/**
 * Global Before/After hooks for browser lifecycle.
 * Creates browser + page per scenario and attaches to Cucumber World.
 */
import { Before, After } from '@cucumber/cucumber';
import { createBrowserContextPage, closeBrowserResources } from '@utils/browser-factory';
import { logger } from '@utils/logger';
import type { BrowserWorld } from './world';

Before(async function (this: BrowserWorld) {
  const { browser, context, page } = await createBrowserContextPage();
  this.browser = browser;
  this.context = context;
  this.page = page;
  this.data = {};
  logger.info('Browser initialized');
});

After(async function (this: BrowserWorld) {
  try {
    await closeBrowserResources({ browser: this.browser, context: this.context, page: this.page });
  } catch (error) {
    logger.error(error as Error, 'Error during browser cleanup');
  }
  logger.info('Browser closed');
});
