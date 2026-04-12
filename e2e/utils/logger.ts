/**
 * Pino-based logger for E2E tests.
 *
 * LOG_FORMAT=json  → structured JSON (CI pipelines)
 * LOG_FORMAT=pretty → colored human-readable (default for local)
 */
import pino from 'pino';
import { LogLevel, testConfigPresets } from '@config/test.config';

const env = process.env.TEST_ENVIRONMENT || 'local';
const preset = testConfigPresets[env] || testConfigPresets.local;

const pinoLevelMap: Record<LogLevel, string> = {
  [LogLevel.ERROR]: 'error',
  [LogLevel.WARN]: 'warn',
  [LogLevel.INFO]: 'info',
  [LogLevel.DEBUG]: 'debug',
};

const logFormat = process.env.LOG_FORMAT || (preset.ci ? 'json' : 'pretty');

const transport = logFormat === 'pretty'
  ? {
      target: 'pino-pretty',
      options: {
        colorize: true,
        translateTime: 'HH:MM:ss.l',
        ignore: 'pid,hostname',
      },
    }
  : undefined; // default JSON output

export const logger = pino({
  level: pinoLevelMap[preset.logLevel] || 'info',
  transport,
});
