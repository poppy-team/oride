const { defineConfig, devices } = require('@playwright/test');

/**
 * Example Playwright Configuration optimized for Visual Regression Testing.
 */
module.exports = defineConfig({
  testDir: './tests/visual',
  timeout: 30000,
  expect: {
    // Threshold settings for image comparisons
    toHaveScreenshot: {
      maxDiffPixels: 100, // Allow up to 100 pixels to be different
      threshold: 0.1,     // Acceptable perceived color difference
      animations: 'disabled', // Automatically disable CSS animations
    },
  },
  use: {
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },
    {
      name: 'mobile-chrome',
      use: { ...devices['Pixel 5'] },
    },
  ],
  outputDir: 'test-results/',
});
