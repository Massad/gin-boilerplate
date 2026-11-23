/**
 * End-to-end scenarios of the unified login (POST /v1/login): the three
 * customer journeys of the plan, on the single origin served by the Go
 * binary.
 */

import { expect, test } from "@playwright/test"
import type { Page } from "@playwright/test"

const ADMIN_EMAIL = "e2e-admin@atrium.dev"
const ADMIN_PASSWORD = "e2e-admin-password-0"
const OWNER_EMAIL = "e2e-owner@atrium.dev"
const OWNER_PASSWORD = "e2e-owner-password-0"

/** The message of a rejected login, as the maquette spells it. */
const LOGIN_ERROR_TEXT = "Adresse e-mail ou mot de passe incorrect. Vérifiez vos identifiants et réessayez."

/** The single key holding the session in the browser storage. */
const SESSION_KEY = "atrium.session"

/** Reads the raw session payload of the given browser storage. */
function readStoredSession(page: Page, storage: "sessionStorage" | "localStorage"): Promise<string | null> {
  return page.evaluate(
    ({ key, store }) => window[store].getItem(key),
    { key: SESSION_KEY, store: storage },
  )
}

/** Fills the login form and submits it. */
async function submitLogin(page: Page, email: string, password: string, remember: boolean) {
  await page.locator("#email").fill(email)
  await page.locator("#password").fill(password)
  if (remember) {
    await page.locator("#remember").check()
  } else {
    await page.locator("#remember").uncheck()
  }
  await page.getByRole("button", { name: "Se connecter" }).click()
}

test.describe("Connexion unifiée", () => {
  test("CUJ1 — le superadmin se connecte sans rester connecté", async ({ page }) => {
    await page.goto("/login")
    await submitLogin(page, ADMIN_EMAIL, ADMIN_PASSWORD, false)

    await expect(page.getByRole("heading", { name: "Admin", level: 1 })).toBeVisible()
    await expect(page.getByText("Bienvenue Prénom Nom")).toBeVisible()

    // A session that is not remembered lives in the session storage only.
    const stored = await readStoredSession(page, "sessionStorage")
    expect(stored).not.toBeNull()
    expect(JSON.parse(stored as string).kind).toBe("superadmin")
    expect(await readStoredSession(page, "localStorage")).toBeNull()
  })

  test("CUJ2 — le tenant reste connecté après un rechargement", async ({ page }) => {
    await page.goto("/login")
    await submitLogin(page, OWNER_EMAIL, OWNER_PASSWORD, true)

    await expect(page.getByRole("heading", { name: "Tenant", level: 1 })).toBeVisible()
    await expect(page.getByText("Bienvenue Camille Laurent")).toBeVisible()

    // A remembered session lives in the local storage only.
    const stored = await readStoredSession(page, "localStorage")
    expect(stored).not.toBeNull()
    expect(JSON.parse(stored as string).kind).toBe("tenant_user")
    expect(await readStoredSession(page, "sessionStorage")).toBeNull()

    await page.reload()
    await expect(page.getByRole("heading", { name: "Tenant", level: 1 })).toBeVisible()
    await expect(page.getByText("Bienvenue Camille Laurent")).toBeVisible()
  })

  test("CUJ3 — un mot de passe erroné est refusé sans session", async ({ page }) => {
    await page.goto("/login")
    await submitLogin(page, "ghost@atrium.dev", "wrong-password", false)

    await expect(page.getByRole("alert")).toHaveText(LOGIN_ERROR_TEXT)
    expect(await readStoredSession(page, "sessionStorage")).toBeNull()
    expect(await readStoredSession(page, "localStorage")).toBeNull()
  })
})
