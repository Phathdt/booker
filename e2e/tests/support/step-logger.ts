/**
 * Cucumber hooks for step and scenario logging
 */
import {
  Before,
  BeforeStep,
  After,
  AfterStep,
  Status,
  ITestCaseHookParameter,
  ITestStepHookParameter,
} from '@cucumber/cucumber';
import { Page } from '@playwright/test';
import { logger } from '@utils/logger';
import { captureAndAttachScreenshot } from '@utils/step-wrapper';

let currentStepText = '';
let stepStartTime = 0;

interface StepLoggerWorld {
  page?: Page;
  attach?: (data: string | Buffer, mediaType: string) => void | Promise<void>;
}

BeforeStep(function ({ pickleStep }) {
  currentStepText = pickleStep?.text || 'Unknown';
  stepStartTime = Date.now();
  const type = pickleStep?.type || '';
  logger.info(`\n  [${type}] ${currentStepText}`);
});

AfterStep(async function (this: StepLoggerWorld, { result }: ITestStepHookParameter) {
  const duration = Date.now() - stepStartTime;
  if (result.status === Status.PASSED) {
    logger.info(`   PASSED (${duration}ms)`);
  } else if (result.status === Status.FAILED) {
    logger.info(`   FAILED (${duration}ms)`);
    try {
      if (this.page && !this.page.isClosed() && typeof this.attach === 'function') {
        await captureAndAttachScreenshot(this.page, this.attach, `Failed step: ${currentStepText}`);
      }
    } catch (error) {
      logger.warn('Failed to capture screenshot at failed step');
    }
  }
});

Before(function ({ pickle, gherkinDocument }: ITestCaseHookParameter) {
  const scenarioName = pickle?.name || 'Unknown Scenario';
  const feature = gherkinDocument?.feature?.name || 'Unknown Feature';
  logger.info(`\n${'='.repeat(80)}`);
  logger.info(`SCENARIO: ${scenarioName}`);
  logger.info(`  Feature: ${feature}`);
  logger.info(`${'='.repeat(80)}`);
});

After(function ({ pickle, result }: ITestCaseHookParameter) {
  const scenarioName = pickle?.name || 'Unknown Scenario';
  logger.info(`\n${'='.repeat(80)}`);
  logger.info(`${result?.status === Status.PASSED ? 'PASSED' : 'FAILED'}: ${scenarioName}`);
  logger.info(`${'='.repeat(80)}\n`);
});
