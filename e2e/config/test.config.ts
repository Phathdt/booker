/**
 * Test Configuration for Booker E2E Tests
 * Centralized configuration with environment presets
 */
import * as path from 'path';
import * as dotenv from 'dotenv';

dotenv.config({ path: path.resolve(__dirname, '..', '.env') });

// =============================================================================
// ENUMS
// =============================================================================

export enum LogLevel {
  ERROR = 'ERROR',
  WARN = 'WARN',
  INFO = 'INFO',
  DEBUG = 'DEBUG',
}

export enum BrowserType {
  CHROMIUM = 'chromium',
  FIREFOX = 'firefox',
  WEBKIT = 'webkit',
}

export enum TestEnvironment {
  LOCAL = 'local',
  CI = 'ci',
  DEBUG = 'debug',
}

export enum TraceMode {
  OFF = 'off',
  ON = 'on',
  RETAIN_ON_FAILURE = 'retain-on-failure',
}

export enum VideoMode {
  OFF = 'off',
  ON = 'on',
  RETAIN_ON_FAILURE = 'retain-on-failure',
}

// =============================================================================
// TIMEOUT CONSTANTS (10-second rule with documented exceptions)
// =============================================================================

export enum TimeoutValue {
  DEFAULT = 10000,
  ACTION = 10000,
  NAVIGATION = 10000,
  EXPECT = 10000,
  QUICK_CHECK = 1000,
  STRATEGIC_PART_DELAY = 2000,
  STRATEGIC_ACTION_DELAY = 300,
  TEST_LOCAL = 30000,
  TEST_CI = 60000,
  TEST_DEBUG = 60000,
  TEST_WORKFLOW = 120000,
}

// =============================================================================
// RETRY CONFIGURATION
// =============================================================================

export const RETRY_TIMES = 3;
export const RETRY_TIMES_ACTION = 3;

// =============================================================================
// VIEWPORT
// =============================================================================

export const VIEWPORT = { width: 1366, height: 768 };

// =============================================================================
// INTERFACES
// =============================================================================

export interface TestConfig {
  headless: boolean;
  workers: number;
  retries: number;
  testTimeout: number;
  actionTimeout: number;
  navigationTimeout: number;
  browser: BrowserType;
  viewportWidth: number;
  viewportHeight: number;
  logLevel: LogLevel;
  trace: TraceMode;
  video: VideoMode;
  testEnvironment: TestEnvironment;
  ci: boolean;
}

// =============================================================================
// PRESETS
// =============================================================================

export const testConfigPresets: Record<string, TestConfig> = {
  local: {
    headless: false,
    workers: 1,
    retries: 0,
    testTimeout: TimeoutValue.TEST_LOCAL,
    actionTimeout: TimeoutValue.ACTION,
    navigationTimeout: TimeoutValue.NAVIGATION,
    browser: BrowserType.CHROMIUM,
    viewportWidth: VIEWPORT.width,
    viewportHeight: VIEWPORT.height,
    logLevel: LogLevel.INFO,
    trace: TraceMode.RETAIN_ON_FAILURE,
    video: VideoMode.RETAIN_ON_FAILURE,
    testEnvironment: TestEnvironment.LOCAL,
    ci: false,
  },
  ci: {
    headless: true,
    workers: 1,
    retries: 2,
    testTimeout: TimeoutValue.TEST_CI,
    actionTimeout: TimeoutValue.ACTION,
    navigationTimeout: TimeoutValue.NAVIGATION,
    browser: BrowserType.CHROMIUM,
    viewportWidth: VIEWPORT.width,
    viewportHeight: VIEWPORT.height,
    logLevel: LogLevel.INFO,
    trace: TraceMode.RETAIN_ON_FAILURE,
    video: VideoMode.RETAIN_ON_FAILURE,
    testEnvironment: TestEnvironment.CI,
    ci: true,
  },
  debug: {
    headless: false,
    workers: 1,
    retries: 0,
    testTimeout: TimeoutValue.TEST_DEBUG,
    actionTimeout: TimeoutValue.ACTION,
    navigationTimeout: TimeoutValue.NAVIGATION,
    browser: BrowserType.CHROMIUM,
    viewportWidth: VIEWPORT.width,
    viewportHeight: VIEWPORT.height,
    logLevel: LogLevel.DEBUG,
    trace: TraceMode.ON,
    video: VideoMode.RETAIN_ON_FAILURE,
    testEnvironment: TestEnvironment.DEBUG,
    ci: false,
  },
};

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

export const parseBoolean = (value: string | undefined, defaultValue = false): boolean => {
  if (value === undefined) return defaultValue;
  return value.toLowerCase() === 'true';
};

export const parseNumber = (value: string | undefined, defaultValue: number): number => {
  if (value === undefined) return defaultValue;
  const parsed = Number(value);
  return isNaN(parsed) ? defaultValue : parsed;
};

export const getRequiredString = (value: string | undefined, envVarName: string): string => {
  if (!value || value.trim() === '') {
    throw new Error(`Required environment variable ${envVarName} is not set.`);
  }
  return value;
};

// =============================================================================
// RESOLVED CONFIG (default export for .cucumber.js)
// =============================================================================

interface CucumberConfig {
  parallel: number;
  workers: number;
}

const getResolvedConfig = (): CucumberConfig => {
  const environment = process.env.TEST_ENVIRONMENT || 'local';
  const preset = testConfigPresets[environment];
  if (!preset) {
    throw new Error(`Invalid TEST_ENVIRONMENT: "${environment}". Valid: ${Object.keys(testConfigPresets).join(', ')}`);
  }

  const config: CucumberConfig = {
    parallel: preset.workers,
    workers: preset.workers,
  };

  if (process.env.CUCUMBER_WORKERS !== undefined) {
    config.parallel = parseNumber(process.env.CUCUMBER_WORKERS, config.parallel);
    config.workers = config.parallel;
  }

  return config;
};

export default getResolvedConfig();
