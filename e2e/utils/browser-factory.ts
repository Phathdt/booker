/**
 * Centralized Playwright browser/context factory for Cucumber steps.
 * Step definitions use these helpers instead of calling Playwright launch/newContext directly.
 */
import path from 'path';
import {
  chromium,
  firefox,
  webkit,
  type Browser,
  type BrowserContext,
  type BrowserContextOptions,
  type LaunchOptions,
  type Page,
} from '@playwright/test';
import {
  BrowserType,
  VIEWPORT,
  VideoMode,
  TraceMode,
  testConfigPresets,
  type TestConfig,
  type TestEnvironment,
} from '@config/test.config';

const activeContexts = new Set<BrowserContext>();

export type BrowserBundle = {
  browser: Browser;
  context: BrowserContext;
  page: Page;
};

type CreateBrowserOptions = {
  overrides?: { browser?: LaunchOptions; context?: BrowserContextOptions };
  testName?: string;
};

const resolvePreset = (): TestConfig => {
  const environment = (process.env.TEST_ENVIRONMENT || 'local') as TestEnvironment;
  const preset = testConfigPresets[environment];
  if (!preset) {
    throw new Error(`Invalid TEST_ENVIRONMENT: "${environment}".`);
  }
  return preset;
};

const pickBrowser = (preset: TestConfig) => {
  switch (preset.browser) {
    case BrowserType.FIREFOX: return firefox;
    case BrowserType.WEBKIT: return webkit;
    default: return chromium;
  }
};

const sanitizeTestName = (name: string): string =>
  name.replace(/[<>:"/\\|?*\s]+/g, '_').replace(/_+/g, '_').replace(/^_|_$/g, '');

export const registerContext = (ctx: BrowserContext): void => { activeContexts.add(ctx); };
export const unregisterContext = (ctx: BrowserContext): void => { activeContexts.delete(ctx); };
export const getActiveContexts = (): ReadonlySet<BrowserContext> => activeContexts;

export const createBrowserContextPage = async (
  options?: CreateBrowserOptions,
): Promise<BrowserBundle> => {
  const preset = resolvePreset();
  const browserType = pickBrowser(preset);
  const overrides = options?.overrides;
  const testName = options?.testName;

  // HEADLESS env var overrides preset (e.g. HEADLESS=true yarn test)
  const headless = process.env.HEADLESS !== undefined
    ? process.env.HEADLESS !== 'false'
    : preset.headless;

  const browser = await browserType.launch({
    headless,
    ...overrides?.browser,
  });

  const shouldRecordVideo = preset.video === VideoMode.ON || preset.video === VideoMode.RETAIN_ON_FAILURE;
  const contextOptions: BrowserContextOptions = {
    viewport: { width: preset.viewportWidth, height: preset.viewportHeight },
    ignoreHTTPSErrors: true,
    ...overrides?.context,
  };

  if (shouldRecordVideo) {
    const videosDir = path.resolve(__dirname, '../test-results/videos');
    contextOptions.recordVideo = {
      dir: videosDir,
      size: { width: 1280, height: 720 },
    };
  }

  const context = await browser.newContext(contextOptions);

  // Start tracing if configured
  const shouldTrace = preset.trace === TraceMode.ON || preset.trace === TraceMode.RETAIN_ON_FAILURE;
  if (shouldTrace) {
    await context.tracing.start({ screenshots: true, snapshots: true });
  }

  registerContext(context);
  const page = await context.newPage();

  return { browser, context, page };
};

export const saveTraceOnFailure = async (
  context: BrowserContext,
  scenarioName: string,
): Promise<string | undefined> => {
  const preset = resolvePreset();
  const shouldTrace = preset.trace === TraceMode.ON || preset.trace === TraceMode.RETAIN_ON_FAILURE;
  if (!shouldTrace || !context) return undefined;

  try {
    const fs = require('fs');
    const tracesDir = path.resolve(__dirname, '../test-results/traces');
    if (!fs.existsSync(tracesDir)) {
      fs.mkdirSync(tracesDir, { recursive: true });
    }

    const tracePath = path.join(tracesDir, `${sanitizeTestName(scenarioName)}-${Date.now()}.zip`);
    await context.tracing.stop({ path: tracePath });
    console.log(`Trace saved to: ${tracePath}`);
    return tracePath;
  } catch (error) {
    console.warn('Failed to save trace:', error);
    return undefined;
  }
};

export const closeBrowserResources = async (
  resources: Partial<BrowserBundle & { testPassed?: boolean }>,
): Promise<void> => {
  if (resources.context) {
    const preset = resolvePreset();
    const shouldSaveTrace = preset.trace === TraceMode.ON ||
      (preset.trace === TraceMode.RETAIN_ON_FAILURE && !resources.testPassed);

    if (shouldSaveTrace) {
      try {
        const tracesDir = path.resolve(__dirname, '../test-results/traces');
        const fs = require('fs');
        if (!fs.existsSync(tracesDir)) fs.mkdirSync(tracesDir, { recursive: true });
        await resources.context.tracing.stop({ path: path.join(tracesDir, `${Date.now()}.zip`) });
      } catch { /* tracing may not be active */ }
    } else {
      try { await resources.context.tracing.stop(); } catch { /* ignore */ }
    }
  }

  if (resources.page) await resources.page.close();
  if (resources.context) {
    await resources.context.close();
    unregisterContext(resources.context);
  }
  if (resources.browser) await resources.browser.close();
};
