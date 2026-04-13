const path = require('path');
require('dotenv').config({ path: path.resolve(__dirname, '.env') });
require('ts-node/register');
require('tsconfig-paths/register');

const testConfig = require('./config/test.config.ts').default;

const DEFAULT_FEATURE_PATHS = [
  'tests/features/**/*.feature'
];

function getCliFeaturePaths() {
  return process.argv
    .slice(2)
    .filter((arg) => !arg.startsWith('-'))
    .filter((arg) => arg.includes('.feature'));
}

const cliFeaturePaths = getCliFeaturePaths();

const common = {
  requireModule: [
    'ts-node/register',
    'tsconfig-paths/register'
  ],
  require: [
    'tests/features/**/*.steps.ts',
    'tests/support/**/*.ts'
  ],
  paths: cliFeaturePaths.length > 0 ? cliFeaturePaths : DEFAULT_FEATURE_PATHS,
  format: [
    process.env.TEST_ENVIRONMENT === 'ci' ? 'progress' : 'progress-bar',
    'json:test-results/cucumber-report.json',
    'message:test-results/cucumber-messages.ndjson',
    'summary:test-results/summary.txt'
  ],
  publishQuiet: true
};

module.exports = {
  default: {
    ...common,
    ...testConfig
  }
};
