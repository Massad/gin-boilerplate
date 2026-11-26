/**
 * Playwright configuration of the end-to-end suite.
 *
 * The suite runs against the Go binary serving the built SPA (app/dist): a
 * single origin, so the specs never carry an API base URL. The binary is
 * started by Playwright itself (webServer) on the e2e port, with the
 * disposable e2e database. The process environment wins over the .env file
 * (godotenv never overrides an already-set variable), so PORT and DB_NAME
 * point at the e2e environment.
 *
 * The projects are sequential: the setup project seeds the tenant owner once,
 * then the chromium project runs the scenarios. One worker keeps the order
 * deterministic.
 */

import { defineConfig, devices } from "@playwright/test"

/** The port of the Go binary serving the SPA for the suite. */
const E2E_PORT = process.env.E2E_PORT ?? "9100"

/** The disposable database recreated by the e2e cycle (cmd/e2edb). */
const E2E_DB_NAME = process.env.E2E_DB_NAME ?? "artisan_e2e"

/** The single origin of the suite: the SPA and the API are served together. */
const BASE_URL = `http://localhost:${E2E_PORT}`

export default defineConfig({
  testDir: "./e2e",
  fullyParallel: false,
  forbidOnly: true,
  retries: 0,
  workers: 1,
  reporter: "list",
  use: {
    baseURL: BASE_URL,
    trace: "retain-on-failure",
  },
  projects: [
    {
      name: "setup",
      testMatch: "**/setup.ts",
    },
    {
      name: "chromium",
      dependencies: ["setup"],
      testIgnore: "**/setup.ts",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
  webServer: {
    command: "go run -mod=readonly ./cmd/api",
    // The binary serves app/dist and loads .env: it must run from the
    // repository root (the directory of this configuration file is app/).
    cwd: "..",
    url: `${BASE_URL}/`,
    reuseExistingServer: false,
    env: { PORT: E2E_PORT, DB_NAME: E2E_DB_NAME, REDIS_DB: "3" }, // index 3 isolates the e2e sessions (suites run on 2, dev on 1; never inherited from the runner or .env)
    timeout: 120_000,
  },
})
