import { defineConfig } from "@playwright/test";

export default defineConfig({
  // Default config runs only the mocked-backend tests. Runtime-specific
  // suites live in sibling directories and are invoked by dedicated configs:
  //   e2e/smoke-prod/    → playwright.smoke.config.ts    (smoke-production CI)
  //   e2e/smoke-compose/ → playwright.smoke-ci.config.ts (compose-smoke CI)
  testDir: "./e2e/mocked",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: "html",
  use: {
    baseURL: "http://localhost:3000",
    trace: "on-first-retry",
  },
  webServer: {
    command: "npm run dev",
    url: "http://localhost:3000",
    reuseExistingServer: !process.env.CI,
  },
});
