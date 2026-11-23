import { beforeEach, describe, expect, it } from "vitest"
import { purgeSession, readSession, writeSession } from "@/auth/session"
import type { SessionData } from "@/auth/session"

const tenantSession: SessionData = {
  kind: "tenant_user",
  user: { id: "u-1", email: "camille@studio-nord.fr", name: "Camille Laurent", tenant_id: "t-1", role: "owner" },
  tenant: { id: "t-1", slug: "studio-nord", name: "Studio Nord" },
  token: { access_token: "access-1", refresh_token: "refresh-1" },
}

const superadminSession: SessionData = {
  kind: "superadmin",
  admin: { id: "a-1", email: "admin@atrium.dev", name: "Laure Moreau" },
  token: { access_token: "admin-access-1" },
}

beforeEach(() => {
  window.localStorage.clear()
  window.sessionStorage.clear()
})

describe("writeSession", () => {
  it("stores a remembered session in localStorage", () => {
    writeSession(tenantSession, { remember: true })

    expect(window.localStorage.getItem("atrium.session")).toContain("camille@studio-nord.fr")
    expect(window.sessionStorage.getItem("atrium.session")).toBeNull()
  })

  it("stores a non-remembered session in sessionStorage", () => {
    writeSession(tenantSession, { remember: false })

    expect(window.sessionStorage.getItem("atrium.session")).toContain("camille@studio-nord.fr")
    expect(window.localStorage.getItem("atrium.session")).toBeNull()
  })

  it("moves a superadmin session to the right storage on rewrite", () => {
    writeSession(superadminSession, { remember: true })
    writeSession(superadminSession, { remember: false })

    expect(window.sessionStorage.getItem("atrium.session")).toContain("superadmin")
    expect(window.localStorage.getItem("atrium.session")).toBeNull()
  })
})

describe("readSession", () => {
  it("round-trips a stored session with its remember choice", () => {
    writeSession(tenantSession, { remember: true })

    expect(readSession()).toEqual({ ...tenantSession, remember: true })
  })

  it("falls back from sessionStorage to localStorage", () => {
    writeSession(tenantSession, { remember: true })
    purgeSession()

    window.localStorage.setItem("atrium.session", JSON.stringify({ ...tenantSession, remember: true }))

    expect(readSession()).toEqual({ ...tenantSession, remember: true })
  })

  it("prefers sessionStorage over localStorage", () => {
    window.sessionStorage.setItem("atrium.session", JSON.stringify({ ...tenantSession, remember: false }))
    window.localStorage.setItem("atrium.session", JSON.stringify({ ...superadminSession, remember: true }))

    expect(readSession()!.kind).toBe("tenant_user")
  })

  it("returns null when nothing is stored", () => {
    expect(readSession()).toBeNull()
  })

  it("returns null on a corrupted payload", () => {
    window.sessionStorage.setItem("atrium.session", "not json")

    expect(readSession()).toBeNull()
  })
})

describe("purgeSession", () => {
  it("removes the session from both storages", () => {
    writeSession(tenantSession, { remember: true })
    window.sessionStorage.setItem("atrium.session", JSON.stringify({ ...superadminSession, remember: false }))

    purgeSession()

    expect(window.localStorage.getItem("atrium.session")).toBeNull()
    expect(window.sessionStorage.getItem("atrium.session")).toBeNull()
    expect(readSession()).toBeNull()
  })
})
