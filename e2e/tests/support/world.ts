/**
 * Custom Cucumber World with Playwright browser resources.
 * All step files use `this.page`, `this.browser`, etc. via Cucumber's World binding.
 */
import { setWorldConstructor, setDefaultTimeout, World, IWorldOptions } from '@cucumber/cucumber';
import { TimeoutValue } from '@config/test.config';

// 120s per step — enough for login + navigation flows
setDefaultTimeout(TimeoutValue.TEST_WORKFLOW);
import { Browser, BrowserContext, Page } from '@playwright/test';

export class BrowserWorld extends World {
  browser!: Browser;
  context!: BrowserContext;
  page!: Page;

  // Scenario-specific data (steps can store anything here)
  data: Record<string, any> = {};

  constructor(options: IWorldOptions) {
    super(options);
  }
}

setWorldConstructor(BrowserWorld);
