/**
 * Step Wrapper - auto-captures screenshots on step failure
 */
import { Page } from '@playwright/test';
import { logger } from '@utils/logger';

export type AttachFunction = (data: string | Buffer, mediaType: string) => void | Promise<void>;

/**
 * Wraps a step function with error handling and screenshot capture
 */
export function withErrorHandling<T extends (...args: any[]) => Promise<any>>(
  stepFn: T,
  page: Page,
  attachFn: AttachFunction,
): T {
  return (async (...args: any[]) => {
    try {
      return await stepFn(...args);
    } catch (error) {
      try {
        const screenshot = await page.screenshot({ fullPage: true });
        await attachFn(screenshot, 'image/png');
        logger.debug('Screenshot captured on step failure');
      } catch (screenshotError) {
        logger.warn('Failed to capture screenshot');
      }
      throw error;
    }
  }) as T;
}

/**
 * Creates a step wrapper bound to a specific page and attach function
 */
export function createStepWrapper(page: Page, attachFn: AttachFunction) {
  return <T extends (...args: any[]) => Promise<any>>(stepFn: T): T => {
    return withErrorHandling(stepFn, page, attachFn);
  };
}

/**
 * Capture and attach a screenshot to the Cucumber report
 */
export async function captureAndAttachScreenshot(
  page: Page,
  attachFn: AttachFunction,
  description?: string,
): Promise<void> {
  const screenshot = await page.screenshot({ fullPage: true });
  await attachFn(screenshot, 'image/png');
  if (description) logger.info(`Screenshot captured: ${description}`);
}
