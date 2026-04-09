import { defineConfig } from "@playwright/test";

// Config for the compose-smoke CI job. Targets a local docker-compose stack
// exposed at http://localhost:8000 via the nginx gateway defined in
// docker-compose.yml. Pair with docker-compose.ci.yml which overrides
// OLLAMA_BASE_URL to point at the mock-ollama stub.
export default defineConfig({
  testDir: "./e2e",
  testMatch: "smoke-ci.spec.ts",
  fullyParallel: false,
  retries: 1,
  workers: 1,
  reporter: "list",
  use: {
    trace: "on-first-retry",
  },
});
