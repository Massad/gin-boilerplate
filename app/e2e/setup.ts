/**
 * Setup project of the e2e suite: it seeds the tenant owner once, before the
 * scenarios run. The database is recreated by the e2e cycle (cmd/e2edb), so
 * the fixture is always fresh.
 */

import { expect, test } from "@playwright/test"

test("seed the tenant owner", async ({ request }) => {
  const response = await request.post("/v1/user/register", {
    data: {
      name: "Camille Laurent",
      email: "owner@e2e.local",
      password: "e2e-owner-password-0",
      company: "Studio Nord E2E",
    },
  })

  const body = await response.text()
  expect(response.status(), `the tenant owner could not be registered: ${body}`).toBe(200)
})
