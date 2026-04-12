/**
 * Global After hook to save Playwright traces on scenario failure (debug mode only)
 */
import { After, Status, ITestCaseHookParameter } from '@cucumber/cucumber';
import { saveTraceOnFailure, getActiveContexts } from '@utils/browser-factory';

const isDebug = process.env.TEST_ENVIRONMENT === 'debug';

if (isDebug) {
  After(async function ({ pickle, result }: ITestCaseHookParameter) {
    if (result?.status !== Status.FAILED) return;

    const activeContexts = getActiveContexts();
    if (activeContexts.size === 0) return;

    const scenarioName = pickle?.name || 'unknown-scenario';
    const firstContext = Array.from(activeContexts)[0];
    await saveTraceOnFailure(firstContext, scenarioName);
  });
}
