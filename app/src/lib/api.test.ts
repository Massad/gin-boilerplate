import { afterAll, afterEach, beforeAll, beforeEach, describe, expect, it } from "vitest"
import { delay, http, HttpResponse } from "msw"
import { setupServer } from "msw/node"
import { purgeSession, readSession, writeSession, type TenantUserSessionData } from "@/auth/session"
import { ApiError, request } from "@/lib/api"

const server = setupServer()

beforeAll(() => {
  server.listen()
})

afterEach(() => {
  server.resetHandlers()
  purgeSession()
})

afterAll(() => {
  server.close()
})

const tenantSession = {
  kind: "tenant_user" as const,
  user: { id: "u-1", email: "camille@studio-nord.fr", name: "Camille Laurent", tenant_id: "t-1", role: "owner" },
  tenant: { id: "t-1", slug: "studio-nord", name: "Studio Nord" },
  token: { access_token: "stale-access", refresh_token: "refresh-1" },
  remember: true,
}

const superadminSession = {
  kind: "superadmin" as const,
  admin: { id: "a-1", email: "admin@atrium.dev", name: "Laure Moreau" },
  token: { access_token: "admin-access" },
  remember: false,
}

beforeEach(() => {
  window.localStorage.clear()
  window.sessionStorage.clear()
})

describe("request", () => {
  it("sends the stored bearer token and returns the parsed JSON", async () => {
    writeSession(tenantSession, { remember: true })
    let seenAuth: string | null = null
    server.use(
      http.get("/v1/protected", ({ request }) => {
        seenAuth = request.headers.get("Authorization")
        return HttpResponse.json({ ok: true, id: 7 })
      }),
    )

    await expect(request<{ ok: boolean; id: number }>("/protected")).resolves.toEqual({ ok: true, id: 7 })
    expect(seenAuth).toBe("Bearer stale-access")
  })

  it("throws an ApiError with the status and body on non-2xx", async () => {
    server.use(
      http.post("/v1/login", () =>
        HttpResponse.json({ message: "Invalid login details" }, { status: 406 }),
      ),
    )

    const error = await request("/login", { method: "POST", json: {} }).catch(
      (caught: unknown) => caught,
    )
    expect(error).toBeInstanceOf(ApiError)
    expect((error as ApiError).status).toBe(406)
    expect((error as ApiError).body).toEqual({ message: "Invalid login details" })
  })

  it("refreshes once, replays once and replaces the stored tokens on a tenant 401", async () => {
    writeSession(tenantSession, { remember: true })
    let refreshCount = 0
    const seenAuths: string[] = []
    server.use(
      http.get("/v1/protected", ({ request }) => {
        const auth = request.headers.get("Authorization")
        seenAuths.push(auth ?? "")
        // A call carrying the stale token is rejected; only replays with the
        // refreshed token go through.
        if (auth === "Bearer stale-access") {
          return HttpResponse.json({ message: "expired" }, { status: 401 })
        }
        return HttpResponse.json({ ok: true })
      }),
      http.post("/v1/token/refresh", async () => {
        refreshCount += 1
        // Hold the refresh so both concurrent 401s share the in-flight
        // promise before it settles.
        await delay(50)
        return HttpResponse.json({ access_token: "fresh-access", refresh_token: "refresh-2" })
      }),
    )

    const [first, second] = await Promise.all([
      request("/protected"),
      request("/protected"),
    ])

    expect(first).toEqual({ ok: true })
    expect(second).toEqual({ ok: true })
    expect(refreshCount).toBe(1)
    // Both concurrent calls start with the stale token, then replay with the
    // fresh one exactly once.
    expect(seenAuths.filter((auth) => auth === "Bearer stale-access")).toHaveLength(2)
    expect(seenAuths.filter((auth) => auth === "Bearer fresh-access")).toHaveLength(2)
     // The refreshed tokens replace the stored pair.
     const session = readSession()
     expect(session).not.toBeNull()
     expect((session as TenantUserSessionData).token.access_token).toBe("fresh-access")
     expect((session as TenantUserSessionData).token.refresh_token).toBe("refresh-2")
  })

  it("purges the session when the refresh fails", async () => {
    writeSession(tenantSession, { remember: true })
    let refreshCount = 0
    server.use(
      http.get("/v1/protected", () => HttpResponse.json({ message: "expired" }, { status: 401 })),
      http.post("/v1/token/refresh", () => {
        refreshCount += 1
        return HttpResponse.json({ message: "Invalid authorization, please login again" }, { status: 401 })
      }),
    )

    await expect(request("/protected")).rejects.toBeInstanceOf(ApiError)
    expect(refreshCount).toBe(1)
    expect(readSession()).toBeNull()
  })

  it("purges the session without refreshing on a superadmin 401", async () => {
    writeSession(superadminSession, { remember: false })
    let refreshCount = 0
    server.use(
      http.get("/v1/admin/tenants", () => HttpResponse.json({ message: "Please login again" }, { status: 401 })),
      http.post("/v1/token/refresh", () => {
        refreshCount += 1
        return HttpResponse.json({})
      }),
    )

    await expect(request("/admin/tenants")).rejects.toBeInstanceOf(ApiError)
    expect(refreshCount).toBe(0)
    expect(readSession()).toBeNull()
  })
})
